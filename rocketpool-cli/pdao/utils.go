package pdao

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func parseFloat(c *cli.Context, name string, value string, isFraction bool) (*big.Int, error) {
	var floatValue float64
	if c.Bool("raw") {
		val, err := cliutils.ValidatePositiveWeiAmount(name, value)
		if err != nil {
			return nil, err
		}
		return val, nil
	} else if isFraction {
		val, err := cliutils.ValidateFraction(name, value)
		if err != nil {
			return nil, err
		}
		floatValue = val
	} else {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		floatValue = val
	}

	trueVal := eth.EthToWei(floatValue)
	fmt.Printf("Your value will be multiplied by 10^18 to be used in the contracts, which results in:\n\n\t[%s]\n\n", trueVal.String())
	if !(c.Bool("yes") || prompt.Confirm("Please make sure this is what you want and does not have any floating-point errors.\n\nIs this result correct?")) {
		value = prompt.Prompt("Please enter the wei amount:", "^[0-9]+$", "Invalid amount")
		val, err := cliutils.ValidatePositiveWeiAmount(name, value)
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return trueVal, nil
}
