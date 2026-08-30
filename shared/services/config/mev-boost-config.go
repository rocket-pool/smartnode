package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	mevBoostTagProd             string = "flashbots/mev-boost:1.12.0"
	mevBoostTagTest             string = "flashbots/mev-boost:1.12.0"
	mevDocsUrl                  string = "https://docs.rocketpool.net/node-staking/mev"
	RegulatedRelayDescription   string = "Select this to enable the relays that comply with government regulations (e.g. OFAC sanctions), "
	UnregulatedRelayDescription string = "Select this to enable the relays that do not follow any sanctions lists (do not censor transactions), "
	NoSandwichRelayDescription  string = "and do not allow front-running or sandwich attacks."
	AllMevRelayDescription      string = "and allow for all types of MEV (including sandwich attacks)."
)

// Configuration for MEV-Boost
type MevBoostConfig struct {
	Title string `yaml:"-"`

	// Ownership mode
	Mode config.Parameter `yaml:"mode,omitempty"`

	// The mode for relay selection
	SelectionMode config.Parameter `yaml:"selectionMode,omitempty"`

	// Regulated, all types
	EnableRegulatedAllMev config.Parameter `yaml:"enableRegulatedAllMev,omitempty"`

	// Unregulated, all types
	EnableUnregulatedAllMev config.Parameter `yaml:"enableUnregulatedAllMev,omitempty"`

	// Flashbots relay
	FlashbotsRelay config.Parameter `yaml:"flashbotsEnabled,omitempty"`

	// bloXroute regulated relay
	BloxRouteRegulatedRelay config.Parameter `yaml:"bloxRouteRegulatedEnabled,omitempty"`

	// Ultra sound relay
	UltrasoundRelay config.Parameter `yaml:"ultrasoundEnabled,omitempty"`

	// Ultra sound filtered relay
	UltrasoundFilteredRelay config.Parameter `yaml:"ultrasoundFilteredEnabled,omitempty"`

	// Aestus relay
	AestusRelay config.Parameter `yaml:"aestusEnabled,omitempty"`

	// Titan Global relay
	TitanGlobalRelay config.Parameter `yaml:"titanGlobalEnabled,omitempty"`

	// Titan Regional relay
	TitanRegionalRelay config.Parameter `yaml:"titanRegionalEnabled,omitempty"`

	// BTCS OFAC+
	BtcsOfacRelay config.Parameter `yaml:"btcsOfacEnabled,omitempty"`

	// The RPC port
	Port config.Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the HTTP port outside of Docker
	OpenRpcPort config.Parameter `yaml:"openRpcPort,omitempty"`

	// The Docker Hub tag for MEV-Boost
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`

	// The URL of an external MEV-Boost client
	ExternalUrl config.Parameter `yaml:"externalUrl"`

	///////////////////////////
	// Non-editable settings //
	///////////////////////////

	parentConfig *RocketPoolConfig                     `yaml:"-"`
	relays       []config.MevRelay                     `yaml:"-"`
	relayMap     map[config.MevRelayID]config.MevRelay `yaml:"-"`
}

// Generates a new MEV-Boost configuration
func NewMevBoostConfig(cfg *RocketPoolConfig) *MevBoostConfig {
	// Generate the relays
	relays := createDefaultRelays(cfg.networks)
	relayMap := map[config.MevRelayID]config.MevRelay{}
	for _, relay := range relays {
		relayMap[relay.ID] = relay
	}

	rpcPortModes := config.PortModes("")

	return &MevBoostConfig{
		Title: "MEV-Boost Settings",

		parentConfig: cfg,

		Mode: config.Parameter{
			ID:                 "mode",
			Name:               "MEV-Boost Mode",
			Description:        "Choose whether to let the Smart Node manage your MEV-Boost instance (Locally Managed), or if you manage your own outside of the Smart Node stack (Externally Managed).",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smart Node to manage the MEV-Boost client for you",
				Value:       config.Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing MEV-Boost client that you manage on your own",
				Value:       config.Mode_External,
			}},
		},

		SelectionMode: config.Parameter{
			ID:                 "selectionMode",
			Name:               "Selection Mode",
			Description:        "Select how the TUI shows you the options for which MEV relays to enable.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.MevSelectionMode_Profile},
			AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Profile Mode",
				Description: "Relays will be bundled up based on whether or not they're regulated, and whether or not they allow sandwich attacks.\nUse this if you simply want to specify which type of relay you want to use without needing to read about each individual relay.",
				Value:       config.MevSelectionMode_Profile,
			}, {
				Name:        "Relay Mode",
				Description: "Each relay will be shown, and you can enable each one individually as you see fit.\nUse this if you already know about the relays and want to customize the ones you will use.",
				Value:       config.MevSelectionMode_Relay,
			}},
		},

		EnableRegulatedAllMev:   generateProfileParameter("enableRegulatedAllMev", relays, true, cfg.networks),
		EnableUnregulatedAllMev: generateProfileParameter("enableUnregulatedAllMev", relays, false, cfg.networks),

		// Explicit relay params
		FlashbotsRelay:          generateRelayParameter("flashbotsEnabled", relayMap[config.MevRelayID_Flashbots]),
		BloxRouteRegulatedRelay: generateRelayParameter("bloxRouteRegulatedEnabled", relayMap[config.MevRelayID_BloxrouteRegulated]),
		UltrasoundRelay:         generateRelayParameter("ultrasoundEnabled", relayMap[config.MevRelayID_Ultrasound]),
		UltrasoundFilteredRelay: generateRelayParameter("ultrasoundFilteredEnabled", relayMap[config.MevRelayID_UltrasoundFiltered]),
		AestusRelay:             generateRelayParameter("aestusEnabled", relayMap[config.MevRelayID_Aestus]),
		TitanGlobalRelay:        generateRelayParameter("titanGlobalEnabled", relayMap[config.MevRelayID_TitanGlobal]),
		TitanRegionalRelay:      generateRelayParameter("titanRegionalEnabled", relayMap[config.MevRelayID_TitanRegional]),
		BtcsOfacRelay:           generateRelayParameter("btcsOfacEnabled", relayMap[config.MevRelayID_BTCSOfac]),

		Port: config.Parameter{
			ID:                 "port",
			Name:               "Port",
			Description:        "The port that MEV-Boost should serve its API on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: uint16(18550)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenRpcPort: config.Parameter{
			ID:                 "openRpcPort",
			Name:               "Expose API Port",
			Description:        "Expose the API port to other processes on your machine, or to your local network so other local machines can access MEV-Boost's API.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.RPC_Closed},
			AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            rpcPortModes,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Container Tag",
			Description:        "The tag name of the MEV-Boost container you want to use on Docker Hub.",
			Type:               config.ParameterType_String,
			Default:            clientTagDefaults(cfg.networks, mevBoostTagProd, mevBoostTagTest),
			AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to MEV-Boost, to take advantage of other settings that the Smart Node's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		ExternalUrl: config.Parameter{
			ID:                 "externalUrl",
			Name:               "External URL",
			Description:        "The URL of the external MEV-Boost client or provider",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		relays:   relays,
		relayMap: relayMap,
	}
}

// Get the config.Parameters for this config
func (cfg *MevBoostConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Mode,
		&cfg.SelectionMode,
		&cfg.EnableRegulatedAllMev,
		&cfg.EnableUnregulatedAllMev,
		&cfg.FlashbotsRelay,
		&cfg.BloxRouteRegulatedRelay,
		&cfg.UltrasoundRelay,
		&cfg.UltrasoundFilteredRelay,
		&cfg.AestusRelay,
		&cfg.TitanGlobalRelay,
		&cfg.TitanRegionalRelay,
		&cfg.BtcsOfacRelay,
		&cfg.Port,
		&cfg.OpenRpcPort,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
		&cfg.ExternalUrl,
	}
}

// The title for the config
func (cfg *MevBoostConfig) GetConfigTitle() string {
	return cfg.Title
}

// Get the profiles that are available for the current network
func (cfg *MevBoostConfig) GetAvailableProfiles() (bool, bool) {
	regulatedAllMev := false
	unregulatedAllMev := false

	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	for _, relay := range cfg.relays {
		if relay.Urls.UrlExists(currentNetwork) {
			regulatedAllMev = regulatedAllMev || relay.Regulated
			unregulatedAllMev = unregulatedAllMev || !relay.Regulated
		}
	}

	return regulatedAllMev, unregulatedAllMev
}

// Get the relays that are available for the current network
func (cfg *MevBoostConfig) GetAvailableRelays() []config.MevRelay {
	relays := []config.MevRelay{}
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	for _, relay := range cfg.relays {
		if relay.Urls.UrlExists(currentNetwork) {
			relays = append(relays, relay)
		}
	}

	return relays
}

// Get which MEV-boost relays are enabled
func (cfg *MevBoostConfig) GetEnabledMevRelays() []config.MevRelay {
	relays := []config.MevRelay{}

	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	switch cfg.SelectionMode.Value.(config.MevSelectionMode) {
	case config.MevSelectionMode_Profile:
		for _, relay := range cfg.relays {
			if !relay.Urls.UrlExists(currentNetwork) {
				// Skip relays that don't exist on the current network
				continue
			}
			if relay.Regulated {
				if cfg.EnableRegulatedAllMev.Value == true {
					relays = append(relays, relay)
				}
			} else {
				if cfg.EnableUnregulatedAllMev.Value == true {
					relays = append(relays, relay)
				}
			}
		}

	case config.MevSelectionMode_Relay:
		relays = cfg.maybeAddRelay(relays, cfg.FlashbotsRelay, config.MevRelayID_Flashbots, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.BloxRouteRegulatedRelay, config.MevRelayID_BloxrouteRegulated, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.UltrasoundRelay, config.MevRelayID_Ultrasound, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.UltrasoundFilteredRelay, config.MevRelayID_UltrasoundFiltered, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.AestusRelay, config.MevRelayID_Aestus, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.TitanGlobalRelay, config.MevRelayID_TitanGlobal, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.TitanRegionalRelay, config.MevRelayID_TitanRegional, currentNetwork)
		relays = cfg.maybeAddRelay(relays, cfg.BtcsOfacRelay, config.MevRelayID_BTCSOfac, currentNetwork)

	}

	return relays
}

func (cfg *MevBoostConfig) GetRelayString() string {
	relayUrls := []string{}
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)

	relays := cfg.GetEnabledMevRelays()
	for _, relay := range relays {
		relayUrls = append(relayUrls, relay.Urls[currentNetwork])
	}

	relayString := strings.Join(relayUrls, ",")
	return relayString
}

// Create the default MEV relays. URLs come from the loaded network YAML.
func createDefaultRelays(networks *NetworksConfig) []config.MevRelay {
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
			Description: "Select this to enable the bloXroute relay.",
			Urls:        relayUrlMap(networks, config.MevRelayID_BloxrouteRegulated),
			Regulated:   true,
		},
		{
			ID:          config.MevRelayID_Ultrasound,
			Name:        "Ultra Sound (non-filtering)",
			Description: "The ultra sound relay is a credibly-neutral and permissionless relay — a public good from the ultrasound.money team.",
			Urls:        relayUrlMap(networks, config.MevRelayID_Ultrasound),
			Regulated:   false,
		},
		{
			ID:          config.MevRelayID_UltrasoundFiltered,
			Name:        "Ultra Sound (filtering)",
			Description: "The ultra sound relay is a credibly-neutral and permissionless relay — a public good from the ultrasound.money team. This is the filtering version.",
			Urls:        relayUrlMap(networks, config.MevRelayID_UltrasoundFiltered),
			Regulated:   true,
		},
		{
			ID:          config.MevRelayID_Aestus,
			Name:        "Aestus",
			Description: "The Aestus MEV-Boost Relay is an independent and non-censoring relay. It is committed to neutrality and the development of a healthy MEV-Boost ecosystem.",
			Urls:        relayUrlMap(networks, config.MevRelayID_Aestus),
			Regulated:   false,
		},
		{
			ID:          config.MevRelayID_TitanGlobal,
			Name:        "Titan Global (non-filtering)",
			Description: "Titan Relay is a neutral, Rust-based MEV-Boost Relay optimized for low latency throughput, geographical distribution, and robustness. Select this to enable the \"non-filtering\" relay from Titan.",
			Urls:        relayUrlMap(networks, config.MevRelayID_TitanGlobal),
			Regulated:   false,
		},
		{
			ID:          config.MevRelayID_TitanRegional,
			Name:        "Titan Regional (filtering)",
			Description: "Titan Relay is a neutral, Rust-based MEV-Boost Relay optimized for low latency throughput, geographical distribution, and robustness. Select this to enable the \"filtering\" relay from Titan.",
			Urls:        relayUrlMap(networks, config.MevRelayID_TitanRegional),
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

// Generate one of the profile parameters
func generateProfileParameter(id string, relays []config.MevRelay, regulated bool, networks *NetworksConfig) config.Parameter {
	name := "Enable "
	description := fmt.Sprintf("[lime]NOTE: You can enable multiple options.\n\nTo learn more about MEV, please visit %s.\n\n[white]", mevDocsUrl)

	if regulated {
		name += "Regulated "
		description += RegulatedRelayDescription
	} else {
		name += "Unregulated "
		description += UnregulatedRelayDescription
	}

	descriptions := map[config.Network]string{}
	defaultDescription := description
	if networks != nil {
		for _, n := range networks.AllNetworks() {
			names := []string{}
			for _, relay := range relays {
				if !relay.Urls.UrlExists(n.ID()) {
					continue
				}
				if relay.Regulated == regulated {
					names = append(names, relay.Name)
				}
			}
			d := description + "\n\nRelays: " + strings.Join(names, ", ")
			descriptions[n.ID()] = d
			if n.Default {
				defaultDescription = d
			}
		}
	}

	return config.Parameter{
		ID:                    id,
		Name:                  name,
		Description:           defaultDescription,
		Type:                  config.ParameterType_Bool,
		Default:               map[config.Network]interface{}{config.Network_All: false},
		AffectsContainers:     []config.ContainerID{config.ContainerID_MevBoost},
		CanBeBlank:            false,
		OverwriteOnUpgrade:    false,
		DescriptionsByNetwork: descriptions,
	}
}

// Generate one of the relay parameters
func generateRelayParameter(id string, relay config.MevRelay) config.Parameter {
	description := fmt.Sprintf("[lime]NOTE: You can enable multiple options.\n\nTo learn more about MEV, please visit %s.\n\n[white]%s\n\n", mevDocsUrl, relay.Description)

	if relay.Regulated {
		description += "Complies with Regulations: YES\n"
	} else {
		description += "Complies with Regulations: NO\n"
	}

	return config.Parameter{
		ID:                 id,
		Name:               fmt.Sprintf("Enable %s", relay.Name),
		Description:        description,
		Type:               config.ParameterType_Bool,
		Default:            map[config.Network]interface{}{config.Network_All: false},
		AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
		CanBeBlank:         false,
		OverwriteOnUpgrade: false,
	}
}

func (cfg *MevBoostConfig) maybeAddRelay(relays []config.MevRelay, relayParam config.Parameter, relayID config.MevRelayID, currentNetwork config.Network) []config.MevRelay {
	if relayParam.Value == true {
		if cfg.relayMap[relayID].Urls.UrlExists(currentNetwork) {
			relays = append(relays, cfg.relayMap[relayID])
		}
	}
	return relays
}
