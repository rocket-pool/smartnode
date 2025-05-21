package pdao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func getRewardsPercentages(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get all PDAO settings
	response, err := rp.PDAOGetRewardsPercentages()
	if err != nil {
		return err
	}

	// Print the settings
	fmt.Printf("Node Operators: %.2f%% (%s)\n", eth.WeiToEth(response.Node)*100, response.Node.String())
	fmt.Printf("Oracle DAO:     %.2f%% (%s)\n", eth.WeiToEth(response.OracleDao)*100, response.OracleDao.String())
	fmt.Printf("Protocol DAO:   %.2f%% (%s)\n", eth.WeiToEth(response.ProtocolDao)*100, response.ProtocolDao.String())
	return nil
}

func proposeRewardsPercentages(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for the raw flag
	rawEnabled := c.Bool("raw")

	// Get the node op percent
	nodeString := c.String("node")
	if nodeString == "" {
		if rawEnabled {
			nodeString = prompt.Prompt("Please enter the new rewards allocation for node operators (as an 18-decimal-fixed-point-integer (wei) value):", "^\\d+$", "Invalid amount")
		} else {
			nodeString = prompt.Prompt("Please enter the new rewards allocation for node operators as a percentage from 0 to 1:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		}
	}
	nodePercent, err := parseFloat(c, "node-percent", nodeString, true)
	if err != nil {
		return err
	}

	// Get the oDAO percent
	odaoString := c.String("odao")
	if odaoString == "" {
		if rawEnabled {
			odaoString = prompt.Prompt("Please enter the new rewards allocation for the Oracle DAO (as an 18-decimal-fixed-point-integer (wei) value):", "^\\d+$", "Invalid amount")
		} else {
			odaoString = prompt.Prompt("Please enter the new rewards allocation for the Oracle DAO as a percentage from 0 to 1:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		}
	}
	odaoPercent, err := parseFloat(c, "odao-percent", odaoString, true)
	if err != nil {
		return err
	}

	// Get the pDAO percent
	pdaoString := c.String("pdao")
	if pdaoString == "" {
		if rawEnabled {
			pdaoString = prompt.Prompt("Please enter the new rewards allocation for the Protocol DAO treasury (as an 18-decimal-fixed-point-integer (wei) value):", "^\\d+$", "Invalid amount")
		} else {
			pdaoString = prompt.Prompt("Please enter the new rewards allocation for the Protocol DAO treasury as a percentage from 0 to 1:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		}
	}
	pdaoPercent, err := parseFloat(c, "pdao-percent", pdaoString, true)
	if err != nil {
		return err
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeRewardsPercentages(nodePercent, odaoPercent, pdaoPercent)
	if err != nil {
		return err
	}
	if !canResponse.CanPropose {
		fmt.Println("Cannot propose rewards precentages:")
		if canResponse.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	// Assign max fee
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to propose new rewards allocations?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeRewardsPercentages(nodePercent, odaoPercent, pdaoPercent, canResponse.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Proposing new allocations...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
