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
