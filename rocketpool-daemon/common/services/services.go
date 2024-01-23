package services

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet"
	lhkeystore "github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore/lighthouse"
	lokeystore "github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore/lodestar"
	nmkeystore "github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore/nimbus"
	prkeystore "github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore/prysm"
	tkkeystore "github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore/teku"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/docker"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// A container for all of the various services used by the Smartnode
type ServiceProvider struct {
	cfg                *config.RocketPoolConfig
	nodeWallet         *wallet.LocalWallet
	ecManager          *ExecutionClientManager
	bcManager          *BeaconClientManager
	rocketPool         *rocketpool.RocketPool
	rplFaucet          *contracts.RplFaucet
	snapshotDelegation *contracts.SnapshotDelegation
	docker             *client.Client
}

// Creates a new ServiceProvider instance
func NewServiceProvider(c *cli.Context) (*ServiceProvider, error) {
	// Config
	settingsFile := os.ExpandEnv(c.GlobalString("settings"))
	cfg, err := rp.LoadConfigFromFile(settingsFile)
	if err != nil {
		return nil, fmt.Errorf("error loading Smartnode config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("Smartnode config settings file [%s] not found", settingsFile)
	}

	// Wallet
	chainID := cfg.Smartnode.GetChainID()
	nodeAddressPath := os.ExpandEnv(cfg.Smartnode.GetNodeAddressPath())
	keystorePath := os.ExpandEnv(cfg.Smartnode.GetWalletPath())
	passwordPath := os.ExpandEnv(cfg.Smartnode.GetPasswordPath())
	nodeWallet, err := wallet.NewLocalWallet(nodeAddressPath, keystorePath, passwordPath, chainID, true)
	if err != nil {
		return nil, fmt.Errorf("error creating node wallet: %w", err)
	}

	// Keystores
	validatorKeychainPath := os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath())
	lighthouseKeystore := lhkeystore.NewKeystore(validatorKeychainPath)
	lodestarKeystore := lokeystore.NewKeystore(validatorKeychainPath)
	nimbusKeystore := nmkeystore.NewKeystore(validatorKeychainPath)
	prysmKeystore := prkeystore.NewKeystore(validatorKeychainPath)
	tekuKeystore := tkkeystore.NewKeystore(validatorKeychainPath)
	nodeWallet.AddValidatorKeystore("lighthouse", lighthouseKeystore)
	nodeWallet.AddValidatorKeystore("lodestar", lodestarKeystore)
	nodeWallet.AddValidatorKeystore("nimbus", nimbusKeystore)
	nodeWallet.AddValidatorKeystore("prysm", prysmKeystore)
	nodeWallet.AddValidatorKeystore("teku", tekuKeystore)

	// EC Manager
	ecManager, err := NewExecutionClientManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating executon client manager: %w", err)
	}

	// Rocket Pool
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

	// Beacon manager
	bcManager, err := NewBeaconClientManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Beacon client manager: %w", err)
	}

	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.WithVersion(docker.DockerApiVersion))
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}

	// Check if the managers should ignore sync checks and/or default to using the fallback (used by the API container when driven by the CLI)
	if c.GlobalBool("ignore-sync-check") {
		ecManager.ignoreSyncCheck = true
		bcManager.ignoreSyncCheck = true
	}
	if c.GlobalBool("force-fallbacks") {
		ecManager.primaryReady = false
		bcManager.primaryReady = false
	}

	// Create the provider
	provider := &ServiceProvider{
		cfg:                cfg,
		nodeWallet:         nodeWallet,
		ecManager:          ecManager,
		bcManager:          bcManager,
		rocketPool:         rp,
		rplFaucet:          rplFaucet,
		snapshotDelegation: snapshotDelegation,
		docker:             dockerClient,
	}
	return provider, nil
}

// ===============
// === Getters ===
// ===============

func (p *ServiceProvider) GetConfig() *config.RocketPoolConfig {
	return p.cfg
}

func (p *ServiceProvider) GetWallet() *wallet.LocalWallet {
	return p.nodeWallet
}

func (p *ServiceProvider) GetEthClient() *ExecutionClientManager {
	return p.ecManager
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

func (p *ServiceProvider) GetBeaconClient() *BeaconClientManager {
	return p.bcManager
}

func (p *ServiceProvider) GetDocker() *client.Client {
	return p.docker
}
