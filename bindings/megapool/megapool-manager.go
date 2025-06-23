package megapool

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
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

	// Both Call and UnpackIntoStruct were not working with this response (which contains a struct inside a struct)
	// For the moment this was the only way for it to work. We should investigate further.
	iface, err := megapoolManager.ABI.Unpack("getValidatorInfo", response)
	if err != nil {
		return ValidatorInfoFromGlobalIndex{}, fmt.Errorf("error unpacking getValidatorInfo response: %w", err)
	}

	src := iface[0].(struct {
		PubKey             []byte `json:"pubKey"`
		LastAssignmentTime uint32 `json:"lastAssignmentTime"`
		LastRequestedValue uint32 `json:"lastRequestedValue"`
		LastRequestedBond  uint32 `json:"lastRequestedBond"`
		DepositValue       uint32 `json:"depositValue"`
		Staked             bool   `json:"staked"`
		Exited             bool   `json:"exited"`
		InQueue            bool   `json:"inQueue"`
		InPrestake         bool   `json:"inPrestake"`
		ExpressUsed        bool   `json:"expressUsed"`
		Dissolved          bool   `json:"dissolved"`
		Exiting            bool   `json:"exiting"`
		ValidatorIndex     uint64 `json:"validatorIndex"`
		ExitBalance        uint64 `json:"exitBalance"`
		WithdrawableEpoch  uint64 `json:"withdrawableEpoch"`
	})
	validatorInfo.ValidatorInfo.PubKey = make([]byte, len(src.PubKey))
	copy(validatorInfo.ValidatorInfo.PubKey[:], src.PubKey)
	validatorInfo.ValidatorInfo.LastAssignmentTime = src.LastAssignmentTime
	validatorInfo.ValidatorInfo.LastRequestedValue = src.LastRequestedValue
	validatorInfo.ValidatorInfo.LastRequestedBond = src.LastRequestedBond
	validatorInfo.ValidatorInfo.Staked = src.Staked
	validatorInfo.ValidatorInfo.DepositValue = src.DepositValue
	validatorInfo.ValidatorInfo.ExitBalance = src.ExitBalance
	validatorInfo.ValidatorInfo.WithdrawableEpoch = src.WithdrawableEpoch
	validatorInfo.ValidatorInfo.Exiting = src.Exiting
	validatorInfo.ValidatorInfo.ValidatorIndex = src.ValidatorIndex
	validatorInfo.ValidatorInfo.Exited = src.Exited
	validatorInfo.ValidatorInfo.InQueue = src.InQueue
	validatorInfo.ValidatorInfo.InPrestake = src.InPrestake
	validatorInfo.ValidatorInfo.ExpressUsed = src.ExpressUsed
	validatorInfo.ValidatorInfo.Dissolved = src.Dissolved
	validatorInfo.MegapoolAddress = iface[1].(common.Address)
	validatorInfo.ValidatorId = iface[2].(uint32)

	return *validatorInfo, nil
}

// Estimate the gas of Stake
func EstimateStakeGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, validatorProof ValidatorProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "stake", megapoolAddress, validatorId, validatorProof)
}

// Progress the prelaunch megapool to staking
func Stake(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, validatorProof ValidatorProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "stake", megapoolAddress, validatorId, validatorProof)
	if err != nil {
		return nil, fmt.Errorf("error staking megapool %s: %w", megapoolAddress, err)
	}
	return tx, nil
}

// Estimate the gas to call NotifyExit
func EstimateNotifyExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, withdrawalEpoch uint64, slot uint64, exitProof [][32]byte, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyExit", megapoolAddress, validatorId, withdrawalEpoch, slot, exitProof)
}

// Notify the megapool that one of its validators is exiting
func NotifyExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, withdrawalEpoch uint64, slot uint64, exitProof [][32]byte, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "notifyExit", megapoolAddress, validatorId, withdrawalEpoch, slot, exitProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify exit: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call NotifyFinalBalance
func EstimateNotifyFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, withdrawalSlot uint64, withdrawalNum *big.Int, withdrawal Withdrawal, slot uint64, exitProof [][32]byte, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyFinalBalance", megapoolAddress, validatorId, withdrawalSlot, withdrawalNum, withdrawal, slot, exitProof)
}

// Notify the megapool of the final balance of an exited validator
func NotifyFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, withdrawalSlot uint64, withdrawalNum *big.Int, withdrawal Withdrawal, slot uint64, exitProof [][32]byte, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "notifyFinalBalance", megapoolAddress, validatorId, withdrawalSlot, withdrawalNum, withdrawal, slot, exitProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify final balance: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call DissolveWithProof
func EstimateDissolveWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, validatorProof ValidatorProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "dissolve", megapoolAddress, validatorId, validatorProof)
}

// Dissolve a validator using a proof that it used wrong credentials
func DissolveWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, validatorProof ValidatorProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "dissolve", megapoolAddress, validatorId, validatorProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify final balance: %w", err)
	}
	return tx, nil
}

// Get contracts
var rocketMegapoolManagerLock sync.Mutex

func getRocketMegapoolManager(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMegapoolManagerLock.Lock()
	defer rocketMegapoolManagerLock.Unlock()
	return rp.GetContract("rocketMegapoolManager", opts)
}
