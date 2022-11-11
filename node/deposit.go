package node

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Estimate the gas of Deposit
func EstimateDepositGas(rp *rocketpool.RocketPool, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, salt *big.Int, expectedMinipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeDeposit.GetTransactionGasInfo(opts, "deposit", eth.EthToWei(minimumNodeFee), validatorPubkey[:], validatorSignature[:], depositDataRoot, salt, expectedMinipoolAddress)
}

// Make a node deposit
func Deposit(rp *rocketpool.RocketPool, minimumNodeFee float64, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, salt *big.Int, expectedMinipoolAddress common.Address, opts *bind.TransactOpts) (*types.Transaction, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := rocketNodeDeposit.Transact(opts, "deposit", eth.EthToWei(minimumNodeFee), validatorPubkey[:], validatorSignature[:], depositDataRoot, salt, expectedMinipoolAddress)
	if err != nil {
		return nil, fmt.Errorf("Could not make node deposit: %w", err)
	}
	return tx, nil
}

// Get the type of a deposit based on the amount
func GetDepositType(rp *rocketpool.RocketPool, amount *big.Int, opts *bind.CallOpts) (rptypes.MinipoolDeposit, error) {
	rocketNodeDeposit, err := getRocketNodeDeposit(rp, opts)
	if err != nil {
		return rptypes.Empty, err
	}

	depositType := new(uint8)
	if err := rocketNodeDeposit.Call(opts, depositType, "getDepositType", amount); err != nil {
		return rptypes.Empty, fmt.Errorf("Could not get deposit type: %w", err)
	}
	return rptypes.MinipoolDeposit(*depositType), nil
}

// Get contracts
var rocketNodeDepositLock sync.Mutex

func getRocketNodeDeposit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeDepositLock.Lock()
	defer rocketNodeDepositLock.Unlock()
	return rp.GetContract("rocketNodeDeposit", opts)
}
