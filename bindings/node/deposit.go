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

type NodeDeposit struct {
	BondAmount         *big.Int    `json:"bondAmount"`
	UseExpressTicket   bool        `json:"useExpressTicket"`
	ValidatorPubkey    []byte      `json:"validatorPubkey"`
	ValidatorSignature []byte      `json:"validatorSignature"`
	DepositDataRoot    common.Hash `json:"depositDataRoot"`
}

type Deposits []NodeDeposit

// Estimate the gas of Deposit
func EstimateDepositGas(rp *rocketpool.RocketPool, bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "deposit", bondAmount, useExpressTicket, validatorPubkey[:], validatorSignature[:], depositDataRoot)
}

// Make a node deposit
func Deposit(rp *rocketpool.RocketPool, bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "deposit", bondAmount, useExpressTicket, validatorPubkey[:], validatorSignature[:], depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error making node deposit: %w", err)
	}
	return tx, nil
}

// Estimate the gas of DepositMulti
func EstimateDepositMultiGas(rp *rocketpool.RocketPool, deposits Deposits, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "depositMulti", deposits)
}

// Make multiple node deposits
func DepositMulti(rp *rocketpool.RocketPool, deposits Deposits, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "depositMulti", deposits)
	if err != nil {
		return nil, fmt.Errorf("error making multiple node deposits: %w", err)
	}
	return tx, nil
}

// Estimate the gas to WithdrawETH
func EstimateWithdrawEthGas(rp *rocketpool.RocketPool, nodeAccount common.Address, ethAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "withdrawEth", nodeAccount, ethAmount)
}

// Withdraw unused Ether that was staked on behalf of the node
func WithdrawEth(rp *rocketpool.RocketPool, nodeAccount common.Address, ethAmount *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "withdrawEth", nodeAccount, ethAmount)
	if err != nil {
		return nil, fmt.Errorf("error trying to withdraw ETH: %w", err)
	}
	return tx, nil
}

// Estimate the gas required to withdraw credit
func EstimateWithdrawCreditGas(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDepositPool.GetTransactionGasInfo(opts, "withdrawCredit", amount)
}

// Withdraws credit store on a node as rETH
func WithdrawCredit(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketDepositPool.Transact(opts, "withdrawCredit", amount)
	if err != nil {
		return nil, fmt.Errorf("error withdrawing credit: %w", err)
	}
	return tx, nil
}

// Estimate the gas of DepositWithCredit
func EstimateDepositWithCreditGas(rp *rocketpool.RocketPool, bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "depositWithCredit", bondAmount, useExpressTicket, validatorPubkey[:], validatorSignature[:], depositDataRoot)
}

// Make a node deposit by using the credit balance
func DepositWithCredit(rp *rocketpool.RocketPool, bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "depositWithCredit", bondAmount, useExpressTicket, validatorPubkey[:], validatorSignature[:], depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error making node deposit with credit: %w", err)
	}
	return tx, nil
}

// Estimate the gas of CreateVacantMinipool
func EstimateCreateVacantMinipoolGas(rp *rocketpool.RocketPool, bondAmount *big.Int, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, salt *big.Int, expectedMinipoolAddress common.Address, currentBalance *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "createVacantMinipool", bondAmount, eth.EthToWei(minimumNodeFee), validatorPubkey[:], salt, expectedMinipoolAddress, currentBalance)
}

// Make a vacant minipool for solo staker migration
func CreateVacantMinipool(rp *rocketpool.RocketPool, bondAmount *big.Int, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, salt *big.Int, expectedMinipoolAddress common.Address, currentBalance *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "createVacantMinipool", bondAmount, eth.EthToWei(minimumNodeFee), validatorPubkey[:], salt, expectedMinipoolAddress, currentBalance)
	if err != nil {
		return nil, fmt.Errorf("error creating vacant minipool: %w", err)
	}
	return tx, nil
}

// Get the amount of ETH in the node's deposit credit bank
func GetNodeDepositCredit(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return nil, err
	}

	creditBalance := new(*big.Int)
	if err := rocketNodeDeposit.Call(opts, creditBalance, "getNodeDepositCredit", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node deposit credit: %w", err)
	}
	return *creditBalance, nil
}

// Get the current ETH balance for the given node operator
func GetNodeEthBalance(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return nil, err
	}

	creditBalance := new(*big.Int)
	if err := rocketNodeDeposit.Call(opts, creditBalance, "getNodeEthBalance", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH balance: %w", err)
	}
	return *creditBalance, nil
}

// Get the sum of the credit balance of a given node operator and their ETH balance
func GetNodeCreditAndBalance(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return nil, err
	}

	creditAndBalance := new(*big.Int)
	if err := rocketNodeDeposit.Call(opts, creditAndBalance, "getNodeCreditAndBalance", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node credit and ETH balance: %w", err)
	}
	return *creditAndBalance, nil
}

// Get the sum of the amount of ETH credit currently usable by a given node operator and their balance
func GetNodeUsableCreditAndBalance(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return nil, err
	}

	usableCreditBalance := new(*big.Int)
	if err := rocketNodeDeposit.Call(opts, usableCreditBalance, "getNodeUsableCreditAndBalance", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node usable credit and ETH balance: %w", err)
	}
	return *usableCreditBalance, nil
}

// Get the amount of ETH credit currently usable by a given node operator
func GetNodeUsableCredit(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return nil, err
	}

	usableCredit := new(*big.Int)
	if err := rocketNodeDeposit.Call(opts, usableCredit, "getNodeUsableCredit", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node usable credit: %w", err)
	}
	return *usableCredit, nil
}

func GetBondRequirement(rp *rocketpool.RocketPool, numValidators *big.Int, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return nil, err
	}

	bondRequirement := new(*big.Int)
	if err := rocketNodeDeposit.Call(opts, bondRequirement, "getBondRequirement", numValidators); err != nil {
		return nil, fmt.Errorf("error getting the bond requirement: %w", err)
	}
	return *bondRequirement, nil
}

// Get contracts
var rocketNodeDepositLock sync.Mutex

func getRocketNodeDeposit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeDepositLock.Lock()
	defer rocketNodeDepositLock.Unlock()
	return rp.GetContract("rocketNodeDeposit", opts)
}

var rocketDepositPoolLock sync.Mutex

func getRocketDepositPool(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDepositPoolLock.Lock()
	defer rocketDepositPoolLock.Unlock()
	return rp.GetContract("rocketDepositPool", opts)
}
