package migration

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/go-homedir"
	nmc_ids "github.com/rocket-pool/node-manager-core/config/ids"
	gww_ids "github.com/rocket-pool/smartnode/v2/addons/graffiti_wall_writer/ids"
	rn_ids "github.com/rocket-pool/smartnode/v2/addons/rescue_node/ids"
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

// Migrate a legacy v1 config into a new v2 config
func upgradeFromV1(oldConfig map[string]any) (map[string]any, error) {
	// Legacy config sections
	legacyRootConfig, err := getLegacyConfigSection(nil, oldConfig, "root")
	legacySmartnodeConfig, err := getLegacyConfigSection(err, oldConfig, "smartnode")
	legacyExecutionCommonConfig, err := getLegacyConfigSection(err, oldConfig, "executionCommon")
	legacyGethConfig, err := getLegacyConfigSection(err, oldConfig, "geth")
	legacyNethermindConfig, err := getLegacyConfigSection(err, oldConfig, "nethermind")
	legacyBesuConfig, err := getLegacyConfigSection(err, oldConfig, "besu")
	legacyRethConfig, err := getLegacyConfigSection(err, oldConfig, "reth")
	legacyExternalExecutionConfig, err := getLegacyConfigSection(err, oldConfig, "externalExecution")
	legacyConsensusCommonConfig, err := getLegacyConfigSection(err, oldConfig, "consensusCommon")
	legacyLighthouseConfig, err := getLegacyConfigSection(err, oldConfig, "lighthouse")
	legacyLodestarConfig, err := getLegacyConfigSection(err, oldConfig, "lodestar")
	legacyNimbusConfig, err := getLegacyConfigSection(err, oldConfig, "nimbus")
	legacyPrysmConfig, err := getLegacyConfigSection(err, oldConfig, "prysm")
	legacyTekuConfig, err := getLegacyConfigSection(err, oldConfig, "teku")
	legacyExternalLighthouseConfig, err := getLegacyConfigSection(err, oldConfig, "externalLighthouse")
	legacyExternalLodestarConfig, err := getLegacyConfigSection(err, oldConfig, "externalLodestar")
	legacyExternalNimbusConfig, err := getLegacyConfigSection(err, oldConfig, "externalNimbus")
	legacyExternalPrysmConfig, err := getLegacyConfigSection(err, oldConfig, "externalPrysm")
	legacyExternalTekuConfig, err := getLegacyConfigSection(err, oldConfig, "externalTeku")
	legacyFallbackNormalConfig, err := getLegacyConfigSection(err, oldConfig, "fallbackNormal")
	legacyFallbackPrysmConfig, err := getLegacyConfigSection(err, oldConfig, "fallbackPrysm")
	legacyGrafanaConfig, err := getLegacyConfigSection(err, oldConfig, "grafana")
	legacyPrometheusConfig, err := getLegacyConfigSection(err, oldConfig, "prometheus")
	legacyAlertmanagerConfig, err := getLegacyConfigSection(err, oldConfig, "alertmanager")
	legacyExporterConfig, err := getLegacyConfigSection(err, oldConfig, "exporter")
	legacybBitflyNodeMetricsConfig, err := getLegacyConfigSection(err, oldConfig, "bitflyNodeMetrics")
	legacyNativeConfig, err := getLegacyConfigSection(err, oldConfig, "native")
	legacyMevBoostConfig, err := getLegacyConfigSection(err, oldConfig, "mevBoost")
	legacyGwwConfig, err := getLegacyConfigSection(err, oldConfig, "addons-gww")
	legacyRescueNodeConfig, err := getLegacyConfigSection(err, oldConfig, "addons-rescue-node")
	if err != nil {
		return nil, err
	}

	// Top level
	newConfig := map[string]any{}
	newConfig[ids.IsNativeKey] = legacyRootConfig["isNative"]
	newConfig[ids.VersionID] = "v2.0.0-migrate"

	// Smart Node
	newSmartnodeConfig := map[string]any{}
	newSmartnodeConfig[ids.ProjectNameID] = legacySmartnodeConfig["projectName"]
	if err = getSettingWithExpandedPath(newSmartnodeConfig, ids.UserDataPathID, legacySmartnodeConfig, "dataPath"); err != nil {
		return nil, err
	}
	newSmartnodeConfig[ids.NetworkID] = legacySmartnodeConfig["network"]
	newSmartnodeConfig[ids.ClientModeID] = legacyRootConfig["executionClientMode"]
	newSmartnodeConfig[ids.VerifyProposalsID] = legacySmartnodeConfig["verifyProposals"]
	newSmartnodeConfig[ids.AutoTxMaxFeeID] = legacySmartnodeConfig["manualMaxFee"]
	newSmartnodeConfig[ids.MaxPriorityFeeID] = legacySmartnodeConfig["priorityFee"]
	newSmartnodeConfig[ids.AutoTxGasThresholdID] = legacySmartnodeConfig["minipoolStakeGasThreshold"]
	newSmartnodeConfig[ids.DistributeThresholdID] = legacySmartnodeConfig["distributeThreshold"]
	newSmartnodeConfig[ids.RewardsTreeModeID] = legacySmartnodeConfig["rewardsTreeMode"]
	newSmartnodeConfig[ids.RewardsTreeCustomUrlID] = legacySmartnodeConfig["rewardsTreeCustomUrl"]
	newSmartnodeConfig[ids.WatchtowerMaxFeeOverrideID] = legacySmartnodeConfig["watchtowerMaxFeeOverride"]
	newSmartnodeConfig[ids.WatchtowerPriorityFeeOverrideID] = legacySmartnodeConfig["watchtowerPrioFeeOverride"]
	newSmartnodeConfig[ids.ArchiveEcUrlID] = legacySmartnodeConfig["archiveECUrl"]
	newSmartnodeConfig[ids.UseRollingRecordsID] = legacySmartnodeConfig["useRollingRecords"]
	newSmartnodeConfig[ids.RecordCheckpointIntervalID] = legacySmartnodeConfig["recordCheckpointInterval"]
	newSmartnodeConfig[ids.CheckpointRetentionLimitID] = legacySmartnodeConfig["checkpointRetentionLimit"]
	newSmartnodeConfig[ids.WatchtowerStatePath] = legacySmartnodeConfig["watchtowerPath"]
	newSmartnodeConfig[ids.RecordsPathID] = legacySmartnodeConfig["recordsPath"]
	newConfig[ids.SmartNodeID] = newSmartnodeConfig

	// Local execution
	newLocalExecutionConfig := map[string]any{}
	newLocalExecutionConfig[nmc_ids.EcID] = legacyRootConfig["executionClient"]
	newLocalExecutionConfig[nmc_ids.HttpPortID] = legacyExecutionCommonConfig["httpPort"]
	newLocalExecutionConfig[nmc_ids.LocalEcWebsocketPortID] = legacyExecutionCommonConfig["wsPort"]
	newLocalExecutionConfig[nmc_ids.LocalEcEnginePortID] = legacyExecutionCommonConfig["enginePort"]
	newLocalExecutionConfig[nmc_ids.LocalEcOpenApiPortsID] = legacyExecutionCommonConfig["openRpcPorts"]
	newLocalExecutionConfig[nmc_ids.P2pPortID] = legacyExecutionCommonConfig["p2pPort"]
	newSmartnodeConfig[ids.LocalExecutionID] = newLocalExecutionConfig

	// Geth
	newGethConfig := map[string]any{}
	newGethConfig[nmc_ids.GethEnablePbssID] = legacyGethConfig["enablePbss"]
	newGethConfig[nmc_ids.MaxPeersID] = legacyGethConfig["maxPeers"]
	newGethConfig[nmc_ids.ContainerTagID] = legacyGethConfig["containerTag"]
	newGethConfig[nmc_ids.AdditionalFlagsID] = legacyGethConfig["additionalFlags"]
	newLocalExecutionConfig[nmc_ids.LocalEcGethID] = newGethConfig

	// Nethermind
	newNethermindConfig := map[string]any{}
	newNethermindConfig[nmc_ids.CacheSizeID] = legacyNethermindConfig["cache"]
	newNethermindConfig[nmc_ids.MaxPeersID] = legacyNethermindConfig["maxPeers"]
	newNethermindConfig[nmc_ids.NethermindPruneMemSizeID] = legacyNethermindConfig["pruneMemSize"]
	newNethermindConfig[nmc_ids.NethermindAdditionalModulesID] = legacyNethermindConfig["additionalModules"]
	newNethermindConfig[nmc_ids.NethermindAdditionalUrlsID] = legacyNethermindConfig["additionalUrls"]
	newNethermindConfig[nmc_ids.ContainerTagID] = legacyNethermindConfig["containerTag"]
	newNethermindConfig[nmc_ids.AdditionalFlagsID] = legacyNethermindConfig["additionalFlags"]
	newLocalExecutionConfig[nmc_ids.LocalEcNethermindID] = newNethermindConfig

	// Besu
	newBesuConfig := map[string]any{}
	newBesuConfig[nmc_ids.BesuJvmHeapSizeID] = legacyBesuConfig["jvmHeapSize"]
	newBesuConfig[nmc_ids.MaxPeersID] = legacyBesuConfig["maxPeers"]
	newBesuConfig[nmc_ids.BesuMaxBackLayersID] = legacyBesuConfig["maxBackLayers"]
	newBesuConfig[nmc_ids.ContainerTagID] = legacyBesuConfig["containerTag"]
	newBesuConfig[nmc_ids.AdditionalFlagsID] = legacyBesuConfig["additionalFlags"]
	newLocalExecutionConfig[nmc_ids.LocalEcBesuID] = newBesuConfig

	// Reth
	newRethConfig := map[string]any{}
	newRethConfig[nmc_ids.CacheSizeID] = legacyRethConfig["cache"]
	newRethConfig[nmc_ids.MaxPeersID] = legacyRethConfig["maxPeers"]
	newRethConfig[nmc_ids.ContainerTagID] = legacyRethConfig["containerTag"]
	newRethConfig[nmc_ids.AdditionalFlagsID] = legacyRethConfig["additionalFlags"]
	newLocalExecutionConfig[nmc_ids.LocalEcRethID] = newRethConfig

	// External execution
	newExternalExecutionConfig := map[string]any{}
	newExternalExecutionConfig[nmc_ids.EcID] = "" // Smartnode v1 didn't have this unfortunately
	isNative, _ := strconv.ParseBool(legacyRootConfig["isNative"])
	if isNative {
		newExternalExecutionConfig[nmc_ids.HttpUrlID] = legacyNativeConfig["ecHttpUrl"]
	} else {
		newExternalExecutionConfig[nmc_ids.HttpUrlID] = legacyExternalExecutionConfig["httpUrl"]
		newExternalExecutionConfig[nmc_ids.ExternalEcWebsocketUrlID] = legacyExternalExecutionConfig["wsUrl"]
	}
	newSmartnodeConfig[ids.ExternalExecutionID] = newExternalExecutionConfig

	// Local beacon
	newLocalBeaconConfig := map[string]any{}
	newLocalBeaconConfig[nmc_ids.BnID] = legacyRootConfig["consensusClient"]
	newLocalBeaconConfig[nmc_ids.LocalBnCheckpointSyncUrlID] = legacyConsensusCommonConfig["checkpointSyncUrl"]
	newLocalBeaconConfig[nmc_ids.P2pPortID] = legacyConsensusCommonConfig["p2pPort"]
	newLocalBeaconConfig[nmc_ids.HttpPortID] = legacyConsensusCommonConfig["apiPort"]
	newLocalBeaconConfig[nmc_ids.OpenHttpPortsID] = legacyConsensusCommonConfig["openApiPort"]
	newSmartnodeConfig[ids.LocalBeaconID] = newLocalBeaconConfig

	// Lighthouse BN
	newLighthouseBnConfig := map[string]any{}
	newLighthouseBnConfig[nmc_ids.LighthouseQuicPortID] = legacyLighthouseConfig["p2pQuicPort"]
	newLighthouseBnConfig[nmc_ids.MaxPeersID] = legacyLighthouseConfig["maxPeers"]
	newLighthouseBnConfig[nmc_ids.ContainerTagID] = legacyLighthouseConfig["containerTag"]
	newLighthouseBnConfig[nmc_ids.AdditionalFlagsID] = legacyLighthouseConfig["additionalBnFlags"]
	newLocalBeaconConfig[nmc_ids.LocalBnLighthouseID] = newLighthouseBnConfig

	// Lodestar BN
	newLodestarBnConfig := map[string]any{}
	newLodestarBnConfig[nmc_ids.MaxPeersID] = legacyLodestarConfig["maxPeers"]
	newLodestarBnConfig[nmc_ids.ContainerTagID] = legacyLodestarConfig["containerTag"]
	newLodestarBnConfig[nmc_ids.AdditionalFlagsID] = legacyLodestarConfig["additionalBnFlags"]
	newLocalBeaconConfig[nmc_ids.LocalBnLodestarID] = newLodestarBnConfig

	// Nimbus BN
	newNimbusBnConfig := map[string]any{}
	newNimbusBnConfig[nmc_ids.MaxPeersID] = legacyNimbusConfig["maxPeers"]
	newNimbusBnConfig[nmc_ids.NimbusPruningModeID] = legacyNimbusConfig["pruningMode"]
	newNimbusBnConfig[nmc_ids.ContainerTagID] = legacyNimbusConfig["bnContainerTag"]
	newNimbusBnConfig[nmc_ids.AdditionalFlagsID] = legacyNimbusConfig["additionalBnFlags"]
	newLocalBeaconConfig[nmc_ids.LocalBnNimbusID] = newNimbusBnConfig

	// Prysm BN
	newPrysmBnConfig := map[string]any{}
	newPrysmBnConfig[nmc_ids.MaxPeersID] = legacyPrysmConfig["maxPeers"]
	newPrysmBnConfig[nmc_ids.PrysmRpcPortID] = legacyPrysmConfig["rpcPort"]
	newPrysmBnConfig[nmc_ids.PrysmOpenRpcPortID] = legacyPrysmConfig["openRpcPort"]
	newPrysmBnConfig[nmc_ids.ContainerTagID] = legacyPrysmConfig["bnContainerTag"]
	newPrysmBnConfig[nmc_ids.AdditionalFlagsID] = legacyPrysmConfig["additionalBnFlags"]
	newLocalBeaconConfig[nmc_ids.LocalBnPrysmID] = newPrysmBnConfig

	// Teku BN
	newTekuBnConfig := map[string]any{}
	newTekuBnConfig[nmc_ids.TekuJvmHeapSizeID] = legacyTekuConfig["jvmHeapSize"]
	newTekuBnConfig[nmc_ids.MaxPeersID] = legacyTekuConfig["maxPeers"]
	newTekuBnConfig[nmc_ids.TekuArchiveModeID] = legacyTekuConfig["archiveMode"]
	newTekuBnConfig[nmc_ids.ContainerTagID] = legacyTekuConfig["containerTag"]
	newTekuBnConfig[nmc_ids.AdditionalFlagsID] = legacyTekuConfig["additionalBnFlags"]
	newLocalBeaconConfig[nmc_ids.LocalBnTekuID] = newTekuBnConfig

	// External beacon
	newExternalBeaconConfig := map[string]any{}
	if isNative {
		newExternalBeaconConfig[nmc_ids.BnID] = legacyNativeConfig["consensusClient"]
		newExternalBeaconConfig[nmc_ids.HttpUrlID] = legacyNativeConfig["ccHttpUrl"]
	} else {
		newExternalBeaconConfig[nmc_ids.BnID] = legacyRootConfig["externalConsensusClient"]
		switch newExternalBeaconConfig[nmc_ids.BnID] {
		case "lighthouse":
			newExternalBeaconConfig[nmc_ids.HttpUrlID] = legacyExternalLighthouseConfig["httpUrl"]
		case "lodestar":
			newExternalBeaconConfig[nmc_ids.HttpUrlID] = legacyExternalLodestarConfig["httpUrl"]
		case "nimbus":
			newExternalBeaconConfig[nmc_ids.HttpUrlID] = legacyExternalNimbusConfig["httpUrl"]
		case "prysm":
			newExternalBeaconConfig[nmc_ids.HttpUrlID] = legacyExternalPrysmConfig["httpUrl"]
			newExternalBeaconConfig[nmc_ids.PrysmRpcUrlID] = legacyExternalPrysmConfig["jsonRpcUrl"]
		case "teku":
			newExternalBeaconConfig[nmc_ids.HttpUrlID] = legacyExternalTekuConfig["httpUrl"]
		}
	}
	newSmartnodeConfig[ids.ExternalBeaconID] = newExternalBeaconConfig

	// Validator Client
	newValidatorClientConfig := map[string]any{}
	if err = getSettingWithExpandedPath(newValidatorClientConfig, ids.NativeValidatorRestartCommandID, legacyNativeConfig, "validatorRestartCommand"); err != nil {
		return nil, err
	}
	if err = getSettingWithExpandedPath(newValidatorClientConfig, ids.NativeValidatorStopCommandID, legacyNativeConfig, "validatorStopCommand"); err != nil {
		return nil, err
	}
	newSmartnodeConfig[ids.ValidatorClientID] = newValidatorClientConfig

	// Get the VC details based on the old client mode
	newValidatorCommonConfig := map[string]any{}
	newValidatorCommonConfig[nmc_ids.MetricsPortID] = legacyRootConfig["vcMetricsPort"]
	newLighthouseVcConfig := map[string]any{}
	newLodestarVcConfig := map[string]any{}
	newNimbusVcConfig := map[string]any{}
	newPrysmVcConfig := map[string]any{}
	newTekuVcConfig := map[string]any{}
	newValidatorClientConfig[ids.VcCommonID] = newValidatorCommonConfig
	newValidatorClientConfig[ids.LighthouseID] = newLighthouseVcConfig
	newValidatorClientConfig[ids.LodestarID] = newLodestarVcConfig
	newValidatorClientConfig[ids.NimbusID] = newNimbusVcConfig
	newValidatorClientConfig[ids.PrysmID] = newPrysmVcConfig
	newValidatorClientConfig[ids.TekuID] = newTekuVcConfig
	switch legacyRootConfig["consensusClientMode"] {
	case "local":
		// VC Common
		newValidatorCommonConfig[nmc_ids.GraffitiID] = legacyConsensusCommonConfig["graffiti"]
		newValidatorCommonConfig[nmc_ids.DoppelgangerDetectionID] = legacyConsensusCommonConfig["doppelgangerDetection"]

		// Lighthouse
		newLighthouseVcConfig[nmc_ids.ContainerTagID] = legacyLighthouseConfig["containerTag"]
		newLighthouseVcConfig[nmc_ids.AdditionalFlagsID] = legacyLighthouseConfig["additionalVcFlags"]

		// Lodestar
		newLodestarVcConfig[nmc_ids.ContainerTagID] = legacyLodestarConfig["containerTag"]
		newLodestarVcConfig[nmc_ids.AdditionalFlagsID] = legacyLodestarConfig["additionalVcFlags"]

		// Nimbus
		newNimbusVcConfig[nmc_ids.ContainerTagID] = legacyNimbusConfig["containerTag"]
		newNimbusVcConfig[nmc_ids.AdditionalFlagsID] = legacyNimbusConfig["additionalVcFlags"]

		// Prysm
		newPrysmVcConfig[nmc_ids.ContainerTagID] = legacyPrysmConfig["vcContainerTag"]
		newPrysmVcConfig[nmc_ids.AdditionalFlagsID] = legacyPrysmConfig["additionalVcFlags"]

		// Teku
		newTekuVcConfig[nmc_ids.ContainerTagID] = legacyTekuConfig["containerTag"]
		newTekuVcConfig[nmc_ids.AdditionalFlagsID] = legacyTekuConfig["additionalVcFlags"]

	case "external":
		// VC Common
		switch newExternalBeaconConfig[nmc_ids.BnID] {
		case "lighthouse":
			newValidatorCommonConfig[nmc_ids.GraffitiID] = legacyExternalLighthouseConfig["graffiti"]
			newValidatorCommonConfig[nmc_ids.DoppelgangerDetectionID] = legacyExternalLighthouseConfig["doppelgangerDetection"]
		case "lodestar":
			newValidatorCommonConfig[nmc_ids.GraffitiID] = legacyExternalLodestarConfig["graffiti"]
			newValidatorCommonConfig[nmc_ids.DoppelgangerDetectionID] = legacyExternalLodestarConfig["doppelgangerDetection"]
		case "nimbus":
			newValidatorCommonConfig[nmc_ids.GraffitiID] = legacyExternalNimbusConfig["graffiti"]
			newValidatorCommonConfig[nmc_ids.DoppelgangerDetectionID] = legacyExternalNimbusConfig["doppelgangerDetection"]
		case "prysm":
			newValidatorCommonConfig[nmc_ids.GraffitiID] = legacyExternalPrysmConfig["graffiti"]
			newValidatorCommonConfig[nmc_ids.DoppelgangerDetectionID] = legacyExternalPrysmConfig["doppelgangerDetection"]
		case "teku":
			newValidatorCommonConfig[nmc_ids.GraffitiID] = legacyExternalTekuConfig["graffiti"]
			newValidatorCommonConfig[nmc_ids.DoppelgangerDetectionID] = legacyExternalTekuConfig["doppelgangerDetection"]
		}

		// Lighthouse
		newLighthouseVcConfig[nmc_ids.ContainerTagID] = legacyExternalLighthouseConfig["containerTag"]
		newLighthouseVcConfig[nmc_ids.AdditionalFlagsID] = legacyExternalLighthouseConfig["additionalVcFlags"]

		// Lodestar
		newLodestarVcConfig[nmc_ids.ContainerTagID] = legacyExternalLodestarConfig["containerTag"]
		newLodestarVcConfig[nmc_ids.AdditionalFlagsID] = legacyExternalLodestarConfig["additionalVcFlags"]

		// Nimbus
		newNimbusVcConfig[nmc_ids.ContainerTagID] = legacyExternalNimbusConfig["containerTag"]
		newNimbusVcConfig[nmc_ids.AdditionalFlagsID] = legacyExternalNimbusConfig["additionalVcFlags"]

		// Prysm
		newPrysmVcConfig[nmc_ids.ContainerTagID] = legacyExternalPrysmConfig["containerTag"]
		newPrysmVcConfig[nmc_ids.AdditionalFlagsID] = legacyExternalPrysmConfig["additionalVcFlags"]

		// Teku
		newTekuVcConfig[nmc_ids.ContainerTagID] = legacyExternalTekuConfig["containerTag"]
		newTekuVcConfig[nmc_ids.AdditionalFlagsID] = legacyExternalTekuConfig["additionalVcFlags"]
	}

	// Fallback
	newFallbackConfig := map[string]any{}
	newFallbackConfig[nmc_ids.FallbackUseFallbackClientsID] = legacyRootConfig["useFallbackClients"]
	if (legacyRootConfig["consensusClientMode"] == "local" && legacyRootConfig["consensusClient"] == "prysm") || (legacyRootConfig["consensusClientMode"] == "external" && legacyRootConfig["externalConsensusClient"] == "prysm") {
		newFallbackConfig[nmc_ids.FallbackEcHttpUrlID] = legacyFallbackPrysmConfig["ecHttpUrl"]
		newFallbackConfig[nmc_ids.FallbackBnHttpUrlID] = legacyFallbackPrysmConfig["ccHttpUrl"]
		newFallbackConfig[nmc_ids.PrysmRpcUrlID] = legacyFallbackPrysmConfig["jsonRpcUrl"]
	} else {
		newFallbackConfig[nmc_ids.FallbackEcHttpUrlID] = legacyFallbackNormalConfig["ecHttpUrl"]
		newFallbackConfig[nmc_ids.FallbackBnHttpUrlID] = legacyFallbackNormalConfig["ccHttpUrl"]
	}
	newSmartnodeConfig[ids.FallbackID] = newFallbackConfig

	// Metrics
	newMetricsConfig := map[string]any{}
	newMetricsConfig[nmc_ids.MetricsEnableID] = legacyRootConfig["enableMetrics"]
	newMetricsConfig[nmc_ids.MetricsEnableBitflyID] = legacyRootConfig["enableBitflyNodeMetrics"]
	newMetricsConfig[nmc_ids.MetricsEcPortID] = legacyRootConfig["ecMetricsPort"]
	newMetricsConfig[nmc_ids.MetricsBnPortID] = legacyRootConfig["bnMetricsPort"]
	newMetricsConfig[nmc_ids.MetricsDaemonPortID] = legacyRootConfig["nodeMetricsPort"]
	newMetricsConfig[nmc_ids.MetricsExporterPortID] = legacyRootConfig["exporterMetricsPort"]
	newMetricsConfig[ids.MetricsEnableOdaoID] = legacyRootConfig["enableODaoMetrics"]
	newSmartnodeConfig[ids.MetricsID] = newMetricsConfig

	// Grafana
	newGrafanaConfig := map[string]any{}
	newGrafanaConfig[nmc_ids.PortID] = legacyGrafanaConfig["port"]
	newGrafanaConfig[nmc_ids.ContainerTagID] = legacyGrafanaConfig["containerTag"]
	newMetricsConfig[nmc_ids.MetricsGrafanaID] = newGrafanaConfig

	// Prometheus
	newPrometheusConfig := map[string]any{}
	newPrometheusConfig[nmc_ids.PortID] = legacyPrometheusConfig["port"]
	newPrometheusConfig[nmc_ids.OpenPortID] = legacyPrometheusConfig["openPort"]
	newPrometheusConfig[nmc_ids.ContainerTagID] = legacyPrometheusConfig["containerTag"]
	newPrometheusConfig[nmc_ids.AdditionalFlagsID] = legacyPrometheusConfig["additionalFlags"]
	newMetricsConfig[nmc_ids.MetricsPrometheusID] = newPrometheusConfig

	// Alertmanager
	newAlertmanagerConfig := map[string]any{}
	newAlertmanagerConfig[ids.AlertmanagerEnableAlertingID] = legacyAlertmanagerConfig["enableAlerting"]
	newAlertmanagerConfig[nmc_ids.PortID] = legacyAlertmanagerConfig["port"]
	newAlertmanagerConfig[ids.AlertmanagerNativeModeHostID] = legacyAlertmanagerConfig["nativeModeHost"]
	newAlertmanagerConfig[ids.AlertmanagerNativeModePortID] = legacyAlertmanagerConfig["nativeModePort"]
	newAlertmanagerConfig[nmc_ids.OpenPortID] = legacyAlertmanagerConfig["openPort"]
	newAlertmanagerConfig[nmc_ids.ContainerTagID] = legacyAlertmanagerConfig["containerTag"]
	newAlertmanagerConfig[ids.AlertmanagerDiscordWebhookUrlID] = legacyAlertmanagerConfig["discordWebhookURL"]
	newAlertmanagerConfig[ids.AlertmanagerClientSyncStatusBeaconID] = legacyAlertmanagerConfig["alertEnabled_ClientSyncStatusBeacon"]
	newAlertmanagerConfig[ids.AlertmanagerClientSyncStatusExecutionID] = legacyAlertmanagerConfig["alertEnabled_ClientSyncStatusExecution"]
	newAlertmanagerConfig[ids.AlertmanagerUpcomingSyncCommitteeID] = legacyAlertmanagerConfig["alertEnabled_UpcomingSyncCommittee"]
	newAlertmanagerConfig[ids.AlertmanagerActiveSyncCommitteeID] = legacyAlertmanagerConfig["alertEnabled_ActiveSyncCommittee"]
	newAlertmanagerConfig[ids.AlertmanagerUpcomingProposalID] = legacyAlertmanagerConfig["alertEnabled_UpcomingProposal"]
	newAlertmanagerConfig[ids.AlertmanagerRecentProposalID] = legacyAlertmanagerConfig["alertEnabled_RecentProposal"]
	newAlertmanagerConfig[ids.AlertmanagerLowDiskSpaceWarningID] = legacyAlertmanagerConfig["alertEnabled_LowDiskSpaceWarning"]
	newAlertmanagerConfig[ids.AlertmanagerLowDiskSpaceCriticalID] = legacyAlertmanagerConfig["alertEnabled_LowDiskSpaceCritical"]
	newAlertmanagerConfig[ids.AlertmanagerOSUpdatesAvailableID] = legacyAlertmanagerConfig["alertEnabled_OSUpdatesAvailable"]
	newAlertmanagerConfig[ids.AlertmanagerRPUpdatesAvailableID] = legacyAlertmanagerConfig["alertEnabled_RPUpdatesAvailable"]
	newAlertmanagerConfig[ids.AlertmanagerFeeRecipientChangedID] = legacyAlertmanagerConfig["alertEnabled_FeeRecipientChanged"]
	newAlertmanagerConfig[ids.AlertmanagerMinipoolBondReducedID] = legacyAlertmanagerConfig["alertEnabled_MinipoolBondReduced"]
	newAlertmanagerConfig[ids.AlertmanagerMinipoolBalanceDistributedID] = legacyAlertmanagerConfig["alertEnabled_MinipoolBalanceDistributed"]
	newAlertmanagerConfig[ids.AlertmanagerMinipoolPromotedID] = legacyAlertmanagerConfig["alertEnabled_MinipoolPromoted"]
	newAlertmanagerConfig[ids.AlertmanagerMinipoolStakedID] = legacyAlertmanagerConfig["alertEnabled_MinipoolStaked"]
	newAlertmanagerConfig[ids.AlertmanagerExecutionClientSyncCompleteID] = legacyAlertmanagerConfig["alertEnabled_ExecutionClientSyncComplete"]
	newAlertmanagerConfig[ids.AlertmanagerBeaconClientSyncCompleteID] = legacyAlertmanagerConfig["alertEnabled_BeaconClientSyncComplete"]
	newSmartnodeConfig[ids.AlertmanagerID] = newAlertmanagerConfig

	// Exporter
	newExporterConfig := map[string]any{}
	newExporterConfig[nmc_ids.ExporterEnableRootFsID] = legacyExporterConfig["enableRootFs"]
	newExporterConfig[nmc_ids.ContainerTagID] = legacyExporterConfig["containerTag"]
	newExporterConfig[nmc_ids.AdditionalFlagsID] = legacyExporterConfig["additionalFlags"]
	newMetricsConfig[nmc_ids.MetricsExporterID] = newExporterConfig

	// Bitfly
	newBitflyConfig := map[string]any{}
	newBitflyConfig[nmc_ids.BitflySecretID] = legacybBitflyNodeMetricsConfig["bitflySecret"]
	newBitflyConfig[nmc_ids.BitflyEndpointID] = legacybBitflyNodeMetricsConfig["bitflyEndpoint"]
	newBitflyConfig[nmc_ids.BitflyMachineNameID] = legacybBitflyNodeMetricsConfig["bitflyMachineName"]
	newMetricsConfig[nmc_ids.MetricsBitflyID] = newBitflyConfig

	// MEV-Boost
	newMevBoostConfig := map[string]any{}
	newMevBoostConfig[ids.MevBoostEnableID] = legacyRootConfig["enableMevBoost"]
	newMevBoostConfig[ids.MevBoostModeID] = legacyMevBoostConfig["mode"]
	newMevBoostConfig[ids.MevBoostSelectionModeID] = legacyMevBoostConfig["selectionMode"]
	newMevBoostConfig[ids.MevBoostEnableRegulatedAllID] = legacyMevBoostConfig["enableRegulatedAllMev"]
	newMevBoostConfig[ids.MevBoostEnableUnregulatedAllID] = legacyMevBoostConfig["enableUnregulatedAllMev"]
	newMevBoostConfig[ids.MevBoostFlashbotsID] = legacyMevBoostConfig["flashbotsEnabled"]
	newMevBoostConfig[ids.MevBoostBloxRouteMaxProfitID] = legacyMevBoostConfig["bloxRouteMaxProfitEnabled"]
	newMevBoostConfig[ids.MevBoostBloxRouteRegulatedID] = legacyMevBoostConfig["bloxRouteRegulatedEnabled"]
	newMevBoostConfig[ids.MevBoostEdenID] = legacyMevBoostConfig["edenEnabled"]
	newMevBoostConfig[ids.MevBoostUltrasoundID] = legacyMevBoostConfig["ultrasoundEnabled"]
	newMevBoostConfig[ids.MevBoostAestusID] = legacyMevBoostConfig["aestusEnabled"]
	newMevBoostConfig[nmc_ids.PortID] = legacyMevBoostConfig["port"]
	newMevBoostConfig[nmc_ids.OpenPortID] = legacyMevBoostConfig["openRpcPort"]
	newMevBoostConfig[nmc_ids.ContainerTagID] = legacyMevBoostConfig["containerTag"]
	newMevBoostConfig[nmc_ids.AdditionalFlagsID] = legacyMevBoostConfig["additionalFlags"]
	newMevBoostConfig[ids.MevBoostExternalUrlID] = legacyMevBoostConfig["externalUrl"]
	newSmartnodeConfig[ids.MevBoostID] = newMevBoostConfig

	// Addons
	newAddonsConfig := map[string]any{}
	newSmartnodeConfig[ids.AddonsID] = newAddonsConfig

	// GWW
	newGwwConfig := map[string]any{}
	newGwwConfig[gww_ids.GwwEnabledID] = legacyGwwConfig["enabled"]
	newGwwConfig[gww_ids.GwwInputUrlID] = legacyGwwConfig["inputUrl"]
	newGwwConfig[gww_ids.GwwUpdateWallTimeID] = legacyGwwConfig["updateWallTime"]
	newGwwConfig[gww_ids.GwwUpdateInputTimeID] = legacyGwwConfig["updateInputTime"]
	newGwwConfig[gww_ids.GwwUpdatePixelTimeID] = legacyGwwConfig["updatePixelTime"]
	newGwwConfig[nmc_ids.ContainerTagID] = legacyGwwConfig["containerTag"]
	newGwwConfig[nmc_ids.AdditionalFlagsID] = legacyGwwConfig["additionalFlags"]
	newAddonsConfig[ids.AddonsGwwID] = newGwwConfig

	// Rescue node
	newRescueNodeConfig := map[string]any{}
	newRescueNodeConfig[rn_ids.RescueNodeEnabledID] = legacyRescueNodeConfig["enabled"]
	newRescueNodeConfig[rn_ids.RescueNodeUsernameID] = legacyRescueNodeConfig["username"]
	newRescueNodeConfig[rn_ids.RescueNodePasswordID] = legacyRescueNodeConfig["password"]
	newAddonsConfig[ids.AddonsRescueNodeID] = newRescueNodeConfig

	return newConfig, nil
}

func getLegacyConfigSection(previousError error, serializedConfig map[string]any, sectionName string) (map[string]string, error) {
	if previousError != nil {
		return nil, previousError
	}

	// Get the existing section
	legacyEntry, exists := serializedConfig[sectionName]
	if !exists {
		return nil, fmt.Errorf("legacy config is missing the [%s] section", sectionName)
	}

	// Convert it to a map
	legacySection, ok := legacyEntry.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("legacy config has a section named [%s] but it is not a map, it's a %s", sectionName, reflect.TypeOf(legacyEntry))
	}

	// Convert each setting into a string
	legacyConfig := map[string]string{}
	for k, v := range legacySection {
		val, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("legacy config has an entry named [%s.%s] but it is not a string, it's a %s", sectionName, k, reflect.TypeOf(v))
		}
		legacyConfig[k] = val
	}
	return legacyConfig, nil
}

func getSettingWithExpandedPath(section map[string]any, id string, legacySection map[string]string, legacyID string) error {
	var err error
	section[id], err = homedir.Expand(legacySection[legacyID])
	if err != nil {
		return fmt.Errorf("error expanding %s [%s]: %w", id, legacySection[legacyID], err)
	}
	return nil
}
