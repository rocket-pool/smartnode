package network

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"gonum.org/v1/gonum/mathext"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Info for a balances updated event
type BalancesUpdatedEvent struct {
	BlockNumber    *big.Int `json:"blockNumber"`
	SlotTimestamp  *big.Int `json:"slotTimestamp"`
	TotalEth       *big.Int `json:"totalEth"`
	StakingEth     *big.Int `json:"stakingEth"`
	RethSupply     *big.Int `json:"rethSupply"`
	BlockTimestamp *big.Int `json:"blockTimestamp"`
}

// Get the block number which network balances are current for
func GetBalancesBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return 0, err
	}
	balancesBlock := new(*big.Int)
	if err := rocketNetworkBalances.Call(opts, balancesBlock, "getBalancesBlock"); err != nil {
		return 0, fmt.Errorf("error getting network balances block: %w", err)
	}
	return (*balancesBlock).Uint64(), nil
}

// Get the block number which network balances are current for
func GetBalancesBlockRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return nil, err
	}
	balancesBlock := new(*big.Int)
	if err := rocketNetworkBalances.Call(opts, balancesBlock, "getBalancesBlock"); err != nil {
		return nil, fmt.Errorf("error getting network balances block: %w", err)
	}
	return *balancesBlock, nil
}

// Get the current network total ETH balance
func GetTotalETHBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return nil, err
	}
	totalEthBalance := new(*big.Int)
	if err := rocketNetworkBalances.Call(opts, totalEthBalance, "getTotalETHBalance"); err != nil {
		return nil, fmt.Errorf("error getting network total ETH balance: %w", err)
	}
	return *totalEthBalance, nil
}

// Get the current network staking ETH balance
func GetStakingETHBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return nil, err
	}
	stakingEthBalance := new(*big.Int)
	if err := rocketNetworkBalances.Call(opts, stakingEthBalance, "getStakingETHBalance"); err != nil {
		return nil, fmt.Errorf("error getting network staking ETH balance: %w", err)
	}
	return *stakingEthBalance, nil
}

// Get the current network total rETH supply
func GetTotalRETHSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRethSupply := new(*big.Int)
	if err := rocketNetworkBalances.Call(opts, totalRethSupply, "getTotalRETHSupply"); err != nil {
		return nil, fmt.Errorf("error getting network total rETH supply: %w", err)
	}
	return *totalRethSupply, nil
}

// Get the current network ETH utilization rate
func GetETHUtilizationRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return 0, err
	}
	ethUtilizationRate := new(*big.Int)
	if err := rocketNetworkBalances.Call(opts, ethUtilizationRate, "getETHUtilizationRate"); err != nil {
		return 0, fmt.Errorf("error getting network ETH utilization rate: %w", err)
	}
	return eth.WeiToEth(*ethUtilizationRate), nil
}

// Estimate the gas of SubmitBalances
func EstimateSubmitBalancesGas(rp *rocketpool.RocketPool, block uint64, slotTimestamp uint64, totalEth, stakingEth, rethSupply *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkBalances.GetTransactionGasInfo(opts, "submitBalances", big.NewInt(int64(block)), big.NewInt(int64(slotTimestamp)), totalEth, stakingEth, rethSupply)
}

// Submit network balances for an epoch
func SubmitBalances(rp *rocketpool.RocketPool, block uint64, slotTimestamp uint64, totalEth, stakingEth, rethSupply *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkBalances.Transact(opts, "submitBalances", big.NewInt(int64(block)), big.NewInt(int64(slotTimestamp)), stakingEth, rethSupply)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error submitting network balances: %w", err)
	}
	return tx.Hash(), nil
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

func GetBalancesUpdatedEvent(rp *rocketpool.RocketPool, blockNumber uint64, opts *bind.CallOpts) (bool, BalancesUpdatedEvent, error) {
	// Get contracts
	rocketNetworkBalances, err := getRocketNetworkBalances(rp, opts)
	if err != nil {
		return false, BalancesUpdatedEvent{}, err
	}

	// Create the list of addresses to check
	currentAddress := *rocketNetworkBalances.Address
	rocketNetworkBalancesAddress := []common.Address{currentAddress}

	// Construct a filter query for relevant logs
	balancesUpdatedEvent := rocketNetworkBalances.ABI.Events["BalancesUpdated"]
	indexBytes := [32]byte{}
	addressFilter := rocketNetworkBalancesAddress
	topicFilter := [][]common.Hash{{balancesUpdatedEvent.ID}, {indexBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, big.NewInt(1), big.NewInt(int64(blockNumber)), big.NewInt(int64(blockNumber)), nil)
	if err != nil {
		return false, BalancesUpdatedEvent{}, err
	}
	if len(logs) == 0 {
		return false, BalancesUpdatedEvent{}, nil
	}

	// Get the log info values
	values, err := balancesUpdatedEvent.Inputs.Unpack(logs[0].Data)
	if err != nil {
		return false, BalancesUpdatedEvent{}, fmt.Errorf("error unpacking price updated event data: %w", err)
	}

	// Convert to a native struct
	var eventData BalancesUpdatedEvent
	err = balancesUpdatedEvent.Inputs.Copy(&eventData, values)
	if err != nil {
		return false, BalancesUpdatedEvent{}, fmt.Errorf("error converting price updated event data to struct: %w", err)
	}

	return true, eventData, nil
}

// TODO: will be adjusted/removed
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

// TODO: will be adjusted/removed
// Calculates the participation rate of every trusted node on balance submission since the last block that member count changed
func CalculateTrustedNodeBalancesParticipation(rp *rocketpool.RocketPool, intervalSize *big.Int, opts *bind.CallOpts) (*node.TrustedNodeParticipation, error) {
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
	latestMemberCountChangedBlock, err := trustednode.GetLatestMemberCountChangedBlock(rp, minBlock, intervalSize, opts)
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
	participation := node.TrustedNodeParticipation{
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

// Get contracts
var rocketNetworkBalancesLock sync.Mutex

func getRocketNetworkBalances(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkBalancesLock.Lock()
	defer rocketNetworkBalancesLock.Unlock()
	return rp.GetContract("rocketNetworkBalances", opts)
}
