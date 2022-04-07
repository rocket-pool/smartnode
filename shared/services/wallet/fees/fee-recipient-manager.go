package fees

import (
	"github.com/ethereum/go-ethereum/common"
)

// Implementations of this interface will manage the fee recipient file for the corresponding Consensus client.
type FeeRecipientManager interface {
	// Checks if the fee recipient file exists and has the correct distributor address in it.
	// If it does, this returns true - the file is up to date.
	// Otherwise, this writes the file and returns false indicating that the VC should be restarted to pick up the new file.
	CheckAndUpdateFeeRecipientFile(distributor common.Address) (bool, error)
}
