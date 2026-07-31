package wallet

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// Each test gets a clean tracker, since it is a package-level singleton
func resetActiveRecovery() {
	activeRecovery = &recoveryTracker{}
}

func TestRecoveryLockRejectsSecondCaller(t *testing.T) {
	resetActiveRecovery()

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		_, err := withRecoveryLock("wallet recover", func() (*struct{}, error) {
			activeRecovery.setKeysTotal(15)
			activeRecovery.setKeysFound(7)
			close(started)
			<-release
			return &struct{}{}, nil
		})
		done <- err
	}()

	<-started

	// A second attempt must be turned away while the first holds the lock
	resp, err := withRecoveryLock("wallet rebuild", func() (*struct{}, error) {
		t.Error("second caller ran while the first held the lock")
		return &struct{}{}, nil
	})
	if err == nil {
		t.Fatal("expected the second caller to be rejected, got nil error")
	}
	if resp != nil {
		t.Errorf("expected a nil response for a rejected caller, got %v", resp)
	}

	// The rejection must describe what is actually running, not just that something is
	msg := err.Error()
	for _, want := range []string{"wallet recover", "7 of 15"} {
		if !strings.Contains(msg, want) {
			t.Errorf("expected the rejection to mention %q, got: %s", want, msg)
		}
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first caller failed: %v", err)
	}

	// The lock must be free again once the first caller returns
	if _, err := withRecoveryLock("wallet rebuild", func() (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		t.Errorf("expected the lock to be released, got: %v", err)
	}
}

func TestRecoveryLockReleasesOnError(t *testing.T) {
	resetActiveRecovery()

	sentinel := errors.New("recovery blew up")
	if _, err := withRecoveryLock("wallet recover", func() (*struct{}, error) {
		return nil, sentinel
	}); !errors.Is(err, sentinel) {
		t.Fatalf("expected the operation's error to propagate, got: %v", err)
	}

	if activeRecovery.status().Running {
		t.Error("lock still held after the operation returned an error")
	}
}

// A panic in a recovery must not wedge the daemon into permanently refusing
// every later attempt, since the lock only clears on a daemon restart.
func TestRecoveryLockReleasesOnPanic(t *testing.T) {
	resetActiveRecovery()

	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected the panic to propagate to the caller")
			}
		}()
		_, _ = withRecoveryLock("wallet recover", func() (*struct{}, error) {
			panic("boom")
		})
	}()

	if activeRecovery.status().Running {
		t.Error("lock still held after the operation panicked")
	}
}

func TestRecoveryStatusReportsProgress(t *testing.T) {
	resetActiveRecovery()

	if status := activeRecovery.status(); status.Running {
		t.Error("expected an idle tracker to report Running=false")
	}

	acquired, _ := activeRecovery.begin("wallet rebuild")
	if !acquired {
		t.Fatal("could not claim an idle tracker")
	}
	defer activeRecovery.end()

	// Before the daemon knows how many keys the node has
	status := activeRecovery.status()
	if !status.Running || status.Operation != "wallet rebuild" {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.TotalKnown {
		t.Error("expected TotalKnown to be false before the key count is known")
	}

	activeRecovery.setKeysTotal(15)
	activeRecovery.setKeysFound(3)

	status = activeRecovery.status()
	if !status.TotalKnown || status.KeysTotal != 15 || status.KeysFound != 3 {
		t.Errorf("expected 3 of 15 keys with the total known, got: %+v", status)
	}
	if status.ElapsedSeconds < 0 || status.StartedAt.IsZero() {
		t.Errorf("expected a sane elapsed time, got: %+v", status)
	}
}

// The tracker is read by the status endpoint while a recovery writes to it.
func TestRecoveryTrackerIsRaceFree(t *testing.T) {
	resetActiveRecovery()

	acquired, _ := activeRecovery.begin("wallet recover")
	if !acquired {
		t.Fatal("could not claim an idle tracker")
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			activeRecovery.setKeysFound(i)
		}
		close(stop)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				activeRecovery.status()
			}
		}
	}()

	wg.Wait()
	activeRecovery.end()

	if activeRecovery.status().Running {
		t.Error("expected the tracker to be idle after end()")
	}
}
