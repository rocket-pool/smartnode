package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Config
const (
	AuctionSettingsContractName      string = "rocketDAOProtocolSettingsAuction"
	CreateLotEnabledSettingPath      string = "auction.lot.create.enabled"
	BidOnLotEnabledSettingPath       string = "auction.lot.bidding.enabled"
	LotMinimumEthValueSettingPath    string = "auction.lot.value.minimum"
	LotMaximumEthValueSettingPath    string = "auction.lot.value.maximum"
	LotDurationSettingPath           string = "auction.lot.duration"
	LotStartingPriceRatioSettingPath string = "auction.price.start"
	LotReservePriceRatioSettingPath  string = "auction.price.reserve"
)

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
func ProposeCreateLotEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", CreateLotEnabledSettingPath), AuctionSettingsContractName, CreateLotEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeCreateLotEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", CreateLotEnabledSettingPath), AuctionSettingsContractName, CreateLotEnabledSettingPath, value, blockNumber, treeNodes, opts)
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
func ProposeBidOnLotEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", BidOnLotEnabledSettingPath), AuctionSettingsContractName, BidOnLotEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeBidOnLotEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", BidOnLotEnabledSettingPath), AuctionSettingsContractName, BidOnLotEnabledSettingPath, value, blockNumber, treeNodes, opts)
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
func ProposeLotMinimumEthValue(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", LotMinimumEthValueSettingPath), AuctionSettingsContractName, LotMinimumEthValueSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeLotMinimumEthValueGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", LotMinimumEthValueSettingPath), AuctionSettingsContractName, LotMinimumEthValueSettingPath, value, blockNumber, treeNodes, opts)
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
func ProposeLotMaximumEthValue(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", LotMaximumEthValueSettingPath), AuctionSettingsContractName, LotMaximumEthValueSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeLotMaximumEthValueGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", LotMaximumEthValueSettingPath), AuctionSettingsContractName, LotMaximumEthValueSettingPath, value, blockNumber, treeNodes, opts)
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
func ProposeLotDuration(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", LotDurationSettingPath), AuctionSettingsContractName, LotDurationSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeLotDurationGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", LotDurationSettingPath), AuctionSettingsContractName, LotDurationSettingPath, value, blockNumber, treeNodes, opts)
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
func ProposeLotStartingPriceRatio(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", LotStartingPriceRatioSettingPath), AuctionSettingsContractName, LotStartingPriceRatioSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeLotStartingPriceRatioGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", LotStartingPriceRatioSettingPath), AuctionSettingsContractName, LotStartingPriceRatioSettingPath, value, blockNumber, treeNodes, opts)
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
func ProposeLotReservePriceRatio(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", LotReservePriceRatioSettingPath), AuctionSettingsContractName, LotReservePriceRatioSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeLotReservePriceRatioGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", LotReservePriceRatioSettingPath), AuctionSettingsContractName, LotReservePriceRatioSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var auctionSettingsContractLock sync.Mutex

func getAuctionSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	auctionSettingsContractLock.Lock()
	defer auctionSettingsContractLock.Unlock()
	return rp.GetContract(AuctionSettingsContractName, opts)
}
