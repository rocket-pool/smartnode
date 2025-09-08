package rewards

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/bindings/types"
)

type PerformanceFile_v1 struct {
	RewardsFileVersion  uint64                                     `json:"rewardsFileVersion"`
	RulesetVersion      uint64                                     `json:"rulesetVersion"`
	Index               uint64                                     `json:"index"`
	Network             string                                     `json:"network"`
	StartTime           time.Time                                  `json:"startTime"`
	EndTime             time.Time                                  `json:"endTime"`
	ConsensusStartBlock uint64                                     `json:"consensusStartBlock"`
	ConsensusEndBlock   uint64                                     `json:"consensusEndBlock"`
	ExecutionStartBlock uint64                                     `json:"executionStartBlock"`
	ExecutionEndBlock   uint64                                     `json:"executionEndBlock"`
	MinipoolPerformance map[common.Address]*MinipoolPerformance_v2 `json:"minipoolPerformance"`
	MegapoolPerformance map[common.Address]*MegapoolPerformance_v1 `json:"megapoolPerformance"`
	BonusScalar         *QuotedBigInt                              `json:"bonusScalar"`
}

// Type assertion to implement IPerformanceFile
var _ IPerformanceFile = (*PerformanceFile_v1)(nil)

type MegapoolPerformanceMap map[types.ValidatorPubkey]*MegapoolValidatorPerformance_v1

type MegapoolPerformance_v1 struct {
	VoterShare           *QuotedBigInt          `json:"voterShare"`
	ValidatorPerformance MegapoolPerformanceMap `json:"validatorPerformance"`
}

// MegapoolPerformanceMap has a custom JSON marshaler to avoid the issue with ValidatorPubkey not being a valid dict key.
// encoding/json/v2 will fix this once it's stable, and the custom marshaler can be removed.
func (m MegapoolPerformanceMap) MarshalJSON() ([]byte, error) {
	out := make(map[string]*MegapoolValidatorPerformance_v1)
	for pubkey, perf := range m {
		out[pubkey.Hex()] = perf
	}
	return json.Marshal(out)
}

// And a custom unmarshaler to avoid the issue with ValidatorPubkey not being a valid dict key.
// encoding/json/v2 will fix this once it's stable, and the custom unmarshaler can be removed.
func (m *MegapoolPerformanceMap) UnmarshalJSON(data []byte) error {
	var out map[string]*MegapoolValidatorPerformance_v1
	err := json.Unmarshal(data, &out)
	if err != nil {
		return err
	}
	*m = make(MegapoolPerformanceMap, len(out))
	for pubkey, perf := range out {
		pubkeyBytes, err := hex.DecodeString(pubkey)
		if err != nil {
			return fmt.Errorf("error decoding pubkey %s: %w", pubkey, err)
		}
		(*m)[types.ValidatorPubkey(pubkeyBytes)] = perf
	}
	return nil
}

// Conveniently, v2 minipool performance tracks all the same fields
// as a single megapool validator, but has 3 extras.
// Those fields are omitempty anyway, so we will just leave them nil
type MegapoolValidatorPerformance_v1 = MinipoolPerformance_v2

// Type assertion to implement ISmoothingPoolPerformance
var _ ISmoothingPoolPerformance = (*MegapoolValidatorPerformance_v1)(nil)

func (f *PerformanceFile_v1) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

func (f *PerformanceFile_v1) SerializeSSZ() ([]byte, error) {
	return nil, fmt.Errorf("ssz format not implemented for performance files")
}

func (f *PerformanceFile_v1) SerializeHuman() ([]byte, error) {
	return json.MarshalIndent(f, "", "\t")
}

func (f *PerformanceFile_v1) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

func (f *PerformanceFile_v1) GetMinipoolAddresses() []common.Address {
	addresses := make([]common.Address, len(f.MinipoolPerformance))
	i := 0
	for address := range f.MinipoolPerformance {
		addresses[i] = address
		i++
	}
	return addresses
}

func (f *PerformanceFile_v1) GetMegapoolAddresses() []common.Address {
	addresses := make([]common.Address, len(f.MegapoolPerformance))
	i := 0
	for address := range f.MegapoolPerformance {
		addresses[i] = address
		i++
	}
	return addresses
}

func (f *PerformanceFile_v1) GetMegapoolValidatorPubkeys(megapoolAddress common.Address) ([]types.ValidatorPubkey, error) {
	perf, exists := f.MegapoolPerformance[megapoolAddress]
	if !exists {
		return nil, fmt.Errorf("megapool %s not found", megapoolAddress)
	}
	numValidators := len(perf.ValidatorPerformance)
	pubkeys := make([]types.ValidatorPubkey, numValidators)
	i := 0
	for pubkey := range perf.ValidatorPerformance {
		pubkeys[i] = pubkey
		i++
	}
	return pubkeys, nil
}

func (f *PerformanceFile_v1) GetMegapoolPerformance(megapoolAddress common.Address, pubkey types.ValidatorPubkey) (ISmoothingPoolPerformance, bool) {
	megapoolPerf, exists := f.MegapoolPerformance[megapoolAddress]
	if !exists {
		return nil, false
	}
	validatorPerf, exists := megapoolPerf.ValidatorPerformance[pubkey]
	if !exists {
		return nil, false
	}
	return validatorPerf, true
}

func (f *PerformanceFile_v1) GetMinipoolPerformance(minipoolAddress common.Address) (ISmoothingPoolPerformance, bool) {
	perf, exists := f.MinipoolPerformance[minipoolAddress]
	return perf, exists
}
