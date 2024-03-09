package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/node/services"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// A container for all of the various services used by the Smartnode
type ServiceProvider struct {
	*services.ServiceProvider

	// Services
	cfg                *config.RocketPoolConfig
	rocketPool         *rocketpool.RocketPool
	rplFaucet          *contracts.RplFaucet
	snapshotDelegation *contracts.SnapshotDelegation

	// Internal use
	contractLoadBlock uint64
	userDir           string
}

// Creates a new ServiceProvider instance
func NewServiceProvider(settingsFile string) (*ServiceProvider, error) {
	// Config
	cfg, err := rp.LoadConfigFromFile(settingsFile)
	if err != nil {
		return nil, fmt.Errorf("error loading Smartnode config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("Smartnode config settings file [%s] not found", settingsFile)
	}

	// Core provider
	sp, err := services.NewServiceProvider(cfg, rpconfig.ClientTimeout, rpconfig.DebugMode.Value)
	if err != nil {
		return nil, fmt.Errorf("error creating core service provider: %w", err)
	}

	// Rocket Pool
	ecManager := sp.GetEthClient()
	rp, err := rocketpool.NewRocketPool(
		ecManager,
		common.HexToAddress(cfg.Smartnode.GetStorageAddress()),
		common.HexToAddress(cfg.Smartnode.GetMulticallAddress()),
		common.HexToAddress(cfg.Smartnode.GetBalanceBatcherAddress()),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating Rocket Pool binding: %w", err)
	}

	// RPL Faucet
	var rplFaucet *contracts.RplFaucet
	faucetAddress := cfg.Smartnode.GetRplFaucetAddress()
	if faucetAddress != "" {
		rplFaucet, err = contracts.NewRplFaucet(common.HexToAddress(faucetAddress), ecManager)
		if err != nil {
			return nil, fmt.Errorf("error creating RPL faucet binding: %w", err)
		}
	}

	// Snapshot delegation
	var snapshotDelegation *contracts.SnapshotDelegation
	snapshotAddress := cfg.Smartnode.GetSnapshotDelegationAddress()
	if snapshotAddress != "" {
		snapshotDelegation, err = contracts.NewSnapshotDelegation(common.HexToAddress(snapshotAddress), ecManager)
		if err != nil {
			return nil, fmt.Errorf("error creating snapshot delegation binding: %w", err)
		}
	}

	// Create the provider
	provider := &ServiceProvider{
		ServiceProvider:    sp,
		cfg:                cfg,
		rocketPool:         rp,
		rplFaucet:          rplFaucet,
		snapshotDelegation: snapshotDelegation,
	}
	return provider, nil
}

// ===============
// === Getters ===
// ===============

func (p *ServiceProvider) GetConfig() *config.RocketPoolConfig {
	return p.cfg
}

func (p *ServiceProvider) GetRocketPool() *rocketpool.RocketPool {
	return p.rocketPool
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
