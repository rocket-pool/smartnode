package services

import (
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	uc "github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/beacon/lighthouse"
	"github.com/rocket-pool/smartnode/shared/services/beacon/nimbus"
	"github.com/rocket-pool/smartnode/shared/services/beacon/prysm"
	"github.com/rocket-pool/smartnode/shared/services/beacon/teku"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	lhkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
	nmkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
	prkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
	tkkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Config
const (
	DockerAPIVersion        string = "1.40"
	EcContainerName         string = "eth1"
	FallbackEcContainerName string = "eth1-fallback"
	BnContainerName         string = "eth2"
)

// Service instances & initializers
var (
	cfg             *config.RocketPoolConfig
	passwordManager *passwords.PasswordManager
	nodeWallet      *wallet.Wallet
	ethClientProxy  *uc.EthClientProxy
	rocketPool      *rocketpool.RocketPool
	oneInchOracle   *contracts.OneInchOracle
	rplFaucet       *contracts.RPLFaucet
	beaconClient    beacon.Client
	docker          *client.Client

	initCfg             sync.Once
	initPasswordManager sync.Once
	initNodeWallet      sync.Once
	initEthClientProxy  sync.Once
	initRocketPool      sync.Once
	initOneInchOracle   sync.Once
	initRplFaucet       sync.Once
	initBeaconClient    sync.Once
	initDocker          sync.Once
)

//
// Service providers
//

func GetConfig(c *cli.Context) (*config.RocketPoolConfig, error) {
	return getConfig(c)
}

func GetPasswordManager(c *cli.Context) (*passwords.PasswordManager, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	return getPasswordManager(cfg), nil
}

func GetWallet(c *cli.Context) (*wallet.Wallet, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	pm := getPasswordManager(cfg)
	return getWallet(c, cfg, pm)
}

func GetEthClientProxy(c *cli.Context) (*uc.EthClientProxy, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	ec, err := getEthClientProxy(cfg)
	if err != nil {
		return nil, err
	}
	return ec, nil
}

func GetRocketPool(c *cli.Context) (*rocketpool.RocketPool, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	ec, err := getEthClientProxy(cfg)
	if err != nil {
		return nil, err
	}

	return getRocketPool(cfg, ec)
}

func GetOneInchOracle(c *cli.Context) (*contracts.OneInchOracle, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	ec, err := getEthClientProxy(cfg)
	if err != nil {
		return nil, err
	}
	return getOneInchOracle(cfg, ec)
}

func GetRplFaucet(c *cli.Context) (*contracts.RPLFaucet, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	ec, err := getEthClientProxy(cfg)
	if err != nil {
		return nil, err
	}
	return getRplFaucet(cfg, ec)
}

func GetBeaconClient(c *cli.Context) (beacon.Client, error) {
	cfg, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	return getBeaconClient(cfg)
}

func GetDocker(c *cli.Context) (*client.Client, error) {
	return getDocker()
}

//
// Service instance getters
//

func getConfig(c *cli.Context) (*config.RocketPoolConfig, error) {
	var err error
	initCfg.Do(func() {
		settingsFile := os.ExpandEnv(c.GlobalString("settings"))
		cfg, err = rp.LoadConfigFromFile(settingsFile)
		if cfg == nil && err == nil {
			err = fmt.Errorf("Settings file [%s] not found.", settingsFile)
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

func getWallet(c *cli.Context, cfg *config.RocketPoolConfig, pm *passwords.PasswordManager) (*wallet.Wallet, error) {
	var err error
	initNodeWallet.Do(func() {
		var maxFee *big.Int
		maxFeeFloat := c.GlobalFloat64("maxFee")
		if maxFeeFloat == 0 {
			maxFeeFloat = cfg.Smartnode.ManualMaxFee.Value.(float64)
		}
		if maxFeeFloat != 0 {
			maxFee = eth.GweiToWei(maxFeeFloat)
		}

		var maxPriorityFee *big.Int
		maxPriorityFeeFloat := c.GlobalFloat64("maxPrioFee")
		if maxPriorityFeeFloat == 0 {
			maxPriorityFeeFloat = cfg.Smartnode.PriorityFee.Value.(float64)
		}
		if maxPriorityFeeFloat != 0 {
			maxPriorityFee = eth.GweiToWei(maxPriorityFeeFloat)
		}

		chainId := cfg.Smartnode.GetChainID()

		nodeWallet, err = wallet.NewWallet(os.ExpandEnv(cfg.Smartnode.GetWalletPath()), chainId, maxFee, maxPriorityFee, 0, pm)
		if err != nil {
			return
		}
		lighthouseKeystore := lhkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		nimbusKeystore := nmkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		prysmKeystore := prkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		tekuKeystore := tkkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.GetValidatorKeychainPath()), pm)
		nodeWallet.AddKeystore("lighthouse", lighthouseKeystore)
		nodeWallet.AddKeystore("nimbus", nimbusKeystore)
		nodeWallet.AddKeystore("prysm", prysmKeystore)
		nodeWallet.AddKeystore("teku", tekuKeystore)
	})
	return nodeWallet, err
}

func getEthClientProxy(cfg *config.RocketPoolConfig) (*uc.EthClientProxy, error) {
	var err error
	initEthClientProxy.Do(func() {
		reconnectDelay, err := time.ParseDuration(cfg.ReconnectDelay.Value.(string))
		if err != nil {
			return
		}

		// Get the provider URL of the primary execution client
		var primaryProvider string
		if cfg.IsNativeMode {
			primaryProvider = cfg.Native.EcHttpUrl.Value.(string)
		} else if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
			primaryProvider = fmt.Sprintf("http://%s:%d", EcContainerName, cfg.ExecutionCommon.HttpPort.Value.(uint16))
		} else {
			primaryProvider = cfg.ExternalExecution.HttpUrl.Value.(string)
		}

		if cfg.UseFallbackExecutionClient.Value == false || cfg.IsNativeMode {
			ethClientProxy = uc.NewEth1ClientProxy(reconnectDelay, primaryProvider)
		} else {
			// Get the provider URL of the fallback execution client
			var fallbackProvider string
			if cfg.FallbackExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
				fallbackProvider = fmt.Sprintf("http://%s:%d", FallbackEcContainerName, cfg.FallbackExecutionCommon.HttpPort.Value.(uint16))
			} else {
				fallbackProvider = cfg.FallbackExternalExecution.HttpUrl.Value.(string)
			}

			ethClientProxy = uc.NewEth1ClientProxy(reconnectDelay, primaryProvider, fallbackProvider)
		}
	})
	return ethClientProxy, err
}

func getRocketPool(cfg *config.RocketPoolConfig, client *uc.EthClientProxy) (*rocketpool.RocketPool, error) {
	var err error
	initRocketPool.Do(func() {
		rocketPool, err = rocketpool.NewRocketPool(client, common.HexToAddress(cfg.Smartnode.GetStorageAddress()))
	})
	return rocketPool, err
}

func getOneInchOracle(cfg *config.RocketPoolConfig, client *uc.EthClientProxy) (*contracts.OneInchOracle, error) {
	var err error
	initOneInchOracle.Do(func() {
		oneInchOracle, err = contracts.NewOneInchOracle(common.HexToAddress(cfg.Smartnode.GetOneInchOracleAddress()), client)
	})
	return oneInchOracle, err
}

func getRplFaucet(cfg *config.RocketPoolConfig, client *uc.EthClientProxy) (*contracts.RPLFaucet, error) {
	var err error
	initRplFaucet.Do(func() {
		rplFaucet, err = contracts.NewRPLFaucet(common.HexToAddress(cfg.Smartnode.GetRplFaucetAddress()), client)
	})
	return rplFaucet, err
}

func getBeaconClient(cfg *config.RocketPoolConfig) (beacon.Client, error) {
	var err error
	initBeaconClient.Do(func() {
		var provider string
		var selectedCC config.ConsensusClient
		if cfg.IsNativeMode {
			provider = cfg.Native.CcHttpUrl.Value.(string)
			selectedCC = cfg.Native.ConsensusClient.Value.(config.ConsensusClient)
		} else if cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local {
			provider = fmt.Sprintf("http://%s:%d", BnContainerName, cfg.ConsensusCommon.ApiPort.Value.(uint16))
			selectedCC = cfg.ConsensusClient.Value.(config.ConsensusClient)
		} else if cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_External {
			var selectedConsensusConfig config.ConsensusConfig
			selectedConsensusConfig, err = cfg.GetSelectedConsensusClientConfig()
			if err != nil {
				return
			}
			provider = selectedConsensusConfig.(config.ExternalConsensusConfig).GetApiUrl()
			selectedCC = cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
		} else {
			err = fmt.Errorf("Unknown Consensus client mode '%v'", cfg.ConsensusClientMode.Value)
		}

		switch selectedCC {
		case config.ConsensusClient_Lighthouse:
			beaconClient = lighthouse.NewClient(provider)
		case config.ConsensusClient_Nimbus:
			beaconClient = nimbus.NewClient(provider)
		case config.ConsensusClient_Prysm:
			beaconClient = prysm.NewClient(provider)
		case config.ConsensusClient_Teku:
			beaconClient = teku.NewClient(provider)
		default:
			err = fmt.Errorf("Unknown Consensus client '%v' selected", cfg.ConsensusClient.Value)
		}

	})
	return beaconClient, err
}

func getDocker() (*client.Client, error) {
	var err error
	initDocker.Do(func() {
		docker, err = client.NewClientWithOpts(client.WithVersion(DockerAPIVersion))
	})
	return docker, err
}
