package teku

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type FeeRecipientFileContents struct {
	DefaultConfig  ProposerFeeRecipient            `json:"default_config"`
	ProposerConfig map[string]ProposerFeeRecipient `json:"proposer_config"`
}

type ProposerFeeRecipient struct {
	FeeRecipient string `json:"fee_recipient"`
}

// Creates a fee recipient file that points all of this node's validators to the node distributor address.
func (c *Client) GenerateFeeRecipientFile(rp *rocketpool.RocketPool, nodeAddress common.Address) ([]byte, error) {

	// Get the distributor address for this node
	distributor, err := node.GetDistributorAddress(rp, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting distributor address for node [%s]: %w", nodeAddress.Hex(), err)
	}

	// Write out the default
	distributorAddress := distributor.Hex()
	fileContents := FeeRecipientFileContents{
		DefaultConfig: ProposerFeeRecipient{
			FeeRecipient: distributorAddress,
		},
		ProposerConfig: map[string]ProposerFeeRecipient{},
	}

	// Get all of the minipool addresses for the node
	minipoolAddresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses for node [%s]: %w", nodeAddress.Hex(), err)
	}

	// Write all of the validator addresses
	for _, minipoolAddress := range minipoolAddresses {
		pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting validator pubkey for minipool [%s]: %w", minipoolAddress.Hex(), err)
		}
		fileContents.ProposerConfig[pubkey.Hex()] = ProposerFeeRecipient{
			FeeRecipient: distributorAddress,
		}
	}

	// Serialize the file contents
	bytes, err := json.Marshal(fileContents)
	if err != nil {
		return nil, fmt.Errorf("error serializing file contents to JSON: %w", err)
	}

	return bytes, nil

}
