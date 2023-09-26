package rewards

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
)

// Legacy carryover from rocketpool-go v1 for interval 4 and 5 generators
type MinipoolDetails struct {
	Address      common.Address
	Exists       bool
	Status       types.MinipoolStatus
	Pubkey       types.ValidatorPubkey
	PenaltyCount uint64
	NodeFee      *big.Int
}

type MinipoolInfo struct {
	Address                 common.Address
	ValidatorPubkey         types.ValidatorPubkey
	ValidatorIndex          string
	NodeAddress             common.Address
	NodeIndex               uint64
	Fee                     *big.Int
	MissedAttestations      uint64
	GoodAttestations        uint64
	MinipoolShare           *big.Int
	MissingAttestationSlots map[uint64]bool
	WasActive               bool
	StartSlot               uint64
	EndSlot                 uint64
	AttestationScore        *big.Int
	CompletedAttestations   map[uint64]bool
	AttestationCount        int
}

type IntervalDutiesInfo struct {
	Index uint64
	Slots map[uint64]*SlotInfo
}

type SlotInfo struct {
	Index      uint64
	Committees map[uint64]*CommitteeInfo
}

type CommitteeInfo struct {
	Index     uint64
	Positions map[int]*MinipoolInfo
}

// Details about a node for the Smoothing Pool
type NodeSmoothingDetails struct {
	Address          common.Address
	IsEligible       bool
	IsOptedIn        bool
	StatusChangeTime time.Time
	Minipools        []*MinipoolInfo
	EligibleSeconds  *big.Int
	StartSlot        uint64
	EndSlot          uint64
	SmoothingPoolEth *big.Int
	RewardsNetwork   uint64

	// v2 Fields
	OptInTime  time.Time
	OptOutTime time.Time
}
