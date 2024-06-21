package beacon

import (
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
