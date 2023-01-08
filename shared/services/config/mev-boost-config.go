package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/sys"
)

// Constants
const (
	mevBoostPortableTag         string = "flashbots/mev-boost:1.4.0-portable"
	mevBoostModernTag           string = "flashbots/mev-boost:1.4.0"
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

	// Regulated, no sandwiching
	EnableRegulatedNoSandwich config.Parameter `yaml:"enableRegulatedNoSandwich,omitempty"`

	// Unregulated, all types
	EnableUnregulatedAllMev config.Parameter `yaml:"enableUnregulatedAllMev,omitempty"`

	// Unregulated, no sandwiching
	EnableUnregulatedNoSandwich config.Parameter `yaml:"enableUnregulatedNoSandwich,omitempty"`

	// Flashbots relay
	FlashbotsRelay config.Parameter `yaml:"flashbotsEnabled,omitempty"`

	// bloXroute ethical relay
	BloxRouteEthicalRelay config.Parameter `yaml:"bloxRouteEthicalEnabled,omitempty"`

	// bloXroute max profit relay
	BloxRouteMaxProfitRelay config.Parameter `yaml:"bloxRouteMaxProfitEnabled,omitempty"`

	// bloXroute regulated relay
	BloxRouteRegulatedRelay config.Parameter `yaml:"bloxRouteRegulatedEnabled,omitempty"`

	// Blocknative relay
	BlocknativeRelay config.Parameter `yaml:"blocknativeEnabled,omitempty"`

	// Eden relay
	EdenRelay config.Parameter `yaml:"edenEnabled,omitempty"`

	// Ultra sound relay
	UltrasoundRelay config.Parameter `yaml:"ultrasoundEnabled,omitempty"`

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

		EnableRegulatedAllMev:       generateProfileParameter("enableRegulatedAllMev", relays, true, false),
		EnableRegulatedNoSandwich:   generateProfileParameter("enableRegulatedNoSandwich", relays, true, true),
		EnableUnregulatedAllMev:     generateProfileParameter("enableUnregulatedAllMev", relays, false, false),
		EnableUnregulatedNoSandwich: generateProfileParameter("enableUnregulatedNoSandwich", relays, false, true),

		// Explicit relay params
		FlashbotsRelay:          generateRelayParameter("flashbotsEnabled", relayMap[config.MevRelayID_Flashbots]),
		BloxRouteMaxProfitRelay: generateRelayParameter("bloxRouteMaxProfitEnabled", relayMap[config.MevRelayID_BloxrouteMaxProfit]),
		BloxRouteEthicalRelay:   generateRelayParameter("bloxRouteEthicalEnabled", relayMap[config.MevRelayID_BloxrouteEthical]),
		BloxRouteRegulatedRelay: generateRelayParameter("bloxRouteRegulatedEnabled", relayMap[config.MevRelayID_BloxrouteRegulated]),
		BlocknativeRelay:        generateRelayParameter("blocknativeEnabled", relayMap[config.MevRelayID_Blocknative]),
		EdenRelay:               generateRelayParameter("edenEnabled", relayMap[config.MevRelayID_Eden]),
		UltrasoundRelay:         generateRelayParameter("ultrasoundEnabled", relayMap[config.MevRelayID_Ultrasound]),

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
			Description:          "Expose the API port to your local network, so other local machines can access MEV-Boost's API.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{config.ContainerID_MevBoost},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the MEV-Boost container you want to use on Docker Hub.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: getMevBoostTag()},
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
		&cfg.EnableRegulatedNoSandwich,
		&cfg.EnableUnregulatedAllMev,
		&cfg.EnableUnregulatedNoSandwich,
		&cfg.FlashbotsRelay,
		&cfg.BloxRouteEthicalRelay,
		&cfg.BloxRouteMaxProfitRelay,
		&cfg.BloxRouteRegulatedRelay,
		&cfg.BlocknativeRelay,
		&cfg.EdenRelay,
		&cfg.UltrasoundRelay,
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
func (cfg *MevBoostConfig) GetAvailableProfiles() (bool, bool, bool, bool) {
	regulatedAllMev := false
	regulatedNoSandwich := false
	unregulatedAllMev := false
	unregulatedNoSandwich := false

	currentNetwork := cfg.parentConfig.Smartnode.Network.Value.(config.Network)
	for _, relay := range cfg.relays {
		_, exists := relay.Urls[currentNetwork]
		if !exists {
			continue
		}
		regulatedAllMev = regulatedAllMev || (relay.Regulated && !relay.NoSandwiching)
		regulatedNoSandwich = regulatedNoSandwich || (relay.Regulated && relay.NoSandwiching)
		unregulatedAllMev = unregulatedAllMev || (!relay.Regulated && !relay.NoSandwiching)
		unregulatedNoSandwich = unregulatedNoSandwich || (!relay.Regulated && relay.NoSandwiching)
	}

	return regulatedAllMev, regulatedNoSandwich, unregulatedAllMev, unregulatedNoSandwich
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
				if relay.NoSandwiching {
					if cfg.EnableRegulatedNoSandwich.Value == true {
						relays = append(relays, relay)
					}
				} else {
					if cfg.EnableRegulatedAllMev.Value == true {
						relays = append(relays, relay)
					}
				}
			} else {
				if relay.NoSandwiching {
					if cfg.EnableUnregulatedNoSandwich.Value == true {
						relays = append(relays, relay)
					}
				} else {
					if cfg.EnableUnregulatedAllMev.Value == true {
						relays = append(relays, relay)
					}
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
		if cfg.BloxRouteEthicalRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_BloxrouteEthical].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_BloxrouteEthical])
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
		if cfg.BlocknativeRelay.Value == true {
			_, exists := cfg.relayMap[config.MevRelayID_Blocknative].Urls[currentNetwork]
			if exists {
				relays = append(relays, cfg.relayMap[config.MevRelayID_Blocknative])
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
			Regulated:     true,
			NoSandwiching: false,
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
			Regulated:     false,
			NoSandwiching: false,
		},

		// bloXroute Ethical
		{
			ID:          config.MevRelayID_BloxrouteEthical,
			Name:        "bloXroute Ethical",
			Description: "Select this to enable the \"ethical\" relay from bloXroute.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xad0a8bb54565c2211cee576363f3a347089d2f07cf72679d16911d740262694cadb62d7fd7483f27afd714ca0f1b9118@bloxroute.ethical.blxrbdn.com?id=rocketpool",
			},
			Regulated:     false,
			NoSandwiching: true,
		},

		// bloXroute Regulated
		{
			ID:          config.MevRelayID_BloxrouteRegulated,
			Name:        "bloXroute Regulated",
			Description: "Select this to enable the \"regulated\" relay from bloXroute.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xb0b07cd0abef743db4260b0ed50619cf6ad4d82064cb4fbec9d3ec530f7c5e6793d9f286c4e082c0244ffb9f2658fe88@bloxroute.regulated.blxrbdn.com?id=rocketpool",
			},
			Regulated:     true,
			NoSandwiching: false,
		},

		// Blocknative
		{
			ID:          config.MevRelayID_Blocknative,
			Name:        "Blocknative",
			Description: "Blocknative is a large blockchain infrastructure company that provides a popular MEV relay.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0x9000009807ed12c1f08bf4e81c6da3ba8e3fc3d953898ce0102433094e5f22f21102ec057841fcb81978ed1ea0fa8246@builder-relay-mainnet.blocknative.com?id=rocketpool",
				config.Network_Prater:  "https://0x8f7b17a74569b7a57e9bdafd2e159380759f5dc3ccbd4bf600414147e8c4e1dc6ebada83c0139ac15850eb6c975e82d0@builder-relay-goerli.blocknative.com?id=rocketpool",
				config.Network_Devnet:  "https://0x8f7b17a74569b7a57e9bdafd2e159380759f5dc3ccbd4bf600414147e8c4e1dc6ebada83c0139ac15850eb6c975e82d0@builder-relay-goerli.blocknative.com?id=rocketpool",
			},
			Regulated:     true,
			NoSandwiching: false,
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
			Regulated:     true,
			NoSandwiching: false,
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
			Regulated:     false,
			NoSandwiching: false,
		},
	}

	return relays
}

// Generate one of the profile parameters
func generateProfileParameter(id string, relays []config.MevRelay, regulated bool, noSandwiching bool) config.Parameter {
	name := "Enable "
	description := fmt.Sprintf("[lime]NOTE: You can enable multiple options.\n\nTo learn more about MEV, please visit %s.\n\n[white]", mevDocsUrl)

	if regulated {
		name += "Regulated "
		description += RegulatedRelayDescription
	} else {
		name += "Unregulated "
		description += UnregulatedRelayDescription
	}

	if noSandwiching {
		name += "(No Sandwiching)"
		description += NoSandwichRelayDescription
	} else {
		name += "(All MEV Types)"
		description += AllMevRelayDescription
	}

	// Generate the Mainnet description
	mainnetRelays := []string{}
	mainnetDescription := description + "\n\nRelays: "
	for _, relay := range relays {
		_, exists := relay.Urls[config.Network_Mainnet]
		if !exists {
			continue
		}
		if relay.Regulated == regulated && relay.NoSandwiching == noSandwiching {
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
		if relay.Regulated == regulated && relay.NoSandwiching == noSandwiching {
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

	if relay.NoSandwiching {
		description += "Allows Sandwich Attacks: NO"
	} else {
		description += "Allows Sandwich Attacks: YES"
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

// Get the appropriate MEV-Boost default tag
func getMevBoostTag() string {
	missingFeatures := sys.GetMissingModernCpuFeatures()
	if len(missingFeatures) > 0 {
		return mevBoostPortableTag
	}
	return mevBoostModernTag
}
