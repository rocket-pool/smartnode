package node

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	rpstate "github.com/rocket-pool/rocketpool-go/v2/utils/state"

	"github.com/rocket-pool/node-manager-core/beacon"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Stake prelaunch minipools task
type StakePrelaunchMinipools struct {
	sp             *services.ServiceProvider
	logger         *slog.Logger
	cfg            *config.SmartNodeConfig
	w              *wallet.Wallet
	vMgr           *validator.ValidatorManager
	rp             *rocketpool.RocketPool
	bc             beacon.IBeaconClient
	d              *client.Client
	mpMgr          *minipool.MinipoolManager
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
}

// Create stake prelaunch minipools task
func NewStakePrelaunchMinipools(sp *services.ServiceProvider, logger *log.Logger) *StakePrelaunchMinipools {
	cfg := sp.GetConfig()
	log := logger.With(slog.String(keys.RoutineKey, "Prelaunch Stake"))
	maxFee, maxPriorityFee := getAutoTxInfo(cfg, log)
	return &StakePrelaunchMinipools{
		sp:             sp,
		logger:         log,
		cfg:            sp.GetConfig(),
		w:              sp.GetWallet(),
		vMgr:           sp.GetValidatorManager(),
		rp:             sp.GetRocketPool(),
		bc:             sp.GetBeaconClient(),
		d:              sp.GetDocker(),
		gasThreshold:   cfg.AutoTxGasThreshold.Value,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
	}
}

// Stake prelaunch minipools
func (t *StakePrelaunchMinipools) Run(state *state.NetworkState) error {
	// Log
	t.logger.Info("Starting check for minipools to launch.")

	// Get prelaunch minipools
	nodeAddress, _ := t.w.GetAddress()
	minipools, err := t.getPrelaunchMinipools(nodeAddress, state)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.logger.Info("Minipools are ready for staking.", slog.Int(keys.CountKey, len(minipools)))
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Stake minipools
	txSubmissions := make([]*eth.TransactionSubmission, len(minipools))
	for i, mpd := range minipools {
		txSubmissions[i], err = t.createStakeMinipoolTx(mpd, state)
		if err != nil {
			t.logger.Error("Error preparing submission to stake minipool", slog.String(keys.MinipoolKey, mpd.MinipoolAddress.Hex()), log.Err(err))
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
			if err := validator.RestartValidator(t.cfg, t.bc, t.log, t.d); err != nil {
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
				t.logger.Info(fmt.Sprintf("Minipool %s has %s left until it can be staked.", mpd.MinipoolAddress.Hex(), remainingTime))
			}
		}
	}

	// Return
	return prelaunchMinipools, nil
}

// Get submission info for staking a minipool
func (t *StakePrelaunchMinipools) createStakeMinipoolTx(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState) (*eth.TransactionSubmission, error) {
	// Log
	t.logger.Info("Preparing to stake minipool...", slog.String(keys.MinipoolKey, mpd.MinipoolAddress.Hex()))

	// Get the updated minipool interface
	mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Get minipool withdrawal credentials
	withdrawalCredentials := mpd.WithdrawalCredentials

	// Get the validator key for the minipool
	validatorPubkey := mpd.Pubkey
	validatorKey, err := t.vMgr.LoadValidatorKey(validatorPubkey)
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
	rs := t.cfg.GetNetworkResources()
	depositData, err := nmc_validator.GetDepositData(validatorKey, withdrawalCredentials, rs.GenesisForkVersion, depositAmount, rs.EthNetworkName)
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Get the tx info
	signature := beacon.ValidatorSignature(depositData.Signature)
	depositDataRoot := common.BytesToHash(depositData.DepositDataRoot)
	txInfo, err := mp.Common().Stake(signature, depositDataRoot, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required to stake the minipool: %w", err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return nil, fmt.Errorf("simulating stake minipool tx for %s failed: %s", mpd.MinipoolAddress.Hex(), txInfo.SimulationResult.SimulationError)
	}

	submission, err := eth.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating stake tx submission for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	return submission, nil
}

// Stake all available minipools
func (t *StakePrelaunchMinipools) stakeMinipools(submissions []*eth.TransactionSubmission, minipools []*rpstate.NativeMinipoolDetails, minipoolLaunchTimeout time.Duration) (bool, error) {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return false, err
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(t.logger)
		if err != nil {
			return false, err
		}
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee

	// Print the gas info
	forceSubmissions := []*eth.TransactionSubmission{}
	forceMinipools := []*rpstate.NativeMinipoolDetails{}
	if !gas.PrintAndCheckGasInfoForBatch(submissions, true, t.gasThreshold, t.logger, maxFee) {
		// Check for the timeout buffers
		for i, mpd := range minipools {
			prelaunchTime := time.Unix(mpd.StatusTime.Int64(), 0)
			isDue, timeUntilDue := tx.IsTransactionDue(prelaunchTime, minipoolLaunchTimeout)
			if !isDue {
				t.logger.Info(fmt.Sprintf("Time until staking minipool %s will be forced for safety: %s", mpd.MinipoolAddress.Hex(), timeUntilDue))
				alerting.AlertMinipoolStaked(t.cfg, mpd.MinipoolAddress, false)
				continue
			}
			t.logger.Warn("NOTICE: Minipool has exceeded half of the timeout period, so it will be force-staked at the current gas price.", slog.String(keys.MinipoolKey, mpd.MinipoolAddress.Hex()))
			forceSubmissions = append(forceSubmissions, submissions[i])
			forceMinipools = append(forceMinipools, mpd)
		}

		if len(forceSubmissions) == 0 {
			return false, nil
		}
		submissions = forceSubmissions
		minipools = forceMinipools
	}

	// Create callbacks
	callbacks := make([]func(err error), len(minipools))
	for i, mp := range minipools {
		callbacks[i] = func(err error) {
			alerting.AlertMinipoolStaked(t.cfg, mp.MinipoolAddress, err == nil)
		}
	}

	// Print TX info and wait for them to be included in a block
	err = tx.PrintAndWaitForTransactionBatch(t.cfg, t.rp, t.logger, submissions, callbacks, opts)
	if err != nil {
		return false, err
	}

	// Log
	t.logger.Info("Successfully staked all minipools.")
	return true, nil
}
