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

    "github.com/rocket-pool/smartnode-cli/shared/contracts"
    "github.com/rocket-pool/smartnode-cli/shared/utils/eth"
)


// Contract manager
type ContractManager struct {
    client          *ethclient.Client
    RocketStorage   *contracts.RocketStorage
    Addresses       map[string]*common.Address
    Abis            map[string]*abi.ABI
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
        Addresses: make(map[string]*common.Address),
        Abis: make(map[string]*abi.ABI),
        Contracts: make(map[string]*bind.BoundContract),
    }, nil

}


/**
 * Load and initialise contracts
 */
func (cm *ContractManager) LoadContracts(contractNames []string) error {

    // Load contract addresses and ABIs
    addressChannels := make(map[string]chan *common.Address)
    abiChannels := make(map[string]chan *abi.ABI)
    errorChannels := make(map[string]chan error)
    for _, contractName := range contractNames {
        addressChannels[contractName] = make(chan *common.Address)
        abiChannels[contractName] = make(chan *abi.ABI)
        errorChannels[contractName] = make(chan error)
        go loadContractAddress(cm.RocketStorage, contractName, addressChannels[contractName], errorChannels[contractName])
        go loadContractABI(cm.RocketStorage, contractName, abiChannels[contractName], errorChannels[contractName])
    }

    // Receive loaded contract data and initialise
    errs := []string{"Error loading Rocket Pool contracts:"}
    for _, contractName := range contractNames {

        // Receive contract data
        received := 0
        for received != -1 && received < 2 {
            select {
                case cm.Addresses[contractName] = <-addressChannels[contractName]:
                    received++
                case cm.Abis[contractName] = <-abiChannels[contractName]:
                    received++
                case err := <-errorChannels[contractName]:
                    errs = append(errs, "Error loading contract " + contractName + ": " + err.Error())
                    received = -1
            }
        }

        // Initialise contract
        if received != -1 {
            cm.Contracts[contractName] = bind.NewBoundContract(*(cm.Addresses[contractName]), *(cm.Abis[contractName]), cm.client, cm.client, cm.client)
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
            case cm.Abis[contractName] = <-abiChannels[contractName]:
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
    abi, ok := cm.Abis[contractName]
    if !ok {
        return nil, errors.New(fmt.Sprintf("Error initialising Rocket Pool contract %s: ABI not loaded", contractName))
    }

    // Initialise and return contract
    return bind.NewBoundContract(*address, *abi, cm.client, cm.client, cm.client), nil

}


/**
 * Load a contract address from stored chain data
 */
func loadContractAddress(rocketStorage *contracts.RocketStorage, name string, addressChannel chan *common.Address, errorChannel chan error) {

    // Get contract address
    if address, err := rocketStorage.GetAddress(nil, eth.KeccakStr("contract.name" + name)); err != nil {
        errorChannel <- errors.New("Error retrieving contract address: " + err.Error())
    } else {
        addressChannel <- &address
    }

}


/**
 * Load a contract ABI from stored chain data
 */
func loadContractABI(rocketStorage *contracts.RocketStorage, name string, abiChannel chan *abi.ABI, errorChannel chan error) {

    // Get contract ABI
    if abiEncoded, err := rocketStorage.GetString(nil, eth.KeccakStr("contract.abi" + name)); err != nil {
        errorChannel <- errors.New("Error retrieving contract ABI: " + err.Error())
    } else {

        // Decode, decompress, parse & send ABI
        if abi, err := decodeAbi(abiEncoded); err != nil {
            errorChannel <- err
        } else {
            abiChannel <- abi
        }

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

