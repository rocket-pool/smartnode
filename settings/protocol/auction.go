package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Config
const AuctionSettingsContractName = "rocketDAOProtocolSettingsAuction"

// Lot creation currently enabled
func GetCreateLotEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := auctionSettingsContract.Call(opts, value, "getCreateLotEnabled"); err != nil {
		return false, fmt.Errorf("error getting lot creation enabled status: %w", err)
	}
	return *value, nil
}

// Lot bidding currently enabled
func GetBidOnLotEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := auctionSettingsContract.Call(opts, value, "getBidOnLotEnabled"); err != nil {
		return false, fmt.Errorf("error getting lot bidding enabled status: %w", err)
	}
	return *value, nil
}

// The minimum lot size in ETH value
func GetLotMinimumEthValue(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := auctionSettingsContract.Call(opts, value, "getLotMinimumEthValue"); err != nil {
		return nil, fmt.Errorf("error getting lot minimum ETH value: %w", err)
	}
	return *value, nil
}

// The maximum lot size in ETH value
func GetLotMaximumEthValue(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := auctionSettingsContract.Call(opts, value, "getLotMaximumEthValue"); err != nil {
		return nil, fmt.Errorf("error getting lot maximum ETH value: %w", err)
	}
	return *value, nil
}

// The lot duration in blocks
func GetLotDuration(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := auctionSettingsContract.Call(opts, value, "getLotDuration"); err != nil {
		return 0, fmt.Errorf("error getting lot duration: %w", err)
	}
	return (*value).Uint64(), nil
}

// The starting price relative to current ETH price, as a fraction
func GetLotStartingPriceRatio(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := auctionSettingsContract.Call(opts, value, "getStartingPriceRatio"); err != nil {
		return 0, fmt.Errorf("error getting lot starting price ratio: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// The reserve price relative to current ETH price, as a fraction
func GetLotReservePriceRatio(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	auctionSettingsContract, err := getAuctionSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := auctionSettingsContract.Call(opts, value, "getReservePriceRatio"); err != nil {
		return 0, fmt.Errorf("error getting lot reserve price ratio: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// Get contracts
var auctionSettingsContractLock sync.Mutex

func getAuctionSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	auctionSettingsContractLock.Lock()
	defer auctionSettingsContractLock.Unlock()
	return rp.GetContract(AuctionSettingsContractName, opts)
}
