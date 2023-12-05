package pdao

import (
	"fmt"
	"sort"
	"strconv"

	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func claimBonds(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Get protocol DAO proposals
	claimableBondsResponse, err := rp.PDAOGetClaimableBonds()
	if err != nil {
		return fmt.Errorf("error checking for claimable bonds: %w", err)
	}
	claimableBonds := claimableBondsResponse.ClaimableBonds

	// Check for executable proposals
	if len(claimableBonds) == 0 {
		fmt.Println("You do not have any unlockable bonds or claimable rewards.")
		return nil
	}

	// Get selected proposal
	var selectedClaims []api.BondClaimResult
	if c.String("proposal") == "all" {

		// Select all proposals
		selectedClaims = claimableBonds

	} else if c.String("proposal") != "" {

		// Get selected proposal ID
		selectedId, err := strconv.ParseUint(c.String("proposal"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid proposal ID '%s': %w", c.String("proposal"), err)
		}

		// Get matching proposal
		found := false
		for _, bond := range claimableBonds {
			if bond.ProposalID == selectedId {
				selectedClaims = []api.BondClaimResult{bond}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Proposal %d does not have any unlockable bonds or claimable rewards.", selectedId)
		}

	} else {

		// Prompt for proposal selection
		options := make([]string, len(claimableBonds)+1)
		options[0] = "All available proposals"
		for pi, bond := range claimableBonds {
			options[pi+1] = fmt.Sprintf("Proposal %d (proposer: %t, unlockable: %.2f RPL, rewards: %.2f RPL)", bond.ProposalID, bond.IsProposer, eth.WeiToEth(bond.UnlockAmount), eth.WeiToEth(bond.RewardAmount))
		}
		selected, _ := cliutils.Select("Please select a proposal to unlock bonds / claim rewards from:", options)

		// Get proposals
		if selected == 0 {
			selectedClaims = claimableBonds
		} else {
			selectedClaims = []api.BondClaimResult{claimableBonds[selected-1]}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, bond := range selectedClaims {
		indices := getClaimIndicesForBond(bond)
		canResponse, err := rp.PDAOCanClaimBonds(bond.ProposalID, indices)
		if err != nil {
			return fmt.Errorf("error simulating claim-bond on proposal %d: %s", bond.ProposalID, err.Error())
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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to claim bonds and rewards from %d proposals?", len(selectedClaims)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Claim bonds from proposals
	for _, bond := range selectedClaims {
		indices := getClaimIndicesForBond(bond)
		response, err := rp.PDAOClaimBonds(bond.IsProposer, bond.ProposalID, indices)
		if err != nil {
			fmt.Printf("Could not claim bonds from proposal %d: %s.\n", bond.ProposalID, err)
			continue
		}

		fmt.Printf("Claiming bonds from proposal %d...\n", bond.ProposalID)
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not claim bonds from proposal %d: %s.\n", bond.ProposalID, err)
		} else {
			fmt.Printf("Successfully claimed bonds from proposal %d.\n", bond.ProposalID)
		}
	}

	// Return
	return nil

}

func getClaimIndicesForBond(bond api.BondClaimResult) []uint64 {
	indexMap := map[uint64]bool{}
	for _, index := range bond.UnlockableIndices {
		indexMap[index] = true
	}
	for _, index := range bond.RewardableIndices {
		indexMap[index] = true
	}

	indices := make([]uint64, 0, len(indexMap))
	for index, _ := range indexMap {
		indices = append(indices, index)
	}

	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i] < indices[j]
	})

	return indices
}
