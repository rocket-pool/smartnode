package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
)

// Flags
var proposeRewardsPercentagesNodeFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "node",
	Aliases: []string{"n"},
	Usage:   fmt.Sprintf("The node operator's rewards allocation (a percentage from 0 to 1 if '--%s' is not set)", utils.RawFlag.Name),
}
var proposeRewardsPercentagesOdaoFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "odao",
	Aliases: []string{"o"},
	Usage:   fmt.Sprintf("The Oracle DAO's rewards allocation (a percentage from 0 to 1 if '--%s' is not set)", utils.RawFlag.Name),
}
var proposeRewardsPercentagesPdaoFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "pdao",
	Aliases: []string{"p"},
	Usage:   fmt.Sprintf("The Protocol DAO's rewards allocation (a percentage from 0 to 1 if '--%s' is not set)", utils.RawFlag.Name),
}

func proposeRewardsPercentages(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check for the raw flag
	rawEnabled := c.Bool(utils.RawFlag.Name)

	// Get the node op percent
	nodeString := c.String(proposeRewardsPercentagesNodeFlag.Name)
	if nodeString == "" {
		if rawEnabled {
			nodeString = utils.Prompt("Please enter the new rewards allocation for node operators (as an 18-decimal-fixed-point-integer (wei) value):", "^\\d+$", "Invalid amount")
		} else {
			nodeString = utils.Prompt("Please enter the new rewards allocation for node operators as a percentage from 0 to 1:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		}
	}
	nodePercent, err := utils.ParseFloat(c, "node-percent", nodeString, true)
	if err != nil {
		return err
	}

	// Get the oDAO percent
	odaoString := c.String(proposeRewardsPercentagesOdaoFlag.Name)
	if odaoString == "" {
		if rawEnabled {
			odaoString = utils.Prompt("Please enter the new rewards allocation for the Oracle DAO (as an 18-decimal-fixed-point-integer (wei) value):", "^\\d+$", "Invalid amount")
		} else {
			odaoString = utils.Prompt("Please enter the new rewards allocation for the Oracle DAO as a percentage from 0 to 1:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		}
	}
	odaoPercent, err := utils.ParseFloat(c, "odao-percent", odaoString, true)
	if err != nil {
		return err
	}

	// Get the pDAO percent
	pdaoString := c.String(proposeRewardsPercentagesPdaoFlag.Name)
	if pdaoString == "" {
		if rawEnabled {
			pdaoString = utils.Prompt("Please enter the new rewards allocation for the Protocol DAO treasury (as an 18-decimal-fixed-point-integer (wei) value):", "^\\d+$", "Invalid amount")
		} else {
			pdaoString = utils.Prompt("Please enter the new rewards allocation for the Protocol DAO treasury as a percentage from 0 to 1:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		}
	}
	pdaoPercent, err := utils.ParseFloat(c, "pdao-percent", pdaoString, true)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.ProposeRewardsPercentages(nodePercent, odaoPercent, pdaoPercent)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("You cannot currently submit this proposal:")
		if response.Data.InsufficientRpl {
			fmt.Printf("You do not have enough unlocked RPL (proposals require locking %.6f RPL, but you only have %.6f RPL staked and unlocked).", eth.WeiToEth(response.Data.ProposalBond), eth.WeiToEth(big.NewInt(0).Sub(response.Data.StakedRpl, response.Data.LockedRpl)))
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to propose new rewards allocations?",
		"new rewards allocation proposal",
		"Proposing new allocations...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
