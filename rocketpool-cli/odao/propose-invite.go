package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func proposeInvite(c *cli.Context, memberAddress common.Address, memberId, memberEmail string) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeInviteToTNDAO(memberAddress)
    if err != nil {
        return err
    }
    if !canPropose.CanPropose {
        fmt.Println("Cannot propose inviting member:")
        if canPropose.ProposalCooldownActive {
            fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
        }
        if canPropose.MemberAlreadyExists {
            fmt.Printf("The node %s is already a member of the oracle DAO.\n", memberAddress.Hex())
        }
        return nil
    }

    // Submit proposal
    response, err := rp.ProposeInviteToTNDAO(memberAddress, memberId, memberEmail)
    if err != nil {
        return err
    }

    fmt.Printf("Inviting %s to the oracle DAO...\n", memberAddress.Hex())
    cliutils.PrintTransactionHash(response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted an invite proposal with ID %d for node %s.\n", response.ProposalId, memberAddress.Hex())
    return nil

}

