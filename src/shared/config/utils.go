package config

import (
	"fmt"
	"path/filepath"

	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/log"
)

func (cfg *SmartNodeConfig) GetNetworkResources() *config.NetworkResources {
	return cfg.GetRocketPoolResources().NetworkResources
}

func (cfg *SmartNodeConfig) GetRocketPoolResources() *RocketPoolResources {
	return cfg.resources
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

func (cfg *SmartNodeConfig) GetNextAccountFilePath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserNextAccountFilename)
}

func (cfg *SmartNodeConfig) GetValidatorsFolderPath() string {
	return filepath.Join(cfg.UserDataPath.Value, ValidatorsFolderName)
}

func (cfg *SmartNodeConfig) GetCustomKeyPath() string {
	return filepath.Join(cfg.UserDataPath.Value, CustomKeysFolderName)
}

func (cfg *SmartNodeConfig) GetCustomKeyPasswordFilePath() string {
	return filepath.Join(cfg.UserDataPath.Value, CustomKeyPasswordFilename)
}

func (cfg *SmartNodeConfig) GetFeeRecipientFilePath() string {
	return filepath.Join(cfg.UserDataPath.Value, ValidatorsFolderName, FeeRecipientFilename)
}

func (cfg *SmartNodeConfig) GetWatchtowerFolder() string {
	return filepath.Join(cfg.UserDataPath.Value, WatchtowerFolder)
}

func (cfg *SmartNodeConfig) GetMinipoolPerformancePath(interval uint64) string {
	return filepath.Join(cfg.UserDataPath.Value, RewardsTreesFolder, fmt.Sprintf(MinipoolPerformanceFilenameFormat, string(cfg.Network.Value), interval))
}

func (cfg *SmartNodeConfig) GetNodeAddressPath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserAddressFilename)
}

func (cfg *SmartNodeConfig) GetWalletPath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserWalletDataFilename)
}

func (cfg *SmartNodeConfig) GetPasswordPath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserPasswordFilename)
}

// Check a port setting to see if it's already used elsewhere
func addAndCheckForDuplicate(portMap map[uint16]bool, param config.Parameter[uint16], errors []string) (map[uint16]bool, []string) {
	port := param.Value
	if portMap[port] {
		return portMap, append(errors, fmt.Sprintf("Port %d for %s is already in use", port, param.GetCommon().Name))
	} else {
		portMap[port] = true
	}
	return portMap, errors
}

func (cfg *SmartNodeConfig) GetApiLogFilePath() string {
	return filepath.Join(cfg.rocketPoolDirectory, LogDir, ApiLogName)
}

func (cfg *SmartNodeConfig) GetTasksLogFilePath() string {
	return filepath.Join(cfg.rocketPoolDirectory, LogDir, TasksLogName)
}

func (cfg *SmartNodeConfig) GetWatchtowerLogFilePath() string {
	return filepath.Join(cfg.rocketPoolDirectory, LogDir, WatchtowerLogName)
}

func (cfg *SmartNodeConfig) GetLoggerOptions() log.LoggerOptions {
	return cfg.Logging.GetOptions()
}
