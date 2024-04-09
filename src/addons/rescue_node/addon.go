package rescue_node

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"

	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/addons/rescue_node/pb"
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
	if c.solo {
		return time.Until(c.issued.Add(soloAuthValidity))
	}

	return time.Until(c.issued.Add(rpAuthValidity))
}

type RescueNode struct {
	cfg *RescueNodeConfig `yaml:"config,omitempty"`
}

func NewRescueNode(cfg *RescueNodeConfig) *RescueNode {
	return &RescueNode{
		cfg: cfg,
	}
}

func (r *RescueNode) GetName() string {
	return "Rescue Node"
}

func (r *RescueNode) GetDescription() string {
	return `Rocket Rescue Node

The Rocket Rescue Node is a community-run, trust-minimized, and secured fallback node for emergencies and maintenance.

For more information, see https://rescuenode.com`
}

func (r *RescueNode) getCredentialDetails() (*credentialDetails, error) {
	if !r.cfg.Enabled.Value {
		panic("getCredentialDetails() should not be called without checking if RN plugin is enabled")
	}

	password := r.cfg.Password.Value
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
	if !r.cfg.Enabled.Value {
		panic("getCredentialNodeId() should not be called without checking if RN plugin is enabled")
	}

	username := r.cfg.Username.Value
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
	if !r.cfg.Enabled.Value {
		return
	}

	fmt.Printf("%s=== Rescue Node Add-on Enabled ===%s\n", colorYellow, colorReset)
	// Check the Username
	usernameNodeAddr, err := r.getCredentialNodeId()
	if err != nil {
		fmt.Printf("%s%v%s\n", colorRed, err, colorReset)
	} else {
		fmt.Printf("Using a credential issued to %s%s%s.\n", colorBlue, usernameNodeAddr.String(), colorReset)
		if !bytes.Equal(usernameNodeAddr.Bytes(), nodeAddr.Bytes()) {
			fmt.Printf("%s - WARNING: This does not match the Node Account!%s\n", colorYellow, colorReset)
		}
	}

	credentialDetails, err := r.getCredentialDetails()
	if err != nil {
		fmt.Printf("%s%v%s\n", colorRed, err, colorReset)
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

type RescueNodeOverrides struct {
	BnApiEndpoint     string
	CcRpcEndpoint     string
	VcAdditionalFlags string
}

func (r *RescueNode) GetOverrides(bn config.BeaconNode) (*RescueNodeOverrides, error) {
	if !r.cfg.Enabled.Value {
		return nil, nil
	}

	username := r.cfg.Username.Value
	password := r.cfg.Password.Value

	if username == "" || password == "" {
		return nil, fmt.Errorf("Rescue Node can not be enabled without a Username and Password configured.")
	}

	switch bn {
	case config.BeaconNode_Unknown:
		return nil, fmt.Errorf("Unable to generate rescue node URLs for unknown consensus client")
	case config.BeaconNode_Lighthouse,
		config.BeaconNode_Nimbus,
		config.BeaconNode_Lodestar,
		config.BeaconNode_Teku:

		rescueURL := fmt.Sprintf("https://%s:%s@%s.rescuenode.com", username, password, bn)

		return &RescueNodeOverrides{
			BnApiEndpoint: rescueURL,
		}, nil
	case config.BeaconNode_Prysm:
		return &RescueNodeOverrides{
			CcRpcEndpoint:     "prysm-grpc.rescuenode.com:443",
			VcAdditionalFlags: fmt.Sprintf("--grpc-headers=rprnauth=%s:%s --tls-cert=/etc/ssl/certs/ca-certificates.crt", username, password),
		}, nil
	}

	return nil, nil
}
