package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	psettings "github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

const (
	nodeNamespace string = "node"
)

// Node registrations currently enabled
func ProposeNodeRegistrationEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.NodeRegistrationEnabledSettingPath), nodeNamespace, psettings.NodeRegistrationEnabledSettingPath, value, opts)
}
func EstimateProposeNodeRegistrationEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.NodeRegistrationEnabledSettingPath), nodeNamespace, psettings.NodeRegistrationEnabledSettingPath, value, opts)
}

// Smoothing pool joining currently enabled
func ProposeSmoothingPoolRegistrationEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.SmoothingPoolRegistrationEnabledSettingPath), nodeNamespace, psettings.SmoothingPoolRegistrationEnabledSettingPath, value, opts)
}
func EstimateProposeSmoothingPoolRegistrationEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.SmoothingPoolRegistrationEnabledSettingPath), nodeNamespace, psettings.SmoothingPoolRegistrationEnabledSettingPath, value, opts)
}

// Node deposits currently enabled
func ProposeNodeDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.NodeDepositEnabledSettingPath), nodeNamespace, psettings.NodeDepositEnabledSettingPath, value, opts)
}
func EstimateProposeNodeDepositEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.NodeDepositEnabledSettingPath), nodeNamespace, psettings.NodeDepositEnabledSettingPath, value, opts)
}

// Vacant minipools currently enabled
func ProposeVacantMinipoolsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.VacantMinipoolsEnabledSettingPath), nodeNamespace, psettings.VacantMinipoolsEnabledSettingPath, value, opts)
}
func EstimateProposeVacantMinipoolsEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.VacantMinipoolsEnabledSettingPath), nodeNamespace, psettings.VacantMinipoolsEnabledSettingPath, value, opts)
}
