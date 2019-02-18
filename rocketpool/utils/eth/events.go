package eth

import (
    "bytes"
    "context"
    "errors"
    "reflect"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Get contract events from a transaction
// eventPrototype must be an event struct and not a pointer to one
func GetTransactionEvents(client *ethclient.Client, tx *types.Transaction, contractAddress *common.Address, contractAbi *abi.ABI, eventName string, eventPrototype interface{}) ([]interface{}, error) {

    // Create contract instance
    contract := bind.NewBoundContract(*contractAddress, *contractAbi, client, client, client)

    // Get event type from prototype
    eventType := reflect.TypeOf(eventPrototype)

    // Get transaction receipt
    receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
    if err != nil {
        return nil, errors.New("Error retrieving transaction receipt: " + err.Error())
    }

    // Process transaction receipt logs
    events := make([]interface{}, 0)
    for _, log := range receipt.Logs {

        // Check log address matches contract address
        if !bytes.Equal(log.Address.Bytes(), contractAddress.Bytes()) {
            continue
        }

        // Check log first topic matches event ID
        if len(log.Topics) == 0 || !bytes.Equal(log.Topics[0].Bytes(), contractAbi.Events[eventName].Id().Bytes()) {
            continue
        }

        // Unpack event
        event := reflect.New(eventType).Interface()
        if err := contract.UnpackLog(event, eventName, *log); err != nil {
            return nil, errors.New("Error unpacking event: " + err.Error())
        }
        events = append(events, event)

    }

    // Return events
    return events, nil

}

