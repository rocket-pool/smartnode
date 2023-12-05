package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func parseFloat(c *cli.Context, name string, value string) (*big.Int, error) {
	if c.Bool("raw") {
		val, err := cliutils.ValidateBigInt(name, value)
		if err != nil {
			return nil, err
		}
		return val, nil
	} else {
		val, err := cliutils.ValidateFraction(name, value)
		if err != nil {
			return nil, err
		}

		trueVal := eth.EthToWei(val)
		fmt.Printf("Your value will be multiplied by 10^18 to be used in the contracts, which results in:\n\n\t[%s]\n\n", trueVal.String())
		if !(c.Bool("yes") || cliutils.Confirm("Please make sure this is what you want and does not have any floating-point errors.\n\nIs this result correct?")) {
			fmt.Println("Cancelled. Please try again with the '--raw' flag and provide an explicit value instead.")
			return nil, nil
		}
		return trueVal, nil
	}
}
