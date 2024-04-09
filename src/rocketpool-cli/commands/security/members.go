package security

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
)

func getMembers(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get security council members
	members, err := rp.Api.Security.Members()
	if err != nil {
		return err
	}

	// Print & return
	if len(members.Data.Members) > 0 {
		fmt.Printf("The security council has %d members:\n", len(members.Data.Members))
		fmt.Println("")
	} else {
		fmt.Println("The security council does not have any members yet.")
	}
	for _, member := range members.Data.Members {
		fmt.Printf("--------------------\n")
		fmt.Printf("\n")
		fmt.Printf("Member ID:            %s\n", member.ID)
		fmt.Printf("Node address:         %s\n", member.Address.Hex())
		fmt.Printf("Joined at:            %s\n", utils.GetDateTimeStringOfTime(member.JoinedTime))
		fmt.Printf("\n")
	}
	return nil
}
