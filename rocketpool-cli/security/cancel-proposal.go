package security

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/types"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func cancelProposal(proposal string, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get security council proposals
	proposals, err := rp.SecurityProposals()
	if err != nil {
		return err
	}

	// Get wallet status
	wallet, err := rp.WalletStatus()
	if err != nil {
		return err
	}

	// Get cancelable proposals
	cancelableProposals := []dao.ProposalDetails{}
	for _, proposal := range proposals.Proposals {
		if bytes.Equal(proposal.ProposerAddress.Bytes(), wallet.AccountAddress.Bytes()) && (proposal.State == types.Pending || proposal.State == types.Active) {
			cancelableProposals = append(cancelableProposals, proposal)
		}
	}

	// Check for cancelable proposals
	if len(cancelableProposals) == 0 {
		fmt.Println("No proposals can be cancelled.")
		return nil
	}

	// Get selected proposal
	var selectedProposal dao.ProposalDetails
	if proposal != "" {

		// Get selected proposal ID
		selectedId, err := strconv.ParseUint(proposal, 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid proposal ID '%s': %w", proposal, err)
		}

		// Get matching proposal
		found := false
		for _, proposal := range cancelableProposals {
			if proposal.ID == selectedId {
				selectedProposal = proposal
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Proposal %d can not be cancelled.", selectedId)
		}

	} else {

		// Prompt for proposal selection
		options := make([]string, len(cancelableProposals))
		for pi, proposal := range cancelableProposals {
			options[pi] = fmt.Sprintf("proposal %d (message: '%s', payload: %s)", proposal.ID, proposal.Message, proposal.PayloadStr)
		}
		selected, _ := prompt.Select("Please select a proposal to cancel:", options)
		selectedProposal = cancelableProposals[selected]

	}

	// Display gas estimate
	canResponse, err := rp.SecurityCanCancelProposal(selectedProposal.ID)
	if err != nil {
		return err
	}
	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to cancel proposal %d?", selectedProposal.ID) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Cancel proposal
	response, err := rp.SecurityCancelProposal(selectedProposal.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Canceling proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully cancelled proposal %d.\n", selectedProposal.ID)
	return nil

}
