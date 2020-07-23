package services

import (
    "sync"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
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

