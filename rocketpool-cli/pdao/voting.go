package pdao

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func networkSetVotingDelegate(c *cli.Context, nameOrAddress string) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var address common.Address
	var addressString string
	if strings.Contains(nameOrAddress, ".") {
		response, err := rp.ResolveEnsName(nameOrAddress)
		if err != nil {
			return err
		}
		address = response.Address
		addressString = fmt.Sprintf("%s (%s)", nameOrAddress, address.Hex())
	} else {
		address, err = cliutils.ValidateAddress("delegate", nameOrAddress)
		if err != nil {
			return err
		}
		addressString = address.Hex()
	}

	// Get the gas estimation
	gasEstimate, err := rp.EstimateSetVotingDelegateGas(address)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasEstimate.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want %s to represent your node in Rocket Pool on-chain governance proposals?", addressString))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set delegate
	response, err := rp.SetVotingDelegate(address)
	if err != nil {
		return err
	}

	fmt.Printf("Setting delegate...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node's voting delegate was successfuly set to %s.\n", addressString)
	return nil

}
