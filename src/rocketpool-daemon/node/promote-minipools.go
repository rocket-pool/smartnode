package node

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	rpstate "github.com/rocket-pool/rocketpool-go/v2/utils/state"

	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Promote minipools task
type PromoteMinipools struct {
	sp             *services.ServiceProvider
	logger         *slog.Logger
	cfg            *config.SmartNodeConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	mpMgr          *minipool.MinipoolManager
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
}

// Create promote minipools task
func NewPromoteMinipools(sp *services.ServiceProvider, logger *log.Logger) *PromoteMinipools {
	cfg := sp.GetConfig()
	log := logger.With(slog.String(keys.RoutineKey, "Promote Minipools"))
	maxFee, maxPriorityFee := getAutoTxInfo(cfg, log)
	return &PromoteMinipools{
		sp:             sp,
		logger:         log,
		cfg:            cfg,
		w:              sp.GetWallet(),
		rp:             sp.GetRocketPool(),
		gasThreshold:   cfg.AutoTxGasThreshold.Value,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
	}
}

// Stake prelaunch minipools
func (t *PromoteMinipools) Run(state *state.NetworkState) error {
	// Log
	t.logger.Info("Starting check for minipools to promote...")

	// Get prelaunch minipools
	nodeAddress, _ := t.w.GetAddress()
	minipools, err := t.getVacantMinipools(nodeAddress, state)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.logger.Info("Minipool are ready for promotion.", slog.Int(keys.CountKey, len(minipools)))
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	// Get promote minipools submissions
	txSubmissions := make([]*eth.TransactionSubmission, len(minipools))
	for i, mpd := range minipools {
		txSubmissions[i], err = t.createPromoteMinipoolTx(mpd)
		if err != nil {
			t.logger.Error("Error preparing submission to promote minipool.", slog.String(keys.MinipoolKey, mpd.MinipoolAddress.Hex()), log.Err(err))
			return err
		}
	}

	// Promote
	timeoutBig := state.NetworkDetails.MinipoolLaunchTimeout
	timeout := time.Duration(timeoutBig.Uint64()) * time.Second
	err = t.promoteMinipools(txSubmissions, minipools, timeout)
	if err != nil {
		return fmt.Errorf("error promoting minipools: %w", err)
	}

	// Return
	return nil
}

// Get vacant minipools
func (t *PromoteMinipools) getVacantMinipools(nodeAddress common.Address, state *state.NetworkState) ([]*rpstate.NativeMinipoolDetails, error) {
	vacantMinipools := []*rpstate.NativeMinipoolDetails{}

	// Get the scrub period
	scrubPeriod := state.NetworkDetails.PromotionScrubPeriod

	// Get the time of the target block
	block, err := t.rp.Client.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(state.ElBlockNumber))
	if err != nil {
		return nil, fmt.Errorf("error getting the latest block time: %w", err)
	}
	blockTime := time.Unix(int64(block.Time), 0)

	// Filter by vacancy
	mpds := state.MinipoolDetailsByNode[nodeAddress]
	for _, mpd := range mpds {
		if mpd.IsVacant && mpd.Status == types.MinipoolStatus_Prelaunch {
			creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
			remainingTime := creationTime.Add(scrubPeriod).Sub(blockTime)
			if remainingTime < 0 {
				vacantMinipools = append(vacantMinipools, mpd)
			} else {
				t.logger.Info(fmt.Sprintf("Minipool %s has %s left until it can be promoted.", mpd.MinipoolAddress.Hex(), remainingTime))
			}
		}
	}

	// Return
	return vacantMinipools, nil
}

// Get submission info for promoting a minipool
func (t *PromoteMinipools) createPromoteMinipoolTx(mpd *rpstate.NativeMinipoolDetails) (*eth.TransactionSubmission, error) {
	// Log
	t.logger.Info("Preparing to promote minipool...", slog.String(keys.MinipoolKey, mpd.MinipoolAddress.Hex()))

	// Get the updated minipool interface
	mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("cannot promote minipool %s because its delegate version is too low (v%d); please update the delegate to promote it", mpd.MinipoolAddress.Hex(), mpd.Version)
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return nil, err
	}

	// Get the tx info
	txInfo, err := mpv3.Promote(opts)
	if err != nil {
		return nil, fmt.Errorf("error getting promote minipool tx for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return nil, fmt.Errorf("simulating promote minipool tx for %s failed: %s", mpd.MinipoolAddress.Hex(), txInfo.SimulationResult.SimulationError)
	}

	submission, err := eth.CreateTxSubmissionFromInfo(txInfo, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating promote tx submission for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	return submission, nil
}

// Promote all available minipools
func (t *PromoteMinipools) promoteMinipools(submissions []*eth.TransactionSubmission, minipools []*rpstate.NativeMinipoolDetails, minipoolLaunchTimeout time.Duration) error {
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = gas.GetMaxFeeWeiForDaemon(t.logger)
		if err != nil {
			return err
		}
	}
	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee

	// Print the gas info
	forceSubmissions := []*eth.TransactionSubmission{}
	if !gas.PrintAndCheckGasInfoForBatch(submissions, true, t.gasThreshold, t.logger, maxFee) {
		// Check for the timeout buffers
		for i, mpd := range minipools {
			creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
			isDue, timeUntilDue := tx.IsTransactionDue(creationTime, minipoolLaunchTimeout)
			if !isDue {
				t.logger.Info(fmt.Sprintf("Time until promoting minipool %s will be forced for safety: %s", mpd.MinipoolAddress.Hex(), timeUntilDue))
				continue
			}
			t.logger.Info("NOTICE: Minipool has exceeded half of the timeout period, so it will be force-promoted at the current gas price.", slog.String(keys.MinipoolKey, mpd.MinipoolAddress.Hex()))
			forceSubmissions = append(forceSubmissions, submissions[i])
		}

		if len(forceSubmissions) == 0 {
			return nil
		}
		submissions = forceSubmissions
	}

	// Create callbacks
	callbacks := make([]func(err error), len(minipools))
	for i, mp := range minipools {
		callbacks[i] = func(err error) {
			alerting.AlertMinipoolPromoted(t.cfg, mp.MinipoolAddress, err == nil)
		}
	}

	// Print TX info and wait for them to be included in a block
	err = tx.PrintAndWaitForTransactionBatch(t.cfg, t.rp, t.logger, submissions, callbacks, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully promoted all minipools.")
	return nil
}
