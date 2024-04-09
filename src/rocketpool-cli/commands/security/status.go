package security

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

	// Get security council status
	status, err := rp.Api.Security.Status()
	if err != nil {
		return err
	}

	// Get failed proposal count
	failedProposalCount := (status.Data.ProposalCounts.Cancelled + status.Data.ProposalCounts.Defeated + status.Data.ProposalCounts.Expired)

	// Membership status
	if status.Data.IsMember {
		fmt.Println("The node is a member of the security council - it can propose changing certain Protocol DAO settings without a cooldown.")
		if status.Data.CanLeave {
			fmt.Println("The node has an executed proposal to leave - you can leave the security council with 'rocketpool security leave'")
		}
	} else {
		fmt.Println("The node is not a member of the security council.")
		if status.Data.CanJoin {
			fmt.Println("The node has an executed proposal to join - you can join the security council with 'rocketpool security join'")
		}
	}
	fmt.Println()

	// Members
	fmt.Printf("There are currently %d member(s) in the security council.\n", status.Data.TotalMembers)
	fmt.Println()

	// Proposals
	if status.Data.ProposalCounts.Total > 0 {
		fmt.Printf("There are %d security council proposal(s) in total:\n", status.Data.ProposalCounts.Total)
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
		fmt.Println("There are no security council proposals.")
	}

	// Return
	return nil
}
