package config

import "testing"

func mustNewRocketPoolConfig(t testing.TB, rpDir string, native bool) *RocketPoolConfig {
	t.Helper()
	cfg, err := NewRocketPoolConfig(rpDir, native)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}
