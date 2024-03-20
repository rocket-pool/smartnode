package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"

	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/alerting"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/node/collectors"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Config
const (
	tasksInterval               time.Duration = time.Minute * 5
	taskCooldown                time.Duration = time.Second * 10
	totalEffectiveStakeCooldown time.Duration = time.Hour * 1
	metricsShutdownTimeout      time.Duration = time.Second * 5
)

const (
	MaxConcurrentEth1Requests = 200

	StakePrelaunchMinipoolsColor = color.FgBlue
	DownloadRewardsTreesColor    = color.FgGreen
	MetricsColor                 = color.FgHiYellow
	ManageFeeRecipientColor      = color.FgHiCyan
	PromoteMinipoolsColor        = color.FgMagenta
	ReduceBondAmountColor        = color.FgHiBlue
	DefendPdaoPropsColor         = color.FgYellow
	VerifyPdaoPropsColor         = color.FgYellow
	DistributeMinipoolsColor     = color.FgHiGreen
	ErrorColor                   = color.FgRed
	WarningColor                 = color.FgYellow
	UpdateColor                  = color.FgHiWhite
)

type TaskLoop struct {
	ctx           context.Context
	cancel        context.CancelFunc
	sp            *services.ServiceProvider
	wg            *sync.WaitGroup
	metricsServer *http.Server
}

func NewTaskLoop(sp *services.ServiceProvider, wg *sync.WaitGroup) *TaskLoop {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskLoop{
		ctx:    ctx,
		cancel: cancel,
		sp:     sp,
		wg:     wg,
	}
}

// Run daemon
func (t *TaskLoop) Run() error {
	// Get services
	cfg := t.sp.GetConfig()
	rp := t.sp.GetRocketPool()
	ec := t.sp.GetEthClient()
	bc := t.sp.GetBeaconClient()

	// Print the current mode
	if cfg.IsNativeMode {
		fmt.Println("Starting node daemon in Native Mode.")
	} else {
		fmt.Println("Starting node daemon in Docker Mode.")
	}

	// Handle the initial fee recipient file deployment
	err := deployDefaultFeeRecipientFile(cfg)
	if err != nil {
		return err
	}

	// Initialize loggers
	errorLog := log.NewColorLogger(ErrorColor)
	updateLog := log.NewColorLogger(UpdateColor)

	// Create the state manager
	m, err := state.NewNetworkStateManager(t.ctx, rp, cfg, ec, bc, &updateLog)
	if err != nil {
		return err
	}
	stateLocker := collectors.NewStateLocker()

	// Initialize tasks
	manageFeeRecipient := NewManageFeeRecipient(t.ctx, t.sp, log.NewColorLogger(ManageFeeRecipientColor))
	distributeMinipools := NewDistributeMinipools(t.sp, log.NewColorLogger(DistributeMinipoolsColor))
	stakePrelaunchMinipools := NewStakePrelaunchMinipools(t.sp, log.NewColorLogger(StakePrelaunchMinipoolsColor))
	promoteMinipools := NewPromoteMinipools(t.sp, log.NewColorLogger(PromoteMinipoolsColor))
	downloadRewardsTrees := NewDownloadRewardsTrees(t.sp, log.NewColorLogger(DownloadRewardsTreesColor))
	reduceBonds := NewReduceBonds(t.sp, log.NewColorLogger(ReduceBondAmountColor))
	defendPdaoProps := NewDefendPdaoProps(t.ctx, t.sp, log.NewColorLogger(DefendPdaoPropsColor))
	var verifyPdaoProps *VerifyPdaoProps
	// Make sure the user opted into this duty
	verifyEnabled := cfg.VerifyProposals.Value
	if verifyEnabled {
		verifyPdaoProps = NewVerifyPdaoProps(t.ctx, t.sp, log.NewColorLogger(VerifyPdaoPropsColor))
		if err != nil {
			return err
		}
	}

	// Timestamp for caching total effective RPL stake
	lastTotalEffectiveStakeTime := time.Unix(0, 0)

	// Run task loop
	t.wg.Add(1)
	isHoustonDeployedMasterFlag := false
	go func() {
		defer t.wg.Done()

		// Wait until node is registered
		err := t.sp.WaitNodeRegistered(t.ctx, true)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				errorLog.Printlnf("error waiting for node registration: %s", err.Error())
			}
			return
		}

		// we assume clients are synced on startup so that we don't send unnecessary alerts
		wasExecutionClientSynced := true
		wasBeaconClientSynced := true
		for {
			// Check the EC status
			err := t.sp.WaitEthClientSynced(t.ctx, false) // Force refresh the primary / fallback EC status
			if err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}
				wasExecutionClientSynced = false
				errorLog.Printlnf("Execution Client not synced: %s. Waiting for sync...", err.Error())
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			if !wasExecutionClientSynced {
				updateLog.Println("Execution Client is now synced.")
				wasExecutionClientSynced = true
				alerting.AlertExecutionClientSyncComplete(cfg)
			}

			// Check the BC status
			err = t.sp.WaitBeaconClientSynced(t.ctx, false) // Force refresh the primary / fallback BC status
			if err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}
				// NOTE: if not synced, it returns an error - so there isn't necessarily an underlying issue
				wasBeaconClientSynced = false
				errorLog.Printlnf("Beacon Node not synced: %s. Waiting for sync...", err.Error())
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			if !wasBeaconClientSynced {
				updateLog.Println("Beacon Node is now synced.")
				wasBeaconClientSynced = true
				alerting.AlertBeaconClientSyncComplete(cfg)
			}

			// Load contracts
			err = t.sp.LoadContractsIfStale()
			if err != nil {
				errorLog.Println(fmt.Sprintf("error loading contract bindings: %s", err.Error()))
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Update the network state
			updateTotalEffectiveStake := false
			if time.Since(lastTotalEffectiveStakeTime) > totalEffectiveStakeCooldown {
				updateTotalEffectiveStake = true
				lastTotalEffectiveStakeTime = time.Now() // Even if the call below errors out, this will prevent contant errors related to this flag
			}
			nodeAddress, hasNodeAddress := t.sp.GetWallet().GetAddress()
			if !hasNodeAddress {
				continue
			}
			state, totalEffectiveStake, err := updateNetworkState(t.ctx, m, &updateLog, nodeAddress, updateTotalEffectiveStake)
			if err != nil {
				errorLog.Println(err)
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}
			stateLocker.UpdateState(state, totalEffectiveStake)

			// Check for Houston
			if !isHoustonDeployedMasterFlag && state.IsHoustonDeployed {
				printHoustonMessage(&updateLog)
				isHoustonDeployedMasterFlag = true
			}

			// Manage the fee recipient for the node
			if err := manageFeeRecipient.Run(state); err != nil {
				errorLog.Println(err)
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			// Run the rewards download check
			if err := downloadRewardsTrees.Run(state); err != nil {
				errorLog.Println(err)
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			if state.IsHoustonDeployed {
				// Run the pDAO proposal defender
				if err := defendPdaoProps.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the pDAO proposal verifier
				if verifyPdaoProps != nil {
					if err := verifyPdaoProps.Run(state); err != nil {
						errorLog.Println(err)
					}
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
				}
			}

			// Run the minipool stake check
			if err := stakePrelaunchMinipools.Run(state); err != nil {
				errorLog.Println(err)
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			// Run the balance distribution check
			if err := distributeMinipools.Run(state); err != nil {
				errorLog.Println(err)
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			// Run the reduce bond check
			if err := reduceBonds.Run(state); err != nil {
				errorLog.Println(err)
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			// Run the minipool promotion check
			if err := promoteMinipools.Run(state); err != nil {
				errorLog.Println(err)
			}

			if utils.SleepWithCancel(t.ctx, tasksInterval) {
				break
			}
		}
	}()

	// Run metrics loop
	t.metricsServer = runMetricsServer(t.ctx, t.sp, log.NewColorLogger(MetricsColor), stateLocker, t.wg)

	return nil
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

func (t *TaskLoop) Stop() {
	t.cancel()
	if t.metricsServer != nil {
		// Shut down the metrics server
		ctx, cancel := context.WithTimeout(context.Background(), metricsShutdownTimeout)
		defer cancel()
		t.metricsServer.Shutdown(ctx)
	}
}

// Update the latest network state at each cycle
func updateNetworkState(ctx context.Context, m *state.NetworkStateManager, log *log.ColorLogger, nodeAddress common.Address, calculateTotalEffectiveStake bool) (*state.NetworkState, *big.Int, error) {
	// Get the state of the network
	state, totalEffectiveStake, err := m.GetHeadStateForNode(ctx, nodeAddress, calculateTotalEffectiveStake)
	if err != nil {
		return nil, nil, fmt.Errorf("error updating network state: %w", err)
	}
	return state, totalEffectiveStake, nil
}

// Check if Houston has been deployed yet
func printHoustonMessage(log *log.ColorLogger) {
	log.Println(`
*       .
*      / \
*     |.'.|
*     |'.'|
*   ,'|   |'.
*  |,-'-|-'-.|
*   __|_| |         _        _      _____           _
*  | ___ \|        | |      | |    | ___ \         | |
*  | |_/ /|__   ___| | _____| |_   | |_/ /__   ___ | |
*  |    // _ \ / __| |/ / _ \ __|  |  __/ _ \ / _ \| |
*  | |\ \ (_) | (__|   <  __/ |_   | | | (_) | (_) | |
*  \_| \_\___/ \___|_|\_\___|\__|  \_|  \___/ \___/|_|
* +---------------------------------------------------+
* |    DECENTRALISED STAKING PROTOCOL FOR ETHEREUM    |
* +---------------------------------------------------+
*
* =============== Houston has launched! ===============
`)
}
