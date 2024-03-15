package config

import "github.com/rocket-pool/node-manager-core/config"

const (
	// Container Names
	NodeSuffix            string = "node"
	ExecutionClientSuffix string = "eth1"
	BeaconNodeSuffix      string = "eth2"
	ValidatorClientSuffix string = "validator"
)

func GetContainerName(id config.ContainerID) string {
	switch id {
	case config.ContainerID_Daemon:
		return NodeSuffix
	case config.ContainerID_BeaconNode:
		return BeaconNodeSuffix
	case config.ContainerID_ExecutionClient:
		return ExecutionClientSuffix
	default:
		return string(id)
	}
}
