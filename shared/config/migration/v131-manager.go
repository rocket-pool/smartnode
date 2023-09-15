package migration

import "fmt"

func upgradeFromV131(serializedConfig map[string]map[string]string) error {
	// v1.3.1 had some of the common EC parameters stored inside the Geth config
	gethSettings, exists := serializedConfig["geth"]
	if !exists {
		return fmt.Errorf("expected a section called `geth` but it didn't exist")
	}
	p2pPort, exists := gethSettings["p2pPort"]
	if !exists {
		return fmt.Errorf("expected a Geth setting named `p2pPort` but it didn't exist")
	}
	ethstatsLabel, exists := gethSettings["ethstatsLabel"]
	if !exists {
		return fmt.Errorf("expected a Geth setting named `ethstatsLabel` but it didn't exist")
	}
	ethstatsLogin, exists := gethSettings["ethstatsLogin"]
	if !exists {
		return fmt.Errorf("expected a Geth setting named `ethstatsLogin` but it didn't exist")
	}

	// Update the config with them
	executionCommonSettings, exists := serializedConfig["executionCommon"]
	if !exists {
		return fmt.Errorf("expected a section called `executionCommon` but it didn't exist")
	}
	executionCommonSettings["p2pPort"] = p2pPort
	executionCommonSettings["ethstatsLabel"] = ethstatsLabel
	executionCommonSettings["ethstatsLogin"] = ethstatsLogin
	serializedConfig["executionCommon"] = executionCommonSettings

	return nil
}
