package rocketpool

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "reflect"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Contract type wraps go-ethereum bound contract
type Contract struct {
    Contract *bind.BoundContract
    Address *common.Address
    ABI *abi.ABI
    Client *ethclient.Client
}


// Call a contract method
func (c *Contract) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
    return c.Contract.Call(opts, result, method, params...)
}


// Transact on a contract method and wait for a receipt
func (c *Contract) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Receipt, error) {

    // Send transaction
    tx, err := c.Contract.Transact(opts, method, params...)
    if err != nil {
        return nil, err
    }

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


// Transfer ETH to a contract
func (c *Contract) Transfer(opts *bind.TransactOpts) (*types.Receipt, error) {

    // Send transaction
    tx, err := c.Contract.Transfer(opts)
    if err != nil {
        return nil, err
    }

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

