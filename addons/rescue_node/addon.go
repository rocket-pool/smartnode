package rescue_node

import (
	"fmt"
	"os"

	"github.com/rocket-pool/smartnode/shared/types/addons"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

type RescueNode struct {
	cfg *RescueNodeConfig `yaml:"config,omitempty"`
}

func NewRescueNode() addons.SmartnodeAddon {
	return &RescueNode{
		cfg: NewConfig(),
	}
}

func (r *RescueNode) GetName() string {
	return "Rescue Node"
}

func (r *RescueNode) GetConfig() cfgtypes.Config {
	return r.cfg
}

func (r *RescueNode) GetContainerName() string {
	return ""
}

func (r *RescueNode) GetContainerTag() string {
	return ""
}

func (r *RescueNode) GetDescription() string {
	return `Rocket Rescue Node

The Rocket Rescue Node is a community-run, trust-minimized, and secured fallback node for emergencies and maintenance.

For more information, see https://rescuenode.com`
}

func (r *RescueNode) GetEnabledParameter() *cfgtypes.Parameter {
	return &r.cfg.Enabled
}

func (r *RescueNode) UpdateEnvVars(envVars map[string]string) error {
	return nil
}

func (r *RescueNode) ApplyValidatorOverrides(cc cfgtypes.ConsensusClient) (func(), error) {
	nop := func() {}
	if !r.cfg.Enabled.Value.(bool) {
		return nop, nil
	}

	username := r.cfg.Username.Value.(string)
	password := r.cfg.Password.Value.(string)

	if username == "" || password == "" {
		return nop, fmt.Errorf("Rescue Node can not be enabled without a Username and Password configured.")
	}

	switch cc {
	case cfgtypes.ConsensusClient_Unknown:
		return nop, fmt.Errorf("Unable to generate rescue node URLs for unknown consensus client")
	case cfgtypes.ConsensusClient_Lighthouse,
		cfgtypes.ConsensusClient_Nimbus,
		cfgtypes.ConsensusClient_Lodestar,
		cfgtypes.ConsensusClient_Teku:

		rescueURL := fmt.Sprintf("https://%s:%s@%s.rescuenode.com", username, password, cc)

		oldCC := os.Getenv("CC_API_ENDPOINT")
		cleanup := func() {
			os.Setenv("CC_API_ENDPOINT", oldCC)
		}
		os.Setenv("CC_API_ENDPOINT", rescueURL)
		return cleanup, nil
	case cfgtypes.ConsensusClient_Prysm:
		extraFlags := fmt.Sprintf("--grpc-headers=rprnauth=%s:%s --tls-cert=/etc/ssl/certs/ca-certificates.crt", username, password)

		oldExtraFlags := os.Getenv("VC_ADDITIONAL_FLAGS")
		if oldExtraFlags != "" {
			extraFlags = fmt.Sprintf("%s %s", oldExtraFlags, extraFlags)
		}

		oldCC := os.Getenv("CC_RPC_ENDPOINT")

		cleanup := func() {
			os.Setenv("CC_RPC_ENDPOINT", oldCC)
			os.Setenv("VC_ADDITIONAL_FLAGS", oldExtraFlags)
		}
		os.Setenv("CC_RPC_ENDPOINT", "prysm-grpc.rescuenode.com:443")
		os.Setenv("VC_ADDITIONAL_FLAGS", extraFlags)

		return cleanup, nil
	}

	return nop, nil
}
