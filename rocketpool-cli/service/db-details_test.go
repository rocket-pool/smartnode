package service

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func TestGetDBDetails(t *testing.T) {
	for _, tc := range []struct {
		name, response, want string
		client               cfgtypes.ExecutionClient
		external, wantErr    bool
	}{
		{name: "flat", client: cfgtypes.ExecutionClient_Nethermind, response: "flat", want: "Nethermind is using FlatDB"},
		{name: "patricia", client: cfgtypes.ExecutionClient_Nethermind, response: "patricia", want: "legacy Patricia"},
		{name: "no nethermind state", client: cfgtypes.ExecutionClient_Nethermind, response: "none", want: "no persisted state"},
		{name: "unknown nethermind layout", client: cfgtypes.ExecutionClient_Nethermind, response: "unknown", wantErr: true},
		{name: "geth v1", client: cfgtypes.ExecutionClient_Geth, response: "CURRENT\nOPTIONS-000001", want: "Geth is using Pebble v1"},
		{name: "geth v2", client: cfgtypes.ExecutionClient_Geth, response: "marker.manifest.000001.MANIFEST-000001\nmarker.format-version.000001.013", want: "Geth is using Pebble v2"},
		{name: "leveldb", client: cfgtypes.ExecutionClient_Geth, response: "CURRENT", want: "Geth is using LevelDB"},
		{name: "no geth database", client: cfgtypes.ExecutionClient_Geth, want: "Geth has no database yet"},
		{name: "container unavailable", client: cfgtypes.ExecutionClient_Geth, response: "fail", wantErr: true},
		{name: "external", client: cfgtypes.ExecutionClient_Geth, external: true, response: "fail", want: "only available for execution clients managed by Smart Node"},
		{name: "unsupported", client: cfgtypes.ExecutionClient_Besu, response: "fail", want: "Nethermind and Geth only"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg, err := config.NewRocketPoolConfig(dir, false)
			if err != nil {
				t.Fatal(err)
			}
			cfg.IsCLI = true
			cfg.Smartnode.DataPath.Value = filepath.Join(dir, "data")
			cfg.ExecutionClient.Value = tc.client
			cfg.ExecutionClientMode.Value = cfgtypes.Mode_Local
			if tc.external {
				cfg.ExecutionClientMode.Value = cfgtypes.Mode_External
			}
			if err := cfg.Save(dir, rocketpool.SettingsFile); err != nil {
				t.Fatal(err)
			}
			oldDefaults := rocketpool.Defaults
			rocketpool.Defaults.ConfigPath = dir
			t.Cleanup(func() { rocketpool.Defaults = oldDefaults })
			// Exercise the Docker command boundary without a real execution node.
			script := `#!/bin/sh
[ "$1" = exec ] && [ "$2" = rocketpool_eth1 ] && [ "$3" = sh ] || exit 2
if [ "$4" = -c ]; then
    [ "$#" = 5 ] || exit 2
    case "$5" in *'/ethclient/geth/geth/chaindata'*) ;; *) exit 2 ;; esac
else
    [ "$#" = 4 ] && [ "$4" = /setup/nethermind-db.sh ] || exit 2
fi
[ "$TEST_DB_RESPONSE" != fail ] || exit 1
printf '%s\n' "$TEST_DB_RESPONSE"
`
			if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
			t.Setenv("TEST_DB_RESPONSE", tc.response)
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			stdout := os.Stdout
			os.Stdout = writer
			defer func() { os.Stdout = stdout }()
			err = getDBDetails()
			writer.Close()
			os.Stdout = stdout
			output, readErr := io.ReadAll(reader)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if (err != nil) != tc.wantErr || !strings.Contains(string(output), tc.want) {
				t.Fatalf("got %q, %v; want %q, error=%v", output, err, tc.want, tc.wantErr)
			}
		})
	}
}
