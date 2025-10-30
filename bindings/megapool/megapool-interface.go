package megapool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
)

type Megapool interface {
	GetContract() *rocketpool.Contract
	GetAddress() common.Address
	GetVersion() uint8
	GetValidatorCount(opts *bind.CallOpts) (uint32, error)
	GetActiveValidatorCount(opts *bind.CallOpts) (uint32, error)
	GetLockedValidatorCount(opts *bind.CallOpts) (uint32, error)
	GetExitingValidatorCount(opts *bind.CallOpts) (uint32, error)
	GetSoonestWithdrawableEpoch(opts *bind.CallOpts) (uint32, error)
	GetValidatorInfo(validatorId uint32, opts *bind.CallOpts) (ValidatorInfo, error)
	GetValidatorPubkey(validatorId uint32, opts *bind.CallOpts) (rptypes.ValidatorPubkey, error)
	GetValidatorInfoAndPubkey(validatorId uint32, opts *bind.CallOpts) (ValidatorInfoWithPubkey, error)
	GetLastDistributionTime(opts *bind.CallOpts) (uint64, error)
	GetAssignedValue(opts *bind.CallOpts) (*big.Int, error)
	GetDebt(opts *bind.CallOpts) (*big.Int, error)
	GetRefundValue(opts *bind.CallOpts) (*big.Int, error)
	GetNodeBond(opts *bind.CallOpts) (*big.Int, error)
	GetUserCapital(opts *bind.CallOpts) (*big.Int, error)
	CalculatePendingRewards(opts *bind.CallOpts) (RewardSplit, error)
	CalculateRewards(amount *big.Int, opts *bind.CallOpts) (RewardSplit, error)
	GetPendingRewards(opts *bind.CallOpts) (*big.Int, error)
	GetNodeAddress(opts *bind.CallOpts) (common.Address, error)
	EstimateNewValidatorGas(validatorId uint32, validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	NewValidator(bondAmount *big.Int, useExpressTicket bool, validatorPubkey rptypes.ValidatorPubkey, validatorSignature rptypes.ValidatorSignature, opts *bind.TransactOpts) (common.Hash, error)
	EstimateDequeueGas(validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	Dequeue(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error)
	EstimateDistributeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	Distribute(opts *bind.TransactOpts) (common.Hash, error)
	EstimateAssignFundsGas(validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	AssignFunds(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error)
	EstimateDissolveValidatorGas(validatorId uint32, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	DissolveValidator(validatorId uint32, opts *bind.TransactOpts) (common.Hash, error)
	EstimateClaimRefundGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	ClaimRefund(opts *bind.TransactOpts) (common.Hash, error)
	EstimateRepayDebtGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	RepayDebt(opts *bind.TransactOpts) (common.Hash, error)
	EstimateReduceBondGas(amount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	ReduceBond(amount *big.Int, opts *bind.TransactOpts) (common.Hash, error)
	GetWithdrawalCredentials(opts *bind.CallOpts) (common.Hash, error)
	EstimateRequestUnstakeRPL(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	RequestUnstakeRPL(opts *bind.TransactOpts) (common.Hash, error)
	EstimateSetUseLatestDelegateGas(setting bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	SetUseLatestDelegate(setting bool, opts *bind.TransactOpts) (common.Hash, error)
	GetUseLatestDelegate(opts *bind.CallOpts) (bool, error)
	GetDelegate(opts *bind.CallOpts) (common.Address, error)
	GetDelegateExpired(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error)
	GetEffectiveDelegate(opts *bind.CallOpts) (common.Address, error)
	EstimateDelegateUpgradeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error)
	DelegateUpgrade(opts *bind.TransactOpts) (common.Hash, error)
	GetMegapoolPubkeys(opts *bind.CallOpts) ([]rptypes.ValidatorPubkey, error)
}
