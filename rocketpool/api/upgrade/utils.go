package upgrade

import (
	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"golang.org/x/sync/errgroup"
)

// Settings
const UpgradeProposalStatesBatchSize = 50

// Get all proposal states
func getUpgradeProposalStates(rp *rocketpool.RocketPool) ([]rptypes.UpgradeProposalState, error) {

	// Get proposal IDs
	proposalIds, err := dao.GetDAOProposalIDs(rp, "rocketDAONodeTrustedProposals", nil)
	if err != nil {
		return []rptypes.UpgradeProposalState{}, err
	}

	// Load proposal states in batches
	states := make([]rptypes.UpgradeProposalState, len(proposalIds))
	for bsi := 0; bsi < len(proposalIds); bsi += UpgradeProposalStatesBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + UpgradeProposalStatesBatchSize
		if pei > len(proposalIds) {
			pei = len(proposalIds)
		}

		// Load states
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalState, err := upgrades.GetUpgradeProposalState(rp, proposalIds[pi], nil)
				if err == nil {
					states[pi] = proposalState
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []rptypes.UpgradeProposalState{}, err
		}

	}

	// Return
	return states, nil

}
