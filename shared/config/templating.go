package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/rocket-pool/node-manager-core/config"
	gww "github.com/rocket-pool/smartnode/v2/addons/graffiti_wall_writer"
	rn "github.com/rocket-pool/smartnode/v2/addons/rescue_node"
	"github.com/rocket-pool/smartnode/v2/shared"
)

// =================
// === Constants ===
// =================

func (c *SmartNodeConfig) ExecutionClientContainerName() string {
	return string(ExecutionClientSuffix)
}

func (c *SmartNodeConfig) BeaconNodeContainerName() string {
	return string(BeaconNodeSuffix)
}

func (c *SmartNodeConfig) ValidatorClientContainerName() string {
	return string(ValidatorClientSuffix)
}

func (c *SmartNodeConfig) DaemonContainerName() string {
	return string(NodeSuffix)
}

func (c *SmartNodeConfig) ExporterContainerName() string {
	return string(config.ContainerID_Exporter)
}

func (c *SmartNodeConfig) GrafanaContainerName() string {
	return string(config.ContainerID_Grafana)
}

func (c *SmartNodeConfig) PrometheusContainerName() string {
	return string(config.ContainerID_Prometheus)
}

func (c *SmartNodeConfig) MevBoostContainerName() string {
	return string(config.ContainerID_MevBoost)
}

func (c *SmartNodeConfig) AlertmanagerContainerName() string {
	return string(ContainerID_Alertmanager)
}

func (c *SmartNodeConfig) ExecutionClientDataVolume() string {
	return ExecutionClientDataVolume
}

func (c *SmartNodeConfig) BeaconNodeDataVolume() string {
	return BeaconNodeDataVolume
}

func (c *SmartNodeConfig) AlertmanagerDataVolume() string {
	return AlertmanagerDataVolume
}

func (c *SmartNodeConfig) PrometheusDataVolume() string {
	return PrometheusDataVolume
}

// ===============
// === General ===
// ===============

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) IsLocalMode() bool {
	return cfg.ClientMode.Value == config.ClientMode_Local && !cfg.IsNativeMode
}

// Gets the full name of the Docker container or volume with the provided suffix (name minus the project ID prefix)
func (cfg *SmartNodeConfig) GetDockerArtifactName(entity string) string {
	return fmt.Sprintf("%s_%s", cfg.ProjectName.Value, entity)
}

// Gets the name of the Execution Client start script
func (cfg *SmartNodeConfig) GetEcStartScript() string {
	return EcStartScript
}

// Gets the name of the Beacon Node start script
func (cfg *SmartNodeConfig) GetBnStartScript() string {
	return BnStartScript
}

// Gets the name of the Validator Client start script
func (cfg *SmartNodeConfig) GetVcStartScript() string {
	return VcStartScript
}

func (cfg *SmartNodeConfig) BnHttpUrl() (string, error) {
	// Check if Rescue Node is in-use
	bn := cfg.GetSelectedBeaconNode()
	rescueNode := rn.NewRescueNode(cfg.Addons.RescueNode)
	overrides, err := rescueNode.GetOverrides(bn)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}
	if overrides != nil {
		// Use the rescue node
		return overrides.BnApiEndpoint, nil
	}
	return cfg.GetBnHttpEndpoint(), nil
}

func (cfg *SmartNodeConfig) BnRpcUrl() (string, error) {
	// Check if Rescue Node is in-use
	bn := cfg.GetSelectedBeaconNode()
	if bn != config.BeaconNode_Prysm {
		return "", nil
	}

	rescueNode := rn.NewRescueNode(cfg.Addons.RescueNode)
	overrides, err := rescueNode.GetOverrides(bn)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}
	if overrides != nil {
		// Use the rescue node
		return overrides.CcRpcEndpoint, nil
	}
	if cfg.IsLocalMode() {
		return fmt.Sprintf("%s:%d", BeaconNodeSuffix, cfg.LocalBeaconClient.Prysm.RpcPort.Value), nil
	}
	return cfg.ExternalBeaconClient.PrysmRpcUrl.Value, nil
}

func (cfg *SmartNodeConfig) FallbackBnHttpUrl() string {
	if !cfg.Fallback.UseFallbackClients.Value {
		return ""
	}
	return cfg.Fallback.BnHttpUrl.Value
}

func (cfg *SmartNodeConfig) FallbackBnRpcUrl() string {
	if !cfg.Fallback.UseFallbackClients.Value {
		return ""
	}
	return cfg.Fallback.PrysmRpcUrl.Value
}

func (cfg *SmartNodeConfig) AutoTxMaxFeeInt() uint64 {
	return uint64(cfg.AutoTxMaxFee.Value)
}

func (cfg *SmartNodeConfig) AutoTxGasThresholdInt() uint64 {
	return uint64(cfg.AutoTxGasThreshold.Value)
}

// ===============
// === Service ===
// ===============

func (cfg *SmartNodeConfig) GetSmartNodeContainerTag() string {
	return smartnodeTag
}

// ========================
// === Execution Client ===
// ========================

// Get the selected Beacon Node
func (cfg *SmartNodeConfig) GetSelectedExecutionClient() config.ExecutionClient {
	if cfg.IsLocalMode() {
		return cfg.LocalExecutionClient.ExecutionClient.Value
	}
	return cfg.ExternalExecutionClient.ExecutionClient.Value
}

// Gets the port mapping of the ec container
// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcOpenApiPorts() string {
	return cfg.LocalExecutionClient.GetOpenApiPortMapping()
}

// Gets the max peers of the ec container
// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcMaxPeers() (uint16, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return 0, fmt.Errorf("Execution client is external, there is no max peers")
	}
	return cfg.LocalExecutionClient.GetMaxPeers(), nil
}

// Gets the tag of the ec container
// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcContainerTag() (string, error) {
	if !cfg.IsLocalMode() {
		return "", fmt.Errorf("Execution client is external, there is no container tag")
	}
	return cfg.LocalExecutionClient.GetContainerTag(), nil
}

// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcAdditionalFlags() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Execution client is external, there are no additional flags")
	}
	return cfg.LocalExecutionClient.GetAdditionalFlags(), nil
}

// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetExternalIP() string {
	// Get the external IP address
	ip, err := config.GetExternalIP()
	if err != nil {
		fmt.Println("Warning: couldn't get external IP address; if you're using Nimbus or Besu, it may have trouble finding peers:")
		fmt.Println(err.Error())
		return ""
	}

	if ip.To4() == nil {
		fmt.Println("Warning: external IP address is v6; if you're using Nimbus or Besu, it may have trouble finding peers:")
	}

	return ip.String()
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetEcHttpEndpoint() string {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return fmt.Sprintf("http://%s:%d", ExecutionClientSuffix, cfg.LocalExecutionClient.HttpPort.Value)
	}

	return cfg.ExternalExecutionClient.HttpUrl.Value
}

// Get the endpoints of the EC, including the fallback if applicable
func (cfg *SmartNodeConfig) GetEcHttpEndpointsWithFallback() string {
	endpoints := cfg.GetEcHttpEndpoint()

	if cfg.Fallback.UseFallbackClients.Value {
		endpoints = fmt.Sprintf("%s,%s", endpoints, cfg.Fallback.EcHttpUrl.Value)
	}
	return endpoints
}

// ===================
// === Beacon Node ===
// ===================

// Get the selected Beacon Node
func (cfg *SmartNodeConfig) GetSelectedBeaconNode() config.BeaconNode {
	if cfg.IsLocalMode() {
		return cfg.LocalBeaconClient.BeaconNode.Value
	}
	return cfg.ExternalBeaconClient.BeaconNode.Value
}

// Gets the tag of the bn container
// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnContainerTag() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Beacon Node is external, there is no container tag")
	}
	return cfg.LocalBeaconClient.GetContainerTag(), nil
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnOpenPorts() []string {
	return cfg.LocalBeaconClient.GetOpenApiPortMapping()
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetEcWsEndpoint() string {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return fmt.Sprintf("ws://%s:%d", ExecutionClientSuffix, cfg.LocalExecutionClient.WebsocketPort.Value)
	}

	return cfg.ExternalExecutionClient.WebsocketUrl.Value
}

// Gets the max peers of the bn container
// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnMaxPeers() (uint16, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return 0, fmt.Errorf("Beacon Node is external, there is no max peers")
	}
	return cfg.LocalBeaconClient.GetMaxPeers(), nil
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnAdditionalFlags() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Beacon Node is external, there are no additional flags")
	}
	return cfg.LocalBeaconClient.GetAdditionalFlags(), nil
}

// Get the HTTP API endpoint for the provided BN
func (cfg *SmartNodeConfig) GetBnHttpEndpoint() string {
	if cfg.IsLocalMode() {
		return fmt.Sprintf("http://%s:%d", BeaconNodeSuffix, cfg.LocalBeaconClient.HttpPort.Value)
	}

	return cfg.ExternalBeaconClient.HttpUrl.Value
}

// Get the endpoints of the BN, including the fallback if applicable
func (cfg *SmartNodeConfig) GetBnHttpEndpointsWithFallback() string {
	endpoints := cfg.GetBnHttpEndpoint()

	if cfg.Fallback.UseFallbackClients.Value {
		endpoints = fmt.Sprintf("%s,%s", endpoints, cfg.Fallback.BnHttpUrl.Value)
	}
	return endpoints
}

// ===============
// === Metrics ===
// ===============

// Used by text/template to format exporter.yml
func (cfg *SmartNodeConfig) GetExporterAdditionalFlags() []string {
	flags := strings.Trim(cfg.Metrics.Exporter.AdditionalFlags.Value, " ")
	if flags == "" {
		return nil
	}
	return strings.Split(flags, " ")
}

// Used by text/template to format prometheus.yml
func (cfg *SmartNodeConfig) GetPrometheusAdditionalFlags() []string {
	flags := strings.Trim(cfg.Metrics.Prometheus.AdditionalFlags.Value, " ")
	if flags == "" {
		return nil
	}
	return strings.Split(flags, " ")
}

// Used by text/template to format prometheus.yml
func (cfg *SmartNodeConfig) GetPrometheusOpenPorts() string {
	portMode := cfg.Metrics.Prometheus.OpenPort.Value
	if !portMode.IsOpen() {
		return ""
	}
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(cfg.Metrics.Prometheus.Port.Value))
}

// Gets the hostname portion of the Execution Client's URI.
// Used by text/template to format prometheus.yml.
func (cfg *SmartNodeConfig) GetExecutionHostname() (string, error) {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return ExecutionClientSuffix, nil
	}
	ecUrl, err := url.Parse(cfg.ExternalExecutionClient.HttpUrl.Value)
	if err != nil {
		return "", fmt.Errorf("Invalid External Execution URL %s: %w", cfg.ExternalExecutionClient.HttpUrl.Value, err)
	}

	return ecUrl.Hostname(), nil
}

// Gets the hostname portion of the Beacon Node's URI.
// Used by text/template to format prometheus.yml.
func (cfg *SmartNodeConfig) GetBeaconHostname() (string, error) {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return BeaconNodeSuffix, nil
	}
	bnUrl, err := url.Parse(cfg.ExternalBeaconClient.HttpUrl.Value)
	if err != nil {
		return "", fmt.Errorf("Invalid External Consensus URL %s: %w", cfg.ExternalBeaconClient.HttpUrl.Value, err)
	}

	return bnUrl.Hostname(), nil
}

// ========================
// === Validator Client ===
// ========================

// Gets the tag of the VC container
func (cfg *SmartNodeConfig) GetVcContainerTag() string {
	bn := cfg.GetSelectedBeaconNode()
	return cfg.ValidatorClient.GetVcContainerTag(bn)
}

// Gets the additional flags of the selected VC
func (cfg *SmartNodeConfig) GetVcAdditionalFlags() string {
	bn := cfg.GetSelectedBeaconNode()
	return cfg.ValidatorClient.GetVcAdditionalFlags(bn)
}

// Check if doppelganger detection is enabled
func (cfg *SmartNodeConfig) IsDoppelgangerEnabled() bool {
	return cfg.ValidatorClient.VcCommon.DoppelgangerDetection.Value
}

// Used by text/template to format validator.yml
// Only returns the user-entered value, not the prefixed value
func (cfg *SmartNodeConfig) CustomGraffiti() string {
	return cfg.ValidatorClient.VcCommon.Graffiti.Value
}

// Used by text/template to format validator.yml
// Only returns the the prefix
func (cfg *SmartNodeConfig) GraffitiPrefix() string {
	// Graffiti
	identifier := ""
	versionString := fmt.Sprintf("v%s", shared.RocketPoolVersion)
	if len(versionString) < 8 {
		ecInitial := strings.ToUpper(string(cfg.GetSelectedExecutionClient())[:1])

		var bnInitial string
		bn := cfg.GetSelectedBeaconNode()
		switch bn {
		case config.BeaconNode_Lodestar:
			bnInitial = "S" // Lodestar is special because it conflicts with Lighthouse
		default:
			bnInitial = strings.ToUpper(string(bn)[:1])
		}

		var modeFlag string
		if cfg.IsNativeMode {
			modeFlag = "N" // I don't think this will ever actually get used but whatever
		} else if cfg.ClientMode.Value == config.ClientMode_Local {
			modeFlag = "L"
		} else {
			modeFlag = "X"
		}
		identifier = fmt.Sprintf("%s%s%s", ecInitial, bnInitial, modeFlag)
	}

	return fmt.Sprintf("RP%s %s", identifier, versionString)
}

// Used by text/template to format validator.yml
func (cfg *SmartNodeConfig) Graffiti() (string, error) {
	prefix := cfg.GraffitiPrefix()
	customGraffiti := cfg.CustomGraffiti()
	if customGraffiti == "" {
		return prefix, nil
	}
	return fmt.Sprintf("%s (%s)", prefix, customGraffiti), nil
}

// Used by text/template to format validator.yml
func (cfg *SmartNodeConfig) FeeRecipientFile() string {
	return FeeRecipientFilename
}

// Used by text/template to format validator.yml
func (cfg *SmartNodeConfig) MevBoostUrl() string {
	if !cfg.MevBoost.Enable.Value {
		return ""
	}

	if cfg.MevBoost.Mode.Value == config.ClientMode_Local {
		return fmt.Sprintf("http://%s:%d", config.ContainerID_MevBoost, cfg.MevBoost.Port.Value)
	}

	return cfg.MevBoost.ExternalUrl.Value
}

// =================
// === MEV-Boost ===
// =================

// Gets the name of the MEV-Boost start script
func (cfg *SmartNodeConfig) GetMevBoostStartScript() string {
	return MevBoostStartScript
}

// Used by text/template to format mev-boost.yml
func (cfg *SmartNodeConfig) GetMevBoostOpenPorts() string {
	portMode := cfg.MevBoost.OpenRpcPort.Value
	if !portMode.IsOpen() {
		return ""
	}
	port := cfg.MevBoost.Port.Value
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(port))
}

// ==============
// === Addons ===
// ==============

func (cfg *SmartNodeConfig) GetAddonsFolderPath() string {
	return filepath.Join(cfg.rocketPoolDirectory, AddonsFolderName)
}

func (cfg *SmartNodeConfig) GetGwwPath() string {
	return filepath.Join(cfg.GetAddonsFolderPath(), gww.FolderName)
}

func (cfg *SmartNodeConfig) GetGwwGraffitiFilePath() string {
	return filepath.Join(cfg.GetGwwPath(), gww.GraffitiFilename)
}
