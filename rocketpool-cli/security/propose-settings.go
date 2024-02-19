package security

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func proposeSettingAuctionIsCreateLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, trueValue)
}

func proposeSettingAuctionIsBidOnLotEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.AuctionSettingsContractName, protocol.BidOnLotEnabledSettingPath, trueValue)
}

func proposeSettingDepositIsDepositingEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.DepositEnabledSettingPath, trueValue)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.DepositSettingsContractName, protocol.AssignDepositsEnabledSettingPath, trueValue)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue)
}

func proposeSettingMinipoolIsBondReductionEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.MinipoolSettingsContractName, protocol.BondReductionEnabledSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitBalancesEnabledSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitPricesEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitPricesEnabledSettingPath, trueValue)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.SubmitRewardsEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsRegistrationEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.NodeRegistrationEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue)
}

func proposeSettingNodeIsDepositingEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.NodeDepositEnabledSettingPath, trueValue)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(c *cli.Context, value bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NodeSettingsContractName, protocol.VacantMinipoolsEnabledSettingPath, trueValue)
}

// Master general proposal function
func proposeSetting(c *cli.Context, contract string, setting string, value string) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Check if proposal can be made
	canPropose, err := rp.SecurityCanProposeSetting(contract, setting, value)
	if err != nil {
		return err
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
	response, err := rp.SecurityProposeSetting(contract, setting, value)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a %s setting update proposal.\n", setting)
	return nil
}
