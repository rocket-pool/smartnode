package odao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
	}

	// Get oracle DAO status
	status, err := rp.TNDAOStatus()
	if err != nil {
		return err
	}

	// Get failed proposal count
	failedProposalCount := (status.ProposalCounts.Cancelled + status.ProposalCounts.Defeated + status.ProposalCounts.Expired)

	// Membership status
	if status.IsMember {
		fmt.Println("The node is a member of the oracle DAO - it can create unbonded minipools, vote on DAO proposals and perform watchtower duties.")
		if status.CanLeave {
			fmt.Println("The node has an executed proposal to leave - you can leave the oracle DAO with 'rocketpool odao leave'")
		}
		if status.CanReplace {
			fmt.Println("The node has an executed proposal to replace itself - you can replace your position in the oracle DAO with 'rocketpool odao replace'")
		}
	} else {
		fmt.Println("The node is not a member of the oracle DAO.")
		if status.CanJoin {
			fmt.Println("The node has an executed proposal to join - you can join the oracle DAO with 'rocketpool odao join'")
		}
	}
	fmt.Println("")

	// Members
	fmt.Printf("There are currently %d member(s) in the oracle DAO.\n", status.TotalMembers)
	fmt.Println("")

	// Proposals
	if status.ProposalCounts.Total > 0 {
		fmt.Printf("There are %d oracle DAO proposal(s) in total:\n", status.ProposalCounts.Total)
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
		fmt.Println("There are no oracle DAO proposals.")
	}

	// Return
	return nil

}
