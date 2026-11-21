package performancewatcher

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type retry struct {
	attempts uint
	after    time.Time
}

// Historical committee reads can block a beacon client's event loop while it
// replays states. Concurrency limits alone do not leave time for validator work.
func backfillCooldown(elapsed time.Duration) time.Duration {
	return max(30*time.Second, elapsed*9)
}

func nextCollectionEpoch(latest, oldest uint64, count int, coverage map[uint64]int, retries map[uint64]retry, backfillAfter, now time.Time) (uint64, bool) {
	if latest < oldest {
		return 0, false
	}
	for epoch := latest; ; epoch-- {
		// Newly finalized work bypasses the backfill cooldown.
		if epoch != latest && now.Before(backfillAfter) {
			return 0, false
		}
		if coverage[epoch] != count || coverage[epoch] == 0 {
			if !now.Before(retries[epoch].after) {
				return epoch, true
			}
		}
		if epoch == oldest {
			return 0, false
		}
	}
}

type registryResult struct {
	validators []Validator
	err        error
}
type Worker struct {
	mu         sync.RWMutex
	status     Status
	store      *Store
	validators []Validator
	candidates []Candidate
	beacon     *Beacon
	source     ExecutionSource
	days       uint64
	backfill   bool
	path       string
	log        func(string, ...any)
	lastLog    string
	export     chan struct{}
}

func New(path string, days uint64, backfill bool, b *Beacon, source ExecutionSource, log func(string, ...any)) *Worker {
	return &Worker{status: Status{Enabled: true, BackfillEnabled: backfill, Health: "starting", RetentionSource: "bootstrap"}, path: path, days: days, backfill: backfill, beacon: b, source: source, log: log, export: make(chan struct{}, 1)}
}
func DisabledStatus(enabled bool, reason string) Status {
	return Status{Enabled: enabled, Health: reason}
}
func (w *Worker) Status() Status {
	w.mu.RLock()
	defer w.mu.RUnlock()
	s := w.status
	if s.Parameters.Threshold != nil {
		s.Parameters.Threshold = new(big.Int).Set(s.Parameters.Threshold)
	}
	if s.ActiveEpoch != nil {
		epoch := *s.ActiveEpoch
		s.ActiveEpoch = &epoch
	}
	return s
}
func (w *Worker) report(err error) {
	w.mu.Lock()
	w.status.Health = "degraded"
	w.status.LastError = err.Error()
	w.mu.Unlock()
	if w.lastLog != err.Error() {
		w.log("Performance watcher: %s", err)
		w.lastLog = err.Error()
	}
}
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		err := w.beacon.Init(ctx)
		if err == nil {
			break
		}
		w.report(err)
		if !sleep(ctx, 30*time.Second) {
			return
		}
	}
	store, err := Open(filepath.Join(w.path, w.beacon.GenesisRoot))
	if err != nil {
		w.report(err)
		return
	}
	w.mu.Lock()
	w.store = store
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		w.store = nil
		_ = store.Close()
		w.status.ParametersVerified = false
		if ctx.Err() != nil {
			w.status.Health = "stopped"
		}
		w.mu.Unlock()
	}()
	validators, err := store.Registry()
	if err != nil {
		w.report(err)
		return
	}
	var params Parameters
	if raw, e := store.db.Get([]byte("parameters"), nil); e == nil {
		if err = json.Unmarshal(raw, &params); err != nil {
			w.report(err)
			return
		}
	} else if e != leveldb.ErrNotFound {
		w.report(e)
		return
	}
	coverage := map[uint64]int{}
	if err = store.Scan(0, func(r EpochRecord) error { coverage[r.Epoch] = r.Count; return nil }); err != nil {
		w.report(err)
		return
	}
	retries := map[uint64]retry{}
	var cp Checkpoint
	if raw, e := store.db.Get([]byte("checkpoint"), nil); e == nil {
		if err = json.Unmarshal(raw, &cp); err != nil {
			w.report(err)
			return
		}
	} else if e != leveldb.ErrNotFound {
		w.report(e)
		return
	}
	registryResults := make(chan registryResult, 1)
	registryRunning := false
	registryCtx, cancelRegistry := context.WithCancel(ctx)
	var registryWG sync.WaitGroup
	defer func() { cancelRegistry(); registryWG.Wait() }()
	compactResults := make(chan error, 1)
	var compactWG sync.WaitGroup
	defer compactWG.Wait()
	compacting := false
	var refreshed, published, registryAt, compacted, backfillAfter time.Time
	var oldest, latest, current uint64
	verified, registryReady := false, false
	dirty := true
	var candidateDigest [32]byte
	for ctx.Err() == nil {
		now := time.Now()
		select {
		case e := <-compactResults:
			compacting = false
			if e != nil {
				w.report(e)
			}
		default:
		}
		select {
		case result := <-registryResults:
			dirty = true
			published = time.Time{}
			registryRunning = false
			if result.err != nil {
				w.report(result.err)
				registryAt = now.Add(-time.Hour + time.Minute)
			} else {
				validators, err = store.UpdateRegistry(result.validators)
				if err != nil {
					w.report(err)
					return
				}
				if err = store.Prune(oldest, validators); err != nil {
					w.report(err)
					return
				}
				// Re-evaluate explicit unknowns when an existing identity becomes
				// eligible again; row length alone only detects appended identities.
				err = store.Scan(oldest, func(r EpochRecord) error {
					coverage[r.Epoch] = r.Count
					for i, v := range validators {
						if v.Eligible && r.Epoch >= v.Activation && r.Epoch < v.Exit && r.Get(i) == Unknown {
							coverage[r.Epoch] = -1
							break
						}
					}
					return nil
				})
				if err != nil {
					w.report(err)
					return
				}
				registryReady = true
				registryAt = now
				w.mu.Lock()
				w.validators = validators
				w.status.Validators = len(validators)
				w.mu.Unlock()
			}
		default:
		}
		if now.Sub(refreshed) >= 30*time.Second {
			refreshed = now
			verified = false
			p, e := w.source.Snapshot(ctx, true)
			policyChanged := e == nil && !sameParameters(params, p)
			if e == nil {
				if policyChanged {
					dirty = true
					published = time.Time{}
				}
				if params.Deployed && !p.Deployed {
					w.report(fmt.Errorf("finalized Saturn 2 deployment disappeared"))
					return
				}
				params = p
				verified = true
				raw, _ := json.Marshal(params)
				e = store.db.Put([]byte("parameters"), raw, &opt.WriteOptions{Sync: true})
				if e != nil {
					w.report(e)
					return
				}
			}
			if e != nil {
				dirty = true
				w.report(e)
			}
			paused := errors.Is(e, ErrSyncing)
			previousEpoch := current
			current = 0
			if uint64(now.Unix()) >= w.beacon.GenesisTime {
				current = (uint64(now.Unix()) - w.beacon.GenesisTime) / (w.beacon.SecondsPerSlot * w.beacon.SlotsPerEpoch)
			}
			if current != previousEpoch {
				dirty = true
			}
			var length uint64
			oldest, length, err = Retention(params, w.days, w.beacon.SecondsPerSlot*w.beacon.SlotsPerEpoch, current)
			if err != nil {
				w.report(err)
				if !sleep(ctx, 30*time.Second) {
					return
				}
				continue
			}
			// Age pruning continues during finality/parameter RPC outages.
			if err = store.PruneBefore(oldest); err != nil {
				w.report(err)
				return
			}
			if !compacting && now.Sub(compacted) >= time.Hour {
				compacting = true
				compacted = now
				compactWG.Add(1)
				go func(before uint64) { defer compactWG.Done(); compactResults <- store.Compact(before) }(oldest)
			}
			for epoch := range coverage {
				if epoch < oldest {
					delete(coverage, epoch)
					delete(retries, epoch)
				}
			}
			for epoch := range retries {
				if epoch < oldest {
					delete(retries, epoch)
				}
			}
			w.mu.Lock()
			w.status.Parameters = params
			w.status.ParametersVerified = verified
			w.status.CurrentEpoch = current
			w.status.OldestEpoch = oldest
			w.status.RetentionEpochs = length
			w.status.RetentionSource = "bootstrap"
			if params.Deployed {
				w.status.RetentionSource = "on-chain"
			}
			w.status.UpdatedAt = now
			w.status.RPCFailures = w.beacon.Failures.Load()
			if !verified || policyChanged {
				w.candidates = nil
				w.status.CandidateCount = 0
			}
			w.mu.Unlock()
			if paused {
				if !sleep(ctx, 30*time.Second) {
					return
				}
				continue
			}
			next, e := w.beacon.Checkpoint(ctx)
			if e != nil {
				w.report(e)
				if !sleep(ctx, 30*time.Second) {
					return
				}
				continue
			}
			if cp.Root != "" && (next.Epoch < cp.Epoch || (next.Epoch == cp.Epoch && next.Root != cp.Root)) {
				w.report(fmt.Errorf("conflicting finalized checkpoints"))
				return
			}
			cp = next
			raw, _ := json.Marshal(cp)
			if err = store.db.Put([]byte("checkpoint"), raw, &opt.WriteOptions{Sync: true}); err != nil {
				w.report(err)
				return
			}
			if cp.Epoch < 2 {
				if !sleep(ctx, 30*time.Second) {
					return
				}
				continue
			}
			latest = cp.Epoch - 2
			if verified && !registryRunning && now.Sub(registryAt) >= time.Hour {
				registryRunning = true
				registryWG.Add(1)
				go func(p Parameters, checkpoint Checkpoint, known []Validator) {
					defer registryWG.Done()
					ctx, cancel := context.WithTimeout(registryCtx, 10*time.Minute)
					defer cancel()
					v, e := w.source.Registry(ctx, p)
					requested := make(map[string]bool, len(v))
					for _, item := range v {
						requested[item.Pubkey] = true
					}
					// This reader has independent caches and shares the global RPC budget.
					b := NewBeacon(w.beacon.clients)
					b.GenesisRoot = w.beacon.GenesisRoot
					if e == nil {
						v, e = b.ResolveValidators(ctx, checkpoint, v)
					}
					if e == nil {
						resolved := make(map[string]bool, len(v))
						for _, item := range v {
							resolved[item.Pubkey] = true
						}
						for _, item := range known {
							if requested[item.Pubkey] && !resolved[item.Pubkey] {
								e = fmt.Errorf("previously known validator %d missing from finalized registry response", item.Index)
								break
							}
						}
					}
					registryResults <- registryResult{v, e}
				}(params, cp, validators)
			}
			w.mu.Lock()
			w.status.FinalizedEpoch = cp.Epoch
			w.status.LatestMeasurableEpoch = latest
			w.status.Validators = len(validators)
			w.mu.Unlock()
		}
		if !registryReady || len(validators) == 0 {
			if !sleep(ctx, time.Second) {
				return
			}
			continue
		}
		collectionStart := oldest
		firstActivation := uint64(math.MaxUint64)
		for _, v := range validators {
			if v.Eligible && v.Activation < firstActivation {
				firstActivation = v.Activation
			}
		}
		collectionStart = max(collectionStart, firstActivation)
		// Live-only mode never schedules older gaps, including on restart or
		// when finality jumps. Keep retained evidence independent of this bound.
		jobStart := collectionStart
		if !w.backfill {
			jobStart = max(jobStart, latest)
			for epoch := range retries {
				if epoch < latest {
					delete(retries, epoch)
				}
			}
		}
		if collectionStart > oldest {
			if err = store.PruneBefore(collectionStart); err != nil {
				w.report(err)
				return
			}
		}
		if dirty && time.Since(published) >= 30*time.Second {
			var candidates []Candidate
			if verified {
				candidates, err = store.Candidates(ctx, validators, params, current, oldest, latest, nil, false)
				if err != nil {
					w.report(err)
					return
				}
			}
			complete := uint64(0)
			var last uint64
			for e, n := range coverage {
				if e >= collectionStart && e <= latest && n == len(validators) {
					complete++
					if e > last {
						last = e
					}
				}
			}
			pending := uint64(0)
			if w.backfill && latest >= collectionStart {
				pending = latest - collectionStart + 1 - complete
			} else if latest >= jobStart && coverage[latest] != len(validators) {
				pending = 1
			}
			w.mu.Lock()
			w.candidates = candidates
			w.status.CandidateCount = len(candidates)
			w.status.CompleteEpochs = complete
			w.status.PendingEpochs = pending
			w.status.LatestCollectedEpoch = last
			w.status.StorageBytes = store.Size()
			if verified && len(retries) == 0 && time.Since(registryAt) < time.Hour {
				w.status.Health = "collecting"
				w.status.LastError = ""
				w.lastLog = ""
			}
			w.mu.Unlock()
			hash := sha256.New()
			var tuple [24]byte
			for _, candidate := range candidates {
				binary.BigEndian.PutUint64(tuple[:8], candidate.Index)
				binary.BigEndian.PutUint64(tuple[8:16], candidate.Start)
				binary.BigEndian.PutUint64(tuple[16:], candidate.Misses)
				_, _ = hash.Write(tuple[:])
			}
			var digest [32]byte
			copy(digest[:], hash.Sum(nil))
			if digest != candidateDigest {
				w.log("Performance watcher: %d candidates", len(candidates))
				candidateDigest = digest
			}
			published = time.Now()
			dirty = false
		}
		// Descending selection always gives the newest live epoch priority.
		job, found := nextCollectionEpoch(latest, jobStart, len(validators), coverage, retries, backfillAfter, time.Now())
		if !found {
			w.mu.Lock()
			w.status.Mode = "idle"
			w.status.ActiveEpoch = nil
			w.mu.Unlock()
			if !sleep(ctx, time.Second) {
				return
			}
			continue
		}
		// Do not let a long previous job retain newly expired observations.
		w.mu.Lock()
		w.status.Mode = "backfill"
		if job == latest {
			w.status.Mode = "live"
		}
		w.status.ActiveEpoch = &job
		w.mu.Unlock()
		jobCurrent := w.currentEpoch()
		jobOldest, _, e := Retention(params, w.days, w.beacon.SecondsPerSlot*w.beacon.SlotsPerEpoch, jobCurrent)
		if e != nil {
			w.report(e)
			return
		}
		if job < jobOldest {
			refreshed = time.Time{}
			continue
		}
		jobCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		began := time.Now()
		record, e := w.beacon.Collect(jobCtx, job, cp, validators)
		cancel()
		elapsed := time.Since(began)
		backfillAfter = time.Now().Add(backfillCooldown(elapsed))
		w.mu.Lock()
		w.status.ProcessingSeconds = elapsed.Seconds()
		w.mu.Unlock()
		if e != nil {
			if errors.Is(e, ErrFinalityConflict) {
				w.report(e)
				return
			}
			r := retries[job]
			r.attempts = min(r.attempts+1, 10)
			r.after = time.Now().Add(time.Duration(1<<r.attempts) * 30 * time.Second)
			retries[job] = r
			w.report(fmt.Errorf("epoch %d remains unknown: %w", job, e))
			continue
		}
		for i, v := range validators {
			if !v.Eligible || job < v.Activation {
				record.Set(i, Unknown)
			}
		}
		finishedOldest, _, e := Retention(params, w.days, w.beacon.SecondsPerSlot*w.beacon.SlotsPerEpoch, w.currentEpoch())
		if e != nil {
			w.report(e)
			return
		}
		if job < finishedOldest {
			refreshed = time.Time{}
			continue
		}
		if e = store.Put(record); e != nil {
			w.report(e)
			return
		}
		coverage[job] = record.Count
		dirty = true
		delete(retries, job)
	}
}

func (w *Worker) List(offset, limit int) ([]Candidate, int, string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	reason := ""
	if !w.status.Parameters.Deployed {
		reason = "Performance challenges are unavailable before Saturn 2."
	} else if !w.status.ParametersVerified {
		reason = "Current challenge parameters are unavailable."
	} else if !w.status.Parameters.ExitsEnabled {
		reason = "Performance exits are paused."
	}
	if reason != "" {
		return []Candidate{}, 0, reason
	}
	valid := make([]Candidate, 0, len(w.candidates))
	now := uint64(time.Now().Unix())
	current := w.status.CurrentEpoch
	if now >= w.beacon.GenesisTime {
		current = (now - w.beacon.GenesisTime) / (w.beacon.SecondsPerSlot * w.beacon.SlotsPerEpoch)
	}
	for _, c := range w.candidates {
		if c.Expiry > current {
			valid = append(valid, c)
		}
	}
	count := len(valid)
	offset = min(max(offset, 0), count)
	limit = min(max(limit, 1), 100)
	return valid[offset:min(offset+limit, count)], count, ""
}

func (w *Worker) Candidate(ctx context.Context, index uint64) (*Candidate, error) {
	w.mu.RLock()
	available := w.store != nil && w.status.ParametersVerified
	w.mu.RUnlock()
	if !available {
		return nil, fmt.Errorf("watcher is unavailable")
	}
	select {
	case w.export <- struct{}{}:
		defer func() { <-w.export }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	params, err := w.source.Snapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	if !params.Deployed {
		return nil, fmt.Errorf("performance challenges are unavailable before Saturn 2")
	}
	if !params.ExitsEnabled {
		return nil, fmt.Errorf("performance exits are paused")
	}
	current := w.currentEpoch()
	oldest, _, err := Retention(params, w.days, w.beacon.SecondsPerSlot*w.beacon.SlotsPerEpoch, current)
	if err != nil {
		return nil, err
	}
	results, err := func() ([]Candidate, error) {
		w.mu.RLock()
		defer w.mu.RUnlock()
		if w.store == nil || !w.status.ParametersVerified {
			return nil, fmt.Errorf("watcher is unavailable")
		}
		return w.store.Candidates(ctx, w.validators, params, current, oldest, w.status.LatestMeasurableEpoch, &index, true)
	}()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("validator %d has no supported challenge window", index)
	}
	ok, err := w.source.Eligible(ctx, params, results[0].Validator)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("validator is no longer eligible")
	}
	if results[0].Expiry <= w.currentEpoch() {
		return nil, fmt.Errorf("challenge window expired during revalidation")
	}
	return &results[0], nil
}

func (w *Worker) currentEpoch() uint64 {
	now := uint64(time.Now().Unix())
	if now < w.beacon.GenesisTime {
		return 0
	}
	return (now - w.beacon.GenesisTime) / (w.beacon.SecondsPerSlot * w.beacon.SlotsPerEpoch)
}
func sameParameters(a, b Parameters) bool {
	if a.Deployed != b.Deployed || a.ExitsEnabled != b.ExitsEnabled || a.Period != b.Period || a.Buffer != b.Buffer {
		return false
	}
	if a.Threshold == nil || b.Threshold == nil {
		return a.Threshold == nil && b.Threshold == nil
	}
	return a.Threshold.Cmp(b.Threshold) == 0
}
