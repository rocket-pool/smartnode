package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBeaconSlotDuration(t *testing.T) {
	for _, tc := range []struct {
		name, fields string
		wantError    bool
	}{
		{"legacy", `"SECONDS_PER_SLOT":"12"`, false},
		{"plataberget", `"SLOT_DURATION_MS":"12000"`, false},
		{"both", `"SECONDS_PER_SLOT":"12","SLOT_DURATION_MS":"12000"`, false},
		{"missing", `"SLOT_DURATION_MS":"0"`, true},
		{"fractional", `"SLOT_DURATION_MS":"12500"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			eth2ConfigCache.Store(nil)
			t.Cleanup(func() { eth2ConfigCache.Store(nil) })
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == RequestEth2ConfigPath {
					fmt.Fprintf(w, `{"data":{%s,"SLOTS_PER_EPOCH":"32","GLOAS_FORK_EPOCH":"1536"}}`, tc.fields)
				} else {
					fmt.Fprint(w, `{"data":{"genesis_time":"1786622400","genesis_fork_version":"0x10733183","genesis_validators_root":"0xbb4a1a9e3f7f4e10edcd734e4acc3b5ffd4f830efe0af2748fa458cfee5d2658"}}`)
				}
			}))
			defer server.Close()
			cfg, err := NewStandardHttpClient(server.URL).GetEth2Config()
			if tc.wantError {
				if err == nil {
					t.Fatal("invalid slot duration accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.SecondsPerSlot != 12 || cfg.SecondsPerEpoch != 384 || cfg.GloasForkEpoch != 1536 {
				t.Fatalf("incorrect slot/fork config: %+v", cfg)
			}
		})
	}
}
