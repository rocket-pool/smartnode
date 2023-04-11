package migration

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

func upgradeFromV191(serializedConfig map[string]map[string]string) error {
	// v1.9.1 had the BN API port mode as a boolean
	consensusCommon, exists := serializedConfig["consensusCommon"]
	if !exists {
		return fmt.Errorf("expected a section called `consensusCommon` but it didn't exist")
	}
	openApiPort, exists := consensusCommon["openApiPort"]
	if !exists {
		return fmt.Errorf("expected a consensusCommon setting named `openApiPort` but it didn't exist")
	}

	// Update the config
	if openApiPort == "true" {
		consensusCommon["openApiPort"] = string(config.RPC_OpenLocalhost)
	} else {
		consensusCommon["openApiPort"] = string(config.RPC_Closed)
	}
	serializedConfig["consensusCommon"] = consensusCommon

	// v1.9.1 had the EC API ports mode as a boolean
	executionCommon, exists := serializedConfig["executionCommon"]
	if !exists {
		return fmt.Errorf("expected a section called `executionCommon` but it didn't exist")
	}
	openRPCPorts, exists := executionCommon["openRpcPorts"]
	if !exists {
		return fmt.Errorf("expected a executionCommon setting named `openRPCPorts` but it didn't exist")
	}

	// Update the config
	if openRPCPorts == "true" {
		executionCommon["openRPCPorts"] = string(config.RPC_OpenLocalhost)
	} else {
		executionCommon["openRPCPorts"] = string(config.RPC_Closed)
	}
	serializedConfig["executionCommon"] = executionCommon
	return nil
}
