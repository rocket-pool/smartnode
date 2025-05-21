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
	MinipoolSettingsContractName  = "rocketDAONodeTrustedSettingsMinipool"
	DissolvePeriodPath            = "megapool.dissolve.period"
	ScrubPeriodPath               = "minipool.scrub.period"
	PromotionScrubPeriodPath      = "minipool.promotion.scrub.period"
	ScrubPenaltyEnabledPath       = "minipool.scrub.penalty.enabled"
	BondReductionWindowStartPath  = "minipool.bond.reduction.window.start"
	BondReductionWindowLengthPath = "minipool.bond.reduction.window.length"
)

// The amount of time, in seconds, the scrub check lasts before a minipool can move from prelaunch to staking
func GetScrubPeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getScrubPeriod"); err != nil {
		return 0, fmt.Errorf("error getting scrub period: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeScrubPeriod(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", ScrubPeriodPath), MinipoolSettingsContractName, ScrubPeriodPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeScrubPeriodGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ScrubPeriodPath), MinipoolSettingsContractName, ScrubPeriodPath, big.NewInt(int64(value)), opts)
}

// The amount of time, in seconds, the promotion scrub check lasts before a vacant minipool can be promoted
func GetPromotionScrubPeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getPromotionScrubPeriod"); err != nil {
		return 0, fmt.Errorf("error getting promotion scrub period: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposePromotionScrubPeriod(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", PromotionScrubPeriodPath), MinipoolSettingsContractName, PromotionScrubPeriodPath, big.NewInt(int64(value)), opts)
}
func EstimateProposePromotionScrubPeriodGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PromotionScrubPeriodPath), MinipoolSettingsContractName, PromotionScrubPeriodPath, big.NewInt(int64(value)), opts)
}

// Whether or not the RPL slashing penalty is applied to scrubbed minipools
func GetScrubPenaltyEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := minipoolSettingsContract.Call(opts, value, "getScrubPenaltyEnabled"); err != nil {
		return false, fmt.Errorf("error getting scrub penalty setting: %w", err)
	}
	return (*value), nil
}
func ProposeScrubPenaltyEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetBool(rp, fmt.Sprintf("set %s", ScrubPenaltyEnabledPath), MinipoolSettingsContractName, ScrubPenaltyEnabledPath, value, opts)
}
func EstimateProposeScrubPenaltyEnabledGas(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", ScrubPenaltyEnabledPath), MinipoolSettingsContractName, ScrubPenaltyEnabledPath, value, opts)
}

// The amount of time, in seconds, a minipool must wait after beginning a bond reduction before it can apply the bond reduction (how long the Oracle DAO has to cancel the reduction if required)
func GetBondReductionWindowStart(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getBondReductionWindowStart"); err != nil {
		return 0, fmt.Errorf("error getting bond reduction window start: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeBondReductionWindowStart(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", BondReductionWindowStartPath), MinipoolSettingsContractName, BondReductionWindowStartPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeBondReductionWindowStartGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", BondReductionWindowStartPath), MinipoolSettingsContractName, BondReductionWindowStartPath, big.NewInt(int64(value)), opts)
}

// The amount of time, in seconds, a minipool has to reduce its bond once it has passed the check window
func GetBondReductionWindowLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getBondReductionWindowLength"); err != nil {
		return 0, fmt.Errorf("error getting bond reduction window length: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeBondReductionWindowLength(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", BondReductionWindowLengthPath), MinipoolSettingsContractName, BondReductionWindowLengthPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeBondReductionWindowLengthGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", BondReductionWindowLengthPath), MinipoolSettingsContractName, BondReductionWindowLengthPath, big.NewInt(int64(value)), opts)
}

// The amount of time, in seconds, how long after assignment before a validator can be dissolved
func GetDissolvePeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getDissolvePeriod"); err != nil {
		return 0, fmt.Errorf("error getting dissolve period: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeDissolvePeriod(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", DissolvePeriodPath), MinipoolSettingsContractName, DissolvePeriodPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeDissolvePeriodGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DissolvePeriodPath), MinipoolSettingsContractName, DissolvePeriodPath, big.NewInt(int64(value)), opts)
}

// Get contracts
var minipoolSettingsContractLock sync.Mutex

func getMinipoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	minipoolSettingsContractLock.Lock()
	defer minipoolSettingsContractLock.Unlock()
	return rp.GetContract(MinipoolSettingsContractName, opts)
}
