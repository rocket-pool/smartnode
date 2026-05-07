package megapool

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// makeValidators builds a slice of MegapoolValidatorDetails with only the
// fields that sortExitingValidatorsBySweep cares about populated. ValidatorId
// is set to a distinct value derived from the index so we can also assert that
// the original elements (not just their indices) are preserved.
func makeValidators(indices ...uint64) []api.MegapoolValidatorDetails {
	out := make([]api.MegapoolValidatorDetails, 0, len(indices))
	for i, idx := range indices {
		out = append(out, api.MegapoolValidatorDetails{
			ValidatorId:    uint32(1000 + i),
			ValidatorIndex: idx,
		})
	}
	return out
}

func indexesOf(validators []api.MegapoolValidatorDetails) []uint64 {
	out := make([]uint64, len(validators))
	for i, v := range validators {
		out[i] = v.ValidatorIndex
	}
	return out
}

func TestSortExitingValidatorsBySweep(t *testing.T) {
	tests := []struct {
		name                  string
		input                 []uint64
		lastWithdrawnIndex    uint64
		hasLastWithdrawnIndex bool
		want                  []uint64
	}{
		{
			name:                  "empty slice does not panic",
			input:                 []uint64{},
			lastWithdrawnIndex:    100,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{},
		},
		{
			name:                  "single validator above pivot",
			input:                 []uint64{200},
			lastWithdrawnIndex:    100,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{200},
		},
		{
			name:                  "single validator below pivot",
			input:                 []uint64{50},
			lastWithdrawnIndex:    100,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{50},
		},
		{
			name:                  "no pivot falls back to ascending order",
			input:                 []uint64{300, 50, 200, 100},
			hasLastWithdrawnIndex: false,
			want:                  []uint64{50, 100, 200, 300},
		},
		{
			name:                  "pivot in the middle splits and sorts each half",
			input:                 []uint64{300, 50, 200, 100, 400, 25},
			lastWithdrawnIndex:    150,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{200, 300, 400, 25, 50, 100},
		},
		{
			name:                  "pivot equal to a validator index puts that validator after the wrap",
			input:                 []uint64{100, 50, 150, 200},
			lastWithdrawnIndex:    100,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{150, 200, 50, 100},
		},
		{
			name:                  "pivot above all validators wraps everyone (plain ascending)",
			input:                 []uint64{50, 30, 10, 20},
			lastWithdrawnIndex:    1000,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{10, 20, 30, 50},
		},
		{
			name:                  "pivot below all validators leaves them all 'after' (plain ascending)",
			input:                 []uint64{50, 30, 10, 20},
			lastWithdrawnIndex:    0,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{10, 20, 30, 50},
		},
		{
			name:                  "already in sweep order is unchanged",
			input:                 []uint64{200, 300, 400, 25, 50, 100},
			lastWithdrawnIndex:    150,
			hasLastWithdrawnIndex: true,
			want:                  []uint64{200, 300, 400, 25, 50, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validators := makeValidators(tt.input...)

			sortExitingValidatorsBySweep(validators, tt.lastWithdrawnIndex, tt.hasLastWithdrawnIndex)

			got := indexesOf(validators)
			// Full struct is unreadable in %v; print index order and compact id@index chain.
			t.Logf("final sweep order (validator indices): %v", got)
			parts := make([]string, len(validators))
			for i, v := range validators {
				parts[i] = fmt.Sprintf("%d@%d", v.ValidatorId, v.ValidatorIndex)
			}
			idIndexLine := strings.Join(parts, " → ")
			if idIndexLine == "" {
				idIndexLine = "(empty)"
			}
			t.Logf("final sweep order (id@index): %s", idIndexLine)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected order: got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSortExitingValidatorsBySweepPreservesElements confirms that the helper
// reorders the original elements (preserving fields like ValidatorId) rather
// than producing copies that drop unrelated data.
func TestSortExitingValidatorsBySweepPreservesElements(t *testing.T) {
	validators := []api.MegapoolValidatorDetails{
		{ValidatorId: 11, ValidatorIndex: 300},
		{ValidatorId: 22, ValidatorIndex: 50},
		{ValidatorId: 33, ValidatorIndex: 200},
		{ValidatorId: 44, ValidatorIndex: 100},
	}

	sortExitingValidatorsBySweep(validators, 150, true)

	wantIndexes := []uint64{200, 300, 50, 100}
	wantIDs := []uint32{33, 11, 22, 44}
	for i, v := range validators {
		if v.ValidatorIndex != wantIndexes[i] || v.ValidatorId != wantIDs[i] {
			t.Fatalf("position %d: got (id=%d, index=%d), want (id=%d, index=%d)",
				i, v.ValidatorId, v.ValidatorIndex, wantIDs[i], wantIndexes[i])
		}
	}
}
