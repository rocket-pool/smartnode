package odao

import (
	"fmt"

	"github.com/urfave/cli"

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
    fmt.Printf("ODAO Voting Quorum Threshold: %f", response.Quorum)
    fmt.Printf("Required Member RPL Bond: %s", response.RPLBond.String())
    fmt.Printf("Max Number of Unbonded Minipools: %d", response.MinipoolUnbondedMax)
    return nil

}