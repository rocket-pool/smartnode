package performancewatcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

func testRoot(n uint64) string { return fmt.Sprintf("0x%064x", n) }
func testParams(period, buffer uint64, threshold int64) Parameters {
	return Parameters{Deployed: true, ExitsEnabled: true, Period: period, Buffer: buffer, Threshold: big.NewInt(threshold)}
}
func testValidator(index uint64) Validator {
	return Validator{Index: index, Pubkey: fmt.Sprintf("0x%096x", index+1), Pool: testRoot(index), Kind: "minipool", Eligible: true, Exit: math.MaxUint64}
}
func testStore(t testing.TB) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}
func putObservations(t testing.TB, s *Store, epoch uint64, observations ...Observation) {
	t.Helper()
	r := NewRecord(epoch, len(observations))
	r.TargetRoot = testRoot(epoch + 1)
	r.FinalizedRoot = testRoot(999)
	for i, o := range observations {
		r.Set(i, o)
	}
	if err := s.Put(r); err != nil {
		t.Fatal(err)
	}
}

func TestRetentionAndThreshold(t *testing.T) {
	for _, tc := range []struct {
		name                                   string
		p                                      Parameters
		days, seconds, current, oldest, length uint64
		invalid                                bool
	}{
		{"bootstrap", Parameters{}, 201, 384, 50000, 4776, 45225, false},
		{"round up", Parameters{}, 1, 1000, 100, 14, 87, false},
		{"chain", testParams(44032, 225, 940000000000000000), 201, 384, 50000, 5744, 44257, false},
		{"genesis", testParams(10, 2, 940000000000000000), 201, 384, 1, 0, 12, false},
		{"zero days", Parameters{}, 0, 384, 1, 0, 0, true},
		{"overflow", testParams(math.MaxUint64, 1, 0), 201, 384, 1, 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lo, n, err := Retention(tc.p, tc.days, tc.seconds, tc.current)
			if (err != nil) != tc.invalid || !tc.invalid && (lo != tc.oldest || n != tc.length) {
				t.Fatalf("got %d %d %v", lo, n, err)
			}
		})
	}
	p := testParams(44032, 225, 940000000000000000)
	n, err := p.RequiredMisses()
	if err != nil || n != 2642 {
		t.Fatalf("misses %d %v", n, err)
	}
	p.Threshold = big.NewInt(1000000000000000001)
	if _, err = p.RequiredMisses(); err == nil {
		t.Fatal("accepted invalid threshold")
	}
}

func TestCandidatesOlderWindowPartialCoverageAndExpiry(t *testing.T) {
	s := testStore(t)
	validators := []Validator{testValidator(10), testValidator(11)}
	if _, err := s.UpdateRegistry(validators); err != nil {
		t.Fatal(err)
	}
	// The newest two misses (4, 20) do not fit. The older pair (3, 4)
	// still qualifies through the proof buffer; unknowns supply no misses.
	putObservations(t, s, 3, Missed, Timely)
	putObservations(t, s, 4, Missed, Unknown)
	putObservations(t, s, 20, Missed, Missed)
	p := testParams(10, 10, 800000000000000000)
	c, err := s.Candidates(context.Background(), validators, p, 21, 2, 20, nil, true)
	if err != nil || len(c) != 1 {
		t.Fatalf("%v %v", c, err)
	}
	if c[0].Start != 3 || c[0].Expiry != 23 || c[0].Observed != 2 || c[0].Unknown != 8 || c[0].Participation[0].Uint64() != 3 {
		t.Fatalf("bad candidate %+v", c[0])
	}
	c, err = s.Candidates(context.Background(), validators, p, 23, 4, 22, nil, false)
	if err != nil || len(c) != 0 {
		t.Fatalf("expired candidate %v %v", c, err)
	}
	p.ExitsEnabled = false
	c, err = s.Candidates(context.Background(), validators, p, 21, 2, 20, nil, true)
	if err != nil || len(c) != 0 {
		t.Fatal("paused candidates")
	}
	p.Deployed = false
	p.ExitsEnabled = true
	c, err = s.Candidates(context.Background(), validators, p, 21, 2, 20, nil, true)
	if err != nil || len(c) != 0 {
		t.Fatal("predeployment candidates")
	}
}

func TestStoreResumePruneAndDiscovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	v := []Validator{testValidator(10)}
	_, err = s.UpdateRegistry(v)
	if err != nil {
		t.Fatal(err)
	}
	putObservations(t, s, 5, Missed)
	putObservations(t, s, 6, Timely)
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	v, err = s.UpdateRegistry([]Validator{testValidator(11), testValidator(10)})
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 2 || v[0].Index != 10 || v[1].Index != 11 {
		t.Fatal("unstable ordinals")
	}
	if err = s.PruneBefore(6); err != nil {
		t.Fatal(err)
	}
	n := 0
	err = s.Scan(0, func(r EpochRecord) error {
		n++
		if r.Epoch != 6 || r.Get(0) != Timely || r.Get(1) != Unknown {
			t.Fatalf("bad retained record %+v", r)
		}
		return nil
	})
	if err != nil || n != 1 {
		t.Fatal(err, n)
	}
	v[0].Eligible = false
	if err = s.Prune(6, v); err != nil {
		t.Fatal(err)
	}
	err = s.Scan(0, func(r EpochRecord) error {
		if r.Get(0) != Unknown {
			t.Fatal("ineligible history retained")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCorruptionAndSchemaErrors(t *testing.T) {
	s := testStore(t)
	if err := s.db.Put(epochKey(10), []byte{0}, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Scan(0, func(EpochRecord) error { return nil }); err == nil {
		t.Fatal("corruption silently accepted")
	}
	path := t.TempDir()
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Put([]byte("schema"), []byte("2"), nil)
	_ = db.Close()
	if db, err := Open(path); err == nil {
		_ = db.Close()
		t.Fatal("unsupported schema accepted")
	}
}

type fixtureRequester struct {
	mu        sync.Mutex
	responses map[string][]byte
	requests  []string
}

func (f *fixtureRequester) WatcherRequest(ctx context.Context, path string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.requests = append(f.requests, path)
	if data, ok := f.responses[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("unavailable %s", path)
}

type mockExecutionSource struct {
	params        Parameters
	validators    []Validator
	eligible      bool
	pools         []string
	poolErr       error
	requestedNode string
}

func (s *mockExecutionSource) Snapshot(context.Context, bool) (Parameters, error) {
	return s.params, nil
}
func (s *mockExecutionSource) Registry(context.Context, Parameters) ([]Validator, error) {
	return s.validators, nil
}
func (s *mockExecutionSource) Eligible(context.Context, Parameters, Validator) (bool, error) {
	return s.eligible, nil
}

func (s *mockExecutionSource) NodePools(ctx context.Context, address string) ([]string, error) {
	s.requestedNode = address
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.pools, s.poolErr
}

func TestBeaconInitStructuredSpec(t *testing.T) {
	for _, tc := range []struct {
		name    string
		timing  map[string]any
		invalid bool
	}{
		{"seconds", map[string]any{"SECONDS_PER_SLOT": "12"}, false},
		{"milliseconds", map[string]any{"SLOT_DURATION_MS": "12000"}, false},
		{"fractional seconds", map[string]any{"SLOT_DURATION_MS": "12500"}, true},
		{"missing duration", map[string]any{}, true},
		{"missing epoch length", map[string]any{"SECONDS_PER_SLOT": "12", "SLOTS_PER_EPOCH": nil}, true},
		{"missing fork epoch", map[string]any{"SECONDS_PER_SLOT": "12", "DENEB_FORK_EPOCH": nil}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &fixtureRequester{responses: map[string][]byte{}}
			f.put("/eth/v1/beacon/genesis", map[string]any{"data": map[string]any{"genesis_validators_root": testRoot(999), "genesis_time": "1"}})
			spec := map[string]any{
				"SLOTS_PER_EPOCH": "32", "DENEB_FORK_EPOCH": "0",
				"BLOB_SCHEDULE": []any{map[string]any{"EPOCH": "1", "MAX_BLOBS_PER_BLOCK": "9"}},
			}
			for k, v := range tc.timing {
				spec[k] = v
			}
			f.put("/eth/v1/config/spec", map[string]any{"data": spec})
			b := NewBeacon([]Requester{f})
			err := b.Init(context.Background())
			if (err != nil) != tc.invalid {
				t.Fatalf("Init error: %v", err)
			}
			if err == nil && (b.SlotsPerEpoch != 32 || b.SecondsPerSlot != 12 || b.DenebEpoch != 0) {
				t.Fatalf("unexpected timing: %d, %d, %d", b.SlotsPerEpoch, b.SecondsPerSlot, b.DenebEpoch)
			}
		})
	}
}

func TestCheckpointRejectsStaleHead(t *testing.T) {
	for _, head := range []uint64{97, 98, 100} {
		t.Run(fmt.Sprint(head), func(t *testing.T) {
			b, f, cp, _ := fixture("gloas")
			b.GenesisTime = uint64(time.Now().Unix()) - 1200
			f.put("/eth/v1/node/syncing", map[string]any{"data": map[string]any{"head_slot": fmt.Sprint(head), "is_syncing": false, "is_optimistic": false, "el_offline": false}})
			f.put("/eth/v1/beacon/states/head/finality_checkpoints", map[string]any{"execution_optimistic": false, "data": map[string]any{"finalized": map[string]any{"epoch": "3", "root": cp.Root}}})
			_, err := b.Checkpoint(context.Background())
			if head == 97 {
				if !errors.Is(err, ErrSyncing) {
					t.Fatalf("stale head accepted: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckpointFollowsSyncedClient(t *testing.T) {
	genesis := uint64(time.Now().Unix()) - 1200
	syncing := func(head uint64) map[string]any {
		return map[string]any{"data": map[string]any{"head_slot": fmt.Sprint(head), "is_syncing": false, "is_optimistic": false, "el_offline": false}}
	}
	finality := func(epoch uint64, root string) map[string]any {
		return map[string]any{"execution_optimistic": false, "data": map[string]any{"finalized": map[string]any{"epoch": fmt.Sprint(epoch), "root": root}}}
	}
	onlySync := func(t *testing.T, f *fixtureRequester) {
		t.Helper()
		if len(f.requests) != 1 || f.requests[0] != "/eth/v1/node/syncing" {
			t.Fatalf("client was used after failing sync: %v", f.requests)
		}
	}
	t.Run("lagging primary", func(t *testing.T) {
		b, synced, cp, _ := fixture("gloas")
		b.GenesisTime = genesis
		synced.put("/eth/v1/node/syncing", syncing(1000))
		synced.put("/eth/v1/beacon/states/head/finality_checkpoints", finality(cp.Epoch, cp.Root))
		lagging := &fixtureRequester{responses: map[string][]byte{}}
		lagging.put("/eth/v1/node/syncing", syncing(90))
		lagging.put("/eth/v1/beacon/states/head/finality_checkpoints", finality(1, testRoot(1)))
		lagging.header(1, 0)
		b.clients = []Requester{lagging, synced}
		b.verified = []bool{true, true}
		got, err := b.Checkpoint(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got != cp {
			t.Fatalf("checkpoint %+v", got)
		}
		onlySync(t, lagging)
	})
	t.Run("fallback skips lagging client", func(t *testing.T) {
		b, synced, cp, _ := fixture("gloas")
		b.GenesisTime = genesis
		blind := &fixtureRequester{responses: map[string][]byte{}}
		blind.put("/eth/v1/node/syncing", syncing(1000))
		lagging := &fixtureRequester{responses: map[string][]byte{}}
		lagging.put("/eth/v1/node/syncing", syncing(90))
		lagging.put("/eth/v1/beacon/states/head/finality_checkpoints", finality(1, testRoot(1)))
		lagging.header(1, 0)
		lagging.header(6, 5)
		synced.put("/eth/v1/node/syncing", syncing(1000))
		synced.put("/eth/v1/beacon/states/head/finality_checkpoints", finality(cp.Epoch, cp.Root))
		b.clients = []Requester{blind, lagging, synced}
		b.verified = []bool{true, true, true}
		got, err := b.Checkpoint(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got != cp {
			t.Fatalf("checkpoint %+v", got)
		}
		onlySync(t, lagging)
		if len(blind.requests) != 2 || blind.requests[1] != "/eth/v1/beacon/states/head/finality_checkpoints" {
			t.Fatalf("blind requests %v", blind.requests)
		}
	})
	t.Run("header stays on synced client", func(t *testing.T) {
		b, synced, cp, _ := fixture("gloas")
		b.GenesisTime = genesis
		synced.put("/eth/v1/node/syncing", syncing(1000))
		synced.put("/eth/v1/beacon/states/head/finality_checkpoints", finality(cp.Epoch, cp.Root))
		delete(synced.responses, "/eth/v1/beacon/headers/"+cp.Root)
		lagging := &fixtureRequester{responses: map[string][]byte{}}
		lagging.put("/eth/v1/node/syncing", syncing(90))
		lagging.header(6, 5)
		b.clients = []Requester{synced, lagging}
		b.verified = []bool{true, true}
		if _, err := b.Checkpoint(context.Background()); err == nil {
			t.Fatal("accepted header from a client that failed sync")
		}
		onlySync(t, lagging)
	})
}

func TestBackfillPacingPreservesLivePriority(t *testing.T) {
	now := time.Now()
	coverage := map[uint64]int{100: 3}
	retries := map[uint64]retry{}
	after := now.Add(backfillCooldown(10 * time.Second))
	if backfillCooldown(time.Second) != 30*time.Second || after.Sub(now) != 90*time.Second {
		t.Fatal("backfill must allow at least 30 seconds and nine times the job duration for client recovery")
	}
	if _, found := nextCollectionEpoch(100, 90, 3, coverage, retries, after, now); found {
		t.Fatal("backfill ignored cooldown")
	}
	if epoch, found := nextCollectionEpoch(101, 90, 3, coverage, retries, after, now); !found || epoch != 101 {
		t.Fatal("live collection blocked by backfill cooldown")
	}
	if epoch, found := nextCollectionEpoch(100, 90, 3, coverage, retries, after, after); !found || epoch != 99 {
		t.Fatal("backfill failed to resume newest first")
	}
	retries[101] = retry{after: after}
	if _, found := nextCollectionEpoch(101, 90, 3, coverage, retries, after, now); found {
		t.Fatal("retry or backfill cooldown bypassed")
	}
}

func TestWorkerPreDeploymentRestartAndTransition(t *testing.T) {
	b, f, cp, validators := fixture("electra")
	f.put("/eth/v1/beacon/genesis", map[string]any{"data": map[string]any{"genesis_validators_root": testRoot(999), "genesis_time": fmt.Sprint(time.Now().Unix() - 72)}})
	f.put("/eth/v1/config/spec", map[string]any{"data": map[string]any{"SLOTS_PER_EPOCH": "2", "SECONDS_PER_SLOT": "12", "DENEB_FORK_EPOCH": "0"}})
	f.put("/eth/v1/node/syncing", map[string]any{"data": map[string]any{"head_slot": "6", "is_syncing": false, "is_optimistic": false, "el_offline": false}})
	f.put("/eth/v1/beacon/states/head/finality_checkpoints", map[string]any{"execution_optimistic": false, "data": map[string]any{"finalized": map[string]any{"epoch": "3", "root": cp.Root}}})
	var response []any
	for _, v := range validators {
		response = append(response, map[string]any{"index": fmt.Sprint(v.Index), "validator": map[string]any{"pubkey": v.Pubkey, "activation_epoch": "0", "exit_epoch": fmt.Sprint(uint64(math.MaxUint64))}})
	}
	f.put("/eth/v1/beacon/states/"+cp.StateRoot+"/validators?id="+validators[0].Pubkey+","+validators[1].Pubkey, map[string]any{"finalized": true, "execution_optimistic": false, "data": response})
	source := &mockExecutionSource{validators: validators, eligible: true}
	path := t.TempDir()
	for _, deployed := range []bool{false, true} {
		if deployed {
			source.params = testParams(10, 2, 900000000000000000)
			b = NewBeacon([]Requester{f})
		}
		w := New(path, 201, true, b, source, func(string, ...any) {})
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() { defer close(done); w.Run(ctx) }()
		deadline := time.Now().Add(4 * time.Second)
		collected := false
		for time.Now().Before(deadline) {
			w.mu.RLock()
			if w.store != nil && len(w.validators) == 2 {
				raw, err := w.store.db.Get(epochKey(1), nil)
				if err == nil {
					r, e := decodeRecord(1, raw)
					collected = e == nil && r.Get(0) == Timely && r.Get(1) == Missed
				}
			}
			w.mu.RUnlock()
			if collected {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		status := w.Status()
		cancel()
		<-done
		if !collected {
			t.Fatalf("deployed=%v: no collection: %+v", deployed, status)
		}
		if status.Parameters.Deployed != deployed || status.RetentionSource != map[bool]string{false: "bootstrap", true: "on-chain"}[deployed] {
			t.Fatalf("bad parameter transition %+v", status)
		}
		if !deployed {
			if _, _, reason := w.List(0, 10); !strings.Contains(reason, "before Saturn 2") {
				t.Fatal(reason)
			}
		}
	}
}

func TestWorkerLiveOnlySkipsGapsAndCanEnableBackfill(t *testing.T) {
	path := t.TempDir()
	genesis := time.Now().Unix() - 120
	_, _, _, validators := fixture("electra")
	store, err := Open(filepath.Join(path, testRoot(999)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.UpdateRegistry(validators); err != nil {
		t.Fatal(err)
	}
	putObservations(t, store, 0, Timely, Timely)
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []struct {
		name     string
		latest   uint64
		backfill bool
		want     uint64
	}{
		{"live startup", 1, false, 1},
		{"restart after downtime", 3, false, 3},
		{"restart with a gap and current epoch already collected", 3, false, 3},
		{"enable backfill", 3, true, 2},
	} {
		t.Run(phase.name, func(t *testing.T) {
			b, f, cp, _ := fixture("electra")
			if phase.latest == 3 {
				f.header(10, 9)
				f.header(9, 7)
				f.header(7, 5)
				cp = Checkpoint{Epoch: 5, Slot: 10, Root: testRoot(10), StateRoot: testRoot(110)}
				a := map[string]any{"aggregation_bits": "0x05", "committee_bits": "0x0100000000000000", "data": map[string]any{"slot": "7", "index": "0", "target": map[string]any{"epoch": "3", "root": testRoot(5)}}}
				f.attestations(9, "electra", []any{a})
				f.attestations(7, "electra", []any{})
				for _, epoch := range []uint64{2, 3} {
					slot := epoch*2 + 1
					f.put(fmt.Sprintf("/eth/v1/beacon/states/%s/committees?epoch=%d", testRoot(100+slot), epoch), map[string]any{"finalized": true, "execution_optimistic": false, "data": []any{map[string]any{"index": "0", "slot": fmt.Sprint(slot), "validators": []string{"10", "11"}}}})
				}
			}
			f.put("/eth/v1/beacon/genesis", map[string]any{"data": map[string]any{"genesis_validators_root": testRoot(999), "genesis_time": fmt.Sprint(genesis)}})
			f.put("/eth/v1/config/spec", map[string]any{"data": map[string]any{"SLOTS_PER_EPOCH": "2", "SECONDS_PER_SLOT": "12", "DENEB_FORK_EPOCH": "0"}})
			f.put("/eth/v1/node/syncing", map[string]any{"data": map[string]any{"head_slot": "10", "is_syncing": false, "is_optimistic": false, "el_offline": false}})
			f.put("/eth/v1/beacon/states/head/finality_checkpoints", map[string]any{"execution_optimistic": false, "data": map[string]any{"finalized": map[string]any{"epoch": fmt.Sprint(cp.Epoch), "root": cp.Root}}})
			var resolved []any
			for _, v := range validators {
				resolved = append(resolved, map[string]any{"index": fmt.Sprint(v.Index), "validator": map[string]any{"pubkey": v.Pubkey, "activation_epoch": "0", "exit_epoch": fmt.Sprint(uint64(math.MaxUint64))}})
			}
			f.put("/eth/v1/beacon/states/"+cp.StateRoot+"/validators?id="+validators[0].Pubkey+","+validators[1].Pubkey, map[string]any{"finalized": true, "execution_optimistic": false, "data": resolved})
			source := &mockExecutionSource{validators: validators, eligible: true}
			w := New(path, 201, phase.backfill, b, source, func(string, ...any) {})
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() { defer close(done); w.Run(ctx) }()
			ready := false
			deadline := time.Now().Add(4 * time.Second)
			for time.Now().Before(deadline) {
				w.mu.RLock()
				if w.store != nil && w.status.Mode == "idle" {
					_, err := w.store.db.Get(epochKey(phase.want), nil)
					ready = err == nil
				}
				w.mu.RUnlock()
				if ready {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			status := w.Status()
			cancel()
			<-done
			if !ready || status.BackfillEnabled != phase.backfill {
				t.Fatalf("collection failed: %+v", status)
			}
			if !phase.backfill {
				for _, request := range f.requests {
					if strings.Contains(request, "/committees?epoch=") && !strings.HasSuffix(request, fmt.Sprintf("epoch=%d", phase.latest)) {
						t.Fatalf("live-only mode fetched historical committees: %s", request)
					}
				}
				if strings.Contains(phase.name, "already collected") && status.PendingEpochs != 0 {
					t.Fatalf("unscheduled historical gaps reported as pending: %+v", status)
				}
			}
			s, err := Open(filepath.Join(path, testRoot(999)))
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			if _, err := s.db.Get(epochKey(0), nil); err != nil {
				t.Fatal("existing history removed", err)
			}
			_, gapErr := s.db.Get(epochKey(2), nil)
			if !phase.backfill && gapErr != leveldb.ErrNotFound || phase.backfill && gapErr != nil {
				t.Fatalf("unexpected gap coverage: %v", gapErr)
			}
		})
	}
}

func TestCandidateExportRevalidatesEligibilityAndParameters(t *testing.T) {
	s := testStore(t)
	validators := []Validator{testValidator(1)}
	_, _ = s.UpdateRegistry(validators)
	putObservations(t, s, 3, Missed)
	putObservations(t, s, 4, Missed)
	source := &mockExecutionSource{params: testParams(10, 10, 800000000000000000), eligible: true}
	b := NewBeacon(nil)
	b.GenesisTime = uint64(time.Now().Unix()) - 21
	b.SlotsPerEpoch = 1
	b.SecondsPerSlot = 1
	w := New("", 201, false, b, source, func(string, ...any) {})
	w.store = s
	w.validators = validators
	w.status.ParametersVerified = true
	w.status.LatestMeasurableEpoch = 19
	candidate, err := w.Candidate(context.Background(), 1)
	if err != nil || candidate.Start != 3 {
		t.Fatal(candidate, err)
	}
	source.eligible = false
	if _, err = w.Candidate(context.Background(), 1); err == nil {
		t.Fatal("exported ineligible validator")
	}
	source.eligible = true
	source.params.Deployed = false
	if _, err = w.Candidate(context.Background(), 1); err == nil || !strings.Contains(err.Error(), "before Saturn 2") {
		t.Fatalf("expected pre-deployment explanation, got %v", err)
	}
	source.params.Deployed = true
	source.params.ExitsEnabled = false
	if _, err = w.Candidate(context.Background(), 1); err == nil || !strings.Contains(err.Error(), "performance exits are paused") {
		t.Fatalf("expected paused-challenges explanation, got %v", err)
	}
	source.params.ExitsEnabled = true
	source.params.Buffer = 1
	if _, err = w.Candidate(context.Background(), 1); err == nil {
		t.Fatal("exported expired window under new parameters")
	}
}
func (f *fixtureRequester) put(path string, value any) {
	data, _ := json.Marshal(value)
	f.responses[path] = data
}
func (f *fixtureRequester) header(slot, parent uint64) {
	r := map[string]any{"finalized": true, "execution_optimistic": false, "data": map[string]any{"root": testRoot(slot), "canonical": true, "header": map[string]any{"message": map[string]any{"slot": fmt.Sprint(slot), "parent_root": testRoot(parent), "state_root": testRoot(100 + slot)}}}}
	f.put("/eth/v1/beacon/headers/"+testRoot(slot), r)
	f.put(fmt.Sprintf("/eth/v1/beacon/headers/%d", slot), r)
}
func (f *fixtureRequester) attestations(slot uint64, version string, attestations []any) {
	f.put("/eth/v2/beacon/blocks/"+testRoot(slot)+"/attestations", map[string]any{"version": version, "finalized": true, "execution_optimistic": false, "data": attestations})
}
func fixture(version string) (*Beacon, *fixtureRequester, Checkpoint, []Validator) {
	f := &fixtureRequester{responses: map[string][]byte{}}
	f.header(6, 5)
	f.header(5, 3)
	f.header(3, 1)
	f.header(1, 0)
	a := map[string]any{"aggregation_bits": "0x05", "data": map[string]any{"slot": "3", "index": "0", "target": map[string]any{"epoch": "1", "root": testRoot(1)}}}
	if version == "electra" || version == "fulu" || version == "gloas" {
		a["committee_bits"] = "0x0100000000000000"
	}
	f.attestations(5, version, []any{a, a})
	f.attestations(3, version, []any{})
	f.put("/eth/v1/beacon/states/"+testRoot(103)+"/committees?epoch=1", map[string]any{"finalized": true, "execution_optimistic": false, "data": []any{map[string]any{"index": "0", "slot": "3", "validators": []string{"10", "11"}}}})
	b := NewBeacon([]Requester{f})
	b.SlotsPerEpoch = 2
	b.SecondsPerSlot = 12
	b.GenesisRoot = testRoot(999)
	b.verified[0] = true
	return b, f, Checkpoint{Epoch: 3, Slot: 6, Root: testRoot(6), StateRoot: testRoot(106)}, []Validator{testValidator(10), testValidator(11)}
}

func TestFinalizedCollection(t *testing.T) {
	for _, version := range []string{"deneb", "electra", "fulu", "gloas"} {
		t.Run(version, func(t *testing.T) {
			b, f, cp, v := fixture(version)
			r, err := b.Collect(context.Background(), 1, cp, v)
			if err != nil {
				t.Fatal(err)
			}
			// Equivalent finalized participation flags are [TIMELY_TARGET, 0].
			flags := []byte{2, 0}
			for i, flag := range flags {
				want := Missed
				if flag&2 != 0 {
					want = Timely
				}
				if r.Get(i) != want {
					t.Fatalf("validator %d: %v", i, r.Get(i))
				}
			}
			if r.TargetRoot != testRoot(1) {
				t.Fatal("skipped boundary target")
			}
			requests := len(f.requests)
			if _, err = b.Collect(context.Background(), 1, cp, v); err != nil {
				t.Fatal(err)
			}
			if len(f.requests) != requests+1 {
				t.Fatal("raw block data not reused")
			}
			for _, path := range f.requests {
				if strings.Contains(path, "/eth/v2/beacon/blocks/") && !strings.HasSuffix(path, "/attestations") {
					t.Fatal("full block requested")
				}
				if strings.Contains(path, "states/head/") {
					t.Fatal("unfinalized state requested")
				}
			}
		})
	}
}

func TestGloasPayloadIndex(t *testing.T) {
	for _, version := range []string{"electra", "fulu", "gloas"} {
		for _, index := range []string{"0", "1", "2"} {
			t.Run(version+"/"+index, func(t *testing.T) {
				b, f, cp, v := fixture(version)
				path := "/eth/v2/beacon/blocks/" + testRoot(5) + "/attestations"
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), `"index":"0"`, `"index":"`+index+`"`))
				r, err := b.Collect(context.Background(), 1, cp, v)
				valid := index == "0" || version == "gloas" && index == "1"
				if (err == nil) != valid {
					t.Fatalf("index %s accepted=%v: %v", index, err == nil, err)
				}
				if valid && (r.Get(0) != Timely || r.Get(1) != Missed) {
					t.Fatalf("payload index changed timely target: %v", r.Bits)
				}
			})
		}
	}
}

func TestUnknownHistoryNeverBecomesMisses(t *testing.T) {
	for _, mode := range []string{"missing parent", "missing attestations", "missing committees", "optimistic", "bad bitlist", "unfinalized window", "wrong root", "missing header slot", "missing attestation epoch", "missing committee slot"} {
		t.Run(mode, func(t *testing.T) {
			b, f, cp, v := fixture("deneb")
			switch mode {
			case "missing parent":
				delete(f.responses, "/eth/v1/beacon/headers/"+testRoot(3))
			case "missing attestations":
				delete(f.responses, "/eth/v2/beacon/blocks/"+testRoot(5)+"/attestations")
			case "missing committees":
				delete(f.responses, "/eth/v1/beacon/states/"+testRoot(103)+"/committees?epoch=1")
			case "optimistic":
				path := "/eth/v2/beacon/blocks/" + testRoot(5) + "/attestations"
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), `"execution_optimistic":false`, `"execution_optimistic":true`))
			case "bad bitlist":
				path := "/eth/v2/beacon/blocks/" + testRoot(5) + "/attestations"
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), "0x05", "0x01"))
			case "unfinalized window":
				cp.Epoch = 2
			case "wrong root":
				path := "/eth/v1/beacon/headers/" + testRoot(3)
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), `"root":"`+testRoot(3), `"root":"`+testRoot(9)))
			case "missing header slot":
				path := "/eth/v1/beacon/headers/" + testRoot(3)
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), `"slot":"3",`, ""))
			case "missing attestation epoch":
				path := "/eth/v2/beacon/blocks/" + testRoot(5) + "/attestations"
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), `"epoch":"1",`, ""))
			case "missing committee slot":
				path := "/eth/v1/beacon/states/" + testRoot(103) + "/committees?epoch=1"
				f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), `"slot":"3",`, ""))
			}
			if _, err := b.Collect(context.Background(), 1, cp, v); err == nil {
				t.Fatal("incomplete/invalid history accepted")
			}
		})
	}
}

func TestEntireSkippedEpochAndWrongTargets(t *testing.T) {
	b, f, cp, v := fixture("deneb")
	f.header(6, 1)
	f.responses["/eth/v1/beacon/states/2/committees?epoch=1"] = f.responses["/eth/v1/beacon/states/"+testRoot(103)+"/committees?epoch=1"]
	r, err := b.Collect(context.Background(), 1, cp, v)
	if err != nil {
		t.Fatal(err)
	}
	if r.Get(0) != Missed || r.Get(1) != Missed {
		t.Fatal("skipped inclusion window")
	}
	b, f, cp, v = fixture("deneb")
	path := "/eth/v2/beacon/blocks/" + testRoot(5) + "/attestations"
	f.responses[path] = []byte(strings.ReplaceAll(string(f.responses[path]), testRoot(1), testRoot(99)))
	r, err = b.Collect(context.Background(), 1, cp, v)
	if err != nil || r.Get(0) != Missed {
		t.Fatal("wrong target counted", err)
	}
}

func TestFallbackAndCancellation(t *testing.T) {
	b, good, cp, v := fixture("electra")
	bad := &fixtureRequester{responses: map[string][]byte{}}
	b.clients = []Requester{bad, good}
	b.verified = []bool{true, true}
	if _, err := b.Collect(context.Background(), 1, cp, v); err != nil {
		t.Fatal(err)
	}
	if len(bad.requests) == 0 || b.Failures.Load() == 0 {
		t.Fatal("fallback not attempted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := b.Collect(ctx, 1, cp, v); err == nil {
		t.Fatal("cancellation ignored")
	}
}

func TestActivationAndNewValidatorCoverage(t *testing.T) {
	b, _, cp, v := fixture("deneb")
	v[1].Activation = 2
	r, err := b.Collect(context.Background(), 1, cp, v)
	if err != nil || r.Get(1) != Inactive {
		t.Fatal(r, err)
	}
	s := testStore(t)
	_, _ = s.UpdateRegistry(v)
	if err = s.Put(r); err != nil {
		t.Fatal(err)
	}
	v = append(v, testValidator(12))
	_, err = s.UpdateRegistry(v)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Scan(0, func(saved EpochRecord) error {
		if saved.Get(2) != Unknown {
			t.Fatal("new validator inherited old coverage")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestThresholdZeroAndRegistryIdentity(t *testing.T) {
	s := testStore(t)
	v := []Validator{testValidator(5)}
	_, _ = s.UpdateRegistry(v)
	p := testParams(10, 1, 1000000000000000000)
	c, err := s.Candidates(context.Background(), v, p, 5, 0, 3, nil, true)
	if err != nil || len(c) != 1 || c[0].Start != 4 || !reflect.DeepEqual(c[0].Participation, []*big.Int{new(big.Int)}) {
		t.Fatal(c, err)
	}
	v[0].Pubkey = "different"
	if _, err = s.UpdateRegistry(v); err == nil {
		t.Fatal("identity mutation accepted")
	}
}

type blockingRequester struct{}

func (blockingRequester) WatcherRequest(ctx context.Context, _ string) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
func TestWorkerCancellationBeforeInitialization(t *testing.T) {
	w := New(t.TempDir(), 201, false, NewBeacon([]Requester{blockingRequester{}}), nil, func(string, ...any) {})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); w.Run(ctx) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not cancel")
	}
}
