package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	psettings "github.com/rocket-pool/rocketpool-go/settings/protocol"
)

const (
	auctionNamespace string = "auction"
)

// Lot creation currently enabled
func ProposeCreateLotEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.CreateLotEnabledSettingPath), auctionNamespace, psettings.CreateLotEnabledSettingPath, value, opts)
}
func EstimateProposeCreateLotEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.CreateLotEnabledSettingPath), auctionNamespace, psettings.CreateLotEnabledSettingPath, value, opts)
}

// Lot bidding currently enabled
func ProposeBidOnLotEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.BidOnLotEnabledSettingPath), auctionNamespace, psettings.BidOnLotEnabledSettingPath, value, opts)
}
func EstimateProposeBidOnLotEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.BidOnLotEnabledSettingPath), auctionNamespace, psettings.BidOnLotEnabledSettingPath, value, opts)
}
