package node

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)


func nodeClaimRpl(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check for rewards
    canClaim, err := rp.CanNodeClaimRpl()
    if err != nil {
        return err
    }
    if canClaim.RplAmount.Cmp(big.NewInt(0)) == 0 {
        fmt.Println("The node does not have any available RPL rewards to claim.")
        return nil
    } else {
        fmt.Printf("%.6f RPL is available to claim.\n", math.RoundDown(eth.WeiToEth(canClaim.RplAmount), 6))
    }

    // Assign max fees
    err = services.AssignMaxFeeAndLimit(canClaim.GasInfo, rp, c.Bool("yes"))
    if err != nil{
        return err
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to claim your RPL?")) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Claim rewards
    response, err := rp.NodeClaimRpl()
    if err != nil {
        return err
    }

    fmt.Printf("Claiming RPL...\n")
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully claimed %.6f RPL in rewards.", math.RoundDown(eth.WeiToEth(canClaim.RplAmount), 6))
    return nil

}

