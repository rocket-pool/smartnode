package accounts

import (
    "errors"
    "os"

    "github.com/ethereum/go-ethereum/accounts"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/accounts/keystore"
)


// Keystore passphrase
const PASSPHRASE = ""


// Account manager
type AccountManager struct {
    ks *keystore.KeyStore
}


/**
 * Create new account manager
 */
func NewAccountManager(keychainPath string) *AccountManager {
    return &AccountManager{
        ks: keystore.NewKeyStore(keychainPath, keystore.StandardScryptN, keystore.StandardScryptP),
    }
}


/**
 * Check if the node account exists
 */
func (am *AccountManager) NodeAccountExists() bool {
    return len(am.ks.Accounts()) > 0
}


/**
 * Get the node account
 */
func (am *AccountManager) GetNodeAccount() accounts.Account {
    return am.ks.Accounts()[0]
}


/**
 * Create the node account
 */
func (am *AccountManager) CreateNodeAccount() (accounts.Account, error) {
    return am.ks.NewAccount(PASSPHRASE)
}


/**
 * Get a transactor for the node account
 */
func (am *AccountManager) GetNodeAccountTransactor() (*bind.TransactOpts, error) {

    // Open node account file
    nodeAccountFile, err := os.Open(am.GetNodeAccount().URL.Path)
    if err != nil {
        return nil, errors.New("Error opening node account file: " + err.Error())
    }

    // Create node account transactor
    transactor, err := bind.NewTransactor(nodeAccountFile, PASSPHRASE)
    if err != nil {
        return nil, errors.New("Error creating node account transactor: " + err.Error())
    }

    // Return
    return transactor, nil

}

