package assets

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

const Mainnet20ELHeaderTime = 1710394571

//go:embed rp-rewards-mainnet-20.json.gz
var mainnet20RewardsJSONGZ []byte
var mainnet20RewardsJSON []byte

func GetMainnet20RewardsJSON() []byte {
	if mainnet20RewardsJSON != nil {
		return mainnet20RewardsJSON
	}

	gz, err := gzip.NewReader(bytes.NewBuffer(mainnet20RewardsJSONGZ))
	if err != nil {
		panic(err)
	}
	defer gz.Close()
	mainnet20RewardsJSON, err = io.ReadAll(gz)
	if err != nil {
		panic(err)
	}
	return mainnet20RewardsJSON
}

//go:embed rp-minipool-performance-mainnet-20.json.gz
var Mainnet20MinipoolPerformanceJSONGZ []byte
var Mainnet20MinipoolPerformanceJSON []byte

func GetMainnet20MinipoolPerformanceJSON() []byte {
	if Mainnet20MinipoolPerformanceJSON != nil {
		return Mainnet20MinipoolPerformanceJSON
	}

	gz, err := gzip.NewReader(bytes.NewBuffer(Mainnet20MinipoolPerformanceJSONGZ))
	if err != nil {
		panic(err)
	}
	defer gz.Close()
	Mainnet20MinipoolPerformanceJSON, err = io.ReadAll(gz)
	if err != nil {
		panic(err)
	}
	return Mainnet20MinipoolPerformanceJSON
}

//go:embed rp-network-state-mainnet-20.json.gz
var Mainnet20NetworkStateJSONGZ []byte

var mainnet20RewardsState *state.NetworkState

func GetMainnet20RewardsState() *state.NetworkState {
	if mainnet20RewardsState != nil {
		return mainnet20RewardsState
	}

	// GUnzip the embedded bytes
	gz, err := gzip.NewReader(bytes.NewBuffer(Mainnet20NetworkStateJSONGZ))
	if err != nil {
		panic(err)
	}
	defer gz.Close()

	// Create a JSON decoder
	dec := json.NewDecoder(gz)

	// Decode the JSON
	result := state.NetworkState{}
	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}

	// Memoize the result
	mainnet20RewardsState = &result

	return mainnet20RewardsState
}

func GetRewardSnapshotEventInterval19() rewards.RewardsEvent {
	var rewardSnapshotEventInterval19 = rewards.RewardsEvent{
		Index:             big.NewInt(19),
		ExecutionBlock:    big.NewInt(19231284),
		ConsensusBlock:    big.NewInt(8429279),
		MerkleRoot:        common.HexToHash("0x35d1be64d49aa71dc5b5ea13dd6f91d8613c81aef2593796d6dee599cd228aea"),
		MerkleTreeCID:     "bafybeiazkzsqe7molppbhbxg2khdgocrip36eoezroa7anbe53za7mxjpq",
		IntervalsPassed:   big.NewInt(1),
		TreasuryRPL:       big.NewInt(0), // Set below
		TrustedNodeRPL:    []*big.Int{},  // XXX Not set, but probably not needed
		NodeRPL:           []*big.Int{},  // XXX Not set, but probably not needed
		NodeETH:           []*big.Int{},  // XXX Not set, but probably not needed
		UserETH:           big.NewInt(0), // XXX Not set, but probably not needed
		IntervalStartTime: time.Unix(1705556139, 0),
		IntervalEndTime:   time.Unix(1707975339, 0),
		SubmissionTime:    time.Unix(1707976475, 0),
	}
	rewardSnapshotEventInterval19.TreasuryRPL.SetString("0x0000000000000000000000000000000000000000000000f0a1e7585cd758ffe2", 16)
	return rewardSnapshotEventInterval19
}

//go:embed rp-network-critical-duties-mainnet-20.json.gz
var mainnet20CriticalDutiesSlotsGZ []byte
var mainnet20CriticalDutiesSlots *state.CriticalDutiesSlots

func GetMainnet20CriticalDutiesSlots() *state.CriticalDutiesSlots {
	if mainnet20CriticalDutiesSlots != nil {
		return mainnet20CriticalDutiesSlots
	}

	jsonReader, err := gzip.NewReader(bytes.NewBuffer(mainnet20CriticalDutiesSlotsGZ))
	if err != nil {
		panic(err)
	}
	defer jsonReader.Close()

	// Create a JSON decoder
	dec := json.NewDecoder(jsonReader)

	// Decode the JSON
	result := state.CriticalDutiesSlots{}
	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}

	mainnet20CriticalDutiesSlots = &result
	return mainnet20CriticalDutiesSlots
}
