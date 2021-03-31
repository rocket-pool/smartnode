package odao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
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
    return nil

}