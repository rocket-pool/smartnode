package migration

import "fmt"

func upgradeFromV1153(serializedConfig map[string]map[string]string) error {
	// v1.15.3 had the max block gas default as 30M. We're changing the default to mean the SN will use the EC default.

	executionCommonSettings, exists := serializedConfig["executionCommon"]
	if !exists {
		return fmt.Errorf("expected a section called `executionCommon` but it didn't exist")
	}

	config, exists := executionCommonSettings["suggestedBlockGasLimit"]
	if !exists {
		return nil
	}

	// If using the previous or current defaults, change to the new one
	if config == "30000000" || config == "36000000" {
		executionCommonSettings["suggestedBlockGasLimit"] = ""
	}

	serializedConfig["executionCommon"] = executionCommonSettings

	return nil
}
