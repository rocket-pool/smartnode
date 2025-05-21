package node

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/storage"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	"github.com/rocket-pool/smartnode/bindings/utils/strings"
)

// Settings
const (
	nodeAddressFastBatchSize    int    = 1000
	NodeAddressBatchSize               = 50
	NodeDetailsBatchSize               = 20
	SmoothingPoolCountBatchSize uint64 = 2000
	NativeNodeDetailsBatchSize         = 10000
)

// Node details
type NodeDetails struct {
	Address                         common.Address `json:"address"`
	Exists                          bool           `json:"exists"`
	PrimaryWithdrawalAddress        common.Address `json:"primaryWithdrawalAddress"`
	PendingPrimaryWithdrawalAddress common.Address `json:"pendingPrimaryWithdrawalAddress"`
	IsRPLWithdrawalAddressSet       bool           `json:"isRPLWithdrawalAddressSet"`
	RPLWithdrawalAddress            common.Address `json:"rplWithdrawalAddress"`
	PendingRPLWithdrawalAddress     common.Address `json:"pendingRPLWithdrawalAddress"`
	TimezoneLocation                string         `json:"timezoneLocation"`
}

// Count of nodes belonging to a timezone
type TimezoneCount struct {
	Timezone string   `abi:"timezone"`
	Count    *big.Int `abi:"count"`
}

// The results of the trusted node participation calculation
type TrustedNodeParticipation struct {
	StartBlock          uint64
	UpdateFrequency     uint64
	UpdateCount         uint64
	Probability         float64
	ExpectedSubmissions float64
	ActualSubmissions   map[common.Address]float64
	Participation       map[common.Address][]bool
}

// Get the version of the Node Manager contract
func GetNodeManagerVersion(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint8, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return 0, err
	}
	return rocketpool.GetContractVersion(rp, *rocketNodeManager.Address, opts)
}

// Get all node details
// The 'includeRplWithdrawalAddress' flag is used for backwards compatibility with Atlas, - set it to `false` if Houston hasn't been deployed yet
func GetNodes(rp *rocketpool.RocketPool, includeRplWithdrawalAddress bool, opts *bind.CallOpts) ([]NodeDetails, error) {

	// Get node addresses
	nodeAddresses, err := GetNodeAddresses(rp, opts)
	if err != nil {
		return []NodeDetails{}, err
	}

	// Load node details in batches
	details := make([]NodeDetails, len(nodeAddresses))
	for bsi := 0; bsi < len(nodeAddresses); bsi += NodeDetailsBatchSize {

		// Get batch start & end index
		nsi := bsi
		nei := bsi + NodeDetailsBatchSize
		if nei > len(nodeAddresses) {
			nei = len(nodeAddresses)
		}

		// Load details
		var wg errgroup.Group
		for ni := nsi; ni < nei; ni++ {
			ni := ni
			wg.Go(func() error {
				nodeAddress := nodeAddresses[ni]
				nodeDetails, err := GetNodeDetails(rp, nodeAddress, includeRplWithdrawalAddress, opts)
				if err == nil {
					details[ni] = nodeDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []NodeDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get all node addresses
func GetNodeAddresses(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]common.Address, error) {

	// Get node count
	nodeCount, err := GetNodeCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Load node addresses in batches
	addresses := make([]common.Address, nodeCount)
	for bsi := uint64(0); bsi < nodeCount; bsi += NodeAddressBatchSize {

		// Get batch start & end index
		nsi := bsi
		nei := bsi + NodeAddressBatchSize
		if nei > nodeCount {
			nei = nodeCount
		}

		// Load addresses
		var wg errgroup.Group
		for ni := nsi; ni < nei; ni++ {
			ni := ni
			wg.Go(func() error {
				address, err := GetNodeAt(rp, ni, opts)
				if err == nil {
					addresses[ni] = address
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

// Get all node addresses using a multicaller
func GetNodeAddressesFast(rp *rocketpool.RocketPool, multicallAddress common.Address, opts *bind.CallOpts) ([]common.Address, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return nil, err
	}

	// Get minipool count
	nodeCount, err := GetNodeCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	addresses := make([]common.Address, nodeCount)

	// Run the getters in batches
	count := int(nodeCount)
	for i := 0; i < count; i += nodeAddressFastBatchSize {
		i := i
		max := i + nodeAddressFastBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(rocketNodeManager, &addresses[j], "getNodeAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	return addresses, nil
}

// Get a node's details
// The 'includeRplWithdrawalAddress' flag is used for backwards compatibility with Atlas, - set it to `false` if Houston hasn't been deployed yet
func GetNodeDetails(rp *rocketpool.RocketPool, nodeAddress common.Address, includeRplWithdrawalAddress bool, opts *bind.CallOpts) (NodeDetails, error) {

	// Data
	var wg errgroup.Group
	var exists bool
	var primaryWithdrawalAddress common.Address
	var pendingPrimaryWithdrawalAddress common.Address
	var isRPLWithdrawalAddressSet bool
	var rplWithdrawalAddress common.Address
	var pendingRPLWithdrawalAddress common.Address
	var timezoneLocation string

	// Load data
	wg.Go(func() error {
		var err error
		exists, err = GetNodeExists(rp, nodeAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		primaryWithdrawalAddress, err = storage.GetNodeWithdrawalAddress(rp, nodeAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		pendingPrimaryWithdrawalAddress, err = storage.GetNodePendingWithdrawalAddress(rp, nodeAddress, opts)
		return err
	})
	if includeRplWithdrawalAddress {
		wg.Go(func() error {
			var err error
			isRPLWithdrawalAddressSet, err = GetNodeRPLWithdrawalAddressIsSet(rp, nodeAddress, opts)
			return err
		})
		wg.Go(func() error {
			var err error
			rplWithdrawalAddress, err = GetNodeRPLWithdrawalAddress(rp, nodeAddress, opts)
			return err
		})
		wg.Go(func() error {
			var err error
			pendingRPLWithdrawalAddress, err = GetNodePendingRPLWithdrawalAddress(rp, nodeAddress, opts)
			return err
		})
	}
	wg.Go(func() error {
		var err error
		timezoneLocation, err = GetNodeTimezoneLocation(rp, nodeAddress, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return NodeDetails{}, err
	}

	// Return
	return NodeDetails{
		Address:                         nodeAddress,
		Exists:                          exists,
		PrimaryWithdrawalAddress:        primaryWithdrawalAddress,
		PendingPrimaryWithdrawalAddress: pendingPrimaryWithdrawalAddress,
		IsRPLWithdrawalAddressSet:       isRPLWithdrawalAddressSet,
		RPLWithdrawalAddress:            rplWithdrawalAddress,
		PendingRPLWithdrawalAddress:     pendingRPLWithdrawalAddress,
		TimezoneLocation:                timezoneLocation,
	}, nil

}

// Get the number of nodes in the network
func GetNodeCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return 0, err
	}
	nodeCount := new(*big.Int)
	if err := rocketNodeManager.Call(opts, nodeCount, "getNodeCount"); err != nil {
		return 0, fmt.Errorf("error getting node count: %w", err)
	}
	return (*nodeCount).Uint64(), nil
}

// Get a breakdown of the number of nodes per timezone
func GetNodeCountPerTimezone(rp *rocketpool.RocketPool, offset, limit *big.Int, opts *bind.CallOpts) ([]TimezoneCount, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return []TimezoneCount{}, err
	}
	timezoneCounts := new([]TimezoneCount)
	if err := rocketNodeManager.Call(opts, timezoneCounts, "getNodeCountPerTimezone", offset, limit); err != nil {
		return []TimezoneCount{}, fmt.Errorf("error getting node count: %w", err)
	}
	return *timezoneCounts, nil
}

// Get a node address by index
func GetNodeAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	nodeAddress := new(common.Address)
	if err := rocketNodeManager.Call(opts, nodeAddress, "getNodeAt", big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("error getting node %d address: %w", index, err)
	}
	return *nodeAddress, nil
}

// Check whether a node exists
func GetNodeExists(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return false, err
	}
	exists := new(bool)
	if err := rocketNodeManager.Call(opts, exists, "getNodeExists", nodeAddress); err != nil {
		return false, fmt.Errorf("error getting node %s exists status: %w", nodeAddress.Hex(), err)
	}
	return *exists, nil
}

// Get a node's timezone location
func GetNodeTimezoneLocation(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (string, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return "", err
	}
	timezoneLocation := new(string)
	if err := rocketNodeManager.Call(opts, timezoneLocation, "getNodeTimezoneLocation", nodeAddress); err != nil {
		return "", fmt.Errorf("error getting node %s timezone location: %w", nodeAddress.Hex(), err)
	}
	return strings.Sanitize(*timezoneLocation), nil
}

// Estimate the gas of RegisterNode
func EstimateRegisterNodeGas(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	_, err = time.LoadLocation(timezoneLocation)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error verifying timezone [%s]: %w", timezoneLocation, err)
	}
	return rocketNodeManager.GetTransactionGasInfo(opts, "registerNode", timezoneLocation)
}

// Register a node
func RegisterNode(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	_, err = time.LoadLocation(timezoneLocation)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error verifying timezone [%s]: %w", timezoneLocation, err)
	}
	tx, err := rocketNodeManager.Transact(opts, "registerNode", timezoneLocation)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error registering node: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of SetTimezoneLocation
func EstimateSetTimezoneLocationGas(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	_, err = time.LoadLocation(timezoneLocation)
	if err != nil {
		return rocketpool.GasInfo{}, fmt.Errorf("error verifying timezone [%s]: %w", timezoneLocation, err)
	}
	return rocketNodeManager.GetTransactionGasInfo(opts, "setTimezoneLocation", timezoneLocation)
}

// Set a node's timezone location
func SetTimezoneLocation(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	_, err = time.LoadLocation(timezoneLocation)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error verifying timezone [%s]: %w", timezoneLocation, err)
	}
	tx, err := rocketNodeManager.Transact(opts, "setTimezoneLocation", timezoneLocation)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting node timezone location: %w", err)
	}
	return tx.Hash(), nil
}

// Get the network ID for a node's rewards
func GetRewardNetwork(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return 0, err
	}
	rewardNetwork := new(*big.Int)
	if err := rocketNodeManager.Call(opts, rewardNetwork, "getRewardNetwork", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node %s reward network: %w", nodeAddress.Hex(), err)
	}
	return (*rewardNetwork).Uint64(), nil
}

// Get the network ID for a node's rewards
func GetRewardNetworkRaw(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return nil, err
	}
	rewardNetwork := new(*big.Int)
	if err := rocketNodeManager.Call(opts, rewardNetwork, "getRewardNetwork", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node %s reward network: %w", nodeAddress.Hex(), err)
	}
	return *rewardNetwork, nil
}

// Check if a node's fee distributor has been initialized yet
func GetFeeDistributorInitialized(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return false, err
	}
	isInitialized := new(bool)
	if err := rocketNodeManager.Call(opts, isInitialized, "getFeeDistributorInitialised", nodeAddress); err != nil {
		return false, fmt.Errorf("error checking if node %s's fee distributor is initialized: %w", nodeAddress.Hex(), err)
	}
	return *isInitialized, nil
}

// Estimate the gas for creating the fee distributor contract for a node
func EstimateInitializeFeeDistributorGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeManager.GetTransactionGasInfo(opts, "initialiseFeeDistributor")
}

// Create the fee distributor contract for a node
func InitializeFeeDistributor(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeManager.Transact(opts, "initialiseFeeDistributor")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error initializing fee distributor: %w", err)
	}
	return tx.Hash(), nil
}

// Get a node's average minipool fee
func GetNodeAverageFee(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (float64, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return 0, err
	}
	avgFee := new(*big.Int)
	if err := rocketNodeManager.Call(opts, avgFee, "getAverageNodeFee", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node %s average fee: %w", nodeAddress.Hex(), err)
	}
	return eth.WeiToEth(*avgFee), nil
}

// Get a node's average minipool fee
func GetNodeAverageFeeRaw(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return nil, err
	}
	avgFee := new(*big.Int)
	if err := rocketNodeManager.Call(opts, avgFee, "getAverageNodeFee", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node %s average fee: %w", nodeAddress.Hex(), err)
	}
	return *avgFee, nil
}

// Get the time that the user registered as a claimer
func GetNodeRegistrationTime(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (time.Time, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	registrationTime := new(*big.Int)
	if err := rocketNodeManager.Call(opts, registrationTime, "getNodeRegistrationTime", address); err != nil {
		return time.Time{}, fmt.Errorf("error getting registration time for %s: %w", address.Hex(), err)
	}
	return time.Unix((*registrationTime).Int64(), 0), nil
}

// Get the time that the user registered as a claimer
func GetNodeRegistrationTimeRaw(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return nil, err
	}
	registrationTime := new(*big.Int)
	if err := rocketNodeManager.Call(opts, registrationTime, "getNodeRegistrationTime", address); err != nil {
		return nil, fmt.Errorf("error getting registration time for %s: %w", address.Hex(), err)
	}
	return *registrationTime, nil
}

// Get the smoothing pool opt-in status of a node
func GetSmoothingPoolRegistrationState(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return false, err
	}
	state := new(bool)
	if err := rocketNodeManager.Call(opts, state, "getSmoothingPoolRegistrationState", nodeAddress); err != nil {
		return false, fmt.Errorf("error getting node %s smoothing pool registration status: %w", nodeAddress.Hex(), err)
	}
	return *state, nil
}

// Get the time of the previous smoothing pool opt-in / opt-out
func GetSmoothingPoolRegistrationChanged(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (time.Time, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	timestamp := new(*big.Int)
	if err := rocketNodeManager.Call(opts, timestamp, "getSmoothingPoolRegistrationChanged", nodeAddress); err != nil {
		return time.Time{}, fmt.Errorf("error getting node %s's last smoothing pool registration change time: %w", nodeAddress.Hex(), err)
	}
	return time.Unix((*timestamp).Int64(), 0), nil
}

// Get the time of the previous smoothing pool opt-in / opt-out
func GetSmoothingPoolRegistrationChangedRaw(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return nil, err
	}
	timestamp := new(*big.Int)
	if err := rocketNodeManager.Call(opts, timestamp, "getSmoothingPoolRegistrationChanged", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node %s's last smoothing pool registration change time: %w", nodeAddress.Hex(), err)
	}
	return *timestamp, nil
}

// Estimate the gas for opting into / out of the smoothing pool
func EstimateSetSmoothingPoolRegistrationStateGas(rp *rocketpool.RocketPool, optIn bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeManager.GetTransactionGasInfo(opts, "setSmoothingPoolRegistrationState", optIn)
}

// Opt into / out of the smoothing pool
func SetSmoothingPoolRegistrationState(rp *rocketpool.RocketPool, optIn bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeManager.Transact(opts, "setSmoothingPoolRegistrationState", optIn)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting smoothing pool registration state: %w", err)
	}
	return tx.Hash(), nil
}

// Get the number of nodes in the Smoothing Pool
func GetSmoothingPoolRegisteredNodeCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return 0, err
	}

	// Get the number of nodes
	nodeCount, err := GetNodeCount(rp, opts)
	if err != nil {
		return 0, err
	}

	iterations := uint64(math.Ceil(float64(nodeCount) / float64(SmoothingPoolCountBatchSize)))
	iterationCounts := make([]*big.Int, iterations)

	// Load addresses
	var wg errgroup.Group
	for i := uint64(0); i < iterations; i++ {
		i := i
		offset := i * SmoothingPoolCountBatchSize
		limit := SmoothingPoolCountBatchSize
		if nodeCount-offset < SmoothingPoolCountBatchSize {
			limit = nodeCount - offset
		}
		wg.Go(func() error {
			count := new(*big.Int)
			err := rocketNodeManager.Call(opts, count, "getSmoothingPoolRegisteredNodeCount", big.NewInt(int64(offset)), big.NewInt(int64(limit)))
			if err != nil {
				return fmt.Errorf("error getting smoothing pool opt-in count for batch starting at %d: %w", offset, err)
			}

			iterationCounts[i] = *count
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return 0, err
	}

	total := uint64(0)
	for _, count := range iterationCounts {
		total += count.Uint64()
	}

	return total, nil

}

// Check if the RPL-specific withdrawal address has been set
func GetNodeRPLWithdrawalAddressIsSet(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketNodeManager.Call(opts, value, "getNodeRPLWithdrawalAddressIsSet", nodeAddress); err != nil {
		return false, fmt.Errorf("error getting node %s's RPL withdrawal address status: %w", nodeAddress.Hex(), err)
	}
	return *value, nil
}

// Get the RPL-specific withdrawal address
func GetNodeRPLWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	value := new(common.Address)
	if err := rocketNodeManager.Call(opts, value, "getNodeRPLWithdrawalAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("error getting node %s's RPL withdrawal address: %w", nodeAddress.Hex(), err)
	}
	return *value, nil
}

// Get the pending RPL-specific withdrawal address
func GetNodePendingRPLWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	value := new(common.Address)
	if err := rocketNodeManager.Call(opts, value, "getNodePendingRPLWithdrawalAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("error getting node %s's pending RPL withdrawal address: %w", nodeAddress.Hex(), err)
	}
	return *value, nil
}

// Estimate the gas for setting the RPL-specific withdrawal address
func EstimateSetRPLWithdrawalAddressGas(rp *rocketpool.RocketPool, nodeAddress common.Address, withdrawalAddress common.Address, confirm bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeManager.GetTransactionGasInfo(opts, "setRPLWithdrawalAddress", nodeAddress, withdrawalAddress, confirm)
}

// Set the RPL-specific withdrawal address
func SetRPLWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, withdrawalAddress common.Address, confirm bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeManager.Transact(opts, "setRPLWithdrawalAddress", nodeAddress, withdrawalAddress, confirm)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting RPL withdrawal address: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas for confirming the RPL-specific withdrawal address
func EstimateConfirmRPLWithdrawalAddressGas(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeManager.GetTransactionGasInfo(opts, "confirmRPLWithdrawalAddress", nodeAddress)
}

// Confirm the RPL-specific withdrawal address
func ConfirmRPLWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeManager.Transact(opts, "confirmRPLWithdrawalAddress", nodeAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error confirming RPL withdrawal address: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketNodeManagerLock sync.Mutex

func getRocketNodeManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeManagerLock.Lock()
	defer rocketNodeManagerLock.Unlock()
	return rp.GetContract("rocketNodeManager", opts)
}

var rocketNetworkPricesLock sync.Mutex

func getRocketNetworkPrices(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkPricesLock.Lock()
	defer rocketNetworkPricesLock.Unlock()
	return rp.GetContract("rocketNetworkPrices", opts)
}

var rocketNetworkBalancesLock sync.Mutex

func getRocketNetworkBalances(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkBalancesLock.Lock()
	defer rocketNetworkBalancesLock.Unlock()
	return rp.GetContract("rocketNetworkBalances", opts)
}

var rocketDAONodeTrustedActionsLock sync.Mutex

func getRocketDAONodeTrustedActions(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAONodeTrustedActionsLock.Lock()
	defer rocketDAONodeTrustedActionsLock.Unlock()
	return rp.GetContract("rocketDAONodeTrustedActions", opts)
}
