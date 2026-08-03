package wallet

import (
	"fmt"
	"sync"
	"time"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Guards validator key recovery, which is the one wallet operation that routinely
// outlives the CLI invocation that started it
type recoveryTracker struct {
	lock       sync.Mutex
	running    bool
	operation  string
	startedAt  time.Time
	keysFound  int
	keysTotal  int
	totalKnown bool
}

var activeRecovery = &recoveryTracker{}

// Claims the tracker for the named operation. If another recovery already holds
// it, returns false along with a snapshot of that recovery's progress.
func (t *recoveryTracker) begin(operation string) (bool, api.KeyRecoveryStatus) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.running {
		return false, t.snapshot()
	}

	t.running = true
	t.operation = operation
	t.startedAt = time.Now()
	t.keysFound = 0
	t.keysTotal = 0
	t.totalKnown = false
	return true, api.KeyRecoveryStatus{}
}

// Releases the tracker
func (t *recoveryTracker) end() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.running = false
}

// Records how many validator keys the running recovery is looking for
func (t *recoveryTracker) setKeysTotal(total int) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.keysTotal = total
	t.totalKnown = true
}

// Records how many of those keys have been recovered so far
func (t *recoveryTracker) setKeysFound(found int) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.keysFound = found
}

// Returns a snapshot of the current state for reporting
func (t *recoveryTracker) status() api.KeyRecoveryStatus {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.snapshot()
}

// Builds a status value; the caller must hold the lock
func (t *recoveryTracker) snapshot() api.KeyRecoveryStatus {
	if !t.running {
		return api.KeyRecoveryStatus{}
	}
	return api.KeyRecoveryStatus{
		Running:        true,
		Operation:      t.operation,
		StartedAt:      t.startedAt,
		ElapsedSeconds: time.Since(t.startedAt).Seconds(),
		KeysFound:      t.keysFound,
		KeysTotal:      t.keysTotal,
		TotalKnown:     t.totalKnown,
	}
}

// Runs fn only if no other validator key recovery is in flight, so a user who
// closed the CLI and re-ran the command gets told what is already happening
// instead of starting a second competing recovery
func withRecoveryLock[T any](operation string, fn func() (*T, error)) (*T, error) {
	acquired, busy := activeRecovery.begin(operation)
	if !acquired {
		return nil, recoveryInProgressError(busy)
	}
	defer activeRecovery.end()
	return fn()
}

// Describes a rejected attempt to start a second recovery
func recoveryInProgressError(busy api.KeyRecoveryStatus) error {
	elapsed := time.Duration(busy.ElapsedSeconds * float64(time.Second)).Round(time.Second)
	progress := "no validator keys have been recovered yet"
	if busy.TotalKnown {
		progress = fmt.Sprintf("%d of %d validator keys have been recovered", busy.KeysFound, busy.KeysTotal)
	}
	return fmt.Errorf("a validator key recovery (%s) is already running on the node daemon: it started %s ago and %s. Closing the CLI does not stop it; wait for it to finish and try again, or restart the node daemon if it is stuck", busy.Operation, elapsed, progress)
}
