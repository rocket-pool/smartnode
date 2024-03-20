package services

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/fatih/color"
	"github.com/rocket-pool/node-manager-core/node/services"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/shared/config"
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

	// Internal use
	contractLoadBlock uint64
	userDir           string
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

	// Attempt a wallet upgrade before anything
	upgradeLog := log.NewColorLogger(color.FgHiWhite)
	err = validator.CheckAndUpgradeWallet(cfg.GetWalletFilePath(), cfg.GetNextAccountFilePath(), &upgradeLog)
	if err != nil {
		return nil, fmt.Errorf("error checking for legacy wallet upgrade: %w", err)
	}

	// Make the core provider
	sp, err := services.NewServiceProvider(cfg, config.ClientTimeout, cfg.DebugMode.Value)
	if err != nil {
		return nil, fmt.Errorf("error creating core service provider: %w", err)
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
	provider := &ServiceProvider{
		userDir:            userDir,
		ServiceProvider:    sp,
		cfg:                cfg,
		rocketPool:         rp,
		validatorManager:   vMgr,
		rplFaucet:          rplFaucet,
		snapshotDelegation: snapshotDelegation,
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

// =============
// === Utils ===
// =============

func (p *ServiceProvider) LoadContractsIfStale() error {
	if p.contractLoadBlock > 0 {
		return nil
	}

	// Get the current block
	var err error
	p.contractLoadBlock, err = p.GetEthClient().BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting latest block: %w", err)
	}

	return p.rocketPool.LoadAllContracts(&bind.CallOpts{
		BlockNumber: big.NewInt(int64(p.contractLoadBlock)),
	})
}
