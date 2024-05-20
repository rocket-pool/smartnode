package beacon

import (
	"time"
)

type Eth2Config struct {
	GenesisForkVersion           []byte
	GenesisValidatorsRoot        []byte
	GenesisEpoch                 uint64
	GenesisTime                  uint64
	SecondsPerSlot               uint64
	SlotsPerEpoch                uint64
	SecondsPerEpoch              uint64
	EpochsPerSyncCommitteePeriod uint64
}

// GetSlotTime returns the time of a given slot for the network described by Eth2Config.
func (c *Eth2Config) GetSlotTime(slot uint64) time.Time {
	// In the interest of keeping this pure, we'll just return genesis time for slots before genesis
	if slot <= c.GenesisEpoch*c.SlotsPerEpoch {
		return time.Unix(int64(c.GenesisTime), 0)
	}
	// Genesis is slot 0 on mainnet, so we can subtract it safely
	slotsSinceGenesis := slot - (c.GenesisEpoch * c.SlotsPerEpoch)
	return time.Unix(int64(slotsSinceGenesis*c.SecondsPerSlot+c.GenesisTime), 0)
}

// FirstSlotAtLeast returns the first slot with a timestamp greater than or equal to t
func (c *Eth2Config) FirstSlotAtLeast(t int64) uint64 {
	if t <= 0 {
		return c.GenesisEpoch * c.SlotsPerEpoch
	}

	if uint64(t) <= c.GenesisTime {
		return c.GenesisEpoch * c.SlotsPerEpoch
	}

	secondsSinceGenesis := uint64(t) - c.GenesisTime

	var slotsSinceGenesis uint64
	// Avoid float error triggering ceil on quality with a modulo check
	if secondsSinceGenesis%c.SecondsPerSlot == 0 {
		slotsSinceGenesis = secondsSinceGenesis / c.SecondsPerSlot
	} else {
		// There must be a remainder
		slotsSinceGenesis = secondsSinceGenesis/c.SecondsPerSlot + 1
	}
	return c.GenesisEpoch*c.SlotsPerEpoch + slotsSinceGenesis
}
