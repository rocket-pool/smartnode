package ids

const (
	// Root IDs
	VersionID   string = "version"
	IsNativeKey string = "isNative"
	SmartNodeID string = "smartNode"

	// Smart Node parameter IDs
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
	LoggingID           string = "logging"
	FallbackID          string = "fallback"
	LocalExecutionID    string = "localExecution"
	ExternalExecutionID string = "externalExecution"
	LocalBeaconID       string = "localBeacon"
	ExternalBeaconID    string = "externalBeacon"
	ValidatorClientID   string = "validator"
	MetricsID           string = "metrics"
	AlertmanagerID      string = "alertmanager"
	MevBoostID          string = "mevBoost"
	AddonsID            string = "addons"

	// Metrics
	MetricsEnableOdaoID string = "enableODaoMetrics"

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

	// Alertmanager
	AlertmanagerEnableAlertingID              string = "enableAlerting"
	AlertmanagerNativeModeHostID              string = "nativeModeHost"
	AlertmanagerNativeModePortID              string = "nativeModePort"
	AlertmanagerDiscordWebhookUrlID           string = "discordWebhookURL"
	AlertmanagerClientSyncStatusBeaconID      string = "alertEnabled_ClientSyncStatusBeacon"
	AlertmanagerClientSyncStatusExecutionID   string = "alertEnabled_ClientSyncStatusExecution"
	AlertmanagerUpcomingSyncCommitteeID       string = "alertEnabled_UpcomingSyncCommittee"
	AlertmanagerActiveSyncCommitteeID         string = "alertEnabled_ActiveSyncCommittee"
	AlertmanagerUpcomingProposalID            string = "alertEnabled_UpcomingProposal"
	AlertmanagerRecentProposalID              string = "alertEnabled_RecentProposal"
	AlertmanagerLowDiskSpaceWarningID         string = "alertEnabled_LowDiskSpaceWarning"
	AlertmanagerLowDiskSpaceCriticalID        string = "alertEnabled_LowDiskSpaceCritical"
	AlertmanagerOSUpdatesAvailableID          string = "alertEnabled_OSUpdatesAvailable"
	AlertmanagerRPUpdatesAvailableID          string = "alertEnabled_RPUpdatesAvailable"
	AlertmanagerFeeRecipientChangedID         string = "alertEnabled_FeeRecipientChanged"
	AlertmanagerMinipoolBondReducedID         string = "alertEnabled_MinipoolBondReduced"
	AlertmanagerMinipoolBalanceDistributedID  string = "alertEnabled_MinipoolBalanceDistributed"
	AlertmanagerMinipoolPromotedID            string = "alertEnabled_MinipoolPromoted"
	AlertmanagerMinipoolStakedID              string = "alertEnabled_MinipoolStaked"
	AlertmanagerExecutionClientSyncCompleteID string = "alertEnabled_ExecutionClientSyncComplete"
	AlertmanagerBeaconClientSyncCompleteID    string = "alertEnabled_BeaconClientSyncComplete"
)
