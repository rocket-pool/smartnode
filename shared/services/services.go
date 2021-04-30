package services

import (
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/beacon/lighthouse"
	"github.com/rocket-pool/smartnode/shared/services/beacon/nimbus"
	"github.com/rocket-pool/smartnode/shared/services/beacon/prysm"
	"github.com/rocket-pool/smartnode/shared/services/beacon/teku"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	rpcli "github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	lhkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
	nmkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
	prkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
	tkkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
)

// Config
const DockerAPIVersion = "1.40"


// Service instances & initializers
var (
    cfg config.RocketPoolConfig
    passwordManager *passwords.PasswordManager
    nodeWallet *wallet.Wallet
    ethClient *ethclient.Client
    mainnetEthClient *ethclient.Client
    rocketPool *rocketpool.RocketPool
    oneInchOracle *contracts.OneInchOracle
    beaconClient beacon.Client
    docker *client.Client

    initCfg sync.Once
    initPasswordManager sync.Once
    initNodeWallet sync.Once
    initEthClient sync.Once
    initMainnetEthClient sync.Once
    initRocketPool sync.Once
    initOneInchOracle sync.Once
    initBeaconClient sync.Once
    initDocker sync.Once
)


//
// Service providers
//


func GetConfig(c *cli.Context) (config.RocketPoolConfig, error) {
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
    return getWallet(cfg, pm)
}


func GetEthClient(c *cli.Context) (*ethclient.Client, error) {
    cfg, err := getConfig(c)
    if err != nil {
        return nil, err
    }
    return getEthClient(cfg)
}


func GetMainnetEthClient(c *cli.Context) (*ethclient.Client, error) {
    cfg, err := getConfig(c)
    if err != nil {
        return nil, err
    }
    return getMainnetEthClient(cfg)
}


func GetRocketPool(c *cli.Context) (*rocketpool.RocketPool, error) {
    cfg, err := getConfig(c)
    if err != nil {
        return nil, err
    }
    ec, err := getEthClient(cfg)
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
    mnec, err := getMainnetEthClient(cfg)
    if err != nil {
        return nil, err
    }
    return getOneInchOracle(cfg, mnec)
}


func GetBeaconClient(c *cli.Context) (beacon.Client, error) {
    cfg, err := getConfig(c)
    if err != nil {
        return nil, err
    }
    return getBeaconClient(cfg)
}


func GetBeaconClientFromCLI(rp *rpcli.Client) (beacon.Client, error) {
    cfg, err := rp.LoadGlobalConfig()
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


func getConfig(c *cli.Context) (config.RocketPoolConfig, error) {
    var err error
    initCfg.Do(func() {
        cfg, err = config.Load(c)
    })
    return cfg, err
}


func getPasswordManager(cfg config.RocketPoolConfig) *passwords.PasswordManager {
    initPasswordManager.Do(func() {
        passwordManager = passwords.NewPasswordManager(os.ExpandEnv(cfg.Smartnode.PasswordPath))
    })
    return passwordManager
}


func getWallet(cfg config.RocketPoolConfig, pm *passwords.PasswordManager) (*wallet.Wallet, error) {
    var err error
    initNodeWallet.Do(func() {
        var gasPrice *big.Int
        var gasLimit uint64
        gasPrice, err = cfg.GetGasPrice()
        if err != nil { return }
        gasLimit, err = cfg.GetGasLimit()
        if err != nil { return }
        nodeWallet, err = wallet.NewWallet(os.ExpandEnv(cfg.Smartnode.WalletPath), cfg.Chains.Eth1.ChainID, gasPrice, gasLimit, pm)
        if err != nil { return }
        lighthouseKeystore := lhkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.ValidatorKeychainPath), pm)
        nimbusKeystore := nmkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.ValidatorKeychainPath), pm)
        prysmKeystore := prkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.ValidatorKeychainPath), pm)
        tekuKeystore := tkkeystore.NewKeystore(os.ExpandEnv(cfg.Smartnode.ValidatorKeychainPath), pm)
        nodeWallet.AddKeystore("lighthouse", lighthouseKeystore)
        nodeWallet.AddKeystore("nimbus", nimbusKeystore)
        nodeWallet.AddKeystore("prysm", prysmKeystore)
        nodeWallet.AddKeystore("teku", tekuKeystore)
    })
    return nodeWallet, err
}


func getEthClient(cfg config.RocketPoolConfig) (*ethclient.Client, error) {
    var err error
    initEthClient.Do(func() {
        ethClient, err = ethclient.Dial(cfg.Chains.Eth1.Provider)
    })
    return ethClient, err
}


func getMainnetEthClient(cfg config.RocketPoolConfig) (*ethclient.Client, error) {
    var err error
    initMainnetEthClient.Do(func() {
        mainnetEthClient, err = ethclient.Dial(cfg.Chains.Eth1.MainnetProvider)
    })
    return mainnetEthClient, err
}


func getRocketPool(cfg config.RocketPoolConfig, client *ethclient.Client) (*rocketpool.RocketPool, error) {
    var err error
    initRocketPool.Do(func() {
        rocketPool, err = rocketpool.NewRocketPool(client, common.HexToAddress(cfg.Rocketpool.StorageAddress))
    })
    return rocketPool, err
}


func getOneInchOracle(cfg config.RocketPoolConfig, client *ethclient.Client) (*contracts.OneInchOracle, error) {
    var err error
    initOneInchOracle.Do(func() {
        oneInchOracle, err = contracts.NewOneInchOracle(common.HexToAddress(cfg.Rocketpool.OneInchOracleAddress), client)
    })
    return oneInchOracle, err
}


func getBeaconClient(cfg config.RocketPoolConfig) (beacon.Client, error) {
    var err error
    initBeaconClient.Do(func() {
        switch cfg.Chains.Eth2.Client.Selected {
            case "lighthouse":
                beaconClient = lighthouse.NewClient(cfg.Chains.Eth2.Provider)
            case "nimbus":
                beaconClient, err = nimbus.NewClient(cfg.Chains.Eth2.Provider)
            case "prysm":
                beaconClient, err = prysm.NewClient(cfg.Chains.Eth2.Provider)
            case "teku":
                beaconClient = teku.NewClient(cfg.Chains.Eth2.Provider)
            default:
                err = fmt.Errorf("Unknown Eth 2.0 client '%s' selected", cfg.Chains.Eth2.Client.Selected)
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
