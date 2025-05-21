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

// Info for a price updated event
type PriceUpdatedEvent struct {
	BlockNumber   *big.Int `json:"blockNumber"`
	SlotTimestamp *big.Int `json:"slotTimestamp"`
	RplPrice      *big.Int `json:"rplPrice"`
	Time          *big.Int `json:"time"`
}

// Get the block number which network prices are current for
func GetPricesBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, opts)
	if err != nil {
		return 0, err
	}
	pricesBlock := new(*big.Int)
	if err := rocketNetworkPrices.Call(opts, pricesBlock, "getPricesBlock"); err != nil {
		return 0, fmt.Errorf("error getting network prices block: %w", err)
	}
	return (*pricesBlock).Uint64(), nil
}

// Get the current network RPL price in ETH
func GetRPLPrice(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, opts)
	if err != nil {
		return nil, err
	}
	rplPrice := new(*big.Int)
	if err := rocketNetworkPrices.Call(opts, rplPrice, "getRPLPrice"); err != nil {
		return nil, fmt.Errorf("error getting network RPL price: %w", err)
	}
	return *rplPrice, nil
}

// Estimate the gas of SubmitPrices
func EstimateSubmitPricesGas(rp *rocketpool.RocketPool, block uint64, slotTimestamp uint64, rplPrice *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkPrices.GetTransactionGasInfo(opts, "submitPrices", big.NewInt(int64(block)), big.NewInt(int64(slotTimestamp)), rplPrice)
}

// Submit network prices and total effective RPL stake for an epoch
func SubmitPrices(rp *rocketpool.RocketPool, block uint64, slotTimestamp uint64, rplPrice *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkPrices.Transact(opts, "submitPrices", big.NewInt(int64(block)), big.NewInt(int64(slotTimestamp)), rplPrice)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error submitting network prices: %w", err)
	}
	return tx.Hash(), nil
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
	topicFilter := [][]common.Hash{{rocketNetworkPrices.ABI.Events["PricesSubmitted"].ID}, {common.BytesToHash(nodeAddress.Bytes())}}

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

// Get the event info for a price update
func GetPriceUpdatedEvent(rp *rocketpool.RocketPool, blockNumber uint64, opts *bind.CallOpts) (bool, PriceUpdatedEvent, error) {
	// Get contracts
	rocketNetworkPrices, err := getRocketNetworkPrices(rp, opts)
	if err != nil {
		return false, PriceUpdatedEvent{}, err
	}

	indexBig := big.NewInt(0).SetUint64(blockNumber)

	// Create the list of addresses to check
	currentAddress := *rocketNetworkPrices.Address
	rocketNetworkPricesAddress := []common.Address{currentAddress}

	// Construct a filter query for relevant logs
	pricesUpdatedEvent := rocketNetworkPrices.ABI.Events["PricesUpdated"]
	indexBytes := [32]byte{}
	indexBig.FillBytes(indexBytes[:])
	addressFilter := rocketNetworkPricesAddress
	topicFilter := [][]common.Hash{{pricesUpdatedEvent.ID}, {indexBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, big.NewInt(100), big.NewInt(int64(blockNumber)), big.NewInt(int64(blockNumber+1000)), nil)
	if err != nil {
		return false, PriceUpdatedEvent{}, err
	}
	if len(logs) == 0 {
		return false, PriceUpdatedEvent{}, nil
	}

	// Get the log info values
	values, err := pricesUpdatedEvent.Inputs.Unpack(logs[0].Data)
	if err != nil {
		return false, PriceUpdatedEvent{}, fmt.Errorf("error unpacking price updated event data: %w", err)
	}

	// Convert to a native struct
	var eventData PriceUpdatedEvent
	err = pricesUpdatedEvent.Inputs.Copy(&eventData, values)
	if err != nil {
		return false, PriceUpdatedEvent{}, fmt.Errorf("error converting price updated event data to struct: %w", err)
	}

	return true, eventData, nil
}

// Get contracts
var rocketNetworkPricesLock sync.Mutex

func getRocketNetworkPrices(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkPricesLock.Lock()
	defer rocketNetworkPricesLock.Unlock()
	return rp.GetContract("rocketNetworkPrices", opts)
}
