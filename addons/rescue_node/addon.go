package rescue_node

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"

	"github.com/rocket-pool/smartnode/addons/rescue_node/pb"
	"github.com/rocket-pool/smartnode/shared/types/addons"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
	colorBlue   string = "\033[36m"

	soloAuthValidity = 10 * time.Hour * 24
	rpAuthValidity   = 15 * time.Hour * 24
)

type credentialDetails struct {
	solo   bool
	issued time.Time
}

func (c *credentialDetails) GetTimeLeft() time.Duration {

	if c.solo == true {
		return time.Until(c.issued.Add(soloAuthValidity))
	}

	return time.Until(c.issued.Add(rpAuthValidity))
}

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

func (r *RescueNode) getCredentialDetails() (*credentialDetails, error) {
	if !r.cfg.Enabled.Value.(bool) {
		panic("getCredentialDetails() should not be called without checking if RN plugin is enabled")
	}

	password := r.cfg.Password.Value.(string)
	if password == "" {
		return nil, fmt.Errorf("Rescue Node enabled, but no Username provided")
	}

	protoBytes, err := base64.URLEncoding.DecodeString(password)
	if err != nil {
		return nil, fmt.Errorf("Rescue Node enabled, but Password is not valid - error decoding base64: %w", err)
	}

	// To avoid a dependency on Rescue Node code, we will parse the protobuf by hand.
	msg := new(pb.AuthenticatedCredential)
	err = proto.Unmarshal(protoBytes, msg)
	if err != nil {
		return nil, fmt.Errorf("Rescue Node enabled, but Password is not valid - error decoding proto: %w", err)
	}

	return &credentialDetails{
		solo:   msg.Credential.OperatorType == pb.OperatorType_OT_SOLO,
		issued: time.Unix(msg.Credential.Timestamp, 0),
	}, nil
}

func (r *RescueNode) getCredentialNodeId() (*common.Address, error) {
	if !r.cfg.Enabled.Value.(bool) {
		panic("getCredentialNodeId() should not be called without checking if RN plugin is enabled")
	}

	username := r.cfg.Username.Value.(string)
	if username == "" {
		return nil, fmt.Errorf("Rescue Node enabled, but no Username provided")
	}

	addr, err := base64.URLEncoding.DecodeString(username)
	if err != nil {
		return nil, fmt.Errorf("Rescue Node enabled, but Username is not valid - error decoding base64: %w", err)
	}

	out := common.BytesToAddress(addr)
	return &out, nil
}

func (r *RescueNode) PrintStatusText(nodeAddr common.Address) {
	if !r.cfg.Enabled.Value.(bool) {
		return
	}

	fmt.Printf("%s=== Rescue Node Add-on Enabled ===%s\n", colorYellow, colorReset)
	// Check the Username
	usernameNodeAddr, err := r.getCredentialNodeId()
	if err != nil {
		fmt.Printf("%s%w%s", colorRed, err, colorReset)
	} else {
		fmt.Printf("Using a credential issued to %s%s%s.\n", colorBlue, usernameNodeAddr.String(), colorReset)
		if !bytes.Equal(usernameNodeAddr.Bytes(), nodeAddr.Bytes()) {
			fmt.Printf("%s - WARNING: This does not match the Node Account!%s\n", colorYellow, colorReset)
		}
	}

	credentialDetails, err := r.getCredentialDetails()
	if err != nil {
		fmt.Printf("%s%w%s", colorRed, err, colorReset)
	} else {
		if credentialDetails.solo {
			fmt.Printf("%s - WARNING: This credential was issued to a solo staker!%s\n", colorYellow, colorReset)
		}

		timeLeft := credentialDetails.GetTimeLeft().Truncate(time.Second)
		if timeLeft < 0 {
			fmt.Printf("%sWARNING: This credential expired %s ago!%s\n", colorRed, timeLeft.String(), colorReset)
		} else if timeLeft < time.Hour*24 {
			fmt.Printf("%sWARNING: This credential expires in %s!%s\n", colorYellow, timeLeft.String(), colorReset)
		} else {
			fmt.Printf("%sThis credential expires in %s.%s\n", colorGreen, timeLeft.String(), colorReset)
		}
	}
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
