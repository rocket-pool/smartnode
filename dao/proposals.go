package dao

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    rptypes "github.com/rocket-pool/rocketpool-go/types"
)


// Get the proposal count
func GetProposalCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    proposalCount := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, proposalCount, "getTotal"); err != nil {
        return 0, fmt.Errorf("Could not get proposal count: %w", err)
    }
    return (*proposalCount).Uint64(), nil
}


// Proposal details
func GetProposalDAO(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (string, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return "", err
    }
    daoName := new(string)
    if err := rocketDAOProposal.Call(opts, daoName, "getDAO", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d DAO: %w", proposalId, err)
    }
    return *daoName, nil
}
func GetProposalProposerAddress(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (common.Address, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return common.Address{}, err
    }
    proposerAddress := new(common.Address)
    if err := rocketDAOProposal.Call(opts, proposerAddress, "getProposer", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d proposer address: %w", proposalId, err)
    }
    return *proposerAddress, nil
}
func GetProposalCreatedBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    createdBlock := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, createdBlock, "getCreated", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d created block: %w", proposalId, err)
    }
    return (*createdBlock).Uint64(), nil
}
func GetProposalStartBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    startBlock := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, startBlock, "getStart", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d start block: %w", proposalId, err)
    }
    return (*startBlock).Uint64(), nil
}
func GetProposalEndBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    endBlock := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, endBlock, "getEnd", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d end block: %w", proposalId, err)
    }
    return (*endBlock).Uint64(), nil
}
func GetProposalExpiryBlock(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    expiryBlock := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, expiryBlock, "getExpires", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d expiry block: %w", proposalId, err)
    }
    return (*expiryBlock).Uint64(), nil
}
func GetProposalVotesRequired(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    votesRequired := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, votesRequired, "getVotesRequired", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d votes required: %w", proposalId, err)
    }
    return (*votesRequired).Uint64(), nil
}
func GetProposalVotesFor(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    votesFor := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, votesFor, "getVotesFor", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d votes for: %w", proposalId, err)
    }
    return (*votesFor).Uint64(), nil
}
func GetProposalVotesAgainst(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    votesAgainst := new(*big.Int)
    if err := rocketDAOProposal.Call(opts, votesAgainst, "getVotesAgainst", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d votes against: %w", proposalId, err)
    }
    return (*votesAgainst).Uint64(), nil
}
func GetProposalIsCancelled(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return false, err
    }
    cancelled := new(bool)
    if err := rocketDAOProposal.Call(opts, cancelled, "getCancelled", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d cancelled status: %w", proposalId, err)
    }
    return *cancelled, nil
}
func GetProposalIsExecuted(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return false, err
    }
    executed := new(bool)
    if err := rocketDAOProposal.Call(opts, executed, "getExecuted", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d executed status: %w", proposalId, err)
    }
    return *executed, nil
}
func GetProposalPayload(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) ([]byte, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return []byte{}, err
    }
    payload := new([]byte)
    if err := rocketDAOProposal.Call(opts, payload, "getPayload", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d payload: %w", proposalId, err)
    }
    return *payload, nil
}
func GetProposalState(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (rptypes.ProposalState, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return 0, err
    }
    state := new(uint8)
    if err := rocketDAOProposal.Call(opts, state, "getState", big.NewInt(int64(proposalId))); err != nil {
        return 0, fmt.Errorf("Could not get proposal %d state: %w", proposalId, err)
    }
    return rptypes.ProposalState(*state), nil
}


// Get whether a member has voted on a proposal
func GetProposalMemberVoted(rp *rocketpool.RocketPool, proposalId uint64, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return false, err
    }
    voted := new(bool)
    if err := rocketDAOProposal.Call(opts, voted, "getReceiptHasVoted", big.NewInt(int64(proposalId)), memberAddress); err != nil {
        return false, fmt.Errorf("Could not get proposal %d member %s voted status: %w", proposalId, memberAddress.Hex(), err)
    }
    return *voted, nil
}


// Get whether a member has voted in support of a proposal
func GetProposalMemberSupported(rp *rocketpool.RocketPool, proposalId uint64, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
    rocketDAOProposal, err := getRocketDAOProposal(rp)
    if err != nil {
        return false, err
    }
    supported := new(bool)
    if err := rocketDAOProposal.Call(opts, supported, "getReceiptSupported", big.NewInt(int64(proposalId)), memberAddress); err != nil {
        return false, fmt.Errorf("Could not get proposal %d member %s supported status: %w", proposalId, memberAddress.Hex(), err)
    }
    return *supported, nil
}


// Get contracts
var rocketDAOProposalLock sync.Mutex
func getRocketDAOProposal(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAOProposalLock.Lock()
    defer rocketDAOProposalLock.Unlock()
    return rp.GetContract("rocketDAOProposal")
}

