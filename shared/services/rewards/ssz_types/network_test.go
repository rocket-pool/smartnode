package ssz_types

import (
	"encoding/json"
	"testing"
)

func TestPlatabergetNetworkRoundTrip(t *testing.T) {
	network, ok := NetworkFromString("plataberget")
	if !ok || network != 7091047534 {
		t.Fatalf("wrong rewards network: %d, %v", network, ok)
	}
	encoded, err := json.Marshal(network)
	if err != nil || string(encoded) != `"plataberget"` {
		t.Fatalf("marshal network: %s, %v", encoded, err)
	}
	var decoded Network
	if err := json.Unmarshal(encoded, &decoded); err != nil || decoded != network {
		t.Fatalf("unmarshal network: %d, %v", decoded, err)
	}
}
