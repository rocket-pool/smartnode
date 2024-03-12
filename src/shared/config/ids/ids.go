package ids

const (
	// Root IDs
	VersionID        string = "version"
	UserDirectoryKey string = "rpUserDir"
	IsNativeKey      string = "isNative"
	SmartNodeID      string = "smartNode"

	// Smart Node parameter IDs
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
	MevBoostID          string = "mevBoost"
	AddonsID            string = "addons"

	// Metrics
	MetricsWatchtowerPortID string = "watchtowerMetricsPort"
	MetricsEnableOdaoID     string = "enableODaoMetrics"

	// MEV-Boost
	MevBoostEnableID               string = "enableMevBoost"
	MevBoostModeID                 string = "mode"
	MevBoostSelectionModeID        string = "selectionMode"
	MevBoostOpenRpcPortID          string = "openRpcPort"
	MevBoostExternalUrlID          string = "externalUrl"
	MevBoostEnableRegulatedAllID   string = "enableRegulatedAllMev"
	MevBoostEnableUnregulatedAllID string = "enableUnregulatedAllMev"
	MevBoostFlashbotsID            string = "flashbotsEnabled"
	MevBoostBloxRouteMaxProfitID   string = "bloxRouteMaxProfitEnabled"
	MevBoostBloxRouteRegulatedID   string = "bloxRouteRegulatedEnabled"
	MevBoostEdenID                 string = "edenEnabled"
	MevBoostUltrasoundID           string = "ultrasoundEnabled"
	MevBoostAestusID               string = "aestusEnabled"

	// Native
	NativeValidatorRestartCommandID string = "nativeValidatorRestartCommand"
	NativeValidatorStopCommandID    string = "nativeValidatorStopCommand"

	// VC Subconfigs
	VcCommonID   string = "common"
	LighthouseID string = "lighthouse"
	LodestarID   string = "lodestar"
	NimbusID     string = "nimbus"
	PrysmID      string = "prysm"
	TekuID       string = "teku"

	// Addons
	AddonsGwwID        string = "gww"
	AddonsRescueNodeID string = "rescueNode"
)
