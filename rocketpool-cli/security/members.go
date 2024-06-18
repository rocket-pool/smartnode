package security

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func getMembers(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get security council members
	members, err := rp.SecurityMembers()
	if err != nil {
		return err
	}

	// Print & return
	if len(members.Members) > 0 {
		fmt.Printf("The security council has %d members:\n", len(members.Members))
		fmt.Println("")
	} else {
		fmt.Println("The security council does not have any members yet.")
	}
	for _, member := range members.Members {
		fmt.Printf("--------------------\n")
		fmt.Printf("\n")
		fmt.Printf("Member ID:            %s\n", member.ID)
		fmt.Printf("Node address:         %s\n", member.Address.Hex())
		fmt.Printf("Joined at:            %s\n", cliutils.GetDateTimeString(member.JoinedTime))
		fmt.Printf("\n")
	}
	return nil

}
