package config

import (
	"fmt"
	"path/filepath"

	"github.com/rocket-pool/node-manager-core/config"
)

func (cfg *SmartNodeConfig) GetNetworkResources() *config.NetworkResources {
	return cfg.GetRocketPoolResources().NetworkResources
}

func (cfg *SmartNodeConfig) GetRocketPoolResources() *SmartNodeResources {
	return newSmartNodeResources(cfg.Network.Value)
}

func (cfg *SmartNodeConfig) GetVotingPath() string {
	return filepath.Join(cfg.UserDataPath.Value, VotingFolder, string(cfg.Network.Value))
}

func (cfg *SmartNodeConfig) GetRecordsPath() string {
	return filepath.Join(cfg.UserDataPath.Value, RecordsFolder)
}

func (cfg *SmartNodeConfig) GetRewardsTreePath(interval uint64) string {
	return filepath.Join(cfg.UserDataPath.Value, RewardsTreesFolder, fmt.Sprintf(RewardsTreeFilenameFormat, string(cfg.Network.Value), interval))
}

func (cfg *SmartNodeConfig) GetVotingSnapshotID() [32]byte {
	// So the contract wants a Keccak'd hash of the voting ID, but Snapshot's service wants ASCII so it can display the ID in plain text; we have to do this to make it play nicely with Snapshot
	buffer := [32]byte{}
	idBytes := []byte(SnapshotID)
	copy(buffer[0:], idBytes)
	return buffer
}

func (cfg *SmartNodeConfig) GetRegenerateRewardsTreeRequestPath(interval uint64) string {
	return filepath.Join(cfg.UserDataPath.Value, WatchtowerFolder, fmt.Sprintf(RegenerateRewardsTreeRequestFormat, interval))
}
