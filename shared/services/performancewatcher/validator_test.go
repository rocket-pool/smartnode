package performancewatcher

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func (s *Store) validatorPerformance(ctx context.Context, v Validator, ordinal int, oldest, latest uint64) (*ValidatorPerformance, error) {
	results, err := s.validatorsPerformance(ctx, []indexedValidator{{v, ordinal}}, oldest, latest)
	if err != nil {
		return nil, err
	}
	return &results[0], nil
}

func TestValidatorPerformanceCoverageAndLifecycle(t *testing.T) {
	s := testStore(t)
	v := testValidator(11)
	v.Activation, v.Exit = 2, 8
	putObservations(t, s, 2, Missed, Missed) // Outside retention.
	putObservations(t, s, 3, Timely)         // Before this validator was discovered.
	putObservations(t, s, 4, Timely, Missed)
	// Epoch 5 has no row at all.
	putObservations(t, s, 6, Timely, Timely)
	putObservations(t, s, 7, Timely, Unknown)
	putObservations(t, s, 8, Missed, Missed) // At/after exit, no duty.
	putObservations(t, s, 10, Missed, Missed)
	r, err := s.validatorPerformance(context.Background(), v, 1, 3, 9)
	if err != nil {
		t.Fatal(err)
	}
	if r.Index != 11 || *r.StartEpoch != 3 || *r.EndEpoch != 7 || r.ExpectedEpochs != 5 {
		t.Fatalf("incorrect active range: %+v", r)
	}
	if r.TimelyEpochs != 1 || r.MissedEpochs != 1 || r.ObservedEpochs != 2 || r.UnknownEpochs != 3 || r.CoveragePercent != 40 || r.ObservedPerformancePercent == nil || *r.ObservedPerformancePercent != 50 {
		t.Fatalf("unknown epochs affected performance: %+v", r)
	}
	if *r.FirstObservedEpoch != 4 || *r.LastObservedEpoch != 6 || len(r.RecentMissedEpochs) != 1 || r.RecentMissedEpochs[0] != 4 {
		t.Fatalf("incorrect observed epochs: %+v", r)
	}

	v.Activation = 6
	r, err = s.validatorPerformance(context.Background(), v, 1, 0, 6)
	if err != nil || r.ExpectedEpochs != 1 || r.TimelyEpochs != 1 || r.MissedEpochs != 0 {
		t.Fatalf("activation/finality bounds ignored: %+v %v", r, err)
	}
}

func TestValidatorPerformanceEmptyAndRecentMisses(t *testing.T) {
	s := testStore(t)
	v := testValidator(10)
	r, err := s.validatorPerformance(context.Background(), v, 0, 1, 25)
	if err != nil || r.UnknownEpochs != 25 || r.ObservedPerformancePercent != nil || r.FirstObservedEpoch != nil || r.LastObservedEpoch != nil {
		t.Fatalf("empty history claims performance: %+v %v", r, err)
	}
	if _, err := json.Marshal(r); err != nil {
		t.Fatal(err)
	}
	for epoch := uint64(1); epoch <= 25; epoch++ {
		putObservations(t, s, epoch, Missed)
	}
	r, err = s.validatorPerformance(context.Background(), v, 0, 1, 25)
	if err != nil || r.MissedEpochs != 25 || r.UnknownEpochs != 0 || r.CoveragePercent != 100 || *r.ObservedPerformancePercent != 0 || len(r.RecentMissedEpochs) != 20 {
		t.Fatalf("incorrect complete history: %+v %v", r, err)
	}
	for i, epoch := range r.RecentMissedEpochs {
		if epoch != 25-uint64(i) {
			t.Fatal("recent misses must be bounded and newest first")
		}
	}
	v.Activation = 26
	r, err = s.validatorPerformance(context.Background(), v, 0, 1, 25)
	if err != nil || r.ExpectedEpochs != 0 || r.StartEpoch != nil || r.EndEpoch != nil || r.ObservedPerformancePercent != nil {
		t.Fatalf("no active range: %+v %v", r, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = s.validatorPerformance(ctx, v, 0, 1, 25); !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored cancellation: %v", err)
	}
}

func TestValidatorInspectionBeforeSaturnWithoutClients(t *testing.T) {
	s := testStore(t)
	v := testValidator(10)
	putObservations(t, s, 7, Missed)
	putObservations(t, s, 8, Timely)
	putObservations(t, s, 9, Missed) // Not measurable yet.
	b := NewBeacon(nil)
	b.GenesisTime = uint64(time.Now().Unix()) - 1000
	b.SlotsPerEpoch, b.SecondsPerSlot = 1, 100
	// Neither an execution source nor beacon requester is available. Inspection
	// must work from local data, even when parameters cannot be verified.
	w := New("", 201, false, b, nil, func(string, ...any) {})
	w.store, w.validators = s, []Validator{v}
	w.status.FinalizedEpoch, w.status.LatestMeasurableEpoch = 10, 8
	r, err := w.Validator(context.Background(), 10)
	if err != nil || r.ObservedEpochs != 2 || r.UnknownEpochs != 7 || r.MissedEpochs != 1 {
		t.Fatalf("pre-Saturn inspection failed: %+v %v", r, err)
	}
	if _, err := w.Validator(context.Background(), 11); !errors.Is(err, ErrValidatorNotTracked) {
		t.Fatalf("unknown validator: %v", err)
	}
	// On-chain retention still bounds the report when exits are paused or the
	// latest settings refresh is unavailable; expired disk rows are excluded.
	w.status.Parameters = testParams(2, 1, 940000000000000000)
	w.status.Parameters.ExitsEnabled = false
	r, err = w.Validator(context.Background(), 10)
	if err != nil || *r.StartEpoch != 8 || r.ObservedEpochs != 1 || r.MissedEpochs != 0 || r.TimelyEpochs != 1 {
		t.Fatalf("retention/paused inspection: %+v %v", r, err)
	}
}

func TestNodeValidatorInspection(t *testing.T) {
	s := testStore(t)
	mini, other, mega, mega2 := testValidator(10), testValidator(11), testValidator(12), testValidator(13)
	mini.Pool, mini.Kind = "0xAbCd", "minipool"
	other.Pool = "0x1234"
	mega.Pool, mega.Kind = "0xBeEf", "megapool"
	mega2.Pool, mega2.Kind, mega2.Activation = mega.Pool, mega.Kind, 8
	putObservations(t, s, 7, Missed, Missed, Timely) // Before the fourth validator's discovery/activation.
	putObservations(t, s, 8, Timely, Missed, Missed, Unknown)
	putObservations(t, s, 9, Missed, Missed, Missed, Missed) // Not measurable.
	b := NewBeacon(nil)
	b.GenesisTime = uint64(time.Now().Unix()) - 1000
	b.SlotsPerEpoch, b.SecondsPerSlot = 1, 100
	source := &mockExecutionSource{pools: []string{"0xabcd", "0xbeef", "0xabcd"}}
	w := New("", 201, false, b, source, func(string, ...any) {})
	w.store, w.validators = s, []Validator{mini, other, mega, mega2}
	w.status.FinalizedEpoch, w.status.LatestMeasurableEpoch = 10, 8
	results, err := w.NodeValidators(context.Background(), "node-address")
	if err != nil || len(results) != 3 || source.requestedNode != "node-address" {
		t.Fatalf("node lookup failed: %+v %v", results, err)
	}
	if results[0].Index != 10 || results[1].Index != 12 || results[2].Index != 13 {
		t.Fatalf("wrong ownership or duplicated validators: %+v", results)
	}
	if results[0].MissedEpochs != 1 || results[0].TimelyEpochs != 1 || results[1].MissedEpochs != 1 || results[1].RecentMissedEpochs[0] != 8 {
		t.Fatalf("incorrect registry ordinals: %+v", results)
	}
	if results[2].UnknownEpochs != 1 || results[2].ObservedPerformancePercent != nil {
		t.Fatalf("unknown history counted as misses: %+v", results[2])
	}
	source.pools = nil
	results, err = w.NodeValidators(context.Background(), "node-address")
	if err != nil || results == nil || len(results) != 0 {
		t.Fatalf("expected empty JSON array: %+v %v", results, err)
	}
	source.poolErr = errors.New("execution unavailable")
	if _, err = w.NodeValidators(context.Background(), "node-address"); !errors.Is(err, source.poolErr) {
		t.Fatalf("ownership failure hidden: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = w.NodeValidators(ctx, "node-address"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored cancellation: %v", err)
	}
}
