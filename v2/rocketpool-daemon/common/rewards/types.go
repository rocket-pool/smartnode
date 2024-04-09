package rewards

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

// Legacy carryover from rocketpool-go v1 for interval 4 and 5 generators
type MinipoolDetails struct {
	Address      common.Address
	Exists       bool
	Status       types.MinipoolStatus
	Pubkey       beacon.ValidatorPubkey
	PenaltyCount uint64
	NodeFee      *big.Int
}

type MinipoolInfo struct {
	Address                 common.Address            `json:"address"`
	ValidatorPubkey         beacon.ValidatorPubkey    `json:"pubkey"`
	ValidatorIndex          string                    `json:"index"`
	NodeAddress             common.Address            `json:"nodeAddress"`
	NodeIndex               uint64                    `json:"-"`
	Fee                     *big.Int                  `json:"-"`
	MissedAttestations      uint64                    `json:"-"`
	GoodAttestations        uint64                    `json:"-"`
	MinipoolShare           *big.Int                  `json:"-"`
	MissingAttestationSlots map[uint64]bool           `json:"missingAttestationSlots"`
	WasActive               bool                      `json:"-"`
	StartSlot               uint64                    `json:"-"`
	EndSlot                 uint64                    `json:"-"`
	AttestationScore        *sharedtypes.QuotedBigInt `json:"attestationScore"`
	CompletedAttestations   map[uint64]bool           `json:"-"`
	AttestationCount        int                       `json:"attestationCount"`
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
