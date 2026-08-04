package odao

import (
	"fmt"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getMembers() error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get oracle DAO members
	members, err := rp.TNDAOMembers()
	if err != nil {
		return err
	}

	// Print & return
	if len(members.Members) > 0 {
		fmt.Printf("The oracle DAO has %d members:\n", len(members.Members))
		fmt.Println("")
	} else {
		fmt.Println("The oracle DAO does not have any members yet.")
	}
	for _, member := range members.Members {
		fmt.Printf("--------------------\n")
		fmt.Printf("\n")
		fmt.Printf("Member ID:            %s\n", member.ID)
		fmt.Printf("URL:                  %s\n", member.Url)
		fmt.Printf("Node address:         %s\n", member.Address.Hex())
		fmt.Printf("Joined at:            %s\n", cliutils.GetDateTimeString(member.JoinedTime))
		fmt.Printf("Last proposal:        %s\n", cliutils.GetDateTimeString(member.LastProposalTime))
		fmt.Printf("RPL bond amount:      %.6f\n", math.RoundDown(math.WeiToEth(member.RPLBondAmount), 6))
		fmt.Printf("\n")
	}
	return nil

}
