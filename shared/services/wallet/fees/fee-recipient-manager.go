package fees

import (
	"github.com/ethereum/go-ethereum/common"
)

// Implementations of this interface will manage the fee recipient file for the corresponding Consensus client.
type FeeRecipientManager interface {
	// Checks if the fee recipient file exists and has the correct distributor address in it.
	// The first return value is for file existence, the second is for validation of the fee recipient address inside.
	CheckFeeRecipientFile(distributor common.Address) (bool, bool, error)

	// Writes the given address to the fee recipient file. The VC should be restarted to pick up the new file.
	UpdateFeeRecipientFile(distributor common.Address) error
}
