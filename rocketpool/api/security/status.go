package security

import (
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.SecurityStatusResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SecurityStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get membership status
	isMember, err := security.GetMemberExists(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.IsMember = isMember

	// Sync
	var wg errgroup.Group

	// Get pending executed proposal statuses
	if isMember {

		// Check if node can leave
		wg.Go(func() error {
			leaveActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "leave")
			if err == nil {
				response.CanLeave = leaveActionable
			}
			return err
		})

	} else {

		// Check if node can join
		wg.Go(func() error {
			joinActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "invited")
			if err == nil {
				response.CanJoin = joinActionable
			}
			return err
		})

	}

	// Get total DAO members
	wg.Go(func() error {
		memberCount, err := security.GetMemberCount(rp, nil)
		if err == nil {
			response.TotalMembers = memberCount
		}
		return err
	})

	// Get proposal counts
	wg.Go(func() error {
		proposalStates, err := getProposalStates(rp)
		if err == nil {
			response.ProposalCounts.Total = len(proposalStates)
			for _, state := range proposalStates {
				switch state {
				case rptypes.Pending:
					response.ProposalCounts.Pending++
				case rptypes.Active:
					response.ProposalCounts.Active++
				case rptypes.Cancelled:
					response.ProposalCounts.Cancelled++
				case rptypes.Defeated:
					response.ProposalCounts.Defeated++
				case rptypes.Succeeded:
					response.ProposalCounts.Succeeded++
				case rptypes.Expired:
					response.ProposalCounts.Expired++
				case rptypes.Executed:
					response.ProposalCounts.Executed++
				}
			}
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
