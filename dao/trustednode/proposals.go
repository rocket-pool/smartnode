package trustednode

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Submit a proposal to invite a new member to the trusted node DAO
func ProposeInviteMember(rp *rocketpool.RocketPool, message string, newMemberAddress common.Address, newMemberId, newMemberEmail string, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalInvite", newMemberId, newMemberEmail, newMemberAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not encode invite member proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a proposal for a member to leave the trusted node DAO
func ProposeMemberLeave(rp *rocketpool.RocketPool, message string, memberAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalLeave", memberAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not encode member leave proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a proposal to replace a member in the trusted node DAO
func ProposeReplaceMember(rp *rocketpool.RocketPool, message string, memberAddress, newMemberAddress common.Address, newMemberId, newMemberEmail string, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalReplace", memberAddress, newMemberId, newMemberEmail, newMemberAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not encode replace member proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a proposal to kick a member from the trusted node DAO
func ProposeKickMember(rp *rocketpool.RocketPool, message string, memberAddress common.Address, rplFineAmount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalKick", memberAddress, rplFineAmount)
    if err != nil {
        return nil, fmt.Errorf("Could not encode kick member proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a proposal to update a bool trusted node DAO setting
func ProposeSetBool(rp *rocketpool.RocketPool, message, contractName, settingPath string, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalSettingBool", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not encode set bool setting proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a proposal to update a uint trusted node DAO setting
func ProposeSetUint(rp *rocketpool.RocketPool, message, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalSettingUint", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not encode set uint setting proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a proposal to upgrade a contract
func ProposeUpgradeContract(rp *rocketpool.RocketPool, message, upgradeType, contractName, contractAbi string, contractAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    compressedAbi, err := rocketpool.EncodeAbiStr(contractAbi)
    if err != nil {
        return nil, err
    }
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    payload, err := rocketDAONodeTrustedProposals.ABI.Pack("proposalUpgrade", upgradeType, contractName, compressedAbi, contractAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not encode upgrade contract proposal payload: %w", err)
    }
    return SubmitProposal(rp, message, payload, opts)
}


// Submit a trusted node DAO proposal
func SubmitProposal(rp *rocketpool.RocketPool, message string, payload []byte, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedProposals.Transact(opts, "propose", message, payload)
    if err != nil {
        return nil, fmt.Errorf("Could not submit trusted node DAO proposal: %w")
    }
    return txReceipt, nil
}


// Cancel a submitted proposal
func CancelProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedProposals.Transact(opts, "cancel", big.NewInt(int64(proposalId)))
    if err != nil {
        return nil, fmt.Errorf("Could not cancel trusted node DAO proposal %d: %w", proposalId, err)
    }
    return txReceipt, nil
}


// Vote on a submitted proposal
func VoteOnProposal(rp *rocketpool.RocketPool, proposalId uint64, support bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedProposals.Transact(opts, "vote", big.NewInt(int64(proposalId)), support)
    if err != nil {
        return nil, fmt.Errorf("Could not vote on trusted node DAO proposal %d: %w", proposalId, err)
    }
    return txReceipt, nil
}


// Execute a submitted proposal
func ExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedProposals, err := getRocketDAONodeTrustedProposals(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedProposals.Transact(opts, "execute", big.NewInt(int64(proposalId)))
    if err != nil {
        return nil, fmt.Errorf("Could not execute trusted node DAO proposal %d: %w", proposalId, err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDAONodeTrustedProposalsLock sync.Mutex
func getRocketDAONodeTrustedProposals(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAONodeTrustedProposalsLock.Lock()
    defer rocketDAONodeTrustedProposalsLock.Unlock()
    return rp.GetContract("rocketDAONodeTrustedProposals")
}

