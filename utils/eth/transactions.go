package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Send a transaction to an address
func SendTransaction(client *ethclient.Client, toAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
    var err error

    // Get from address nonce
    var nonce uint64
    if opts.Nonce == nil {
        nonce, err = client.PendingNonceAt(context.Background(), opts.From)
        if err != nil {
            return common.Hash{}, err
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
            return common.Hash{}, err
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
            return common.Hash{}, err
        }
    }

    // Initialize transaction
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, []byte{})

    // Sign transaction
    signedTx, err := opts.Signer(opts.From, tx)
    if err != nil {
        return common.Hash{}, err
    }

    // Send transaction
    if err = client.SendTransaction(context.Background(), signedTx); err != nil {
        return common.Hash{}, err
    }

    return signedTx.Hash(), nil

}

