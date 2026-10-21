package rocketpool

import "testing"

func TestServiceLogsArgs(t *testing.T) {
	for _, tc := range []struct {
		name, tail, since, until string
		services                 []string
		want                     string
	}{
		{"default", "100", "", "", nil, "logs -f --tail 100"},
		{"mapped services", "100", "", "", []string{"eth1", "eth2", "validator"}, "logs -f --tail 100 eth1 eth2 validator"},
		{"since only", "all", "1h", "", []string{"eth2"}, "logs -f --tail all --since 1h eth2"},
		{"until only", "all", "", "30m", []string{"eth2"}, "logs --tail all --until 30m eth2"},
		{"since and until", "all", "1h", "30m", []string{"eth2"}, "logs --tail all --since 1h --until 30m eth2"},
		{"explicit tail with since", "50", "1h", "", []string{"eth2"}, "logs -f --tail 50 --since 1h eth2"},
		{"tail all without window", "all", "", "", nil, "logs -f --tail all"},
		{"since needs quotes", "all", "2026-10-20 10:00", "", []string{"eth2"}, "logs -f --tail all --since '2026-10-20 10:00' eth2"},
		{"until needs quotes, no services", "all", "", "2026-10-20 10:00", nil, "logs --tail all --until '2026-10-20 10:00'"},
		{"passthrough args", "100", "", "", []string{"--timestamps", "eth2"}, "logs -f --tail 100 --timestamps eth2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := serviceLogsArgs(tc.tail, tc.since, tc.until, tc.services); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
