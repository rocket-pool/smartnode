package megapool

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func deficitTestWei(eth string) *big.Int {
	value, ok := new(big.Rat).SetString(eth)
	if !ok {
		panic("invalid test amount")
	}
	value.Mul(value, new(big.Rat).SetInt(big.NewInt(1e18)))
	if !value.IsInt() {
		panic("test amount is not an integer wei value")
	}
	return new(big.Int).Set(value.Num())
}

func TestPlanDeficitExits(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		debt, assets, bond, queued string
		active, exiting, wantCount uint32
		wantBefore, wantAfter      string
		wantCanExit                bool
	}{
		{"below threshold", "0.1", "0", "16", "0", 4, 0, 0, "0.1", "0.1", false},
		{"at threshold", "0.2", "0", "16", "0", 4, 0, 1, "0.2", "0", true},
		{"strictly below threshold", "4.2", "0", "16", "0", 4, 0, 2, "4.2", "0", true},
		{"assets offset debt", "4.4", "0.3", "16", "0", 4, 0, 1, "4.1", "0.1", true},
		{"assets exceed debt", "1", "2", "16", "0", 4, 0, 0, "0", "0", false},
		{"pending exits already enough", "4.1", "0", "16", "0", 4, 1, 0, "0.1", "0.1", false},
		{"pending exits counted once", "4.2", "0", "16", "0", 4, 1, 1, "0.2", "0", true},
		{"underbonded release clamped", "1", "0", "6", "0", 3, 0, 2, "1", "0", true},
		{"queued bond included", "1", "0", "8", "4", 3, 0, 1, "1", "0", true},
		{"principal release capped", "32.2", "0", "100", "0", 3, 0, 2, "32.2", "0", true},
		{"release capped at node bond", "1.2", "0", "1", "100", 3, 0, 3, "1.2", "0.2", true},
		{"deficit remains above threshold", "10", "0", "4", "0", 1, 0, 1, "10", "6", true},
		{"exit every validator while above threshold", "10", "0", "4", "0", 3, 0, 3, "10", "6", true},
		{"no validators", "1", "0", "0", "0", 0, 0, 0, "1", "1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			debt, assets, bond, queued := deficitTestWei(tc.debt), deficitTestWei(tc.assets), deficitTestWei(tc.bond), deficitTestWei(tc.queued)
			plan, err := planDeficitExits(debt, assets, bond, queued, deficitTestWei("0.2"), tc.active, tc.exiting, func(count uint32) (*big.Int, error) {
				return new(big.Int).Mul(new(big.Int).SetUint64(uint64(count)), deficitTestWei("4")), nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if plan.CanExit != tc.wantCanExit || plan.ValidatorsRequired != tc.wantCount {
				t.Fatalf("got canExit=%v count=%d, want %v %d (%s)", plan.CanExit, plan.ValidatorsRequired, tc.wantCanExit, tc.wantCount, plan.Reason)
			}
			if plan.DeficitAfterPendingExits.Cmp(deficitTestWei(tc.wantBefore)) != 0 || plan.ProjectedDeficit.Cmp(deficitTestWei(tc.wantAfter)) != 0 {
				t.Fatalf("got before=%s after=%s, want %s ETH and %s ETH", plan.DeficitAfterPendingExits, plan.ProjectedDeficit, tc.wantBefore, tc.wantAfter)
			}
			if debt.Cmp(deficitTestWei(tc.debt)) != 0 || assets.Cmp(deficitTestWei(tc.assets)) != 0 || bond.Cmp(deficitTestWei(tc.bond)) != 0 || queued.Cmp(deficitTestWei(tc.queued)) != 0 {
				t.Fatal("planning mutated its input balances")
			}
		})
	}
}

func TestPlanDeficitExitsUsesBondCurve(t *testing.T) {
	curve := []string{"0", "4", "8", "10", "12"}
	plan, err := planDeficitExits(deficitTestWei("2.2"), new(big.Int), deficitTestWei("12"), new(big.Int), deficitTestWei("0.2"), 4, 0, func(count uint32) (*big.Int, error) {
		return deficitTestWei(curve[count]), nil
	})
	if err != nil || plan.ValidatorsRequired != 2 {
		t.Fatalf("expected two exits across the bond curve, got %+v, %v", plan, err)
	}
}

func TestPlanDeficitExitsErrors(t *testing.T) {
	wantErr := errors.New("bond query failed")
	_, err := planDeficitExits(deficitTestWei("1"), new(big.Int), deficitTestWei("4"), new(big.Int), deficitTestWei("0.2"), 1, 0, func(uint32) (*big.Int, error) {
		return nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v, want bond query error", err)
	}
	_, err = planDeficitExits(nil, nil, nil, nil, nil, 1, 2, nil)
	if err == nil {
		t.Fatal("expected inconsistent counts to be rejected")
	}
}

func TestApplyEligibleValidators(t *testing.T) {
	project := func(additional uint32) (*big.Int, bool, error) {
		if additional == 0 {
			return deficitTestWei("10"), false, nil
		}
		return deficitTestWei("6"), false, nil
	}
	result := &api.MegapoolDeficitResponse{
		CanExit: true, ValidatorsRequired: 3, ProjectedDeficit: deficitTestWei("0"), DeficitAfterPendingExits: deficitTestWei("10"),
	}
	if err := applyEligibleValidators(result, []uint32{4}, project); err != nil {
		t.Fatal(err)
	}
	if !result.CanExit || result.ValidatorsRequired != 1 || len(result.ValidatorIds) != 1 || result.ValidatorIds[0] != 4 || result.ProjectedDeficit.Cmp(deficitTestWei("6")) != 0 {
		t.Fatalf("shorter list was not accepted: %+v", result)
	}

	refused := &api.MegapoolDeficitResponse{CanExit: true, ValidatorsRequired: 2, ProjectedDeficit: deficitTestWei("0")}
	if err := applyEligibleValidators(refused, []uint32{}, func(uint32) (*big.Int, bool, error) {
		return deficitTestWei("0"), true, nil
	}); err != nil || refused.CanExit || refused.ValidatorsRequired != 2 {
		t.Fatalf("empty eligible set should refuse the planned count, got %+v, %v", refused, err)
	}

	disallowed := &api.MegapoolDeficitResponse{CanExit: true, ValidatorsRequired: 3, ProjectedDeficit: deficitTestWei("0")}
	if err := applyEligibleValidators(disallowed, []uint32{1}, func(additional uint32) (*big.Int, bool, error) {
		return deficitTestWei("0.1"), additional == 0, nil
	}); err != nil || disallowed.CanExit || disallowed.ValidatorsRequired != 3 {
		t.Fatalf("shorter list below the floor should be refused, got %+v, %v", disallowed, err)
	}
}

type deficitValidatorPool struct {
	megapool.Megapool
	validators []megapool.ValidatorInfo
}

func (mp *deficitValidatorPool) GetValidatorCount(opts *bind.CallOpts) (uint32, error) {
	if err := requirePinnedBlock(opts); err != nil {
		return 0, err
	}
	return uint32(len(mp.validators)), nil
}

func (mp *deficitValidatorPool) GetValidatorInfo(id uint32, opts *bind.CallOpts) (megapool.ValidatorInfo, error) {
	if err := requirePinnedBlock(opts); err != nil {
		return megapool.ValidatorInfo{}, err
	}
	return mp.validators[id], nil
}

func requirePinnedBlock(opts *bind.CallOpts) error {
	if opts == nil || opts.BlockNumber == nil || opts.BlockNumber.Cmp(big.NewInt(123)) != 0 {
		return errors.New("validator info was not read at the pinned block")
	}
	return nil
}

func TestFirstDeficitExitValidators(t *testing.T) {
	mp := &deficitValidatorPool{validators: []megapool.ValidatorInfo{
		{InQueue: true},
		{Staked: true, Exiting: true},
		{Staked: true},
		{Staked: true, Dissolved: true},
		{Staked: true, Exited: true},
		{Staked: true, Locked: true}, // forceExit also unlocks the validator.
		{Staked: true},
	}}
	for _, tc := range []struct {
		count uint32
		want  []uint32
	}{{2, []uint32{2, 5}}, {4, []uint32{2, 5, 6}}, {0, []uint32{}}} {
		ids, err := firstDeficitExitValidators(mp, tc.count, &bind.CallOpts{BlockNumber: big.NewInt(123)})
		if err != nil || !reflect.DeepEqual(ids, tc.want) {
			t.Fatalf("count %d: got %v, %v; want %v", tc.count, ids, err, tc.want)
		}
	}
}
