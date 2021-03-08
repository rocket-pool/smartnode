package rocketpool

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "math/big"
    "reflect"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    "golang.org/x/sync/errgroup"
)


// Transaction settings
const (
    GasLimitPadding = 100000
    MaxGasLimit = 12000000
)


// Contract type wraps go-ethereum bound contract
type Contract struct {
    Contract *bind.BoundContract
    Address *common.Address
    ABI *abi.ABI
    Client *ethclient.Client
}


// Response for gas prices and limits from network and from user request
type GasInfo struct {
    EstGasPrice *big.Int            `json:"estGasPrice"`
    EstGasLimit uint64              `json:"estGasLimit"`
    ReqGasPrice *big.Int            `json:"reqGasPrice"`
    ReqGasLimit uint64              `json:"reqGasLimit"`
}


// Call a contract method
func (c *Contract) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
    results := make([]interface{}, 1)
    results[0] = result
    return c.Contract.Call(opts, &results, method, params...)
}


// Transact on a contract method and wait for a receipt
func (c *Contract) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Receipt, error) {

    // Estimate gas limit
    if opts.GasLimit == 0 {
        input, err := c.ABI.Pack(method, params...)
        if err != nil {
            return nil, fmt.Errorf("Could not encode input data: %w", err)
        }
        gasLimit, err := c.estimateGasLimit(opts, input)
        if err != nil {
            return nil, err
        }
        opts.GasLimit = gasLimit
    }

    // Send transaction
    tx, err := c.Contract.Transact(opts, method, params...)
    if err != nil {
        return nil, err
    }

    // Get & return transaction receipt
    return c.getTransactionReceipt(tx)

}


// Transfer ETH to a contract and wait for a receipt
func (c *Contract) Transfer(opts *bind.TransactOpts) (*types.Receipt, error) {

    // Estimate gas limit
    if opts.GasLimit == 0 {
        gasLimit, err := c.estimateGasLimit(opts, []byte{})
        if err != nil {
            return nil, err
        }
        opts.GasLimit = gasLimit
    }

    // Send transaction
    tx, err := c.Contract.Transfer(opts)
    if err != nil {
        return nil, err
    }

    // Get & return transaction receipt
    return c.getTransactionReceipt(tx)

}


// Get Gas Price and Gas Limit for transaction
func (c *Contract) GetGasInfo(methodName string, opts *bind.TransactOpts, params ...interface{}) (GasInfo, error) {

    // Find user option for gas price and gas limit
    response := GasInfo {
        ReqGasPrice: opts.GasPrice,
        ReqGasLimit: opts.GasLimit,
    }

    input, err := c.ABI.Pack(methodName, params...)
    if err != nil {
        return response, fmt.Errorf("Could not encode input data: %w", err)
    }

    // Sync
    var wg errgroup.Group

    wg.Go(func() error {
        estGasPrice, err := c.Client.SuggestGasPrice(context.Background())
        if err == nil {
            response.EstGasPrice = estGasPrice
        }
        return err
    })

    wg.Go(func() error {
        estGasLimit, err := c.estimateGasLimit(opts, input)
        if err == nil {
            response.EstGasLimit = estGasLimit
        }
        return err
    })

    // Wait for data
    err = wg.Wait()
    return response, err
}


// Estimate the gas limit for a contract transaction
func (c *Contract) estimateGasLimit(opts *bind.TransactOpts, input []byte) (uint64, error) {

    // Estimate gas limit
    gasLimit, err := c.Client.EstimateGas(context.Background(), ethereum.CallMsg{
        From: opts.From,
        To: c.Address,
        GasPrice: opts.GasPrice,
        Value: opts.Value,
        Data: input,
    })
    if err != nil {
        return 0, fmt.Errorf("Could not estimate gas needed: %w", err)
    }

    // Pad and return gas limit
    gasLimit += GasLimitPadding
    if gasLimit > MaxGasLimit { gasLimit = MaxGasLimit }
    return gasLimit, nil

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

