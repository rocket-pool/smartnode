package pdao

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
)

func setSnapshotDelegate(c *cli.Context, snapshotAddress common.Address, signature string) error {

	// fmt.Printf("address: %s, signature: %s", snapshotAddress.String(), signature)

	// TODO:
	// Get the RP client

	// Check for Houston

	// Parse the signature string to extract _r, _s, and _v

	// Assign max fees

	// Prompt for confirmation

	// Network call for set-snapshot-delegate on RocketSignerRegistry
	// rp.SetSnapshotDelegate call here

	// Log & Retrn

	return nil
}
