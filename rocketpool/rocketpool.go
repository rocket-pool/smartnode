package rocketpool

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/contracts"
)

// Cache settings
const CacheTTL = 300 // 5 minutes

// Cached data types
type cachedAddress struct {
	address *common.Address
	time    int64
}
type cachedABI struct {
	abi  *abi.ABI
	time int64
}
type cachedContract struct {
	contract *Contract
	time     int64
}

// Rocket Pool contract manager
type RocketPool struct {
	Client                ExecutionClient
	RocketStorage         *contracts.RocketStorage
	RocketStorageContract *Contract
	VersionManager        *VersionManager
	addresses             map[string]cachedAddress
	abis                  map[string]cachedABI
	contracts             map[string]cachedContract
	addressesLock         sync.RWMutex
	abisLock              sync.RWMutex
	contractsLock         sync.RWMutex
}

// Create new contract manager
func NewRocketPool(client ExecutionClient, rocketStorageAddress common.Address) (*RocketPool, error) {

	// Initialize RocketStorage contract
	rocketStorage, err := contracts.NewRocketStorage(rocketStorageAddress, client)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize Rocket Pool storage contract: %w", err)
	}

	// Create a Contract for it
	rsAbi, err := abi.JSON(strings.NewReader(contracts.RocketStorageABI))
	if err != nil {
		return nil, err
	}
	contract := &Contract{
		Contract: bind.NewBoundContract(rocketStorageAddress, rsAbi, client, client, client),
		Address:  &rocketStorageAddress,
		ABI:      &rsAbi,
		Client:   client,
	}

	// Create and return
	rp := &RocketPool{
		Client:                client,
		RocketStorage:         rocketStorage,
		RocketStorageContract: contract,
		addresses:             make(map[string]cachedAddress),
		abis:                  make(map[string]cachedABI),
		contracts:             make(map[string]cachedContract),
	}
	rp.VersionManager = NewVersionManager(rp)

	return rp, nil

}

// Load Rocket Pool contract addresses
func (rp *RocketPool) GetAddress(contractName string, opts *bind.CallOpts) (*common.Address, error) {

	// Check for cached address
	if opts == nil {
		if cached, ok := rp.getCachedAddress(contractName); ok {
			if time.Now().Unix()-cached.time <= CacheTTL {
				return cached.address, nil
			} else {
				rp.deleteCachedAddress(contractName)
			}
		}
	}

	// Get address
	address, err := rp.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.address"), []byte(contractName)))
	if err != nil {
		return nil, fmt.Errorf("Could not load contract %s address: %w", contractName, err)
	}

	// Cache address
	if opts == nil {
		rp.setCachedAddress(contractName, cachedAddress{
			address: &address,
			time:    time.Now().Unix(),
		})
	}

	// Return
	return &address, nil

}

func (rp *RocketPool) GetAddresses(opts *bind.CallOpts, contractNames ...string) ([]*common.Address, error) {

	// Data
	var wg errgroup.Group
	addresses := make([]*common.Address, len(contractNames))

	// Load addresses
	for ci, contractName := range contractNames {
		ci, contractName := ci, contractName
		wg.Go(func() error {
			address, err := rp.GetAddress(contractName, opts)
			if err == nil {
				addresses[ci] = address
			}
			return err
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return
	return addresses, nil

}

// Load Rocket Pool contract ABIs
func (rp *RocketPool) GetABI(contractName string, opts *bind.CallOpts) (*abi.ABI, error) {

	// Check for cached ABI
	if opts == nil {
		if cached, ok := rp.getCachedABI(contractName); ok {
			if time.Now().Unix()-cached.time <= CacheTTL {
				return cached.abi, nil
			} else {
				rp.deleteCachedABI(contractName)
			}
		}
	}

	// Get ABI
	abiEncoded, err := rp.RocketStorage.GetString(opts, crypto.Keccak256Hash([]byte("contract.abi"), []byte(contractName)))
	if err != nil {
		return nil, fmt.Errorf("Could not load contract %s ABI: %w", contractName, err)
	}

	// Decode ABI
	abi, err := DecodeAbi(abiEncoded)
	if err != nil {
		return nil, fmt.Errorf("Could not decode contract %s ABI: %w", contractName, err)
	}

	// Cache ABI
	if opts == nil {
		rp.setCachedABI(contractName, cachedABI{
			abi:  abi,
			time: time.Now().Unix(),
		})
	}

	// Return
	return abi, nil

}
func (rp *RocketPool) GetABIs(opts *bind.CallOpts, contractNames ...string) ([]*abi.ABI, error) {

	// Data
	var wg errgroup.Group
	abis := make([]*abi.ABI, len(contractNames))

	// Load ABIs
	for ci, contractName := range contractNames {
		ci, contractName := ci, contractName
		wg.Go(func() error {
			abi, err := rp.GetABI(contractName, opts)
			if err == nil {
				abis[ci] = abi
			}
			return err
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return
	return abis, nil

}

// Load Rocket Pool contracts
func (rp *RocketPool) GetContract(contractName string, opts *bind.CallOpts) (*Contract, error) {

	// Check for cached contract
	if opts == nil {
		if cached, ok := rp.getCachedContract(contractName); ok {
			if time.Now().Unix()-cached.time <= CacheTTL {
				return cached.contract, nil
			} else {
				rp.deleteCachedContract(contractName)
			}
		}
	}

	// Data
	var wg errgroup.Group
	var address *common.Address
	var abi *abi.ABI

	// Load data
	wg.Go(func() error {
		var err error
		address, err = rp.GetAddress(contractName, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		abi, err = rp.GetABI(contractName, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Create contract
	contract := &Contract{
		Contract: bind.NewBoundContract(*address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  address,
		ABI:      abi,
		Client:   rp.Client,
	}

	// Cache contract
	rp.setCachedContract(contractName, cachedContract{
		contract: contract,
		time:     time.Now().Unix(),
	})

	// Return
	return contract, nil

}
func (rp *RocketPool) GetContracts(opts *bind.CallOpts, contractNames ...string) ([]*Contract, error) {

	// Data
	var wg errgroup.Group
	contracts := make([]*Contract, len(contractNames))

	// Load contracts
	for ci, contractName := range contractNames {
		ci, contractName := ci, contractName
		wg.Go(func() error {
			contract, err := rp.GetContract(contractName, opts)
			if err == nil {
				contracts[ci] = contract
			}
			return err
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return
	return contracts, nil

}

// Create a Rocket Pool contract instance
func (rp *RocketPool) MakeContract(contractName string, address common.Address, opts *bind.CallOpts) (*Contract, error) {

	// Load ABI
	abi, err := rp.GetABI(contractName, opts)
	if err != nil {
		return nil, err
	}

	// Create and return
	return &Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}, nil

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

// Contract cache control
func (rp *RocketPool) getCachedContract(contractName string) (cachedContract, bool) {
	rp.contractsLock.RLock()
	defer rp.contractsLock.RUnlock()
	value, ok := rp.contracts[contractName]
	return value, ok
}
func (rp *RocketPool) setCachedContract(contractName string, value cachedContract) {
	rp.contractsLock.Lock()
	defer rp.contractsLock.Unlock()
	rp.contracts[contractName] = value
}
func (rp *RocketPool) deleteCachedContract(contractName string) {
	rp.contractsLock.Lock()
	defer rp.contractsLock.Unlock()
	delete(rp.contracts, contractName)
}
