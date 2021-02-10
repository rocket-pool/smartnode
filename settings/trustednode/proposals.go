package trustednode

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Config
const (
    ProposalsSettingsContractName = "rocketDAONodeTrustedSettingsProposals"
    CooldownSettingPath = "proposal.cooldown"
    VoteBlocksSettingPath = "proposal.vote.blocks"
    VoteDelayBlocksSettingPath = "proposal.vote.delay.blocks"
    ExecuteBlocksSettingPath = "proposal.execute.blocks"
    ActionBlocksSettingPath = "proposal.action.blocks"
)


// The cooldown period a member must wait after making a proposal before making another in blocks
func GetCooldown(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    proposalsSettingsContract, err := getProposalsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := proposalsSettingsContract.Call(opts, value, "getCooldown"); err != nil {
        return 0, fmt.Errorf("Could not get proposal cooldown period: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapCooldown(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, CooldownSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeCooldown(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", CooldownSettingPath), ProposalsSettingsContractName, CooldownSettingPath, big.NewInt(int64(value)), opts)
}


// The period a proposal can be voted on for in blocks
func GetVoteBlocks(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    proposalsSettingsContract, err := getProposalsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := proposalsSettingsContract.Call(opts, value, "getVoteBlocks"); err != nil {
        return 0, fmt.Errorf("Could not get proposal voting period: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapVoteBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, VoteBlocksSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeVoteBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", VoteBlocksSettingPath), ProposalsSettingsContractName, VoteBlocksSettingPath, big.NewInt(int64(value)), opts)
}


// The delay after creation before a proposal can be voted on in blocks
func GetVoteDelayBlocks(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    proposalsSettingsContract, err := getProposalsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := proposalsSettingsContract.Call(opts, value, "getVoteDelayBlocks"); err != nil {
        return 0, fmt.Errorf("Could not get proposal voting delay: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapVoteDelayBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, VoteDelayBlocksSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeVoteDelayBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", VoteDelayBlocksSettingPath), ProposalsSettingsContractName, VoteDelayBlocksSettingPath, big.NewInt(int64(value)), opts)
}


// The period during which a passed proposal can be executed in blocks
func GetExecuteBlocks(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    proposalsSettingsContract, err := getProposalsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := proposalsSettingsContract.Call(opts, value, "getExecuteBlocks"); err != nil {
        return 0, fmt.Errorf("Could not get proposal execution period: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapExecuteBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, ExecuteBlocksSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeExecuteBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", ExecuteBlocksSettingPath), ProposalsSettingsContractName, ExecuteBlocksSettingPath, big.NewInt(int64(value)), opts)
}


// The period during which an action can be performed on an executed proposal in blocks
func GetActionBlocks(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    proposalsSettingsContract, err := getProposalsSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := proposalsSettingsContract.Call(opts, value, "getActionBlocks"); err != nil {
        return 0, fmt.Errorf("Could not get proposal action period: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapActionBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, ActionBlocksSettingPath, big.NewInt(int64(value)), opts)
}
func ProposeActionBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return trustednode.ProposeSetUint(rp, fmt.Sprintf("set %s", ActionBlocksSettingPath), ProposalsSettingsContractName, ActionBlocksSettingPath, big.NewInt(int64(value)), opts)
}


// Get contracts
var proposalsSettingsContractLock sync.Mutex
func getProposalsSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    proposalsSettingsContractLock.Lock()
    defer proposalsSettingsContractLock.Unlock()
    return rp.GetContract(ProposalsSettingsContractName)
}

