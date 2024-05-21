package pdao

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
)

func setSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) error {

	// fmt.Printf("address: %s, signature: %s", snapshotAddress.String(), signature)

	// TODO:
	// Get the RP client

	// Check for Houston

	// Parse the signature string to extract _r, _s, and _v

	// Assign max fees

	// Prompt for confirmation

	// Network call for set-snapshot-address on RocketSignerRegistry
	// rp.SetSnapshotAddress call here

	// Log & Retrn

	return nil
}
