package contract

import (
    "bytes"
    "errors"
    "fmt"
    "reflect"

    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get contract events from a transaction
// eventPrototype must be an event struct type
// Returns a slice of untyped values; assert returned events to event struct type
func GetTransactionEvents(client *ethclient.Client, contract *rocketpool.Contract, txReceipt *types.Receipt, eventName string, eventPrototype interface{}) ([]interface{}, error) {

    // Get event type
    eventType := reflect.TypeOf(eventPrototype)
    if eventType.Kind() != reflect.Struct {
        return nil, errors.New("Invalid event type")
    }

    // Get ABI event
    abiEvent, ok := contract.ABI.Events[eventName]
    if !ok {
        return nil, fmt.Errorf("Event '%s' does not exist on contract", eventName)
    }

    // Process transaction receipt logs
    events := make([]interface{}, 0)
    for _, log := range txReceipt.Logs {

        // Check log address matches contract address
        if !bytes.Equal(log.Address.Bytes(), contract.Address.Bytes()) {
            continue
        }

        // Check log first topic matches event ID
        if len(log.Topics) == 0 || !bytes.Equal(log.Topics[0].Bytes(), abiEvent.ID.Bytes()) {
            continue
        }

        // Unpack event
        event := reflect.New(eventType)
        if err := contract.Contract.UnpackLog(event.Interface(), eventName, *log); err != nil {
            return nil, fmt.Errorf("Could not unpack event data: %w", err)
        }
        events = append(events, reflect.Indirect(event).Interface())

    }

    // Return events
    return events, nil

}

