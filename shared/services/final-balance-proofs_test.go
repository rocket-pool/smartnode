package services

import (
	"errors"
	"math/big"
	"slices"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

func TestFinalBalanceProofVersionWithdrawalBoundary(t *testing.T) {
	config := beacon.Eth2Config{GloasForkEpoch: 100, SlotsPerEpoch: 32}
	for _, tt := range []struct {
		name           string
		version        int64
		withdrawalSlot uint64
		wantError      bool
		wantBoundary   bool
	}{
		{name: "v1 before Gloas", version: 1, withdrawalSlot: 3199},
		{name: "v1 at Gloas must switch to v2", version: 1, withdrawalSlot: 3200, wantError: true, wantBoundary: true},
		{name: "v1 after Gloas must switch to v2", version: 1, withdrawalSlot: 3201, wantError: true, wantBoundary: true},
		{name: "v2 before Gloas rejected", version: 2, withdrawalSlot: 3199, wantError: true},
		{name: "v2 at Gloas", version: 2, withdrawalSlot: 3200},
		{name: "v2 after Gloas", version: 2, withdrawalSlot: 3201},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFinalBalanceProofVersion(big.NewInt(tt.version), tt.withdrawalSlot, config)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, want error = %t", err, tt.wantError)
			}
			if errors.Is(err, ErrGloasBoundaryReached) != tt.wantBoundary {
				t.Fatalf("error = %v, want Gloas fallback = %t", err, tt.wantBoundary)
			}
		})
	}
}

type finalBalanceBoundaryClient struct {
	beacon.Client
	requestedSlots []uint64
}

func (c *finalBalanceBoundaryClient) GetEth2Config() (beacon.Eth2Config, error) {
	return beacon.Eth2Config{GloasForkEpoch: 100, SlotsPerEpoch: 32}, nil
}

func (c *finalBalanceBoundaryClient) GetBeaconBlock(string) (beacon.BeaconBlock, bool, error) {
	// Finality has advanced past Gloas, but the withdrawal hint may predate it.
	return beacon.BeaconBlock{Slot: 3264, HasExecutionPayload: true}, true, nil
}

func (c *finalBalanceBoundaryClient) GetBeaconBlockSSZ(slot uint64) (*beacon.BeaconBlockSSZ, bool, error) {
	c.requestedSlots = append(c.requestedSlots, slot)
	return nil, false, nil
}

func TestFinalBalanceWithdrawalScanCrossesGloas(t *testing.T) {
	for _, tt := range []struct {
		name      string
		hint      uint64
		wantSlots []uint64
	}{
		{name: "pre-Gloas hint crosses activation", hint: 3198, wantSlots: []uint64{3198, 3199}},
		{name: "hint at activation", hint: 3200},
		{name: "hint after activation", hint: 3201},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bc := &finalBalanceBoundaryClient{}
			_, _, _, _, _, err := FindWithdrawalBlockAndArrayPosition(tt.hint, 42, bc)
			if !errors.Is(err, ErrGloasBoundaryReached) {
				t.Fatalf("expected Gloas fallback, got %v", err)
			}
			if !slices.Equal(bc.requestedSlots, tt.wantSlots) {
				t.Fatalf("requested slots = %v, want %v", bc.requestedSlots, tt.wantSlots)
			}
		})
	}
}
