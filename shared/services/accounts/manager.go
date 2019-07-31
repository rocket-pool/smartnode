package accounts

import (
    "errors"
    "os"

    "github.com/ethereum/go-ethereum/accounts"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/accounts/keystore"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
)


// Account manager
type AccountManager struct {
    ks *keystore.KeyStore
    pm *passwords.PasswordManager
}


/**
 * Create new account manager
 */
func NewAccountManager(keychainPath string, passwordManager *passwords.PasswordManager) *AccountManager {
    return &AccountManager{
        ks: keystore.NewKeyStore(keychainPath, keystore.StandardScryptN, keystore.StandardScryptP),
        pm: passwordManager,
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
func (am *AccountManager) GetNodeAccount() (accounts.Account, error) {

    // Check node account exists
    if !am.NodeAccountExists() { return accounts.Account{}, errors.New("Node account does not exist.") }

    // Return
    return am.ks.Accounts()[0], nil

}


/**
 * Create the node account
 */
func (am *AccountManager) CreateNodeAccount() (accounts.Account, error) {

    // Check node account does not exist
    if am.NodeAccountExists() { return accounts.Account{}, errors.New("Node account already exists.") }

    // Get keystore passphrase
    passphrase, err := am.pm.GetPassphrase()
    if err != nil {
        return accounts.Account{}, errors.New("Error retrieving node keystore passphrase: " + err.Error())
    }

    // Get node account
    account, err := am.ks.NewAccount(passphrase)
    if err != nil {
        return accounts.Account{}, errors.New("Error creating node account: " + err.Error())
    }

    // Return
    return account, nil

}


/**
 * Get a transactor for the node account
 */
func (am *AccountManager) GetNodeAccountTransactor() (*bind.TransactOpts, error) {

    // Check node account exists
    if !am.NodeAccountExists() { return nil, errors.New("Node account does not exist.") }

    // Open node account file
    nodeAccount, _ := am.GetNodeAccount()
    nodeAccountFile, err := os.Open(nodeAccount.URL.Path)
    if err != nil {
        return nil, errors.New("Error opening node account file: " + err.Error())
    }

    // Get keystore passphrase
    passphrase, err := am.pm.GetPassphrase()
    if err != nil {
        return nil, errors.New("Error retrieving node keystore passphrase: " + err.Error())
    }

    // Create node account transactor
    transactor, err := bind.NewTransactor(nodeAccountFile, passphrase)
    if err != nil {
        return nil, errors.New("Error creating node account transactor: " + err.Error())
    }

    // Return
    return transactor, nil

}

