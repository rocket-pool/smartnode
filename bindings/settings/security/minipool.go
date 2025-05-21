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
	minipoolNamespace string = "minipool"
)

// Minipool withdrawable event submissions currently enabled
func ProposeMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.MinipoolSubmitWithdrawableEnabledSettingPath), minipoolNamespace, psettings.MinipoolSubmitWithdrawableEnabledSettingPath, value, opts)
}
func EstimateProposeMinipoolSubmitWithdrawableEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.MinipoolSubmitWithdrawableEnabledSettingPath), minipoolNamespace, psettings.MinipoolSubmitWithdrawableEnabledSettingPath, value, opts)
}

// Minipool bond reductions currently enabled
func ProposeBondReductionEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.BondReductionEnabledSettingPath), minipoolNamespace, psettings.BondReductionEnabledSettingPath, value, opts)
}
func EstimateProposeBondReductionEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.BondReductionEnabledSettingPath), minipoolNamespace, psettings.BondReductionEnabledSettingPath, value, opts)
}
