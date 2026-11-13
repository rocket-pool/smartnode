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
	megapoolNamespace string = "megapool"
)

// Proofless staking currently enabled
func ProposeStakeWithoutProofEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return security.ProposeSetBool(rp, fmt.Sprintf("set %s", psettings.MegapoolStakeWithoutProofEnabledPath), megapoolNamespace, psettings.MegapoolStakeWithoutProofEnabledPath, value, opts)
}
func EstimateProposeStakeWithoutProofEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return security.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", psettings.MegapoolStakeWithoutProofEnabledPath), megapoolNamespace, psettings.MegapoolStakeWithoutProofEnabledPath, value, opts)
}
