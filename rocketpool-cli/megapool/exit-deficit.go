package megapool

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func exitDeficit(address common.Address, yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	plan, err := rp.MegapoolDeficit(address)
	if err != nil {
		return err
	}
	if !plan.Saturn2Deployed {
		fmt.Println(cliutils.Saturn2NotDeployedMessage)
		return nil
	}
	formatETH := func(value *big.Int) string {
		return new(big.Rat).SetFrac(value, big.NewInt(1e18)).FloatString(18)
	}
	fmt.Printf("Megapool: %s\n", address.Hex())
	fmt.Printf("Current deficit: %s ETH\nDeficit required to permit an exit: %s ETH\n", formatETH(plan.Deficit), formatETH(plan.ExitDeficit))
	if plan.ExistingExits > 0 {
		fmt.Printf("Validators already exiting: %d\nDeficit after those exits: %s ETH\n", plan.ExistingExits, formatETH(plan.DeficitAfterPendingExits))
	}
	if !plan.CanExit {
		fmt.Println(plan.Reason)
		return nil
	}
	fmt.Printf("Validators to exit: %d\nProjected deficit after these exits: %s ETH\nValidator IDs: %v\n", len(plan.ValidatorIds), formatETH(plan.ProjectedDeficit), plan.ValidatorIds)
	if plan.ProjectedDeficit.Cmp(plan.ExitDeficit) >= 0 {
		fmt.Println("The deficit stays at or above the amount required to permit an exit.")
	}
	ids := make([]string, len(plan.ValidatorIds))
	for i, id := range plan.ValidatorIds {
		ids[i] = strconv.FormatUint(uint64(id), 10)
	}
	validatorIds := strings.Join(ids, ",")
	can, err := rp.CanExitMegapoolDeficit(address, validatorIds)
	if err != nil {
		return err
	}
	if !can.CanExit {
		return fmt.Errorf("the selected validators cannot be exited based on this megapool's deficit")
	}
	if can.ExitFee == nil {
		return fmt.Errorf("the API did not return the total exit fee")
	}

	exitFeeEth := math.RoundDown(math.WeiToEth(can.ExitFee), 6)
	fmt.Printf("Total EIP-7002 exit fee: %.6f ETH, in addition to transaction gas.\n", exitFeeEth)
	if err := gas.AssignMaxFeeAndLimit(can.GasLimits, rp, yes); err != nil {
		return err
	}
	if prompt.Declined(yes, "Force exit %d validators from megapool %s? This pays %.6f ETH in EIP-7002 exit fees plus gas.", len(can.ValidatorIds), address.Hex(), exitFeeEth) {
		fmt.Println("Cancelled.")
		return nil
	}

	result, err := rp.ExitMegapoolDeficit(address, validatorIds, can.ExitFee)
	if err != nil {
		return err
	}
	fmt.Println("Submitting the deficit-triggered exits...")
	cliutils.PrintTransactionHash(rp, result.TxHash)
	if _, err := rp.WaitForTransaction(result.TxHash); err != nil {
		return err
	}
	fmt.Printf("Successfully submitted exit requests for %d validators. Beacon-chain exit processing is pending.\n", len(can.ValidatorIds))
	return nil
}
