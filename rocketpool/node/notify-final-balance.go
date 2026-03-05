package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Notify final balance task
type notifyFinalBalance struct {
	c              *cli.Context
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
func newNotifyFinalBalance(c *cli.Context, logger log.ColorLogger) (*notifyFinalBalance, error) {

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

	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		priorityFee = eth.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Return task
	return &notifyFinalBalance{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		d:              d,
		gasThreshold:   gasThreshold,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}, nil

}

// Notify Final Balance
func (t *notifyFinalBalance) run(state *state.NetworkState) error {
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

	validatorDetailsToProve := make(map[uint32]beacon.ValidatorStatus)
	pubkeys := state.MegapoolToPubkeysMap[megapoolAddress]
	for _, pubkey := range pubkeys {
		validatorDetails, exists := state.MegapoolValidatorDetails[pubkey]
		if !exists {
			// Log
			t.log.Printlnf("Validator %s not found in the megapool validator details", pubkey.String())
			continue
		}

		validatorInfo, exists := state.MegapoolValidatorInfo[pubkey]
		if !exists {
			// Log
			t.log.Printlnf("Validator %s not found in the megapool validator info map", pubkey.String())
			continue
		}

		if validatorDetails.Status == beacon.ValidatorState_WithdrawalDone && validatorInfo.ValidatorInfo.Exiting && !validatorInfo.ValidatorInfo.Exited && validatorDetails.EffectiveBalance == 0 {
			validatorDetailsToProve[validatorInfo.ValidatorId] = validatorDetails
		}
	}

	// Check if there are any validators to notify
	if len(validatorDetailsToProve) == 0 {
		return nil
	}

	// Notify the validators
	for validatorId, validatorDetails := range validatorDetailsToProve {
		// Log
		t.log.Printlnf("The validator id %d needs a final balance proof", validatorId)

		err := t.createFinalBalanceProof(t.rp, mp, state, validatorId, validatorDetails, opts)
		// dont return if there was an error, just log it so we can continue with the next validator
		if err != nil {
			t.log.Printlnf("Error creating final balance proof for validator %d: %w", validatorId, err)
		}
	}

	// Return
	return nil

}

func (t *notifyFinalBalance) createFinalBalanceProof(rp *rocketpool.RocketPool, mp megapool.Megapool, state *state.NetworkState, validatorId uint32, validatorDetails beacon.ValidatorStatus, callopts *bind.CallOpts) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Crafting a final balance proof. This process can take several seconds and is CPU and memory intensive. If you don't see a [FINISHED] log entry your system may not have enough resources to perform this operation.")

	validatorIndexStr, err := t.bc.GetValidatorIndex(validatorDetails.Pubkey)
	if err != nil {
		return err
	}

	validatorIndex, err := strconv.ParseUint(validatorIndexStr, 10, 64)
	if err != nil {
		return err
	}

	slot := validatorDetails.WithdrawableEpoch * 32

	withdrawalProof, proofSlot, stateUsed, err := services.GetWithdrawalProofForSlot(t.c, slot, validatorIndex)
	if err != nil {
		return fmt.Errorf("error getting withdrawal proof for validator 0x%s (index: %d): %w", validatorDetails.Pubkey.String(), validatorIndex, err)
	}
	t.log.Printlnf("The Beacon WithdrawalSlot for validator index %d is: %d", validatorDetails.Index, withdrawalProof.WithdrawalSlot)

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(t.c, proofSlot, t.w, state.BeaconConfig, mp.GetAddress(), validatorDetails.Pubkey, stateUsed)
	if err != nil {
		t.log.Printlnf("There was an error during the proof creation process: %w", err)
		return err
	}

	withdrawal := megapool.Withdrawal{
		Index:                 withdrawalProof.WithdrawalIndex,
		ValidatorIndex:        validatorIndex,
		WithdrawalCredentials: withdrawalProof.WithdrawalAddress,
		AmountInGwei:          withdrawalProof.Amount.Uint64(),
	}

	finalBalanceProof := megapool.WithdrawalProof{
		WithdrawalSlot: withdrawalProof.WithdrawalSlot,
		WithdrawalNum:  uint16(withdrawalProof.IndexInWithdrawalsArray),
		Withdrawal:     withdrawal,
		Witnesses:      withdrawalProof.Witnesses,
	}

	t.log.Printlnf("The validator final balance proof has been successfully created.")

	// Get the gas limit
	gasInfo, err := megapool.EstimateNotifyFinalBalance(rp, mp.GetAddress(), validatorId, slotTimestamp, finalBalanceProof, validatorProof, slotProof, opts)
	if err != nil {
		t.log.Printlnf("Could not estimate the gas required to notify final balance on megapool validator %d: %w", validatorId, err)
		return err
	}
	gas := big.NewInt(int64(gasInfo.SafeGasLimit))
	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Call Notify Final Balance
	tx, err := megapool.NotifyFinalBalance(rp, mp.GetAddress(), validatorId, slotTimestamp, finalBalanceProof, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully notified validator %d final balance.", validatorId)

	// Return
	return nil
}
