package megapool

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func GetValidatorCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint32, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, opts)
	if err != nil {
		return 0, err
	}
	var validatorCount *big.Int
	if err := megapoolManager.Call(opts, &validatorCount, "getValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool manager validator count: %w", err)
	}
	return uint32((*validatorCount).Uint64()), nil
}

func GetValidatorInfo(rp *rocketpool.RocketPool, index uint32, opts *bind.CallOpts) (ValidatorInfoFromGlobalIndex, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, opts)
	if err != nil {
		return ValidatorInfoFromGlobalIndex{}, err
	}

	validatorInfo := new(ValidatorInfoFromGlobalIndex)

	indexBig := new(big.Int).SetUint64(uint64(index))

	callData, err := megapoolManager.ABI.Pack("getValidatorInfo", indexBig)
	if err != nil {
		return ValidatorInfoFromGlobalIndex{}, fmt.Errorf("error creating calldata for getValidatorInfo: %w", err)
	}

	response, err := megapoolManager.Client.CallContract(context.Background(), ethereum.CallMsg{To: megapoolManager.Address, Data: callData}, nil)
	if err != nil {
		return ValidatorInfoFromGlobalIndex{}, fmt.Errorf("error calling getValidatorInfo: %w", err)
	}

	err = megapoolManager.ABI.UnpackIntoInterface(&validatorInfo, "getValidatorInfo", response)
	if err != nil {
		return ValidatorInfoFromGlobalIndex{}, fmt.Errorf("error unpacking getValidatorInfo response: %w", err)
	}

	return *validatorInfo, nil
}

// Get contracts
var rocketMegapoolManagerLock sync.Mutex

func getRocketMegapoolManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMegapoolManagerLock.Lock()
	defer rocketMegapoolManagerLock.Unlock()
	return rp.GetContract("rocketMegapoolManager", opts)
}
