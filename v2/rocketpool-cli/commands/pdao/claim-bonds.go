package pdao

import (
	"fmt"
	"sort"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

func claimBonds(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get protocol DAO proposals
	claimableBondsResponse, err := rp.Api.PDao.GetClaimableBonds()
	if err != nil {
		return fmt.Errorf("error checking for claimable bonds: %w", err)
	}
	claimableBonds := claimableBondsResponse.Data.ClaimableBonds

	// Check for executable proposals
	if len(claimableBonds) == 0 {
		fmt.Println("You do not have any unlockable bonds or claimable rewards.")
		return nil
	}

	// Get selected proposals
	options := make([]utils.SelectionOption[api.BondClaimResult], len(claimableBonds))
	for i, bond := range claimableBonds {
		option := &options[i]
		option.Element = &claimableBonds[i]
		option.ID = fmt.Sprint(bond.ProposalID)
		option.Display = fmt.Sprintf("Proposal %d (proposer: %t, unlockable: %.2f RPL, rewards: %.2f RPL)", bond.ProposalID, bond.IsProposer, eth.WeiToEth(bond.UnlockAmount), eth.WeiToEth(bond.RewardAmount))
	}
	selectedClaims, err := utils.GetMultiselectIndices(c, proposalFlag.Name, options, "Please select a proposal to unlock bonds / claim rewards from:")
	if err != nil {
		return fmt.Errorf("error determining proposal selection: %w", err)
	}

	// Build the TXs
	claims := make([]api.ProtocolDaoClaimBonds, len(selectedClaims))
	for i, bond := range selectedClaims {
		claims[i].ProposalID = bond.ProposalID
		claims[i].Indices = getClaimIndicesForBond(bond)
	}
	response, err := rp.Api.PDao.ClaimBonds(claims)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedClaims))
	for i, bond := range selectedClaims {
		data := response.Data.Batch[i]
		if !data.CanClaim {
			fmt.Printf("Cannot claim rewards from proposal %d:", bond.ProposalID)
			if data.DoesNotExist {
				fmt.Println("The proposal does not exist.")
			}
			if data.InvalidState {
				fmt.Println("The proposal is not in a claimable state.")
			}
			return nil
		}
		txs[i] = data.TxInfo
	}

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to claim bonds and rewards from %d proposals?", len(selectedClaims)),
		func(i int) string {
			return fmt.Sprintf("claim of proposal %d", selectedClaims[i].ProposalID)
		},
		"Claiming bonds from proposals...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully claimed bonds from all selected proposals.")
	return nil
}

func getClaimIndicesForBond(bond *api.BondClaimResult) []uint64 {
	indexMap := map[uint64]bool{}
	for _, index := range bond.UnlockableIndices {
		indexMap[index] = true
	}
	for _, index := range bond.RewardableIndices {
		indexMap[index] = true
	}

	indices := make([]uint64, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}

	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i] < indices[j]
	})

	return indices
}
