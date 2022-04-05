package prysm

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
)

// Config
const (
	FileMode fs.FileMode = 0600
)

type FeeRecipientManager struct {
	keystore keystore.Keystore
}

type FeeRecipientFileContents struct {
	DefaultConfig  ProposerFeeRecipient            `json:"default_config"`
	ProposerConfig map[string]ProposerFeeRecipient `json:"proposer_config"`
}

type ProposerFeeRecipient struct {
	FeeRecipient string `json:"fee_recipient"`
}

// Creates a new fee recipient manager
func NewFeeRecipientManager(keystore keystore.Keystore) *FeeRecipientManager {
	return &FeeRecipientManager{
		keystore: keystore,
	}
}

// Creates a fee recipient file that points all of this node's validators to the node distributor address.
func (fm *FeeRecipientManager) StoreFeeRecipientFile(rp *rocketpool.RocketPool, nodeAddress common.Address) (common.Address, error) {

	// Get the distributor address for this node
	distributor, err := node.GetDistributorAddress(rp, nodeAddress, nil)
	if err != nil {
		return common.Address{}, fmt.Errorf("error getting distributor address for node [%s]: %w", nodeAddress.Hex(), err)
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
		return common.Address{}, fmt.Errorf("error getting minipool addresses for node [%s]: %w", nodeAddress.Hex(), err)
	}

	// Write all of the validator addresses
	for _, minipoolAddress := range minipoolAddresses {
		pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
		if err != nil {
			return common.Address{}, fmt.Errorf("error getting validator pubkey for minipool [%s]: %w", minipoolAddress.Hex(), err)
		}
		fileContents.ProposerConfig[pubkey.Hex()] = ProposerFeeRecipient{
			FeeRecipient: distributorAddress,
		}
	}

	// Serialize the file contents
	bytes, err := json.Marshal(fileContents)
	if err != nil {
		return common.Address{}, fmt.Errorf("error serializing file contents to JSON: %w", err)
	}

	// Write the contents out to the file
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.PrysmFeeRecipientFilename)
	err = ioutil.WriteFile(path, bytes, FileMode)
	if err != nil {
		return common.Address{}, fmt.Errorf("error writing fee recipient file: %w", err)
	}

	// TODO: WAIT FOR PRYSM TO ADD SUPPORT FOR THIS, SEE
	// https://github.com/prysmaticlabs/prysm/pull/10312
	return common.Address{}, fmt.Errorf("Prysm currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

	return distributor, nil

}
