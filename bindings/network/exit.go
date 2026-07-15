package network

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Get the amount of ETH currently requested to exit
func GetRequestedEth(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketNetworkExit.Call(opts, value, "getRequestedEth"); err != nil {
		return nil, fmt.Errorf("error getting requested ETH: %w", err)
	}
	return *value, nil
}

// Get the start of the cooperative exit phase for a Minipool
func GetMinipoolCooperativeExitStart(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (time.Time, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketNetworkExit.Call(opts, value, "getMinipoolCooperativeExitStart", minipoolAddress); err != nil {
		return time.Time{}, fmt.Errorf("error getting minipool cooperative exit start for %s: %w", minipoolAddress.Hex(), err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the start of the cooperative exit phase for a Megapool validator
func GetMegapoolCooperativeExitStart(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, opts *bind.CallOpts) (time.Time, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketNetworkExit.Call(opts, value, "getMegapoolCooperativeExitStart", megapoolAddress, validatorId); err != nil {
		return time.Time{}, fmt.Errorf("error getting megapool cooperative exit start for %s validator %d: %w", megapoolAddress.Hex(), validatorId, err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Get the timestamp of the last exit request for a Minipool (or zero time if never requested)
func GetMinipoolLastExit(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (time.Time, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	value := new(*big.Int)
	if err := rocketNetworkExit.Call(opts, value, "getMinipoolLastExit", minipoolAddress); err != nil {
		return time.Time{}, fmt.Errorf("error getting minipool last exit for %s: %w", minipoolAddress.Hex(), err)
	}
	return time.Unix((*value).Int64(), 0), nil
}

// Estimate the gas of RequestMinipoolExit
func EstimateRequestMinipoolExitGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkExit.GetTransactionGasInfo(opts, "requestMinipoolExit", minipoolAddress)
}

// Request a Minipool to exit cooperatively
func RequestMinipoolExit(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkExit.Transact(opts, "requestMinipoolExit", minipoolAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error requesting minipool exit for %s: %w", minipoolAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ForceMinipoolExit
func EstimateForceMinipoolExitGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkExit.GetTransactionGasInfo(opts, "forceMinipoolExit", minipoolAddress)
}

// Force a Minipool to exit after the cooperative exit phase
func ForceMinipoolExit(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkExit.Transact(opts, "forceMinipoolExit", minipoolAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error forcing minipool exit for %s: %w", minipoolAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of PenaliseMinipool
func EstimatePenaliseMinipoolGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkExit.GetTransactionGasInfo(opts, "penaliseMinipool", minipoolAddress, slotTimestamp, validatorProof, slotProof)
}

// Penalise a Minipool that failed to exit cooperatively (or force-exits if the delegate supports it)
func PenaliseMinipool(rp *rocketpool.RocketPool, minipoolAddress common.Address, slotTimestamp uint64, validatorProof megapool.ValidatorProof, slotProof megapool.SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkExit.Transact(opts, "penaliseMinipool", minipoolAddress, slotTimestamp, validatorProof, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error penalising minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of RequestMegapoolExit
func EstimateRequestMegapoolExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkExit.GetTransactionGasInfo(opts, "requestMegapoolExit", megapoolAddress, validatorId)
}

// Request a Megapool validator to exit cooperatively
func RequestMegapoolExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkExit.Transact(opts, "requestMegapoolExit", megapoolAddress, validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error requesting megapool exit for %s validator %d: %w", megapoolAddress.Hex(), validatorId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ForceMegapoolExit
func EstimateForceMegapoolExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkExit.GetTransactionGasInfo(opts, "forceMegapoolExit", megapoolAddress, validatorId)
}

// Force a Megapool validator to exit after the cooperative exit phase
func ForceMegapoolExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkExit, err := getRocketNetworkExit(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkExit.Transact(opts, "forceMegapoolExit", megapoolAddress, validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error forcing megapool exit for %s validator %d: %w", megapoolAddress.Hex(), validatorId, err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketNetworkExitLock sync.Mutex

func getRocketNetworkExit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkExitLock.Lock()
	defer rocketNetworkExitLock.Unlock()
	return rp.GetContract("rocketNetworkExit", opts)
}
