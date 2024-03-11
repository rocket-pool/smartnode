package ids

const (
	// Smart Node parameter IDs
	RootConfigID                    string = "smartnode"
	VersionID                       string = "version"
	UserDirectoryKey                string = "rpUserDir"
	IsNativeKey                     string = "isNative"
	DebugModeID                     string = "debugMode"
	NetworkID                       string = "network"
	ClientModeID                    string = "clientMode"
	UserDataPathID                  string = "rpUserDataDir"
	ProjectNameID                   string = "projectName"
	WatchtowerStatePath             string = "watchtowerStatePath"
	AutoTxMaxFeeID                  string = "autoTxMaxFee"
	MaxPriorityFeeID                string = "maxPriorityFee"
	AutoTxGasThresholdID            string = "autoTxGasThreshold"
	DistributeThresholdID           string = "distributeThreshold"
	RewardsTreeModeID               string = "rewardsTreeMode"
	RewardsTreeCustomUrlID          string = "rewardsTreeCustomUrl"
	ArchiveEcUrlID                  string = "archiveEcUrl"
	WatchtowerMaxFeeOverrideID      string = "watchtowerMaxFeeOverride"
	WatchtowerPriorityFeeOverrideID string = "watchtowerPriorityFeeOverride"
	UseRollingRecordsID             string = "useRollingRecords"
	RecordCheckpointIntervalID      string = "recordCheckpointInterval"
	CheckpointRetentionLimitID      string = "checkpointRetentionLimit"
	RecordsPathID                   string = "recordsPath"
	VerifyProposalsID               string = "verifyProposals"

	// Subconfig IDs
	FallbackID          string = "fallback"
	LocalExecutionID    string = "localExecution"
	ExternalExecutionID string = "externalExecution"
	LocalBeaconID       string = "localBeacon"
	ExternalBeaconID    string = "externalBeacon"
	ValidatorClientID   string = "validator"
	MetricsID           string = "metrics"
	NativeID            string = "native"
	MevBoostID          string = "mevBoost"

	// Metrics
	MetricsWatchtowerPortID string = "watchtowerMetricsPort"
	MetricsEnableOdaoID     string = "enableODaoMetrics"

	// MEV-Boost
	MevBoostEnableID        string = "enableMevBoost"
	MevBoostModeID          string = "mode"
	MevBoostSelectionModeID string = "selectionMode"
	MevBoostPortID          string = "port"
	MevBoostOpenRpcPortID   string = "openRpcPort"
	MevBoostExternalUrlID   string = "externalUrl"

	// Native
	NativeValidatorRestartCommandID string = "validatorRestartCommand"
	NativeValidatorStopCommandID    string = "validatorStopCommand"

	// VC Subconfigs
	VcCommonID   string = "common"
	LighthouseID string = "lighthouse"
	LodestarID   string = "lodestar"
	NimbusID     string = "nimbus"
	PrysmID      string = "prysm"
	TekuID       string = "teku"
)
