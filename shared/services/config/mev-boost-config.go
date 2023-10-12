package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	mevBoostTag                 string = "flashbots/mev-boost:1.6"
	mevBoostUrlEnvVar           string = "MEV_BOOST_URL"
	mevBoostRelaysEnvVar        string = "MEV_BOOST_RELAYS"
	mevDocsUrl                  string = "https://docs.rocketpool.net/guides/node/mev.html"
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

	// bloXroute max profit relay
	BloxRouteMaxProfitRelay config.Parameter `yaml:"bloxRouteMaxProfitEnabled,omitempty"`

	// bloXroute regulated relay
	BloxRouteRegulatedRelay config.Parameter `yaml:"bloxRouteRegulatedEnabled,omitempty"`

	// Eden relay
	EdenRelay config.Parameter `yaml:"edenEnabled,omitempty"`

	// Ultra sound relay
	UltrasoundRelay config.Parameter `yaml:"ultrasoundEnabled,omitempty"`

	// Aestus relay
	AestusRelay config.Parameter `yaml:"aestusEnabled,omitempty"`

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
	relays := createDefaultRelays()
	relayMap := map[config.MevRelayID]config.MevRelay{}
	for _, relay := range relays {
		relayMap[relay.ID] = relay
	}

	rpcPortModes := config.PortModes("")

	return &MevBoostConfig{
		Title: "MEV-Boost Settings",

		parentConfig: cfg,

		Mode: config.Parameter{
			ID:                   "mode",
			Name:                 "MEV-Boost Mode",
			Description:          "Choose whether to let the Smartnode manage your MEV-Boost instance (Locally Managed), or if you manage your own outside of the Smartnode stack (Externally Managed).",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []config.ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage the MEV-Boost client for you",
				Value:       config.Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing MEV-Boost client that you manage on your own",
				Value:       config.Mode_External,
			}},
		},

		SelectionMode: config.Parameter{
			ID:                   "selectionMode",
			Name:                 "Selection Mode",
			Description:          "Select how the TUI shows you the options for which MEV relays to enable.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.MevSelectionMode_Profile},
			AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
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

		EnableRegulatedAllMev:   generateProfileParameter("enableRegulatedAllMev", relays, true),
		EnableUnregulatedAllMev: generateProfileParameter("enableUnregulatedAllMev", relays, false),

		// Explicit relay params
		FlashbotsRelay:          generateRelayParameter("flashbotsEnabled", relayMap[config.MevRelayID_Flashbots]),
		BloxRouteMaxProfitRelay: generateRelayParameter("bloxRouteMaxProfitEnabled", relayMap[config.MevRelayID_BloxrouteMaxProfit]),
		BloxRouteRegulatedRelay: generateRelayParameter("bloxRouteRegulatedEnabled", relayMap[config.MevRelayID_BloxrouteRegulated]),
		EdenRelay:               generateRelayParameter("edenEnabled", relayMap[config.MevRelayID_Eden]),
		UltrasoundRelay:         generateRelayParameter("ultrasoundEnabled", relayMap[config.MevRelayID_Ultrasound]),
		AestusRelay:             generateRelayParameter("aestusEnabled", relayMap[config.MevRelayID_Aestus]),

		Port: config.Parameter{
			ID:                   "port",
			Name:                 "Port",
			Description:          "The port that MEV-Boost should serve its API on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: uint16(18550)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		OpenRpcPort: config.Parameter{
			ID:                   "openRpcPort",
			Name:                 "Expose API Port",
			Description:          "Expose the API port to other processes on your machine, or to your local network so other local machines can access MEV-Boost's API.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.RPC_Closed},
			AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options:              rpcPortModes,
		},

		ContainerTag: config.Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the MEV-Boost container you want to use on Docker Hub.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: mevBoostTag},
			AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: config.Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to MEV-Boost, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ExternalUrl: config.Parameter{
			ID:                   "externalUrl",
			Name:                 "External URL",
			Description:          "The URL of the external MEV-Boost client or provider",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2},
			EnvironmentVariables: []string{mevBoostUrlEnvVar},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
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
		&cfg.BloxRouteMaxProfitRelay,
		&cfg.BloxRouteRegulatedRelay,
		&cfg.EdenRelay,
		&cfg.UltrasoundRelay,
		&cfg.AestusRelay,
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
		_, exists := relay.Urls[currentNetwork]
		if !exists {
			continue
		}
		regulatedAllMev = regulatedAllMev || relay.Regulated
		unregulatedAllMev = unregulatedAllMev || !relay.Regulated
	}

	return regulatedAllMev, unregulatedAllMev
}

// Get the relays that are available for the current network
func (cfg *MevBoostConfig) GetAvailableRelays() []config.MevRelay {
	relays := []config.MevRelay{}
	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	for _, relay := range cfg.relays {
		_, exists := relay.Urls[currentNetwork]
		if !exists {
			continue
		}
		relays = append(relays, relay)
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
			_, exists := relay.Urls[currentNetwork]
			if !exists {
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
		if cfg.FlashbotsRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_Flashbots].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_Flashbots])
			}
		}
		if cfg.BloxRouteMaxProfitRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_BloxrouteMaxProfit].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_BloxrouteMaxProfit])
			}
		}
		if cfg.BloxRouteRegulatedRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_BloxrouteRegulated].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_BloxrouteRegulated])
			}
		}
		if cfg.EdenRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_Eden].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_Eden])
			}
		}
		if cfg.UltrasoundRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_Ultrasound].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_Ultrasound])
			}
		}
		if cfg.AestusRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_Aestus].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_Aestus])
			}
		}
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

// Create the default MEV relays
func createDefaultRelays() []config.MevRelay {
	relays := []config.MevRelay{
		// Flashbots
		{
			ID:          config.MevRelayID_Flashbots,
			Name:        "Flashbots",
			Description: "Flashbots is the developer of MEV-Boost, and one of the best-known and most trusted relays in the space.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xac6e77dfe25ecd6110b8e780608cce0dab71fdd5ebea22a16c0205200f2f8e2e3ad3b71d3499c54ad14d6c21b41a37ae@boost-relay.flashbots.net?id=rocketpool",
				config.Network_Prater:  "https://0xafa4c6985aa049fb79dd37010438cfebeb0f2bd42b115b89dd678dab0670c1de38da0c4e9138c9290a398ecd9a0b3110@builder-relay-goerli.flashbots.net?id=rocketpool",
				config.Network_Devnet:  "https://0xafa4c6985aa049fb79dd37010438cfebeb0f2bd42b115b89dd678dab0670c1de38da0c4e9138c9290a398ecd9a0b3110@builder-relay-goerli.flashbots.net?id=rocketpool",
			},
			Regulated: true,
		},

		// bloXroute Max Profit
		{
			ID:          config.MevRelayID_BloxrouteMaxProfit,
			Name:        "bloXroute Max Profit",
			Description: "Select this to enable the \"max profit\" relay from bloXroute.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0x8b5d2e73e2a3a55c6c87b8b6eb92e0149a125c852751db1422fa951e42a09b82c142c3ea98d0d9930b056a3bc9896b8f@bloxroute.max-profit.blxrbdn.com?id=rocketpool",
				config.Network_Prater:  "https://0x821f2a65afb70e7f2e820a925a9b4c80a159620582c1766b1b09729fec178b11ea22abb3a51f07b288be815a1a2ff516@bloxroute.max-profit.builder.goerli.blxrbdn.com?id=rocketpool",
				config.Network_Devnet:  "https://0x821f2a65afb70e7f2e820a925a9b4c80a159620582c1766b1b09729fec178b11ea22abb3a51f07b288be815a1a2ff516@bloxroute.max-profit.builder.goerli.blxrbdn.com?id=rocketpool",
			},
			Regulated: false,
		},

		// bloXroute Regulated
		{
			ID:          config.MevRelayID_BloxrouteRegulated,
			Name:        "bloXroute Regulated",
			Description: "Select this to enable the \"regulated\" relay from bloXroute.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xb0b07cd0abef743db4260b0ed50619cf6ad4d82064cb4fbec9d3ec530f7c5e6793d9f286c4e082c0244ffb9f2658fe88@bloxroute.regulated.blxrbdn.com?id=rocketpool",
			},
			Regulated: true,
		},

		// Eden
		{
			ID:          config.MevRelayID_Eden,
			Name:        "Eden Network",
			Description: "Eden Network is the home of Eden Relay, a block building hub focused on optimising block rewards for validators.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xb3ee7afcf27f1f1259ac1787876318c6584ee353097a50ed84f51a1f21a323b3736f271a895c7ce918c038e4265918be@relay.edennetwork.io?id=rocketpool",
				config.Network_Prater:  "https://0xaa1488eae4b06a1fff840a2b6db167afc520758dc2c8af0dfb57037954df3431b747e2f900fe8805f05d635e9a29717b@relay-goerli.edennetwork.io?id=rocketpool",
				config.Network_Devnet:  "https://0xaa1488eae4b06a1fff840a2b6db167afc520758dc2c8af0dfb57037954df3431b747e2f900fe8805f05d635e9a29717b@relay-goerli.edennetwork.io?id=rocketpool",
			},
			Regulated: true,
		},

		// Ultrasound
		{
			ID:          config.MevRelayID_Ultrasound,
			Name:        "Ultra Sound",
			Description: "The ultra sound relay is a credibly-neutral and permissionless relay — a public good from the ultrasound.money team.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xa1559ace749633b997cb3fdacffb890aeebdb0f5a3b6aaa7eeeaf1a38af0a8fe88b9e4b1f61f236d2e64d95733327a62@relay.ultrasound.money?id=rocketpool",
				config.Network_Prater:  "https://0xb1559beef7b5ba3127485bbbb090362d9f497ba64e177ee2c8e7db74746306efad687f2cf8574e38d70067d40ef136dc@relay-stag.ultrasound.money?id=rocketpool",
				config.Network_Devnet:  "https://0xb1559beef7b5ba3127485bbbb090362d9f497ba64e177ee2c8e7db74746306efad687f2cf8574e38d70067d40ef136dc@relay-stag.ultrasound.money?id=rocketpool",
			},
			Regulated: false,
		},

		// Aestus
		{
			ID:          config.MevRelayID_Aestus,
			Name:        "Aestus",
			Description: "The Aestus MEV-Boost Relay is an independent and non-censoring relay. It is committed to neutrality and the development of a healthy MEV-Boost ecosystem.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xa15b52576bcbf1072f4a011c0f99f9fb6c66f3e1ff321f11f461d15e31b1cb359caa092c71bbded0bae5b5ea401aab7e@aestus.live?id=rocketpool",
				config.Network_Prater:  "https://0xab78bf8c781c58078c3beb5710c57940874dd96aef2835e7742c866b4c7c0406754376c2c8285a36c630346aa5c5f833@goerli.aestus.live?id=rocketpool",
				config.Network_Devnet:  "https://0xab78bf8c781c58078c3beb5710c57940874dd96aef2835e7742c866b4c7c0406754376c2c8285a36c630346aa5c5f833@goerli.aestus.live?id=rocketpool",
			},
			Regulated: false,
		},
	}

	return relays
}

// Generate one of the profile parameters
func generateProfileParameter(id string, relays []config.MevRelay, regulated bool) config.Parameter {
	name := "Enable "
	description := fmt.Sprintf("[lime]NOTE: You can enable multiple options.\n\nTo learn more about MEV, please visit %s.\n\n[white]", mevDocsUrl)

	if regulated {
		name += "Regulated "
		description += RegulatedRelayDescription
	} else {
		name += "Unregulated "
		description += UnregulatedRelayDescription
	}

	// Generate the Mainnet description
	mainnetRelays := []string{}
	mainnetDescription := description + "\n\nRelays: "
	for _, relay := range relays {
		_, exists := relay.Urls[config.Network_Mainnet]
		if !exists {
			continue
		}
		if relay.Regulated == regulated {
			mainnetRelays = append(mainnetRelays, relay.Name)
		}
	}
	mainnetDescription += strings.Join(mainnetRelays, ", ")

	// Generate the Prater description
	praterRelays := []string{}
	praterDescription := description + "\n\nRelays:\n"
	for _, relay := range relays {
		_, exists := relay.Urls[config.Network_Prater]
		if !exists {
			continue
		}
		if relay.Regulated == regulated {
			praterRelays = append(praterRelays, relay.Name)
		}
	}
	praterDescription += strings.Join(praterRelays, ", ")

	return config.Parameter{
		ID:                   id,
		Name:                 name,
		Description:          mainnetDescription,
		Type:                 config.ParameterType_Bool,
		Default:              map[config.Network]interface{}{config.Network_All: false},
		AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
		EnvironmentVariables: []string{},
		CanBeBlank:           false,
		OverwriteOnUpgrade:   false,
		DescriptionsByNetwork: map[config.Network]string{
			config.Network_Mainnet: mainnetDescription,
			config.Network_Prater:  praterDescription,
		},
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
		ID:                   id,
		Name:                 fmt.Sprintf("Enable %s", relay.Name),
		Description:          description,
		Type:                 config.ParameterType_Bool,
		Default:              map[config.Network]interface{}{config.Network_All: false},
		AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
		EnvironmentVariables: []string{},
		CanBeBlank:           false,
		OverwriteOnUpgrade:   false,
	}
}
