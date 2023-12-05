package security

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getStatus(c *cli.Context) error {

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

	// Get security council status
	status, err := rp.SecurityStatus()
	if err != nil {
		return err
	}

	// Get failed proposal count
	failedProposalCount := (status.ProposalCounts.Cancelled + status.ProposalCounts.Defeated + status.ProposalCounts.Expired)

	// Membership status
	if status.IsMember {
		fmt.Println("The node is a member of the security council - it can propose changing certain Protocol DAO settings without a cooldown.")
		if status.CanLeave {
			fmt.Println("The node has an executed proposal to leave - you can leave the security council with 'rocketpool security leave'")
		}
	} else {
		fmt.Println("The node is not a member of the security council.")
		if status.CanJoin {
			fmt.Println("The node has an executed proposal to join - you can join the security council with 'rocketpool security join'")
		}
	}
	fmt.Println("")

	// Members
	fmt.Printf("There are currently %d member(s) in the security council.\n", status.TotalMembers)
	fmt.Println("")

	// Proposals
	if status.ProposalCounts.Total > 0 {
		fmt.Printf("There are %d security council proposal(s) in total:\n", status.ProposalCounts.Total)
		if status.ProposalCounts.Pending > 0 {
			fmt.Printf("- %d proposal(s) are pending and cannot be voted on yet\n", status.ProposalCounts.Pending)
		}
		if status.ProposalCounts.Active > 0 {
			fmt.Printf("- %d proposal(s) are active and can be voted on\n", status.ProposalCounts.Active)
		}
		if status.ProposalCounts.Succeeded > 0 {
			fmt.Printf("- %d proposal(s) have passed and can be executed\n", status.ProposalCounts.Succeeded)
		}
		if status.ProposalCounts.Executed > 0 {
			fmt.Printf("- %d proposal(s) have passed and been executed\n", status.ProposalCounts.Executed)
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
