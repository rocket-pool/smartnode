package trustednode

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Join the trusted node DAO
// Requires an executed invite proposal
func Join(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedActions.Transact(opts, "actionJoin")
    if err != nil {
        return nil, fmt.Errorf("Could not join the trusted node DAO: %w")
    }
    return txReceipt, nil
}


// Leave the trusted node DAO
// Requires an executed leave proposal
func Leave(rp *rocketpool.RocketPool, bondRefundAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedActions.Transact(opts, "actionLeave", bondRefundAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not leave the trusted node DAO: %w")
    }
    return txReceipt, nil
}


// Replace node's position in the trusted node DAO with another node
// Requires an executed replace proposal
func Replace(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrustedActions.Transact(opts, "actionLeave")
    if err != nil {
        return nil, fmt.Errorf("Could not replace node's position in the trusted node DAO: %w")
    }
    return txReceipt, nil
}


// Get contracts
var rocketDAONodeTrustedActionsLock sync.Mutex
func getRocketDAONodeTrustedActions(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAONodeTrustedActionsLock.Lock()
    defer rocketDAONodeTrustedActionsLock.Unlock()
    return rp.GetContract("rocketDAONodeTrustedActions")
}

