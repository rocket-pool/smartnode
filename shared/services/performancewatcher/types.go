// Package performancewatcher collects finalized, bounded attestation evidence.
// It never submits transactions and never treats unavailable history as a miss.
package performancewatcher

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"
)

var ErrSyncing = errors.New("client syncing or execution unavailable")
var ErrFinalityConflict = errors.New("conflicting finalized roots")
var ErrValidatorNotTracked = errors.New("validator is not tracked by the performance watcher")

type Observation byte

const (
	Unknown Observation = iota
	Timely
	Missed
	Inactive
)

type Validator struct {
	Index       uint64 `json:"validatorIndex"`
	Pubkey      string `json:"pubkey"`
	Pool        string `json:"pool"`
	Kind        string `json:"kind"`
	ValidatorID uint32 `json:"validatorId"`
	Activation  uint64 `json:"activationEpoch"`
	Exit        uint64 `json:"exitEpoch"`
	Eligible    bool   `json:"eligible"`
}

type Parameters struct {
	Deployed     bool     `json:"saturn2Deployed"`
	ExitsEnabled bool     `json:"exitsEnabled"`
	Period       uint64   `json:"periodEpochs"`
	Buffer       uint64   `json:"proofBufferEpochs"`
	Threshold    *big.Int `json:"thresholdWei"`
	Block        uint64   `json:"executionBlock"`
	BlockHash    string   `json:"executionBlockHash"`
}

func (p Parameters) RequiredMisses() (uint64, error) {
	base := big.NewInt(1e18)
	if p.Period == 0 || p.Period > math.MaxUint64-p.Buffer || p.Threshold == nil || p.Threshold.Sign() < 0 || p.Threshold.Cmp(base) > 0 {
		return 0, fmt.Errorf("invalid performance challenge parameters")
	}
	n := new(big.Int).Sub(base, p.Threshold)
	n.Mul(n, new(big.Int).SetUint64(p.Period)).Add(n, new(big.Int).Sub(base, big.NewInt(1))).Div(n, base)
	return n.Uint64(), nil
}

func Retention(p Parameters, days, secondsPerEpoch, current uint64) (oldest, length uint64, err error) {
	if p.Deployed {
		if _, err = p.RequiredMisses(); err != nil {
			return
		}
		length = p.Period + p.Buffer
	} else {
		if days == 0 || secondsPerEpoch == 0 || days > math.MaxUint64/86400 {
			return 0, 0, fmt.Errorf("invalid bootstrap retention")
		}
		seconds := days * 86400
		length = seconds / secondsPerEpoch
		if seconds%secondsPerEpoch != 0 {
			length++
		}
	}
	if current >= length {
		oldest = current - length + 1
	}
	return
}

type Checkpoint struct {
	Epoch, Slot     uint64
	Root, StateRoot string
}

type EpochRecord struct {
	Epoch         uint64
	TargetRoot    string
	FinalizedRoot string
	Count         int
	Bits          []byte
}

func NewRecord(epoch uint64, count int) EpochRecord {
	return EpochRecord{Epoch: epoch, Count: count, Bits: make([]byte, (count+3)/4)}
}
func (r EpochRecord) Get(i int) Observation {
	if i < 0 || i >= r.Count {
		return Unknown
	}
	return Observation(r.Bits[i/4] >> (uint(i%4) * 2) & 3)
}
func (r EpochRecord) Set(i int, o Observation) {
	s := uint(i%4) * 2
	r.Bits[i/4] = r.Bits[i/4] & ^(byte(3)<<s) | byte(o)<<s
}

type Candidate struct {
	Validator
	Start         uint64     `json:"startEpoch"`
	Expiry        uint64     `json:"expiryEpoch"`
	Misses        uint64     `json:"confirmedMisses"`
	Required      uint64     `json:"requiredMisses"`
	Observed      uint64     `json:"observedEpochs"`
	Unknown       uint64     `json:"unknownEpochs"`
	Participation []*big.Int `json:"participation,omitempty"`
}

// ValidatorPerformance describes local observations, independent of whether a
// validator can be challenged. Performance excludes unknown epochs; coverage
// exposes that denominator separately.
type ValidatorPerformance struct {
	Validator
	StartEpoch                 *uint64  `json:"startEpoch"`
	EndEpoch                   *uint64  `json:"endEpoch"`
	ExpectedEpochs             uint64   `json:"expectedEpochs"`
	ObservedEpochs             uint64   `json:"observedEpochs"`
	TimelyEpochs               uint64   `json:"timelyEpochs"`
	MissedEpochs               uint64   `json:"missedEpochs"`
	UnknownEpochs              uint64   `json:"unknownEpochs"`
	CoveragePercent            float64  `json:"coveragePercent"`
	ObservedPerformancePercent *float64 `json:"observedPerformancePercent"`
	FirstObservedEpoch         *uint64  `json:"firstObservedEpoch"`
	LastObservedEpoch          *uint64  `json:"lastObservedEpoch"`
	RecentMissedEpochs         []uint64 `json:"recentMissedEpochs"`
}

type Status struct {
	Enabled               bool       `json:"enabled"`
	BackfillEnabled       bool       `json:"backfillEnabled"`
	Health                string     `json:"health"`
	Mode                  string     `json:"mode"`
	ActiveEpoch           *uint64    `json:"activeEpoch,omitempty"`
	LastError             string     `json:"lastError,omitempty"`
	Parameters            Parameters `json:"parameters"`
	ParametersVerified    bool       `json:"parametersVerified"`
	RetentionSource       string     `json:"retentionSource"`
	RetentionEpochs       uint64     `json:"retentionEpochs"`
	OldestEpoch           uint64     `json:"oldestEpoch"`
	CurrentEpoch          uint64     `json:"currentEpoch"`
	FinalizedEpoch        uint64     `json:"finalizedEpoch"`
	LatestMeasurableEpoch uint64     `json:"latestMeasurableEpoch"`
	LatestCollectedEpoch  uint64     `json:"latestCollectedEpoch"`
	Validators            int        `json:"validators"`
	CompleteEpochs        uint64     `json:"completeEpochs"`
	PendingEpochs         uint64     `json:"pendingEpochs"`
	CandidateCount        int        `json:"candidateCount"`
	RPCFailures           uint64     `json:"rpcFailures"`
	ProcessingSeconds     float64    `json:"processingSeconds"`
	StorageBytes          uint64     `json:"storageBytes"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

// ExecutionSource pins every registry/parameter read to a single execution block.
type ExecutionSource interface {
	Snapshot(context.Context, bool) (Parameters, error) // true requests finalized, false requests head
	Registry(context.Context, Parameters) ([]Validator, error)
	Eligible(context.Context, Parameters, Validator) (bool, error)
	NodePools(context.Context, string) ([]string, error) // finalized pool ownership, independent of Saturn 2
}
