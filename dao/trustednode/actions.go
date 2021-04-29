package trustednode

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Join the trusted node DAO
// Requires an executed invite proposal
func Join(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketDAONodeTrustedActions.Transact(opts, "actionJoin")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not join the trusted node DAO: %w", err)
    }
    return hash, nil
}


// Leave the trusted node DAO
// Requires an executed leave proposal
func Leave(rp *rocketpool.RocketPool, rplBondRefundAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketDAONodeTrustedActions.Transact(opts, "actionLeave", rplBondRefundAddress)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not leave the trusted node DAO: %w", err)
    }
    return hash, nil
}


// Make a challenge against a node
func MakeChallenge(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketDAONodeTrustedActions.Transact(opts, "actionChallengeMake", memberAddress)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not challenge trusted node DAO member %s: %w", memberAddress.Hex(), err)
    }
    return hash, nil
}


// Decide a challenge against a node
func DecideChallenge(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketDAONodeTrustedActions.Transact(opts, "actionChallengeDecide", memberAddress)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not decide the challenge against trusted node DAO member %s: %w", memberAddress.Hex(), err)
    }
    return hash, nil
}


// Get contracts
var rocketDAONodeTrustedActionsLock sync.Mutex
func getRocketDAONodeTrustedActions(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAONodeTrustedActionsLock.Lock()
    defer rocketDAONodeTrustedActionsLock.Unlock()
    return rp.GetContract("rocketDAONodeTrustedActions")
}

