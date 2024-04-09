package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/services"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// A container for all of the various services used by the Smartnode
type ServiceProvider struct {
	*services.ServiceProvider

	// Services
	cfg                *config.SmartNodeConfig
	rocketPool         *rocketpool.RocketPool
	validatorManager   *validator.ValidatorManager
	rplFaucet          *contracts.RplFaucet
	snapshotDelegation *contracts.SnapshotDelegation
	watchtowerLog      *log.Logger

	// Internal use
	loadedContractVersion *version.Version
	refreshLock           *sync.Mutex
	userDir               string
}

// Creates a new ServiceProvider instance
func NewServiceProvider(userDir string) (*ServiceProvider, error) {
	// Config
	cfgPath := filepath.Join(userDir, config.ConfigFilename)
	cfg, err := client.LoadConfigFromFile(os.ExpandEnv(cfgPath))
	if err != nil {
		return nil, fmt.Errorf("error loading Smartnode config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("smart node config settings file [%s] not found", cfgPath)
	}

	// Make the core provider
	sp, err := services.NewServiceProvider(cfg, config.ClientTimeout)
	if err != nil {
		return nil, fmt.Errorf("error creating core service provider: %w", err)
	}

	// Make the watchtower log
	loggerOpts := cfg.GetLoggerOptions()
	watchtowerLogger, err := log.NewLogger(cfg.GetWatchtowerLogFilePath(), loggerOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating watchtower logger: %w", err)
	}

	// Attempt a wallet upgrade before anything
	tasksLogger := sp.GetTasksLogger().Logger
	upgraded, err := validator.CheckAndUpgradeWallet(cfg.GetWalletFilePath(), cfg.GetNextAccountFilePath(), tasksLogger)
	if err != nil {
		return nil, fmt.Errorf("error checking for legacy wallet upgrade: %w", err)
	}
	if upgraded {
		err = sp.GetWallet().Reload(tasksLogger)
		if err != nil {
			return nil, fmt.Errorf("error reloading wallet after upgrade: %w", err)
		}
	}

	// Rocket Pool
	ecManager := sp.GetEthClient()
	resources := cfg.GetRocketPoolResources()
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

	// RPL Faucet
	var rplFaucet *contracts.RplFaucet
	faucetAddress := resources.RplFaucetAddress
	if faucetAddress != nil {
		rplFaucet, err = contracts.NewRplFaucet(*faucetAddress, sp.GetEthClient(), sp.GetTransactionManager())
		if err != nil {
			return nil, fmt.Errorf("error creating RPL faucet binding: %w", err)
		}
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
	provider := &ServiceProvider{
		userDir:               userDir,
		ServiceProvider:       sp,
		cfg:                   cfg,
		rocketPool:            rp,
		validatorManager:      vMgr,
		rplFaucet:             rplFaucet,
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

func (p *ServiceProvider) GetUserDir() string {
	return p.userDir
}

func (p *ServiceProvider) GetConfig() *config.SmartNodeConfig {
	return p.cfg
}

func (p *ServiceProvider) GetRocketPool() *rocketpool.RocketPool {
	return p.rocketPool
}

func (p *ServiceProvider) GetValidatorManager() *validator.ValidatorManager {
	return p.validatorManager
}

func (p *ServiceProvider) GetRplFaucet() *contracts.RplFaucet {
	return p.rplFaucet
}

func (p *ServiceProvider) GetSnapshotDelegation() *contracts.SnapshotDelegation {
	return p.snapshotDelegation
}

func (p *ServiceProvider) GetWatchtowerLogger() *log.Logger {
	return p.watchtowerLog
}

func (p *ServiceProvider) Close() {
	p.watchtowerLog.Close()
	p.ServiceProvider.Close()
}

// =============
// === Utils ===
// =============

// Refresh the Rocket Pool contracts if they've been updated since they were last loaded
func (p *ServiceProvider) RefreshRocketPoolContracts() error {
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
