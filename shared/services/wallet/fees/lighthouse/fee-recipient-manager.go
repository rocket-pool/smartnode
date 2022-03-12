package lighthouse

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

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

// Creates a new fee recipient manager
func NewFeeRecipientManager(keystore keystore.Keystore) *FeeRecipientManager {
	return &FeeRecipientManager{
		keystore: keystore,
	}
}

// Creates a fee recipient file that points all of this node's validators to the node distributor address.
func (fm *FeeRecipientManager) StoreFeeRecipientFile(rp *rocketpool.RocketPool, nodeAddress common.Address) error {

	// Get the distributor address for this node
	distributor, err := node.GetDistributorAddress(rp, nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("error getting distributor address for node [%s]: %w", nodeAddress.Hex(), err)
	}

	// Write out the default
	distributorAddress := distributor.Hex()
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("default: %s\n", distributorAddress))

	// Get all of the minipool addresses for the node
	minipoolAddresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool addresses for node [%s]: %w", nodeAddress.Hex(), err)
	}

	// Write all of the validator addresses
	for _, minipoolAddress := range minipoolAddresses {
		pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting validator pubkey for minipool [%s]: %w", minipoolAddress.Hex(), err)
		}

		builder.WriteString(fmt.Sprintf("%s: %s\n", pubkey.Hex(), distributorAddress))
	}

	// Write the string out to the file
	bytes := []byte(builder.String())
	path := filepath.Join(fm.keystore.GetKeystoreDir(), config.LighthouseFeeRecipientFilename)
	err = ioutil.WriteFile(path, bytes, FileMode)
	if err != nil {
		return fmt.Errorf("error writing fee recipient file: %w", err)
	}

	return nil

}
