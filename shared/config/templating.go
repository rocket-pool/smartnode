package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rocket-pool/node-manager-core/config"
	rn "github.com/rocket-pool/smartnode/addons/rescue_node"
	"github.com/rocket-pool/smartnode/shared"
)

// =================
// === Constants ===
// =================

func (c *SmartNodeConfig) BeaconNodeContainerName() string {
	return string(config.ContainerID_BeaconNode)
}

func (c *SmartNodeConfig) DaemonContainerName() string {
	return string(config.ContainerID_Daemon)
}

func (c *SmartNodeConfig) ExecutionClientContainerName() string {
	return string(config.ContainerID_ExecutionClient)
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

func (c *SmartNodeConfig) ExecutionClientDataVolume() string {
	return ExecutionClientDataVolume
}

func (c *SmartNodeConfig) BeaconNodeDataVolume() string {
	return BeaconNodeDataVolume
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
	rescueNode := rn.NewRescueNode(cfg.AddonsConfig.RescueNode)
	overrides, err := rescueNode.GetOverrides(bn)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}
	if overrides != nil {
		// Use the rescue node
		return overrides.CcApiEndpoint, nil
	}
	return cfg.GetBnHttpEndpoint(), nil
}

func (cfg *SmartNodeConfig) BnRpcUrl() (string, error) {
	// Check if Rescue Node is in-use
	bn := cfg.GetSelectedBeaconNode()
	if bn != config.BeaconNode_Prysm {
		return "", nil
	}

	rescueNode := rn.NewRescueNode(cfg.AddonsConfig.RescueNode)
	overrides, err := rescueNode.GetOverrides(bn)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}
	if overrides != nil {
		// Use the rescue node
		return overrides.CcRpcEndpoint, nil
	}
	if cfg.IsLocalMode() {
		return fmt.Sprintf("%s:%d", config.ContainerID_BeaconNode, cfg.LocalBeaconConfig.Prysm.RpcPort.Value), nil
	}
	return cfg.ExternalBeaconConfig.PrysmRpcUrl.Value, nil
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
		return cfg.LocalExecutionConfig.ExecutionClient.Value
	}
	return cfg.ExternalExecutionConfig.ExecutionClient.Value
}

// Gets the port mapping of the ec container
// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcOpenApiPorts() string {
	return cfg.LocalExecutionConfig.GetOpenApiPortMapping()
}

// Gets the max peers of the ec container
// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcMaxPeers() (uint16, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return 0, fmt.Errorf("Execution client is external, there is no max peers")
	}
	return cfg.LocalExecutionConfig.GetMaxPeers(), nil
}

// Gets the tag of the ec container
// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcContainerTag() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Execution client is external, there is no container tag")
	}
	return cfg.LocalExecutionConfig.GetContainerTag(), nil
}

// Used by text/template to format ec.yml
func (cfg *SmartNodeConfig) GetEcAdditionalFlags() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Execution client is external, there are no additional flags")
	}
	return cfg.LocalExecutionConfig.GetAdditionalFlags(), nil
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
		return fmt.Sprintf("http://%s:%d", config.ContainerID_ExecutionClient, cfg.LocalExecutionConfig.HttpPort.Value)
	}

	return cfg.ExternalExecutionConfig.HttpUrl.Value
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
		return cfg.LocalBeaconConfig.BeaconNode.Value
	}
	return cfg.ExternalBeaconConfig.BeaconNode.Value
}

// Gets the tag of the bn container
// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnContainerTag() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Beacon Node is external, there is no container tag")
	}
	return cfg.LocalBeaconConfig.GetContainerTag(), nil
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnOpenPorts() []string {
	return cfg.LocalBeaconConfig.GetOpenApiPortMapping()
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetEcWsEndpoint() string {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return fmt.Sprintf("ws://%s:%d", config.ContainerID_ExecutionClient, cfg.LocalExecutionConfig.WebsocketPort.Value)
	}

	return cfg.ExternalExecutionConfig.WebsocketUrl.Value
}

// Gets the max peers of the bn container
// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnMaxPeers() (uint16, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return 0, fmt.Errorf("Beacon Node is external, there is no max peers")
	}
	return cfg.LocalBeaconConfig.GetMaxPeers(), nil
}

// Used by text/template to format bn.yml
func (cfg *SmartNodeConfig) GetBnAdditionalFlags() (string, error) {
	if cfg.ClientMode.Value != config.ClientMode_Local {
		return "", fmt.Errorf("Beacon Node is external, there is no additional flags")
	}
	return cfg.LocalBeaconConfig.GetAdditionalFlags(), nil
}

// Get the HTTP API endpoint for the provided BN
func (cfg *SmartNodeConfig) GetBnHttpEndpoint() string {
	if cfg.IsLocalMode() {
		return fmt.Sprintf("http://%s:%d", config.ContainerID_BeaconNode, cfg.LocalBeaconConfig.HttpPort.Value)
	}

	return cfg.ExternalBeaconConfig.HttpUrl.Value
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

// Gets the hostname portion of the Execution Client's URI.
// Used by text/template to format prometheus.yml.
func (cfg *SmartNodeConfig) GetExecutionHostname() (string, error) {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return string(config.ContainerID_ExecutionClient), nil
	}
	ecUrl, err := url.Parse(cfg.ExternalExecutionConfig.HttpUrl.Value)
	if err != nil {
		return "", fmt.Errorf("Invalid External Execution URL %s: %w", cfg.ExternalExecutionConfig.HttpUrl.Value, err)
	}

	return ecUrl.Hostname(), nil
}

// Gets the hostname portion of the Beacon Node's URI.
// Used by text/template to format prometheus.yml.
func (cfg *SmartNodeConfig) GetBeaconHostname() (string, error) {
	if cfg.ClientMode.Value == config.ClientMode_Local {
		return string(config.ContainerID_BeaconNode), nil
	}
	ccUrl, err := url.Parse(cfg.ExternalBeaconConfig.HttpUrl.Value)
	if err != nil {
		return "", fmt.Errorf("Invalid External Consensus URL %s: %w", cfg.ExternalBeaconConfig.HttpUrl.Value, err)
	}

	return ccUrl.Hostname(), nil
}

// ========================
// === Validator Client ===
// ========================

// Gets the tag of the VC container
func (cfg *SmartNodeConfig) GetVcContainerTag() string {
	bn := cfg.GetSelectedBeaconNode()
	return cfg.ValidatorClientConfig.GetVcContainerTag(bn)
}

// Gets the additional flags of the selected VC
func (cfg *SmartNodeConfig) GetVcAdditionalFlags() string {
	bn := cfg.GetSelectedBeaconNode()
	return cfg.ValidatorClientConfig.GetVcAdditionalFlags(bn)
}

// Check if doppelganger detection is enabled
func (cfg *SmartNodeConfig) IsDoppelgangerEnabled() bool {
	return cfg.ValidatorClientConfig.VcCommon.DoppelgangerDetection.Value
}

// Used by text/template to format validator.yml
// Only returns the user-entered value, not the prefixed value
func (cfg *SmartNodeConfig) CustomGraffiti() string {
	return cfg.ValidatorClientConfig.VcCommon.Graffiti.Value
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
