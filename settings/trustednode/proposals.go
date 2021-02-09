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
const ProposalsSettingsContractName = "rocketDAONodeTrustedSettingsProposals"


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
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, "proposal.cooldown", big.NewInt(int64(value)), opts)
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
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, "proposal.vote.blocks", big.NewInt(int64(value)), opts)
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
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, "proposal.vote.delay.blocks", big.NewInt(int64(value)), opts)
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
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, "proposal.execute.blocks", big.NewInt(int64(value)), opts)
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
    return trustednode.BootstrapUint(rp, ProposalsSettingsContractName, "proposal.action.blocks", big.NewInt(int64(value)), opts)
}


// Get contracts
var proposalsSettingsContractLock sync.Mutex
func getProposalsSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    proposalsSettingsContractLock.Lock()
    defer proposalsSettingsContractLock.Unlock()
    return rp.GetContract(ProposalsSettingsContractName)
}

