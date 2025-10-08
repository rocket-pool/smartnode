package megapool

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"golang.org/x/sync/errgroup"
)

type SlotProof struct {
	Slot      uint64
	Witnesses [][32]byte
}

type ValidatorProof struct {
	ValidatorIndex *big.Int
	Validator      ProvedValidator
	Witnesses      [][32]byte
}

type ProvedValidator struct {
	Pubkey                     []byte   `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials      [32]byte `json:"withdrawal_credentials" ssz-size:"32"`
	EffectiveBalance           uint64   `json:"effective_balance"`
	Slashed                    bool     `json:"slashed"`
	ActivationEligibilityEpoch uint64   `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64   `json:"activation_epoch"`
	ExitEpoch                  uint64   `json:"exit_epoch"`
	WithdrawableEpoch          uint64   `json:"withdrawable_epoch"`
}

type FinalBalanceProof struct {
	WithdrawalSlot uint64
	ValidatorIndex uint64
	Amount         *big.Int
	Witnesses      [][32]byte

	// Contract refers to this as _withdrawalNum
	IndexInWithdrawalsArray uint
	// Part of the Withdrawal calldata
	WithdrawalIndex   uint64
	WithdrawalAddress common.Address
}

type Withdrawal struct {
	Index                 uint64   `json:"index"`
	ValidatorIndex        uint64   `json:"validatorIndex"`
	WithdrawalCredentials [20]byte `json:"withdrawalCredentials"`
	AmountInGwei          uint64   `json:"amountInGwei"`
}

type RewardSplit struct {
	NodeRewards        *big.Int `abi:"nodeRewards"`
	VoterRewards       *big.Int `abi:"voterRewards"`
	ProtocolDAORewards *big.Int `abi:"protocolDAORewards"`
	RethRewards        *big.Int `abi:"rethRewards"`
}

type MegapoolV1 interface {
	Megapool
}

type ValidatorInfo struct {
	LastAssignmentTime uint32 `abi:"lastAssignmentTime"`
	LastRequestedValue uint32 `abi:"lastRequestedValue"`
	LastRequestedBond  uint32 `abi:"lastRequestedBond"`
	DepositValue       uint32 `abi:"depositValue"`
	Staked             bool   `abi:"staked"`
	Exited             bool   `abi:"exited"`
	InQueue            bool   `abi:"inQueue"`
	InPrestake         bool   `abi:"inPrestake"`
	ExpressUsed        bool   `abi:"expressUsed"`
	Dissolved          bool   `abi:"dissolved"`
	Exiting            bool   `abi:"exiting"`
	Locked             bool   `abi:"locked"`
	ValidatorIndex     uint64 `abi:"validatorIndex"`
	ExitBalance        uint64 `abi:"exitBalance"`
	WithdrawableEpoch  uint64 `abi:"withdrawableEpoch"`
	LockedSlot         uint64 `abi:"lockedSlot"`
}

type ValidatorInfoFromGlobalIndex struct {
	Pubkey          []byte         `abi:"pubkey"`
	ValidatorInfo   ValidatorInfo  `abi:"validatorInfo"`
	MegapoolAddress common.Address `abi:"megapoolAddress"`
	ValidatorId     uint32         `abi:"validatorId"`
}

// Megapool contract
type megapoolV1 struct {
	Address    common.Address
	Version    uint8
	Contract   *rocketpool.Contract
	RocketPool *rocketpool.RocketPool
}

const (
	megapoolV1EncodedAbi string = "eJztWktv2zgQ/isLn4M9tLs99OYk7iJA2k2dePdQFAYtjm0iNKmSIztGsf99h7b1smRLsiRDjXNqHZHDbx6cJ7/97DGl1XqhA9v7OGXSwlVPKD9A+vntJ/2XwwvwxCcEo5h8WvvQ+9gL6Pe7Pz/0rnqKLdwfGBFSSL8xveC/q+q0UNA/WUrfowWfYcZ8reW1VnwIPPCIeLQelkBA3Lm/PoO3MME75Rlg9jWz+Kp1+ABEBdd935eiDo9oghKwlkwKzlCbO35mPv8JT+5bK2bqIli9hR8BBJfBqrBWy+VF8DpQran1/bsSnNKiMzH6IvAyNEqMCjW7AE7vtfd8ESp9RHYZjI6UrKnTLqZHQ1gxw+2NZLSyfd6U5tBvkL+lpu9NEjSA8ybp+Uaj9ii/ZrpJsmX1SukCGjEJDoSXlCaPnzg+aJbxoYyS6/Uu0aaFOsAdeVpikSF8DpBNhBT02dmC8tmaTWSCk2mgPBRalUCXDOHjozE8AXCTE38KFLfF+EqCi6l7THqBJDIkAk5hbqeD1EnFgnZXJN6Zaysl7kQtCpHR9v+uRcddpsMEcoW+FLA6yRwqG2ukrTc1nVFNp9xab86kBDUDlz426VcSR7j41w5pDr4Bjyh10iPybencTWy7WjdKwVpSUBQj26E/A+zTggQfNzvnVOBvkjLLldNp9zKNbNclImx7VlDsN5p0FilQrhnZFSy7krWLuhu8+MIwt+hpmw52QmD3zGKUdtLSa1c2dQbcpoTrojK/uMqIcwO2IBNh0aIMovBTg5DcbKcrynN4vm46gl1CVSvZbw3VEGhRt7z6yIK5Yb5AJrsEaWtRHQPWEQ/VWCqXZOlOTfUeR55e+FqBKneopAizzVsWtOUp0wqJ+vaFVIYu+7UYpz/1qOzcUmUiVCJoK7A0iInWMt5t9zqwm6+Fu2Bv5lBul1Bfd+VCxW0PFJgczuowfRfSRrY6Vp4dlZUXTWpKUW7bfnv40C6n1w9/JBoB8cXg8JLWP60rQcEBvmaSKQ9O2b4F/ig1Znd/z/QiKbMLPPxt6DZh2CV/JPhsBvdsTdf69/2rnnFMGPhUQXXQL/UVfwgmz5BuX745qDcH9eagXr+DEttfaSeVI9U1QqL+8kOHgcnPHXNuuV7tOF/NcZSC9K/AOTdsxeSNAU7+UTBZUDZtzs9PccNPrU4TJuQ2j0/vjty6cWBhsHUST8KZZIlrm1ZErO+HfFOrQOGRghLDwEBZIinD28WBW4ZsqJO3MamGcEACq1YatzWvR7G3Ga9CCyU0A19781N81tiAR8Z9YH9CTBrFdN30ZKN9IW0nbHfqrxWIMvLJtK3GnpvqmHhv1L2qoiJ5JgV9EkQpDmG/kKIshU2XghLKhV/M6BeNZzHGyvNas3mdnOn8NTamMUBbMpOHOi8D2tdtUWYWcxfmti0IbgnGus9nb53VeZmUcUda8luQlC0iHHJJ1QhS/KtFsPqbn/C4kT8zjDf+KC/D4dToRTOsnfXJ2wDn7uGFB6LWM+p8OJnKF/MryoZ5ovLzno6yOPLJH+SzVcaPUUieMO/58FuGlIW14082A+Do4nRiEhbi2UxcoWD0lDaBfe03NPqdTmHzpqBrkooMsRywlmRVIvtUqXlrjjCinoBwFaqw7aZ9ezVjruNIRPMiQZ8Gr4SHMFu/SWD+B1DmSQo="
)

// The decoded ABI for megapools
var megapoolV1Abi *abi.ABI

// Create new megapool contract
func NewMegaPoolV1(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Megapool, error) {

	var contract *rocketpool.Contract
	var err error
	if megapoolV1Abi == nil {
		// Get contract
		contract, err = createMegapoolContractFromEncodedAbi(rp, address, megapoolV1EncodedAbi)
	} else {
		contract, err = createMegapoolContractFromAbi(rp, address, megapoolV1Abi)
	}
	if err != nil {
		return nil, err
	} else if megapoolV1Abi == nil {
		megapoolV1Abi = contract.ABI
	}

	// Create and return
	return &megapoolV1{
		Address:    address,
		Version:    1,
		Contract:   contract,
		RocketPool: rp,
	}, nil
}

// Get the contract
func (mp *megapoolV1) GetContract() *rocketpool.Contract {
	return mp.Contract
}

// Get the contract address
func (mp *megapoolV1) GetAddress() common.Address {
	return mp.Address
}

// Get the contract version
func (mp *megapoolV1) GetVersion() uint8 {
	return mp.Version
}

// Get the count of all validators on a megapool
func (mp *megapoolV1) GetValidatorCount(opts *bind.CallOpts) (uint32, error) {
	var validatorCount uint32
	if err := mp.Contract.Call(opts, &validatorCount, "getValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s validator count: %w", mp.Address.Hex(), err)
	}
	return validatorCount, nil
}

// Get the count of validators on a megapool, excluding inactive validators
func (mp *megapoolV1) GetActiveValidatorCount(opts *bind.CallOpts) (uint32, error) {
	var validatorCount uint32
	if err := mp.Contract.Call(opts, &validatorCount, "getActiveValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s active validator count: %w", mp.Address.Hex(), err)
	}
	return validatorCount, nil
}

// Get the count of validators on a megapool, excluding inactive validators
func (mp *megapoolV1) GetLockedValidatorCount(opts *bind.CallOpts) (uint32, error) {
	var validatorCount uint32
	if err := mp.Contract.Call(opts, &validatorCount, "getLockedValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s locked validator count: %w", mp.Address.Hex(), err)
	}
	return validatorCount, nil
}

func (mp *megapoolV1) GetValidatorInfo(validatorId uint32, opts *bind.CallOpts) (ValidatorInfo, error) {
	validatorInfo := new(ValidatorInfo)

	callData, err := mp.Contract.ABI.Pack("getValidatorInfo", validatorId)
	if err != nil {
		return ValidatorInfo{}, fmt.Errorf("error creating calldata for getValidatorInfo: %w", err)
	}

	response, err := mp.Contract.Client.CallContract(context.Background(), ethereum.CallMsg{To: mp.Contract.Address, Data: callData}, nil)
	if err != nil {
		return ValidatorInfo{}, fmt.Errorf("error calling getValidatorInfo: %w", err)
	}

	err = mp.Contract.ABI.UnpackIntoInterface(&validatorInfo, "getValidatorInfo", response)
	if err != nil {
		return ValidatorInfo{}, fmt.Errorf("error unpacking getValidatorInfo response: %w", err)
	}

	return *validatorInfo, nil
}

func (mp *megapoolV1) GetValidatorPubkey(validatorId uint32, opts *bind.CallOpts) (rptypes.ValidatorPubkey, error) {
	pubkey := new(rptypes.ValidatorPubkey)

	callData, err := mp.Contract.ABI.Pack("getValidatorPubkey", validatorId)
	if err != nil {
		return rptypes.ValidatorPubkey{}, fmt.Errorf("error creating calldata for getValidatorPubkey: %w", err)
	}

	response, err := mp.Contract.Client.CallContract(context.Background(), ethereum.CallMsg{To: mp.Contract.Address, Data: callData}, nil)
	if err != nil {
		return rptypes.ValidatorPubkey{}, fmt.Errorf("error calling getValidatorPubkey: %w", err)
	}

	err = mp.Contract.ABI.UnpackIntoInterface(&pubkey, "getValidatorPubkey", response)
	if err != nil {
		return rptypes.ValidatorPubkey{}, fmt.Errorf("error unpacking getValidatorPubkey response: %w", err)
	}

	return *pubkey, nil
}

type ValidatorInfoWithPubkey struct {
	Pubkey        []byte `abi:"pubkey"`
	ValidatorInfo `abi:"validatorInfo"`
}

func (mp *megapoolV1) GetValidatorInfoAndPubkey(validatorId uint32, opts *bind.CallOpts) (ValidatorInfoWithPubkey, error) {

	validator := new(ValidatorInfoWithPubkey)

	callData, err := mp.Contract.ABI.Pack("getValidatorInfoAndPubkey", validatorId)
	if err != nil {
		return ValidatorInfoWithPubkey{}, fmt.Errorf("error creating calldata for getValidatorInfoAndPubkey: %w", err)
	}

	response, err := mp.Contract.Client.CallContract(context.Background(), ethereum.CallMsg{To: mp.Contract.Address, Data: callData}, nil)
	if err != nil {
		return ValidatorInfoWithPubkey{}, fmt.Errorf("error calling getValidatorInfoAndPubkey: %w", err)
	}

	iface, err := mp.Contract.ABI.Unpack("getValidatorInfoAndPubkey", response)
	if err != nil {
		return ValidatorInfoWithPubkey{}, fmt.Errorf("error unpacking getValidatorInfoAndPubkey response: %w", err)
	}

	src := iface[0].(struct {
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
		WithdrawableEpoch  uint64 `json:"withdrawableEpoch"`
		LockedSlot         uint64 `json:"lockedSlot"`
	})
	// validatorInfo.ValidatorInfo.PubKey = make([]byte, len(src.PubKey))
	// copy(validatorInfo.ValidatorInfo.PubKey[:], src.PubKey)
	validator.Pubkey = iface[1].([]byte)
	validator.ValidatorInfo.LastAssignmentTime = src.LastAssignmentTime
	validator.ValidatorInfo.LastRequestedValue = src.LastRequestedValue
	validator.ValidatorInfo.LastRequestedBond = src.LastRequestedBond
	validator.ValidatorInfo.Staked = src.Staked
	validator.ValidatorInfo.DepositValue = src.DepositValue
	validator.ValidatorInfo.ExitBalance = src.ExitBalance
	validator.ValidatorInfo.WithdrawableEpoch = src.WithdrawableEpoch
	validator.ValidatorInfo.Exiting = src.Exiting
	validator.ValidatorInfo.ValidatorIndex = src.ValidatorIndex
	validator.ValidatorInfo.Exited = src.Exited
	validator.ValidatorInfo.InQueue = src.InQueue
	validator.ValidatorInfo.InPrestake = src.InPrestake
	validator.ValidatorInfo.ExpressUsed = src.ExpressUsed
	validator.ValidatorInfo.Dissolved = src.Dissolved
	validator.ValidatorInfo.Locked = src.Locked
	validator.ValidatorInfo.LockedSlot = src.LockedSlot

	return *validator, nil
}

// Get the number of validators currently exiting
func (mp *megapoolV1) GetExitingValidatorCount(opts *bind.CallOpts) (uint32, error) {
	var exitingValidatorCount uint32
	if err := mp.Contract.Call(opts, &exitingValidatorCount, "getExitingValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s exiting validator count: %w", mp.Address.Hex(), err)
	}
	return exitingValidatorCount, nil
}

// Gets the soonest epoch a validator within this megapool can be withdrawn
func (mp *megapoolV1) GetSoonestWithdrawableEpoch(opts *bind.CallOpts) (uint32, error) {
	var soonestWithdrawableEpoch uint32
	if err := mp.Contract.Call(opts, &soonestWithdrawableEpoch, "getSoonestWithdrawableEpoch"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s soonest withdrawable epoch: %w", mp.Address.Hex(), err)
	}
	return soonestWithdrawableEpoch, nil
}

func (mp *megapoolV1) GetLastDistributionBlock(opts *bind.CallOpts) (uint64, error) {
	lastDistributionBlock := new(*big.Int)
	if err := mp.Contract.Call(opts, lastDistributionBlock, "getLastDistributionBlock"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s lastDistributionBlock: %w", mp.Address.Hex(), err)
	}
	return (*lastDistributionBlock).Uint64(), nil
}

func (mp *megapoolV1) GetAssignedValue(opts *bind.CallOpts) (*big.Int, error) {
	assignedValue := new(*big.Int)
	if err := mp.Contract.Call(opts, assignedValue, "getAssignedValue"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s assigned value: %w", mp.Address.Hex(), err)
	}
	return *assignedValue, nil
}

func (mp *megapoolV1) GetDebt(opts *bind.CallOpts) (*big.Int, error) {
	debt := new(*big.Int)
	if err := mp.Contract.Call(opts, debt, "getDebt"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s debt: %w", mp.Address.Hex(), err)
	}
	return *debt, nil
}

func (mp *megapoolV1) GetRefundValue(opts *bind.CallOpts) (*big.Int, error) {
	refundValue := new(*big.Int)
	if err := mp.Contract.Call(opts, refundValue, "getRefundValue"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s refund value: %w", mp.Address.Hex(), err)
	}
	return *refundValue, nil
}

func (mp *megapoolV1) GetNodeBond(opts *bind.CallOpts) (*big.Int, error) {
	nodeBond := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeBond, "getNodeBond"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s debt: %w", mp.Address.Hex(), err)
	}
	return *nodeBond, nil
}

func (mp *megapoolV1) GetUserCapital(opts *bind.CallOpts) (*big.Int, error) {
	userCapital := new(*big.Int)
	if err := mp.Contract.Call(opts, userCapital, "getUserCapital"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s user capital: %w", mp.Address.Hex(), err)
	}
	return *userCapital, nil
}

func (mp *megapoolV1) CalculatePendingRewards(opts *bind.CallOpts) (RewardSplit, error) {
	rewardSplits := new(RewardSplit)
	if err := mp.Contract.Call(opts, rewardSplits, "calculatePendingRewards"); err != nil {
		return RewardSplit{}, fmt.Errorf("error calculating the pending rewards for megapool %s: %w", mp.Address.Hex(), err)
	}
	return *rewardSplits, nil
}

func (mp *megapoolV1) CalculateRewards(amount *big.Int, opts *bind.CallOpts) (RewardSplit, error) {
	rewardSplits := new(RewardSplit)
	if err := mp.Contract.Call(opts, rewardSplits, "calculateRewards", amount); err != nil {
		return RewardSplit{}, fmt.Errorf("error calculating the rewards for amount %s: %w", amount, err)
	}
	return *rewardSplits, nil
}

func (mp *megapoolV1) GetPendingRewards(opts *bind.CallOpts) (*big.Int, error) {
	pendingRewards := new(*big.Int)
	if err := mp.Contract.Call(opts, pendingRewards, "getPendingRewards"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s pending rewards: %w", mp.Address.Hex(), err)
	}
	return *pendingRewards, nil
}

func (mp *megapoolV1) GetNodeAddress(opts *bind.CallOpts) (common.Address, error) {
	nodeAddress := new(common.Address)
	if err := mp.Contract.Call(opts, nodeAddress, "getNodeAddress"); err != nil {
		return common.Address{}, fmt.Errorf("error getting megapool %s node address: %w", mp.Address.Hex(), err)
	}
	return *nodeAddress, nil
}

// Estimate the gas required to create a new validator as part of a megapool
func (mp *megapoolV1) EstimateNewValidatorGas(validatorId uint32, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "newValidator", validatorId, validatorSignature[:], depositDataRoot)
}

// Create a new validator as part of a megapool
func (mp *megapoolV1) NewValidator(bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "newValidator", bondAmount, useExpressTicket, validatorPubkey[:], validatorSignature[:])
	if err != nil {
		return common.Hash{}, fmt.Errorf("error creating new validator %s: %w", validatorPubkey.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to remove a validator from the deposit queue
func (mp *megapoolV1) EstimateDequeueGas(validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "dequeue", validatorId)
}

// Remove a validator from the deposit queue
func (mp *megapoolV1) Dequeue(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "dequeue", validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error dequeuing validator ID %d: %w", validatorId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to accept requested funds from the deposit pool
func (mp *megapoolV1) EstimateAssignFundsGas(validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "assignFunds", validatorId)
}

// Accept requested funds from the deposit pool
func (mp *megapoolV1) AssignFunds(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "assignFunds", validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error assigning funds to validator ID %d: %w", validatorId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to dissolve a validator that has not staked within the required period
func (mp *megapoolV1) EstimateDissolveValidatorGas(validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "dissolveValidator", validatorId)
}

// Dissolve a validator that has not staked within the required period
func (mp *megapoolV1) DissolveValidator(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "dissolveValidator", validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error dissolving validator ID %d: %w", validatorId, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to repay megapool debt
func (mp *megapoolV1) EstimateRepayDebtGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "repayDebt")
}

// Receive ETH, which is sent to the rETH contract, to repay a megapool debt
func (mp *megapoolV1) RepayDebt(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "repayDebt")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error repaying debt for megapool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to reduce a megapool bond
func (mp *megapoolV1) EstimateReduceBondGas(amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "reduceBond", amount)
}

// If the megapool is overbonded, reduce the bond by the specified amount
func (mp *megapoolV1) ReduceBond(amount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "reduceBond", amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error reducing the megapool bond %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to claim a megapool refund
func (mp *megapoolV1) EstimateClaimRefundGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "claim")
}

// Claim megapool rewards that were distributed but not yet claimed
func (mp *megapoolV1) ClaimRefund(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "claim")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error claiming megapool refund %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Get the expected withdrawal credentials for any validator within this megapool
func (mp *megapoolV1) GetWithdrawalCredentials(opts *bind.CallOpts) (common.Hash, error) {
	withdrawalCredentials := new(common.Hash)
	if err := mp.Contract.Call(opts, withdrawalCredentials, "getWithdrawalCredentials"); err != nil {
		return common.Hash{}, fmt.Errorf("error getting megapool %s withdrawal credentials: %w", mp.Address.Hex(), err)
	}
	return *withdrawalCredentials, nil
}

// Estimate the gas required to Request RPL previously staked on this megapool to be unstaked
func (mp *megapoolV1) EstimateRequestUnstakeRPL(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "requestUnstakeRPL")
}

// RequestUnstakeRPL is not yet implemented in RocketMegapoolDelegate.sol
// Request RPL previously staked on this megapool to be unstaked
func (mp *megapoolV1) RequestUnstakeRPL(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "requestUnstakeRPL")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error requesting unstake rpl for megapool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to distribute megapool rewards
func (mp *megapoolV1) EstimateDistributeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "distribute")
}

// Distribute megapool rewards
func (mp *megapoolV1) Distribute(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "distribute")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error distributing megapool rewards: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of SetUseLatestDelegate
func (mp *megapoolV1) EstimateSetUseLatestDelegateGas(setting bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "setUseLatestDelegate", setting)
}

// If set to true, will automatically use the latest delegate contract
func (mp *megapoolV1) SetUseLatestDelegate(setting bool, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "setUseLatestDelegate", setting)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting use latest delegate for megapool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Getter for useLatestDelegate setting
func (mp *megapoolV1) GetUseLatestDelegate(opts *bind.CallOpts) (bool, error) {
	setting := new(bool)
	if err := mp.Contract.Call(opts, setting, "getUseLatestDelegate"); err != nil {
		return false, fmt.Errorf("error getting use latest delegate for megapool %s: %w", mp.Address.Hex(), err)
	}
	return *setting, nil
}

// Returns the address of the megapool's stored delegate
func (mp *megapoolV1) GetDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("error getting delegate for megapool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Returns the delegate which will be used when calling this minipool taking into account useLatestDelegate setting
func (mp *megapoolV1) GetEffectiveDelegate(opts *bind.CallOpts) (common.Address, error) {
	address := new(common.Address)
	if err := mp.Contract.Call(opts, address, "getEffectiveDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("error getting effective delegate for megapool %s: %w", mp.Address.Hex(), err)
	}
	return *address, nil
}

// Returns true if the megapools current delegate has expired
func (mp *megapoolV1) GetDelegateExpired(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	delegateExpired := new(bool)
	if err := mp.Contract.Call(opts, delegateExpired, "getDelegateExpired"); err != nil {
		return false, fmt.Errorf("error checking if the megapool's delegate has expired:, %w", err)
	}
	return *delegateExpired, nil
}

// Estimate the gas of DelegateUpgrade
func (mp *megapoolV1) EstimateDelegateUpgradeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "delegateUpgrade")
}

// Upgrade this megapool to the latest network delegate contract
func (mp *megapoolV1) DelegateUpgrade(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "delegateUpgrade")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error upgrading delegate for megapool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

var ValidatorBatchSize = uint32(50)

func (mp *megapoolV1) GetMegapoolPubkeys(opts *bind.CallOpts) ([]rptypes.ValidatorPubkey, error) {
	validatorCount, err := mp.GetValidatorCount(opts)
	if err != nil {
		return []rptypes.ValidatorPubkey{}, err
	}

	// Load pubkeys in batches
	var lock = sync.RWMutex{}
	pubkeys := make([]rptypes.ValidatorPubkey, validatorCount)
	for bsi := uint32(0); bsi < validatorCount; bsi += ValidatorBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + ValidatorBatchSize
		if mei > validatorCount {
			mei = validatorCount
		}

		// Load pubkeys
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				validator, err := mp.GetValidatorInfoAndPubkey(mi, opts)
				if err != nil {
					return err
				}
				lock.Lock()
				pubkeys[mi] = rptypes.BytesToValidatorPubkey(validator.Pubkey)
				lock.Unlock()
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return []rptypes.ValidatorPubkey{}, err
		}

	}
	// Return
	return pubkeys, nil
}

// Create a megapool contract directly from its ABI
func createMegapoolContractFromAbi(rp *rocketpool.RocketPool, address common.Address, abi *abi.ABI) (*rocketpool.Contract, error) {
	// Create and return
	return &rocketpool.Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}, nil
}

// Create a megapool contract directly from its ABI, encoded in string form
func createMegapoolContractFromEncodedAbi(rp *rocketpool.RocketPool, address common.Address, encodedAbi string) (*rocketpool.Contract, error) {
	// Decode ABI
	abi, err := rocketpool.DecodeAbi(encodedAbi)
	if err != nil {
		return nil, fmt.Errorf("error decoding megapool %s ABI: %w", address, err)
	}

	// Create and return
	return &rocketpool.Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}, nil
}
