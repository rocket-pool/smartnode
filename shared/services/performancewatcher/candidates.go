package performancewatcher

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"
)

// Candidates transposes packed epoch records into packed per-validator histories.
// Memory is two bits per validator/retained epoch, independent of miss density.
// Only one validator's miss list and bitmap are assembled at a time.
func (s *Store) Candidates(ctx context.Context, validators []Validator, p Parameters, current, oldest, latest uint64, only *uint64, bitmap bool) ([]Candidate, error) {
	if current == 0 {
		return nil, nil
	}
	latest = min(latest, current-1)
	if !p.Deployed || !p.ExitsEnabled || latest < oldest {
		return nil, nil
	}
	required, err := p.RequiredMisses()
	if err != nil {
		return nil, err
	}
	length := latest - oldest + 1
	if length > uint64(math.MaxInt)-3 {
		return nil, fmt.Errorf("retention interval too large")
	}
	width := int((length + 3) / 4)
	selected := make([]int, 0, len(validators))
	for i, v := range validators {
		if v.Eligible && (only == nil || *only == v.Index) {
			selected = append(selected, i)
		}
	}
	// Bound peak scan memory even if the validator registry grows. Each batch
	// uses a fresh consistent database snapshot; individual epoch rows are atomic.
	const maxHistoryBytes = 128 << 20
	if width > maxHistoryBytes {
		return nil, fmt.Errorf("retained history exceeds scan limit")
	}
	batchSize := max(1, maxHistoryBytes/width)
	var out []Candidate
	for offset := 0; offset < len(selected); offset += batchSize {
		items, err := s.scanCandidates(ctx, validators, selected[offset:min(offset+batchSize, len(selected))], p, current, oldest, latest, required, width, bitmap)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out, nil
}

func (s *Store) scanCandidates(ctx context.Context, validators []Validator, selected []int, p Parameters, current, oldest, latest, required uint64, width int, bitmap bool) ([]Candidate, error) {
	history := make([]byte, width*len(selected))
	err := s.Scan(oldest, func(r EpochRecord) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if r.Epoch > latest {
			return nil
		}
		offset := r.Epoch - oldest
		for j, ordinal := range selected {
			history[j*width+int(offset/4)] |= byte(r.Get(ordinal)) << (offset % 4 * 2)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var out []Candidate
	for j, ordinal := range selected {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		v := validators[ordinal]
		bits := history[j*width : (j+1)*width]
		get := func(epoch uint64) Observation {
			if epoch < oldest || epoch > latest || epoch >= current {
				return Unknown
			}
			offset := epoch - oldest
			return Observation(bits[offset/4] >> (offset % 4 * 2) & 3)
		}
		var misses []uint64
		for e := max(oldest, v.Activation); e <= latest && e < current; e++ {
			if get(e) == Missed {
				misses = append(misses, e)
			}
		}
		var start uint64
		found := false
		if required == 0 && current > 0 && current-1 >= max(oldest, v.Activation) {
			start = current - 1
			found = true
		}
		if required > 0 && uint64(len(misses)) >= required {
			for last := len(misses) - 1; last >= int(required)-1; last-- {
				first := last - int(required) + 1
				if misses[last]-misses[first] < p.Period {
					start = misses[first]
					found = true
					break
				}
			}
		}
		if !found {
			continue
		}
		if start > math.MaxUint64-p.Period-p.Buffer {
			return nil, fmt.Errorf("challenge expiry overflow")
		}
		candidate := Candidate{Validator: v, Start: start, Expiry: start + p.Period + p.Buffer, Required: required}
		if candidate.Expiry <= current {
			continue
		}
		if bitmap {
			words := p.Period / 256
			if p.Period%256 != 0 {
				words++
			}
			// Refuse an impractically large export rather than allocating
			// unbounded memory following an erroneous governance setting.
			if words > 1<<20 {
				return nil, fmt.Errorf("challenge bitmap exceeds export size limit")
			}
			candidate.Participation = make([]*big.Int, words)
			for i := range candidate.Participation {
				candidate.Participation[i] = new(big.Int)
			}
		}
		for e := start; e < min(start+p.Period, latest+1); e++ {
			o := get(e)
			if o == Timely || o == Missed {
				candidate.Observed++
			}
			if o == Missed {
				candidate.Misses++
				if bitmap {
					offset := e - start
					candidate.Participation[offset/256].SetBit(candidate.Participation[offset/256], int(offset%256), 1)
				}
			}
		}
		candidate.Unknown = p.Period - candidate.Observed
		out = append(out, candidate)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out, nil
}
