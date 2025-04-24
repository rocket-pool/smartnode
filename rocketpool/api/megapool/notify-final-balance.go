package megapool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canNotifyFinalBalance(c *cli.Context, validatorId uint32) (*api.CanNotifyFinalBalanceResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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

	// Response
	response := api.CanNotifyFinalBalanceResponse{}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	validatorInfo, err := mp.GetValidatorInfo(validatorId, nil)
	if err != nil {
		return nil, err
	}

	// Check validator status
	if !validatorInfo.Staked {
		response.InvalidStatus = true
	}

	// Update & return response
	response.CanNotify = !response.InvalidStatus
	return &response, nil

}

func notifyFinalBalance(c *cli.Context, validatorId uint32) (*api.NotifyFinalBalanceResponse, error) {

	// Get services
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NotifyFinalBalanceResponse{}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the validator pubkey
	validatorInfo, err := mp.GetValidatorInfo(validatorId, nil)
	if err != nil {
		return nil, err
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Get the first slot of the finalized epoch to use for the withdrawal proof
	head, err := bc.GetBeaconHead()
	if err != nil {
		return nil, err
	}

	referenceSlot := head.FinalizedEpoch * eth2Config.SlotsPerEpoch

	proof, err := services.GetWithdrawalProofForSlot(c, referenceSlot, validatorInfo.ValidatorIndex)
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	withdrawalObj := megapool.Withdrawal{
		Index:                 proof.WithdrawalIndex,
		AmountInGwei:          proof.Amount.Uint64(),
		ValidatorIndex:        big.NewInt(int64(proof.ValidatorIndex)),
		WithdrawalCredentials: proof.WithdrawalAddress,
	}

	// Notify the validator final balance
	hash, err := mp.NotifyFinalBalance(validatorId, big.NewInt(int64(proof.WithdrawalSlot)), big.NewInt(int64(proof.IndexInWithdrawalsArray)), withdrawalObj, big.NewInt(int64(referenceSlot)), proof.Proof, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
