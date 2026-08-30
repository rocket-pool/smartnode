package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	CommitBoostConfigFile     string = "cb_config.toml"
	CommitBoostConfigTemplate string = "commit-boost-config"
	commitBoostProdTag        string = "ghcr.io/commit-boost/commit-boost:v0.10.0"
	commitBoostTestTag        string = "ghcr.io/commit-boost/commit-boost:v0.10.0"
)

// Relay selection mode for Commit-Boost PBS
type PbsRelaySelectionMode string

const (
	PbsRelaySelectionMode_All    PbsRelaySelectionMode = "all"
	PbsRelaySelectionMode_Manual PbsRelaySelectionMode = "manual"
)

// A relay entry ready for template rendering
type PbsRelayInfo struct {
	ID  string
	URL string
}

// Configuration for Commit-Boost's service
type CommitBoostConfig struct {
	// Ownership mode
	Mode config.Parameter `yaml:"mode,omitempty"`

	// The mode for relay selection
	RelaySelectionMode config.Parameter `yaml:"relaySelectionMode,omitempty"`

	// Flashbots relay
	FlashbotsRelay config.Parameter `yaml:"cbFlashbotsEnabled,omitempty"`

	// bloXroute regulated relay
	BloxRouteRegulatedRelay config.Parameter `yaml:"cbBloxRouteRegulatedEnabled,omitempty"`

	// Titan regional relay
	TitanRegionalRelay config.Parameter `yaml:"cbTitanRegionalEnabled,omitempty"`

	// Ultra Sound filtered relay
	UltrasoundFilteredRelay config.Parameter `yaml:"cbUltrasoundFilteredEnabled,omitempty"`

	// BTCS OFAC+ relay
	BtcsOfacRelay config.Parameter `yaml:"cbBtcsOfacEnabled,omitempty"`

	// Custom relays provided by the user (comma-separated URLs)
	CustomRelays config.Parameter `yaml:"customRelays,omitempty"`

	// The URL of an external Commit-Boost client
	ExternalUrl config.Parameter `yaml:"externalUrl"`

	// The Docker Hub tag for Commit-Boost
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// The port that Commit-Boost should serve its API on
	Port config.Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the HTTP port outside of Docker
	OpenRpcPort config.Parameter `yaml:"openRpcPort,omitempty"`

	///////////////////////////
	// Non-editable settings //
	///////////////////////////

	parentConfig *RocketPoolConfig                     `yaml:"-"`
	relays       []config.MevRelay                     `yaml:"-"`
	relayMap     map[config.MevRelayID]config.MevRelay `yaml:"-"`
}

// Generates a new Commit-Boost PBS service configuration
func NewCommitBoostConfig(cfg *RocketPoolConfig) *CommitBoostConfig {
	// Generate the relays
	relays := createCommitBoostRelays(cfg.networks)
	relayMap := map[config.MevRelayID]config.MevRelay{}
	for _, relay := range relays {
		relayMap[relay.ID] = relay
	}

	portModes := config.PortModes("")

	return &CommitBoostConfig{
		parentConfig: cfg,
		relays:       relays,
		relayMap:     relayMap,

		Mode: config.Parameter{
			ID:                 "mode",
			Name:               "Commit-Boost Mode",
			Description:        "Choose whether to let the Smartnode manage your Commit-Boost instance (Locally Managed), or if you manage your own outside of the Smartnode stack (Externally Managed).",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage the Commit-Boost client for you",
				Value:       config.Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing Commit-Boost client that you manage on your own",
				Value:       config.Mode_External,
			}},
		},

		RelaySelectionMode: config.Parameter{
			ID:                 "relaySelectionMode",
			Name:               "Relay Selection Mode",
			Description:        "Select how the TUI shows you the options for which PBS relays to enable.\n\nNote that all of the built-in relays support regional sanction lists (such as the US OFAC list) and are compliant with regulations.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: PbsRelaySelectionMode_All},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Use All Relays",
				Description: "Use this if you simply want to enable all of the built-in relays without needing to read about each individual relay. If new relays get added, you'll automatically start using those too.\n\nNote that all of the built-in relays support regional sanction lists (such as the US OFAC list) and are compliant with regulations.",
				Value:       PbsRelaySelectionMode_All,
			}, {
				Name:        "Manual Mode",
				Description: "Each relay will be shown, and you can enable each one individually as you see fit.\nUse this if you already know about the relays and want to customize the ones you will use.",
				Value:       PbsRelaySelectionMode_Manual,
			}},
		},

		FlashbotsRelay:          generateCbRelayParameter("cbFlashbotsEnabled", relayMap[config.MevRelayID_Flashbots]),
		BloxRouteRegulatedRelay: generateCbRelayParameter("cbBloxRouteRegulatedEnabled", relayMap[config.MevRelayID_BloxrouteRegulated]),
		TitanRegionalRelay:      generateCbRelayParameter("cbTitanRegionalEnabled", relayMap[config.MevRelayID_TitanRegional]),
		UltrasoundFilteredRelay: generateCbRelayParameter("cbUltrasoundFilteredEnabled", relayMap[config.MevRelayID_UltrasoundFiltered]),
		BtcsOfacRelay:           generateCbRelayParameter("cbBtcsOfacEnabled", relayMap[config.MevRelayID_BTCSOfac]),

		CustomRelays: config.Parameter{
			ID:                 "customRelays",
			Name:               "Custom Relays",
			Description:        "Add custom relay URLs to Commit-Boost that aren't part of the built-in set. You can add multiple relays by separating each one with a comma. Any relay URLs can be used as long as they match your selected Ethereum network.\n\nFor a comprehensive list of available relays, we recommend the list maintained by ETHStaker:\nhttps://github.com/eth-educators/ethstaker-guides/blob/main/MEV-relay-list.md",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		Port: config.Parameter{
			ID:                 "port",
			Name:               "Port",
			Description:        "The port that Commit-Boost should serve its API on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: uint16(18550)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenRpcPort: config.Parameter{
			ID:                 "openRpcPort",
			Name:               "Expose API Port",
			Description:        "Expose the API port to other processes on your machine, or to your local network so other local machines can access Commit-Boost's API.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.RPC_Closed},
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            portModes,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Container Tag",
			Description:        "The tag name of the Commit-Boost container you want to use.",
			AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
			Type:               config.ParameterType_String,
			Default:            clientTagDefaults(cfg.networks, commitBoostProdTag, commitBoostTestTag),
		},

		ExternalUrl: config.Parameter{
			ID:          "externalUrl",
			Name:        "External URL",
			Description: "The URL of the external Commit-Boost client or provider.",
			Type:        config.ParameterType_String,
			Default:     map[config.Network]interface{}{config.Network_All: ""},
		},
	}
}

// The title for the config
func (cfg *CommitBoostConfig) GetConfigTitle() string {
	return "Commit-Boost Settings"
}

// Get the Parameters for this config
func (cfg *CommitBoostConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Mode,
		&cfg.RelaySelectionMode,
		&cfg.FlashbotsRelay,
		&cfg.BloxRouteRegulatedRelay,
		&cfg.TitanRegionalRelay,
		&cfg.UltrasoundFilteredRelay,
		&cfg.BtcsOfacRelay,
		&cfg.CustomRelays,
		&cfg.Port,
		&cfg.OpenRpcPort,
		&cfg.ContainerTag,
		&cfg.ExternalUrl,
	}
}

// Get the filename for the Commit-Boost PBS config
func (cfg *CommitBoostConfig) GetCommitBoostConfigFilename() string {
	return CommitBoostConfigFile
}

func (cfg *CommitBoostConfig) GetCommitBoostOpenPorts() string {
	portMode := cfg.OpenRpcPort.Value.(config.RPCMode)
	if !portMode.Open() {
		return ""
	}
	port := cfg.Port.Value.(uint16)
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(port))
}

// Get the chain name for the Commit-Boost config file.
// Commit-Boost expects actual Ethereum network names, not generic labels.
func (cfg *CommitBoostConfig) GetChainName(network config.Network) (string, error) {
	if cfg.parentConfig != nil {
		if info := cfg.parentConfig.networks.GetNetwork(network); info != nil && info.CommitBoostChainName != "" {
			return info.CommitBoostChainName, nil
		}
	}
	return "", fmt.Errorf("unsupported network %s for Commit-Boost PBS config", network)
}

// Get the chain name for the current network (for use in templates)
func (cfg *CommitBoostConfig) GetCurrentChainName() string {
	network := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	name, _ := cfg.GetChainName(network)
	return name
}

// Get the relays that are available for the current network
func (cfg *CommitBoostConfig) GetAvailableRelays() []config.MevRelay {
	relays := []config.MevRelay{}
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	for _, relay := range cfg.relays {
		if relay.Urls.UrlExists(currentNetwork) {
			relays = append(relays, relay)
		}
	}
	return relays
}

// Get the enabled PBS relays, formatted for template rendering.
func (cfg *CommitBoostConfig) GetEnabledPbsRelayInfo() []PbsRelayInfo {
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	relays := cfg.getEnabledRelays()
	result := []PbsRelayInfo{}
	for _, relay := range relays {
		if url, ok := relay.Urls[currentNetwork]; ok && url != "" {
			result = append(result, PbsRelayInfo{
				ID:  string(relay.ID),
				URL: url,
			})
		}
	}
	return result
}

// Get which PBS relays are enabled based on the relay selection mode
func (cfg *CommitBoostConfig) getEnabledRelays() []config.MevRelay {
	relays := []config.MevRelay{}
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)

	switch cfg.RelaySelectionMode.Value.(PbsRelaySelectionMode) {
	case PbsRelaySelectionMode_All:
		for _, relay := range cfg.relays {
			if relay.Urls.UrlExists(currentNetwork) {
				relays = append(relays, relay)
			}
		}

	case PbsRelaySelectionMode_Manual:
		relays = cfg.maybeAddRelay(relays, cfg.FlashbotsRelay, config.MevRelayID_Flashbots, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.BloxRouteRegulatedRelay, config.MevRelayID_BloxrouteRegulated, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.TitanRegionalRelay, config.MevRelayID_TitanRegional, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.UltrasoundFilteredRelay, config.MevRelayID_UltrasoundFiltered, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.BtcsOfacRelay, config.MevRelayID_BTCSOfac, currentNetwork)
	}

	return relays
}

func (cfg *CommitBoostConfig) maybeAddRelay(relays []config.MevRelay, relayParam config.Parameter, relayID config.MevRelayID, currentNetwork config.Network) []config.MevRelay {
	if relayParam.Value == true {
		if cfg.relayMap[relayID].Urls.UrlExists(currentNetwork) {
			relays = append(relays, cfg.relayMap[relayID])
		}
	}
	return relays
}

// Get custom relay URLs (comma-separated in CustomRelays) as a slice
func (cfg *CommitBoostConfig) GetCustomRelays() []string {
	customRelays, ok := cfg.CustomRelays.Value.(string)
	if !ok || customRelays == "" {
		return nil
	}
	result := []string{}
	for relay := range strings.SplitSeq(customRelays, ",") {
		trimmed := strings.TrimSpace(relay)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Get the relay string for all enabled relays (built-in + custom), comma-separated
func (cfg *CommitBoostConfig) GetRelayString() string {
	relayUrls := []string{}
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)

	relays := cfg.getEnabledRelays()
	for _, relay := range relays {
		relayUrls = append(relayUrls, relay.Urls[currentNetwork])
	}
	if customRelays := cfg.GetCustomRelays(); len(customRelays) > 0 {
		relayUrls = append(relayUrls, customRelays...)
	}

	return strings.Join(relayUrls, ",")
}

// Get the container tag value as a string (for use in templates)
func (cfg *CommitBoostConfig) GetContainerTag() string {
	return fmt.Sprint(cfg.ContainerTag.Value)
}

// Create the default Commit-Boost PBS relays. URLs come from the loaded network YAML.
func createCommitBoostRelays(networks *NetworksConfig) []config.MevRelay {
	return []config.MevRelay{
		{
			ID:          config.MevRelayID_Flashbots,
			Name:        "Flashbots",
			Description: "Flashbots is the developer of MEV-Boost, and one of the best-known and most trusted relays in the space.",
			Urls:        relayUrlMap(networks, config.MevRelayID_Flashbots),
			Regulated:   true,
		},
		{
			ID:          config.MevRelayID_BloxrouteRegulated,
			Name:        "bloXroute Regulated",
			Description: "Select this to enable the \"regulated\" relay from bloXroute.",
			Urls:        relayUrlMap(networks, config.MevRelayID_BloxrouteRegulated),
			Regulated:   true,
		},
		{
			ID:          config.MevRelayID_TitanRegional,
			Name:        "Titan Regional",
			Description: "Titan Relay is a neutral, Rust-based PBS Relay optimized for low latency throughput, geographical distribution, and robustness. This is the regulated (filtering) version.",
			Urls:        relayUrlMap(networks, config.MevRelayID_TitanRegional),
			Regulated:   true,
		},
		{
			ID:          config.MevRelayID_UltrasoundFiltered,
			Name:        "Ultra Sound (filtering)",
			Description: "The ultra sound relay is a credibly-neutral and permissionless relay — a public good from the ultrasound.money team. This is the OFAC-filtering version.",
			Urls:        relayUrlMap(networks, config.MevRelayID_UltrasoundFiltered),
			Regulated:   true,
		},
		{
			ID:          config.MevRelayID_BTCSOfac,
			Name:        "BTCS OFAC+",
			Description: "Select this to enable the BTCS OFAC+ regulated relay.",
			Urls:        relayUrlMap(networks, config.MevRelayID_BTCSOfac),
			Regulated:   true,
		},
	}
}

// Generate one of the relay parameters for Commit-Boost
func generateCbRelayParameter(id string, relay config.MevRelay) config.Parameter {
	description := fmt.Sprintf("[lime]NOTE: You can enable multiple options.\n\n[white]%s\n\n", relay.Description)
	description += "Complies with Regulations: YES\n"

	return config.Parameter{
		ID:                 id,
		Name:               fmt.Sprintf("Enable %s", relay.Name),
		Description:        description,
		Type:               config.ParameterType_Bool,
		Default:            map[config.Network]interface{}{config.Network_All: false},
		AffectsContainers:  []config.ContainerID{config.ContainerID_CommitBoost},
		CanBeBlank:         false,
		OverwriteOnUpgrade: false,
	}
}
