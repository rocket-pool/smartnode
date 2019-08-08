package eth

import (
    "context"
    "errors"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Config
const GAS_LIMIT_PADDING uint64 = 100000
const MAX_GAS_LIMIT uint64 = 8000000


// Executes a transaction on a contract method and returns a transaction receipt
// May block for long periods of time due to waiting for the transaction to be mined
func ExecuteContractTransaction(client *ethclient.Client, txor *bind.TransactOpts, contractAddress *common.Address, contractAbi *abi.ABI, method string, params ...interface{}) (*types.Receipt, error) {

    // Create contract instance
    contract := bind.NewBoundContract(*contractAddress, *contractAbi, client, client, client)

    // Estimate gas limit if not set
    if txor.GasLimit == 0 {

        // Pack transaction input
        input, err := contractAbi.Pack(method, params...)
        if err != nil {
            return nil, errors.New("Error packing transaction input: " + err.Error())
        }

        // Get transaction gas limit
        if gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{From: txor.From, To: contractAddress, Value: txor.Value, Data: input}); err != nil {
            return nil, errors.New("Error estimating transaction gas limit: " + err.Error())
        } else {
            txor.GasLimit = (gasLimit + GAS_LIMIT_PADDING)
            if txor.GasLimit > MAX_GAS_LIMIT { txor.GasLimit = MAX_GAS_LIMIT }
        }

    }

    // Execute transaction
    tx, err := contract.Transact(txor, method, params...)
    if err != nil {
        return nil, errors.New("Error executing transaction: " + err.Error())
    }

    // Wait for transaction to be mined before continuing
    txReceipt, err := bind.WaitMined(context.Background(), client, tx)
    if err != nil {
        return nil, errors.New("Error retrieving transaction receipt: " + err.Error())
    }

    // Check transaction status
    if txReceipt.Status == 0 {
        return txReceipt, errors.New("Transaction failed with status 0")
    }

    // Return
    return txReceipt, nil

}

