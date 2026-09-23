package rocketpool

import "testing"

func TestGethDBVersionFromFiles(t *testing.T) {
	for _, tc := range []struct {
		name, files, want string
		wantErr           bool
	}{
		{"empty", "", "none", false},
		{"leveldb", "CURRENT\nMANIFEST-000001\n", "leveldb", false},
		{"legacy pebble", "CURRENT\nOPTIONS-000001\n", "v1", false},
		{"v1 marker", "marker.manifest.000001.MANIFEST-000001\nmarker.format-version.000001.012\n", "v1", false},
		{"migrated v2", "marker.manifest.000001.MANIFEST-000001\nmarker.format-version.000002.013\n", "v2", false},
		{"newer v2", "marker.manifest.000001.MANIFEST-000001\nmarker.format-version.000002.016\n", "v2", false},
		{"stale marker", "CURRENT\nOPTIONS-000001\nmarker.format-version.000002.013\nmarker.format-version.000001.012\n", "v2", false},
		{"incomplete database", "OPTIONS-000001\n", "none", false},
		{"bad marker", "marker.format-version.bad\n", "", true},
		{"bad generation", "marker.format-version.bad.013\n", "", true},
		{"bad version", "marker.format-version.000001.bad\n", "", true},
		{"zero version", "marker.format-version.000001.000\n", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gethDBVersionFromFiles(tc.files)
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("got %q, %v; want %q, error=%v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
