package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

var builderAmountPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// EpbsConfig stores builder settings shared by all consensus clients.
type EpbsConfig struct {
	Title string `yaml:"-"`

	MinBid                  config.Parameter `yaml:"minBid,omitempty"`
	MaxExecutionPayment     config.Parameter `yaml:"maxExecutionPayment,omitempty"`
	FlashbotsRelay          config.Parameter `yaml:"flashbotsEnabled,omitempty"`
	BloxRouteRegulatedRelay config.Parameter `yaml:"bloxRouteRegulatedEnabled,omitempty"`
	UltrasoundRelay         config.Parameter `yaml:"ultrasoundEnabled,omitempty"`
	UltrasoundFilteredRelay config.Parameter `yaml:"ultrasoundFilteredEnabled,omitempty"`
	AestusRelay             config.Parameter `yaml:"aestusEnabled,omitempty"`
	TitanGlobalRelay        config.Parameter `yaml:"titanGlobalEnabled,omitempty"`
	TitanRegionalRelay      config.Parameter `yaml:"titanRegionalEnabled,omitempty"`
	BtcsOfacRelay           config.Parameter `yaml:"btcsOfacEnabled,omitempty"`

	parentConfig *RocketPoolConfig `yaml:"-"`
	relays       []config.MevRelay `yaml:"-"`
}

func NewEpbsConfig(cfg *RocketPoolConfig) *EpbsConfig {
	epbs := &EpbsConfig{
		Title:        "Builder Settings",
		parentConfig: cfg,
		relays:       createDefaultRelays(cfg.networks),
		MinBid: config.Parameter{
			ID:                 "minBid",
			Name:               "Minimum Bid (ETH)",
			Description:        "The minimum payment to accept from an ePBS builder, in ETH. For example, enter 0.05 for 0.05 ETH. Leave blank to use the consensus client's default.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
		MaxExecutionPayment: config.Parameter{
			ID:                 "maxExecutionPayment",
			Name:               "Max Execution Payment (ETH)",
			Description:        "The maximum execution-layer payment to count toward an ePBS builder bid, in ETH. These payments rely on the builder's promise and are not backed by staked collateral. Set to 0 to count only trustless payments. For example, enter 0.05 to count up to 0.05 ETH of promised execution-layer payments.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: "0"},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
	for _, option := range epbs.relayOptions() {
		for _, relay := range epbs.relays {
			if relay.ID != option.id {
				continue
			}
			*option.parameter = config.Parameter{
				ID:                 option.parameterID,
				Name:               fmt.Sprintf("Enable %s", relay.Name),
				Description:        fmt.Sprintf("Select %s for the ePBS builder list. You can select multiple operators.\n\nUses this operator's existing relay endpoint for the selected network. ePBS availability at that endpoint has not yet been verified.", relay.Name),
				Type:               config.ParameterType_Bool,
				Default:            map[config.Network]interface{}{config.Network_All: false},
				AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			}
		}
	}
	return epbs
}

type epbsRelayOption struct {
	id          config.MevRelayID
	parameterID string
	parameter   *config.Parameter
}

func (cfg *EpbsConfig) relayOptions() []epbsRelayOption {
	return []epbsRelayOption{
		{config.MevRelayID_Flashbots, "flashbotsEnabled", &cfg.FlashbotsRelay},
		{config.MevRelayID_BloxrouteRegulated, "bloxRouteRegulatedEnabled", &cfg.BloxRouteRegulatedRelay},
		{config.MevRelayID_Ultrasound, "ultrasoundEnabled", &cfg.UltrasoundRelay},
		{config.MevRelayID_UltrasoundFiltered, "ultrasoundFilteredEnabled", &cfg.UltrasoundFilteredRelay},
		{config.MevRelayID_Aestus, "aestusEnabled", &cfg.AestusRelay},
		{config.MevRelayID_TitanGlobal, "titanGlobalEnabled", &cfg.TitanGlobalRelay},
		{config.MevRelayID_TitanRegional, "titanRegionalEnabled", &cfg.TitanRegionalRelay},
		{config.MevRelayID_BTCSOfac, "btcsOfacEnabled", &cfg.BtcsOfacRelay},
	}
}

func (cfg *EpbsConfig) GetParameters() []*config.Parameter {
	params := []*config.Parameter{&cfg.MinBid, &cfg.MaxExecutionPayment}
	for _, option := range cfg.relayOptions() {
		params = append(params, option.parameter)
	}
	return params
}

func (cfg *EpbsConfig) GetConfigTitle() string {
	return cfg.Title
}

// GetAvailableRelayParameters uses the same operator catalog and network
// availability as MEV-Boost, with independent selections for ePBS.
func (cfg *EpbsConfig) GetAvailableRelayParameters() []*config.Parameter {
	params := []*config.Parameter{}
	network := cfg.parentConfig.GetNetwork()
	for _, option := range cfg.relayOptions() {
		for _, relay := range cfg.relays {
			if relay.ID == option.id && relay.Urls.UrlExists(network) {
				params = append(params, option.parameter)
			}
		}
	}
	return params
}

// GetBuilderUrls returns the selected network's endpoints in catalog order.
// These are seeded from the relay catalog, not verified ePBS endpoints. Relay
// signing pubkeys in URL userinfo are not builder authentication credentials.
func (cfg *EpbsConfig) GetBuilderUrls() ([]string, error) {
	urls := []string{}
	seen := map[string]bool{}
	network := cfg.parentConfig.GetNetwork()
	for _, option := range cfg.relayOptions() {
		if option.parameter.Value != true {
			continue
		}
		for _, relay := range cfg.relays {
			if relay.ID != option.id || !relay.Urls.UrlExists(network) {
				continue
			}
			endpoint, err := url.Parse(relay.Urls[network])
			if err != nil || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.Hostname() == "" {
				return nil, fmt.Errorf("ePBS builder endpoint for %s must be an HTTP or HTTPS URL", relay.Name)
			}
			endpoint.User = nil
			builderURL := endpoint.String()
			if !seen[builderURL] {
				urls = append(urls, builderURL)
				seen[builderURL] = true
			}
		}
	}
	return urls, nil
}

// GetBuilderUrlString formats the selected endpoints for client CLI arguments.
func (cfg *EpbsConfig) GetBuilderUrlString() (string, error) {
	urls, err := cfg.GetBuilderUrls()
	if err != nil {
		return "", err
	}
	return strings.Join(urls, ","), nil
}

// GetMinBidGwei converts the ETH setting exactly, without floating point or
// rounding. A blank value lets client integrations retain their own defaults.
func (cfg *EpbsConfig) GetMinBidGwei() (string, error) {
	return builderAmountToGwei(cfg.MinBid.Value.(string), cfg.MinBid.Name)
}

// GetMaxExecutionPaymentGwei converts the ETH cap exactly. Blank input uses
// the default of zero, so only trustless payments count toward a builder bid.
func (cfg *EpbsConfig) GetMaxExecutionPaymentGwei() (string, error) {
	gwei, err := builderAmountToGwei(cfg.MaxExecutionPayment.Value.(string), cfg.MaxExecutionPayment.Name)
	if err == nil && gwei == "" {
		return "0", nil
	}
	return gwei, err
}

func builderAmountToGwei(value, name string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if !builderAmountPattern.MatchString(value) {
		return "", fmt.Errorf("ePBS %s must be a non-negative decimal amount, for example 0.05", name)
	}
	whole, fraction, _ := strings.Cut(value, ".")
	fraction = strings.TrimRight(fraction, "0")
	if len(fraction) > 9 {
		return "", fmt.Errorf("ePBS %s", name)
	}
	gwei, err := strconv.ParseUint(whole+fraction+strings.Repeat("0", 9-len(fraction)), 10, 64)
	if err != nil {
		return "", fmt.Errorf("ePBS %s must not exceed 18446744073.709551615 ETH", name)
	}
	return strconv.FormatUint(gwei, 10), nil
}
