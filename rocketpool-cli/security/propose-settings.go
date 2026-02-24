package security

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
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

func proposeSettingNodeComissionShareSecurityCouncilAdder(c *cli.Context, value *big.Int) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(c, protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath, trueValue)
}

// Master general proposal function
func proposeSetting(c *cli.Context, contract string, setting string, value string) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

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
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
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
