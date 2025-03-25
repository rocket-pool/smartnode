package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func addAddressToStakeRplWhitelist(c *cli.Context, addressOrENS string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var address common.Address
	var addressString string
	if strings.Contains(addressOrENS, ".") {
		response, err := rp.ResolveEnsName(addressOrENS)
		if err != nil {
			return err
		}
		address = response.Address
		addressString = fmt.Sprintf("%s (%s)", addressOrENS, address.Hex())
	} else {
		address, err = cliutils.ValidateAddress("address", addressOrENS)
		if err != nil {
			return err
		}
		addressString = address.Hex()
	}

	// Get the gas estimate
	canResponse, err := rp.CanSetStakeRPLForAllowed(address, true)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to allow %s to stake RPL for your node?", addressString))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the allow status
	response, err := rp.SetStakeRPLForAllowed(address, true)
	if err != nil {
		return err
	}

	fmt.Printf("Adding address to RPL stake whitelist...\n")
	cliutils.PrintTransactionHash(rp, response.SetTxHash)
	if _, err = rp.WaitForTransaction(response.SetTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully added %s to your node's RPL staking whitelist.\n", addressString)
	return nil
}

func removeAddressFromStakeRplWhitelist(c *cli.Context, addressOrENS string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var address common.Address
	var addressString string
	if strings.Contains(addressOrENS, ".") {
		response, err := rp.ResolveEnsName(addressOrENS)
		if err != nil {
			return err
		}
		address = response.Address
		addressString = fmt.Sprintf("%s (%s)", addressOrENS, address.Hex())
	} else {
		address, err = cliutils.ValidateAddress("address", addressOrENS)
		if err != nil {
			return err
		}
		addressString = address.Hex()
	}

	// Get the gas estimate
	canResponse, err := rp.CanSetStakeRPLForAllowed(address, false)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to remove %s from your RPL staking whitelist?", addressString))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the allow status
	response, err := rp.SetStakeRPLForAllowed(address, false)
	if err != nil {
		return err
	}

	fmt.Printf("Removing address from RPL stake whitelist...\n")
	cliutils.PrintTransactionHash(rp, response.SetTxHash)
	if _, err = rp.WaitForTransaction(response.SetTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully removed %s from your node's RPL staking whitelist.\n", addressString)
	return nil
}
