package pdao

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
	"github.com/urfave/cli/v2"
)

func setSignallingAddress(c *cli.Context, signallingAddress common.Address, signature string) error {
	// // Get RP client
	// rp, err := client.NewClientFromCtx(c)
	// if err != nil {
	// 	return err
	// }

	// // Build the TX
	// response, err := rp.Api.PDao.SetSignallingAddress()

	// Test Strings
	fmt.Printf("Signalling Address: %s\n", signallingAddress)
	fmt.Printf("Signature: %s\n", signature)
	fmt.Println()

	sig, err := utils.ParseEIP712(signature)
	if err != nil {
		return err
	}

	fmt.Println("EIP712Components:")
	fmt.Printf("V: %d\n", sig.V)
	fmt.Printf("R: %s\n", hex.EncodeToString(sig.R[:]))
	fmt.Printf("S: %s\n", hex.EncodeToString(sig.S[:]))

	return nil

}

func clearSignallingAddress(c *cli.Context) error {

	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.ClearSignallingAddress()

	// Test Strings
	fmt.Printf("Test\n")
	fmt.Printf("Response %v\n", response) // response returns nil
	fmt.Printf("Txinfo %v\n", response.Data.TxInfo)

	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to clear your current snapshot address?",
		"clearing snapshot address",
		"Clearing snapshot address...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("The node's signalling address has been sucessfully cleared.")
	return nil
}
