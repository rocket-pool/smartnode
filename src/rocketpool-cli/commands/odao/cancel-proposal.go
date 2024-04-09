package odao

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

func cancelProposal(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get oracle DAO proposals
	proposals, err := rp.Api.ODao.Proposals()
	if err != nil {
		return err
	}

	// Get wallet status
	wallet, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	status := wallet.Data.WalletStatus

	// Get cancelable proposals
	cancelableProposals := []api.OracleDaoProposalDetails{}
	for _, proposal := range proposals.Data.Proposals {
		if proposal.ProposerAddress == status.Address.NodeAddress && (proposal.State == types.ProposalState_Pending || proposal.State == types.ProposalState_Active) {
			cancelableProposals = append(cancelableProposals, proposal)
		}
	}

	// Check for cancelable proposals
	if len(cancelableProposals) == 0 {
		fmt.Println("No proposals can be cancelled.")
		return nil
	}

	// Get selected proposal
	var selectedProposal api.OracleDaoProposalDetails
	if c.Uint64(proposalFlag.Name) != 0 {
		// Get matching proposal
		selectedId := c.Uint64(proposalFlag.Name)
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
		selected, _ := utils.Select("Please select a proposal to cancel:", options)
		selectedProposal = cancelableProposals[selected]
	}

	// Build the TX
	response, err := rp.Api.ODao.CancelProposal(selectedProposal.ID)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanCancel {
		fmt.Println("Cannot cancel proposal:")
		if response.Data.DoesNotExist {
			fmt.Println("The proposal does not exist.")
		}
		if response.Data.InvalidProposer {
			fmt.Println("You are not the proposer of this proposal.")
		}
		if response.Data.InvalidState {
			fmt.Println("The proposal is not in a cancellable state.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to cancel proposal %d?", selectedProposal.ID),
		"proposal cancellation",
		"Canceling proposal...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully cancelled proposal %d.\n", selectedProposal.ID)
	return nil
}
