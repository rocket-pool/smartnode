package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

type MegapoolStatusResponse struct {
	Status         string          `json:"status"`
	Error          string          `json:"error"`
	Megapool       MegapoolDetails `json:"megapoolDetails"`
	LatestDelegate common.Address  `json:"latestDelegate"`
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
	NodeDebt                 *big.Int                   `json:"nodeDebt"`
	RefundValue              *big.Int                   `json:"refundValue"`
	DelegateExpiry           uint64                     `json:"delegateExpiry"`
	DelegateExpired          bool                       `json:"delegateExpired"`
	PendingRewards           *big.Int                   `json:"pendingRewards"`
	NodeExpressTicketCount   uint64                     `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool                       `json:"useLatestDelegate"`
	AssignedValue            *big.Int                   `json:"assignedValue"`
	NodeBond                 *big.Int                   `json:"nodeBond"`
	UserCapital              *big.Int                   `json:"userCapital"`
	NodeShare                *big.Int                   `json:"nodeShare"`
	BondRequirement          *big.Int                   `json:"bondRequirement"`
	RevenueSplit             network.RevenueSplit       `json:"revenueSplit"`
	Balances                 tokens.Balances            `json:"balances"`
	LastDistributionBlock    uint64                     `json:"lastDistributionBlock"`
	PendingRewardSplit       megapool.RewardSplit       `json:"pendingRewardSplit"`
	ReducedBond              *big.Int                   `json:"reducedBond"`
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
	LockedSlot         uint64                 `json:"lockedSlot"`
	Activated          bool                   `json:"activated"`
	BeaconStatus       beacon.ValidatorStatus `json:"beaconStatus"`
}

type MegapoolValidatorMapAndRewardsResponse struct {
	Status               string                                `json:"status"`
	Error                string                                `json:"error"`
	MegapoolValidatorMap map[string][]MegapoolValidatorDetails `json:"megapoolValidatorMap"`
	TotalBeaconBalance   *big.Int                              `json:"totalBeaconBalance"`
	NodeShareOfCLBalance *big.Int                              `json:"nodeShareOfCLBalance"`
	NodeBond             *big.Int                              `json:"nodeBond"`
}

type MegapoolRewardSplitResponse struct {
	Status      string               `json:"status"`
	Error       string               `json:"error"`
	RewardSplit megapool.RewardSplit `json:"rewardSplit"`
}

type QueueDetails struct {
	ExpressQueueLength  *big.Int `json:"expressQueueLength"`
	StandardQueueLength *big.Int `json:"standardQueueLength"`
	QueueIndex          *big.Int `json:"queueIndex"`
	ExpressQueueRate    uint64   `json:"expressQueueRate"`
}

type MegapoolCanDelegateUpgradeResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type MegapoolDelegateUpgradeResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type MegapoolGetDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type MegapoolCanSetUseLatestDelegateResponse struct {
	Status                string             `json:"status"`
	Error                 string             `json:"error"`
	GasInfo               rocketpool.GasInfo `json:"gasInfo"`
	MatchesCurrentSetting bool               `json:"matchesCurrentSetting"`
}
type MegapoolSetUseLatestDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type MegapoolGetUseLatestDelegateResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Setting bool   `json:"setting"`
}

type MegapoolGetEffectiveDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type CanDistributeMegapoolResponse struct {
	Status                string             `json:"status"`
	Error                 string             `json:"error"`
	MegapoolAddress       common.Address     `json:"megapoolAddress"`
	MegapoolNotDeployed   bool               `json:"megapoolNotDeployed"`
	LastDistributionBlock uint64             `json:"lastDistributionBlock"`
	LockedValidatorCount  uint32             `json:"lockedValidatorCount"`
	ExitingValidatorCount uint32             `json:"exitingValidatorCount"`
	CanDistribute         bool               `json:"canDistribute"`
	Details               MegapoolDetails    `json:"details"`
	GasInfo               rocketpool.GasInfo `json:"gasInfo"`
}

type DistributeMegapoolResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type ValidatorWithdrawableEpochProof struct {
	Slot              uint64
	ValidatorIndex    *big.Int
	Pubkey            []byte
	WithdrawableEpoch uint64
	Witnesses         [][32]byte
}
