package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	apitypes "github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
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

	// Check if the megapool is deployed
	deployed, err := megapool.GetMegapoolDeployed(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return err
	}
	if !deployed {
		return nil
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(t.rp, megapoolAddress, nil)
	if err != nil {
		return err
	}

	// Iterate over megapool validators checking whether they're ready to submit a final balance proof
	validatorCount, err := mp.GetValidatorCount(nil)
	if err != nil {
		return err
	}
	validatorInfo, err := services.GetMegapoolValidatorDetails(t.rp, t.bc, mp, megapoolAddress, uint32(validatorCount), opts)
	if err != nil {
		return err
	}

	// Get the beacon state
	beaconState, err := services.GetBeaconState(t.bc)
	if err != nil {
		return err
	}

	for i := uint32(0); i < uint32(validatorCount); i++ {
		if validatorInfo[i].BeaconStatus.Status == "withdrawal_done" && validatorInfo[i].Exiting && !validatorInfo[i].Exited && validatorInfo[i].BeaconStatus.EffectiveBalance == 0 {
			// Log
			t.log.Printlnf("The validator ID %d needs a final balance proof", validatorInfo[i].ValidatorId)

			t.createFinalBalanceProof(t.rp, mp, validatorInfo[i], state, types.ValidatorPubkey(validatorInfo[i].PubKey), beaconState, opts)
		}
	}

	// Return
	return nil

}

func (t *notifyFinalBalance) createFinalBalanceProof(rp *rocketpool.RocketPool, mp megapool.Megapool, validatorInfo apitypes.MegapoolValidatorDetails, state *state.NetworkState, validatorPubkey types.ValidatorPubkey, beaconState eth2.BeaconState, callopts *bind.CallOpts) error {

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	t.log.Printlnf("Crafting a final balance proof. This process can take several seconds and is CPU and memory intensive. If you don't see a [FINISHED] log entry your system may not have enough resources to perform this operation.")

	validatorIndexStr, err := t.bc.GetValidatorIndex(validatorPubkey)
	if err != nil {
		return err
	}

	validatorIndex, err := strconv.ParseUint(validatorIndexStr, 10, 64)
	if err != nil {
		return err
	}

	slot := validatorInfo.WithdrawableEpoch * 32

	withdrawalProof, proofSlot, stateUsed, err := services.GetWithdrawalProofForSlot(t.c, slot, validatorIndex)
	if err != nil {
		fmt.Printf("An error occurred: %s\n", err)
	}
	t.log.Printlnf("The Beacon WithdrawalSlot for validator ID %d is: %d", validatorInfo.ValidatorId, withdrawalProof.WithdrawalSlot)

	validatorProof, slotTimestamp, slotProof, err := services.GetValidatorProof(t.c, proofSlot, t.w, state.BeaconConfig, mp.GetAddress(), validatorPubkey, stateUsed)
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
	gasInfo, err := megapool.EstimateNotifyFinalBalance(rp, mp.GetAddress(), validatorInfo.ValidatorId, slotTimestamp, finalBalanceProof, validatorProof, slotProof, opts)
	if err != nil {
		t.log.Printlnf("Could not estimate the gas required to notify final balance on megapool validator %d: %w", validatorInfo.ValidatorId, err)
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
	tx, err := megapool.NotifyFinalBalance(rp, mp.GetAddress(), validatorInfo.ValidatorId, slotTimestamp, finalBalanceProof, validatorProof, slotProof, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, tx.Hash(), t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully notified validator %d final balance.", validatorInfo.ValidatorId)

	// Return
	return nil
}
