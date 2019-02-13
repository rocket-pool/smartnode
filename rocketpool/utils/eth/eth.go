package eth

import (
    "bytes"
    "context"
    "errors"
    "math/big"
    "reflect"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto/sha3"
    "github.com/ethereum/go-ethereum/ethclient"
)


// Conversion factor from wei to eth
const WEI_PER_ETH = 1000000000000000000


// Convert wei to eth
func WeiToEth(wei *big.Int) float64 {
    var weiFloat big.Float
    var eth big.Float
    weiFloat.SetInt(wei)
    eth.Quo(&weiFloat, big.NewFloat(WEI_PER_ETH))
    eth64, _ := eth.Float64()
    return eth64
}


// Convert eth to wei
func EthToWei(eth float64) *big.Int {
    var weiFloat big.Float
    var wei big.Int
    weiFloat.Mul(big.NewFloat(eth), big.NewFloat(WEI_PER_ETH))
    weiFloat.Int(&wei)
    return &wei
}


// Make a keccak256 hash of a source byte slice and return as a 32-byte array
func KeccakBytes(src []byte) [32]byte {

    // Hash source data
    hash := sha3.NewKeccak256()
    hash.Write(src[:])

    // Copy hashed data to byte array
    var bytes [32]byte
    copy(bytes[:], hash.Sum(nil))

    // Return
    return bytes

}


// Make a keccak256 hash of a source string and return as a 32-byte array
func KeccakStr(src string) [32]byte {
    return KeccakBytes([]byte(src))
}


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
        err = contract.UnpackLog(event, eventName, *log)
        if err != nil {
            return nil, errors.New("Error unpacking event: " + err.Error())
        }
        events = append(events, event)

    }

    // Return events
    return events, nil

}

