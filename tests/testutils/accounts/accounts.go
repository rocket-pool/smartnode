package accounts

import (
    "context"
    "crypto/ecdsa"
    "encoding/hex"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"

    "github.com/rocket-pool/rocketpool-go/tests"
)


// An account containing a keypair and address
type Account struct {
    PrivateKey  *ecdsa.PrivateKey
    Address     common.Address
}


// Get an account by index
func GetAccount(index uint8) (*Account, error) {

    // Get private key data
    privateKeyBytes, err := hex.DecodeString(tests.AccountPrivateKeys[index])
    if err != nil { return nil, err }

    // Get private key
    privateKey, err := crypto.ToECDSA(privateKeyBytes)
    if err != nil { return nil, err }

    // Return account
    return &Account{
        PrivateKey: privateKey,
        Address: crypto.PubkeyToAddress(privateKey.PublicKey),
    }, nil

}


// Get a transactor for an account
func (a *Account) GetTransactor() *bind.TransactOpts {
    opts := bind.NewKeyedTransactor(a.PrivateKey)
    opts.Context = context.Background()
    return opts
}

