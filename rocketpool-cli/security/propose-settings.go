package security

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/settings/protocol"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeSettingAuctionIsCreateLotEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.CreateLotEnabledSettingPath, trueValue, yes)
}

func proposeSettingAuctionIsBidOnLotEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.AuctionSettingsContractName, protocol.BidOnLotEnabledSettingPath, trueValue, yes)
}

func proposeSettingDepositIsDepositingEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.DepositEnabledSettingPath, trueValue, yes)
}

func proposeSettingDepositAreDepositAssignmentsEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.DepositSettingsContractName, protocol.AssignDepositsEnabledSettingPath, trueValue, yes)
}

func proposeSettingMinipoolIsSubmitWithdrawableEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.MinipoolSubmitWithdrawableEnabledSettingPath, trueValue, yes)
}

func proposeSettingMinipoolIsBondReductionEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.MinipoolSettingsContractName, protocol.BondReductionEnabledSettingPath, trueValue, yes)
}

func proposeSettingNetworkIsSubmitBalancesEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitBalancesEnabledSettingPath, trueValue, yes)
}

func proposeSettingNetworkIsSubmitPricesEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitPricesEnabledSettingPath, trueValue, yes)
}

func proposeSettingNetworkIsSubmitRewardsEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.SubmitRewardsEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeIsRegistrationEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeRegistrationEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeIsSmoothingPoolRegistrationEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.SmoothingPoolRegistrationEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeIsDepositingEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.NodeDepositEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeAreVacantMinipoolsEnabled(value bool, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NodeSettingsContractName, protocol.VacantMinipoolsEnabledSettingPath, trueValue, yes)
}

func proposeSettingNodeComissionShareSecurityCouncilAdder(value *big.Int, yes bool) error {
	trueValue := fmt.Sprint(value)
	return proposeSetting(protocol.NetworkSettingsContractName, protocol.NetworkNodeCommissionShareSecurityCouncilAdderPath, trueValue, yes)
}

// Master general proposal function
func proposeSetting(contract string, setting string, value string, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
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
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to submit this proposal?")) {
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
