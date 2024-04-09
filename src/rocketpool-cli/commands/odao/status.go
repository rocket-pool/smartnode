package odao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get oracle DAO status
	status, err := rp.Api.ODao.Status()
	if err != nil {
		return err
	}

	// Get failed proposal count
	failedProposalCount := (status.Data.ProposalCounts.Cancelled + status.Data.ProposalCounts.Defeated + status.Data.ProposalCounts.Expired)

	// Membership status
	if status.Data.IsMember {
		fmt.Println("The node is a member of the oracle DAO - it can create unbonded minipools, vote on DAO proposals and perform watchtower duties.")
		if status.Data.CanLeave {
			fmt.Println("The node has an executed proposal to leave - you can leave the oracle DAO with 'rocketpool odao leave'")
		}
		if status.Data.CanReplace {
			fmt.Println("The node has an executed proposal to replace itself - you can replace your position in the oracle DAO with 'rocketpool odao replace'")
		}
	} else {
		fmt.Println("The node is not a member of the oracle DAO.")
		if status.Data.CanJoin {
			fmt.Println("The node has an executed proposal to join - you can join the oracle DAO with 'rocketpool odao join'")
		}
	}
	fmt.Println("")

	// Members
	fmt.Printf("There are currently %d member(s) in the oracle DAO.\n", status.Data.TotalMembers)
	fmt.Println("")

	// Proposals
	if status.Data.ProposalCounts.Total > 0 {
		fmt.Printf("There are %d oracle DAO proposal(s) in total:\n", status.Data.ProposalCounts.Total)
		if status.Data.ProposalCounts.Pending > 0 {
			fmt.Printf("- %d proposal(s) are pending and cannot be voted on yet\n", status.Data.ProposalCounts.Pending)
		}
		if status.Data.ProposalCounts.Active > 0 {
			fmt.Printf("- %d proposal(s) are active and can be voted on\n", status.Data.ProposalCounts.Active)
		}
		if status.Data.ProposalCounts.Succeeded > 0 {
			fmt.Printf("- %d proposal(s) have passed and can be executed\n", status.Data.ProposalCounts.Succeeded)
		}
		if status.Data.ProposalCounts.Executed > 0 {
			fmt.Printf("- %d proposal(s) have passed and been executed\n", status.Data.ProposalCounts.Executed)
		}
		if failedProposalCount > 0 {
			fmt.Printf("- %d proposal(s) were cancelled, defeated, or have expired\n", failedProposalCount)
		}
	} else {
		fmt.Println("There are no oracle DAO proposals.")
	}

	// Return
	return nil
}
