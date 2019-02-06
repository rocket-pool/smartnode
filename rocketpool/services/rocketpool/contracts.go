package rocketpool

import (
    "bytes"
    "compress/zlib"
    "encoding/base64"
    "errors"
    "fmt"
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
    abis            map[string]*abi.ABI
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
        abis: make(map[string]*abi.ABI),
    }, nil

}


/**
 * Load and initialise contracts
 */
func (cm *ContractManager) LoadContracts(contractNames []string) error {

    // Load contracts
    contractChannels := make(map[string]chan *bind.BoundContract)
    errorChannels := make(map[string]chan error)
    for _, contractName := range contractNames {
        contractChannels[contractName] = make(chan *bind.BoundContract)
        errorChannels[contractName] = make(chan error)
        go loadContract(cm.client, cm.RocketStorage, contractName, contractChannels[contractName], errorChannels[contractName])
    }

    // Receive loaded contracts
    errs := []string{"Error loading Rocket Pool contracts:"}
    for _, contractName := range contractNames {
        select {
            case cm.Contracts[contractName] = <-contractChannels[contractName]:
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
 * Load contract ABIs
 */
func (cm *ContractManager) LoadABIs(contractNames []string) error {

    // Load ABIs
    abiChannels := make(map[string]chan *abi.ABI)
    errorChannels := make(map[string]chan error)
    for _, contractName := range contractNames {
        abiChannels[contractName] = make(chan *abi.ABI)
        errorChannels[contractName] = make(chan error)
        go loadContractABI(cm.RocketStorage, contractName, abiChannels[contractName], errorChannels[contractName])
    }

    // Receive loaded ABIs
    errs := []string{"Error loading Rocket Pool contract ABIs:"}
    for _, contractName := range contractNames {
        select {
            case cm.abis[contractName] = <-abiChannels[contractName]:
            case err := <-errorChannels[contractName]:
                errs = append(errs, "Error loading contract " + contractName + " ABI: " + err.Error())
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
 * Initialise a new contract from an address and a contract ABI name
 */
func (cm *ContractManager) NewContract(address *common.Address, contractName string) (*bind.BoundContract, error) {

    // Get ABI
    abi, ok := cm.abis[contractName]
    if !ok {
        return nil, errors.New(fmt.Sprintf("Error initialising Rocket Pool contract %s: ABI not loaded", contractName))
    }

    // Initialise and return contract
    return bind.NewBoundContract(*address, *abi, cm.client, cm.client, cm.client), nil

}


/**
 * Load and initialise a contract from stored chain data
 */
func loadContract(client bind.ContractBackend, rocketStorage *contracts.RocketStorage, name string, contractChannel chan *bind.BoundContract, errorChannel chan error) {

    // Load contract address from storage
    addressChannel := make(chan common.Address)
    go loadContractAddress(rocketStorage, name, addressChannel, errorChannel)

    // Load contract ABI from storage
    abiChannel := make(chan *abi.ABI)
    go loadContractABI(rocketStorage, name, abiChannel, errorChannel)

    // Initialise and send contract
    contractChannel <- bind.NewBoundContract(<-addressChannel, *(<-abiChannel), client, client, client)

}


/**
 * Load a contract address from stored chain data
 */
func loadContractAddress(rocketStorage *contracts.RocketStorage, name string, addressChannel chan common.Address, errorChannel chan error) {

    // Get contract address
    address, err := rocketStorage.GetAddress(nil, eth.KeccakStr("contract.name" + name))
    if err == nil {
        addressChannel <- address
    } else {
        errorChannel <- errors.New("Error retrieving contract address: " + err.Error())
    }

}


/**
 * Load a contract ABI from stored chain data
 */
func loadContractABI(rocketStorage *contracts.RocketStorage, name string, abiChannel chan *abi.ABI, errorChannel chan error) {

    // Get contract ABI
    abiEncoded, err := rocketStorage.GetString(nil, eth.KeccakStr("contract.abi" + name))
    if err == nil {

        // Decode, decompress, parse & send ABI
        abi, err := decodeAbi(abiEncoded)
        if err == nil {
            abiChannel <- abi
        } else {
            errorChannel <- err
        }

    } else {
        errorChannel <- errors.New("Error retrieving contract ABI: " + err.Error())
    }

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

