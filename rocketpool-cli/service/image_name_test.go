package service

import "testing"

func TestGetDockerImageName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"sigp/lighthouse:v8.2.2", "sigp/lighthouse"},
		{"sigp/lighthouse:v8.2.2@sha256:abc", "sigp/lighthouse"},
		{"statusim/nimbus-validator-client:multiarch-v26.8.0", "statusim/nimbus-validator-client"},
		{"statusim/nimbus-validator-client:multiarch-v26.8.0@sha256:deadbeef", "statusim/nimbus-validator-client"},
		{"gcr.io/offchainlabs/prysm/validator:v7.1.8", "gcr.io/offchainlabs/prysm/validator"},
		{"localhost:5000/lighthouse:v1", "localhost:5000/lighthouse"},
		{"sigp/lighthouse@sha256:abc", "sigp/lighthouse"},
	}
	for _, tt := range tests {
		got, err := getDockerImageName(tt.in)
		if err != nil {
			t.Fatalf("%q: %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("%q: got %q want %q", tt.in, got, tt.want)
		}
	}

	running, err := getDockerImageName("sigp/lighthouse:v8.2.1@sha256:aaa")
	if err != nil {
		t.Fatal(err)
	}
	pending, err := getDockerImageName("sigp/lighthouse:v8.2.2")
	if err != nil {
		t.Fatal(err)
	}
	if running != pending {
		t.Fatalf("same client should match across tag/digest: %q vs %q", running, pending)
	}

	oldClient, _ := getDockerImageName("statusim/nimbus-validator-client:multiarch-v26.8.0@sha256:abc")
	newClient, _ := getDockerImageName("sigp/lighthouse:v8.2.2")
	if oldClient == newClient {
		t.Fatal("different clients should not match")
	}
}
