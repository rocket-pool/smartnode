package migration

import (
	"maps"
	"reflect"
	"slices"
	"strconv"
	"testing"
)

func TestUpdateConfigEpbsMigration(t *testing.T) {
	regulated := []string{"flashbotsEnabled", "bloxRouteRegulatedEnabled", "ultrasoundFilteredEnabled", "titanRegionalEnabled", "btcsOfacEnabled"}
	unregulated := []string{"ultrasoundEnabled", "aestusEnabled", "titanGlobalEnabled"}
	all := slices.Concat(regulated, unregulated)
	for _, test := range []struct {
		name string
		root map[string]string
		mev  map[string]string
		cb   map[string]string
		want []string
	}{
		{
			name: "MEV manual ignores profiles",
			mev: map[string]string{
				"selectionMode": "relay", "aestusEnabled": "true", "btcsOfacEnabled": "true",
				"enableRegulatedAllMev": "true", "enableUnregulatedAllMev": "true",
			},
			want: []string{"aestusEnabled", "btcsOfacEnabled"},
		},
		{
			name: "MEV regulated profile ignores manual choices",
			mev:  map[string]string{"selectionMode": "profile", "enableRegulatedAllMev": "true", "aestusEnabled": "true"},
			want: regulated,
		},
		{
			name: "MEV unregulated profile",
			mev:  map[string]string{"selectionMode": "profile", "enableUnregulatedAllMev": "true"},
			want: unregulated,
		},
		{
			name: "MEV both profiles",
			mev:  map[string]string{"selectionMode": "profile", "enableRegulatedAllMev": "true", "enableUnregulatedAllMev": "true"},
			want: all,
		},
		{
			name: "MEV missing settings use legacy defaults",
			mev:  map[string]string{"enableUnregulatedAllMev": "true"},
			want: unregulated,
		},
		{
			name: "MEV empty profile ignores stale manual choices",
			mev:  map[string]string{"selectionMode": "profile", "flashbotsEnabled": "true"},
		},
		{
			name: "Commit-Boost manual",
			root: map[string]string{"enableMevBoost": "false", "enableCommitBoost": "true"},
			cb:   map[string]string{"relaySelectionMode": "manual", "cbTitanRegionalEnabled": "true", "cbBtcsOfacEnabled": "true"},
			want: []string{"titanRegionalEnabled", "btcsOfacEnabled"},
		},
		{
			name: "Commit-Boost all ignores disabled manual choices",
			root: map[string]string{"enableMevBoost": "false", "enableCommitBoost": "true"},
			cb:   map[string]string{"relaySelectionMode": "all", "cbFlashbotsEnabled": "false"},
			want: regulated,
		},
		{
			name: "Commit-Boost missing settings use legacy defaults",
			root: map[string]string{"enableMevBoost": "false", "enableCommitBoost": "true"},
			want: regulated,
		},
		{
			name: "disabled services retain their settings without selecting builders",
			root: map[string]string{"enableMevBoost": "false", "enableCommitBoost": "false"},
			mev:  map[string]string{"selectionMode": "profile", "enableRegulatedAllMev": "true"},
			cb:   map[string]string{"relaySelectionMode": "all"},
		},
		{
			name: "external MEV service has no known relay selection",
			mev:  map[string]string{"mode": "external", "externalUrl": "http://mev.example:18550", "enableRegulatedAllMev": "true"},
		},
		{
			name: "external Commit-Boost service has no known relay selection",
			root: map[string]string{"enableMevBoost": "false", "enableCommitBoost": "true"},
			cb:   map[string]string{"mode": "external", "externalUrl": "http://cb.example:18550", "relaySelectionMode": "all"},
		},
		{
			name: "custom relays remain in Commit-Boost",
			root: map[string]string{"enableMevBoost": "false", "enableCommitBoost": "true"},
			cb:   map[string]string{"relaySelectionMode": "manual", "customRelays": "https://0x123@custom.example", "cbFlashbotsEnabled": "true"},
			want: []string{"flashbotsEnabled"},
		},
		{
			name: "both enabled retains existing MEV precedence",
			root: map[string]string{"enableMevBoost": "true", "enableCommitBoost": "true"},
			mev:  map[string]string{"selectionMode": "relay", "aestusEnabled": "true"},
			cb:   map[string]string{"relaySelectionMode": "all"},
			want: []string{"aestusEnabled"},
		},
		{
			name: "external MEV does not fall back to inactive local Commit-Boost",
			root: map[string]string{"enableMevBoost": "true", "enableCommitBoost": "true"},
			mev:  map[string]string{"mode": "external"},
			cb:   map[string]string{"relaySelectionMode": "all"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := map[string]string{"version": "v1.24.4"}
			maps.Copy(root, test.root)
			serialized := map[string]map[string]string{
				"root": root, "mevBoost": test.mev, "commitBoostConfig": test.cb,
			}
			before := make(map[string]map[string]string)
			for section, params := range serialized {
				before[section] = maps.Clone(params)
			}
			if err := UpdateConfig(serialized); err != nil {
				t.Fatal(err)
			}
			epbs := serialized["epbs"]
			if len(epbs) != 10 || epbs["minBid"] != "" || epbs["maxExecutionPayment"] != "0" {
				t.Fatalf("unexpected builder settings: %v", epbs)
			}
			for _, key := range all {
				want := strconv.FormatBool(slices.Contains(test.want, key))
				if epbs[key] != want {
					t.Errorf("%s = %q, want %q", key, epbs[key], want)
				}
			}
			for section, params := range before {
				if !reflect.DeepEqual(serialized[section], params) {
					t.Errorf("migration changed existing section %s", section)
				}
			}
			epbs = maps.Clone(epbs)
			if err := UpdateConfig(serialized); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(serialized["epbs"], epbs) {
				t.Fatal("repeated migration changed builder settings")
			}
		})
	}
}

func TestUpdateConfigPreservesExistingEpbsSettings(t *testing.T) {
	for _, epbs := range []map[string]string{
		{},
		{"minBid": ""},
		{"minBid": "0.05", "flashbotsEnabled": "false", "aestusEnabled": "true"},
		{"minBid": "0.05", "maxExecutionPayment": "0"},
		{"minBid": "0.05", "maxExecutionPayment": "0.02"},
	} {
		serialized := map[string]map[string]string{
			"root":     {"version": "v1.24.4", "enableMevBoost": "true"},
			"mevBoost": {"selectionMode": "profile", "enableRegulatedAllMev": "true"},
			"epbs":     maps.Clone(epbs),
		}
		if err := UpdateConfig(serialized); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(serialized["epbs"], epbs) {
			t.Fatalf("overwrote existing settings: got %v, want %v", serialized["epbs"], epbs)
		}
	}
}

func TestUpdateConfigEpbsVersionBoundary(t *testing.T) {
	for _, test := range []struct {
		version string
		migrate bool
	}{
		{"v1.24.3", true},
		{"1.24.4", true},
		{"v1.24.4-dev", true},
		{"v1.24.5-dev", true},
		{"v1.24.5", true},
		{"v1.24.6", false},
		{"v1.25.0", false},
	} {
		t.Run(test.version, func(t *testing.T) {
			serialized := map[string]map[string]string{"root": {"version": test.version}}
			if err := UpdateConfig(serialized); err != nil {
				t.Fatal(err)
			}
			if _, migrated := serialized["epbs"]; migrated != test.migrate {
				t.Fatalf("migrated = %v, want %v", migrated, test.migrate)
			}
		})
	}
}
