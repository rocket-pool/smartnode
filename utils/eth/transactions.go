package eth

import (
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Transaction settings
const DefaultGasLimit = 21000


// Send ether to an address
func SendEther(client *ethclient.Client, toAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {

    // Get from address nonce
    if opts.Nonce == nil {
        nonce, err := client.PendingNonceAt(context.Background(), opts.From)
        if err != nil {
            return nil, err
        }
        opts.Nonce = big.NewInt(int64(nonce))
    }

    // Set default gas limit
    if opts.GasLimit == 0 {
        opts.GasLimit = DefaultGasLimit
    }

    // Get suggested gas price
    if opts.GasPrice == nil {
        gasPrice, err := client.SuggestGasPrice(context.Background())
        if err != nil {
            return nil, err
        }
        opts.GasPrice = gasPrice
    }

    // Initialize transaction
    tx := types.NewTransaction(opts.Nonce.Uint64(), toAddress, opts.Value, opts.GasLimit, opts.GasPrice, []byte{})

    // Sign transaction
    signedTx, err := opts.Signer(types.HomesteadSigner{}, opts.From, tx)
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

