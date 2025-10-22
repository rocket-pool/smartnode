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

type ExitChallenge struct {
	Megapool     common.Address `json:"megapool"`
	ValidatorIds []uint32       `json:"validatorIds"`
}

type WithdrawalProof struct {
	Slot           uint64     `json:"slot"`
	WithdrawalSlot uint64     `json:"withdrawalSlot"`
	WithdrawalNum  uint16     `json:"withdrawalNum"`
	Withdrawal     Withdrawal `json:"withdrawal"`
	Witnesses      [][32]byte `json:"witnesses"`
}

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

	validator := new(ValidatorInfoFromGlobalIndex)

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

	src := iface[1].(struct {
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
		Locked             bool   `json:"locked"`
		ValidatorIndex     uint64 `json:"validatorIndex"`
		ExitBalance        uint64 `json:"exitBalance"`
		LockedSlot         uint64 `json:"lockedSlot"`
	})
	// validatorInfo.ValidatorInfo.PubKey = make([]byte, len(src.PubKey))
	// copy(validatorInfo.ValidatorInfo.PubKey[:], src.PubKey)
	validator.Pubkey = iface[0].([]byte)
	validator.ValidatorInfo.LastAssignmentTime = src.LastAssignmentTime
	validator.ValidatorInfo.LastRequestedValue = src.LastRequestedValue
	validator.ValidatorInfo.LastRequestedBond = src.LastRequestedBond
	validator.ValidatorInfo.Staked = src.Staked
	validator.ValidatorInfo.DepositValue = src.DepositValue
	validator.ValidatorInfo.ExitBalance = src.ExitBalance
	validator.ValidatorInfo.Exiting = src.Exiting
	validator.ValidatorInfo.ValidatorIndex = src.ValidatorIndex
	validator.ValidatorInfo.Exited = src.Exited
	validator.ValidatorInfo.InQueue = src.InQueue
	validator.ValidatorInfo.InPrestake = src.InPrestake
	validator.ValidatorInfo.ExpressUsed = src.ExpressUsed
	validator.ValidatorInfo.Dissolved = src.Dissolved
	validator.ValidatorInfo.Locked = src.Locked
	validator.ValidatorInfo.LockedSlot = src.LockedSlot
	validator.MegapoolAddress = iface[2].(common.Address)
	validator.ValidatorId = iface[3].(uint32)

	return *validator, nil
}

// Estimate the gas of Stake
func EstimateStakeGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "stake", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Progress the prelaunch megapool to staking
func Stake(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
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
func EstimateNotifyExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyExit", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Notify the megapool that one of its validators is exiting
func NotifyExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
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
func EstimateNotifyNotExitGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyNotExit", megapoolAddress, validatorId, slotTimestamp, validatorProof)
}

// Used to prove a validator is not exiting after a challenge-exit
func NotifyNotExit(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "notifyNotExit", megapoolAddress, slotTimestamp, validatorId, validatorProof)
	if err != nil {
		return nil, fmt.Errorf("error calling notify not exit: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call ChallengeExit
func EstimateChallengeExitGas(rp *rocketpool.RocketPool, exitChallenge []ExitChallenge, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "challengeExit", exitChallenge)
}

// Used to challenge a validator that is exiting without an exit notification
func ChallengeExit(rp *rocketpool.RocketPool, exitChallenge []ExitChallenge, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "challengeExit", exitChallenge)
	if err != nil {
		return nil, fmt.Errorf("error calling challengeExit: %w", err)
	}
	return tx, nil
}

// Estimate the gas to call NotifyFinalBalance
func EstimateNotifyFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, withdrawalProof WithdrawalProof, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "notifyFinalBalance", megapoolAddress, validatorId, slotTimestamp, withdrawalProof, validatorProof, slotProof)
}

// Notify the megapool of the final balance of an exited validator
func NotifyFinalBalance(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, withdrawalProof WithdrawalProof, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
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
func EstimateDissolveWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return megapoolManager.GetTransactionGasInfo(opts, "dissolve", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Dissolve a validator using a proof that it used wrong credentials
func DissolveWithProof(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (*types.Transaction, error) {
	megapoolManager, err := getRocketMegapoolManager(rp, nil)
	if err != nil {
		return nil, err
	}
	tx, err := megapoolManager.Transact(opts, "dissolve", megapoolAddress, validatorId, slotTimestamp, validatorProof, slotProof)
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
