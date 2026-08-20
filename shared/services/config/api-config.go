package config

import (
	"os"
	"path/filepath"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	apiPortID         string = "apiPort"
	openApiPortID     string = "openApiPort"
	apiTokenID        string = "apiToken"
	tokenScopeID      string = "tokenScope"
	rateLimitID       string = "rateLimit"
	apiTokenFile      string = "api-token"
	defaultApiPort    uint16 = 8280
	defaultOpenPort          = config.RPC_OpenLocalhost
	defaultTokenScope        = config.APITokenScope_All
	defaultRateLimit  uint16 = 5
)

// Configuration for the Smart Node HTTP API
type ApiConfig struct {
	Title string `yaml:"-"`

	parent *RocketPoolConfig

	// Port the node's HTTP API server listens on
	ApiPort config.Parameter `yaml:"apiPort,omitempty"`

	// How the API port is published
	OpenApiPort config.Parameter `yaml:"openApiPort,omitempty"`

	// Bearer token (stored in a sidecar file)
	APIToken config.Parameter `yaml:"-"`

	// Which routes require the bearer token
	TokenScope config.Parameter `yaml:"tokenScope,omitempty"`

	// Maximum requests per second (0 disables the limit)
	RateLimit config.Parameter `yaml:"rateLimit,omitempty"`
}

func NewApiConfig(cfg *RocketPoolConfig) *ApiConfig {
	portModes := config.PortModes("Allow connections from external hosts. The Smart Node API can export your wallet, send funds, and change node settings. Do not expose this to the public internet without TLS (put it behind a reverse proxy). Trusted LAN only otherwise.")
	portModes[0].Description = "Do not publish the API port to the host. The rocketpool CLI on this machine will not be able to reach the API."

	return &ApiConfig{
		Title:  "API Settings",
		parent: cfg,

		ApiPort: config.Parameter{
			ID:                 apiPortID,
			Name:               "API Port",
			Description:        "The port your Smartnode's HTTP API server should listen on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultApiPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenApiPort: config.Parameter{
			ID:                 openApiPortID,
			Name:               "Expose API Port",
			Description:        "Expose the Smart Node HTTP API to other processes on your machine, or to your local network so other machines can access it. Closed means the rocketpool CLI on this host cannot reach the API.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultOpenPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            portModes,
		},

		APIToken: config.Parameter{
			ID:   apiTokenID,
			Name: "API Token",
			Description: "Bearer token required by the API according to Token Requirement below. Treat this like a password: copy it to a password manager. " +
				"Clients send `Authorization: Bearer <token>`. The rocketpool CLI on this machine sends it automatically. " +
				"Clearing the field regenerates a new token on save.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
			Sensitive:          true,
		},

		TokenScope: config.Parameter{
			ID:                 tokenScopeID,
			Name:               "Token Requirement",
			Description:        "Which API routes require the bearer token. Sensitive endpoints include every route that submits an on-chain or validator-exit transaction, plus wallet operations and similar mutating actions. Status, balances, and gas estimates (can-*) stay open if you choose sensitive-only.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultTokenScope},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Require token for all endpoints",
				Description: "Every API request except /healthz must include the bearer token.",
				Value:       config.APITokenScope_All,
			}, {
				Name:        "Require token for sensitive endpoints only",
				Description: "Only mutating routes need the token: every endpoint that submits a transaction, plus wallet operations, exiting validators, staking, withdrawals, DAO votes, and similar. Read-only status and can-* estimates do not.",
				Value:       config.APITokenScope_Sensitive,
			}},
		},

		RateLimit: config.Parameter{
			ID:                 rateLimitID,
			Name:               "API Rate Limit",
			Description:        "Maximum number of API requests per second. The default is 5. Set to 0 to disable rate limiting.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultRateLimit},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

func (cfg *ApiConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.ApiPort,
		&cfg.OpenApiPort,
		&cfg.APIToken,
		&cfg.TokenScope,
		&cfg.RateLimit,
	}
}

func (cfg *ApiConfig) GetConfigTitle() string {
	return cfg.Title
}

// GetAPITokenPath is the token file path as seen by the node daemon.
func (cfg *ApiConfig) GetAPITokenPath() string {
	if cfg.parent != nil && cfg.parent.IsNativeMode {
		return tokenPath(cfg.parent.Smartnode.DataPath.Value.(string))
	}
	return tokenPath(DaemonDataPath)
}

// GetAPITokenPathInCLI is the token file path as seen by the host CLI / TUI.
func (cfg *ApiConfig) GetAPITokenPathInCLI() string {
	return tokenPath(cfg.parent.Smartnode.DataPath.Value.(string))
}

func tokenPath(dataDir string) string {
	return filepath.Join(os.ExpandEnv(dataDir), apiTokenFile)
}
