package nimbus

import (
	"fmt"
	"io/fs"

	"github.com/ethereum/go-ethereum/common"
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

// Checks if the fee recipient file exists and has the correct distributor address in it.
// If it does, this returns true - the file is up to date.
// Otherwise, this writes the file and returns false indicating that the VC should be restarted to pick up the new file.
func (fm *FeeRecipientManager) CheckAndUpdateFeeRecipientFile(distributor common.Address) (bool, error) {

	return false, fmt.Errorf("Nimbus currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

}
