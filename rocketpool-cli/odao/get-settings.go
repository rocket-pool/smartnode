package odao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func getMemberSettings(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get oracle DAO member settings
    response, err := rp.GetTNDAOMemberSettings()
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("ODAO Voting Quorum Threshold: %f%%\n", response.Quorum * 100)
    fmt.Printf("Required Member RPL Bond: %f RPL\n", eth.WeiToEth(response.RPLBond))
    fmt.Printf("Max Number of Unbonded Minipools: %d\n", response.MinipoolUnbondedMax)
    fmt.Printf("Consecutive Challenge Cooldown: %d Blocks\n", response.ChallengeCooldown)
    fmt.Printf("Challenge Meeting Window: %d Blocks\n", response.ChallengeWindow)
    fmt.Printf("Cost for Non-members to Challenge Members: %f ETH\n", eth.WeiToEth(response.ChallengeCost))
    return nil

}


func getProposalSettings(c* cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get oracle DAO proposal settings
    response, err := rp.GetTNDAOProposalSettings()
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Cooldown Between Proposals: %d Blocks\n", response.Cooldown)
    fmt.Printf("Proposal Voting Window: %d Blocks\n", response.VoteBlocks)
    fmt.Printf("Delay Before Voting on a Proposal is Allowed: %d Blocks\n", response.VoteDelayBlocks)
    fmt.Printf("Window to Execute an Accepted Proposal: %d Blocks\n", response.ExecuteBlocks)
    fmt.Printf("Window to Act on an Executed Proposal: %d Blocks\n", response.ActionBlocks)
    return nil

}

