package rocketpool

import (
	"slices"
	"testing"
)

func TestParseVolumeSizes(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		want         []VolumeSize
		wantErr      bool
	}{
		{"empty", "", nil, false},
		{"mixed", "rocketpool_prometheus-data\trocketpool\t842.5MB\nrocketpoolother_data\trocketpoolother\t9GB\nrocketpool_unlabeled\t\t0B\nrocketpool_eth1clientdata\trocketpool\t1.235TB\n", []VolumeSize{{"rocketpool_eth1clientdata", "1.235TB"}, {"rocketpool_prometheus-data", "842.5MB"}}, false},
		{"size not available", "rocketpool_eth2clientdata\trocketpool\tN/A\n", []VolumeSize{{"rocketpool_eth2clientdata", "N/A"}}, false},
		{"too few fields", "rocketpool_eth1clientdata\trocketpool\n", nil, true},
		{"too many fields", "rocketpool_eth1clientdata\trocketpool\t50MB\textra\n", nil, true},
		{"crlf and blank lines", "\r\n\nrocketpool_eth1clientdata\trocketpool\t50MB\r\n\n", []VolumeSize{{"rocketpool_eth1clientdata", "50MB"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseVolumeSizes(tc.output, "rocketpool")
			if !slices.Equal(got, tc.want) || (err != nil) != tc.wantErr {
				t.Fatalf("got %v, %v; want %v, error=%v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
