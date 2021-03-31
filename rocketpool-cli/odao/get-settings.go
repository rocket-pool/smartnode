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

    // Submit proposal
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