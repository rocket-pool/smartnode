package performancewatcher

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// Run with -run '^$' -bench 'Benchmark.*20000' -benchtime=1x -benchmem.
// Both benchmarks are synthetic local workloads; neither measures archive RPC
// latency nor claims to emulate all Ethereum committees or network bandwidth.
func BenchmarkLiveCollection20000(b *testing.B) {
	const count = 20000
	for i := 0; i < b.N; i++ {
		collector, f, cp, _ := fixture("electra")
		validators := make([]Validator, count)
		var committees []any
		for j := range validators {
			validators[j] = testValidator(uint64(j))
			validators[j].Exit = math.MaxUint64
		}
		for start, index := 0, 0; start < count; start, index = start+313, index+1 {
			members := []string{}
			for j := start; j < min(start+313, count); j++ {
				members = append(members, fmt.Sprint(j))
			}
			committees = append(committees, map[string]any{"slot": "3", "index": fmt.Sprint(index), "validators": members})
		}
		f.put("/eth/v1/beacon/states/"+testRoot(103)+"/committees?epoch=1", map[string]any{"finalized": true, "execution_optimistic": false, "data": committees})
		bits := make([]byte, count/8+1)
		for j := range bits {
			bits[j] = 255
		}
		bits[len(bits)-1] = 1
		a := map[string]any{"aggregation_bits": "0x" + hex.EncodeToString(bits), "committee_bits": "0xffffffffffffffff", "data": map[string]any{"slot": "3", "index": "0", "target": map[string]any{"epoch": "1", "root": testRoot(1)}}}
		f.attestations(5, "electra", []any{a})
		r, err := collector.Collect(context.Background(), 1, cp, validators)
		if err != nil {
			b.Fatal(err)
		}
		for j := range validators {
			if r.Get(j) != Timely {
				b.Fatalf("validator %d", j)
			}
		}
		b.ReportMetric(float64(collector.Requests.Load()), "RPCs/epoch")
		b.ReportMetric(float64(len(collector.attestations)), "cached-blocks")
	}
}

func BenchmarkBackfillHistory20000(b *testing.B) {
	const count = 20000
	p := testParams(44032, 225, 940000000000000000)
	for iteration := 0; iteration < b.N; iteration++ {
		s := testStore(b)
		validators := make([]Validator, count)
		for i := range validators {
			validators[i] = testValidator(uint64(i))
		}
		if _, err := s.UpdateRegistry(validators); err != nil {
			b.Fatal(err)
		}
		batch := new(leveldb.Batch)
		for epoch := uint64(1); epoch <= p.Period+p.Buffer; epoch++ {
			r := NewRecord(epoch, count)
			r.TargetRoot = testRoot(epoch)
			r.FinalizedRoot = testRoot(epoch + 2)
			for i := range validators {
				o := Timely
				if i < 1000 {
					o = Missed
				}
				r.Set(i, o)
			}
			data, err := encodeRecord(r)
			if err != nil {
				b.Fatal(err)
			}
			batch.Put(epochKey(epoch), data)
			if batch.Len() == 128 {
				if err = s.db.Write(batch, &opt.WriteOptions{Sync: true}); err != nil {
					b.Fatal(err)
				}
				batch.Reset()
			}
		}
		if err := s.db.Write(batch, &opt.WriteOptions{Sync: true}); err != nil {
			b.Fatal(err)
		}
		candidates, err := s.Candidates(context.Background(), validators, p, p.Period+p.Buffer+1, 2, p.Period+p.Buffer, nil, false)
		if err != nil {
			b.Fatal(err)
		}
		if len(candidates) != 1000 {
			b.Fatalf("candidates: %d", len(candidates))
		}
		b.ReportMetric(float64(s.Size()), "disk-bytes")
		b.ReportMetric(float64((count+3)/4*(int(p.Period+p.Buffer))), "packed-history-bytes")
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		b.ReportMetric(float64(m.HeapInuse), "heap-inuse-bytes")
	}
}
