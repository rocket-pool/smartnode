package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/shared"
)

// =========================
// === Smart Node Config ===
// =========================

const (
	// Tags
	smartnodeTag string = "rocketpool/smartnode:v" + shared.RocketPoolVersion
)

// The master configuration struct
type SmartNodeConfig struct {
	// Smart Node settings
	DebugMode                 config.Parameter[bool]
	Network                   config.Parameter[config.Network]
	ClientMode                config.Parameter[config.ClientMode]
	ProjectName               config.Parameter[string]
	UserDataPath              config.Parameter[string]
	WatchtowerStatePath       config.Parameter[string]
	AutoTxMaxFee              config.Parameter[float64]
	MaxPriorityFee            config.Parameter[float64]
	AutoTxGasThreshold        config.Parameter[float64]
	DistributeThreshold       config.Parameter[float64]
	RewardsTreeMode           config.Parameter[RewardsMode]
	RewardsTreeCustomUrl      config.Parameter[string]
	ArchiveECUrl              config.Parameter[string]
	WatchtowerMaxFeeOverride  config.Parameter[float64]
	WatchtowerPrioFeeOverride config.Parameter[float64]
	UseRollingRecords         config.Parameter[bool]
	RecordCheckpointInterval  config.Parameter[uint64]
	CheckpointRetentionLimit  config.Parameter[uint64]
	RecordsPath               config.Parameter[string]
	VerifyProposals           config.Parameter[bool]

	// Execution client settings
	LocalExecutionConfig    *config.LocalExecutionConfig
	ExternalExecutionConfig *config.ExternalExecutionConfig

	// Beacon node settings
	LocalBeaconConfig    *config.LocalBeaconConfig
	ExternalBeaconConfig *config.ExternalBeaconConfig

	// Fallback clients
	Fallback *config.FallbackConfig

	// Metrics
	Metrics *MetricsConfig

	// Native mode
	Native *NativeConfig

	// MEV-Boost
	EnableMevBoost config.Parameter[bool]
	MevBoost       *MevBoostConfig

	// Addons
	Addons map[string]any

	// Internal fields
	Version             string
	RocketPoolDirectory string
	IsNativeMode        bool
}
