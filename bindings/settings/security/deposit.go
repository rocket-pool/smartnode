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
	depositNamespace string = "deposit"
)

// Deposits currently enabled
func ProposeDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.DepositEnabledSettingPath), depositNamespace, psettings.DepositEnabledSettingPath, value, opts)
}
func EstimateProposeDepositEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.DepositEnabledSettingPath), depositNamespace, psettings.DepositEnabledSettingPath, value, opts)
}

// Deposit assignments currently enabled
func ProposeAssignDepositsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.AssignDepositsEnabledSettingPath), depositNamespace, psettings.AssignDepositsEnabledSettingPath, value, opts)
}
func EstimateProposeAssignDepositsEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.AssignDepositsEnabledSettingPath), depositNamespace, psettings.AssignDepositsEnabledSettingPath, value, opts)
}
