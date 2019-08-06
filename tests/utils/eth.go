package utils

import (
    "context"
    "crypto/ecdsa"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Get the owner account
func OwnerAccount() (*ecdsa.PrivateKey, common.Address, error) {

    // Initialise private key
    privateKey, err := crypto.HexToECDSA(OWNER_PRIVATE_KEY)
    if err != nil { return nil, common.Address{}, err }

    // Get public key
    publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
    if !ok { return nil, common.Address{}, errors.New("Failed to get owner account public key") }

    // Return
    return privateKey, crypto.PubkeyToAddress(*publicKey), nil

}


// Seed an address with ether from the owner account
func SeedAccount(client *ethclient.Client, address common.Address, amount *big.Int) error {

    // Get owner account
    ownerPrivateKey, ownerAddress, err := OwnerAccount()
    if err != nil { return err }

    // Get owner tx nonce
    nonce, err := client.PendingNonceAt(context.Background(), ownerAddress)
    if err != nil { return err }

    // Initialise tx
    tx := types.NewTransaction(nonce, address, amount, 8000000, big.NewInt(20000), []byte{})

    // Get chain ID
    chainID, err := client.NetworkID(context.Background())
    if err != nil { return err }

    // Sign tx
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ownerPrivateKey)
    if err != nil { return err }

    // Send tx and wait until mined
    if err = client.SendTransaction(context.Background(), signedTx); err != nil { return err }
    if _, err := bind.WaitMined(context.Background(), client, signedTx); err != nil { return err }

    // Return
    return nil

}

