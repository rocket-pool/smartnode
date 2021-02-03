package services

import (
    "fmt"
    "sync"

    "github.com/docker/docker/client"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/services/beacon/lighthouse"
    "github.com/rocket-pool/smartnode/shared/services/beacon/prysm"
    "github.com/rocket-pool/smartnode/shared/services/beacon/teku"
    "github.com/rocket-pool/smartnode/shared/services/beacon/nimbus"
    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
    lhkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
    prkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
    tkkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
    nmkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
)


// Config
const DockerAPIVersion = "1.40"


// Service instances & initializers
var (
    cfg config.RocketPoolConfig
    passwordManager *passwords.PasswordManager
    nodeWallet *wallet.Wallet
    ethClient *ethclient.Client
    rocketPool *rocketpool.RocketPool
    beaconClient beacon.Client
    docker *client.Client

    initCfg sync.Once
    initPasswordManager sync.Once
    initNodeWallet sync.Once
    initEthClient sync.Once
    initRocketPool sync.Once
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


func getConfig(c *cli.Context) (config.RocketPoolConfig, error) {
    var err error
    initCfg.Do(func() {
        cfg, err = config.Load(c)
    })
    return cfg, err
}


func getPasswordManager(cfg config.RocketPoolConfig) *passwords.PasswordManager {
    initPasswordManager.Do(func() {
        passwordManager = passwords.NewPasswordManager(cfg.Smartnode.PasswordPath)
    })
    return passwordManager
}


func getWallet(cfg config.RocketPoolConfig, pm *passwords.PasswordManager) (*wallet.Wallet, error) {
    var err error
    initNodeWallet.Do(func() {
        nodeWallet, err = wallet.NewWallet(cfg.Smartnode.WalletPath, pm)
        if err == nil {
            lighthouseKeystore := lhkeystore.NewKeystore(cfg.Smartnode.ValidatorKeychainPath, pm)
            prysmKeystore := prkeystore.NewKeystore(cfg.Smartnode.ValidatorKeychainPath, pm)
            tekuKeystore := tkkeystore.NewKeystore(cfg.Smartnode.ValidatorKeychainPath, pm)
			nimbusKeystore := nmkeystore.NewKeystore(cfg.Smartnode.ValidatorKeychainPath, pm)
			nodeWallet.AddKeystore("lighthouse", lighthouseKeystore)
            nodeWallet.AddKeystore("prysm", prysmKeystore)
            nodeWallet.AddKeystore("teku", tekuKeystore)
            nodeWallet.AddKeystore("nimbus", nimbusKeystore)
        }
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


func getRocketPool(cfg config.RocketPoolConfig, client *ethclient.Client) (*rocketpool.RocketPool, error) {
    var err error
    initRocketPool.Do(func() {
        rocketPool, err = rocketpool.NewRocketPool(client, common.HexToAddress(cfg.Rocketpool.StorageAddress))
    })
    return rocketPool, err
}


func getBeaconClient(cfg config.RocketPoolConfig) (beacon.Client, error) {
    var err error
    initBeaconClient.Do(func() {
        switch cfg.Chains.Eth2.Client.Selected {
            case "lighthouse":
                beaconClient = lighthouse.NewClient(cfg.Chains.Eth2.Provider)
            case "prysm":
                beaconClient, err = prysm.NewClient(cfg.Chains.Eth2.Provider)
            case "teku":
				beaconClient = teku.NewClient(cfg.Chains.Eth2.Provider)
			case "nimbus":
                beaconClient, err = nimbus.NewClient(cfg.Chains.Eth2.Provider)
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
