package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"

	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool/common/log"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/state"
	"github.com/rocket-pool/smartnode/rocketpool/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool/common/validator"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Stake prelaunch minipools task
type StakePrelaunchMinipools struct {
	sp             *services.ServiceProvider
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.LocalWallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	d              *client.Client
	mpMgr          *minipool.MinipoolManager
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
}

// Create stake prelaunch minipools task
func NewStakePrelaunchMinipools(sp *services.ServiceProvider, logger log.ColorLogger) *StakePrelaunchMinipools {
	return &StakePrelaunchMinipools{
		sp:  sp,
		log: logger,
	}
}

// Stake prelaunch minipools
func (t *StakePrelaunchMinipools) Run(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.bc = t.sp.GetBeaconClient()
	t.w = t.sp.GetWallet()
	t.d = t.sp.GetDocker()
	nodeAddress, _ := t.w.GetAddress()
	t.maxFee, t.maxPriorityFee = getAutoTxInfo(t.cfg, &t.log)
	t.gasThreshold = t.cfg.Smartnode.AutoTxGasThreshold.Value.(float64)

	// Log
	t.log.Println("Checking for minipools to launch...")

	// Get prelaunch minipools
	minipools, err := t.getPrelaunchMinipools(nodeAddress, state)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for staking...", len(minipools))
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Stake minipools
	txSubmissions := make([]*core.TransactionSubmission, len(minipools))
	for i, mpd := range minipools {
		txSubmissions[i], err = t.createStakeMinipoolTx(mpd, state)
		if err != nil {
			t.log.Println(fmt.Errorf("error preparing submission to stake minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
	}

	// Stake
	timeoutBig := state.NetworkDetails.MinipoolLaunchTimeout
	timeout := time.Duration(timeoutBig.Uint64()) * time.Second
	_, err = t.stakeMinipools(txSubmissions, minipools, timeout)
	if err != nil {
		return fmt.Errorf("error staking minipools: %w", err)
	}

	/*
		NOTE: This is prompted by the CLI now, so automatic restarting may be obviated
		// Restart validator process if any minipools were staked successfully
		if stakedMinipools {
			if err := validator.RestartValidator(t.cfg, t.bc, &t.log, t.d); err != nil {
				return err
			}
		}
	*/

	// Return
	return nil

}

// Get prelaunch minipools
func (t *StakePrelaunchMinipools) getPrelaunchMinipools(nodeAddress common.Address, state *state.NetworkState) ([]*rpstate.NativeMinipoolDetails, error) {
	// Get the scrub period
	scrubPeriod := state.NetworkDetails.ScrubPeriod

	// Get the time of the target block
	block, err := t.rp.Client.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(state.ElBlockNumber))
	if err != nil {
		return nil, fmt.Errorf("error getting the latest block time: %w", err)
	}
	blockTime := time.Unix(int64(block.Time), 0)

	// Filter minipools by status
	prelaunchMinipools := []*rpstate.NativeMinipoolDetails{}
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
		if mpd.Status == rptypes.MinipoolStatus_Prelaunch {
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

// Get submission info for staking a minipool
func (t *StakePrelaunchMinipools) createStakeMinipoolTx(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState) (*core.TransactionSubmission, error) {
	// Log
	t.log.Printlnf("Preparing to stake minipool %s...", mpd.MinipoolAddress.Hex())

	// Get the updated minipool interface
	mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Get minipool withdrawal credentials
	withdrawalCredentials := mpd.WithdrawalCredentials

	// Get the validator key for the minipool
	validatorPubkey := mpd.Pubkey
	validatorKey, err := t.w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("error staking minipool %s: unknown deposit type %d", mpd.MinipoolAddress.Hex(), depositType)
	}

	// Get validator deposit data
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, state.BeaconConfig, depositAmount)
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Get the tx info
	signature := rptypes.BytesToValidatorSignature(depositData.Signature)
	txInfo, err := mp.Common().Stake(signature, depositDataRoot, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required to stake the minipool: %w", err)
	}
	if txInfo.SimError != "" {
		return nil, fmt.Errorf("simulating stake minipool tx for %s failed: %s", mpd.MinipoolAddress.Hex(), txInfo.SimError)
	}

	submission, err := core.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating stake tx submission for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	return submission, nil
}

// Stake all available minipools
func (t *StakePrelaunchMinipools) stakeMinipools(submissions []*core.TransactionSubmission, minipools []*rpstate.NativeMinipoolDetails, minipoolLaunchTimeout time.Duration) (bool, error) {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return false, err
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(&t.log)
		if err != nil {
			return false, err
		}
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee

	// Print the gas info
	forceSubmissions := []*core.TransactionSubmission{}
	if !gas.PrintAndCheckGasInfoForBatch(submissions, true, t.gasThreshold, &t.log, maxFee) {
		// Check for the timeout buffers
		for i, mpd := range minipools {
			prelaunchTime := time.Unix(mpd.StatusTime.Int64(), 0)
			isDue, timeUntilDue := tx.IsTransactionDue(prelaunchTime, minipoolLaunchTimeout)
			if !isDue {
				t.log.Printlnf("Time until staking minipool %s will be forced for safety: %s", mpd.MinipoolAddress.Hex(), timeUntilDue)
				continue
			}
			t.log.Printlnf("NOTICE: Minipool %s has exceeded half of the timeout period, so it will be force-staked at the current gas price.", mpd.MinipoolAddress.Hex())
			forceSubmissions = append(forceSubmissions, submissions[i])
		}

		if len(forceSubmissions) == 0 {
			return false, nil
		}
		submissions = forceSubmissions
	}

	// Print TX info and wait for them to be included in a block
	err = tx.PrintAndWaitForTransactionBatch(t.cfg, t.rp, &t.log, submissions, opts)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Println("Successfully staked all minipools.")
	return true, nil
}
