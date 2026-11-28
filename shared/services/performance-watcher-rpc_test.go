package services

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcherSharedRequestBudget(t *testing.T) {
	budget := make(watcherBudget, 4)
	var active, peak atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := budget.acquire(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
			n := active.Add(1)
			for {
				old := peak.Load()
				if old >= n || peak.CompareAndSwap(old, n) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			active.Add(-1)
			release()
		}()
	}
	wg.Wait()
	if peak.Load() != 4 {
		t.Fatalf("max concurrency %d", peak.Load())
	}
	for i := 0; i < 4; i++ {
		budget <- struct{}{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := budget.acquire(ctx); err == nil {
		t.Fatal("canceled acquisition succeeded")
	}
}
