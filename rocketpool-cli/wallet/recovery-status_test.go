package wallet

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Collects reporter output; the reporter writes from its own goroutine
type syncBuffer struct {
	lock sync.Mutex
	buf  bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.buf.String()
}

func TestFormatRecoveryProgress(t *testing.T) {
	tests := []struct {
		name     string
		recovery api.KeyRecoveryStatus
		want     string
	}{
		{
			name:     "before the key count is known",
			recovery: api.KeyRecoveryStatus{Running: true, ElapsedSeconds: 3},
			want:     "Looking up this node's validator keys... (3s elapsed)",
		},
		{
			name:     "partway through",
			recovery: api.KeyRecoveryStatus{Running: true, ElapsedSeconds: 65, KeysFound: 7, KeysTotal: 15, TotalKnown: true},
			want:     "Recovered 7 of 15 validator keys... (1m5s elapsed)",
		},
		{
			name:     "node with no validators",
			recovery: api.KeyRecoveryStatus{Running: true, ElapsedSeconds: 2, TotalKnown: true},
			want:     "No validator keys to recover (2s elapsed)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := strings.TrimSpace(formatRecoveryProgress(test.recovery)); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestProgressReporterPrintsUpdates(t *testing.T) {
	out := &syncBuffer{}
	var polls int
	var lock sync.Mutex

	stop := startProgressReporter(out, false, time.Millisecond, func() (api.KeyRecoveryStatus, error) {
		lock.Lock()
		defer lock.Unlock()
		polls++
		return api.KeyRecoveryStatus{
			Running: true, KeysFound: polls, KeysTotal: 15, TotalKnown: true, ElapsedSeconds: 1,
		}, nil
	})

	// Let it tick a few times, then shut it down
	time.Sleep(50 * time.Millisecond)
	stop()

	got := out.String()
	if !strings.Contains(got, "of 15 validator keys") {
		t.Errorf("expected progress output, got %q", got)
	}

	// Non-interactive output goes one update per line, with no terminal escapes
	if strings.Contains(got, "\r") || strings.Contains(got, "\033[K") {
		t.Errorf("expected no terminal escapes in non-interactive output, got %q", got)
	}

	// Nothing may be written after stop returns
	after := out.String()
	time.Sleep(20 * time.Millisecond)
	if out.String() != after {
		t.Error("reporter kept writing after stop returned")
	}
}

func TestProgressReporterInteractiveRewritesOneLine(t *testing.T) {
	out := &syncBuffer{}
	stop := startProgressReporter(out, true, time.Millisecond, func() (api.KeyRecoveryStatus, error) {
		return api.KeyRecoveryStatus{
			Running: true, KeysFound: 3, KeysTotal: 15, TotalKnown: true, ElapsedSeconds: 1,
		}, nil
	})
	time.Sleep(30 * time.Millisecond)
	stop()

	got := out.String()
	if !strings.Contains(got, "\r\033[K") {
		t.Errorf("expected in-place rewrites on a terminal, got %q", got)
	}
	// The overwritten line has to be closed off so later output starts fresh
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("expected a trailing newline after stop, got %q", got)
	}
}

// A probe that fails, or one that reports nothing running, must stay silent:
// the operation's own result is what tells the user whether it worked.
func TestProgressReporterStaysQuietWhenNothingToReport(t *testing.T) {
	for _, test := range []struct {
		name string
		poll func() (api.KeyRecoveryStatus, error)
	}{
		{"probe fails", func() (api.KeyRecoveryStatus, error) {
			return api.KeyRecoveryStatus{}, errors.New("connection refused")
		}},
		{"nothing running", func() (api.KeyRecoveryStatus, error) {
			return api.KeyRecoveryStatus{Running: false}, nil
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := &syncBuffer{}
			stop := startProgressReporter(out, true, time.Millisecond, test.poll)
			time.Sleep(30 * time.Millisecond)
			stop()
			if got := out.String(); got != "" {
				t.Errorf("expected no output, got %q", got)
			}
		})
	}
}

// A fast operation finishes before the first tick and must print nothing at all
func TestProgressReporterSilentWhenStoppedImmediately(t *testing.T) {
	out := &syncBuffer{}
	stop := startProgressReporter(out, true, time.Hour, func() (api.KeyRecoveryStatus, error) {
		t.Error("reporter polled before its first tick")
		return api.KeyRecoveryStatus{}, nil
	})
	stop()
	if got := out.String(); got != "" {
		t.Errorf("expected no output for a fast operation, got %q", got)
	}
}
