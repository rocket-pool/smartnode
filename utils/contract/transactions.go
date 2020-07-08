package contract

import (
    "context"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Transact on a contract method and wait for a receipt
func Transact(client *ethclient.Client, contract *bind.BoundContract, opts *bind.TransactOpts, method string, params ...interface{}) (*types.Receipt, error) {

    // Send transaction
    tx, err := contract.Transact(opts, method, params...)
    if err != nil {
        return nil, fmt.Errorf("Could not transact on contract: %w", err)
    }

    // Wait for transaction to be mined
    txReceipt, err := bind.WaitMined(context.Background(), client, tx)
    if err != nil {
        return nil, fmt.Errorf("Could not get transaction receipt: %w", err)
    }

    // Check transaction status
    if txReceipt.Status == 0 {
        return txReceipt, errors.New("Transaction failed with status 0")
    }

    // Return
    return txReceipt, nil

}

