package node

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/node/collectors"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower"
	wc "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/collectors"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// Config
const (
	minTasksInterval            time.Duration = time.Minute * 4
	maxTasksInterval            time.Duration = time.Minute * 6
	taskCooldown                time.Duration = time.Second * 5
	totalEffectiveStakeCooldown time.Duration = time.Hour * 1
	metricsShutdownTimeout      time.Duration = time.Second * 5
)

type waitUntilReadyResult int

const (
	waitUntilReadyExit waitUntilReadyResult = iota
	waitUntilReadyContinue
	waitUntilReadySuccess
)

type TaskLoop struct {
	// Services
	logger            *log.Logger
	ctx               context.Context
	sp                *services.ServiceProvider
	wg                *sync.WaitGroup
	cfg               *config.SmartNodeConfig
	rp                *rocketpool.RocketPool
	ec                eth.IExecutionClient
	bc                beacon.IBeaconClient
	metricsServer     *http.Server
	stateLocker       *collectors.StateLocker
	stateMgr          *state.NetworkStateManager
	watchtowerTaskMgr *watchtower.TaskManager

	// Tasks
	manageFeeRecipient      *ManageFeeRecipient
	distributeMinipools     *DistributeMinipools
	stakePrelaunchMinipools *StakePrelaunchMinipools
	promoteMinipools        *PromoteMinipools
	downloadRewardsTrees    *DownloadRewardsTrees
	reduceBonds             *ReduceBonds
	defendPdaoProps         *DefendPdaoProps
	verifyPdaoProps         *VerifyPdaoProps

	// Watchtower metrics
	scrubCollector         *wc.ScrubCollector
	bondReductionCollector *wc.BondReductionCollector
	soloMigrationCollector *wc.SoloMigrationCollector

	// Internal
	wasExecutionClientSynced    bool
	wasBeaconClientSynced       bool
	lastTotalEffectiveStakeTime time.Time
	secondsDelta                float64
}

func NewTaskLoop(sp *services.ServiceProvider, wg *sync.WaitGroup) *TaskLoop {
	logger := sp.GetTasksLogger()
	ctx := logger.CreateContextWithLogger(sp.GetBaseContext())
	t := &TaskLoop{
		sp:                          sp,
		logger:                      logger,
		ctx:                         ctx,
		wg:                          wg,
		cfg:                         sp.GetConfig(),
		rp:                          sp.GetRocketPool(),
		ec:                          sp.GetEthClient(),
		bc:                          sp.GetBeaconClient(),
		stateLocker:                 collectors.NewStateLocker(),
		lastTotalEffectiveStakeTime: time.Unix(0, 0),
		manageFeeRecipient:          NewManageFeeRecipient(ctx, sp, logger),
		distributeMinipools:         NewDistributeMinipools(sp, logger),
		stakePrelaunchMinipools:     NewStakePrelaunchMinipools(sp, logger),
		promoteMinipools:            NewPromoteMinipools(sp, logger),
		downloadRewardsTrees:        NewDownloadRewardsTrees(sp, logger),
		reduceBonds:                 NewReduceBonds(sp, logger),
		defendPdaoProps:             NewDefendPdaoProps(ctx, sp, logger),
		scrubCollector:              wc.NewScrubCollector(),
		bondReductionCollector:      wc.NewBondReductionCollector(),
		soloMigrationCollector:      wc.NewSoloMigrationCollector(),

		// We assume clients are synced on startup so that we don't send unnecessary alerts
		wasExecutionClientSynced: true,
		wasBeaconClientSynced:    true,

		// Delta between min and max to wait between loops
		secondsDelta: (maxTasksInterval - minTasksInterval).Seconds(),
	}

	// Create the prop verifier if the user enabled it
	if t.cfg.VerifyProposals.Value {
		t.verifyPdaoProps = NewVerifyPdaoProps(t.ctx, t.sp, t.logger)
	}

	return t
}

// Run the daemon task loop
func (t *TaskLoop) Run() error {
	// Print the current mode
	if t.cfg.IsNativeMode {
		fmt.Println("Starting node daemon in Native Mode.")
	} else {
		fmt.Println("Starting node daemon in Docker Mode.")
	}

	// Handle the initial fee recipient file deployment
	err := deployDefaultFeeRecipientFile(t.cfg)
	if err != nil {
		return err
	}

	// Run task loop
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		for {
			// Make sure all of the resources are ready for task processing
			readyResult := t.waitUntilReady()
			switch readyResult {
			case waitUntilReadyExit:
				return
			case waitUntilReadyContinue:
				continue
			}

			// === Task execution ===
			if t.runTasks() {
				return
			}
		}
	}()

	// Run metrics loop
	t.metricsServer = runMetricsServer(t.ctx, t.sp, t.logger, t.stateLocker, t.wg, t.scrubCollector, t.bondReductionCollector, t.soloMigrationCollector)

	return nil
}

// Stop the daemon
func (t *TaskLoop) Stop() {
	if t.metricsServer != nil {
		// Shut down the metrics server
		ctx, cancel := context.WithTimeout(context.Background(), metricsShutdownTimeout)
		defer cancel()
		t.metricsServer.Shutdown(ctx)
	}
}

// Wait until the chains and other resources are ready to be queried
// Returns true if the owning loop needs to exit, false if it can continue
func (t *TaskLoop) waitUntilReady() waitUntilReadyResult {
	// Check the EC status
	err := t.sp.WaitEthClientSynced(t.ctx, false) // Force refresh the primary / fallback EC status
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		t.wasExecutionClientSynced = false
		t.logger.Error("Execution Client not synced. Waiting for sync...", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	if !t.wasExecutionClientSynced {
		t.logger.Info("Execution Client is now synced.")
		t.wasExecutionClientSynced = true
		alerting.AlertExecutionClientSyncComplete(t.cfg)
	}

	// Check the BC status
	err = t.sp.WaitBeaconClientSynced(t.ctx, false) // Force refresh the primary / fallback BC status
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		// NOTE: if not synced, it returns an error - so there isn't necessarily an underlying issue
		t.wasBeaconClientSynced = false
		t.logger.Error("Beacon Node not synced. Waiting for sync...", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	if !t.wasBeaconClientSynced {
		t.logger.Info("Beacon Node is now synced.")
		t.wasBeaconClientSynced = true
		alerting.AlertBeaconClientSyncComplete(t.cfg)
	}

	// Load contracts
	err = t.sp.RefreshRocketPoolContracts()
	if err != nil {
		t.logger.Error("Error loading contract bindings", log.Err(err))
		return t.sleepAndReturnReadyResult()
	}

	// Wait until node is registered
	err = t.sp.WaitNodeRegistered(t.ctx, true)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context canceled") {
			return waitUntilReadyExit
		}
		t.logger.Error("Error waiting for node registration", slog.String(log.ErrorKey, errMsg))
		return t.sleepAndReturnReadyResult()
	}

	// Create the network state manager
	if t.stateMgr == nil {
		t.stateMgr, err = state.NewNetworkStateManager(t.ctx, t.rp, t.cfg, t.ec, t.bc, t.logger.Logger)
		if err != nil {
			t.logger.Error("Error creating network state manager", log.Err(err))
			return t.sleepAndReturnReadyResult()
		}
	}

	// Create the watchtower task manager
	if t.watchtowerTaskMgr == nil {
		t.watchtowerTaskMgr = watchtower.NewTaskManager(t.sp, t.stateMgr, t.scrubCollector, t.bondReductionCollector, t.soloMigrationCollector)
		err = t.watchtowerTaskMgr.Initialize(t.stateMgr)
		if err != nil {
			t.logger.Error("Error creating watchtower task manager", log.Err(err))
			return t.sleepAndReturnReadyResult()
		}
	}

	return waitUntilReadySuccess
}

// Sleep on the context for the task cooldown time, and return either exit or continue
// based on whether the context was cancelled.
func (t *TaskLoop) sleepAndReturnReadyResult() waitUntilReadyResult {
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return waitUntilReadyExit
	} else {
		return waitUntilReadyContinue
	}
}

// Runs an iteration of the node tasks.
// Returns true if the task loop should exit, false if it should continue.
func (t *TaskLoop) runTasks() bool {
	nodeAddress, hasAddress := t.sp.GetWallet().GetAddress()
	if !hasAddress {
		t.logger.Error("Node address not set")
		return utils.SleepWithCancel(t.ctx, taskCooldown)
	}

	// Get the Beacon block
	latestBlock, err := t.stateMgr.GetLatestBeaconBlock(t.ctx)
	if err != nil {
		t.logger.Error("error getting latest Beacon block", log.Err(err))
		return utils.SleepWithCancel(t.ctx, taskCooldown)
	}

	// Check if on the Oracle DAO
	isOnOdao, err := isOnOracleDAO(t.rp, nodeAddress, latestBlock)
	if err != nil {
		t.logger.Error(err.Error())
		return utils.SleepWithCancel(t.ctx, taskCooldown)
	}

	// Get the latest appropriate state
	var state *state.NetworkState
	var totalEffectiveStake *big.Int
	if isOnOdao {
		// Get the state of the entire network
		state, err = t.stateMgr.GetStateForSlot(t.ctx, latestBlock.Header.Slot)
		totalEffectiveStake = calculateTotalEffectiveStakeForNetwork(state)
		t.lastTotalEffectiveStakeTime = time.Now()
	} else {
		updateTotalEffectiveStake := false
		if time.Since(t.lastTotalEffectiveStakeTime) > totalEffectiveStakeCooldown {
			updateTotalEffectiveStake = true
			t.lastTotalEffectiveStakeTime = time.Now()
		}
		state, totalEffectiveStake, err = t.stateMgr.GetNodeStateForSlot(t.ctx, nodeAddress, latestBlock.Header.Slot, updateTotalEffectiveStake)
	}
	if err != nil {
		t.logger.Error(err.Error())
		return utils.SleepWithCancel(t.ctx, taskCooldown)
	}
	t.stateLocker.UpdateState(state, totalEffectiveStake)

	// Run watchtower duties in parallel
	var watchtowerWg errgroup.Group
	defer watchtowerWg.Wait()
	watchtowerWg.Go(func() error {
		return t.watchtowerTaskMgr.Run(isOnOdao, state)
	})

	// Manage the fee recipient for the node
	if err := t.manageFeeRecipient.Run(state); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Run the rewards download check
	if err := t.downloadRewardsTrees.Run(state); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Run the pDAO proposal defender
	if err := t.defendPdaoProps.Run(state); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Run the pDAO proposal verifier
	if t.verifyPdaoProps != nil {
		if err := t.verifyPdaoProps.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return true
		}
	}

	// Run the minipool stake check
	if err := t.stakePrelaunchMinipools.Run(state); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Run the balance distribution check
	if err := t.distributeMinipools.Run(state); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Run the reduce bond check
	if err := t.reduceBonds.Run(state); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return true
	}

	// Run the minipool promotion check
	if err := t.promoteMinipools.Run(state); err != nil {
		t.logger.Error(err.Error())
	}

	// Wait for a random amount of time between the min and max durations
	randomSeconds := rand.Intn(int(t.secondsDelta))
	interval := minTasksInterval + time.Duration(randomSeconds)*time.Second
	return utils.SleepWithCancel(t.ctx, interval)
}

// Copy the default fee recipient file into the proper location
func deployDefaultFeeRecipientFile(cfg *config.SmartNodeConfig) error {
	feeRecipientPath := cfg.GetFeeRecipientFilePath()
	_, err := os.Stat(feeRecipientPath)
	if os.IsNotExist(err) {
		// Make sure the validators dir is created
		validatorsFolder := filepath.Dir(feeRecipientPath)
		err = os.MkdirAll(validatorsFolder, 0755)
		if err != nil {
			return fmt.Errorf("could not create validators directory: %w", err)
		}

		// Create the file
		rs := cfg.GetRocketPoolResources()
		var defaultFeeRecipientFileContents string
		if cfg.IsNativeMode {
			// Native mode needs an environment variable definition
			defaultFeeRecipientFileContents = fmt.Sprintf("FEE_RECIPIENT=%s", rs.RethAddress.Hex())
		} else {
			// Docker and Hybrid just need the address itself
			defaultFeeRecipientFileContents = rs.RethAddress.Hex()
		}
		err := os.WriteFile(feeRecipientPath, []byte(defaultFeeRecipientFileContents), 0664)
		if err != nil {
			return fmt.Errorf("could not write default fee recipient file to %s: %w", feeRecipientPath, err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking fee recipient file status: %w", err)
	}

	return nil
}

// Check if this node is on the Oracle DAO
func isOnOracleDAO(rp *rocketpool.RocketPool, nodeAddress common.Address, block beacon.BeaconBlock) (bool, error) {
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(block.ExecutionBlockNumber),
	}

	member, err := oracle.NewOracleDaoMember(rp, nodeAddress)
	if err != nil {
		return false, fmt.Errorf("error creating oDAO member binding: %w", err)
	}
	err = rp.Query(nil, opts, member.Exists)
	if err != nil {
		return false, fmt.Errorf("error checking if node is in the Oracle DAO for Beacon block %d, EL block %d: %w", block.Header.Slot, block.ExecutionBlockNumber, err)
	}
	return member.Exists.Get(), nil
}

// Calculates the total effective stake for the network from a state that captured the whole network,
// since the information normally required by GetTotalEffectiveRplStake() is already in the state.
func calculateTotalEffectiveStakeForNetwork(state *state.NetworkState) *big.Int {
	total := big.NewInt(0)
	for _, node := range state.NodeDetails {
		if node.EffectiveRPLStake.Cmp(node.MinimumRPLStake) > 0 {
			total.Add(total, node.EffectiveRPLStake)
		}
	}
	return total
}
