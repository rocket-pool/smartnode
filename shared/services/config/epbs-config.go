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

	MinBid              config.Parameter `yaml:"minBid,omitempty"`
	MaxExecutionPayment config.Parameter `yaml:"maxExecutionPayment,omitempty"`
	BuilderBoostFactor  config.Parameter `yaml:"builderBoostFactor,omitempty"`
	Aestus              config.Parameter `yaml:"aestusEnabled,omitempty"`
	Flashbots           config.Parameter `yaml:"flashbotsEnabled,omitempty"`
	Nflaig              config.Parameter `yaml:"nflaigEnabled,omitempty"`
	Titan               config.Parameter `yaml:"titanEnabled,omitempty"`
	Ultrasound          config.Parameter `yaml:"ultrasoundEnabled,omitempty"`
	Pk910               config.Parameter `yaml:"pk910Enabled,omitempty"`

	parentConfig *RocketPoolConfig `yaml:"-"`
}

func NewEpbsConfig(cfg *RocketPoolConfig) *EpbsConfig {
	epbs := &EpbsConfig{
		Title:        "Builder Settings",
		parentConfig: cfg,
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
		BuilderBoostFactor: config.Parameter{
			ID:                 "builderBoostFactor",
			Name:               "Builder Boost Factor",
			Description:        "The percentage applied to a builder bid when comparing it with a locally built payload. 100 compares them equally. A value below 100 favors the local payload, and 0 uses the local payload whenever it is available. For example, 90 selects a builder bid only when it is about 11% higher than the local payload.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: uint64(100)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
	for _, option := range epbs.builderOptions() {
		*option.parameter = config.Parameter{
			ID:                 option.id,
			Name:               fmt.Sprintf("Enable %s", option.name),
			Description:        fmt.Sprintf("Select %s for the ePBS builder list. You can select multiple operators.\n\n%s", option.name, option.url),
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		}
	}
	return epbs
}

type epbsBuilderOption struct {
	id        string
	name      string
	url       string
	parameter *config.Parameter
}

func (cfg *EpbsConfig) builderOptions() []epbsBuilderOption {
	return []epbsBuilderOption{
		{"aestusEnabled", "Aestus", "https://hoodi.aestus.live", &cfg.Aestus},
		{"flashbotsEnabled", "Flashbots", "https://boost-relay-hoodi.flashbots.net", &cfg.Flashbots},
		{"nflaigEnabled", "nflaig", "https://builder-hoodi.nflaig.dev", &cfg.Nflaig},
		{"titanEnabled", "Titan", "https://hoodi.titanrelay.xyz", &cfg.Titan},
		{"ultrasoundEnabled", "Ultra Sound", "https://relay-hoodi.ultrasound.money", &cfg.Ultrasound},
		{"pk910Enabled", "pk910", "https://builder-hoodi.pk910.de", &cfg.Pk910},
	}
}

func (cfg *EpbsConfig) GetParameters() []*config.Parameter {
	params := []*config.Parameter{&cfg.MinBid, &cfg.MaxExecutionPayment, &cfg.BuilderBoostFactor}
	for _, option := range cfg.builderOptions() {
		params = append(params, option.parameter)
	}
	return params
}

func (cfg *EpbsConfig) GetConfigTitle() string {
	return cfg.Title
}

func (cfg *EpbsConfig) hoodiBuildersAvailable() bool {
	info := cfg.parentConfig.GetNetworkInfo()
	return info != nil && info.BeaconNetwork == "hoodi"
}

// GetAvailableBuilderParameters returns the Hoodi ePBS builder catalog.
func (cfg *EpbsConfig) GetAvailableBuilderParameters() []*config.Parameter {
	if !cfg.hoodiBuildersAvailable() {
		return nil
	}
	params := []*config.Parameter{}
	for _, option := range cfg.builderOptions() {
		params = append(params, option.parameter)
	}
	return params
}

// GetBuilderUrls returns the selected endpoints in catalog order as scheme://host.
func (cfg *EpbsConfig) GetBuilderUrls() ([]string, error) {
	if !cfg.hoodiBuildersAvailable() {
		return nil, nil
	}
	urls := []string{}
	seen := map[string]bool{}
	for _, option := range cfg.builderOptions() {
		if option.parameter.Value != true {
			continue
		}
		endpoint, err := url.Parse(option.url)
		if err != nil || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.Hostname() == "" {
			return nil, fmt.Errorf("ePBS builder endpoint for %s must be an HTTP or HTTPS URL", option.name)
		}
		builderURL := (&url.URL{
			Scheme: endpoint.Scheme,
			Host:   endpoint.Host,
		}).String()
		if !seen[builderURL] {
			urls = append(urls, builderURL)
			seen[builderURL] = true
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

// GetBuilderBoostFactor returns the percentage applied to a builder bid.
// 100 compares a builder bid and a local payload equally.
func (cfg *EpbsConfig) GetBuilderBoostFactor() (string, error) {
	value, ok := cfg.BuilderBoostFactor.Value.(uint64)
	if !ok {
		return "", fmt.Errorf("ePBS builder boost factor must be a whole number from 0 to 18446744073709551615")
	}
	return strconv.FormatUint(value, 10), nil
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
		return "", fmt.Errorf("ePBS %s must not have more than 9 decimal places", name)
	}
	gwei, err := strconv.ParseUint(whole+fraction+strings.Repeat("0", 9-len(fraction)), 10, 64)
	if err != nil {
		return "", fmt.Errorf("ePBS %s must not exceed 18446744073.709551615 ETH", name)
	}
	return strconv.FormatUint(gwei, 10), nil
}
