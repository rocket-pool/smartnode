package services

import (
    "context"
    "errors"
    "sync"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Service instances & initializers
var (
    cfg config.RocketPoolConfig
    passwordManager *passwords.PasswordManager
    accountManager *accounts.AccountManager
    ethClient *ethclient.Client
    rocketPool *rocketpool.RocketPool

    initCfg sync.Once
    initPasswordManager sync.Once
    initAccountManager sync.Once
    initEthClient sync.Once
    initRocketPool sync.Once
)


//
// Service instance getters
//


func getConfig(c *cli.Context) (config.RocketPoolConfig, error) {
    var err error
    initCfg.Do(func() {
        _, cfg, err = config.Load(c)
    })
    return cfg, err
}


func getPasswordManager(cfg config.RocketPoolConfig) *passwords.PasswordManager {
    initPasswordManager.Do(func() {
        passwordManager = passwords.NewPasswordManager(cfg.Smartnode.PasswordPath)
    })
    return passwordManager
}


func getAccountManager(cfg config.RocketPoolConfig, pm *passwords.PasswordManager) *accounts.AccountManager {
    initAccountManager.Do(func() {
        accountManager = accounts.NewAccountManager(cfg.Smartnode.NodeKeychainPath, pm)
    })
    return accountManager
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


func GetAccountManager(c *cli.Context) (*accounts.AccountManager, error) {
    cfg, err := getConfig(c)
    if err != nil {
        return nil, err
    }
    pm := getPasswordManager(cfg)
    return getAccountManager(cfg, pm), nil
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


//
// Service requirements
//


func RequireNodePassword(c *cli.Context) error {
    pm, err := GetPasswordManager(c)
    if err != nil {
        return err
    }
    if !pm.PasswordExists() {
        return errors.New("The node password has not been set. Please initialize the node and try again.")
    }
    return nil
}


func RequireNodeAccount(c *cli.Context) error {
    if err := RequireNodePassword(c); err != nil {
        return err
    }
    am, err := GetAccountManager(c)
    if err != nil {
        return err
    }
    if !am.NodeAccountExists() {
        return errors.New("The node account has not been created. Please initialize the node and try again.")
    }
    return nil
}


func RequireClientSynced(c *cli.Context) error {
    ec, err := GetEthClient(c)
    if err != nil {
        return err
    }
    progress, err := ec.SyncProgress(context.Background())
    if err != nil {
        return err
    }
    if progress != nil {
        return errors.New("The Eth 1.0 node is currently syncing. Please try again later.")
    }
    return nil
}


func RequireRocketStorage(c *cli.Context) error {
    if err := RequireClientSynced(c); err != nil {
        return err
    }
    cfg, err := GetConfig(c)
    if err != nil {
        return err
    }
    ec, err := GetEthClient(c)
    if err != nil {
        return err
    }
    code, err := ec.CodeAt(context.Background(), common.HexToAddress(cfg.Rocketpool.StorageAddress), nil)
    if err != nil {
        return err
    }
    if len(code) == 0 {
        return errors.New("The Rocket Pool storage contract was not found; the configured address may be incorrect, or the Eth 1.0 node may not be synced. Please try again later.")
    }
    return nil
}


func RequireNodeRegistered(c *cli.Context) error {
    if err := RequireNodeAccount(c); err != nil {
        return err
    }
    if err := RequireRocketStorage(c); err != nil {
        return err
    }
    am, err := GetAccountManager(c)
    if err != nil {
        return err
    }
    rp, err := GetRocketPool(c)
    if err != nil {
        return err
    }
    nodeAccount, _ := am.GetNodeAccount()
    exists, err := node.GetNodeExists(rp, nodeAccount.Address)
    if err != nil {
        return err
    }
    if !exists {
        return errors.New("The node is not registered with Rocket Pool. Please register and try again.")
    }
    return nil
}

