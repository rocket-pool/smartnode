package odao

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	rocketpoolapi "github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func getUpgradeProposals(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get upgrade upgradeProposals
	upgradeProposals, err := rp.TNDAOUpgradeProposals()
	if err != nil {
		return err
	}

	fmt.Printf("Found %d upgrade proposals.\n\n", len(upgradeProposals.Proposals))

	typeMap := make(map[string]string)
	typeMap["0x529a09aed0ded46c4cc64a5f9a6cb6dbde240a9e9c966041749f311248110e11"] = "UpgradeContract"
	typeMap["0x19f10da52b60efe9f5ee07f5d429a865c0bda23ae1482284927780c41b724cef"] = "AddContract"
	typeMap["0x1ca99adef8e0f1a6fa2cbc2b46aedc54c66479f38df59bff3575de40893db660"] = "AddABI"
	typeMap["0xbe19c295254203061a6ecbbdd7353a2134a6bae25c11f27532123ce4a4be1600"] = "UpgradeABI"
	// Print upgrade proposals
	for _, proposal := range upgradeProposals.Proposals {
		fmt.Printf("Upgrade proposal %d: %s\n", proposal.ID, proposal.Name)
		fmt.Printf("  State: %s\n", types.UpgradeProposalStates[types.UpgradeProposalState(proposal.State)])
		fmt.Printf("  End time: %s\n", time.Unix(proposal.EndTime.Int64(), 0).Format(time.RFC3339))
		fmt.Printf("  Type: %s\n", typeMap[common.BytesToHash([]byte(proposal.Type[:])).String()])
		fmt.Printf("  Upgrade address: %s\n", proposal.UpgradeAddress)
		fmt.Printf("  Upgrade ABI: %s\n", proposal.UpgradeAbi)

	}

	// Return
	return nil
}

func executeUpgrade(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get upgrade upgradeProposals
	upgradeProposals, err := rp.TNDAOUpgradeProposals()
	if err != nil {
		return err
	}

	// Get executable proposals
	executableProposals := []upgrades.UpgradeProposalDetails{}
	pendingProposals := []upgrades.UpgradeProposalDetails{}
	for _, proposal := range upgradeProposals.Proposals {
		if proposal.State == types.UpgradeProposalState_Succeeded {
			executableProposals = append(executableProposals, proposal)
		} else if proposal.State == types.UpgradeProposalState_Pending {
			pendingProposals = append(pendingProposals, proposal)
		}
	}

	// Check for executable proposals
	if len(executableProposals) == 0 {
		fmt.Println("No upgrade proposals can be executed.\n	")

		if len(pendingProposals) > 0 {
			// Print pending proposals
			fmt.Println("Pending upgrade proposals:")
			for _, proposal := range pendingProposals {
				fmt.Printf("    ID %d: %s\n", proposal.ID, proposal.Name)
				fmt.Printf("    State: %s\n", types.UpgradeProposalStates[types.UpgradeProposalState(proposal.State)])
				fmt.Printf("    End time: %s\n", time.Unix(proposal.EndTime.Int64(), 0).Format(time.RFC3339))
				fmt.Printf("    Upgrade address: %s\n", proposal.UpgradeAddress)
				fmt.Printf("    Upgrade ABI: %s\n\n", proposal.UpgradeAbi)
			}
			return nil
		}
		return nil
	}

	// Get selected proposal
	var selectedProposals []upgrades.UpgradeProposalDetails
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
				selectedProposals = []upgrades.UpgradeProposalDetails{proposal}
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
			options[pi+1] = fmt.Sprintf("proposal %d (name: '%s', type: '%s', upgrade address: '%s', upgrade ABI: '%s')", proposal.ID, proposal.Name, proposal.Type, proposal.UpgradeAddress, proposal.UpgradeAbi)
		}
		selected, _ := prompt.Select("Please select a proposal to execute:", options)

		// Get proposals
		if selected == 0 {
			selectedProposals = executableProposals
		} else {
			selectedProposals = []upgrades.UpgradeProposalDetails{executableProposals[selected-1]}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, proposal := range selectedProposals {
		canResponse, err := rp.CanExecuteUpgradeProposal(proposal.ID)
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

	// Get max fees
	g, err := gas.GetMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to execute %d proposals?", len(selectedProposals)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Execute proposals
	for _, proposal := range selectedProposals {
		g.Assign(rp)
		response, err := rp.ExecuteTNDAOProposal(proposal.ID)
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
