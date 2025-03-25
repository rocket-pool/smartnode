package migration

import "fmt"

const previousGasLimitDefault = "36000000"

func upgradeFromV1153(serializedConfig map[string]map[string]string) error {
	// If the gas limit setting is 36M we're changing the default to mean the SN will use the EC default.

	executionCommonSettings, exists := serializedConfig["executionCommon"]
	if !exists {
		return fmt.Errorf("expected a section called `executionCommon` but it didn't exist")
	}

	config, exists := executionCommonSettings["suggestedBlockGasLimit"]
	if !exists {
		return nil
	}

	// If using the current default, change to blank (to use the clients default)
	if config == previousGasLimitDefault {
		executionCommonSettings["suggestedBlockGasLimit"] = ""
	}

	serializedConfig["executionCommon"] = executionCommonSettings

	consensusCommonSettings, exists := serializedConfig["consensusCommon"]
	if !exists {
		return fmt.Errorf("expected a section called `consensusCommon` but it didn't exist")
	}

	config, exists = consensusCommonSettings["suggestedBlockGasLimit"]
	if !exists {
		return nil
	}

	// If using the current default, change to blank (to use the clients default)
	if config == previousGasLimitDefault {
		consensusCommonSettings["suggestedBlockGasLimit"] = ""
	}

	serializedConfig["consensusCommon"] = consensusCommonSettings

	return nil
}
