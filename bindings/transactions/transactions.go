package transactions

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/bindings/utils"
	log "github.com/rocket-pool/smartnode/shared/logger"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The fraction of the timeout period to trigger overdue transactions
const TimeoutSafetyFactor int = 2

// Print a TX's details to the logger and waits for it to validated.
func PrintAndWaitForTransaction(cfg *config.RocketPoolConfig, hash common.Hash, ec rocketpool.ExecutionClient, logger *log.ColorLogger) error {

	txWatchUrl := cfg.Smartnode.GetTxWatchUrl()
	hashString := hash.String()

	logger.Printlnf("Transaction has been submitted with hash %s.", hashString)
	if txWatchUrl != "" {
		logger.Printlnf("You may follow its progress by visiting:")
		logger.Printlnf("%s/%s\n", txWatchUrl, hashString)
	}
	logger.Println("Waiting for the transaction to be validated...")

	// Wait for the TX to be included in a block
	if _, err := utils.WaitForTransaction(ec, hash); err != nil {
		return fmt.Errorf("Error waiting for transaction: %w", err)
	}

	return nil

}

// True if a transaction is due and needs to bypass the gas threshold
func IsTransactionDue(rp *rocketpool.RocketPool, startTime time.Time) (bool, time.Duration, error) {

	// Get the dissolve timeout
	timeout, err := protocol.GetMinipoolLaunchTimeout(rp, nil)
	if err != nil {
		return false, 0, err
	}

	dueTime := timeout / time.Duration(TimeoutSafetyFactor)
	isDue := time.Since(startTime) > dueTime
	timeUntilDue := time.Until(startTime.Add(dueTime))
	return isDue, timeUntilDue, nil

}

// Estimate the gas of SendTransaction
func EstimateSendTransactionGas(client rocketpool.ExecutionClient, toAddress common.Address, data []byte, useSafeGasLimit bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {

	// User-defined settings
	response := gaslimit.Limits{}

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
		GasPrice: nil,
		Data:     data,
		Value:    value,
	})
	if err != nil {
		return gaslimit.Limits{}, err
	}
	response.Estimated = gasLimit

	if useSafeGasLimit {
		response.Safe = uint64(float64(gasLimit) * rocketpool.GasLimitMultiplier)
	} else {
		response.Safe = gasLimit
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
			GasPrice: nil,
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
