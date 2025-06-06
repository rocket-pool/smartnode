package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	psettings "github.com/rocket-pool/smartnode/bindings/settings/protocol"
)

const (
	networkNamespace string = "network"
)

// Network balance submissions currently enabled
func ProposeSubmitBalancesEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.SubmitBalancesEnabledSettingPath), networkNamespace, psettings.SubmitBalancesEnabledSettingPath, value, opts)
}
func EstimateProposeSubmitBalancesEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.SubmitBalancesEnabledSettingPath), networkNamespace, psettings.SubmitBalancesEnabledSettingPath, value, opts)
}

// Rewards submissions currently enabled
func ProposeSubmitRewardsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.SubmitRewardsEnabledSettingPath), networkNamespace, psettings.SubmitRewardsEnabledSettingPath, value, opts)
}
func EstimateProposeSubmitRewardsEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.SubmitRewardsEnabledSettingPath), networkNamespace, psettings.SubmitRewardsEnabledSettingPath, value, opts)
}
