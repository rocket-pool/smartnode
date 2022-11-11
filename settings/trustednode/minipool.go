package trustednode

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Config
const (
	MinipoolSettingsContractName = "rocketDAONodeTrustedSettingsMinipool"
	ScrubPeriodPath              = "minipool.scrub.period"
	ScrubPenaltyEnabledPath      = "minipool.scrub.penalty.enabled"
)

// The cooldown period a member must wait after making a proposal before making another in seconds
func GetScrubPeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getScrubPeriod"); err != nil {
		return 0, fmt.Errorf("Could not get scrub period: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapScrubPeriod(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapUint(rp, MinipoolSettingsContractName, ScrubPeriodPath, big.NewInt(int64(value)), opts)
}
func ProposeScrubPeriod(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", ScrubPeriodPath), MinipoolSettingsContractName, ScrubPeriodPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeScrubPeriodGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ScrubPeriodPath), MinipoolSettingsContractName, ScrubPeriodPath, big.NewInt(int64(value)), opts)
}

// Whether or not the RPL slashing penalty is applied to scrubbed minipools
func GetScrubPenaltyEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := minipoolSettingsContract.Call(opts, value, "getScrubPenaltyEnabled"); err != nil {
		return false, fmt.Errorf("Could not get scrub penalty setting: %w", err)
	}
	return (*value), nil
}
func BootstrapScrubPenaltyEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapBool(rp, MinipoolSettingsContractName, ScrubPenaltyEnabledPath, value, opts)
}
func ProposeScrubPenaltyEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetBool(rp, fmt.Sprintf("set %s", ScrubPenaltyEnabledPath), MinipoolSettingsContractName, ScrubPenaltyEnabledPath, value, opts)
}
func EstimateProposeScrubPenaltyEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", ScrubPenaltyEnabledPath), MinipoolSettingsContractName, ScrubPenaltyEnabledPath, value, opts)
}

// Get contracts
var minipoolSettingsContractLock sync.Mutex

func getMinipoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	minipoolSettingsContractLock.Lock()
	defer minipoolSettingsContractLock.Unlock()
	return rp.GetContract(MinipoolSettingsContractName, opts)
}
