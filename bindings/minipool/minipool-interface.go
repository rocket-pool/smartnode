package minipool

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/units"
)

// The number of blocks to look for events in at once when scanning
const EventScanInterval = 10000

// Minipool detail types
type StatusDetails struct {
	Status      rptypes.MinipoolStatus `json:"status"`
	StatusBlock uint64                 `json:"statusBlock"`
	StatusTime  time.Time              `json:"statusTime"`
	IsVacant    bool                   `json:"isVacant"`
}
type NodeDetails struct {
	Address         common.Address `json:"address"`
	Fee             units.Eth      `json:"fee"`
	DepositBalance  units.Wei      `json:"depositBalance"`
	RefundBalance   units.Wei      `json:"refundBalance"`
	DepositAssigned bool           `json:"depositAssigned"`
}
type UserDetails struct {
	DepositBalance      units.Wei `json:"depositBalance"`
	DepositAssigned     bool      `json:"depositAssigned"`
	DepositAssignedTime time.Time `json:"depositAssignedTime"`
}

// The data from a minipool's MinipoolPrestaked event
type MinipoolPrestakeEvent struct {
	Pubkey                []byte    `abi:"validatorPubkey"`
	Signature             []byte    `abi:"validatorSignature"`
	DepositDataRoot       [32]byte  `abi:"depositDataRoot"`
	Amount                units.Wei `abi:"amount"`
	WithdrawalCredentials []byte    `abi:"withdrawalCredentials"`
	Time                  *big.Int  `abi:"time"`
}

// Formatted MinipoolPrestaked event data
type PrestakeData struct {
	Pubkey                rptypes.ValidatorPubkey    `json:"pubkey"`
	WithdrawalCredentials common.Hash                `json:"withdrawalCredentials"`
	Amount                units.Wei                  `json:"amount"`
	Signature             rptypes.ValidatorSignature `json:"signature"`
	DepositDataRoot       common.Hash                `json:"depositDataRoot"`
	Time                  time.Time                  `json:"time"`
}

type Minipool interface {
	GetContract() *rocketpool.Contract
	GetAddress() common.Address
	GetVersion() uint8
	GetStatusDetails(opts *bind.CallOpts) (StatusDetails, error)
	GetStatus(opts *bind.CallOpts) (rptypes.MinipoolStatus, error)
	GetStatusBlock(opts *bind.CallOpts) (uint64, error)
	GetStatusTime(opts *bind.CallOpts) (time.Time, error)
	GetFinalised(opts *bind.CallOpts) (bool, error)
	GetDepositType(opts *bind.CallOpts) (rptypes.MinipoolDeposit, error)
	GetNodeDetails(opts *bind.CallOpts) (NodeDetails, error)
	GetNodeAddress(opts *bind.CallOpts) (common.Address, error)
	GetNodeFee(opts *bind.CallOpts) (units.Eth, error)
	GetNodeDepositBalance(opts *bind.CallOpts) (units.Wei, error)
	GetNodeRefundBalance(opts *bind.CallOpts) (units.Wei, error)
	GetNodeDepositAssigned(opts *bind.CallOpts) (bool, error)
	GetUserDetails(opts *bind.CallOpts) (UserDetails, error)
	GetUserDepositBalance(opts *bind.CallOpts) (units.Wei, error)
	GetUserDepositAssigned(opts *bind.CallOpts) (bool, error)
	GetUserDepositAssignedTime(opts *bind.CallOpts) (time.Time, error)
	EstimateRefundGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	Refund(opts *bind.TransactOpts) (common.Hash, error)
	EstimateStakeGas(validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (gaslimit.Limits, error)
	Stake(validatorSignature rptypes.ValidatorSignature, depositDataRoot common.Hash, opts *bind.TransactOpts) (common.Hash, error)
	EstimateDissolveGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	Dissolve(opts *bind.TransactOpts) (common.Hash, error)
	EstimateCloseGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	Close(opts *bind.TransactOpts) (common.Hash, error)
	EstimateFinaliseGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	Finalise(opts *bind.TransactOpts) (common.Hash, error)
	EstimateDelegateUpgradeGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	DelegateUpgrade(opts *bind.TransactOpts) (common.Hash, error)
	EstimateSetUseLatestDelegateGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	SetUseLatestDelegate(opts *bind.TransactOpts) (common.Hash, error)
	GetUseLatestDelegate(opts *bind.CallOpts) (bool, error)
	GetDelegate(opts *bind.CallOpts) (common.Address, error)
	GetPreviousDelegate(opts *bind.CallOpts) (common.Address, error)
	GetEffectiveDelegate(opts *bind.CallOpts) (common.Address, error)
	CalculateNodeShare(balance units.Wei, opts *bind.CallOpts) (units.Wei, error)
	CalculateUserShare(balance units.Wei, opts *bind.CallOpts) (units.Wei, error)
	EstimateVoteScrubGas(opts *bind.TransactOpts) (gaslimit.Limits, error)
	VoteScrub(opts *bind.TransactOpts) (common.Hash, error)
	GetPrestakeEvent(intervalSize *big.Int, opts *bind.CallOpts) (PrestakeData, error)
}
