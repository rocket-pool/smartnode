package node

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"

	"github.com/rocket-pool/smartnode/rocketpool/common/log"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/state"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet/keystore/lighthouse"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet/keystore/nimbus"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet/keystore/prysm"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet/keystore/teku"
	"github.com/rocket-pool/smartnode/rocketpool/node/collectors"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Config
var tasksInterval, _ = time.ParseDuration("5m")
var taskCooldown, _ = time.ParseDuration("10s")
var totalEffectiveStakeCooldown, _ = time.ParseDuration("1h")

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

// Run daemon
func Run(sp *services.ServiceProvider) error {
	// Get services
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()
	bc := sp.GetBeaconClient()

	// Handle the initial fee recipient file deployment
	err := deployDefaultFeeRecipientFile(cfg)
	if err != nil {
		return err
	}

	// Clean up old fee recipient files
	err = removeLegacyFeeRecipientFiles(cfg)
	if err != nil {
		return err
	}

	// Wait until node is registered
	if err := sp.WaitNodeRegistered(true); err != nil {
		return err
	}

	// Print the current mode
	if cfg.IsNativeMode {
		fmt.Println("Starting node daemon in Native Mode.")
	} else {
		fmt.Println("Starting node daemon in Docker Mode.")
	}

	// Initialize loggers
	errorLog := log.NewColorLogger(ErrorColor)
	updateLog := log.NewColorLogger(UpdateColor)

	// Create the state manager
	m, err := state.NewNetworkStateManager(rp, cfg, ec, bc, &updateLog)
	if err != nil {
		return err
	}
	stateLocker := collectors.NewStateLocker()

	// Initialize tasks
	manageFeeRecipient := NewManageFeeRecipient(sp, log.NewColorLogger(ManageFeeRecipientColor))
	distributeMinipools := NewDistributeMinipools(sp, log.NewColorLogger(DistributeMinipoolsColor))
	stakePrelaunchMinipools := NewStakePrelaunchMinipools(sp, log.NewColorLogger(StakePrelaunchMinipoolsColor))
	promoteMinipools := NewPromoteMinipools(sp, log.NewColorLogger(PromoteMinipoolsColor))
	downloadRewardsTrees := NewDownloadRewardsTrees(sp, log.NewColorLogger(DownloadRewardsTreesColor))
	reduceBonds := NewReduceBonds(sp, log.NewColorLogger(ReduceBondAmountColor))
	defendPdaoProps := NewReduceBonds(sp, log.NewColorLogger(DefendPdaoPropsColor))
	var verifyPdaoProps *VerifyPdaoProps
	// Make sure the user opted into this duty
	verifyEnabled := cfg.Smartnode.VerifyProposals.Value.(bool)
	if verifyEnabled {
		verifyPdaoProps = NewVerifyPdaoProps(sp, log.NewColorLogger(VerifyPdaoPropsColor))
		if err != nil {
			return err
		}
	}

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Timestamp for caching total effective RPL stake
	lastTotalEffectiveStakeTime := time.Unix(0, 0)

	// Run task loop
	isHoustonDeployedMasterFlag := false
	go func() {
		for {
			// Check the EC status
			err := sp.WaitEthClientSynced(false) // Force refresh the primary / fallback EC status
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
				continue
			}

			// Check the BC status
			err = sp.WaitBeaconClientSynced(false) // Force refresh the primary / fallback BC status
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
				continue
			}

			// Update the network state
			updateTotalEffectiveStake := false
			if time.Since(lastTotalEffectiveStakeTime) > totalEffectiveStakeCooldown {
				updateTotalEffectiveStake = true
				lastTotalEffectiveStakeTime = time.Now() // Even if the call below errors out, this will prevent contant errors related to this flag
			}
			nodeAddress, hasNodeAddress := sp.GetWallet().GetAddress()
			if !hasNodeAddress {
				continue
			}
			state, totalEffectiveStake, err := updateNetworkState(m, &updateLog, nodeAddress, updateTotalEffectiveStake)
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
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
			time.Sleep(taskCooldown)

			// Run the rewards download check
			if err := downloadRewardsTrees.Run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			if state.IsHoustonDeployed {
				// Run the pDAO proposal defender
				if err := defendPdaoProps.Run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the pDAO proposal verifier
				if verifyPdaoProps != nil {
					if err := verifyPdaoProps.Run(state); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)
				}
			}

			// Run the minipool stake check
			if err := stakePrelaunchMinipools.Run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the balance distribution check
			if err := distributeMinipools.Run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the reduce bond check
			if err := reduceBonds.Run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the minipool promotion check
			if err := promoteMinipools.Run(state); err != nil {
				errorLog.Println(err)
			}

			time.Sleep(tasksInterval)
		}
	}()

	// Run metrics loop
	go func() {
		err := runMetricsServer(sp, log.NewColorLogger(MetricsColor), stateLocker)
		if err != nil {
			errorLog.Println(err)
		}
		wg.Done()
	}()

	// Wait for both threads to stop
	wg.Wait()
	return nil

}

// Copy the default fee recipient file into the proper location
func deployDefaultFeeRecipientFile(cfg *config.RocketPoolConfig) error {
	feeRecipientPath := cfg.Smartnode.GetFeeRecipientFilePath()
	_, err := os.Stat(feeRecipientPath)
	if os.IsNotExist(err) {
		// Make sure the validators dir is created
		validatorsFolder := filepath.Dir(feeRecipientPath)
		err = os.MkdirAll(validatorsFolder, 0755)
		if err != nil {
			return fmt.Errorf("could not create validators directory: %w", err)
		}

		// Create the file
		var defaultFeeRecipientFileContents string
		if cfg.IsNativeMode {
			// Native mode needs an environment variable definition
			defaultFeeRecipientFileContents = fmt.Sprintf("FEE_RECIPIENT=%s", cfg.Smartnode.GetRethAddress().Hex())
		} else {
			// Docker and Hybrid just need the address itself
			defaultFeeRecipientFileContents = cfg.Smartnode.GetRethAddress().Hex()
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

// Remove the old fee recipient files that were created in v1.5.0
func removeLegacyFeeRecipientFiles(cfg *config.RocketPoolConfig) error {
	legacyFeeRecipientFile := "rp-fee-recipient.txt"
	validatorsFolder := cfg.Smartnode.GetValidatorKeychainPath()

	// Remove the legacy files
	keystoreDirs := []string{lighthouse.KeystoreDir, nimbus.KeystoreDir, prysm.KeystoreDir, teku.KeystoreDir}
	for _, keystoreDir := range keystoreDirs {
		oldFile := filepath.Join(validatorsFolder, keystoreDir, legacyFeeRecipientFile)
		_, err := os.Stat(oldFile)
		if !os.IsNotExist(err) {
			err = os.Remove(oldFile)
			if err != nil {
				fmt.Printf("NOTE: Couldn't remove old fee recipient file (%s): %s\nThis file is no longer used, you may remove it manually if you wish.\n", oldFile, err.Error())
			}
		}
	}

	return nil
}

// Update the latest network state at each cycle
func updateNetworkState(m *state.NetworkStateManager, log *log.ColorLogger, nodeAddress common.Address, calculateTotalEffectiveStake bool) (*state.NetworkState, *big.Int, error) {
	// Get the state of the network
	state, totalEffectiveStake, err := m.GetHeadStateForNode(nodeAddress, calculateTotalEffectiveStake)
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
