package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/rocket-pool/smartnode/shared/services/apitoken"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	apiPortID         string = "apiPort"
	openApiPortID     string = "openApiPort"
	apiTokenID        string = "apiToken"
	tokenCommentID    string = "apiTokenComment"
	tokenScopeID      string = "tokenScope"
	rateLimitID       string = "rateLimit"
	apiTokenFile      string = "api-tokens.json"
	defaultApiPort    uint16 = 8280
	defaultOpenPort          = config.RPC_OpenLocalhost
	defaultTokenScope        = config.APITokenScope_All
	defaultRateLimit  uint16 = 0
)

// Configuration for the Smart Node HTTP API
type ApiConfig struct {
	Title string `yaml:"-"`

	parent *RocketPoolConfig

	// Port the node's HTTP API server listens on
	ApiPort config.Parameter `yaml:"apiPort,omitempty"`

	// How the API port is published
	OpenApiPort config.Parameter `yaml:"openApiPort,omitempty"`

	// Bearer token for the CLI write token (stored in the sidecar JSON file)
	APIToken config.Parameter `yaml:"-"`

	// Comment stored with the CLI write token
	TokenComment config.Parameter `yaml:"-"`

	// Whether Read routes may be called without a bearer token. Server config, not per individual token
	TokenScope config.Parameter `yaml:"tokenScope,omitempty"`

	// Maximum requests per second (0 disables the limit)
	RateLimit config.Parameter `yaml:"rateLimit,omitempty"`

	records []apitoken.Record
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
			Description: "Write-scoped bearer token for the rocketpool CLI. Treat this like a password: copy it to a password manager. " +
				"Clients send `Authorization: Bearer <token>`. The rocketpool CLI on this machine sends it automatically. " +
				"Clearing the field regenerates a new token on save. Additional tokens can be added in data/api-tokens.json.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
			Sensitive:          true,
			SkipUserSettings:   true,
		},

		TokenComment: config.Parameter{
			ID:                 tokenCommentID,
			Name:               "API Token Comment",
			Description:        "A short note stored with this token so you remember why it was created. Not a secret.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: apitoken.CLIComment},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
			SkipUserSettings:   true,
		},

		TokenScope: config.Parameter{
			ID:                 tokenScopeID,
			Name:               "Unauthenticated Reads",
			Description:        "Whether read-only API routes (status, balances, can-* estimates) may be called without a bearer token. Write routes always require a write-scoped token. This does not change the privilege of existing tokens.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultTokenScope},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []config.ParameterOption{{
				Name:        "Require a token for read routes",
				Description: "Every API request except /healthz must include a valid bearer token. Read-scoped tokens may call read routes; write-scoped tokens may call everything.",
				Value:       config.APITokenScope_All,
			}, {
				Name:        "Allow unauthenticated reads",
				Description: "Status, balances, and can-* estimates do not require a token. Wallet operations, transactions, and other write routes still need a write-scoped token.",
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
		&cfg.TokenComment,
		&cfg.TokenScope,
		&cfg.RateLimit,
	}
}

func (cfg *ApiConfig) GetConfigTitle() string {
	return cfg.Title
}

// Records is the loaded token list used by the HTTP API for auth.
func (cfg *ApiConfig) Records() []apitoken.Record {
	return cfg.records
}

// UnauthenticatedReads reports whether Read routes may be called without a token.
func (cfg *ApiConfig) UnauthenticatedReads() bool {
	scope, _ := cfg.TokenScope.Value.(config.APITokenScope)
	return scope == config.APITokenScope_Sensitive
}

// GetAPITokenPath is the sidecar JSON file path for this process.
func (cfg *ApiConfig) GetAPITokenPath() string {
	if cfg.parent != nil && (cfg.parent.IsCLI || cfg.parent.IsNativeMode) {
		return tokenPath(cfg.parent.Smartnode.DataPath.Value.(string))
	}
	return tokenPath(DaemonDataPath)
}

func tokenPath(dataDir string) string {
	dataDir = os.ExpandEnv(dataDir)
	expanded, err := homedir.Expand(dataDir)
	if err == nil {
		dataDir = expanded
	}
	return filepath.Join(dataDir, apiTokenFile)
}

func (cfg *ApiConfig) applyFile(f apitoken.File) {
	cfg.records = f.Tokens
	idx := f.CLIIndex()
	if idx < 0 {
		cfg.APIToken.Value = ""
		return
	}
	cfg.APIToken.Value = f.Tokens[idx].Token.String()
	comment := f.Tokens[idx].Comment
	if comment == "" {
		comment = apitoken.CLIComment
	}
	cfg.TokenComment.Value = comment
}

func (cfg *ApiConfig) fileFromForm() (apitoken.File, error) {
	records := append([]apitoken.Record(nil), cfg.records...)
	comment, _ := cfg.TokenComment.Value.(string)
	comment = strings.TrimSpace(comment)
	if comment == "" {
		comment = apitoken.CLIComment
	}

	tokenStr, _ := cfg.APIToken.Value.(string)
	tokenStr = strings.TrimSpace(tokenStr)
	var tok apitoken.Token
	var err error
	if tokenStr == "" {
		tok, err = apitoken.Generate()
		if err != nil {
			return apitoken.File{}, err
		}
	} else {
		tok, err = apitoken.Parse(tokenStr)
		if err != nil {
			return apitoken.File{}, err
		}
	}

	cli := apitoken.Record{Token: tok, Comment: comment, Scope: apitoken.ScopeWrite}
	idx := (apitoken.File{Tokens: records}).CLIIndex()
	if idx >= 0 {
		records[idx] = cli
	} else {
		records = append([]apitoken.Record{cli}, records...)
	}
	return apitoken.File{Tokens: records}, nil
}
