package minipool

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/storage"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

const (
	minipoolV3EncodedAbi string = "eJztWltv2koQ/isVz3lq1SrqW3Jon5qeCJKch6pCY+8Aqyy71l7MQdH57x0bY2Mw2MAau0d9SsDjmW8uO7flx9sApJKrhXJm8HkKwuDNgMvIWfr4443+Zfgvsq1HFrUE8bSKcPB54Ojz+4+fBjcDCYvki0hjzInXvZLsjphKS89smfi/m9P5Slz6Zmk5/dnn9DMnSASOkLmQmOZ0GCMBSOQ1s5vVbh8LMKbRmALLVKtFIWPz+Byt4JoW+mLnqIcYKcNt+0Yi2tCV5FxiKKkY+gwnZ1B7DU9lQdyDABlWOaE1d/7D7ZxpWIJ41Cok67bvWKt+19jfGEueb6JqPMHK4paFYhCcgVX60QWvuCqkrekaKHiI4ZjPJFin8VyeH94XXNk6FQzBwkgpu8OSKK/s1R2ll3lo/6WRkZs4vXWG3qcHywOXPFKKjhQaC6+XHCn/kNRCXZS9fSMah9oFwRUQBZ47itBpTUCvm7Q3VnuBkKSuKMIi0O3n7AUuAtR+8natjmlEvFwWpZVqJbJuCyB0Nq0zZSi3LamUinqOKAlXa5XrULwT4IzLZ+ozhtxYzQNqhuhN5WxBmWiAD85CwAW3q7TPkRGsIBBbeKZOhpYrWRb0VqvVJDgY2gXKEEToBMH4Th3WeA66DLJeSjX3fc1ijstr65RYv2udtmHJLH0fhxNQfqjCkn7vCcg4qWtdwxDKeD0SP3dbm3rmp3PmxigRt3uUy6afaFyCZuZvKVYVbiiAZVmmqGgtWHbKCSZvy28ztNmAmlriaHyidIt3m3o6zB1eeYRvvQQsofuaqc86PjsEJUnYd3mhPwJmrxuo6AP8Qco8cWcMzSp9sVMGqupkdFUNMmRfsVdwRkiErId2elLRc/QCwvUGFfXu34DI5n1D9cBnGhKy7t1Yn4smB0eTklqlmaLjhDLezB1Ni1P+Qru1aS3mXqjwtS/RuIb0tB6veoHoKVnU5tFU48XrwUrnxL5VzgpQfXLlFrzuM90+srwf74En03WT7QRHgxIgS53s4TpA6TRZAnuePposHZRkmypfuSU8tseeNLoOaMahZv9/fN8/qV/4F7aONA49D9ENDX18x1vDoHapW9Iw2b1mm9h25tmoYt/jjblOb7tLF+2tSEkmgnZ4GwFm3u4mpfuTZPZWbd7s54oK2MKqy5Wq2J0Qatl5LYtRm+RxbaG/rYLhr92O6VSnXaRnxzZ7PVSS3OJCiuOL79n2yrESbIgCZ2C3JJ5wYbTHUOLyIoanX9dsxI2UEMjuIR2IMnI/t5P/Z6s9RzMN7M/PqA6YKf0pyQhD5HHju0C2FZAB7MznHm89Sg5sR0i6os/jshdr5y/TKRJNjH0D9pj9vrJvuGg+/UZMTENPdjcZahW+JlskpWF2MA3+XtPlzt2eQWu5nFUYddM+1rnrPIBN+kPIklVDer3OiQT+F40AQik="
)

type MinipoolV3 interface {
	Minipool
	EstimateReduceBondAmountGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	ReduceBondAmount(opts *bind.TransactOpts) (common.Hash, error)
	EstimatePromoteGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	Promote(opts *bind.TransactOpts) (common.Hash, error)
	GetPreMigrationBalance(opts *bind.CallOpts) (*big.Int, error)
	GetUserDistributed(opts *bind.CallOpts) (bool, error)
	EstimateDistributeBalanceGas(rewardsOnly bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	DistributeBalance(rewardsOnly bool, opts *bind.TransactOpts) (common.Hash, error)
}

// Minipool contract
type minipool_v3 struct {
	Address    common.Address
	Version    uint8
	Contract   *rocketpool.Contract
	RocketPool *rocketpool.RocketPool
}

// The decoded ABI for v2 minipools
var minipoolV3Abi *abi.ABI

// Create new minipool contract
func newMinipool_v3(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Minipool, error) {

	var contract *rocketpool.Contract
	var err error
	if minipoolV3Abi == nil {
		// Get contract
		contract, err = createMinipoolContractFromEncodedAbi(rp, address, minipoolV3EncodedAbi)
	} else {
		contract, err = createMinipoolContractFromAbi(rp, address, minipoolV3Abi)
	}
	if err != nil {
		return nil, err
	} else if minipoolV3Abi == nil {
		minipoolV3Abi = contract.ABI
	}

	// Create and return
	return &minipool_v3{
		Address:    address,
		Version:    3,
		Contract:   contract,
		RocketPool: rp,
	}, nil
}

// Get the minipool as a v3 minipool if it implements the required methods
func GetMinipoolAsV3(mp Minipool) (MinipoolV3, bool) {
	castedMp, ok := mp.(MinipoolV3)
	if ok {
		return castedMp, true
	}
	return nil, false
}

// Get the contract
func (mp *minipool_v3) GetContract() *rocketpool.Contract {
	return mp.Contract
}

// Get the contract address
func (mp *minipool_v3) GetAddress() common.Address {
	return mp.Address
}

// Get the contract version
func (mp *minipool_v3) GetVersion() uint8 {
	return mp.Version
}

// Get status details
func (mp *minipool_v3) GetStatusDetails(opts *bind.CallOpts) (StatusDetails, error) {

	// Data
	var wg errgroup.Group
	var status rptypes.MinipoolStatus
	var statusBlock uint64
	var statusTime time.Time
	var isVacant bool

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		statusBlock, err = mp.GetStatusBlock(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		statusTime, err = mp.GetStatusTime(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		isVacant, err = mp.GetVacant(opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return StatusDetails{}, err
	}

	// Return
	return StatusDetails{
		Status:      status,
		StatusBlock: statusBlock,
		StatusTime:  statusTime,
		IsVacant:    isVacant,
	}, nil

}
func (mp *minipool_v3) GetStatus(opts *bind.CallOpts) (rptypes.MinipoolStatus, error) {
	status := new(uint8)
	if err := mp.Contract.Call(opts, status, "getStatus"); err != nil {
		return 0, fmt.Errorf("Could not get minipool %s status: %w", mp.Address.Hex(), err)
	}
	return rptypes.MinipoolStatus(*status), nil
}
func (mp *minipool_v3) GetStatusBlock(opts *bind.CallOpts) (uint64, error) {
	statusBlock := new(*big.Int)
	if err := mp.Contract.Call(opts, statusBlock, "getStatusBlock"); err != nil {
		return 0, fmt.Errorf("Could not get minipool %s status changed block: %w", mp.Address.Hex(), err)
	}
	return (*statusBlock).Uint64(), nil
}
func (mp *minipool_v3) GetStatusTime(opts *bind.CallOpts) (time.Time, error) {
	statusTime := new(*big.Int)
	if err := mp.Contract.Call(opts, statusTime, "getStatusTime"); err != nil {
		return time.Unix(0, 0), fmt.Errorf("Could not get minipool %s status changed time: %w", mp.Address.Hex(), err)
	}
	return time.Unix((*statusTime).Int64(), 0), nil
}
func (mp *minipool_v3) GetFinalised(opts *bind.CallOpts) (bool, error) {
	finalised := new(bool)
	if err := mp.Contract.Call(opts, finalised, "getFinalised"); err != nil {
		return false, fmt.Errorf("Could not get minipool %s finalised: %w", mp.Address.Hex(), err)
	}
	return *finalised, nil
}

// Get deposit type
func (mp *minipool_v3) GetDepositType(opts *bind.CallOpts) (rptypes.MinipoolDeposit, error) {
	depositType := new(uint8)
	if err := mp.Contract.Call(opts, depositType, "getDepositType"); err != nil {
		return 0, fmt.Errorf("Could not get minipool %s deposit type: %w", mp.Address.Hex(), err)
	}
	return rptypes.MinipoolDeposit(*depositType), nil
}

// Get node details
func (mp *minipool_v3) GetNodeDetails(opts *bind.CallOpts) (NodeDetails, error) {

	// Data
	var wg errgroup.Group
	var address common.Address
	var fee float64
	var depositBalance *big.Int
	var refundBalance *big.Int
	var depositAssigned bool

	// Load data
	wg.Go(func() error {
		var err error
		address, err = mp.GetNodeAddress(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		fee, err = mp.GetNodeFee(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		depositBalance, err = mp.GetNodeDepositBalance(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		refundBalance, err = mp.GetNodeRefundBalance(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		depositAssigned, err = mp.GetNodeDepositAssigned(opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return NodeDetails{}, err
	}

	// Return
	return NodeDetails{
		Address:         address,
		Fee:             fee,
		DepositBalance:  depositBalance,
		RefundBalance:   refundBalance,
		DepositAssigned: depositAssigned,
	}, nil

}
func (mp *minipool_v3) GetNodeAddress(opts *bind.CallOpts) (common.Address, error) {
	nodeAddress := new(common.Address)
	if err := mp.Contract.Call(opts, nodeAddress, "getNodeAddress"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get minipool %s node address: %w", mp.Address.Hex(), err)
	}
	return *nodeAddress, nil
}
func (mp *minipool_v3) GetNodeFee(opts *bind.CallOpts) (float64, error) {
	nodeFee := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeFee, "getNodeFee"); err != nil {
		return 0, fmt.Errorf("Could not get minipool %s node fee: %w", mp.Address.Hex(), err)
	}
	return eth.WeiToEth(*nodeFee), nil
}
func (mp *minipool_v3) GetNodeFeeRaw(opts *bind.CallOpts) (*big.Int, error) {
	nodeFee := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeFee, "getNodeFee"); err != nil {
		return nil, fmt.Errorf("Could not get minipool %s node fee: %w", mp.Address.Hex(), err)
	}
	return *nodeFee, nil
}
func (mp *minipool_v3) GetNodeDepositBalance(opts *bind.CallOpts) (*big.Int, error) {
	nodeDepositBalance := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeDepositBalance, "getNodeDepositBalance"); err != nil {
		return nil, fmt.Errorf("Could not get minipool %s node deposit balance: %w", mp.Address.Hex(), err)
	}
	return *nodeDepositBalance, nil
}
func (mp *minipool_v3) GetNodeRefundBalance(opts *bind.CallOpts) (*big.Int, error) {
	nodeRefundBalance := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeRefundBalance, "getNodeRefundBalance"); err != nil {
		return nil, fmt.Errorf("Could not get minipool %s node refund balance: %w", mp.Address.Hex(), err)
	}
	return *nodeRefundBalance, nil
}
func (mp *minipool_v3) GetNodeDepositAssigned(opts *bind.CallOpts) (bool, error) {
	nodeDepositAssigned := new(bool)
	if err := mp.Contract.Call(opts, nodeDepositAssigned, "getNodeDepositAssigned"); err != nil {
		return false, fmt.Errorf("Could not get minipool %s node deposit assigned status: %w", mp.Address.Hex(), err)
	}
	return *nodeDepositAssigned, nil
}
func (mp *minipool_v3) GetVacant(opts *bind.CallOpts) (bool, error) {
	isVacant := new(bool)
	if err := mp.Contract.Call(opts, isVacant, "getVacant"); err != nil {
		return false, fmt.Errorf("Could not get minipool %s vacant status: %w", mp.Address.Hex(), err)
	}
	return *isVacant, nil
}
func (mp *minipool_v3) GetPreMigrationBalance(opts *bind.CallOpts) (*big.Int, error) {
	preMigrationBalance := new(*big.Int)
	if err := mp.Contract.Call(opts, preMigrationBalance, "getPreMigrationBalance"); err != nil {
		return nil, fmt.Errorf("Could not get minipool %s pre-migration balance: %w", mp.Address.Hex(), err)
	}
	return *preMigrationBalance, nil
}

// Get user deposit details
func (mp *minipool_v3) GetUserDetails(opts *bind.CallOpts) (UserDetails, error) {

	// Data
	var wg errgroup.Group
	var depositBalance *big.Int
	var depositAssigned bool
	var depositAssignedTime time.Time

	// Load data
	wg.Go(func() error {
		var err error
		depositBalance, err = mp.GetUserDepositBalance(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		depositAssigned, err = mp.GetUserDepositAssigned(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		depositAssignedTime, err = mp.GetUserDepositAssignedTime(opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return UserDetails{}, err
	}

	// Return
	return UserDetails{
		DepositBalance:      depositBalance,
		DepositAssigned:     depositAssigned,
		DepositAssignedTime: depositAssignedTime,
	}, nil

}
func (mp *minipool_v3) GetUserDepositBalance(opts *bind.CallOpts) (*big.Int, error) {
	userDepositBalance := new(*big.Int)
	if err := mp.Contract.Call(opts, userDepositBalance, "getUserDepositBalance"); err != nil {
		return nil, fmt.Errorf("Could not get minipool %s user deposit balance: %w", mp.Address.Hex(), err)
	}
	return *userDepositBalance, nil
}
func (mp *minipool_v3) GetUserDepositAssigned(opts *bind.CallOpts) (bool, error) {
	userDepositAssigned := new(bool)
	if err := mp.Contract.Call(opts, userDepositAssigned, "getUserDepositAssigned"); err != nil {
		return false, fmt.Errorf("Could not get minipool %s user deposit assigned status: %w", mp.Address.Hex(), err)
	}
	return *userDepositAssigned, nil
}
func (mp *minipool_v3) GetUserDepositAssignedTime(opts *bind.CallOpts) (time.Time, error) {
	depositAssignedTime := new(*big.Int)
	if err := mp.Contract.Call(opts, depositAssignedTime, "getUserDepositAssignedTime"); err != nil {
		return time.Unix(0, 0), fmt.Errorf("Could not get minipool %s user deposit assigned time: %w", mp.Address.Hex(), err)
	}
	return time.Unix((*depositAssignedTime).Int64(), 0), nil
}

// Estimate the gas of Refund
func (mp *minipool_v3) EstimateRefundGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "refund")
}

// Refund node ETH from the minipool
func (mp *minipool_v3) Refund(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "refund")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not refund from minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Check if the minipool's balance has already been distributed
func (mp *minipool_v3) GetUserDistributed(opts *bind.CallOpts) (bool, error) {
	distributed := new(bool)
	if err := mp.Contract.Call(opts, distributed, "getUserDistributed"); err != nil {
		return false, fmt.Errorf("Could not get user distributed status for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *distributed, nil
}

// Estimate the gas of DistributeBalance
func (mp *minipool_v3) EstimateDistributeBalanceGas(rewardsOnly bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "distributeBalance", rewardsOnly)
}

// Distribute the minipool's ETH balance to the node operator and rETH staking pool.
func (mp *minipool_v3) DistributeBalance(rewardsOnly bool, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "distributeBalance", rewardsOnly)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not process withdrawal for minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Stake
func (mp *minipool_v3) EstimateStakeGas(validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "stake", validatorSignature[:], depositDataRoot)
}

// Progress the prelaunch minipool to staking
func (mp *minipool_v3) Stake(validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "stake", validatorSignature[:], depositDataRoot)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not stake minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Dissolve
func (mp *minipool_v3) EstimateDissolveGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "dissolve")
}

// Dissolve the initialized or prelaunch minipool
func (mp *minipool_v3) Dissolve(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "dissolve")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not dissolve minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Close
func (mp *minipool_v3) EstimateCloseGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "close")
}

// Withdraw node balances from the dissolved minipool and close it
func (mp *minipool_v3) Close(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "close")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not close minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Finalise
func (mp *minipool_v3) EstimateFinaliseGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "finalise")
}

// Finalise a minipool to get the RPL stake back
func (mp *minipool_v3) Finalise(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "finalise")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not finalise minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of DelegateUpgrade
func (mp *minipool_v3) EstimateDelegateUpgradeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "delegateUpgrade")
}

// Upgrade this minipool to the latest network delegate contract
func (mp *minipool_v3) DelegateUpgrade(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "delegateUpgrade")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not upgrade delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of DelegateRollback
func (mp *minipool_v3) EstimateDelegateRollbackGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "delegateRollback")
}

// Rollback to previous delegate contract
func (mp *minipool_v3) DelegateRollback(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "delegateRollback")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not rollback delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of SetUseLatestDelegate
func (mp *minipool_v3) EstimateSetUseLatestDelegateGas(setting bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "setUseLatestDelegate", setting)
}

// If set to true, will automatically use the latest delegate contract
func (mp *minipool_v3) SetUseLatestDelegate(setting bool, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "setUseLatestDelegate", setting)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not set use latest delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Getter for useLatestDelegate setting
func (mp *minipool_v3) GetUseLatestDelegate(opts *bind.CallOpts) (bool, error) {
	setting := new(bool)
	if err := mp.Contract.Call(opts, setting, "getUseLatestDelegate"); err != nil {
		return false, fmt.Errorf("Could not get use latest delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *setting, nil
}

// Returns the address of the minipool's stored delegate
func (mp *minipool_v3) GetDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Returns the address of the minipool's previous delegate (or address(0) if not set)
func (mp *minipool_v3) GetPreviousDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getPreviousDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get previous delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Returns the delegate which will be used when calling this minipool taking into account useLatestDelegate setting
func (mp *minipool_v3) GetEffectiveDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getEffectiveDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("Could not get effective delegate for minipool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Estimate the gas required to reduce a minipool's bond
func (mp *minipool_v3) EstimateReduceBondAmountGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "reduceBondAmount")
}

// Reduce a minipool's bond
func (mp *minipool_v3) ReduceBondAmount(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "reduceBondAmount")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not reduce bond for minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Given a validator balance, calculates how much belongs to the node taking into consideration rewards and penalties
func (mp *minipool_v3) CalculateNodeShare(balance *big.Int, opts *bind.CallOpts) (*big.Int, error) {
	nodeAmount := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeAmount, "calculateNodeShare", balance); err != nil {
		return nil, fmt.Errorf("Could not get minipool node portion: %w", err)
	}
	return *nodeAmount, nil
}

// Given a validator balance, calculates how much belongs to rETH users taking into consideration rewards and penalties
func (mp *minipool_v3) CalculateUserShare(balance *big.Int, opts *bind.CallOpts) (*big.Int, error) {
	userAmount := new(*big.Int)
	if err := mp.Contract.Call(opts, userAmount, "calculateUserShare", balance); err != nil {
		return nil, fmt.Errorf("Could not get minipool user portion: %w", err)
	}
	return *userAmount, nil
}

// Estimate the gas required to vote to scrub a minipool
func (mp *minipool_v3) EstimateVoteScrubGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "voteScrub")
}

// Vote to scrub a minipool
func (mp *minipool_v3) VoteScrub(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "voteScrub")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not vote to scrub minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to promote a vacant minipool
func (mp *minipool_v3) EstimatePromoteGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "promote")
}

// Promote a vacant minipool
func (mp *minipool_v3) Promote(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "promote")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not promote minipool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Get the data from this minipool's MinipoolPrestaked event
func (mp *minipool_v3) GetPrestakeEvent(intervalSize *big.Int, opts *bind.CallOpts) (PrestakeData, error) {

	addressFilter := []common.Address{mp.Address}
	topicFilter := [][]common.Hash{{mp.Contract.ABI.Events["MinipoolPrestaked"].ID}}

	// Grab the latest block number
	currentBlock, err := mp.RocketPool.Client.BlockNumber(context.Background())
	if err != nil {
		return PrestakeData{}, fmt.Errorf("Error getting current block %s: %w", mp.Address.Hex(), err)
	}

	// Grab the lowest block number worth querying from (should never have to go back this far in practice)
	fromBlockBig, err := storage.GetDeployBlock(mp.RocketPool)
	if err != nil {
		return PrestakeData{}, fmt.Errorf("Error getting deploy block %s: %w", mp.Address.Hex(), err)
	}

	fromBlock := fromBlockBig.Uint64()
	var log types.Log
	found := false

	// Backwards scan through blocks to find the event
	for i := currentBlock; i >= fromBlock; i -= EventScanInterval {
		from := i - EventScanInterval + 1
		if from < fromBlock {
			from = fromBlock
		}

		fromBig := big.NewInt(0).SetUint64(from)
		toBig := big.NewInt(0).SetUint64(i)

		logs, err := eth.GetLogs(mp.RocketPool, addressFilter, topicFilter, intervalSize, fromBig, toBig, nil)
		if err != nil {
			return PrestakeData{}, fmt.Errorf("Error getting prestake logs for minipool %s: %w", mp.Address.Hex(), err)
		}

		if len(logs) > 0 {
			log = logs[0]
			found = true
			break
		}
	}

	if !found {
		// This should never happen
		return PrestakeData{}, fmt.Errorf("Error finding prestake log for minipool %s", mp.Address.Hex())
	}

	// Decode the event
	prestakeEvent := new(MinipoolPrestakeEvent)
	mp.Contract.Contract.UnpackLog(prestakeEvent, "MinipoolPrestaked", log)
	if err != nil {
		return PrestakeData{}, fmt.Errorf("Error unpacking prestake data: %w", err)
	}

	// Convert the event to a more useable struct
	prestakeData := PrestakeData{
		Pubkey:                rptypes.BytesToValidatorPubkey(prestakeEvent.Pubkey),
		WithdrawalCredentials: common.BytesToHash(prestakeEvent.WithdrawalCredentials),
		Amount:                prestakeEvent.Amount,
		Signature:             rptypes.BytesToValidatorSignature(prestakeEvent.Signature),
		DepositDataRoot:       prestakeEvent.DepositDataRoot,
		Time:                  time.Unix(prestakeEvent.Time.Int64(), 0),
	}
	return prestakeData, nil
}
