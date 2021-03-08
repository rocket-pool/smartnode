package eth

import (
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Send a transaction to an address
func SendTransaction(client *ethclient.Client, toAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    var err error

    // Get from address nonce
    var nonce uint64
    if opts.Nonce == nil {
        nonce, err = client.PendingNonceAt(context.Background(), opts.From)
        if err != nil {
            return nil, err
        }
    } else {
        nonce = opts.Nonce.Uint64()
    }

    // Set default value
    value := opts.Value
    if value == nil {
        value = big.NewInt(0)
    }

    // Get suggested gas price
    gasPrice := opts.GasPrice
    if gasPrice == nil {
        gasPrice, err = client.SuggestGasPrice(context.Background())
        if err != nil {
            return nil, err
        }
    }

    // Estimate gas limit
    gasLimit := opts.GasLimit
    if gasLimit == 0 {
        gasLimit, err = client.EstimateGas(context.Background(), ethereum.CallMsg{
            From: opts.From,
            To: &toAddress,
            GasPrice: gasPrice,
            Value: value,
        })
        if err != nil {
            return nil, err
        }
    }

    // Initialize transaction
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, []byte{})

    // Sign transaction
    signedTx, err := opts.Signer(opts.From, tx)
    if err != nil {
        return nil, err
    }

    // Send transaction
    if err = client.SendTransaction(context.Background(), signedTx); err != nil {
        return nil, err
    }

    // Wait for transaction to be mined
    txReceipt, err := bind.WaitMined(context.Background(), client, signedTx)
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

