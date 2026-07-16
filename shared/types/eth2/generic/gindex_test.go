package generic

import "testing"

func TestContainerFieldGindex(t *testing.T) {
	// Pre-Gloas Deneb-style: 28 fields → width 32; validators at index 11 → 43
	if got := ContainerFieldGindex(28, BeaconStateValidatorsIndex); got != 43 {
		t.Fatalf("validators (28 fields): got %d want 43", got)
	}
	// Slot at index 2 → 34
	if got := ContainerFieldGindex(28, BeaconStateSlotIndex); got != 34 {
		t.Fatalf("slot (28 fields): got %d want 34", got)
	}
	// Fulu-style: 37+ fields → width 64; validators → 75
	if got := ContainerFieldGindex(38, BeaconStateValidatorsIndex); got != 75 {
		t.Fatalf("validators (38 fields): got %d want 75", got)
	}
}

func TestGetGeneralizedIndexForValidatorClassic(t *testing.T) {
	// validators field at classic gindex 43 (deneb), index 0
	// = 43 * 2 * 2^40 + 0
	validatorsField := ContainerFieldGindex(28, BeaconStateValidatorsIndex)
	got := GetGeneralizedIndexForValidator(0, validatorsField)
	want := validatorsField * 2 * beaconStateValidatorsMaxLength
	if got != want {
		t.Fatalf("validator 0: got %d want %d", got, want)
	}
	got1 := GetGeneralizedIndexForValidator(1, validatorsField)
	if got1 != want+1 {
		t.Fatalf("validator 1: got %d want %d", got1, want+1)
	}
}

func TestGetGeneralizedIndexForVectorAndList(t *testing.T) {
	// Vector of length 8 at root 1: element 3 at 1*8+3 = 11
	if got := GetGeneralizedIndexForVectorElement(1, 8, 3); got != 11 {
		t.Fatalf("vector: got %d want 11", got)
	}
	// List capacity 16 at root 1: element 3 at 1*2*16+3 = 35
	if got := GetGeneralizedIndexForListElement(1, 16, 3); got != 35 {
		t.Fatalf("list: got %d want 35", got)
	}
}
