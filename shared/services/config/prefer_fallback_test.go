package config

import "testing"

func TestPreferFallbackDefaultAndDeserialize(t *testing.T) {
	cfg := mustNewRocketPoolConfig(t, "/tmp/rp-test", false)
	if cfg.PreferFallback.Value != false {
		t.Fatalf("default PreferFallback = %v, want false", cfg.PreferFallback.Value)
	}

	serialized := cfg.Serialize()
	if got := serialized["root"]["preferFallback"]; got != "false" {
		t.Fatalf("serialized preferFallback = %q, want false", got)
	}

	delete(serialized["root"], "preferFallback")
	loaded := mustNewRocketPoolConfig(t, "/tmp/rp-test", false)
	if err := loaded.Deserialize(serialized); err != nil {
		t.Fatalf("deserialize without key: %v", err)
	}
	if loaded.PreferFallback.Value != false {
		t.Fatalf("missing key deserialized to %v, want false", loaded.PreferFallback.Value)
	}

	serialized["root"]["preferFallback"] = "true"
	loaded = mustNewRocketPoolConfig(t, "/tmp/rp-test", false)
	if err := loaded.Deserialize(serialized); err != nil {
		t.Fatalf("deserialize true: %v", err)
	}
	if loaded.PreferFallback.Value != true {
		t.Fatalf("true deserialized to %v, want true", loaded.PreferFallback.Value)
	}
}
