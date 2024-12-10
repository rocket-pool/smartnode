package beacon

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Eth2Config struct {
	GenesisForkVersion           []byte `json:"genesis_fork_version"`
	GenesisValidatorsRoot        []byte `json:"genesis_validators_root"`
	GenesisEpoch                 uint64 `json:"genesis_epoch"`
	GenesisTime                  uint64 `json:"genesis_time"`
	SecondsPerSlot               uint64 `json:"seconds_per_slot"`
	SlotsPerEpoch                uint64 `json:"slots_per_epoch"`
	SecondsPerEpoch              uint64 `json:"seconds_per_epoch"`
	EpochsPerSyncCommitteePeriod uint64 `json:"epochs_per_sync_committee_period"`
}

func (c *Eth2Config) MarshalJSON() ([]byte, error) {
	// GenesisForkVersion and GenesisValidatorsRoot are returned as hex strings with 0x prefixes.
	// The other fields are returned as uint64s.
	type Alias Eth2Config
	return json.Marshal(&struct {
		GenesisForkVersion    string `json:"genesis_fork_version"`
		GenesisValidatorsRoot string `json:"genesis_validators_root"`
		*Alias
	}{
		GenesisForkVersion:    hexutil.Encode(c.GenesisForkVersion),
		GenesisValidatorsRoot: hexutil.Encode(c.GenesisValidatorsRoot),
		Alias:                 (*Alias)(c),
	})
}

func (c *Eth2Config) UnmarshalJSON(data []byte) error {
	type Alias Eth2Config
	aux := &struct {
		GenesisForkVersion    string `json:"genesis_fork_version"`
		GenesisValidatorsRoot string `json:"genesis_validators_root"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	c.GenesisForkVersion, err = hexutil.Decode(aux.GenesisForkVersion)
	if err != nil {
		return err
	}
	c.GenesisValidatorsRoot, err = hexutil.Decode(aux.GenesisValidatorsRoot)
	if err != nil {
		return err
	}
	return nil
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

func (c *Eth2Config) SlotToEpoch(slot uint64) uint64 {
	return slot / c.SlotsPerEpoch
}

func (c *Eth2Config) EpochToSlot(epoch uint64) uint64 {
	return epoch * c.SlotsPerEpoch
}

func (c *Eth2Config) SlotOfEpoch(epoch uint64, slot uint64) (uint64, error) {
	if slot > c.SlotsPerEpoch-1 {
		return 0, fmt.Errorf("slot %d is not in range 0 - %d", slot, c.SlotsPerEpoch-1)
	}
	return epoch*c.SlotsPerEpoch + slot, nil
}

func (c *Eth2Config) LastSlotOfEpoch(epoch uint64) uint64 {
	out, err := c.SlotOfEpoch(epoch, c.SlotsPerEpoch-1)
	if err != nil {
		panic("SlotOfEpoch should never return an error when passed SlotsPerEpoch - 1")
	}
	return out
}

func (c *Eth2Config) FirstSlotOfEpoch(epoch uint64) uint64 {
	out, err := c.SlotOfEpoch(epoch, 0)
	if err != nil {
		panic("SlotOfEpoch should never return an error when passed 0")
	}
	return out
}
