package beacon

import (
	"slices"
	"testing"
	"time"
)

var config = &Eth2Config{
	GenesisEpoch:    10,
	GenesisTime:     10000,
	SecondsPerSlot:  4,
	SlotsPerEpoch:   32,
	SecondsPerEpoch: 32 * 4,
}

func TestGetSlotTime(t *testing.T) {
	genesis := config.GetSlotTime(0)
	if !genesis.Equal(time.Unix(int64(config.GenesisTime), 0)) {
		t.Fatalf("slot 0 should be at genesis (%d) but was at %s", config.GenesisTime, genesis)
	}

	slotPlusTen := config.GenesisEpoch*config.SlotsPerEpoch + 10
	slotPlusTenTime := config.GetSlotTime(slotPlusTen)
	expectedTime := time.Unix(int64(config.SecondsPerSlot*10+config.GenesisTime), 0)
	if !slotPlusTenTime.Equal(expectedTime) {
		t.Fatalf("slot +10 should be at genesis (%d) but was at %s", config.GenesisTime, genesis)
	}
}

func TestFirstSlotAtLeast(t *testing.T) {
	genesis := config.FirstSlotAtLeast(30)
	if genesis != config.GenesisEpoch*config.SlotsPerEpoch {
		t.Fatalf("should have gotten the genesis slot (%d), instead got %d", config.GenesisEpoch*config.SlotsPerEpoch, genesis)
	}

	// Whole multiple
	slots := uint64(9000000)
	st := config.GenesisTime + config.SecondsPerSlot*slots
	result := config.FirstSlotAtLeast(int64(st))
	if result != slots+config.GenesisEpoch*config.SlotsPerEpoch {
		t.Fatal("Whole number seconds shouldn't round up")
	}

	// Partial multiple rounds up
	st = config.GenesisTime + config.SecondsPerSlot*slots - config.SecondsPerSlot/2
	result = config.FirstSlotAtLeast(int64(st))
	if result != slots+config.GenesisEpoch*config.SlotsPerEpoch {
		t.Fatal("Whole number seconds shouldn't round up")
	}

	// Smallest fractional amount rounds up
	st = config.GenesisTime + config.SecondsPerSlot*slots - config.SecondsPerSlot + 1
	result = config.FirstSlotAtLeast(int64(st))
	if result != slots+config.GenesisEpoch*config.SlotsPerEpoch {
		t.Fatal("Whole number seconds shouldn't round up")
	}
}

func TestMarshalJSON(t *testing.T) {
	config := &Eth2Config{
		GenesisForkVersion:           []byte{0x00, 0x00, 0x00, 0x08},
		GenesisValidatorsRoot:        []byte{0xfe, 0x44, 0x33, 0x22},
		GenesisEpoch:                 10,
		GenesisTime:                  10000,
		SecondsPerSlot:               4,
		SlotsPerEpoch:                32,
		SecondsPerEpoch:              32 * 4,
		EpochsPerSyncCommitteePeriod: 256,
	}

	json, err := config.MarshalJSON()
	if err != nil {
		t.Fatalf("error marshalling config: %v", err)
	}

	unmarshalled := &Eth2Config{}
	err = unmarshalled.UnmarshalJSON(json)
	if err != nil {
		t.Fatalf("error unmarshalling config: %v", err)
	}

	if !slices.Equal(unmarshalled.GenesisForkVersion, config.GenesisForkVersion) {
		t.Fatalf("genesis fork version should be %v, instead got %v", config.GenesisForkVersion, unmarshalled.GenesisForkVersion)
	}

	if !slices.Equal(unmarshalled.GenesisValidatorsRoot, config.GenesisValidatorsRoot) {
		t.Fatalf("genesis validators root should be %v, instead got %v", config.GenesisValidatorsRoot, unmarshalled.GenesisValidatorsRoot)
	}

	if unmarshalled.GenesisEpoch != config.GenesisEpoch {
		t.Fatalf("genesis epoch should be %v, instead got %v", config.GenesisEpoch, unmarshalled.GenesisEpoch)
	}

	if unmarshalled.GenesisTime != config.GenesisTime {
		t.Fatalf("genesis time should be %v, instead got %v", config.GenesisTime, unmarshalled.GenesisTime)
	}

	if unmarshalled.SecondsPerSlot != config.SecondsPerSlot {
		t.Fatalf("seconds per slot should be %v, instead got %v", config.SecondsPerSlot, unmarshalled.SecondsPerSlot)
	}

	if unmarshalled.SlotsPerEpoch != config.SlotsPerEpoch {
		t.Fatalf("slots per epoch should be %v, instead got %v", config.SlotsPerEpoch, unmarshalled.SlotsPerEpoch)
	}

	if unmarshalled.EpochsPerSyncCommitteePeriod != config.EpochsPerSyncCommitteePeriod {
		t.Fatalf("epochs per sync committee period should be %v, instead got %v", config.EpochsPerSyncCommitteePeriod, unmarshalled.EpochsPerSyncCommitteePeriod)
	}
}
