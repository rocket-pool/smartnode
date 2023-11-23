package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
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

// Get contracts
var rocketNetworkBalancesLock sync.Mutex

func getRocketNetworkBalances(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkBalancesLock.Lock()
	defer rocketNetworkBalancesLock.Unlock()
	return rp.GetContract("rocketNetworkBalances", opts)
}
