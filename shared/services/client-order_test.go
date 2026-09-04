package services

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestClientOrder(t *testing.T) {
	t.Parallel()

	got := clientOrder(false)
	want := []clientRole{primaryClient, fallbackClient}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("clientOrder(false) = %v, want %v", got, want)
	}

	got = clientOrder(true)
	want = []clientRole{fallbackClient, primaryClient}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("clientOrder(true) = %v, want %v", got, want)
	}
}

func TestTryClients(t *testing.T) {
	t.Parallel()

	disconnectErr := errors.New("dial tcp 127.0.0.1:8545: connect: connection refused")
	appErr := errors.New("execution reverted")
	isDisconnected := func(err error) bool {
		return err != nil && strings.Contains(err.Error(), "dial tcp")
	}

	type disconnectEvent struct {
		failed  string
		next    string
		hasNext bool
	}

	tests := []struct {
		name            string
		preferFallback  bool
		primaryReady    bool
		fallbackReady   bool
		primaryErr      error
		fallbackErr     error
		wantCalled      []clientRole
		wantErr         string
		wantPrimary     bool
		wantFallback    bool
		wantDisconnects []disconnectEvent
	}{
		{
			name:          "default both ready uses primary",
			primaryReady:  true,
			fallbackReady: true,
			wantCalled:    []clientRole{primaryClient},
			wantPrimary:   true,
			wantFallback:  true,
		},
		{
			name:            "default primary disconnect uses fallback",
			primaryReady:    true,
			fallbackReady:   true,
			primaryErr:      disconnectErr,
			wantCalled:      []clientRole{primaryClient, fallbackClient},
			wantPrimary:     false,
			wantFallback:    true,
			wantDisconnects: []disconnectEvent{{failed: "Primary", next: "fallback", hasNext: true}},
		},
		{
			name:           "prefer both ready uses fallback",
			preferFallback: true,
			primaryReady:   true,
			fallbackReady:  true,
			wantCalled:     []clientRole{fallbackClient},
			wantPrimary:    true,
			wantFallback:   true,
		},
		{
			name:            "prefer fallback disconnect uses primary",
			preferFallback:  true,
			primaryReady:    true,
			fallbackReady:   true,
			fallbackErr:     disconnectErr,
			wantCalled:      []clientRole{fallbackClient, primaryClient},
			wantPrimary:     true,
			wantFallback:    false,
			wantDisconnects: []disconnectEvent{{failed: "Fallback", next: "primary", hasNext: true}},
		},
		{
			name:           "prefer both disconnect",
			preferFallback: true,
			primaryReady:   true,
			fallbackReady:  true,
			primaryErr:     disconnectErr,
			fallbackErr:    disconnectErr,
			wantCalled:     []clientRole{fallbackClient, primaryClient},
			wantErr:        "no Beacon clients were ready",
			wantPrimary:    false,
			wantFallback:   false,
			wantDisconnects: []disconnectEvent{
				{failed: "Fallback", next: "primary", hasNext: true},
				{failed: "Primary", hasNext: false},
			},
		},
		{
			name:          "non-disconnect error does not try the other client",
			primaryReady:  true,
			fallbackReady: true,
			primaryErr:    appErr,
			wantCalled:    []clientRole{primaryClient},
			wantErr:       appErr.Error(),
			wantPrimary:   true,
			wantFallback:  true,
		},
		{
			name:           "prefer non-disconnect error on fallback does not try primary",
			preferFallback: true,
			primaryReady:   true,
			fallbackReady:  true,
			fallbackErr:    appErr,
			wantCalled:     []clientRole{fallbackClient},
			wantErr:        appErr.Error(),
			wantPrimary:    true,
			wantFallback:   true,
		},
		{
			name:         "no clients ready",
			wantErr:      "no Beacon clients were ready",
			wantPrimary:  false,
			wantFallback: false,
		},
		{
			name:          "default fallback only",
			fallbackReady: true,
			wantCalled:    []clientRole{fallbackClient},
			wantFallback:  true,
		},
		{
			name:           "prefer primary only",
			preferFallback: true,
			primaryReady:   true,
			wantCalled:     []clientRole{primaryClient},
			wantPrimary:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			primaryReady := tt.primaryReady
			fallbackReady := tt.fallbackReady
			var called []clientRole
			var disconnects []disconnectEvent

			err := tryClients(tt.preferFallback, &primaryReady, &fallbackReady, isDisconnected, func(failedName, nextName string, hasNext bool, _ error) {
				disconnects = append(disconnects, disconnectEvent{failed: failedName, next: nextName, hasNext: hasNext})
			}, "Beacon", func(role clientRole) error {
				called = append(called, role)
				switch role {
				case primaryClient:
					return tt.primaryErr
				case fallbackClient:
					return tt.fallbackErr
				default:
					return fmt.Errorf("unknown role %v", role)
				}
			})

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}

			if !reflect.DeepEqual(called, tt.wantCalled) {
				t.Fatalf("called = %v, want %v", called, tt.wantCalled)
			}
			if primaryReady != tt.wantPrimary {
				t.Fatalf("primaryReady = %v, want %v", primaryReady, tt.wantPrimary)
			}
			if fallbackReady != tt.wantFallback {
				t.Fatalf("fallbackReady = %v, want %v", fallbackReady, tt.wantFallback)
			}
			if !reflect.DeepEqual(disconnects, tt.wantDisconnects) {
				t.Fatalf("disconnects = %+v, want %+v", disconnects, tt.wantDisconnects)
			}
		})
	}
}
