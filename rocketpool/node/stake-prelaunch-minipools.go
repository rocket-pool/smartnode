package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

// Stake prelaunch minipools task
type stakePrelaunchMinipools struct {
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

// Create stake prelaunch minipools task
func newStakePrelaunchMinipools(c *cli.Context, logger log.ColorLogger) (*stakePrelaunchMinipools, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetHdWallet(c)
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
	d, err := services.GetDocker(c)
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

	// Get the user-requested priority fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Return task
	return &stakePrelaunchMinipools{
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

// Stake prelaunch minipools
func (t *stakePrelaunchMinipools) run(state *state.NetworkState) error {

	// Reload the wallet (in case a call to `node deposit` changed it)
	if err := t.w.Reload(); err != nil {
		return err
	}

	// Log
	t.log.Println("Checking for minipools to launch...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get prelaunch minipools
	minipools, err := t.getPrelaunchMinipools(nodeAccount.Address, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for staking...", len(minipools))

	// Stake minipools
	successCount := 0
	for _, mpd := range minipools {
		success, err := t.stakeMinipool(mpd, state, opts)
		alerting.AlertMinipoolStaked(t.cfg, mpd.MinipoolAddress, success && err == nil)
		if err != nil {
			t.log.Println(fmt.Errorf("Could not stake minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
		if success {
			successCount++
		}
	}

	// Restart validator process if any minipools were staked successfully
	if successCount > 0 {
		if err := validator.RestartValidator(t.cfg, t.bc, &t.log, t.d); err != nil {
			return err
		}
	}

	// Return
	return nil

}

// Get prelaunch minipools
func (t *stakePrelaunchMinipools) getPrelaunchMinipools(nodeAddress common.Address, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {

	// Get the scrub period
	scrubPeriod := state.NetworkDetails.ScrubPeriod

	// Get the time of the target block
	block, err := t.rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("Can't get the latest block time: %w", err)
	}
	blockTime := time.Unix(int64(block.Time), 0)

	// Filter minipools by status
	prelaunchMinipools := []*rpstate.NativeMinipoolDetails{}
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
		if mpd.Status == rptypes.Prelaunch {
			if mpd.IsVacant {
				// Ignore vacant minipools
				continue
			}
			creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
			remainingTime := creationTime.Add(scrubPeriod).Sub(blockTime)
			if remainingTime < 0 {
				prelaunchMinipools = append(prelaunchMinipools, mpd)
			} else {
				t.log.Printlnf("Minipool %s has %s left until it can be staked.", mpd.MinipoolAddress.Hex(), remainingTime)
			}
		}
	}

	// Return
	return prelaunchMinipools, nil

}

// Stake a minipool
func (t *stakePrelaunchMinipools) stakeMinipool(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState, callOpts *bind.CallOpts) (bool, error) {

	// Log
	t.log.Printlnf("Staking minipool %s...", mpd.MinipoolAddress.Hex())

	mp, err := minipool.NewMinipoolFromVersion(t.rp, mpd.MinipoolAddress, mpd.Version, callOpts)
	if err != nil {
		return false, fmt.Errorf("cannot create binding for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Get minipool withdrawal credentials
	withdrawalCredentials := mpd.WithdrawalCredentials

	// Get the validator key for the minipool
	validatorPubkey := mpd.Pubkey
	validatorKey, err := t.w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return false, err
	}

	// Get the minipool type
	depositType := mpd.DepositType

	var depositAmount uint64
	switch depositType {
	case rptypes.Full, rptypes.Half, rptypes.Empty:
		depositAmount = uint64(16e9) // 16 ETH in gwei
	case rptypes.Variable:
		depositAmount = uint64(31e9) // 31 ETH in gwei
	default:
		return false, fmt.Errorf("error staking minipool %s: unknown deposit type %d", mpd.MinipoolAddress.Hex(), depositType)
	}

	// Get validator deposit data
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, state.BeaconConfig, depositAmount)
	if err != nil {
		return false, err
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)
	gasInfo, err := mp.EstimateStakeGas(signature, depositDataRoot, opts)
	if err != nil {
		return false, fmt.Errorf("Could not estimate the gas required to stake the minipool: %w", err)
	}
	var gas *big.Int
	if t.gasLimit != 0 {
		gas = new(big.Int).SetUint64(t.gasLimit)
	} else {
		gas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return false, err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		// Check for the timeout buffer
		prelaunchTime := time.Unix(mpd.StatusTime.Int64(), 0)
		isDue, timeUntilDue, err := api.IsTransactionDue(t.rp, prelaunchTime)
		if err != nil {
			t.log.Printlnf("Error checking if minipool is due: %s\nStaking now for safety...", err.Error())
		}
		if !isDue {
			t.log.Printlnf("Time until staking will be forced for safety: %s", timeUntilDue)
			return false, nil
		}

		t.log.Println("NOTICE: The minipool has exceeded half of the timeout period, so it will be force-staked at the current gas price.")
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Stake minipool
	hash, err := mp.Stake(
		signature,
		depositDataRoot,
		opts,
	)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully staked minipool %s.", mp.GetAddress().Hex())

	// Return
	return true, nil

}
