package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/utils/client"
)

// Wait for a transaction to get mined
func WaitForTransaction(client *client.EthClientProxy, hash common.Hash) (*types.Receipt, error) {
    
    var tx *types.Transaction
    var err error

    // Get the transaction from its hash, retrying for 30 sec if it wasn't found
    for i := 0; i < 30; i++ {
        if i == 29 {
            return nil, fmt.Errorf("Transaction not found after 30 seconds.")
        }

        tx, _, err = client.TransactionByHash(context.Background(), hash)
        if err != nil {
            if err.Error() == "not found" {
                time.Sleep(1 * time.Second)
                continue;
            }
            return nil, err
        } else {
            break
        }
    }

    // Wait for transaction to be mined
    txReceipt, err := bind.WaitMined(context.Background(), client, tx)
    if err != nil {
        return nil, err
    }

    // Check transaction status
    if txReceipt.Status == 0 {
        return txReceipt, errors.New("Transaction failed with status 0")
    }

    // Return
    return txReceipt, nil
}

