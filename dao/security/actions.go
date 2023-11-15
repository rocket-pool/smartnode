package security

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas of Join
func EstimateJoinGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityActions.GetTransactionGasInfo(opts, "actionJoin")
}

// Join the security DAO
// Requires an executed invite proposal
func Join(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityActions.Transact(opts, "actionJoin")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error joining the security DAO: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Kick
func EstimateKickGas(rp *rocketpool.RocketPool, address common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityActions.GetTransactionGasInfo(opts, "actionKick", address)
}

// Removes a member from the security DAO
func Kick(rp *rocketpool.RocketPool, address common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityActions.Transact(opts, "actionKick", address)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error kicking %s from the security DAO: %w", address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of KickMulti
func EstimateKickMultiGas(rp *rocketpool.RocketPool, addresses []common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityActions.GetTransactionGasInfo(opts, "actionKickMulti", addresses)
}

// Removes multiple members from the security DAO
func KickMulti(rp *rocketpool.RocketPool, addresses []common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityActions.Transact(opts, "actionKickMulti", addresses)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error kicking members from the security DAO: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of RequestLeave
func EstimateRequestLeaveGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityActions.GetTransactionGasInfo(opts, "actionRequestLeave")
}

// A member who wishes to leave the security council can call this method to initiate the process
func RequestLeave(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityActions.Transact(opts, "actionRequestLeave")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error requesting to leave the security DAO: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Leave
func EstimateLeaveGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAOSecurityActions.GetTransactionGasInfo(opts, "actionLeave")
}

// A member who has asked to leave and waited the required time can call this method to formally leave the security council
func Leave(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAOSecurityActions, err := getRocketDAOSecurityActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAOSecurityActions.Transact(opts, "actionLeave")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error leaving the security DAO: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDAOSecurityActionsLock sync.Mutex

func getRocketDAOSecurityActions(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOSecurityActionsLock.Lock()
	defer rocketDAOSecurityActionsLock.Unlock()
	return rp.GetContract("rocketDAOSecurityActions", opts)
}
