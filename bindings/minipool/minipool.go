package minipool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
)

// Settings
const (
	MinipoolPrelaunchBatchSize     = 250
	MinipoolAddressBatchSize       = 50
	MinipoolDetailsBatchSize       = 20
	NativeMinipoolDetailsBatchSize = 1000
)

// Minipool details
type MinipoolDetails struct {
	Address common.Address          `json:"address"`
	Exists  bool                    `json:"exists"`
	Pubkey  rptypes.ValidatorPubkey `json:"pubkey"`
}

// The counts of minipools per status
type MinipoolCountsPerStatus struct {
	Initialized  *big.Int `abi:"initialisedCount"`
	Prelaunch    *big.Int `abi:"prelaunchCount"`
	Staking      *big.Int `abi:"stakingCount"`
	Withdrawable *big.Int `abi:"withdrawableCount"`
	Dissolved    *big.Int `abi:"dissolvedCount"`
}

// Get all minipool details
func GetMinipools(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]MinipoolDetails, error) {
	minipoolAddresses, err := GetMinipoolAddresses(rp, opts)
	if err != nil {
		return []MinipoolDetails{}, err
	}
	return loadMinipoolDetails(rp, minipoolAddresses, opts)
}

// Get a node's minipool details
func GetNodeMinipools(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) ([]MinipoolDetails, error) {
	minipoolAddresses, err := GetNodeMinipoolAddresses(rp, nodeAddress, opts)
	if err != nil {
		return []MinipoolDetails{}, err
	}
	return loadMinipoolDetails(rp, minipoolAddresses, opts)
}

// Load minipool details
func loadMinipoolDetails(rp *rocketpool.RocketPool, minipoolAddresses []common.Address, opts *bind.CallOpts) ([]MinipoolDetails, error) {

	// Load minipool details in batches
	details := make([]MinipoolDetails, len(minipoolAddresses))
	for bsi := 0; bsi < len(minipoolAddresses); bsi += MinipoolDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolDetailsBatchSize
		if mei > len(minipoolAddresses) {
			mei = len(minipoolAddresses)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				minipoolAddress := minipoolAddresses[mi]
				minipoolDetails, err := GetMinipoolDetails(rp, minipoolAddress, opts)
				if err == nil {
					details[mi] = minipoolDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []MinipoolDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get all minipool addresses
func GetMinipoolAddresses(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]common.Address, error) {

	// Get minipool count
	minipoolCount, err := GetMinipoolCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Load minipool addresses in batches
	addresses := make([]common.Address, minipoolCount)
	for bsi := uint64(0); bsi < minipoolCount; bsi += MinipoolAddressBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolAddressBatchSize
		if mei > minipoolCount {
			mei = minipoolCount
		}

		// Load addresses
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address, err := GetMinipoolAt(rp, mi, opts)
				if err == nil {
					addresses[mi] = address
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []common.Address{}, err
		}

	}

	// Return
	return addresses, nil

}

// Get the addresses of all minipools in prelaunch status
func GetPrelaunchMinipoolAddresses(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]common.Address, error) {

	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Get the total number of minipools
	totalMinipoolsUint, err := GetMinipoolCount(rp, nil)
	if err != nil {
		return []common.Address{}, err
	}

	totalMinipools := int64(totalMinipoolsUint)
	addresses := []common.Address{}
	limit := big.NewInt(MinipoolPrelaunchBatchSize)
	for i := int64(0); i < totalMinipools; i += MinipoolPrelaunchBatchSize {
		// Get a batch of addresses
		offset := big.NewInt(i)
		newAddresses := new([]common.Address)
		if err := rocketMinipoolManager.Call(opts, newAddresses, "getPrelaunchMinipools", offset, limit); err != nil {
			return []common.Address{}, fmt.Errorf("error getting prelaunch minipool addresses: %w", err)
		}
		addresses = append(addresses, *newAddresses...)
	}

	return addresses, nil
}

// Get a node's minipool addresses
func GetNodeMinipoolAddresses(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) ([]common.Address, error) {

	// Get minipool count
	minipoolCount, err := GetNodeMinipoolCount(rp, nodeAddress, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Load minipool addresses in batches
	addresses := make([]common.Address, minipoolCount)
	for bsi := uint64(0); bsi < minipoolCount; bsi += MinipoolAddressBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolAddressBatchSize
		if mei > minipoolCount {
			mei = minipoolCount
		}

		// Load addresses
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address, err := GetNodeMinipoolAt(rp, nodeAddress, mi, opts)
				if err == nil {
					addresses[mi] = address
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []common.Address{}, err
		}

	}

	// Return
	return addresses, nil

}

// Get a node's validating minipool pubkeys
func GetNodeValidatingMinipoolPubkeys(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) ([]rptypes.ValidatorPubkey, error) {

	// Get minipool count
	minipoolCount, err := GetNodeValidatingMinipoolCount(rp, nodeAddress, opts)
	if err != nil {
		return []rptypes.ValidatorPubkey{}, err
	}

	// Load pubkeys in batches
	var lock = sync.RWMutex{}
	pubkeys := make([]rptypes.ValidatorPubkey, minipoolCount)
	for bsi := uint64(0); bsi < minipoolCount; bsi += MinipoolAddressBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolAddressBatchSize
		if mei > minipoolCount {
			mei = minipoolCount
		}

		// Load pubkeys
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				minipoolAddress, err := GetNodeValidatingMinipoolAt(rp, nodeAddress, mi, opts)
				if err != nil {
					return err
				}
				pubkey, err := GetMinipoolPubkey(rp, minipoolAddress, opts)
				if err != nil {
					return err
				}
				lock.Lock()
				pubkeys[mi] = pubkey
				lock.Unlock()
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return []rptypes.ValidatorPubkey{}, err
		}

	}

	// Return
	return pubkeys, nil

}

// Get a minipool's details
func GetMinipoolDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (MinipoolDetails, error) {

	// Data
	var wg errgroup.Group
	var exists bool
	var pubkey rptypes.ValidatorPubkey

	// Load data
	wg.Go(func() error {
		var err error
		exists, err = GetMinipoolExists(rp, minipoolAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		pubkey, err = GetMinipoolPubkey(rp, minipoolAddress, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return MinipoolDetails{}, err
	}

	// Return
	return MinipoolDetails{
		Address: minipoolAddress,
		Exists:  exists,
		Pubkey:  pubkey,
	}, nil

}

// Get the minipool count
func GetMinipoolCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getMinipoolCount"); err != nil {
		return 0, fmt.Errorf("error getting minipool count: %w", err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get the number of staking minipools in the network
func GetStakingMinipoolCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getStakingMinipoolCount"); err != nil {
		return 0, fmt.Errorf("error getting staking minipool count: %w", err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get the number of finalised minipools in the network
func GetFinalisedMinipoolCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getFinalisedMinipoolCount"); err != nil {
		return 0, fmt.Errorf("error getting finalised minipool count: %w", err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get the number of active minipools in the network
func GetActiveMinipoolCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getActiveMinipoolCount"); err != nil {
		return 0, fmt.Errorf("error getting finalised minipool count: %w", err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get the minipool count by status
func GetMinipoolCountPerStatus(rp *rocketpool.RocketPool, opts *bind.CallOpts) (MinipoolCountsPerStatus, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return MinipoolCountsPerStatus{}, err
	}

	// Get the total number of minipools
	totalMinipoolsUint, err := GetMinipoolCount(rp, nil)
	if err != nil {
		return MinipoolCountsPerStatus{}, err
	}

	totalMinipools := int64(totalMinipoolsUint)
	minipoolCounts := MinipoolCountsPerStatus{
		Initialized:  big.NewInt(0),
		Prelaunch:    big.NewInt(0),
		Staking:      big.NewInt(0),
		Dissolved:    big.NewInt(0),
		Withdrawable: big.NewInt(0),
	}
	limit := big.NewInt(MinipoolPrelaunchBatchSize)
	for i := int64(0); i < totalMinipools; i += MinipoolPrelaunchBatchSize {
		// Get a batch of counts
		offset := big.NewInt(i)
		newMinipoolCounts := new(MinipoolCountsPerStatus)
		if err := rocketMinipoolManager.Call(opts, newMinipoolCounts, "getMinipoolCountPerStatus", offset, limit); err != nil {
			return MinipoolCountsPerStatus{}, fmt.Errorf("error getting minipool counts: %w", err)
		}
		if newMinipoolCounts != nil {
			if newMinipoolCounts.Initialized != nil {
				minipoolCounts.Initialized.Add(minipoolCounts.Initialized, newMinipoolCounts.Initialized)
			}
			if newMinipoolCounts.Prelaunch != nil {
				minipoolCounts.Prelaunch.Add(minipoolCounts.Prelaunch, newMinipoolCounts.Prelaunch)
			}
			if newMinipoolCounts.Staking != nil {
				minipoolCounts.Staking.Add(minipoolCounts.Staking, newMinipoolCounts.Staking)
			}
			if newMinipoolCounts.Dissolved != nil {
				minipoolCounts.Dissolved.Add(minipoolCounts.Dissolved, newMinipoolCounts.Dissolved)
			}
			if newMinipoolCounts.Withdrawable != nil {
				minipoolCounts.Withdrawable.Add(minipoolCounts.Withdrawable, newMinipoolCounts.Withdrawable)
			}
		}
	}
	return minipoolCounts, nil
}

// Get a minipool address by index
func GetMinipoolAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	minipoolAddress := new(common.Address)
	if err := rocketMinipoolManager.Call(opts, minipoolAddress, "getMinipoolAt", big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("error getting minipool %d address: %w", index, err)
	}
	return *minipoolAddress, nil
}

// Get a node's minipool count
func GetNodeMinipoolCount(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getNodeMinipoolCount", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node %s minipool count: %w", nodeAddress.Hex(), err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get a node's minipool count
func GetNodeMinipoolCountRaw(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return nil, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getNodeMinipoolCount", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node %s minipool count: %w", nodeAddress.Hex(), err)
	}
	return *minipoolCount, nil
}

// Get the number of minipools owned by a node that are not finalised
func GetNodeActiveMinipoolCount(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getNodeActiveMinipoolCount", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node %s minipool count: %w", nodeAddress.Hex(), err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get the number of minipools owned by a node that are finalised
func GetNodeFinalisedMinipoolCount(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getNodeFinalisedMinipoolCount", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node %s minipool count: %w", nodeAddress.Hex(), err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get a node's minipool address by index
func GetNodeMinipoolAt(rp *rocketpool.RocketPool, nodeAddress common.Address, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	minipoolAddress := new(common.Address)
	if err := rocketMinipoolManager.Call(opts, minipoolAddress, "getNodeMinipoolAt", nodeAddress, big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("error getting node %s minipool %d address: %w", nodeAddress.Hex(), index, err)
	}
	return *minipoolAddress, nil
}

// Get a node's validating minipool count
func GetNodeValidatingMinipoolCount(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, minipoolCount, "getNodeValidatingMinipoolCount", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node %s validating minipool count: %w", nodeAddress.Hex(), err)
	}
	return (*minipoolCount).Uint64(), nil
}

// Get a node's validating minipool address by index
func GetNodeValidatingMinipoolAt(rp *rocketpool.RocketPool, nodeAddress common.Address, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	minipoolAddress := new(common.Address)
	if err := rocketMinipoolManager.Call(opts, minipoolAddress, "getNodeValidatingMinipoolAt", nodeAddress, big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("error getting node %s validating minipool %d address: %w", nodeAddress.Hex(), index, err)
	}
	return *minipoolAddress, nil
}

// Get a minipool address by validator pubkey
func GetMinipoolByPubkey(rp *rocketpool.RocketPool, pubkey rptypes.ValidatorPubkey, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	minipoolAddress := new(common.Address)
	if err := rocketMinipoolManager.Call(opts, minipoolAddress, "getMinipoolByPubkey", pubkey[:]); err != nil {
		return common.Address{}, fmt.Errorf("error getting validator %s minipool address: %w", pubkey.Hex(), err)
	}
	return *minipoolAddress, nil
}

// Check whether a minipool exists
func GetMinipoolExists(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return false, err
	}
	exists := new(bool)
	if err := rocketMinipoolManager.Call(opts, exists, "getMinipoolExists", minipoolAddress); err != nil {
		return false, fmt.Errorf("error getting minipool %s exists status: %w", minipoolAddress.Hex(), err)
	}
	return *exists, nil
}

// Get a minipool's validator pubkey
func GetMinipoolPubkey(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (rptypes.ValidatorPubkey, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return rptypes.ValidatorPubkey{}, err
	}
	pubkey := new(rptypes.ValidatorPubkey)
	if err := rocketMinipoolManager.Call(opts, pubkey, "getMinipoolPubkey", minipoolAddress); err != nil {
		return rptypes.ValidatorPubkey{}, fmt.Errorf("error getting minipool %s pubkey: %w", minipoolAddress.Hex(), err)
	}
	return *pubkey, nil
}

// Get the 0x01-based Beacon Chain withdrawal credentials for a given minipool
func GetMinipoolWithdrawalCredentials(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (common.Hash, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return common.Hash{}, err
	}
	withdrawalCredentials := new(common.Hash)
	if err := rocketMinipoolManager.Call(opts, withdrawalCredentials, "getMinipoolWithdrawalCredentials", minipoolAddress); err != nil {
		return common.Hash{}, fmt.Errorf("error getting minipool withdrawal credentials: %w", err)
	}
	return *withdrawalCredentials, nil
}

// Get the number of penalties applied to a minipool
func GetMinipoolPenaltyCount(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	key := crypto.Keccak256Hash([]byte("network.penalties.penalty"), minipoolAddress.Bytes())
	penalties, err := rp.RocketStorage.GetUint(opts, key)
	if err != nil {
		return 0, err
	}
	return penalties.Uint64(), nil
}

// Get the vacant minipool count
func GetVacantMinipoolCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	vacantMinipoolCount := new(*big.Int)
	if err := rocketMinipoolManager.Call(opts, vacantMinipoolCount, "getVacantMinipoolCount"); err != nil {
		return 0, fmt.Errorf("error getting vacant minipool count: %w", err)
	}
	return (*vacantMinipoolCount).Uint64(), nil
}

// Get a vacant minipool address by index
func GetVacantMinipoolAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	vacantMinipoolAddress := new(common.Address)
	if err := rocketMinipoolManager.Call(opts, vacantMinipoolAddress, "getVacantMinipoolAt", big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("error getting vacant minipool %d address: %w", index, err)
	}
	return *vacantMinipoolAddress, nil
}

// Get a minipool's RPL slashing status
func GetMinipoolRPLSlashed(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketMinipoolManager.Call(opts, value, "getMinipoolRPLSlashed", minipoolAddress); err != nil {
		return false, fmt.Errorf("error getting minipool %s slashed status: %w", minipoolAddress.Hex(), err)
	}
	return *value, nil
}

// Get a minipool's deposit type invariant of its delegate version
func GetMinipoolDepositType(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (types.MinipoolDeposit, error) {
	rocketMinipoolManager, err := getRocketMinipoolManager(rp, opts)
	if err != nil {
		return types.None, err
	}
	value := new(uint8)
	if err := rocketMinipoolManager.Call(opts, value, "getMinipoolDepositType", minipoolAddress); err != nil {
		return types.None, fmt.Errorf("error getting minipool %s slashed status: %w", minipoolAddress.Hex(), err)
	}
	return types.MinipoolDeposit(*value), nil
}

// Get contracts
var rocketMinipoolManagerLock sync.Mutex

func getRocketMinipoolManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolManagerLock.Lock()
	defer rocketMinipoolManagerLock.Unlock()
	return rp.GetContract("rocketMinipoolManager", opts)
}
