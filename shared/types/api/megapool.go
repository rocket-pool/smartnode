package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
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
	NodeDebt                 *big.Int                   `json:"nodeDebt"`
	RefundValue              *big.Int                   `json:"refundValue"`
	DelegateExpiry           uint64                     `json:"delegateExpiry"`
	DelegateExpired          bool                       `json:"delegateExpired"`
	PendingRewards           *big.Int                   `json:"pendingRewards"`
	NodeExpressTicketCount   uint64                     `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool                       `json:"useLatestDelegate"`
	AssignedValue            *big.Int                   `json:"assignedValue"`
	NodeCapital              *big.Int                   `json:"nodeCapital"`
	NodeBond                 *big.Int                   `json:"nodeBond"`
	UserCapital              *big.Int                   `json:"userCapital"`
	NodeShare                *big.Int                   `json:"nodeShare"`
	RevenueSplit             network.RevenueSplit       `json:"revenueSplit"`
	LastDistributionBlock    uint64                     `json:"lastDistributionBlock"`
	QueueDetails             QueueDetails               `json:"queueDetails"`
	Validators               []MegapoolValidatorDetails `json:"validators"`
}

type MegapoolValidatorDetails struct {
	ValidatorId        uint32                 `json:"validatorId"`
	PubKey             types.ValidatorPubkey  `json:"pubKey"`
	LastAssignmentTime time.Time              `json:"lastAssignmentTime"`
	LastRequestedValue uint32                 `json:"lastRequestedValue"`
	LastRequestedBond  uint32                 `json:"lastRequestedBond"`
	Staked             bool                   `json:"staked"`
	Exited             bool                   `json:"exited"`
	InQueue            bool                   `json:"inQueue"`
	QueuePosition      *big.Int               `json:"queuePosition"`
	InPrestake         bool                   `json:"inPrestake"`
	ExpressUsed        bool                   `json:"expressUsed"`
	Dissolved          bool                   `json:"dissolved"`
	Activated          bool                   `json:"activated"`
	BeaconStatus       beacon.ValidatorStatus `json:"beaconStatus"`
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
