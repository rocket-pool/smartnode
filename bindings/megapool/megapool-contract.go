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

type ValidatorProof struct {
	Slot                  uint64
	ValidatorIndex        uint64
	Pubkey                []byte
	WithdrawalCredentials [32]byte
	Witnesses             [][32]byte
}

type FinalBalanceProof struct {
	Slot           uint64
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
	Index                 uint64
	ValidatorIndex        uint64
	WithdrawalCredentials [20]byte
	AmountInGwei          uint64
}

type RewardSplit struct {
	NodeRewards  *big.Int `abi:"nodeRewards"`
	VoterRewards *big.Int `abi:"voterRewards"`
	RethRewards  *big.Int `abi:"rethRewards"`
}

type MegapoolV1 interface {
	Megapool
}

type ValidatorInfo struct {
	PubKey             []byte `abi:"pubKey"`
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
	megapoolV1EncodedAbi string = "eJztWktv2zgQ/isLn4Nike320Fse7iJoWmTtZPcQFAYljR0iNKnlw45Q7H/foS3rLYuKpECt99Q6JIffDGe+GQ71+H1CuODRWhg1+bgkTMHZhPLQaPz5+B3/G8ALBJkhDZITdh+FMPk4Mfj7/PcPk7MJJ2v7B4KCuMbfOj/h37P2sjTFf8qSviUTvsCKhEKwS8GDGQTGR+HJfNgAArH7/vgKXoOnb7gvgaifWcXOZ6ilccC0IYwGRAt5E7yxkn8ddr5Qiq74Sah6Df8YMKehKlVKsM1J6Drlgx3rb+cOmuKkN1L0herTOFFUlPLVCWg61+T5Z8ujM9gSGagrRnDm8LpxEcBFj/ptBI73KVCCfupTnusBYA7QknqmhjNyJj++46LWf9JNSRiy6A5Qio5wojA6Fo9TlCYavhhNPMooDttD4yGJiMcymiwN9zUV3AFdlpcXR4k5A3BX6HwyPFDN+BzBpdJ9wnzDUAyaIEDuis8gt1Ozoa0vpysrfcXBeTtJsN5aL6DSWBsK21cdY2snS6z8v3lrHdHSbp8BmIoOIJTgo6RRxnewr+7HiS0ux5O8P9ABJYw/jPwV6AuckNHjKg7ZhijM2qzSTl293iKLL7KIreAFzVE5TCgiKNtLGAuWuKoe49lNX0IqiZ10yYT/PBaL3RKlkypqbOC+2ro3CCSohjRIkkklSIehHiHZDuhYTNSpFhsM1Qxw0rhoai4EB6X/pvopkGRrs8I0FP5TM74P7+vg4UhP6B4UyCsSUk3YWAw2EhLtrUDJqnTDl6KgkS/WIXoIr97UizRk6CU03meI0t32wzXFcRYsQ7rdZ/E1bnVfuuYmjbZGKTNbCyqdFgPdpMSc1loIFsxCUe0MwhOCpatVoTe0G21cBYUmodsqyv+Mi+eWy+4wf1ic7WGGNvM8qPZYg3Jv2900ubZi3bIisaVBZFs1ZZprlmC3viSMcB9es3xbQcxlqi0IwbrF+PqXGdYsoA+dxzlqQVZwSyKM73fFmC8xlDYhXhB6Is0kvbArCQGGOCWsISvvuKOaPA9Dg/YkPIz84727I/62MAqmez+/p/YQHBw2T6Upfd8Z79mJVGslzJFXiTYSXIXk8khMZddEk5kQuvoYDm0W2A5y0e2Y7ZrDbOEQZw5SFBPaZWFsusdvmbWhFGJZsi5OydpXaLqM7HXuB7UumztaqByQqYyvZl0dk8eLlSIq+lpKb58Udod5/msFrRcIUefnO4DZ905v+B9boC1yQ8rIlSYu5YHROP8nipLShDpsELRuU8vd1yOlm3BvfTgJuKTUWurykNE9+NtFnauX9BR0uctJizyatYBLsNaU9WVXR2F44VZQFhB7e3W8Jmn1Lo6UmshJy7bsXWIYZ9yAVHb4zW/mXV6RSz0xwYJrYFgk64zqSXus+ZW2JBCLoE4C2z/7HrZ7CFeSBL1/KFLScCnFuh/V3vTzhKl+sm94PtBOn0dVwynd4HX1zbhnnfAafYtbKf0QIj1Uq+WSGwhjHvGf6x8Acx42DJ/sXk2SwBlFX/uAZ/dMAQ3t7bwLFE+/p/eS5RJ2D3Fjs1TiiG7ABrJVswEWPPd6UmGM5HpAbaKnatgys9A4qCSOTDZvMvTr4DkwhNzzJoL5DzTANyg="
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
				validator, err := mp.GetValidatorInfo(mi, opts)
				if err != nil {
					return err
				}
				lock.Lock()
				pubkeys[mi] = rptypes.BytesToValidatorPubkey(validator.PubKey)
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
