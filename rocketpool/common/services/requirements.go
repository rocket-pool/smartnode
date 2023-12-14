package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/node"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// Settings
const (
	EthClientSyncTimeout    int64 = 16 // 16 seconds
	BeaconClientSyncTimeout int64 = 16 // 16 seconds
)

var (
	checkNodePasswordInterval, _      = time.ParseDuration("15s")
	checkNodeWalletInterval, _        = time.ParseDuration("15s")
	checkRocketStorageInterval, _     = time.ParseDuration("15s")
	checkNodeRegisteredInterval, _    = time.ParseDuration("15s")
	ethClientSyncPollInterval, _      = time.ParseDuration("5s")
	beaconClientSyncPollInterval, _   = time.ParseDuration("5s")
	ethClientRecentBlockThreshold, _  = time.ParseDuration("5m")
	ethClientStatusRefreshInterval, _ = time.ParseDuration("60s")

	ethClientSyncLock    sync.Mutex
	beaconClientSyncLock sync.Mutex
)

// ====================
// === Requirements ===
// ====================

func (sp *ServiceProvider) RequireNodeAddress() error {
	status := sp.nodeWallet.GetStatus()
	if status == sharedtypes.WalletStatus_NoAddress {
		return errors.New("The node currently does not have an address set. Please run 'rocketpool wallet init' and try again.")
	}
	return nil
}

func (sp *ServiceProvider) RequireWalletReady() error {
	status := sp.nodeWallet.GetStatus()
	switch status {
	case sharedtypes.WalletStatus_NoAddress:
		return errors.New("The node currently does not have an address set. Please run 'rocketpool wallet init' and try again.")
	case sharedtypes.WalletStatus_NoKeystore:
		return errors.New("The node currently does not have a node wallet keystore. Please run 'rocketpool wallet init' and try again.")
	case sharedtypes.WalletStatus_NoPassword:
		return errors.New("The node's wallet password has not been set. Please run 'rocketpool wallet enter-password' first.")
	case sharedtypes.WalletStatus_KeystoreMismatch:
		return errors.New("The node's wallet keystore does not match the node address. This node is currently in read-only mode.")
	case sharedtypes.WalletStatus_Ready:
		return nil
	default:
		return fmt.Errorf("error checking if wallet is ready: unknown status [%v]", status)
	}
}

func (sp *ServiceProvider) RequireEthClientSynced() error {
	ethClientSynced, err := sp.waitEthClientSynced(false, EthClientSyncTimeout)
	if err != nil {
		return err
	}
	if !ethClientSynced {
		return errors.New("The Execution client is currently syncing. Please try again later.")
	}
	return nil
}

func (sp *ServiceProvider) RequireBeaconClientSynced() error {
	beaconClientSynced, err := sp.waitBeaconClientSynced(false, BeaconClientSyncTimeout)
	if err != nil {
		return err
	}
	if !beaconClientSynced {
		return errors.New("The Beacon client is currently syncing. Please try again later.")
	}
	return nil
}

func (sp *ServiceProvider) RequireNodeRegistered() error {
	if err := sp.RequireNodeAddress(); err != nil {
		return err
	}
	if err := sp.RequireEthClientSynced(); err != nil {
		return err
	}
	nodeRegistered, err := sp.getNodeRegistered()
	if err != nil {
		return err
	}
	if !nodeRegistered {
		return errors.New("The node is not registered with Rocket Pool. Please run 'rocketpool node register' and try again.")
	}
	return nil
}

func (sp *ServiceProvider) RequireRplFaucet() error {
	if sp.rplFaucet == nil {
		network := string(sp.cfg.Smartnode.Network.Value.(cfgtypes.Network))
		return fmt.Errorf("The RPL faucet is not available on the %s network.", network)
	}
	return nil
}

func (sp *ServiceProvider) RequireSnapshot() error {
	if sp.snapshotDelegation == nil {
		network := string(sp.cfg.Smartnode.Network.Value.(cfgtypes.Network))
		return fmt.Errorf("Snapshot voting is not available on the %s network.", network)
	}
	return nil
}

func (sp *ServiceProvider) RequireOnOracleDao() error {
	if err := sp.RequireNodeAddress(); err != nil {
		return err
	}
	if err := sp.RequireEthClientSynced(); err != nil {
		return err
	}
	nodeTrusted, err := sp.isMemberOfOracleDao()
	if err != nil {
		return err
	}
	if !nodeTrusted {
		return errors.New("The node is not a member of the oracle DAO. Nodes can only join the oracle DAO by invite.")
	}
	return nil
}

func (sp *ServiceProvider) RequireOnSecurityCouncil() error {
	if err := sp.RequireNodeAddress(); err != nil {
		return err
	}
	if err := sp.RequireEthClientSynced(); err != nil {
		return err
	}
	nodeTrusted, err := sp.isMemberOfSecurityCouncil()
	if err != nil {
		return err
	}
	if !nodeTrusted {
		return errors.New("The node is not a member of the security council. Nodes can only join the security council by invite.")
	}
	return nil
}

// ===============================
// === Service Synchronization ===
// ===============================

func (sp *ServiceProvider) WaitWalletReady(verbose bool) error {
	for {
		status := sp.nodeWallet.GetStatus()
		var message string
		switch status {
		case sharedtypes.WalletStatus_NoAddress:
			message = "The node currently does not have an address set"
		case sharedtypes.WalletStatus_NoKeystore:
			message = "The node currently does not have a node wallet keystore"
		case sharedtypes.WalletStatus_NoPassword:
			message = "The node's wallet password has not been set"
		case sharedtypes.WalletStatus_KeystoreMismatch:
			message = "The node's wallet keystore does not match the node address"
		case sharedtypes.WalletStatus_Ready:
			return nil
		default:
			message = fmt.Sprintf("error checking if wallet is ready: unknown status [%v]", status)
		}
		if status == sharedtypes.WalletStatus_Ready {
			return nil
		}
		if verbose {
			log.Printf("%s, retrying in %s...\n", message, checkNodeWalletInterval.String())
		}
		time.Sleep(checkNodeWalletInterval)
	}
}

// Wait for the Executon client to sync; timeout of 0 indicates no timeout
func (sp *ServiceProvider) WaitEthClientSynced(verbose bool) error {
	_, err := sp.waitEthClientSynced(verbose, 0)
	return err
}

// Wait for the Beacon client to sync; timeout of 0 indicates no timeout
func (sp *ServiceProvider) WaitBeaconClientSynced(verbose bool) error {
	_, err := sp.waitBeaconClientSynced(verbose, 0)
	return err
}

// Wait until the node has been registered with the Rocket Pool network
func (sp *ServiceProvider) WaitNodeRegistered(verbose bool) error {
	if err := sp.WaitWalletReady(verbose); err != nil {
		return err
	}
	if err := sp.WaitEthClientSynced(verbose); err != nil {
		return err
	}
	for {
		nodeRegistered, err := sp.getNodeRegistered()
		if err != nil {
			return err
		}
		if nodeRegistered {
			return nil
		}
		if verbose {
			log.Printf("The node is not registered with Rocket Pool, retrying in %s...\n", checkNodeRegisteredInterval.String())
		}
		time.Sleep(checkNodeRegisteredInterval)
	}
}

// ===============
// === Helpers ===
// ===============

// Check if the node is registered
func (sp *ServiceProvider) getNodeRegistered() (bool, error) {
	rp := sp.rocketPool
	address, _ := sp.nodeWallet.GetAddress()

	// Create a node binding
	node, err := node.NewNode(rp, address)
	if err != nil {
		return false, fmt.Errorf("error creating node binding: %w", err)
	}

	// Get contract state
	err = rp.Query(nil, nil, node.Exists)
	if err != nil {
		return false, fmt.Errorf("error getting node registration status: %w", err)
	}
	return node.Exists.Get(), nil
}

// Check if the node is a member of the oracle DAO
func (sp *ServiceProvider) isMemberOfOracleDao() (bool, error) {
	rp := sp.rocketPool
	address, _ := sp.nodeWallet.GetAddress()

	// Create the bindings
	odaoMember, err := oracle.NewOracleDaoMember(rp, address)
	if err != nil {
		return false, fmt.Errorf("error creating oDAO member binding: %w", err)
	}

	// Get contract state
	err = rp.Query(nil, nil, odaoMember.Exists)
	if err != nil {
		return false, fmt.Errorf("error getting oDAO member status: %w", err)
	}
	return odaoMember.Exists.Get(), nil
}

// Check if the node is a member of the security council
func (sp *ServiceProvider) isMemberOfSecurityCouncil() (bool, error) {
	rp := sp.rocketPool
	address, _ := sp.nodeWallet.GetAddress()

	// Create the bindings
	scMember, err := security.NewSecurityCouncilMember(rp, address)
	if err != nil {
		return false, fmt.Errorf("error creating security council member binding: %w", err)
	}

	// Get contract state
	err = rp.Query(nil, nil, scMember.Exists)
	if err != nil {
		return false, fmt.Errorf("error getting security council member status: %w", err)
	}
	return scMember.Exists.Get(), nil
}

// Check if the primary and fallback Execution clients are synced
func (sp *ServiceProvider) checkExecutionClientStatus() (bool, core.ExecutionClient, error) {
	// Check the EC status
	ecMgr := sp.ecManager
	cfg := sp.cfg
	mgrStatus := ecMgr.CheckStatus(cfg)
	if ecMgr.primaryReady {
		return true, nil, nil
	}

	// If the primary isn't synced but there's a fallback and it is, return true
	if ecMgr.fallbackReady {
		if mgrStatus.PrimaryClientStatus.Error != "" {
			log.Printf("Primary execution client is unavailable (%s), using fallback execution client...\n", mgrStatus.PrimaryClientStatus.Error)
		} else {
			log.Printf("Primary execution client is still syncing (%.2f%%), using fallback execution client...\n", mgrStatus.PrimaryClientStatus.SyncProgress*100)
		}
		return true, nil, nil
	}

	// If neither is synced, go through the status to figure out what to do

	// Is the primary working and syncing? If so, wait for it
	if mgrStatus.PrimaryClientStatus.IsWorking && mgrStatus.PrimaryClientStatus.Error == "" {
		log.Printf("Fallback execution client is not configured or unavailable, waiting for primary execution client to finish syncing (%.2f%%)\n", mgrStatus.PrimaryClientStatus.SyncProgress*100)
		return false, ecMgr.primaryEc, nil
	}

	// Is the fallback working and syncing? If so, wait for it
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackClientStatus.IsWorking && mgrStatus.FallbackClientStatus.Error == "" {
		log.Printf("Primary execution client is unavailable (%s), waiting for the fallback execution client to finish syncing (%.2f%%)\n", mgrStatus.PrimaryClientStatus.Error, mgrStatus.FallbackClientStatus.SyncProgress*100)
		return false, ecMgr.fallbackEc, nil
	}

	// If neither client is working, report the errors
	if mgrStatus.FallbackEnabled {
		return false, nil, fmt.Errorf("Primary execution client is unavailable (%s) and fallback execution client is unavailable (%s), no execution clients are ready.", mgrStatus.PrimaryClientStatus.Error, mgrStatus.FallbackClientStatus.Error)
	}

	return false, nil, fmt.Errorf("Primary execution client is unavailable (%s) and no fallback execution client is configured.", mgrStatus.PrimaryClientStatus.Error)
}

// Check if the primary and fallback Beacon clients are synced
func (sp *ServiceProvider) checkBeaconClientStatus() (bool, error) {
	// Check the BC status
	bcMgr := sp.bcManager
	mgrStatus := bcMgr.CheckStatus()
	if bcMgr.primaryReady {
		return true, nil
	}

	// If the primary isn't synced but there's a fallback and it is, return true
	if bcMgr.fallbackReady {
		if mgrStatus.PrimaryClientStatus.Error != "" {
			log.Printf("Primary consensus client is unavailable (%s), using fallback consensus client...\n", mgrStatus.PrimaryClientStatus.Error)
		} else {
			log.Printf("Primary consensus client is still syncing (%.2f%%), using fallback consensus client...\n", mgrStatus.PrimaryClientStatus.SyncProgress*100)
		}
		return true, nil
	}

	// If neither is synced, go through the status to figure out what to do

	// Is the primary working and syncing? If so, wait for it
	if mgrStatus.PrimaryClientStatus.IsWorking && mgrStatus.PrimaryClientStatus.Error == "" {
		log.Printf("Fallback consensus client is not configured or unavailable, waiting for primary consensus client to finish syncing (%.2f%%)\n", mgrStatus.PrimaryClientStatus.SyncProgress*100)
		return false, nil
	}

	// Is the fallback working and syncing? If so, wait for it
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackClientStatus.IsWorking && mgrStatus.FallbackClientStatus.Error == "" {
		log.Printf("Primary cosnensus client is unavailable (%s), waiting for the fallback consensus client to finish syncing (%.2f%%)\n", mgrStatus.PrimaryClientStatus.Error, mgrStatus.FallbackClientStatus.SyncProgress*100)
		return false, nil
	}

	// If neither client is working, report the errors
	if mgrStatus.FallbackEnabled {
		return false, fmt.Errorf("Primary consensus client is unavailable (%s) and fallback consensus client is unavailable (%s), no consensus clients are ready.", mgrStatus.PrimaryClientStatus.Error, mgrStatus.FallbackClientStatus.Error)
	}

	return false, fmt.Errorf("Primary consensus client is unavailable (%s) and no fallback consensus client is configured.", mgrStatus.PrimaryClientStatus.Error)
}

// Wait for the primary or fallback Execution client to be synced
func (sp *ServiceProvider) waitEthClientSynced(verbose bool, timeout int64) (bool, error) {
	// Prevent multiple waiting goroutines from requesting sync progress
	ethClientSyncLock.Lock()
	defer ethClientSyncLock.Unlock()

	synced, clientToCheck, err := sp.checkExecutionClientStatus()
	if err != nil {
		return false, err
	}
	if synced {
		return true, nil
	}

	// Get wait start time
	startTime := time.Now()

	// Get EC status refresh time
	ecRefreshTime := startTime

	// Wait for sync
	for {
		// Check timeout
		if (timeout > 0) && (time.Since(startTime).Seconds() > float64(timeout)) {
			return false, nil
		}

		// Check if the EC status needs to be refreshed
		if time.Since(ecRefreshTime) > ethClientStatusRefreshInterval {
			log.Println("Refreshing primary / fallback execution client status...")
			ecRefreshTime = time.Now()
			synced, clientToCheck, err = sp.checkExecutionClientStatus()
			if err != nil {
				return false, err
			}
			if synced {
				return true, nil
			}
		}

		// Get sync progress
		progress, err := clientToCheck.SyncProgress(context.Background())
		if err != nil {
			return false, err
		}

		// Check sync progress
		if progress != nil {
			if verbose {
				p := float64(progress.CurrentBlock-progress.StartingBlock) / float64(progress.HighestBlock-progress.StartingBlock)
				if p > 1 {
					log.Println("Eth 1.0 node syncing...")
				} else {
					log.Printf("Eth 1.0 node syncing: %.2f%%\n", p*100)
				}
			}
		} else {
			// Eth 1 client is not in "syncing" state but may be behind head
			// Get the latest block it knows about and make sure it's recent compared to system clock time
			isUpToDate, _, err := IsSyncWithinThreshold(clientToCheck)
			if err != nil {
				return false, err
			}
			// Only return true if the last reportedly known block is within our defined threshold
			if isUpToDate {
				return true, nil
			}
		}

		// Pause before next poll
		time.Sleep(ethClientSyncPollInterval)
	}
}

// Wait for the primary or fallback Beacon client to be synced
func (sp *ServiceProvider) waitBeaconClientSynced(verbose bool, timeout int64) (bool, error) {
	// Prevent multiple waiting goroutines from requesting sync progress
	beaconClientSyncLock.Lock()
	defer beaconClientSyncLock.Unlock()

	synced, err := sp.checkBeaconClientStatus()
	if err != nil {
		return false, err
	}
	if synced {
		return true, nil
	}

	// Get wait start time
	startTime := time.Now()

	// Get BC status refresh time
	bcRefreshTime := startTime

	// Wait for sync
	for {
		// Check timeout
		if (timeout > 0) && (time.Since(startTime).Seconds() > float64(timeout)) {
			return false, nil
		}

		// Check if the BC status needs to be refreshed
		if time.Since(bcRefreshTime) > ethClientStatusRefreshInterval {
			log.Println("Refreshing primary / fallback consensus client status...")
			bcRefreshTime = time.Now()
			synced, err = sp.checkBeaconClientStatus()
			if err != nil {
				return false, err
			}
			if synced {
				return true, nil
			}
		}

		// Get sync status
		syncStatus, err := sp.bcManager.GetSyncStatus()
		if err != nil {
			return false, err
		}

		// Check sync status
		if syncStatus.Syncing {
			if verbose {
				log.Println("Eth 2.0 node syncing: %.2f%%\n", syncStatus.Progress*100)
			}
		} else {
			return true, nil
		}

		// Pause before next poll
		time.Sleep(beaconClientSyncPollInterval)
	}
}
