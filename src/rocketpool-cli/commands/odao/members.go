package odao

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
)

func getMembers(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get oracle DAO members
	members, err := rp.Api.ODao.Members()
	if err != nil {
		return err
	}

	// Print & return
	if len(members.Data.Members) > 0 {
		fmt.Printf("The oracle DAO has %d members:\n", len(members.Data.Members))
		fmt.Println("")
	} else {
		fmt.Println("The oracle DAO does not have any members yet.")
	}
	for _, member := range members.Data.Members {
		fmt.Printf("--------------------\n")
		fmt.Printf("\n")
		fmt.Printf("Member ID:            %s\n", member.ID)
		fmt.Printf("URL:                  %s\n", member.Url)
		fmt.Printf("Node address:         %s\n", member.Address.Hex())
		fmt.Printf("Joined at:            %s\n", utils.GetDateTimeStringOfTime(member.JoinedTime))
		fmt.Printf("Last proposal:        %s\n", utils.GetDateTimeStringOfTime(member.LastProposalTime))
		fmt.Printf("RPL bond amount:      %.6f\n", math.RoundDown(eth.WeiToEth(member.RplBondAmount), 6))
		fmt.Printf("\n")
	}
	return nil
}
