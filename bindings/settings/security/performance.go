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
	performanceNamespace string = "performance"
)

// Performance exits currently enabled
func ProposePerformanceExitsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.PerformanceExitsEnabledSettingPath), performanceNamespace, psettings.PerformanceExitsEnabledSettingPath, value, opts)
}
func EstimateProposePerformanceExitsEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.PerformanceExitsEnabledSettingPath), performanceNamespace, psettings.PerformanceExitsEnabledSettingPath, value, opts)
}
