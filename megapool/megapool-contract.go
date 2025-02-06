package megapool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
)

type ValidatorProof struct {
	Slot                  uint64
	ValidatorIndex        *big.Int
	Pubkey                []byte
	WithdrawalCredentials [32]byte
	Witnesses             [][32]byte
}

type withdrawal struct {
	index                 *big.Int
	validatorIndex        *big.Int
	withdrawalCredentials [32]byte
	amountInGwei          *big.Int
}

type MegapoolV1 interface {
	Megapool
}

type ValidatorInfo struct {
	PubKey             []byte `abi:"pubKey"`
	LastAssignmentTime uint32 `abi:"lastAssignmentTime"`
	LastRequestedValue uint32 `abi:"lastRequestedValue"`
	LastRequestedBond  uint32 `abi:"lastRequestedBond"`
	Staked             bool   `abi:"staked"`
	Exited             bool   `abi:"exited"`
	InQueue            bool   `abi:"inQueue"`
	InPrestake         bool   `abi:"inPrestake"`
	ExpressUsed        bool   `abi:"expressUsed"`
	Dissolved          bool   `abi:"dissolved"`
}

// Megapool contract
type megapoolV1 struct {
	Address    common.Address
	Version    uint8
	Contract   *rocketpool.Contract
	RocketPool *rocketpool.RocketPool
}

const (
	megapoolV1EncodedAbi string = "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"
)

// The decoded ABI for megapools
var megapoolV1Abi *abi.ABI

// Create new minipool contract
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

func (mp *megapoolV1) GetValidatorCount(opts *bind.CallOpts) (uint32, error) {
	var validatorCount uint32
	if err := mp.Contract.Call(opts, &validatorCount, "getValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s validator count: %w", mp.Address.Hex(), err)
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

func (mp *megapoolV1) GetNodeCapital(opts *bind.CallOpts) (*big.Int, error) {
	nodeCapital := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeCapital, "getNodeCapital"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s node capital: %w", mp.Address.Hex(), err)
	}
	return *nodeCapital, nil
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

//TODO _calculateRewards is currently a view in RocketMegapoolDelegate.sol

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

// Estimate the gas of Stake
func (mp *megapoolV1) EstimateStakeGas(validatorId uint32, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, validatorProof ValidatorProof, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "stake", validatorId, validatorSignature[:], depositDataRoot, validatorProof)
}

// Progress the prelaunch megapool to staking
func (mp *megapoolV1) Stake(validatorId uint32, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, validatorProof ValidatorProof, opts *bind.TransactOpts) (common.Hash, error) {
	// callData, err := mp.Contract.ABI.Pack("stake", validatorId, validatorSignature[:], depositDataRoot, validatorProof)
	// if err != nil {
	// 	return common.Hash{}, fmt.Errorf("error creating calldata for getValidatorInfo: %w", err)
	// }

	// fmt.Println("call data:\n")
	// fmt.Printf("%s", hex.EncodeToString(callData))

	// tx, err := mp.Contract.Contract.RawTransact(opts, callData)
	// if err != nil {
	// 	return common.Hash{}, fmt.Errorf("error calling getValidatorInfo: %w", err)
	// }

	tx, err := mp.Contract.Transact(opts, "stake", validatorId, validatorSignature[:], depositDataRoot, validatorProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error staking megapool %s: %w", mp.Address.Hex(), err)
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
