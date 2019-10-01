package eth

import (
    "context"
    "errors"
    "math/big"

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


// Send ether to an address
// May block for long periods of time due to waiting for the transaction to be mined
func SendEther(client *ethclient.Client, txor *bind.TransactOpts, toAddress *common.Address, amount *big.Int) (*types.Receipt, error) {

    // Get from address tx nonce
    nonce, err := client.PendingNonceAt(context.Background(), txor.From)
    if err != nil { return nil, err }

    // Estimate gas limit if not set
    if txor.GasLimit == 0 {
        if gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{From: txor.From, To: toAddress, Value: amount, Data: []byte{}}); err != nil {
            return nil, errors.New("Error estimating transaction gas limit: " + err.Error())
        } else {
            txor.GasLimit = (gasLimit + GAS_LIMIT_PADDING)
            if txor.GasLimit > MAX_GAS_LIMIT { txor.GasLimit = MAX_GAS_LIMIT }
        }
    }

    // Get suggested gas price
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil { return nil, err }

    // Initialise tx
    tx := types.NewTransaction(nonce, *toAddress, amount, txor.GasLimit, gasPrice, []byte{})

    // Sign tx
    signedTx, err := txor.Signer(types.HomesteadSigner{}, txor.From, tx)
    if err != nil { return nil, err }

    // Send tx and wait until mined
    if err = client.SendTransaction(context.Background(), signedTx); err != nil { return nil, err }
    txReceipt, err := bind.WaitMined(context.Background(), client, signedTx)
    if err != nil { return nil, err }

    // Return
    return txReceipt, nil

}

