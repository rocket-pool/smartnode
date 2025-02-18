package api

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type MegapoolStatusData struct {
	Megapool       MegapoolDetails `json:"megapoolDetails"`
	LatestDelegate common.Address  `json:"latestDelegate"`
}

type MegapoolDetails struct {
	Address                  common.Address `json:"address"`
	DelegateAddress          common.Address `json:"delegate"`
	EffectiveDelegateAddress common.Address `json:"effectiveDelegateAddress"`
	Deployed                 bool           `json:"deployed"`
	ValidatorCount           uint32         `json:"validatorCount"`
	NodeDebt                 *big.Int       `json:"nodeDebt"`
	RefundValue              *big.Int       `json:"refundValue"`
	DelegateExpiry           uint64         `json:"delegateExpiry"`
	DelegateExpired          bool           `json:"delegateExpired"`
	PendingRewards           *big.Int       `json:"pendingRewards"`
	NodeExpressTicketCount   uint64         `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool           `json:"useLatestDelegate"`
	AssignedValue            *big.Int       `json:"assignedValue"`
	NodeCapital              *big.Int       `json:"nodeCapital"`
	NodeBond                 *big.Int       `json:"nodeBond"`
	UserCapital              *big.Int       `json:"userCapital"`
	NodeShare                *big.Int       `json:"nodeShare"`
	// Leaving these out until they are added to rp-go:tx-refactor
	// RevenueSplit             network.RevenueSplit       `json:"revenueSplit"`
	// Balances                 tokens.Balances            `json:"balances"`
	LastDistributionBlock uint64                     `json:"lastDistributionBlock"`
	QueueDetails          QueueDetails               `json:"queueDetails"`
	Validators            []MegapoolValidatorDetails `json:"validators"`
}

type MegapoolValidatorDetails struct {
	ValidatorId uint32 `json:"validatorId"`
	// PubKey             types.ValidatorPubkey  `json:"pubKey"`
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
