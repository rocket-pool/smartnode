package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// canChallengePerformance checks whether the node can challenge the target-vote
// performance of a group of megapool validators: the node wallet must hold the
// performance_challenge_bond in RPL, and the challengeMegapool call must pass
// gas estimation with a fresh slot proof.
func canChallengePerformance(
	c *cli.Command,
	megapoolAddress common.Address,
	validatorIds []uint32,
	startEpoch uint64,
	participation []*big.Int,
) (*api.CanChallengeMegapoolPerformanceResponse, error) {
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	response := api.CanChallengeMegapoolPerformanceResponse{}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err = resolveChallengedMegapool(rp, nodeAccount.Address, megapoolAddress)
	if err != nil {
		return nil, err
	}

	// Check the node wallet holds the challenge bond in RPL before doing any
	// expensive proof work.
	response.ChallengeBond, err = protocol.GetPerformanceChallengeBond(rp, nil)
	if err != nil {
		return nil, err
	}
	response.RplBalance, err = tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node RPL balance: %w", err)
	}
	if response.RplBalance.Cmp(response.ChallengeBond) < 0 {
		response.InsufficientRplBalance = true
		return &response, nil
	}

	slotTimestamp, slotProof, err := getChallengeSlotProof(c, w, bc, rp, megapoolAddress, validatorIds)
	if err != nil {
		return nil, err
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	response.GasInfo, err = megapool.EstimateChallengeMegapoolGas(rp, megapoolAddress, validatorIds, startEpoch, participation, slotTimestamp, slotProof, opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating challengeMegapool gas: %w", err)
	}

	response.CanChallenge = true
	return &response, nil
}

// challengePerformance submits the challengeMegapool transaction for a group
// of megapool validators.
func challengePerformance(
	c *cli.Command,
	megapoolAddress common.Address,
	validatorIds []uint32,
	startEpoch uint64,
	participation []*big.Int,
	opts *bind.TransactOpts,
) (*api.ChallengeMegapoolPerformanceResponse, error) {
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	response := api.ChallengeMegapoolPerformanceResponse{}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	megapoolAddress, err = resolveChallengedMegapool(rp, nodeAccount.Address, megapoolAddress)
	if err != nil {
		return nil, err
	}

	slotTimestamp, slotProof, err := getChallengeSlotProof(c, w, bc, rp, megapoolAddress, validatorIds)
	if err != nil {
		return nil, err
	}

	response.TxHash, err = megapool.ChallengeMegapool(rp, megapoolAddress, validatorIds, startEpoch, participation, slotTimestamp, slotProof, opts)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// resolveChallengedMegapool returns the megapool address to challenge,
// resolving the zero address to the node's own megapool.
func resolveChallengedMegapool(rp *rocketpool.RocketPool, nodeAddress common.Address, megapoolAddress common.Address) (common.Address, error) {
	if (megapoolAddress != common.Address{}) {
		return megapoolAddress, nil
	}
	megapoolAddress, err := node.GetMegapoolAddress(rp, nodeAddress, nil)
	if err != nil {
		return common.Address{}, fmt.Errorf("error looking up node's megapool address: %w", err)
	}
	if (megapoolAddress == common.Address{}) {
		return common.Address{}, fmt.Errorf("node has no megapool deployed; pass a megapool address to challenge")
	}
	return megapoolAddress, nil
}

// getChallengeSlotProof builds the beacon slot proof and timestamp required by
// challengeMegapool, anchored on the first challenged validator's pubkey (the
// challenge itself only consumes the slot proof, not the validator proof).
func getChallengeSlotProof(
	c *cli.Command,
	w wallet.Wallet,
	bc beacon.Client,
	rp *rocketpool.RocketPool,
	megapoolAddress common.Address,
	validatorIds []uint32,
) (uint64, megapool.SlotProof, error) {
	if len(validatorIds) == 0 {
		return 0, megapool.SlotProof{}, fmt.Errorf("no validator ids to challenge")
	}
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return 0, megapool.SlotProof{}, fmt.Errorf("error creating megapool binding for %s: %w", megapoolAddress.Hex(), err)
	}
	pubkey, err := mp.GetValidatorPubkey(validatorIds[0], nil)
	if err != nil {
		return 0, megapool.SlotProof{}, fmt.Errorf("error getting megapool validator %d pubkey: %w", validatorIds[0], err)
	}
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return 0, megapool.SlotProof{}, fmt.Errorf("error getting beacon config: %w", err)
	}
	_, slotTimestamp, slotProof, err := services.GetValidatorProof(c, 0, w, eth2Config, megapoolAddress, pubkey, nil)
	if err != nil {
		return 0, megapool.SlotProof{}, fmt.Errorf("error building slot proof: %w", err)
	}
	return slotTimestamp, slotProof, nil
}
