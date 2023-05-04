package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func setStakeRPLForAllowed(c *cli.Context, callerAddressOrENS string, allowed bool) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
	}

	var callerAddress common.Address
	var callerAddressString string
	if strings.Contains(callerAddressOrENS, ".") {
		response, err := rp.ResolveEnsName(callerAddressOrENS)
		if err != nil {
			return err
		}
		callerAddress = response.Address
		callerAddressString = fmt.Sprintf("%s (%s)", callerAddressOrENS, callerAddress.Hex())
	} else {
		callerAddress, err = cliutils.ValidateAddress("caller", callerAddressOrENS)
		if err != nil {
			return err
		}
		callerAddressString = callerAddress.Hex()
	}

	// Get the gas estimate
	canResponse, err := rp.CanSetStakeRPLForAllowed(callerAddress, allowed)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to set allow status of %s to %t?", callerAddressString, allowed))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the allow status
	response, err := rp.SetStakeRPLForAllowed(callerAddress, allowed)
	if err != nil {
		return err
	}

	fmt.Printf("Setting RPL stake for allowed...\n")
	cliutils.PrintTransactionHash(rp, response.SetTxHash)
	if _, err = rp.WaitForTransaction(response.SetTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully set stake RPL for allowed state of %s to %t.\n", callerAddressString, allowed)
	return nil
}

