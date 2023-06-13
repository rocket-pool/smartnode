package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas of SendTransaction
func EstimateSendTransactionGas(client rocketpool.ExecutionClient, toAddress common.Address, data []byte, useSafeGasLimit bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {

	// User-defined settings
	response := rocketpool.GasInfo{}

	// Set default value
	value := opts.Value
	if value == nil {
		value = big.NewInt(0)
	}

	// Set default data
	if data == nil {
		data = []byte{}
	}

	// Estimate gas limit
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     opts.From,
		To:       &toAddress,
		GasPrice: big.NewInt(0), // set to 0 for simulation
		Data:     data,
		Value:    value,
	})
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	response.EstGasLimit = gasLimit

	if useSafeGasLimit {
		response.SafeGasLimit = uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
	} else {
		response.SafeGasLimit = gasLimit
	}

	return response, err
}

// Send a transaction to an address
// useSafeGasLimit will amplify the estimated gas limit to by 50% for safety (no effect if the gas limit in opts is already set).
func SendTransaction(client rocketpool.ExecutionClient, toAddress common.Address, chainID *big.Int, data []byte, useSafeGasLimit bool, opts *bind.TransactOpts) (common.Hash, error) {
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

	// Set default data
	if data == nil {
		data = []byte{}
	}

	// Estimate gas limit
	gasLimit := opts.GasLimit
	if gasLimit == 0 {
		gasLimit, err = client.EstimateGas(context.Background(), ethereum.CallMsg{
			From:     opts.From,
			To:       &toAddress,
			GasPrice: big.NewInt(0), // use 0 gwei for simulation
			Data:     data,
			Value:    value,
		})
		if err != nil {
			return common.Hash{}, err
		}

		if useSafeGasLimit {
			gasLimit = uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
		}
	}

	// Initialize transaction
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasTipCap:  opts.GasTipCap,
		GasFeeCap:  opts.GasFeeCap,
		Gas:        gasLimit,
		To:         &toAddress,
		Value:      value,
		Data:       data,
		AccessList: []types.AccessTuple{},
	})

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
