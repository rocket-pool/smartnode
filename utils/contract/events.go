package contract

import (
    "bytes"
    "errors"
    "fmt"
    "reflect"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Get contract events from a transaction
// eventPrototype must be an event struct type
// Returns a slice of pointers to untyped values; assert returned events to *eventType
func GetTransactionEvents(client *ethclient.Client, contractAddress *common.Address, contractAbi *abi.ABI, txReceipt *types.Receipt, eventName string, eventPrototype interface{}) ([]interface{}, error) {

    // Get event type
    eventType := reflect.TypeOf(eventPrototype)
    if eventType.Kind() != reflect.Struct {
        return nil, errors.New("Invalid event type")
    }

    // Get ABI event
    abiEvent, ok := contractAbi.Events[eventName]
    if !ok {
        return nil, fmt.Errorf("Event '%s' does not exist on contract")
    }

    // Create contract instance
    contract := bind.NewBoundContract(*contractAddress, *contractAbi, client, client, client)

    // Process transaction receipt logs
    events := make([]interface{}, 0)
    for _, log := range txReceipt.Logs {

        // Check log address matches contract address
        if !bytes.Equal(log.Address.Bytes(), contractAddress.Bytes()) {
            continue
        }

        // Check log first topic matches event ID
        if len(log.Topics) == 0 || !bytes.Equal(log.Topics[0].Bytes(), abiEvent.ID.Bytes()) {
            continue
        }

        // Unpack event
        event := reflect.New(eventType).Interface()
        if err := contract.UnpackLog(event, eventName, *log); err != nil {
            return nil, fmt.Errorf("Could not unpack event data: %w", err)
        }
        events = append(events, event)

    }

    // Return events
    return events, nil

}

