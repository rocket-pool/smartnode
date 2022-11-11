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
	ProposalsSettingsContractName = "rocketDAONodeTrustedSettingsProposals"
	CooldownTimeSettingPath       = "proposal.cooldown.time"
	VoteTimeSettingPath           = "proposal.vote.time"
	VoteDelayTimeSettingPath      = "proposal.vote.delay.time"
	ExecuteTimeSettingPath        = "proposal.execute.time"
	ActionTimeSettingPath         = "proposal.action.time"
)

// The cooldown period a member must wait after making a proposal before making another in seconds
func GetProposalCooldownTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getCooldownTime"); err != nil {
		return 0, fmt.Errorf("Could not get proposal cooldown period: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapProposalCooldownTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapUint(rp, ProposalsSettingsContractName, CooldownTimeSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeProposalCooldownTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", CooldownTimeSettingPath), ProposalsSettingsContractName, CooldownTimeSettingPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeProposalCooldownTimeGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", CooldownTimeSettingPath), ProposalsSettingsContractName, CooldownTimeSettingPath, big.NewInt(int64(value)), opts)
}

// The period a proposal can be voted on for in seconds
func GetProposalVoteTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getVoteTime"); err != nil {
		return 0, fmt.Errorf("Could not get proposal voting period: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapProposalVoteTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapUint(rp, ProposalsSettingsContractName, VoteTimeSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeProposalVoteTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", VoteTimeSettingPath), ProposalsSettingsContractName, VoteTimeSettingPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeProposalVoteTimeGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", VoteTimeSettingPath), ProposalsSettingsContractName, VoteTimeSettingPath, big.NewInt(int64(value)), opts)
}

// The delay after creation before a proposal can be voted on in seconds
func GetProposalVoteDelayTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getVoteDelayTime"); err != nil {
		return 0, fmt.Errorf("Could not get proposal voting delay: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapProposalVoteDelayTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapUint(rp, ProposalsSettingsContractName, VoteDelayTimeSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeProposalVoteDelayTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", VoteDelayTimeSettingPath), ProposalsSettingsContractName, VoteDelayTimeSettingPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeProposalVoteDelayTimeGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", VoteDelayTimeSettingPath), ProposalsSettingsContractName, VoteDelayTimeSettingPath, big.NewInt(int64(value)), opts)
}

// The period during which a passed proposal can be executed in time
func GetProposalExecuteTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getExecuteTime"); err != nil {
		return 0, fmt.Errorf("Could not get proposal execution period: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapProposalExecuteTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapUint(rp, ProposalsSettingsContractName, ExecuteTimeSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeProposalExecuteTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", ExecuteTimeSettingPath), ProposalsSettingsContractName, ExecuteTimeSettingPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeProposalExecuteTimeGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ExecuteTimeSettingPath), ProposalsSettingsContractName, ExecuteTimeSettingPath, big.NewInt(int64(value)), opts)
}

// The period during which an action can be performed on an executed proposal in seconds
func GetProposalActionTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	proposalsSettingsContract, err := getProposalsSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := proposalsSettingsContract.Call(opts, value, "getActionTime"); err != nil {
		return 0, fmt.Errorf("Could not get proposal action period: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapProposalActionTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return trustednodedao.BootstrapUint(rp, ProposalsSettingsContractName, ActionTimeSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeProposalActionTime(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return trustednodedao.ProposeSetUint(rp, fmt.Sprintf("set %s", ActionTimeSettingPath), ProposalsSettingsContractName, ActionTimeSettingPath, big.NewInt(int64(value)), opts)
}
func EstimateProposeProposalActionTimeGas(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return trustednodedao.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ActionTimeSettingPath), ProposalsSettingsContractName, ActionTimeSettingPath, big.NewInt(int64(value)), opts)
}

// Get contracts
var proposalsSettingsContractLock sync.Mutex

func getProposalsSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	proposalsSettingsContractLock.Lock()
	defer proposalsSettingsContractLock.Unlock()
	return rp.GetContract(ProposalsSettingsContractName, opts)
}
