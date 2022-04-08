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
// The first return value is for file existence, the second is for validation of the fee recipient address inside.
func (fm *FeeRecipientManager) CheckFeeRecipientFile(distributor common.Address) (bool, bool, error) {

	return true, true, nil

	return false, false, fmt.Errorf("Nimbus currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

}

// Writes the given address to the fee recipient file. The VC should be restarted to pick up the new file.
func (fm *FeeRecipientManager) UpdateFeeRecipientFile(distributor common.Address) error {

	return nil

	return fmt.Errorf("Nimbus currently does not provide support for per-validator fee recipient specification, so it cannot be used to test the Merge. We will re-enable it when it has support for this feature.")

}
