package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
)

type validatorProof struct {
	slot                  uint64
	validatorIndex        *big.Int
	pubkey                []byte
	withdrawalCredentials [32]byte
	witnesses             [][32]byte
}

type MegapoolV1 interface {
	Megapool
}

// Megapool contract
type megapoolV1 struct {
	Address    common.Address
	Version    uint8
	Contract   *rocketpool.Contract
	RocketPool *rocketpool.RocketPool
}

const (
	megapoolV1EncodedAbi string = ""
)

// The decoded ABI for megapools
var megapoolV1Abi *abi.ABI

// Create new minipool contract
func newMegaPoolV1(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Megapool, error) {

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
		Version:    3,
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

func (mp *megapoolV1) GetValidatorCount(opts *bind.CallOpts) (uint64, error) {
	validatorCount := new(*big.Int)
	if err := mp.Contract.Call(opts, validatorCount, "getValidatorCount"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s validator count: %w", mp.Address.Hex(), err)
	}
	return (*validatorCount).Uint64(), nil
}

//TODO func GetValidatorInfo()

func (mp *megapoolV1) GetAssignedValue(opts *bind.CallOpts) (uint64, error) {
	validatorCount := new(*big.Int)
	if err := mp.Contract.Call(opts, validatorCount, "getAssignedValue"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s assigned value: %w", mp.Address.Hex(), err)
	}
	return (*validatorCount).Uint64(), nil
}

func (mp *megapoolV1) GetDebt(opts *bind.CallOpts) (uint64, error) {
	assignedValue := new(*big.Int)
	if err := mp.Contract.Call(opts, assignedValue, "getDebt"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s debt: %w", mp.Address.Hex(), err)
	}
	return (*assignedValue).Uint64(), nil
}

func (mp *megapoolV1) GetRefundValue(opts *bind.CallOpts) (uint64, error) {
	refundValue := new(*big.Int)
	if err := mp.Contract.Call(opts, refundValue, "getRefundValue"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s refund value: %w", mp.Address.Hex(), err)
	}
	return (*refundValue).Uint64(), nil
}

func (mp *megapoolV1) GetNodeCapital(opts *bind.CallOpts) (uint64, error) {
	nodeCapital := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeCapital, "getNodeCapital"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s node capital: %w", mp.Address.Hex(), err)
	}
	return (*nodeCapital).Uint64(), nil
}

func (mp *megapoolV1) GetNodeBond(opts *bind.CallOpts) (uint64, error) {
	nodeBond := new(*big.Int)
	if err := mp.Contract.Call(opts, nodeBond, "getNodeBond"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s debt: %w", mp.Address.Hex(), err)
	}
	return (*nodeBond).Uint64(), nil
}

func (mp *megapoolV1) GetUserCapital(opts *bind.CallOpts) (uint64, error) {
	userCapital := new(*big.Int)
	if err := mp.Contract.Call(opts, userCapital, "getUserCapital"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s user capital: %w", mp.Address.Hex(), err)
	}
	return (*userCapital).Uint64(), nil
}

//TODO _calculateRewards is currently a view in RocketMegapoolDelegate.sol

func (mp *megapoolV1) GetPendingRewards(opts *bind.CallOpts) (uint64, error) {
	pendingRewards := new(*big.Int)
	if err := mp.Contract.Call(opts, pendingRewards, "getPendingRewards"); err != nil {
		return 0, fmt.Errorf("error getting megapool %s pending rewards: %w", mp.Address.Hex(), err)
	}
	return (*pendingRewards).Uint64(), nil
}

func (mp *megapoolV1) GetNodeAddress(opts *bind.CallOpts) (common.Address, error) {
	nodeAddress := new(common.Address)
	if err := mp.Contract.Call(opts, nodeAddress, "getNodeAddress"); err != nil {
		return common.Address{}, fmt.Errorf("error getting megapool %s node address: %w", mp.Address.Hex(), err)
	}
	return *nodeAddress, nil
}

// TODO gas estimators for the below functions
func (mp *megapoolV1) NewValidator(bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "newValidator", bondAmount, useExpressTicket, validatorPubkey[:], validatorSignature[:])
	if err != nil {
		return common.Hash{}, fmt.Errorf("error creating new validator %s: %w", validatorPubkey.Hex(), err)
	}
	return tx.Hash(), nil
}

func (mp *megapoolV1) Dequeue(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "dequeue", validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error dequeuing validator ID %d: %w", validatorId, err)
	}
	return tx.Hash(), nil
}

func (mp *megapoolV1) AssignFunds(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "assignFunds", validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error assigning funds to validator ID %d: %w", validatorId, err)
	}
	return tx.Hash(), nil
}

func (mp *megapoolV1) DissolveValidator(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "dissolveValidator", validatorId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error dissolving validator ID %d: %w", validatorId, err)
	}
	return tx.Hash(), nil
}

func (mp *megapoolV1) RepayDebt(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "repayDebt")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error repaying debt for megapool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

func (mp *megapoolV1) GetWithdrawalCredentials(opts *bind.CallOpts) ([]byte, error) {
	withdrawalCredentials := make([]byte, 32)
	if err := mp.Contract.Call(opts, withdrawalCredentials, "getWithdrawalCredentials"); err != nil {
		return nil, fmt.Errorf("error getting megapool %s withdrawal credentials: %w", mp.Address.Hex(), err)
	}
	return withdrawalCredentials, nil
}

// RequestUnstakeRPL is not yet implemented in RocketMegapoolDelegate.sol
func (mp *megapoolV1) RequestUnstakeRPL(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "requestUnstakeRPL")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error requesting unstake rpl for megapool %s: %w", mp.Address.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Stake
func (mp *megapoolV1) EstimateStakeGas(validatorId uint32, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return mp.Contract.GetTransactionGasInfo(opts, "stake", validatorSignature[:], validatorSignature[:], depositDataRoot)
}

// Progress the prelaunch megapool to staking
func (mp *megapoolV1) Stake(validatorId uint32, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, validatorProof validatorProof, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := mp.Contract.Transact(opts, "stake", validatorSignature[:], depositDataRoot, validatorProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error staking megapool %s: %w", mp.Address.Hex(), err)
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
