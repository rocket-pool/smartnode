package megapool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

func EstimatePenaliseGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, block *big.Int, amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolPenalties, err := getRocketMegapoolPenalties(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolPenalties.GetTransactionGasInfo(opts, "penalise", megapoolAddress, block, amount)
}

func Penalise(rp *rocketpool.RocketPool, megapoolAddress common.Address, block *big.Int, amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	megapoolPenalties, err := getRocketMegapoolPenalties(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := megapoolPenalties.Transact(opts, "penalise", megapoolAddress, block, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error voting to penalise megapool: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketMegapoolPenaltiesLock sync.Mutex

func getRocketMegapoolPenalties(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMegapoolPenaltiesLock.Lock()
	defer rocketMegapoolPenaltiesLock.Unlock()
	return rp.GetContract("rocketMegapoolPenalties", opts)
}
