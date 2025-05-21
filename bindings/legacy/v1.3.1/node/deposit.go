package node

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
)

// Estimate the gas of Deposit
func EstimateDepositGas(rp *rocketpool.RocketPool, bondAmount *big.Int, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, salt *big.Int, expectedMinipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "deposit", bondAmount, eth.EthToWei(minimumNodeFee), validatorPubkey[:], validatorSignature[:], depositDataRoot, salt, expectedMinipoolAddress)
}

// Make a node deposit
func Deposit(rp *rocketpool.RocketPool, bondAmount *big.Int, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, salt *big.Int, expectedMinipoolAddress common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "deposit", bondAmount, eth.EthToWei(minimumNodeFee), validatorPubkey[:], validatorSignature[:], depositDataRoot, salt, expectedMinipoolAddress)
	if err != nil {
		return nil, fmt.Errorf("error making node deposit: %w", err)
	}
	return tx, nil
}

// Estimate the gas of DepositWithCredit
func EstimateDepositWithCreditGas(rp *rocketpool.RocketPool, bondAmount *big.Int, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, salt *big.Int, expectedMinipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "depositWithCredit", bondAmount, eth.EthToWei(minimumNodeFee), validatorPubkey[:], validatorSignature[:], depositDataRoot, salt, expectedMinipoolAddress)
}

// Make a node deposit by using the credit balance
func DepositWithCredit(rp *rocketpool.RocketPool, bondAmount *big.Int, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, salt *big.Int, expectedMinipoolAddress common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "depositWithCredit", bondAmount, eth.EthToWei(minimumNodeFee), validatorPubkey[:], validatorSignature[:], depositDataRoot, salt, expectedMinipoolAddress)
	if err != nil {
		return nil, fmt.Errorf("error making node deposit with credit: %w", err)
	}
	return tx, nil
}

// Get contracts
var rocketNodeDepositLock sync.Mutex

func getRocketNodeDeposit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeDepositLock.Lock()
	defer rocketNodeDepositLock.Unlock()
	return rp.GetContract("rocketNodeDeposit", opts)
}

// Estimate the gas of AssignDeposits
func EstimateAssignDepositsGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDepositPool.GetTransactionGasInfo(opts, "assignDeposits")
}

// Assign deposits
func AssignDeposits(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDepositPool.Transact(opts, "assignDeposits")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error assigning deposits: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDepositPoolLock sync.Mutex

func getRocketDepositPool(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDepositPoolLock.Lock()
	defer rocketDepositPoolLock.Unlock()
	return rp.GetContract("rocketDepositPool", opts)
}
