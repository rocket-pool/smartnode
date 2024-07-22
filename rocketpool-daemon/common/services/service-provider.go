package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/services"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// ==================
// === Interfaces ===
// ==================

// Provides access to the Rocket Pool binding and a contract address refreshing method
type IRocketPoolProvider interface {
	// Gets the Rocket Pool binding
	GetRocketPool() *rocketpool.RocketPool

	// Refreshes the Rocket Pool contracts if they've been updated on-chain since they were last loaded by the daemon
	RefreshRocketPoolContracts() error
}

// Provides the Smart Node's configuration
type ISmartNodeConfigProvider interface {
	// Gets the Smart Node's configuration
	GetConfig() *config.SmartNodeConfig

	// Gets the Smart Node's list of resources
	GetResources() *config.MergedResources
}

// Provides the Smart Node's validator manager
type IValidatorManagerProvider interface {
	// Gets the Smart Node's validator manager
	GetValidatorManager() *validator.ValidatorManager
}

// Provides a binding for Snapshot's delegation contract
type ISnapshotDelegationProvider interface {
	// Gets the Snapshot delegation binding
	GetSnapshotDelegation() *contracts.SnapshotDelegation
}

// Provides access to the Smart Node's loggers
type ILoggerProvider interface {
	services.ILoggerProvider

	// Gets the Smart Node's watchtower logger
	GetWatchtowerLogger() *log.Logger
}

// Provides methods for requiring or waiting for various conditions to be met
type IRequirementsProvider interface {
	// Require that the Rocket Pool contracts are loaded
	RequireRocketPoolContracts(ctx context.Context) (types.ResponseStatus, error)

	// Require that the Ethereum client is synced
	RequireEthClientSynced(ctx context.Context) error

	// Require that the Beacon chain client is synced
	RequireBeaconClientSynced(ctx context.Context) error

	// Require the Smart Node has a node address set
	RequireNodeAddress() error

	// Require the Smart Node has a wallet that's loaded and ready for transactions
	RequireWalletReady() error

	// Require the node has been registered with the Rocket Pool contracts
	RequireNodeRegistered(ctx context.Context) (types.ResponseStatus, error)

	// Require the selected network has a binding for the Snapshot delegation contract
	RequireSnapshot() error

	// Require the node is a member of the Oracle DAO
	RequireOnOracleDao(ctx context.Context) (types.ResponseStatus, error)

	// Require the node is a member of the Security Council
	RequireOnSecurityCouncil(ctx context.Context) (types.ResponseStatus, error)

	// Wait for the Ethereum client to be synced
	WaitEthClientSynced(ctx context.Context, verbose bool) error

	// Wait for the Beacon chain client to be synced
	WaitBeaconClientSynced(ctx context.Context, verbose bool) error

	// Wait for the node to have a node address set
	WaitNodeAddress(ctx context.Context, verbose bool) error

	// Wait for the node to have a wallet loaded and ready for transactions
	WaitWalletReady(ctx context.Context, verbose bool) error

	// Wait for the node to be registered with the Rocket Pool contracts
	WaitNodeRegistered(ctx context.Context, verbose bool) error
}

// Provides access to all the services used by the Smart Node
type ISmartNodeServiceProvider interface {
	IRocketPoolProvider
	ISmartNodeConfigProvider
	IValidatorManagerProvider
	ISnapshotDelegationProvider
	ILoggerProvider
	IRequirementsProvider

	// Forwarded from the base provider
	services.IEthClientProvider
	services.IBeaconClientProvider
	services.IDockerProvider
	services.IWalletProvider
	services.IContextProvider
	io.Closer
}

// =======================
// === ServiceProvider ===
// =======================

// A container for all of the various services used by the Smartnode
type SmartNodeServiceProvider struct {
	services.IServiceProvider

	// Services
	cfg                *config.SmartNodeConfig
	resources          *config.MergedResources
	rocketPool         *rocketpool.RocketPool
	validatorManager   *validator.ValidatorManager
	snapshotDelegation *contracts.SnapshotDelegation
	watchtowerLog      *log.Logger

	// Internal use
	loadedContractVersion *version.Version
	refreshLock           *sync.Mutex
	userDir               string
}

// Creates a new ServiceProvider instance
func NewServiceProvider(userDir string, resourcesDir string) (ISmartNodeServiceProvider, error) {
	// Load the network settings
	settingsList, err := config.LoadSettingsFiles(resourcesDir)
	if err != nil {
		return nil, fmt.Errorf("error loading network settings: %w", err)
	}

	// Config
	cfgPath := filepath.Join(userDir, config.ConfigFilename)
	cfg, err := client.LoadConfigFromFile(os.ExpandEnv(cfgPath), settingsList)
	if err != nil {
		return nil, fmt.Errorf("error loading Smart Node config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("smart node config settings file [%s] not found", cfgPath)
	}

	// Get the resources from the selected network
	var selectedResources *config.MergedResources
	for _, network := range settingsList {
		if network.Key == cfg.Network.Value {
			selectedResources = &config.MergedResources{
				NetworkResources:   network.NetworkResources,
				SmartNodeResources: network.SmartNodeResources,
			}
			break
		}
	}
	if selectedResources == nil {
		return nil, fmt.Errorf("no resources found for selected network [%s]", cfg.Network.Value)
	}

	// Make the core provider
	sp, err := services.NewServiceProvider(cfg, selectedResources.NetworkResources, config.ClientTimeout)
	if err != nil {
		return nil, fmt.Errorf("error creating core service provider: %w", err)
	}

	// Attempt a wallet upgrade before anything
	tasksLogger := sp.GetTasksLogger().Logger
	upgraded, err := validator.CheckAndUpgradeWallet(cfg.GetWalletFilePath(), cfg.GetNextAccountFilePath(), tasksLogger)
	if err != nil {
		return nil, fmt.Errorf("error checking for legacy wallet upgrade: %w", err)
	}
	if upgraded {
		wallet := sp.GetWallet()
		err = wallet.Reload(tasksLogger)
		if err != nil {
			return nil, fmt.Errorf("error reloading wallet after upgrade: %w", err)
		}
		err = wallet.RestoreAddressToWallet()
		if err != nil {
			return nil, fmt.Errorf("error restoring node address to wallet address after upgrade: %w", err)
		}
	}

	return CreateServiceProviderFromComponents(cfg, selectedResources, sp)
}

// Creates a ServiceProvider instance from a core service provider and Smart Node config
func CreateServiceProviderFromComponents(cfg *config.SmartNodeConfig, resources *config.MergedResources, sp services.IServiceProvider) (ISmartNodeServiceProvider, error) {
	// Make the watchtower log
	loggerOpts := cfg.GetLoggerOptions()
	watchtowerLogger, err := log.NewLogger(cfg.GetWatchtowerLogFilePath(), loggerOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating watchtower logger: %w", err)
	}

	// Rocket Pool
	ecManager := sp.GetEthClient()
	rp, err := rocketpool.NewRocketPool(
		ecManager,
		resources.StorageAddress,
		resources.MulticallAddress,
		resources.BalanceBatcherAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Rocket Pool binding: %w", err)
	}

	// Validator Manager
	vMgr, err := validator.NewValidatorManager(cfg, rp, sp.GetWallet(), sp.GetQueryManager())
	if err != nil {
		return nil, fmt.Errorf("error creating validator manager: %w", err)
	}
	// Snapshot delegation
	var snapshotDelegation *contracts.SnapshotDelegation
	snapshotAddress := resources.SnapshotDelegationAddress
	if snapshotAddress != nil {
		snapshotDelegation, err = contracts.NewSnapshotDelegation(*snapshotAddress, sp.GetEthClient(), sp.GetTransactionManager())
		if err != nil {
			return nil, fmt.Errorf("error creating snapshot delegation binding: %w", err)
		}
	}

	// Create the provider
	defaultVersion, _ := version.NewSemver("0.0.0")
	provider := &SmartNodeServiceProvider{
		userDir:               cfg.RocketPoolDirectory(),
		IServiceProvider:      sp,
		cfg:                   cfg,
		resources:             resources,
		rocketPool:            rp,
		validatorManager:      vMgr,
		snapshotDelegation:    snapshotDelegation,
		watchtowerLog:         watchtowerLogger,
		loadedContractVersion: defaultVersion,
		refreshLock:           &sync.Mutex{},
	}
	return provider, nil
}

// ===============
// === Getters ===
// ===============

func (p *SmartNodeServiceProvider) GetConfig() *config.SmartNodeConfig {
	return p.cfg
}

func (p *SmartNodeServiceProvider) GetResources() *config.MergedResources {
	return p.resources
}

func (p *SmartNodeServiceProvider) GetRocketPool() *rocketpool.RocketPool {
	return p.rocketPool
}

func (p *SmartNodeServiceProvider) GetValidatorManager() *validator.ValidatorManager {
	return p.validatorManager
}

func (p *SmartNodeServiceProvider) GetSnapshotDelegation() *contracts.SnapshotDelegation {
	return p.snapshotDelegation
}

func (p *SmartNodeServiceProvider) GetWatchtowerLogger() *log.Logger {
	return p.watchtowerLog
}

func (p *SmartNodeServiceProvider) Close() error {
	p.watchtowerLog.Close()
	return p.IServiceProvider.Close()
}

// =============
// === Utils ===
// =============

// Refresh the Rocket Pool contracts if they've been updated since they were last loaded
func (p *SmartNodeServiceProvider) RefreshRocketPoolContracts() error {
	p.refreshLock.Lock()
	defer p.refreshLock.Unlock()

	// Get the version on-chain
	protocolVersion, err := p.rocketPool.GetProtocolVersion(nil)
	if err != nil {
		return err
	}

	// Reload everything if it's different from what we have
	if !p.loadedContractVersion.Equal(protocolVersion) {
		err := p.rocketPool.LoadAllContracts(nil)
		if err != nil {
			return fmt.Errorf("error updating contracts to [%s]: %w", protocolVersion.String(), err)
		}
		p.loadedContractVersion = protocolVersion
	}
	return nil
}
