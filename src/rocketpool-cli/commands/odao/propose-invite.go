package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func proposeInvite(c *cli.Context, memberAddress common.Address, memberId, memberUrl string) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.ODao.ProposeInvite(memberAddress, memberId, memberUrl)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("Cannot propose inviting member:")
		if response.Data.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		if response.Data.MemberAlreadyExists {
			fmt.Printf("The node %s is already a member of the oracle DAO.\n", memberAddress.Hex())
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to submit this proposal?",
		"inviting member",
		"Inviting %s to the oracle DAO...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully submitted an invite proposal for node %s.\n", memberAddress.Hex())
	return nil
}
