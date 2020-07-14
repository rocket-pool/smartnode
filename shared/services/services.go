package services

import (
    "sync"

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
    initCfg sync.Once
    initPasswordManager sync.Once
    initAccountManager sync.Once
)


//
// Service instance getters
//


func getConfig(c *cli.Context) (config.RocketPoolConfig, error) {
    var err error
    initCfg.Do(func() {
        _, cfg, err = config.Load(c.GlobalString("config"), c.GlobalString("settings"))
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

