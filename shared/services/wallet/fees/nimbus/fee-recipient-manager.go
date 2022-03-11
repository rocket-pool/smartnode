package nimbus

import (
	"fmt"
	"io/fs"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
)

// Config
const (
	FeeRecipientFilename string      = ""
	FileMode             fs.FileMode = 0600
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

	return fmt.Errorf("Nimbus currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

}
