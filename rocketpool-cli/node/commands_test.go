package node

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestPendingAndCancelCommandsRegistered(t *testing.T) {
	app := &cli.Command{}
	RegisterCommands(app, "node", []string{"n"})

	if len(app.Commands) != 1 {
		t.Fatalf("expected 1 command group, got %d", len(app.Commands))
	}

	nodeCmd := app.Commands[0]
	if nodeCmd.Name != "node" {
		t.Fatalf("expected command group 'node', got '%s'", nodeCmd.Name)
	}

	foundPending := false
	foundCancel := false

	for _, subcmd := range nodeCmd.Commands {
		if subcmd.Name == "pending-transactions" {
			foundPending = true
			expectedAliases := map[string]bool{"pending-txs": true, "pending": true, "txs": true}
			for _, alias := range subcmd.Aliases {
				delete(expectedAliases, alias)
			}
			if len(expectedAliases) != 0 {
				t.Errorf("missing expected aliases for pending-transactions: %v", expectedAliases)
			}
		}

		if subcmd.Name == "cancel-transaction" {
			foundCancel = true
			expectedAliases := map[string]bool{"cancel-tx": true, "cancel": true}
			for _, alias := range subcmd.Aliases {
				delete(expectedAliases, alias)
			}
			if len(expectedAliases) != 0 {
				t.Errorf("missing expected aliases for cancel-transaction: %v", expectedAliases)
			}

			// Verify flags: nonce, all, yes
			flagsFound := map[string]bool{}
			for _, flag := range subcmd.Flags {
				for _, name := range flag.Names() {
					flagsFound[name] = true
				}
			}
			if !flagsFound["nonce"] || !flagsFound["n"] {
				t.Error("cancel-transaction missing nonce flag")
			}
			if !flagsFound["all"] || !flagsFound["a"] {
				t.Error("cancel-transaction missing all flag")
			}
			if !flagsFound["yes"] || !flagsFound["y"] {
				t.Error("cancel-transaction missing yes flag")
			}
		}
	}

	if !foundPending {
		t.Error("pending-transactions command not registered under node")
	}
	if !foundCancel {
		t.Error("cancel-transaction command not registered under node")
	}
}
