package megapool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

// Estimate the gas of Stake
func EstimateStakeGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "stake", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Progress the prelaunch megapool to staking
func Stake(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "stake", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
	if err != nil {
		return nil, fmt.Errorf("error staking megapool %s: %w", megapoolAddress, err)
	}
	return tx, nil
}

// Estimate the gas to call NotifyExit
func EstimateNotifyExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyExit", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Notify the megapool that one of its validators is exiting
func NotifyExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "notifyExit", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify exit: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call NotifyNotExit
func EstimateNotifyNotExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyNotExit", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Used to prove a validator is not exiting after a challenge-exit
func NotifyNotExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "notifyNotExit", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify not exit: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call NotifyFinalBalance
func EstimateNotifyFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, withdrawalProof megapool.WithdrawalProof, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyFinalBalance", megapoolAddress, validatorId, slotTimestamp, withdrawalProof, validatorProof, slotProof)
}

// Notify the megapool of the final balance of an exited validator
func NotifyFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, withdrawalProof megapool.WithdrawalProof, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "notifyFinalBalance", megapoolAddress, validatorId, slotTimestamp, withdrawalProof, validatorProof, slotProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify final balance: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call DissolveWithProof
func EstimateDissolveWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "dissolve", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Dissolve a validator using a proof that it used wrong credentials
func DissolveWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "dissolve", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
	if err != nil {
		return nil, fmt.Errorf("error calling dissolve with proof: %w", err)
	}
	return tx, nil
}

var rocketMegapoolManagerLock sync.Mutex

func getRocketMegapoolManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMegapoolManagerLock.Lock()
	defer rocketMegapoolManagerLock.Unlock()
	return rp.GetContract("rocketMegapoolManager", opts)
}
