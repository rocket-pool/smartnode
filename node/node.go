package node

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"gonum.org/v1/gonum/mathext"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/storage"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/rocketpool-go/utils/strings"
)

// Settings
const (
	NodeAddressBatchSize               = 50
	NodeDetailsBatchSize               = 20
	SmoothingPoolCountBatchSize uint64 = 2000
	NativeNodeDetailsBatchSize         = 10000
)

// Node details
type NodeDetails struct {
	Address                  common.Address `json:"address"`
	Exists                   bool           `json:"exists"`
	WithdrawalAddress        common.Address `json:"withdrawalAddress"`
	PendingWithdrawalAddress common.Address `json:"pendingWithdrawalAddress"`
	TimezoneLocation         string         `json:"timezoneLocation"`
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
func GetNodes(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]NodeDetails, error) {

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
				nodeDetails, err := GetNodeDetails(rp, nodeAddress, opts)
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

// Get a node's details
func GetNodeDetails(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (NodeDetails, error) {

	// Data
	var wg errgroup.Group
	var exists bool
	var withdrawalAddress common.Address
	var pendingWithdrawalAddress common.Address
	var timezoneLocation string

	// Load data
	wg.Go(func() error {
		var err error
		exists, err = GetNodeExists(rp, nodeAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		withdrawalAddress, err = storage.GetNodeWithdrawalAddress(rp, nodeAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		pendingWithdrawalAddress, err = storage.GetNodePendingWithdrawalAddress(rp, nodeAddress, opts)
		return err
	})
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
		Address:                  nodeAddress,
		Exists:                   exists,
		WithdrawalAddress:        withdrawalAddress,
		PendingWithdrawalAddress: pendingWithdrawalAddress,
		TimezoneLocation:         timezoneLocation,
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
		return 0, fmt.Errorf("Could not get node count: %w", err)
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
		return []TimezoneCount{}, fmt.Errorf("Could not get node count: %w", err)
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
		return common.Address{}, fmt.Errorf("Could not get node %d address: %w", index, err)
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
		return false, fmt.Errorf("Could not get node %s exists status: %w", nodeAddress.Hex(), err)
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
		return "", fmt.Errorf("Could not get node %s timezone location: %w", nodeAddress.Hex(), err)
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
		return rocketpool.GasInfo{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
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
		return common.Hash{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
	}
	tx, err := rocketNodeManager.Transact(opts, "registerNode", timezoneLocation)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not register node: %w", err)
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
		return rocketpool.GasInfo{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
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
		return common.Hash{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
	}
	tx, err := rocketNodeManager.Transact(opts, "setTimezoneLocation", timezoneLocation)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not set node timezone location: %w", err)
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
		return 0, fmt.Errorf("Could not get node %s reward network: %w", nodeAddress.Hex(), err)
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
		return nil, fmt.Errorf("Could not get node %s reward network: %w", nodeAddress.Hex(), err)
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
		return false, fmt.Errorf("Could not check if node %s's fee distributor is initialized: %w", nodeAddress.Hex(), err)
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
		return common.Hash{}, fmt.Errorf("Could not initialize fee distributor: %w", err)
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
		return 0, fmt.Errorf("Could not get node %s average fee: %w", nodeAddress.Hex(), err)
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
		return nil, fmt.Errorf("Could not get node %s average fee: %w", nodeAddress.Hex(), err)
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
		return time.Time{}, fmt.Errorf("Could not get registration time for %s: %w", address.Hex(), err)
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
		return nil, fmt.Errorf("Could not get registration time for %s: %w", address.Hex(), err)
	}
	return *registrationTime, nil
}

// Returns an array of block numbers for prices submissions the given trusted node has submitted since fromBlock
func GetPricesSubmissions(rp *rocketpool.RocketPool, nodeAddress common.Address, fromBlock uint64, intervalSize *big.Int, opts *bind.CallOpts) (*[]uint64, error) {
	// Get contracts
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, opts)
	if err != nil {
		return nil, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketNetworkPrices.Address}
	topicFilter := [][]common.Hash{{rocketNetworkPrices.ABI.Events["PricesSubmitted"].ID}, {nodeAddress.Hash()}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, big.NewInt(int64(fromBlock)), nil, nil)
	if err != nil {
		return nil, err
	}
	timestamps := make([]uint64, len(logs))
	for i, log := range logs {
		values := make(map[string]interface{})
		// Decode the event
		if rocketNetworkPrices.ABI.Events["PricesSubmitted"].Inputs.UnpackIntoMap(values, log.Data) != nil {
			return nil, err
		}
		timestamps[i] = values["block"].(*big.Int).Uint64()
	}
	return &timestamps, nil
}

// Returns an array of block numbers for balances submissions the given trusted node has submitted since fromBlock
func GetBalancesSubmissions(rp *rocketpool.RocketPool, nodeAddress common.Address, fromBlock uint64, intervalSize *big.Int, opts *bind.CallOpts) (*[]uint64, error) {
	// Get contracts
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return nil, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketNetworkBalances.Address}
	topicFilter := [][]common.Hash{{rocketNetworkBalances.ABI.Events["BalancesSubmitted"].ID}, {nodeAddress.Hash()}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, big.NewInt(int64(fromBlock)), nil, nil)
	if err != nil {
		return nil, err
	}

	timestamps := make([]uint64, len(logs))
	for i, log := range logs {
		values := make(map[string]interface{})
		// Decode the event
		if rocketNetworkBalances.ABI.Events["BalancesSubmitted"].Inputs.UnpackIntoMap(values, log.Data) != nil {
			return nil, err
		}
		timestamps[i] = values["block"].(*big.Int).Uint64()
	}
	return &timestamps, nil
}

// Returns the most recent block number that the number of trusted nodes changed since fromBlock
func getLatestMemberCountChangedBlock(rp *rocketpool.RocketPool, fromBlock uint64, intervalSize *big.Int, opts *bind.CallOpts) (uint64, error) {
	// Get contracts
	rocketDaoNodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, opts)
	if err != nil {
		return 0, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketDaoNodeTrustedActions.Address}
	topicFilter := [][]common.Hash{{rocketDaoNodeTrustedActions.ABI.Events["ActionJoined"].ID, rocketDaoNodeTrustedActions.ABI.Events["ActionLeave"].ID, rocketDaoNodeTrustedActions.ABI.Events["ActionKick"].ID, rocketDaoNodeTrustedActions.ABI.Events["ActionChallengeDecided"].ID}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, big.NewInt(int64(fromBlock)), nil, nil)
	if err != nil {
		return 0, err
	}

	for i := range logs {
		log := logs[len(logs)-i-1]
		if log.Topics[0] == rocketDaoNodeTrustedActions.ABI.Events["ActionChallengeDecided"].ID {
			values := make(map[string]interface{})
			// Decode the event
			if rocketDaoNodeTrustedActions.ABI.Events["ActionChallengeDecided"].Inputs.UnpackIntoMap(values, log.Data) != nil {
				return 0, err
			}
			if values["success"].(bool) {
				return log.BlockNumber, nil
			}
		} else {
			return log.BlockNumber, nil
		}
	}
	return fromBlock, nil
}

// Calculates the participation rate of every trusted node on price submission since the last block that member count changed
func CalculateTrustedNodePricesParticipation(rp *rocketpool.RocketPool, intervalSize *big.Int, opts *bind.CallOpts) (*TrustedNodeParticipation, error) {
	// Get the update frequency
	updatePricesFrequency, err := protocol.GetSubmitPricesFrequency(rp, opts)
	if err != nil {
		return nil, err
	}
	// Get the current block
	currentBlock, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	currentBlockNumber := currentBlock.Number.Uint64()
	// Get the block of the most recent member join (limiting to 50 intervals)
	minBlock := (currentBlockNumber/updatePricesFrequency - 50) * updatePricesFrequency
	latestMemberCountChangedBlock, err := getLatestMemberCountChangedBlock(rp, minBlock, intervalSize, opts)
	if err != nil {
		return nil, err
	}
	// Get the number of current members
	memberCount, err := trustednode.GetMemberCount(rp, nil)
	if err != nil {
		return nil, err
	}
	// Start block is the first interval after the latest join
	startBlock := (latestMemberCountChangedBlock/updatePricesFrequency + 1) * updatePricesFrequency
	// The number of members that have to submit each interval
	consensus := math.Floor(float64(memberCount)/2 + 1)
	// Check if any intervals have passed
	intervalsPassed := uint64(0)
	if currentBlockNumber > startBlock {
		// The number of intervals passed
		intervalsPassed = (currentBlockNumber-startBlock)/updatePricesFrequency + 1
	}
	// How many submissions would we expect per member given a random submission
	expected := float64(intervalsPassed) * consensus / float64(memberCount)
	// Get trusted members
	members, err := trustednode.GetMembers(rp, nil)
	if err != nil {
		return nil, err
	}
	// Construct the epoch map
	participationTable := make(map[common.Address][]bool)
	// Iterate members and sum chi-square
	submissions := make(map[common.Address]float64)
	chi := float64(0)
	for _, member := range members {
		participationTable[member.Address] = make([]bool, intervalsPassed)
		actual := 0
		if intervalsPassed > 0 {
			blocks, err := GetPricesSubmissions(rp, member.Address, startBlock, intervalSize, opts)
			if err != nil {
				return nil, err
			}
			actual = len(*blocks)
			delta := float64(actual) - expected
			chi += (delta * delta) / expected
			// Add to participation table
			for _, block := range *blocks {
				// Ignore out of step updates
				if block%updatePricesFrequency == 0 {
					index := block/updatePricesFrequency - startBlock/updatePricesFrequency
					participationTable[member.Address][index] = true
				}
			}
		}
		// Save actual submission
		submissions[member.Address] = float64(actual)
	}
	// Calculate inverse cumulative density function with members-1 DoF
	probability := float64(1)
	if intervalsPassed > 0 {
		probability = 1 - mathext.GammaIncReg(float64(len(members)-1)/2, chi/2)
	}
	// Construct return value
	participation := TrustedNodeParticipation{
		Probability:         probability,
		ExpectedSubmissions: expected,
		ActualSubmissions:   submissions,
		StartBlock:          startBlock,
		UpdateFrequency:     updatePricesFrequency,
		UpdateCount:         intervalsPassed,
		Participation:       participationTable,
	}
	return &participation, nil
}

// Calculates the participation rate of every trusted node on balance submission since the last block that member count changed
func CalculateTrustedNodeBalancesParticipation(rp *rocketpool.RocketPool, intervalSize *big.Int, opts *bind.CallOpts) (*TrustedNodeParticipation, error) {
	// Get the update frequency
	updateBalancesFrequency, err := protocol.GetSubmitBalancesFrequency(rp, opts)
	if err != nil {
		return nil, err
	}
	// Get the current block
	currentBlock, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	currentBlockNumber := currentBlock.Number.Uint64()
	// Get the block of the most recent member join (limiting to 50 intervals)
	minBlock := (currentBlockNumber/updateBalancesFrequency - 50) * updateBalancesFrequency
	latestMemberCountChangedBlock, err := getLatestMemberCountChangedBlock(rp, minBlock, intervalSize, opts)
	if err != nil {
		return nil, err
	}
	// Get the number of current members
	memberCount, err := trustednode.GetMemberCount(rp, nil)
	if err != nil {
		return nil, err
	}
	// Start block is the first interval after the latest join
	startBlock := (latestMemberCountChangedBlock/updateBalancesFrequency + 1) * updateBalancesFrequency
	// The number of members that have to submit each interval
	consensus := math.Floor(float64(memberCount)/2 + 1)
	// Check if any intervals have passed
	intervalsPassed := uint64(0)
	if currentBlockNumber > startBlock {
		// The number of intervals passed
		intervalsPassed = (currentBlockNumber-startBlock)/updateBalancesFrequency + 1
	}
	// How many submissions would we expect per member given a random submission
	expected := float64(intervalsPassed) * consensus / float64(memberCount)
	// Get trusted members
	members, err := trustednode.GetMembers(rp, nil)
	if err != nil {
		return nil, err
	}
	// Construct the epoch map
	participationTable := make(map[common.Address][]bool)
	// Iterate members and sum chi-square
	submissions := make(map[common.Address]float64)
	chi := float64(0)
	for _, member := range members {
		participationTable[member.Address] = make([]bool, intervalsPassed)
		actual := 0
		if intervalsPassed > 0 {
			blocks, err := GetBalancesSubmissions(rp, member.Address, startBlock, intervalSize, opts)
			if err != nil {
				return nil, err
			}
			actual = len(*blocks)
			delta := float64(actual) - expected
			chi += (delta * delta) / expected
			// Add to participation table
			for _, block := range *blocks {
				// Ignore out of step updates
				if block%updateBalancesFrequency == 0 {
					index := block/updateBalancesFrequency - startBlock/updateBalancesFrequency
					participationTable[member.Address][index] = true
				}
			}
		}
		// Save actual submission
		submissions[member.Address] = float64(actual)
	}
	// Calculate inverse cumulative density function with members-1 DoF
	probability := float64(1)
	if intervalsPassed > 0 {
		probability = 1 - mathext.GammaIncReg(float64(len(members)-1)/2, chi/2)
	}
	// Construct return value
	participation := TrustedNodeParticipation{
		Probability:         probability,
		ExpectedSubmissions: expected,
		ActualSubmissions:   submissions,
		StartBlock:          startBlock,
		UpdateFrequency:     updateBalancesFrequency,
		UpdateCount:         intervalsPassed,
		Participation:       participationTable,
	}
	return &participation, nil
}

// Returns an array of members who submitted a balance since fromBlock
func GetLatestBalancesSubmissions(rp *rocketpool.RocketPool, fromBlock uint64, intervalSize *big.Int, opts *bind.CallOpts) ([]common.Address, error) {
	// Get contracts
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return nil, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketNetworkBalances.Address}
	topicFilter := [][]common.Hash{{rocketNetworkBalances.ABI.Events["BalancesSubmitted"].ID}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, big.NewInt(int64(fromBlock)), nil, nil)
	if err != nil {
		return nil, err
	}

	results := make([]common.Address, len(logs))
	for i, log := range logs {
		// Topic 0 is the event, topic 1 is the "from" address
		address := common.BytesToAddress(log.Topics[1].Bytes())
		results[i] = address
	}
	return results, nil
}

// Returns a mapping of members and whether they have submitted balances this interval or not
func GetTrustedNodeLatestBalancesParticipation(rp *rocketpool.RocketPool, intervalSize *big.Int, opts *bind.CallOpts) (map[common.Address]bool, error) {
	// Get the update frequency
	updateBalancesFrequency, err := protocol.GetSubmitBalancesFrequency(rp, opts)
	if err != nil {
		return nil, err
	}
	// Get the current block
	currentBlock, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	currentBlockNumber := currentBlock.Number.Uint64()
	// Get trusted members
	members, err := trustednode.GetMembers(rp, nil)
	if err != nil {
		return nil, err
	}
	// Get submission within the current interval
	fromBlock := currentBlockNumber / updateBalancesFrequency * updateBalancesFrequency
	submissions, err := GetLatestBalancesSubmissions(rp, fromBlock, intervalSize, opts)
	if err != nil {
		return nil, err
	}
	// Build and return result table
	participationTable := make(map[common.Address]bool)
	for _, member := range members {
		participationTable[member.Address] = false
	}
	for _, submission := range submissions {
		participationTable[submission] = true
	}
	return participationTable, nil
}

// Returns an array of members who submitted prices since fromBlock
func GetLatestPricesSubmissions(rp *rocketpool.RocketPool, fromBlock uint64, intervalSize *big.Int, opts *bind.CallOpts) ([]common.Address, error) {
	// Get contracts
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, opts)
	if err != nil {
		return nil, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketNetworkPrices.Address}
	topicFilter := [][]common.Hash{{rocketNetworkPrices.ABI.Events["PricesSubmitted"].ID}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, big.NewInt(int64(fromBlock)), nil, nil)
	if err != nil {
		return nil, err
	}

	results := make([]common.Address, len(logs))
	for i, log := range logs {
		// Topic 0 is the event, topic 1 is the "from" address
		address := common.BytesToAddress(log.Topics[1].Bytes())
		results[i] = address
	}
	return results, nil
}

// Returns a mapping of members and whether they have submitted prices this interval or not
func GetTrustedNodeLatestPricesParticipation(rp *rocketpool.RocketPool, intervalSize *big.Int, opts *bind.CallOpts) (map[common.Address]bool, error) {
	// Get the update frequency
	updatePricesFrequency, err := protocol.GetSubmitPricesFrequency(rp, opts)
	if err != nil {
		return nil, err
	}
	// Get the current block
	currentBlock, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	currentBlockNumber := currentBlock.Number.Uint64()
	// Get trusted members
	members, err := trustednode.GetMembers(rp, nil)
	if err != nil {
		return nil, err
	}
	// Get submission within the current interval
	fromBlock := currentBlockNumber / updatePricesFrequency * updatePricesFrequency
	submissions, err := GetLatestPricesSubmissions(rp, fromBlock, intervalSize, opts)
	if err != nil {
		return nil, err
	}
	// Build and return result table
	participationTable := make(map[common.Address]bool)
	for _, member := range members {
		participationTable[member.Address] = false
	}
	for _, submission := range submissions {
		participationTable[submission] = true
	}
	return participationTable, nil
}

// Get the smoothing pool opt-in status of a node
func GetSmoothingPoolRegistrationState(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNodeManager, err := getRocketNodeManager(rp, opts)
	if err != nil {
		return false, err
	}
	state := new(bool)
	if err := rocketNodeManager.Call(opts, state, "getSmoothingPoolRegistrationState", nodeAddress); err != nil {
		return false, fmt.Errorf("Could not get node %s smoothing pool registration status: %w", nodeAddress.Hex(), err)
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
		return time.Time{}, fmt.Errorf("Could not get node %s's last smoothing pool registration change time: %w", nodeAddress.Hex(), err)
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
		return nil, fmt.Errorf("Could not get node %s's last smoothing pool registration change time: %w", nodeAddress.Hex(), err)
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
		return common.Hash{}, fmt.Errorf("Could not set smoothing pool registration state: %w", err)
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
				return fmt.Errorf("Could not get smoothing pool opt-in count for batch starting at %d: %w", offset, err)
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
