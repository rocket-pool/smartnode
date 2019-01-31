package rocketpool

import (
    "bytes"
    "compress/zlib"
    "encoding/base64"
    "errors"
    "strings"

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
        return nil, errors.New("Error initialising Rocket Pool storage contract: " + err.Error())
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
func (contractManager *ContractManager) LoadContracts(contractNames []string) error {

    // Load contracts
    contractChannels := make(map[string]chan *bind.BoundContract)
    errorChannels := make(map[string]chan error)
    for _, contractName := range contractNames {
        contractChannels[contractName] = make(chan *bind.BoundContract)
        errorChannels[contractName] = make(chan error)
        go loadContract(contractManager.client, contractManager.RocketStorage, contractName, contractChannels[contractName], errorChannels[contractName])
    }

    // Receive loaded contracts
    errs := []string{"Error loading Rocket Pool contracts:"}
    for _, contractName := range contractNames {
        select {
            case contractManager.Contracts[contractName] = <-contractChannels[contractName]:
            case err := <-errorChannels[contractName]:
                errs = append(errs, "Error loading contract " + contractName + ": " + err.Error())
        }
    }

    // Return
    if len(errs) == 1 {
        return nil
    } else {
        return errors.New(strings.Join(errs, "\n"))
    }

}


/**
 * Load and initialise a contract from stored chain data
 */
func loadContract(client bind.ContractBackend, rocketStorage *contracts.RocketStorage, name string, contractChannel chan *bind.BoundContract, errorChannel chan error) {

    // Load contract address from storage
    contractAddress := make(chan common.Address)
    go (func() {

        // Get contract address
        address, err := rocketStorage.GetAddress(nil, eth.KeccakStr("contract.name" + name))
        if err == nil {
            contractAddress <- address
        } else {
            errorChannel <- errors.New("Error retrieving contract address: " + err.Error())
        }

    })()

    // Load contract ABI from storage
    contractAbi := make(chan *abi.ABI)
    go (func() {

        // Get contract ABI
        abiEncoded, err := rocketStorage.GetString(nil, eth.KeccakStr("contract.abi" + name))
        if err == nil {

            // Decode, decompress, parse & send ABI
            abi, err := decodeAbi(abiEncoded)
            if err == nil {
                contractAbi <- abi
            } else {
                errorChannel <- err
            }

        } else {
            errorChannel <- errors.New("Error retrieving contract ABI: " + err.Error())
        }

    })()

    // Initialise and send contract
    contractChannel <- bind.NewBoundContract(<-contractAddress, *(<-contractAbi), client, client, client)

}


/**
 * Decode, decompress and parse zlib-compressed, base64-encoded ABI
 */
func decodeAbi(abiEncoded string) (*abi.ABI, error) {

    // Base 64 decode
    abiCompressed, err := base64.StdEncoding.DecodeString(abiEncoded)
    if err != nil {
        return nil, errors.New("Error decoding contract ABI base64 string: " + err.Error())
    }

    // Zlib decompress
    byteReader := bytes.NewReader(abiCompressed)
    zlibReader, err := zlib.NewReader(byteReader)
    if err != nil {
        return nil, errors.New("Error decompressing contract ABI zlib data: " + err.Error())
    }

    // Parse ABI
    abiParsed, err := abi.JSON(zlibReader)
    if err != nil {
        return nil, errors.New("Error parsing contract ABI JSON: " + err.Error())
    }

    // Return
    return &abiParsed, nil

}

