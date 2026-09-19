package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func getDBDetails() error {
	rp := rocketpool.NewClient()
	defer rp.Close()

	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}
	if cfg.IsNativeMode || cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External {
		fmt.Println("Database details are only available for execution clients managed by Smart Node in Docker mode.")
		return nil
	}

	container := cfg.Smartnode.ProjectName.Value.(string) + ExecutionContainerSuffix
	switch cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient) {
	case cfgtypes.ExecutionClient_Nethermind:
		layout, err := rp.NethermindDBLayout(container)
		if err != nil {
			return err
		}
		switch layout {
		case "flat":
			fmt.Println("Nethermind is using FlatDB. Manual state pruning is not needed or supported.")
		case "patricia":
			fmt.Println("Nethermind is using the legacy Patricia database. State pruning is supported.")
		case "none":
			fmt.Println("Nethermind has no persisted state database yet.")
		}
	case cfgtypes.ExecutionClient_Geth:
		version, err := rp.GethDBVersion(container)
		if err != nil {
			return err
		}
		switch version {
		case "v1":
			fmt.Println("Geth is using Pebble v1. Run `rocketpool service migrate-geth` to migrate to Pebble v2.")
		case "v2":
			fmt.Println("Geth is using Pebble v2. No database migration is needed.")
		case "leveldb":
			fmt.Println("Geth is using LevelDB. Pebble v1/v2 migration does not apply to this database.")
		case "none":
			fmt.Println("Geth has no database yet.")
		}
	default:
		fmt.Println("Database details are currently supported for Nethermind and Geth only.")
	}
	return nil
}
