package security

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao"
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	psettings "github.com/rocket-pool/smartnode/bindings/settings/protocol"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"golang.org/x/sync/errgroup"
)

// Settings
const ProposalStatesBatchSize = 50

// Check if a proposal for a node exists & is actionable
func getProposalIsActionable(rp *rocketpool.RocketPool, nodeAddress common.Address, proposalType string) (bool, error) {
	// Data
	var wg errgroup.Group
	var proposalExecutedTimeUnix uint64
	var actionTime time.Duration

	// Get proposal executed time
	wg.Go(func() error {
		var err error
		proposalExecutedTimeUnix, err = security.GetMemberProposalExecutedTime(rp, proposalType, nodeAddress, nil)
		return err
	})

	// Get action window
	wg.Go(func() error {
		var err error
		actionTime, err = psettings.GetSecurityProposalActionTime(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return false, err
	}

	// Return
	proposalExecutedTime := time.Unix(int64(proposalExecutedTimeUnix), 0)
	return time.Until(proposalExecutedTime.Add(actionTime)) > 0, nil
}

// Get all proposal states
func getProposalStates(rp *rocketpool.RocketPool) ([]rptypes.ProposalState, error) {
	// Get proposal IDs
	proposalIds, err := dao.GetDAOProposalIDs(rp, "rocketDAOSecurityProposals", nil)
	if err != nil {
		return []rptypes.ProposalState{}, err
	}

	// Load proposal states in batches
	states := make([]rptypes.ProposalState, len(proposalIds))
	for bsi := 0; bsi < len(proposalIds); bsi += ProposalStatesBatchSize {
		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalStatesBatchSize
		if pei > len(proposalIds) {
			pei = len(proposalIds)
		}

		// Load states
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalState, err := dao.GetProposalState(rp, proposalIds[pi], nil)
				if err == nil {
					states[pi] = proposalState
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []rptypes.ProposalState{}, err
		}
	}

	// Return
	return states, nil
}
