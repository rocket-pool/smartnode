package security

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	"github.com/urfave/cli/v2"
)

var executeProposalFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "proposal",
	Aliases: []string{"p"},
	Usage:   "The ID of the proposal to execute (or 'all')",
}

func executeProposal(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get security council proposals
	proposals, err := rp.Api.Security.Proposals()
	if err != nil {
		return err
	}

	// Get executable proposals
	executableProposals := []api.SecurityProposalDetails{}
	for _, proposal := range proposals.Data.Proposals {
		if proposal.State == types.ProposalState_Succeeded {
			executableProposals = append(executableProposals, proposal)
		}
	}

	// Check for executable proposals
	if len(executableProposals) == 0 {
		fmt.Println("No proposals can be executed.")
		return nil
	}

	// Get selected proposals
	options := make([]utils.SelectionOption[api.SecurityProposalDetails], len(executableProposals))
	for i, prop := range executableProposals {
		option := &options[i]
		option.Element = &executableProposals[i]
		option.ID = fmt.Sprint(prop.ID)
		option.Display = fmt.Sprintf("proposal %d (message: '%s', payload: %s)", prop.ID, prop.Message, prop.PayloadStr)
	}
	selectedProposals, err := utils.GetMultiselectIndices(c, executeProposalFlag.Name, options, "Please select a proposal to execute:")
	if err != nil {
		return fmt.Errorf("error determining proposal selection: %w", err)
	}

	// Build the TXs
	ids := make([]uint64, len(selectedProposals))
	for i, prop := range selectedProposals {
		ids[i] = prop.ID
	}
	response, err := rp.Api.Security.ExecuteProposals(ids)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedProposals))
	for i, prop := range selectedProposals {
		data := response.Data.Batch[i]
		if !data.CanExecute {
			fmt.Printf("Cannot execute proposal %d:\n", prop.ID)
			if data.DoesNotExist {
				fmt.Println("The proposal does not exist.")
			}
			if data.InvalidState {
				fmt.Println("The proposal is not in an executable state.")
			}
			return nil
		}
		txs[i] = data.TxInfo
	}

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to execute %d proposals?", len(selectedProposals)),
		func(i int) string {
			return fmt.Sprintf("executing proposal %d", selectedProposals[i].ID)
		},
		"Executing proposals...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully executed all selected proposals.")
	return nil
}
