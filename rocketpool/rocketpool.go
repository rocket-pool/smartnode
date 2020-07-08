package rocketpool

import (
    "bytes"
    "compress/zlib"
    "encoding/base64"
    "fmt"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/contracts"
)


// Cache settings
const CACHE_TTL = 60 // Seconds


// Cached data types
type cachedAddress struct {
    address *common.Address
    time int64
}
type cachedABI struct {
    abi *abi.ABI
    time int64
}


// Rocket Pool contract manager
type RocketPool struct {
    Client          *ethclient.Client
    RocketStorage   *contracts.RocketStorage
    addresses       map[string]cachedAddress
    abis            map[string]cachedABI
    addressesLock   sync.RWMutex
    abisLock        sync.RWMutex
}


// Create new contract manager
func NewRocketPool(client *ethclient.Client, rocketStorageAddress common.Address) (*RocketPool, error) {

    // Initialize RocketStorage contract
    rocketStorage, err := contracts.NewRocketStorage(rocketStorageAddress, client)
    if err != nil {
        return nil, fmt.Errorf("Could not initialize Rocket Pool storage contract: %w", err)
    }

    // Create and return
    return &RocketPool{
        Client: client,
        RocketStorage: rocketStorage,
        addresses: make(map[string]cachedAddress),
        abis: make(map[string]cachedABI),
    }, nil

}


// Load a Rocket Pool contract address
func (rp *RocketPool) GetAddress(contractName string) (*common.Address, error) {

    // Check for cached address
    if cached, ok := rp.getCachedAddress(contractName); ok {
        if (time.Now().Unix() - cached.time <= CACHE_TTL) {
            return cached.address, nil
        } else {
            rp.deleteCachedAddress(contractName)
        }
    }

    // Get address
    address, err := rp.RocketStorage.GetAddress(nil, crypto.Keccak256Hash([]byte("contract.name"), []byte(contractName)))
    if err != nil {
        return nil, fmt.Errorf("Could not load contract %v address: %w", contractName, err)
    }

    // Cache address
    rp.setCachedAddress(contractName, cachedAddress{
        address: &address,
        time: time.Now().Unix(),
    })

    // Return
    return &address, nil

}


// Load a Rocket Pool contract ABI
func (rp *RocketPool) GetABI(contractName string) (*abi.ABI, error) {

    // Check for cached ABI
    if cached, ok := rp.getCachedABI(contractName); ok {
        if (time.Now().Unix() - cached.time <= CACHE_TTL) {
            return cached.abi, nil
        } else {
            rp.deleteCachedABI(contractName)
        }
    }

    // Get ABI
    abiEncoded, err := rp.RocketStorage.GetString(nil, crypto.Keccak256Hash([]byte("contract.abi"), []byte(contractName)))
    if err != nil {
        return nil, fmt.Errorf("Could not load contract %v ABI: %w", contractName, err)
    }

    // Decode ABI
    abi, err := decodeAbi(abiEncoded)
    if err != nil {
        return nil, err
    }

    // Cache ABI
    rp.setCachedABI(contractName, cachedABI{
        abi: abi,
        time: time.Now().Unix(),
    })

    // Return
    return abi, nil

}


// Load a Rocket Pool contract
func (rp *RocketPool) GetContract(contractName string) (*bind.BoundContract, error) {

    // Contract data
    var wg errgroup.Group
    var contractAddress *common.Address
    var contractAbi *abi.ABI

    // Load address
    wg.Go(func() error {
        address, err := rp.GetAddress(contractName)
        if err == nil { contractAddress = address }
        return err
    })

    // Load ABI
    wg.Go(func() error {
        abi, err := rp.GetABI(contractName)
        if err == nil { contractAbi = abi }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Create and return
    return bind.NewBoundContract(*contractAddress, *contractAbi, rp.Client, rp.Client, rp.Client), nil

}


// Create a Rocket Pool contract instance
func (rp *RocketPool) MakeContract(contractName string, address common.Address) (*bind.BoundContract, error) {
    return nil, nil
}


// Address cache control
func (rp *RocketPool) getCachedAddress(contractName string) (cachedAddress, bool) {
    rp.addressesLock.RLock()
    defer rp.addressesLock.RUnlock()
    value, ok := rp.addresses[contractName]
    return value, ok
}
func (rp *RocketPool) setCachedAddress(contractName string, value cachedAddress) {
    rp.addressesLock.Lock()
    defer rp.addressesLock.Unlock()
    rp.addresses[contractName] = value
}
func (rp *RocketPool) deleteCachedAddress(contractName string) {
    rp.addressesLock.Lock()
    defer rp.addressesLock.Unlock()
    delete(rp.addresses, contractName)
}


// ABI cache control
func (rp *RocketPool) getCachedABI(contractName string) (cachedABI, bool) {
    rp.abisLock.RLock()
    defer rp.abisLock.RUnlock()
    value, ok := rp.abis[contractName]
    return value, ok
}
func (rp *RocketPool) setCachedABI(contractName string, value cachedABI) {
    rp.abisLock.Lock()
    defer rp.abisLock.Unlock()
    rp.abis[contractName] = value
}
func (rp *RocketPool) deleteCachedABI(contractName string) {
    rp.abisLock.Lock()
    defer rp.abisLock.Unlock()
    delete(rp.abis, contractName)
}


// Decode, decompress and parse zlib-compressed, base64-encoded ABI
func decodeAbi(abiEncoded string) (*abi.ABI, error) {

    // Base 64 decode
    abiCompressed, err := base64.StdEncoding.DecodeString(abiEncoded)
    if err != nil {
        return nil, fmt.Errorf("Could not decode contract ABI base64 string: %w", err)
    }

    // Zlib decompress
    byteReader := bytes.NewReader(abiCompressed)
    zlibReader, err := zlib.NewReader(byteReader)
    if err != nil {
        return nil, fmt.Errorf("Could not decompress contract ABI zlib data: %w", err)
    }

    // Parse ABI
    abiParsed, err := abi.JSON(zlibReader)
    if err != nil {
        return nil, fmt.Errorf("Could not parse contract ABI JSON: %w", err)
    }

    // Return
    return &abiParsed, nil

}

