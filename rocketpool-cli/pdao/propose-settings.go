package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func proposeSettingAuctionIsCreateLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.CreateLotEnabledSettingPath, trueValue)
}

func proposeSettingAuctionIsBidOnLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.BidOnLotEnabledSettingPath, trueValue)
}

func proposeSettingLotMinimumEthValue(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.LotMinimumEthValueSettingPath, trueValue)
}

func proposeSettingLotMaximumEthValue(c *cli.Context, value *big.Int) error {
	trueValue := value.String()
	return proposeSetting(c, protocol.LotMaximumEthValueSettingPath, trueValue)
}

func proposeSettingLotDuration(c *cli.Context, value time.Duration) error {
	trueValue := fmt.Sprint(uint64(value.Seconds()))
	return proposeSetting(c, protocol.LotDurationSettingPath, trueValue)
}

func proposeSettingLotStartingPriceRatio(c *cli.Context, value string) error {
	trueValue, err := parseFloat(c, value)
	if err != nil {
		return err
	}
	if trueValue == "" {
		return nil
	}
	return proposeSetting(c, protocol.LotStartingPriceRatioSettingPath, trueValue)
}

// Master general proposal function
func proposeSetting(c *cli.Context, setting string, value string) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if proposal can be made
	canPropose, err := rp.PDAOCanProposeSetting(setting, value)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.InsufficientRpl {
			fmt.Printf("You do not have enough RPL staked but unlocked to make another proposal (unlocked: %.6f RPL, required: %.6f RPL).\n",
				eth.WeiToEth(big.NewInt(0).Sub(canPropose.StakedRpl, canPropose.LockedRpl)), eth.WeiToEth(canPropose.ProposalBond),
			)
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.PDAOProposeSetting(setting, value, canPropose.BlockNumber, canPropose.Pollard)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a %s setting update proposal with ID %d.\n", setting, response.ProposalId)
	return nil
}

func parseFloat(c *cli.Context, value string) (string, error) {
	if c.Bool("raw") {
		val, err := cliutils.ValidateBigInt("value", value)
		if err != nil {
			return "", err
		}
		return val.String(), nil
	} else {
		val, err := cliutils.ValidateFraction("value", value)
		if err != nil {
			return "", err
		}

		trueVal := eth.EthToWei(val)
		if !cliutils.Confirm("Your value will be multiplied by 10^18 to be stored in the contracts, which results in:\n\n\t[%s]\n\nPlease make sure this is what you want and does not have any rounding errors.\n\nIs this result correct?") {
			fmt.Println("Cancelled. Please try again with the '-raw' flag and provide an explicit value instead.")
			return "", nil
		}
		return trueVal.String(), nil
	}
}
