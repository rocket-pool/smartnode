package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/bindings/types"

	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

type finalBalanceCandidate struct {
	id     uint32
	pubkey types.ValidatorPubkey
}

// Notify final balance task
type notifyFinalBalance struct {
	c              *cli.Command
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	d              *client.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create notify final balance task
func newNotifyFinalBalance(c *cli.Command, logger log.ColorLogger) (*notifyFinalBalance, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	gas := loadAutoTxGas(cfg, &logger)

	// Return task
	return &notifyFinalBalance{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		d:              d,
		gasThreshold:   gas.thresholdGwei,
		maxFee:         gas.maxFee,
		maxPriorityFee: gas.maxPriorityFee,
		gasLimit:       0,
	}, nil

}

// Notify Final Balance
func (t *notifyFinalBalance) run(state *state.NetworkStateIndex) error {
	// Log
	t.log.Println("Checking if there are megapool validators with a final balance withdrawn...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	nodeDetails, exists := state.NodeDetailsByAddress[nodeAccount.Address]
	if !exists {
		return fmt.Errorf("node account %s not found in state", nodeAccount.Address.Hex())
	}

	if !nodeDetails.MegapoolDeployed {
		return nil
	}

	megapoolAddress := nodeDetails.MegapoolAddress

	mp, err := megapool.NewMegaPoolV1(t.rp, megapoolAddress, nil)
	if err != nil {
		return err
	}

	var candidates []finalBalanceCandidate
	pubkeys := state.MegapoolToPubkeysMap[megapoolAddress]
	for _, pubkey := range pubkeys {
		validatorInfo, exists := state.GetMegapoolValidatorInfo(megapoolAddress, pubkey)
		if !exists {
			t.log.Printlnf("Validator %s not found in the megapool validator info map", pubkey.String())
			continue
		}
		if !validatorInfo.ValidatorInfo.Staked {
			continue
		}
		if !validatorInfo.ValidatorInfo.Exiting || validatorInfo.ValidatorInfo.Exited {
			continue
		}
		candidates = append(candidates, finalBalanceCandidate{
			id:     validatorInfo.ValidatorId,
			pubkey: pubkey,
		})
	}

	if len(candidates) == 0 {
		return nil
	}

	head, err := t.bc.GetBeaconHead()
	if err != nil {
		return fmt.Errorf("error getting beacon head: %w", err)
	}
	finalizedEpoch := head.FinalizedEpoch
	candidatePubkeys := make([]types.ValidatorPubkey, len(candidates))
	for i, candidate := range candidates {
		candidatePubkeys[i] = candidate.pubkey
	}
	finalizedStatuses, err := t.bc.GetValidatorStatuses(candidatePubkeys, &beacon.ValidatorStatusOptions{Epoch: &finalizedEpoch})
	if err != nil {
		return fmt.Errorf("error getting finalized validator statuses: %w", err)
	}

	for _, candidate := range candidates {
		finalizedStatus := finalizedStatuses[candidate.pubkey]
		if !beacon.HasFinalBalanceWithdrawal(finalizedStatus) {
			t.log.Printlnf("Validator id %d is not ready for a final balance proof on the finalized beacon state (epoch %d): %s. Will retry on next cycle.", candidate.id, finalizedEpoch, finalBalancePendingReason(finalizedStatus, finalizedEpoch))
			continue
		}

		t.log.Printlnf("The validator id %d needs a final balance proof", candidate.id)
		err := t.createFinalBalanceProof(t.rp, mp, state, candidate.id, finalizedStatus, opts)
		if err != nil {
			t.log.Printlnf("Error creating final balance proof for validator %d: %s", candidate.id, err)
		}
	}

	// Return
	return nil

}

func (t *notifyFinalBalance) createFinalBalanceProof(rp *rocketpool.RocketPool, mp megapool.Megapool, state *state.NetworkStateIndex, validatorId uint32, validatorDetails beacon.ValidatorStatus, callopts *bind.CallOpts) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Crafting a final balance proof.")

	validatorIndex, err := strconv.ParseUint(validatorDetails.Index, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing the validator index: %w", err)
	}

	slotsPerEpoch := state.BeaconConfig.SlotsPerEpoch
	if slotsPerEpoch == 0 {
		slotsPerEpoch = 32
	}
	slot := validatorDetails.WithdrawableEpoch * slotsPerEpoch

	proof, err := services.BuildMegapoolFinalBalanceProof(t.c, rp, mp.GetAddress(), slot, validatorIndex, validatorDetails.Pubkey, t.w)
	if err != nil {
		return fmt.Errorf("error getting withdrawal proof for validator 0x%s (index: %d): %w", validatorDetails.Pubkey.String(), validatorIndex, err)
	}

	t.log.Printlnf("The validator final balance proof has been successfully created.")

	// Get the gas limit
	gasLimits, err := services.EstimateMegapoolNotifyFinalBalanceGas(rp, mp.GetAddress(), validatorId, proof, opts)
	if err != nil {
		t.log.Printlnf("Could not estimate the gas required to notify final balance on megapool validator %d: %s", validatorId, err)
		return err
	}
	gas := big.NewInt(int64(gasLimits.Safe))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !gasLimits.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Call Notify Final Balance
	tx, err := services.NotifyMegapoolFinalBalance(rp, mp.GetAddress(), validatorId, proof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = transactions.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully notified validator %d final balance.", validatorId)

	// Return
	return nil
}

func finalBalancePendingReason(status beacon.ValidatorStatus, currentEpoch uint64) string {
	if !status.Exists {
		return "validator not yet included in the finalized beacon state"
	}
	withdrawableEpoch := status.WithdrawableEpoch
	if withdrawableEpoch == 0 || withdrawableEpoch == beacon.FarFutureEpoch {
		return "withdrawable epoch not yet set on the finalized beacon state"
	}
	if currentEpoch < withdrawableEpoch {
		return fmt.Sprintf("waiting for withdrawable_epoch %d", withdrawableEpoch)
	}

	return beacon.FinalBalanceSweepNote(status)
}
