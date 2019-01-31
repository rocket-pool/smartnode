package rocketpool

import (
    "bytes"
    "compress/zlib"
    "encoding/base64"
    "errors"
    "log"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode-cli/rocketpool/contracts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Contract manager
type ContractManager struct {
    client          *ethclient.Client
    RocketStorage   *contracts.RocketStorage
    Contracts       map[string]*bind.BoundContract
}


/**
 * Create new contract manager
 */
func NewContractManager(client *ethclient.Client, rocketStorageAddress string) (*ContractManager, error) {

    // Initialise RocketStorage contract
    rocketStorage, err := contracts.NewRocketStorage(common.HexToAddress(rocketStorageAddress), client)
    if err != nil {
        return nil, errors.New("Error initialising RocketStorage: " + err.Error())
    }

    // Return
    return &ContractManager{
        client: client,
        RocketStorage: rocketStorage,
        Contracts: make(map[string]*bind.BoundContract),
    }, nil

}


/**
 * Load and initialise contracts
 */
func (contractManager *ContractManager) LoadContracts(contractNames []string) {

    // Load contracts
    contractChannels := make(map[string]chan *bind.BoundContract)
    for _, contractName := range contractNames {
        contractChannels[contractName] = make(chan *bind.BoundContract)
        go loadContract(contractManager.client, contractManager.RocketStorage, contractName, contractChannels[contractName])
    }

    // Receive loaded contracts
    for _, contractName := range contractNames {
        contractManager.Contracts[contractName] = <-contractChannels[contractName]
    }

}


/**
 * Load and initialise a contract from stored chain data
 */
func loadContract(client bind.ContractBackend, rocketStorage *contracts.RocketStorage, name string, contract chan *bind.BoundContract) {

    // Load contract address from storage
    contractAddress := make(chan common.Address)
    go (func() {

        // Get contract address
        address, err := rocketStorage.GetAddress(nil, eth.KeccakStr("contract.name" + name))
        if err != nil {
            log.Fatal("Error retrieving contract address: ", err)
        }

        // Send
        contractAddress <- address

    })()

    // Load contract ABI from storage
    contractAbi := make(chan abi.ABI)
    go (func() {

        // Get contract ABI
        abiEncoded, err := rocketStorage.GetString(nil, eth.KeccakStr("contract.abi" + name))
        if err != nil {
            log.Fatal("Error retrieving contract ABI: ", err)
        }

        // Decode, decompress, parse & send
        contractAbi <- decodeAbi(abiEncoded)

    })()

    // Initialise and send contract
    contract <- bind.NewBoundContract(<-contractAddress, <-contractAbi, client, client, client)

}


/**
 * Decode, decompress and parse zlib-compressed, base64-encoded ABI
 */
func decodeAbi(abiEncoded string) abi.ABI {

    // Base 64 decode
    abiCompressed, err := base64.StdEncoding.DecodeString(abiEncoded)
    if err != nil {
        log.Fatal("Error decoding ABI base64 string: ", err)
    }

    // Zlib decompress
    byteReader := bytes.NewReader(abiCompressed)
    zlibReader, err := zlib.NewReader(byteReader)
    if err != nil {
        log.Fatal("Error decompressing ABI zlib data: ", err)
    }

    // Parse ABI
    abiParsed, err := abi.JSON(zlibReader)
    if err != nil {
        log.Fatal("Error parsing ABI JSON: ", err)
    }

    // Return
    return abiParsed

}

