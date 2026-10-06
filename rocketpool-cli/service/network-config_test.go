package service

import (
	"context"
	"strings"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/urfave/cli/v3"
)

func TestHeadlessNetworkClientSelection(t *testing.T) {
	for _, tc := range []struct {
		name, from, to, client string
		wantError              bool
	}{
		{"defaults", "mainnet", "plataberget", "", false},
		{"supported", "testnet", "plataberget", "nethermind", false},
		{"unsupported", "mainnet", "plataberget", "geth", true},
		{"restore", "plataberget", "testnet", "geth", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.NewRocketPoolConfig(t.TempDir(), false)
			if err != nil {
				t.Fatal(err)
			}
			cfg.ChangeNetwork(cfgtypes.Network(tc.from))
			cmd := &cli.Command{
				Name:   "config",
				Flags:  []cli.Flag{&cli.StringFlag{Name: "smartnode-network"}, &cli.StringFlag{Name: "executionClient"}},
				Action: func(_ context.Context, c *cli.Command) error { return updateConfigFromCliArgs(c, cfg) },
			}
			args := []string{"config", "--smartnode-network", tc.to}
			if tc.client != "" {
				args = append(args, "--executionClient", tc.client)
			}
			err = cmd.Run(context.Background(), args)
			if tc.wantError {
				if err == nil || !strings.Contains(err.Error(), "valid options") {
					t.Fatalf("unsupported client accepted: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.GetNetwork() != cfgtypes.Network(tc.to) {
				t.Fatal("network was not changed")
			}
			if tc.to == "plataberget" && cfg.ExecutionClient.Value != cfgtypes.ExecutionClient_Nethermind {
				t.Fatal("headless selection did not apply the network's default client")
			}
		})
	}
}
