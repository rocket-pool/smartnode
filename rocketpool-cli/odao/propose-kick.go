package odao

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func proposeKick(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get DAO members
	members, err := rp.TNDAOMembers()
	if err != nil {
		return err
	}

	// Get member to propose kicking
	var selectedMember trustednode.MemberDetails
	if c.String("member") != "" {

		// Get matching member
		selectedAddress := common.HexToAddress(c.String("member"))
		for _, member := range members.Members {
			if bytes.Equal(member.Address.Bytes(), selectedAddress.Bytes()) {
				selectedMember = member
				break
			}
		}
		if !selectedMember.Exists {
			return fmt.Errorf("The oracle DAO member %s does not exist.", selectedAddress.Hex())
		}

	} else {

		// Prompt for member selection
		options := make([]string, len(members.Members))
		for mi, member := range members.Members {
			options[mi] = fmt.Sprintf("%s (URL: %s, node: %s)", member.ID, member.Url, member.Address)
		}
		selected, _ := prompt.Select("Please select a member to propose kicking:", options)
		selectedMember = members.Members[selected]

	}

	// Get fine amount
	var fineAmountWei *big.Int
	if c.String("fine") == "max" {

		// Set fine amount to member's entire RPL bond
		fineAmountWei = selectedMember.RPLBondAmount

	} else if c.String("fine") != "" {

		// Parse amount
		fineAmount, err := strconv.ParseFloat(c.String("fine"), 64)
		if err != nil {
			return fmt.Errorf("Invalid fine amount '%s': %w", c.String("fine"), err)
		}
		fineAmountWei = eth.EthToWei(fineAmount)

	} else {

		// Prompt for custom amount
		inputAmount := prompt.Prompt(fmt.Sprintf("Please enter an RPL fine amount to propose (max %.6f RPL):", math.RoundDown(eth.WeiToEth(selectedMember.RPLBondAmount), 6)), "^\\d+(\\.\\d+)?$", "Invalid amount")
		fineAmount, err := strconv.ParseFloat(inputAmount, 64)
		if err != nil {
			return fmt.Errorf("Invalid fine amount '%s': %w", inputAmount, err)
		}
		fineAmountWei = eth.EthToWei(fineAmount)

	}

	// Check if proposal can be made
	canPropose, err := rp.CanProposeKickFromTNDAO(selectedMember.Address, fineAmountWei)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose kicking member:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		if canPropose.InsufficientRplBond {
			fmt.Printf("The fine amount of %.6f RPL is greater than the member's bond of %.6f RPL.\n", math.RoundDown(eth.WeiToEth(fineAmountWei), 6), math.RoundDown(eth.WeiToEth(selectedMember.RPLBondAmount), 6))
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeKickFromTNDAO(selectedMember.Address, fineAmountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Kicking %s from the oracle DAO...\n", selectedMember.Address.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a kick proposal with ID %d for node %s, with a fine of %.6f RPL.\n", response.ProposalId, selectedMember.Address.Hex(), math.RoundDown(eth.WeiToEth(fineAmountWei), 6))
	return nil

}
