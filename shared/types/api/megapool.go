package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/units"
)

type MegapoolStatusResponse struct {
	APIResponse
	Megapool       MegapoolDetails   `json:"megapoolDetails"`
	LatestDelegate common.Address    `json:"latestDelegate"`
	BeaconHead     beacon.BeaconHead `json:"beaconHead"`
	// ShardCommitteePeriod is the number of epochs after activation before voluntary exit is allowed
	ShardCommitteePeriod uint64 `json:"shardCommitteePeriod"`
	// SecondsPerEpoch is used for wall-clock estimates of exit/withdrawal timing
	SecondsPerEpoch uint64 `json:"secondsPerEpoch"`
}

type MegapoolDetails struct {
	Address                  common.Address             `json:"address"`
	DelegateAddress          common.Address             `json:"delegate"`
	EffectiveDelegateAddress common.Address             `json:"effectiveDelegateAddress"`
	Deployed                 bool                       `json:"deployed"`
	ValidatorCount           uint32                     `json:"validatorCount"`
	ActiveValidatorCount     uint32                     `json:"activeValidatorCount"`
	ExitingValidatorCount    uint32                     `json:"exitingValidatorCount"`
	LockedValidatorCount     uint32                     `json:"lockedValidatorCount"`
	NodeDebt                 units.Wei                  `json:"nodeDebt"`
	RefundValue              units.Wei                  `json:"refundValue"`
	DelegateExpiry           uint64                     `json:"delegateExpiry"`
	DelegateExpired          bool                       `json:"delegateExpired"`
	PendingRewards           units.Wei                  `json:"pendingRewards"`
	NodeExpressTicketCount   uint64                     `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool                       `json:"useLatestDelegate"`
	AssignedValue            units.Wei                  `json:"assignedValue"`
	NodeBond                 units.Wei                  `json:"nodeBond"`
	NodeQueuedBond           units.Wei                  `json:"nodeQueuedBond"`
	UserCapital              units.Wei                  `json:"userCapital"`
	NodeShare                units.Wei                  `json:"nodeShare"`
	BondRequirement          units.Wei                  `json:"bondRequirement"`
	RevenueSplit             network.RevenueSplit       `json:"revenueSplit"`
	Balances                 tokens.Balances            `json:"balances"`
	LastDistributionTime     uint64                     `json:"lastDistributionTime"`
	PendingRewardSplit       megapool.RewardSplit       `json:"pendingRewardSplit"`
	ReducedBond              units.Wei                  `json:"reducedBond"`
	QueueDetails             QueueDetails               `json:"queueDetails"`
	Validators               []MegapoolValidatorDetails `json:"validators"`
}

type MegapoolValidatorDetails struct {
	ValidatorId        uint32                 `json:"validatorId"`
	PubKey             types.ValidatorPubkey  `json:"pubKey"`
	LastAssignmentTime time.Time              `json:"lastAssignmentTime"`
	LastRequestedValue uint32                 `json:"lastRequestedValue"`
	LastRequestedBond  uint32                 `json:"lastRequestedBond"`
	DepositValue       uint32                 `json:"DepositValue"`
	Staked             bool                   `json:"staked"`
	Exited             bool                   `json:"exited"`
	InQueue            bool                   `json:"inQueue"`
	QueuePosition      *big.Int               `json:"queuePosition"`
	InPrestake         bool                   `json:"inPrestake"`
	ExpressUsed        bool                   `json:"expressUsed"`
	Dissolved          bool                   `json:"dissolved"`
	Exiting            bool                   `json:"exiting"`
	Locked             bool                   `json:"locked"`
	ValidatorIndex     uint64                 `json:"validatorIndex"`
	ExitBalance        uint64                 `json:"exitBalance"`
	WithdrawableEpoch  uint64                 `json:"withdrawableEpoch"`
	LockedTime         uint64                 `json:"lockedTime"`
	Activated          bool                   `json:"activated"`
	BeaconStatus       beacon.ValidatorStatus `json:"beaconStatus"`
}

type MegapoolValidatorMapAndRewardsResponse struct {
	APIResponse
	MegapoolValidatorMap map[string][]MegapoolValidatorDetails `json:"megapoolValidatorMap"`
	TotalBeaconBalance   units.Wei                             `json:"totalBeaconBalance"`
	NodeShareOfCLBalance units.Wei                             `json:"nodeShareOfCLBalance"`
	NodeBond             units.Wei                             `json:"nodeBond"`
}

type MegapoolRewardSplitResponse struct {
	APIResponse
	RewardSplit megapool.RewardSplit `json:"rewardSplit"`
	RefundValue units.Wei            `json:"refundValue"`
}

type QueueDetails struct {
	ExpressQueueLength  *big.Int `json:"expressQueueLength"`
	StandardQueueLength *big.Int `json:"standardQueueLength"`
	QueueIndex          *big.Int `json:"queueIndex"`
	ExpressQueueRate    uint64   `json:"expressQueueRate"`
}

type MegapoolCanDelegateUpgradeResponse struct {
	APIResponse
	GasLimits gaslimit.Limits `json:"gasLimits"`
}
type MegapoolDelegateUpgradeResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type MegapoolGetDelegateResponse struct {
	APIResponse
	Address common.Address `json:"address"`
}

type MegapoolCanSetUseLatestDelegateResponse struct {
	APIResponse
	GasLimits             gaslimit.Limits `json:"gasLimits"`
	MatchesCurrentSetting bool            `json:"matchesCurrentSetting"`
}
type MegapoolSetUseLatestDelegateResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type MegapoolGetUseLatestDelegateResponse struct {
	APIResponse
	Setting bool `json:"setting"`
}

type MegapoolGetEffectiveDelegateResponse struct {
	APIResponse
	Address common.Address `json:"address"`
}

type CanDistributeMegapoolResponse struct {
	APIResponse
	MegapoolAddress       common.Address  `json:"megapoolAddress"`
	MegapoolNotDeployed   bool            `json:"megapoolNotDeployed"`
	LastDistributionTime  uint64          `json:"lastDistributionTime"`
	LockedValidatorCount  uint32          `json:"lockedValidatorCount"`
	ExitingValidatorCount uint32          `json:"exitingValidatorCount"`
	CanDistribute         bool            `json:"canDistribute"`
	Details               MegapoolDetails `json:"details"`
	GasLimits             gaslimit.Limits `json:"gasLimits"`
}

type DistributeMegapoolResponse struct {
	APIResponse
	TxHash common.Hash `json:"txHash"`
}

type ValidatorWithdrawableEpochProof struct {
	Slot              uint64
	ValidatorIndex    *big.Int
	Pubkey            []byte
	WithdrawableEpoch uint64
	Witnesses         [][32]byte
}
type GetNewValidatorBondRequirementResponse struct {
	APIResponse
	NewValidatorBondRequirement units.Wei `json:"newValidatorBondRequirement"`
}

type GetNodeMegapoolEthBondedResponse struct {
	APIResponse
	EthBonded *big.Int `json:"ethBonded"`
}

type LatestBlockWithdrawalsResponse struct {
	APIResponse
	Slot        uint64                  `json:"slot"`
	BlockNumber uint64                  `json:"blockNumber"`
	Withdrawals []beacon.WithdrawalInfo `json:"withdrawals"`
}

type BeaconWithdrawalQueueEstimateResponse struct {
	APIResponse
	ExitQueueGwei         uint64 `json:"exitQueueGwei"`
	ChurnPerEpochGwei     uint64 `json:"churnPerEpochGwei"`
	SecondsPerEpoch       uint64 `json:"secondsPerEpoch"`
	EstimatedQueueEpochs  uint64 `json:"estimatedQueueEpochs"`
	EstimatedQueueSeconds uint64 `json:"estimatedQueueSeconds"`
}
