package services

import (
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpSettings "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	lhkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
	lokeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lodestar"
	nmkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
	prkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
	tkkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Config
const (
	dockerAPIVersion string = "1.44"
)

// Service instances & initializers
var (
	cfg                  *config.RocketPoolConfig
	passwordManager      *passwords.PasswordManager
	addressManager       *wallet.AddressManager
	nodeWallet           wallet.Wallet
	ecManager            *ExecutionClientManager
	bcManager            *BeaconClientManager
	rocketPool           *rocketpool.RocketPool
	rocketSignerRegistry *contracts.RocketSignerRegistry
	beaconClient         beacon.Client
	docker               *client.Client

	initCfg                  sync.Once
	initPasswordManager      sync.Once
	initAddressManager       sync.Once
	initNodeWallet           sync.Once
	initECManager            sync.Once
	initBCManager            sync.Once
	initRocketPool           sync.Once
	initOneInchOracle        sync.Once
	initRocketSignerRegistry sync.Once
	initBeaconClient         sync.Once
	initDocker               sync.Once
)

//
// Service providers
//

func GetConfig(c *cli.Command) (*config.RocketPoolConfig, error) {
	return getConfig(c)
}

func GetPasswordManager(c *cli.Command) (*passwords.PasswordManager, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	return getPasswordManager(cfg), nil
}

func GetWallet(c *cli.Command) (wallet.Wallet, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	pm := getPasswordManager(cfg)
	am := getAddressManager(cfg)
	return getWallet(c, cfg, pm, am, false)
}

// GetNodeAccountTransactorFromRequest returns a transactor for the node
// account, with gas settings overridden by values sent in the HTTP request
// body (maxFee, maxPrioFee, gasLimit fields in Gwei / uint64).  This ensures
// the fee cap and tip cap the user selected interactively in the CLI are
// honoured by the daemon, which otherwise only knows about the values in the
// config file.
func GetNodeAccountTransactorFromRequest(c *cli.Command, r *http.Request) (*bind.TransactOpts, error) {
	w, err := GetWallet(c)
	if err != nil {
		return nil, err
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	if maxFeeStr := r.FormValue("maxFee"); maxFeeStr != "" {
		if maxFeeGwei, parseErr := strconv.ParseFloat(maxFeeStr, 64); parseErr == nil && maxFeeGwei > 0 {
			opts.GasFeeCap = eth.GweiToWei(maxFeeGwei)
		}
	}
	if maxPrioFeeStr := r.FormValue("maxPrioFee"); maxPrioFeeStr != "" {
		if maxPrioFeeGwei, parseErr := strconv.ParseFloat(maxPrioFeeStr, 64); parseErr == nil && maxPrioFeeGwei > 0 {
			opts.GasTipCap = eth.GweiToWei(maxPrioFeeGwei)
		}
	}
	if gasLimitStr := r.FormValue("gasLimit"); gasLimitStr != "" {
		if gasLimit, parseErr := strconv.ParseUint(gasLimitStr, 10, 64); parseErr == nil {
			opts.GasLimit = gasLimit
		}
	}
	if nonceStr := r.FormValue("nonce"); nonceStr != "" {
		if nonce, ok := new(big.Int).SetString(nonceStr, 0); ok {
			opts.Nonce = nonce
		}
	}
	return opts, nil
}

func GetHdWallet(c *cli.Command) (wallet.Wallet, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	pm := getPasswordManager(cfg)
	am := getAddressManager(cfg)
	return getWallet(c, cfg, pm, am, true)
}

func GetEthClient(c *cli.Command) (*ExecutionClientManager, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	ec, err := getEthClient(c, cfg)
	if err != nil {
		return nil, err
	}
	return ec, nil
}

func dialProtectedEthClient(url string) (*ethClient, error) {
	ec, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &ethClient{ec}, nil
}

func GetRocketPool(c *cli.Command) (*rocketpool.RocketPool, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	var ec rocketpool.ExecutionClient
	if c.Root().Bool("use-protected-api") {
		url := cfg.Smartnode.GetFlashbotsProtectUrl()
		ec, err = dialProtectedEthClient(url)
	} else {
		ec, err = getEthClient(c, cfg)
	}
	if err != nil {
		return nil, err
	}

	return getRocketPool(cfg, ec)
}

func GetRocketSignerRegistry(c *cli.Command) (*contracts.RocketSignerRegistry, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	ec, err := getEthClient(c, cfg)
	if err != nil {
		return nil, err
	}
	return getRocketSignerRegistry(cfg, ec)
}

func GetBeaconClient(c *cli.Command) (*BeaconClientManager, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	return getBeaconClient(c, cfg)
}

func GetDocker(c *cli.Command) (*client.Client, error) {
	var err error
	initDocker.Do(func() {
		docker, err = client.NewClientWithOpts(client.WithVersion(dockerAPIVersion))
	})
	return docker, err
}

func GetBeaconState(bc beacon.Client) (eth2.BeaconState, error) {
	blockToRequest := "finalized"
	var block beacon.BeaconBlock
	var err error
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		block, _, err = bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return nil, err
		}

		if block.HasExecutionPayload {
			break
		}
		if attempts == maxAttempts-1 {
			return nil, fmt.Errorf("failed to find a block with execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", block.Slot-1)
	}

	// Get the beacon state for that slot
	beaconStateResponse, err := bc.GetBeaconStateSSZ(block.Slot)
	if err != nil {
		return nil, err
	}

	beaconState, err := eth2.NewBeaconState(beaconStateResponse.Data, beaconStateResponse.Fork)
	if err != nil {
		return nil, err
	}
	return beaconState, nil
}

//
// Service instance getters
//

func getConfig(c *cli.Command) (*config.RocketPoolConfig, error) {
	var err error
	initCfg.Do(func() {
		settingsFile := c.Root().String("settings")
		if settingsFile == "" {
			configDir := c.Root().String("config-path")
			if configDir != "" {
				settingsFile = filepath.Join(configDir, rpSettings.SettingsFile)
			}
		}
		expanded := os.ExpandEnv(settingsFile)
		cfg, err = rp.LoadConfigFromFile(expanded)
		if cfg == nil && err == nil {
			err = fmt.Errorf("settings file [%s] not found", expanded)
		}
	})
	return cfg, err
}

func getPasswordManager(cfg *config.RocketPoolConfig) *passwords.PasswordManager {
	initPasswordManager.Do(func() {
		passwordManager = passwords.NewPasswordManager(os.ExpandEnv(cfg.Smartnode.GetPasswordPath()))
	})
	return passwordManager
}

func getAddressManager(cfg *config.RocketPoolConfig) *wallet.AddressManager {
	initAddressManager.Do(func() {
		addressManager = wallet.NewAddressManager(os.ExpandEnv(cfg.Smartnode.GetNodeAddressPath()))
	})
	return addressManager
}

func getWallet(c *cli.Command, cfg *config.RocketPoolConfig, pm *passwords.PasswordManager, am *wallet.AddressManager, ignoreMasquerade bool) (wallet.Wallet, error) {
	var err error
	initNodeWallet.Do(func() {
		var maxFee *big.Int
		maxFeeFloat := c.Root().Float64("maxFee")
		if maxFeeFloat == 0 {
			maxFeeFloat = cfg.Smartnode.ManualMaxFee.Value.(float64)
		}
		if maxFeeFloat != 0 {
			maxFee = eth.GweiToWei(maxFeeFloat)
		}

		var maxPriorityFee *big.Int
		maxPriorityFeeFloat := c.Root().Float64("maxPrioFee")
		if maxPriorityFeeFloat == 0 {
			maxPriorityFeeFloat = cfg.Smartnode.PriorityFee.Value.(float64)
		}
		if maxPriorityFeeFloat != 0 {
			maxPriorityFee = eth.GweiToWei(maxPriorityFeeFloat)
		}

		chainId := cfg.Smartnode.GetChainID()

		if ignoreMasquerade {
			nodeWallet, err = wallet.NewHdWallet(os.ExpandEnv(cfg.Smartnode.GetWalletPath()), chainId, maxFee, maxPriorityFee, 0, pm, am)
		} else {
			nodeWallet, err = wallet.NewWallet(os.ExpandEnv(cfg.Smartnode.GetNodeAddressPath()), os.ExpandEnv(cfg.Smartnode.GetWalletPath()), chainId, maxFee, maxPriorityFee, 0, pm, am)
		}
		if err != nil {
			return
		}

		// Keystores
		lighthouseKeystore := lhkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		lodestarKeystore := lokeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		nimbusKeystore := nmkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		prysmKeystore := prkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		tekuKeystore := tkkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		nodeWallet.AddKeystore("lighthouse", lighthouseKeystore)
		nodeWallet.AddKeystore("lodestar", lodestarKeystore)
		nodeWallet.AddKeystore("nimbus", nimbusKeystore)
		nodeWallet.AddKeystore("prysm", prysmKeystore)
		nodeWallet.AddKeystore("teku", tekuKeystore)
	})
	return nodeWallet, err
}

func getEthClient(c *cli.Command, cfg *config.RocketPoolConfig) (*ExecutionClientManager, error) {
	var err error
	initECManager.Do(func() {
		// Create a new client manager
		ecManager, err = NewExecutionClientManager(cfg)
		if err == nil {
			// Check if the manager should ignore sync checks and/or default to using the fallback (used by the API container when driven by the CLI)
			if c.Root().Bool("ignore-sync-check") {
				ecManager.ignoreSyncCheck = true
			}
			if c.Root().Bool("force-fallbacks") {
				ecManager.primaryReady = false
			}
		}
	})
	return ecManager, err
}

func getRocketPool(cfg *config.RocketPoolConfig, client rocketpool.ExecutionClient) (*rocketpool.RocketPool, error) {
	var err error
	initRocketPool.Do(func() {
		rocketPool, err = rocketpool.NewRocketPool(client, common.HexToAddress(cfg.Smartnode.GetStorageAddress()))
	})
	return rocketPool, err
}

func getRocketSignerRegistry(cfg *config.RocketPoolConfig, client rocketpool.ExecutionClient) (*contracts.RocketSignerRegistry, error) {
	var err error
	initRocketSignerRegistry.Do(func() {
		address := cfg.Smartnode.GetRocketSignerRegistryAddress()
		if address != "" {
			rocketSignerRegistry, err = contracts.NewRocketSignerRegistry(common.HexToAddress(address), client)
		}
	})
	return rocketSignerRegistry, err
}

func getBeaconClient(c *cli.Command, cfg *config.RocketPoolConfig) (*BeaconClientManager, error) {
	var err error
	initBCManager.Do(func() {
		// Create a new client manager
		bcManager, err = NewBeaconClientManager(cfg)
		if err == nil {
			// Check if the manager should ignore sync checks and/or default to using the fallback (used by the API container when driven by the CLI)
			if c.Root().Bool("ignore-sync-check") {
				bcManager.ignoreSyncCheck = true
			}
			if c.Root().Bool("force-fallbacks") {
				bcManager.primaryReady = false
			}
		}
	})
	return bcManager, err
}
