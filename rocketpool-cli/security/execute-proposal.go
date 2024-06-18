package security

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/dao"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func executeProposal(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get security council proposals
	proposals, err := rp.SecurityProposals()
	if err != nil {
		return err
	}

	// Get executable proposals
	executableProposals := []dao.ProposalDetails{}
	for _, proposal := range proposals.Proposals {
		if proposal.State == types.Succeeded {
			executableProposals = append(executableProposals, proposal)
		}
	}

	// Check for executable proposals
	if len(executableProposals) == 0 {
		fmt.Println("No proposals can be executed.")
		return nil
	}

	// Get selected proposal
	var selectedProposals []dao.ProposalDetails
	if c.String("proposal") == "all" {

		// Select all proposals
		selectedProposals = executableProposals

	} else if c.String("proposal") != "" {

		// Get selected proposal ID
		selectedId, err := strconv.ParseUint(c.String("proposal"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid proposal ID '%s': %w", c.String("proposal"), err)
		}

		// Get matching proposal
		found := false
		for _, proposal := range executableProposals {
			if proposal.ID == selectedId {
				selectedProposals = []dao.ProposalDetails{proposal}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Proposal %d can not be executed.", selectedId)
		}

	} else {

		// Prompt for proposal selection
		options := make([]string, len(executableProposals)+1)
		options[0] = "All available proposals"
		for pi, proposal := range executableProposals {
			options[pi+1] = fmt.Sprintf("proposal %d (message: '%s', payload: %s)", proposal.ID, proposal.Message, proposal.PayloadStr)
		}
		selected, _ := cliutils.Select("Please select a proposal to execute:", options)

		// Get proposals
		if selected == 0 {
			selectedProposals = executableProposals
		} else {
			selectedProposals = []dao.ProposalDetails{executableProposals[selected-1]}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, proposal := range selectedProposals {
		canResponse, err := rp.SecurityCanExecuteProposal(proposal.ID)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for execute transaction (%s)", err)
			break
		} else {
			gasInfo = canResponse.GasInfo
			totalGas += canResponse.GasInfo.EstGasLimit
			totalSafeGas += canResponse.GasInfo.SafeGasLimit
		}
	}
	gasInfo.EstGasLimit = totalGas
	gasInfo.SafeGasLimit = totalSafeGas

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to execute %d proposals?", len(selectedProposals)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Execute proposals
	for _, proposal := range selectedProposals {
		response, err := rp.SecurityExecuteProposal(proposal.ID)
		if err != nil {
			fmt.Printf("Could not execute proposal %d: %s.\n", proposal.ID, err)
			continue
		}

		fmt.Printf("Executing proposal...\n")
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not execute proposal %d: %s.\n", proposal.ID, err)
		} else {
			fmt.Printf("Successfully executed proposal %d.\n", proposal.ID)
		}
	}

	// Return
	return nil

}
