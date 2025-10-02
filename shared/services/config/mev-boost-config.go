package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	mevBoostTagProd             string = "flashbots/mev-boost:1.9"
	mevBoostTagTest             string = "flashbots/mev-boost:1.10a5"
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
			ID:                 "mode",
			Name:               "MEV-Boost Mode",
			Description:        "Choose whether to let the Smartnode manage your MEV-Boost instance (Locally Managed), or if you manage your own outside of the Smartnode stack (Externally Managed).",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
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

		EnableRegulatedAllMev:   generateProfileParameter("enableRegulatedAllMev", relays, true),
		EnableUnregulatedAllMev: generateProfileParameter("enableUnregulatedAllMev", relays, false),

		// Explicit relay params
		FlashbotsRelay:          generateRelayParameter("flashbotsEnabled", relayMap[config.MevRelayID_Flashbots]),
		BloxRouteMaxProfitRelay: generateRelayParameter("bloxRouteMaxProfitEnabled", relayMap[config.MevRelayID_BloxrouteMaxProfit]),
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
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the MEV-Boost container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: mevBoostTagProd,
				config.Network_Devnet:  mevBoostTagTest,
				config.Network_Testnet: mevBoostTagTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to MEV-Boost, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
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
		&cfg.BloxRouteMaxProfitRelay,
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
		relays = cfg.maybeAddRelay(relays, cfg.BloxRouteMaxProfitRelay, config.MevRelayID_BloxrouteMaxProfit, currentNetwork)
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
				config.Network_Testnet: "https://0xafa4c6985aa049fb79dd37010438cfebeb0f2bd42b115b89dd678dab0670c1de38da0c4e9138c9290a398ecd9a0b3110@boost-relay-hoodi.flashbots.net?id=rocketpool",
				config.Network_Devnet:  "https://0xafa4c6985aa049fb79dd37010438cfebeb0f2bd42b115b89dd678dab0670c1de38da0c4e9138c9290a398ecd9a0b3110@boost-relay-hoodi.flashbots.net?id=rocketpool",
			},
			Regulated: true,
		},

		// bloXroute Max Profit
		{
			ID:          config.MevRelayID_BloxrouteMaxProfit,
			Name:        "bloXroute Max Profit",
			Description: "Select this to enable the relay from bloXroute (formerly) known as \"Max Profit\". (Both bloXroute relays propagate the same transactions...)",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0x8b5d2e73e2a3a55c6c87b8b6eb92e0149a125c852751db1422fa951e42a09b82c142c3ea98d0d9930b056a3bc9896b8f@bloxroute.max-profit.blxrbdn.com?id=rocketpool",
				config.Network_Testnet: "https://0x821f2a65afb70e7f2e820a925a9b4c80a159620582c1766b1b09729fec178b11ea22abb3a51f07b288be815a1a2ff516@bloxroute.hoodi.blxrbdn.com?id=rocketpool",
				config.Network_Devnet:  "https://0x821f2a65afb70e7f2e820a925a9b4c80a159620582c1766b1b09729fec178b11ea22abb3a51f07b288be815a1a2ff516@bloxroute.hoodi.blxrbdn.com?id=rocketpool",
			},
			Regulated: true,
		},

		// bloXroute Regulated
		{
			ID:          config.MevRelayID_BloxrouteRegulated,
			Name:        "bloXroute Regulated",
			Description: "Select this to enable the relay from bloXroute (formerly) known as \"Regulated\". (Both bloXroute relays propagate the same transactions...)",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xb0b07cd0abef743db4260b0ed50619cf6ad4d82064cb4fbec9d3ec530f7c5e6793d9f286c4e082c0244ffb9f2658fe88@bloxroute.regulated.blxrbdn.com?id=rocketpool",
			},
			Regulated: true,
		},

		// Ultrasound
		{
			ID:          config.MevRelayID_Ultrasound,
			Name:        "Ultra Sound (non-filtering)",
			Description: "The ultra sound relay is a credibly-neutral and permissionless relay — a public good from the ultrasound.money team.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xa1559ace749633b997cb3fdacffb890aeebdb0f5a3b6aaa7eeeaf1a38af0a8fe88b9e4b1f61f236d2e64d95733327a62@relay.ultrasound.money?id=rocketpool",
				config.Network_Testnet: "https://0xb1559beef7b5ba3127485bbbb090362d9f497ba64e177ee2c8e7db74746306efad687f2cf8574e38d70067d40ef136dc@relay-hoodi.ultrasound.money?id=rocketpool",
				config.Network_Devnet:  "https://0xb1559beef7b5ba3127485bbbb090362d9f497ba64e177ee2c8e7db74746306efad687f2cf8574e38d70067d40ef136dc@relay-hoodi.ultrasound.money?id=rocketpool",
			},
			Regulated: false,
		},

		// Ultrasound Filtered
		{
			ID:          config.MevRelayID_UltrasoundFiltered,
			Name:        "Ultra Sound (filtering)",
			Description: "The ultra sound relay is a credibly-neutral and permissionless relay — a public good from the ultrasound.money team. This is the filtering version.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xa1559ace749633b997cb3fdacffb890aeebdb0f5a3b6aaa7eeeaf1a38af0a8fe88b9e4b1f61f236d2e64d95733327a62@relay-filtered.ultrasound.money?id=rocketpool",
				config.Network_Testnet: "https://0xb1559beef7b5ba3127485bbbb090362d9f497ba64e177ee2c8e7db74746306efad687f2cf8574e38d70067d40ef136dc@relay-filtered-hoodi.ultrasound.money?id=rocketpool",
				config.Network_Devnet:  "https://0xb1559beef7b5ba3127485bbbb090362d9f497ba64e177ee2c8e7db74746306efad687f2cf8574e38d70067d40ef136dc@relay-filtered-hoodi.ultrasound.money?id=rocketpool",
			},
			Regulated: true,
		},

		// Aestus
		{
			ID:          config.MevRelayID_Aestus,
			Name:        "Aestus",
			Description: "The Aestus MEV-Boost Relay is an independent and non-censoring relay. It is committed to neutrality and the development of a healthy MEV-Boost ecosystem.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xa15b52576bcbf1072f4a011c0f99f9fb6c66f3e1ff321f11f461d15e31b1cb359caa092c71bbded0bae5b5ea401aab7e@aestus.live?id=rocketpool",
				config.Network_Testnet: "https://0x98f0ef62f00780cf8eb06701a7d22725b9437d4768bb19b363e882ae87129945ec206ec2dc16933f31d983f8225772b6@hoodi.aestus.live?id=rocketpool",
				config.Network_Devnet:  "https://0x98f0ef62f00780cf8eb06701a7d22725b9437d4768bb19b363e882ae87129945ec206ec2dc16933f31d983f8225772b6@hoodi.aestus.live?id=rocketpool",
			},
			Regulated: false,
		},

		// Titan Global
		{
			ID:          config.MevRelayID_TitanGlobal,
			Name:        "Titan Global (non-filtering)",
			Description: "Titan Relay is a neutral, Rust-based MEV-Boost Relay optimized for low latency throughput, geographical distribution, and robustness. Select this to enable the \"non-filtering\" relay from Titan.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0x8c4ed5e24fe5c6ae21018437bde147693f68cda427cd1122cf20819c30eda7ed74f72dece09bb313f2a1855595ab677d@global.titanrelay.xyz",
				config.Network_Testnet: "https://0xaa58208899c6105603b74396734a6263cc7d947f444f396a90f7b7d3e65d102aec7e5e5291b27e08d02c50a050825c2f@hoodi.titanrelay.xyz",
				config.Network_Devnet:  "https://0xaa58208899c6105603b74396734a6263cc7d947f444f396a90f7b7d3e65d102aec7e5e5291b27e08d02c50a050825c2f@hoodi.titanrelay.xyz",
			},
			Regulated: false,
		},

		// Titan Regional
		{
			ID:          config.MevRelayID_TitanRegional,
			Name:        "Titan Regional (filtering)",
			Description: "Titan Relay is a neutral, Rust-based MEV-Boost Relay optimized for low latency throughput, geographical distribution, and robustness. Select this to enable the \"filtering\" relay from Titan.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0x8c4ed5e24fe5c6ae21018437bde147693f68cda427cd1122cf20819c30eda7ed74f72dece09bb313f2a1855595ab677d@regional.titanrelay.xyz",
				config.Network_Testnet: "",
				config.Network_Devnet:  "",
			},
			Regulated: true,
		},

		// BTCS OFAC+
		{
			ID:          config.MevRelayID_BTCSOfac,
			Name:        "BTCS OFAC+",
			Description: "Select this to enable the BTCS OFAC+ regulated relay.",
			Urls: map[config.Network]string{
				config.Network_Mainnet: "https://0xb66921e917a8f4cfc3c52e10c1e5c77b1255693d9e6ed6f5f444b71ca4bb610f2eff4fa98178efbf4dd43a30472c497e@relay.btcs.com",
				config.Network_Testnet: "",
				config.Network_Devnet:  "",
			},
			Regulated: true,
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
		if !relay.Urls.UrlExists(config.Network_Mainnet) {
			continue
		}
		if relay.Regulated == regulated {
			mainnetRelays = append(mainnetRelays, relay.Name)
		}
	}
	mainnetDescription += strings.Join(mainnetRelays, ", ")

	// Generate the Testnet description
	testnetRelays := []string{}
	testnetDescription := description + "\n\nRelays: "
	for _, relay := range relays {
		if !relay.Urls.UrlExists(config.Network_Testnet) {
			continue
		}
		if relay.Regulated == regulated {
			testnetRelays = append(testnetRelays, relay.Name)
		}
	}
	testnetDescription += strings.Join(testnetRelays, ", ")

	return config.Parameter{
		ID:                 id,
		Name:               name,
		Description:        mainnetDescription,
		Type:               config.ParameterType_Bool,
		Default:            map[config.Network]interface{}{config.Network_All: false},
		AffectsContainers:  []config.ContainerID{config.ContainerID_MevBoost},
		CanBeBlank:         false,
		OverwriteOnUpgrade: false,
		DescriptionsByNetwork: map[config.Network]string{
			config.Network_Mainnet: mainnetDescription,
			config.Network_Testnet: testnetDescription,
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
