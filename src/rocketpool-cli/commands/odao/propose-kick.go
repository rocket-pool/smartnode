package odao

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

var kickFineFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "fine",
	Aliases: []string{"f"},
	Usage:   "The amount of RPL to fine the member (or 'max')",
}

func proposeKick(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get DAO members
	members, err := rp.Api.ODao.Members()
	if err != nil {
		return err
	}

	// Get member to propose kicking
	var selectedMember api.OracleDaoMemberDetails
	if c.String(memberFlag.Name) != "" {
		// Get matching member
		selectedAddress := common.HexToAddress(c.String(memberFlag.Name))
		for _, member := range members.Data.Members {
			if bytes.Equal(member.Address.Bytes(), selectedAddress.Bytes()) {
				selectedMember = member
				break
			}
		}
		if !selectedMember.Exists {
			return fmt.Errorf("the oracle DAO member %s does not exist", selectedAddress.Hex())
		}
	} else {
		// Prompt for member selection
		options := make([]string, len(members.Data.Members))
		for mi, member := range members.Data.Members {
			options[mi] = fmt.Sprintf("%s (URL: %s, node: %s)", member.ID, member.Url, member.Address)
		}
		selected, _ := utils.Select("Please select a member to propose kicking:", options)
		selectedMember = members.Data.Members[selected]
	}

	// Get fine amount
	var fineAmountWei *big.Int
	if c.String(kickFineFlag.Name) == "max" {
		// Set fine amount to member's entire RPL bond
		fineAmountWei = selectedMember.RplBondAmount
	} else if c.String(kickFineFlag.Name) != "" {
		// Parse amount
		fineAmount, err := strconv.ParseFloat(c.String(kickFineFlag.Name), 64)
		if err != nil {
			return fmt.Errorf("invalid fine amount '%s': %w", c.String(kickFineFlag.Name), err)
		}
		fineAmountWei = eth.EthToWei(fineAmount)
	} else {
		// Prompt for custom amount
		inputAmount := utils.Prompt(fmt.Sprintf("Please enter an RPL fine amount to propose (max %.6f RPL):", math.RoundDown(eth.WeiToEth(selectedMember.RplBondAmount), 6)), "^\\d+(\\.\\d+)?$", "Invalid amount")
		fineAmount, err := strconv.ParseFloat(inputAmount, 64)
		if err != nil {
			return fmt.Errorf("invalid fine amount '%s': %w", inputAmount, err)
		}
		fineAmountWei = eth.EthToWei(fineAmount)
	}

	// Build the TX
	response, err := rp.Api.ODao.ProposeKick(selectedMember.Address, fineAmountWei)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("Cannot propose kicking member:")
		if response.Data.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		if response.Data.InsufficientRplBond {
			fmt.Printf("The fine amount of %.6f RPL is greater than the member's bond of %.6f RPL.\n", math.RoundDown(eth.WeiToEth(fineAmountWei), 6), math.RoundDown(eth.WeiToEth(selectedMember.RplBondAmount), 6))
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to submit this proposal?",
		"proposing kicking Oracle DAO member",
		"Proposing kick of %s from the oracle DAO...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully submitted a kick proposal for node %s, with a fine of %.6f RPL.\n", selectedMember.Address.Hex(), math.RoundDown(eth.WeiToEth(fineAmountWei), 6))
	return nil
}
