package performancewatcher

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// Validator reads retained observations without querying either client or
// requiring deployed, enabled, or freshly verified challenge parameters.
func (w *Worker) Validator(ctx context.Context, index uint64) (*ValidatorPerformance, error) {
	results, err := w.inspectValidators(ctx, func(v Validator) bool { return v.Index == index })
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("%w: %d", ErrValidatorNotTracked, index)
	}
	return &results[0], nil
}

// NodeValidators resolves finalized pool ownership, then reads all matching
// validators in a single local history scan. It makes no beacon requests.
func (w *Worker) NodeValidators(ctx context.Context, address string) ([]ValidatorPerformance, error) {
	pools, err := w.source.NodePools(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("resolve node validator ownership: %w", err)
	}
	owned := make(map[string]bool, len(pools))
	for _, pool := range pools {
		owned[strings.ToLower(pool)] = true
	}
	return w.inspectValidators(ctx, func(v Validator) bool { return owned[strings.ToLower(v.Pool)] })
}

type indexedValidator struct {
	validator Validator
	ordinal   int
}

func (w *Worker) inspectValidators(ctx context.Context, include func(Validator) bool) ([]ValidatorPerformance, error) {
	select {
	case w.export <- struct{}{}:
		defer func() { <-w.export }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.store == nil || w.status.FinalizedEpoch < 2 {
		return nil, fmt.Errorf("watcher observations are unavailable")
	}
	current := w.currentEpoch()
	if current == 0 {
		return nil, fmt.Errorf("no measurable epochs yet")
	}
	oldest, _, err := Retention(w.status.Parameters, w.days, w.beacon.SecondsPerSlot*w.beacon.SlotsPerEpoch, current)
	if err != nil {
		return nil, err
	}
	selected := []indexedValidator{}
	for ordinal, v := range w.validators {
		if include(v) {
			selected = append(selected, indexedValidator{v, ordinal})
		}
	}
	return w.store.validatorsPerformance(ctx, selected, oldest, min(w.status.LatestMeasurableEpoch, current-1))
}

func (s *Store) validatorsPerformance(ctx context.Context, selected []indexedValidator, oldest, latest uint64) ([]ValidatorPerformance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	const recentMissLimit = 20
	results := make([]ValidatorPerformance, len(selected))
	active := false
	for i, entry := range selected {
		v := entry.validator
		results[i] = ValidatorPerformance{Validator: v, RecentMissedEpochs: make([]uint64, 0, recentMissLimit)}
		if v.Exit == 0 {
			continue
		}
		start, end := max(oldest, v.Activation), min(latest, v.Exit-1)
		if start > end {
			continue
		}
		results[i].StartEpoch, results[i].EndEpoch = &start, &end
		results[i].ExpectedEpochs = end - start + 1
		active = true
	}
	if !active {
		return results, nil
	}
	err := s.Scan(oldest, func(r EpochRecord) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if r.Epoch > latest {
			return nil
		}
		for i, entry := range selected {
			result := &results[i]
			if result.StartEpoch == nil || r.Epoch < *result.StartEpoch || r.Epoch > *result.EndEpoch {
				continue
			}
			switch r.Get(entry.ordinal) {
			case Timely:
				result.TimelyEpochs++
			case Missed:
				result.MissedEpochs++
				if len(result.RecentMissedEpochs) == recentMissLimit {
					copy(result.RecentMissedEpochs, result.RecentMissedEpochs[1:])
					result.RecentMissedEpochs[recentMissLimit-1] = r.Epoch
				} else {
					result.RecentMissedEpochs = append(result.RecentMissedEpochs, r.Epoch)
				}
			default:
				// Absent rows, shorter rows from before discovery, explicit unknowns,
				// and inactive observations cannot establish timely/missed duties.
				continue
			}
			result.ObservedEpochs++
			epoch := r.Epoch
			if result.FirstObservedEpoch == nil {
				result.FirstObservedEpoch = &epoch
			}
			result.LastObservedEpoch = &epoch
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for i := range results {
		result := &results[i]
		result.UnknownEpochs = result.ExpectedEpochs - result.ObservedEpochs
		if result.ExpectedEpochs > 0 {
			result.CoveragePercent = 100 * float64(result.ObservedEpochs) / float64(result.ExpectedEpochs)
		}
		if result.ObservedEpochs > 0 {
			performance := 100 * float64(result.TimelyEpochs) / float64(result.ObservedEpochs)
			result.ObservedPerformancePercent = &performance
		}
		slices.Reverse(result.RecentMissedEpochs)
	}
	return results, nil
}
