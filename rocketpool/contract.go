package rocketpool

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Transaction settings
const (
	GasLimitMultiplier    float64 = 1.5
	MaxGasLimit           uint64  = 30000000
	NethermindRevertRegex string  = "Reverted 0x(?P<message>[0-9a-fA-F]+).*"
)

// Contract type wraps go-ethereum bound contract
type Contract struct {
	Contract *bind.BoundContract
	Address  *common.Address
	ABI      *abi.ABI
	Client   ExecutionClient
}

// Response for gas limits from network and from user request
type GasInfo struct {
	EstGasLimit  uint64 `json:"estGasLimit"`
	SafeGasLimit uint64 `json:"safeGasLimit"`
}

// Call a contract method
func (c *Contract) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	results := make([]interface{}, 1)
	results[0] = result
	return c.Contract.Call(opts, &results, method, params...)
}

// Get Gas Limit for transaction
func (c *Contract) GetTransactionGasInfo(opts *bind.TransactOpts, method string, params ...interface{}) (GasInfo, error) {

	response := GasInfo{}

	// Pack transaction Info
	input, err := c.ABI.Pack(method, params...)
	if err != nil {
		return response, fmt.Errorf("Error getting transaction gas info: Could not encode input data: %w", err)
	}

	// Estimate gas limit
	estGasLimit, safeGasLimit, err := c.estimateGasLimit(opts, input)

	if err != nil {
		return response, fmt.Errorf("Error getting transaction gas info: could not estimate gas limit: %w", err)
	}
	response.EstGasLimit = estGasLimit
	response.SafeGasLimit = safeGasLimit

	return response, err
}

// Transact on a contract method and wait for a receipt
func (c *Contract) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {

	// Estimate gas limit
	if opts.GasLimit == 0 {
		input, err := c.ABI.Pack(method, params...)
		if err != nil {
			return nil, fmt.Errorf("Could not encode input data: %w", err)
		}
		_, safeGasLimit, err := c.estimateGasLimit(opts, input)
		if err != nil {
			return nil, err
		}
		opts.GasLimit = safeGasLimit
	}

	// Send transaction
	tx, err := c.Contract.Transact(opts, method, params...)
	if err != nil {
		return nil, c.normalizeErrorMessage(err)
	}

	return tx, nil

}

// Get gas limit for a transfer call
func (c *Contract) GetTransferGasInfo(opts *bind.TransactOpts) (GasInfo, error) {

	response := GasInfo{}

	// Estimate gas limit
	estGasLimit, safeGasLimit, err := c.estimateGasLimit(opts, []byte{})
	if err != nil {
		return response, fmt.Errorf("Error getting transfer gas info: could not estimate gas limit: %w", err)
	}
	response.EstGasLimit = estGasLimit
	response.SafeGasLimit = safeGasLimit

	return response, nil
}

// Transfer ETH to a contract and wait for a receipt
func (c *Contract) Transfer(opts *bind.TransactOpts) (common.Hash, error) {

	// Estimate gas limit
	if opts.GasLimit == 0 {
		_, safeGasLimit, err := c.estimateGasLimit(opts, []byte{})
		if err != nil {
			return common.Hash{}, err
		}
		opts.GasLimit = safeGasLimit
	}

	// Send transaction
	tx, err := c.Contract.Transfer(opts)
	if err != nil {
		return common.Hash{}, c.normalizeErrorMessage(err)
	}

	return tx.Hash(), nil

}

// Estimate the expected and safe gas limits for a contract transaction
func (c *Contract) estimateGasLimit(opts *bind.TransactOpts, input []byte) (uint64, uint64, error) {

	// Estimate gas limit
	gasLimit, err := c.Client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     opts.From,
		To:       c.Address,
		GasPrice: big.NewInt(0), // use 0 gwei for simulation
		Value:    opts.Value,
		Data:     input,
	})

	if err != nil {
		return 0, 0, fmt.Errorf("Could not estimate gas needed: %w", c.normalizeErrorMessage(err))
	}

	// Pad and return gas limit
	safeGasLimit := uint64(float64(gasLimit) * GasLimitMultiplier)
	if gasLimit > MaxGasLimit {
		return 0, 0, fmt.Errorf("estimated gas of %d is greater than the max gas limit of %d", gasLimit, MaxGasLimit)
	}
	if safeGasLimit > MaxGasLimit {
		safeGasLimit = MaxGasLimit
	}
	return gasLimit, safeGasLimit, nil

}

// Wait for a transaction to be mined and get a tx receipt
func (c *Contract) getTransactionReceipt(tx *types.Transaction) (*types.Receipt, error) {

	// Wait for transaction to be mined
	txReceipt, err := bind.WaitMined(context.Background(), c.Client, tx)
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

// Get contract events from a transaction
// eventPrototype must be an event struct type
// Returns a slice of untyped values; assert returned events to event struct type
func (c *Contract) GetTransactionEvents(txReceipt *types.Receipt, eventName string, eventPrototype interface{}) ([]interface{}, error) {

	// Get event type
	eventType := reflect.TypeOf(eventPrototype)
	if eventType.Kind() != reflect.Struct {
		return nil, errors.New("Invalid event type")
	}

	// Get ABI event
	abiEvent, ok := c.ABI.Events[eventName]
	if !ok {
		return nil, fmt.Errorf("Event '%s' does not exist on contract", eventName)
	}

	// Process transaction receipt logs
	events := make([]interface{}, 0)
	for _, log := range txReceipt.Logs {

		// Check log address matches contract address
		if !bytes.Equal(log.Address.Bytes(), c.Address.Bytes()) {
			continue
		}

		// Check log first topic matches event ID
		if len(log.Topics) == 0 || !bytes.Equal(log.Topics[0].Bytes(), abiEvent.ID.Bytes()) {
			continue
		}

		// Unpack event
		event := reflect.New(eventType)
		if err := c.Contract.UnpackLog(event.Interface(), eventName, *log); err != nil {
			return nil, fmt.Errorf("Could not unpack event data: %w", err)
		}
		events = append(events, reflect.Indirect(event).Interface())

	}

	// Return events
	return events, nil

}

// Normalize error messages so they're all in ASCII format
func (c *Contract) normalizeErrorMessage(err error) error {
	if err == nil {
		return err
	}

	// Get the message in hex format, if it exists
	reg := regexp.MustCompile(NethermindRevertRegex)
	matches := reg.FindStringSubmatch(err.Error())
	if matches == nil {
		return err
	}
	messageIndex := reg.SubexpIndex("message")
	if messageIndex == -1 {
		return err
	}
	message := matches[messageIndex]

	// Convert the hex message to ASCII
	bytes, err2 := hex.DecodeString(message)
	if err2 != nil {
		return err // Return the original error if decoding failed somehow
	}

	return fmt.Errorf("Reverted: %s", string(bytes))
}
