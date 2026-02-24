package pdao

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func setAllowListedControllers(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var addressListStr string
	addressListStr = c.String("addressList")
	if addressListStr == "" {
		// Ask the user how many addresses should be included in the list
		numStr := prompt.Prompt(fmt.Sprintf("How many addresses do you want to propose as allowlisted controllers? Enter 0 to propose clearing the list"), "^\\d+$", "Invalid number.")
		numAddressesUint, err := strconv.ParseUint(numStr, 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid number: %w.\n", numStr, err)
		}
		numAddr := int(numAddressesUint)

		// Construct a string of addresses
		var address []string
		for i := 0; i < numAddr; i++ {
			promptMsg := fmt.Sprintf(
				"Please enter address %d of %d:",
				i+1, numAddr,
			)
			promptedAddr := prompt.Prompt(promptMsg, "^0x[0-9a-fA-F]{40}$", "Invalid address")

			// Validate input
			_, err := cliutils.ValidateAddress("address", promptedAddr)
			if err != nil {
				return err
			}
			address = append(address, promptedAddr)
		}
		addressListStr = strings.Join(address, ",")
	} else {
		_, err = cliutils.ValidateAddresses("addressList", addressListStr)
		if err != nil {
			return err
		}
	}

	// Prompt for confirmation
	if addressListStr == "" {
		fmt.Printf("%sYou are proposing to remove all allowlisted controllers%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("%sYou have selected propose %v as the allowlisted controllers%s\n", colorGreen, addressListStr, colorReset)
	}
	fmt.Println()

	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to propose a new list of allowlisted controllers?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	canResponse, err := rp.PDAOCanProposeAllowListedControllers(addressListStr)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Submit
	response, err := rp.PDAOProposeAllowListedControllers(addressListStr, canResponse.BlockNumber)
	if err != nil {
		return err
	}
	hash := response.TxHash

	fmt.Printf("Proposing allow listed controllers...\n")
	cliutils.PrintTransactionHash(rp, hash)
	if _, err = rp.WaitForTransaction(hash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
