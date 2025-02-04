package megapool

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/types/api"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

const (
	TimeFormat        = "2006-01-02, 15:04 -0700 MST"
	colorBlue  string = "\033[36m"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Check if Saturn is deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Get Megapool status
	status, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	// Return if megapool isn't deployed
	if !status.Megapool.Deployed {
		fmt.Println("The node does not have a megapool. Please run 'rocketpool megapool deploy-megapool' and try again.")
		return nil
	}

	fmt.Printf("%s=== Megapool ===%s\n", colorGreen, colorReset)
	fmt.Printf("The node has a megapool deployed at %s%s%s\n", colorBlue, status.Megapool.Address, colorReset)
	fmt.Printf("The node has %d express ticket(s).\n", status.Megapool.NodeExpressTicketCount)
	fmt.Printf("The megapool has %d validators.\n", status.Megapool.ValidatorCount)
	fmt.Println("")

	fmt.Printf("%s=== Megapool Delegate ===%s\n", colorGreen, colorReset)
	if status.Megapool.EffectiveDelegateAddress == status.LatestDelegate {
		fmt.Printf("The megapool is using the latest delegate at %s%s%s\n", colorBlue, status.LatestDelegate, colorReset)
	} else {
		fmt.Printf("The megapool is using an outdated delegate at %s%s%s\n", colorBlue, status.Megapool.DelegateAddress, colorReset)
		fmt.Printf("The megapool can be upgraded to delegate %s%s%s using 'rocketpool megapool delegate-upgrade'.\n", colorBlue, status.LatestDelegate, colorReset)
	}
	fmt.Printf("The megapool's effective delegate address is %s%s%s\n", colorBlue, status.Megapool.EffectiveDelegateAddress, colorReset)

	if status.Megapool.UseLatestDelegate {
		fmt.Println("The megapool is set to automatically upgrade to the latest delegate. You can toggle this setting using 'rocketpool megapool set-use-latest-delegate'.")
	} else {
		fmt.Println("The megapool has automatic delegate upgrades disabled. You can toggle this setting using 'rocketpool megapool set-use-latest-delegate'.")
		if status.Megapool.DelegateExpiry > 0 {
			fmt.Printf("Your current megapool delegate expires at %sblock %d%s.\n", colorBlue, status.Megapool.DelegateExpiry, colorReset)
		}
	}
	fmt.Println("")

	beaconBalance := new(big.Int)
	beaconBalance.Add(status.Megapool.UserCapital, status.Megapool.NodeCapital)

	totalBond := new(big.Int).Mul(status.Megapool.NodeBond, big.NewInt(4))
	rpBond := new(big.Int).Sub(totalBond, status.Megapool.NodeBond)

	fmt.Printf("%s=== Megapool Balance ===%s\n", colorGreen, colorReset)
	fmt.Printf("The megapool has %6f node bonded ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.NodeBond), 6))
	fmt.Printf("The megapool has %6f RP ETH for a total of %6f bonded ETH.\n", math.RoundDown(eth.WeiToEth(rpBond), 6), math.RoundDown(eth.WeiToEth(totalBond), 6))
	fmt.Printf("Beacon balance (CL): %6f ETH.\n", math.RoundDown(eth.WeiToEth(beaconBalance), 6))
	fmt.Printf("Your portion: %6f ETH\n", math.RoundDown(eth.WeiToEth(status.Megapool.UserCapital), 6))
	if status.Megapool.NodeDebt.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("The megapool debt is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.NodeDebt), 6))
	}
	if status.Megapool.RefundValue.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("The megapool refund value is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.RefundValue), 6))
	}
	if status.Megapool.PendingRewards.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("The megapool has %.6f ETH in pending rewards to claim.\n", math.RoundDown(eth.WeiToEth(status.Megapool.PendingRewards), 6))
	} else {
		fmt.Println("The megapool does not have any pending rewards to claim.")
	}

	networkCommission := math.RoundDown(eth.WeiToEth(status.Megapool.NodeShare)*100, 6)
	effectiveNodeShare := math.RoundDown(eth.WeiToEth(status.Megapool.RevenueSplit.NodeShare)*100, 6)

	fmt.Printf("Current network commission: %.6f%%\n", networkCommission)
	if networkCommission != effectiveNodeShare {
		fmt.Printf("Effective node share: %.6f%% (time-weighted average due to universal commission changes since last distribution).\n", effectiveNodeShare)
	}
	fmt.Println("")

	return nil

}

func getValidatorStatus(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Check if Saturn is deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Get Megapool status
	status, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	// Return if megapool isn't deployed
	if !status.Megapool.Deployed {
		fmt.Println("The node does not have a megapool. Please run 'rocketpool megapool deploy-megapool' and try again.")
		return nil
	}

	statusValidators := map[string][]api.MegapoolValidatorDetails{
		"Staking":     {},
		"Exited":      {},
		"Initialized": {},
		"Prelaunch":   {},
	}

	statusName := []string{"Staking", "Exited", "Prelaunch", "Initialized"}

	// Iterate over the validators and append them based on their statuses
	for _, validator := range status.Megapool.Validators {
		if validator.Staked {
			statusValidators["Staking"] = append(statusValidators["Staking"], validator)
		}
		if validator.Exited {
			statusValidators["Exited"] = append(statusValidators["Exited"], validator)
		}
		if validator.InQueue {
			statusValidators["Initialized"] = append(statusValidators["Initialized"], validator)
		}
		if validator.InPrestake {
			statusValidators["Prelaunch"] = append(statusValidators["Prelaunch"], validator)
		}
	}

	// Print validators by status
	noValidators := true
	for _, status := range statusName {
		if validators := statusValidators[status]; len(validators) > 0 {
			noValidators = false
			sort.Slice(validators, func(i, j int) bool {
				return validators[i].ValidatorId < validators[j].ValidatorId
			})

			fmt.Printf("%d %s validator(s):\n", len(validators), status)
			fmt.Println("")
			for _, validator := range validators {
				printValidatorDetails(validator, status)
			}
		}
	}
	if noValidators {
		fmt.Println("The megapool does not have any validators yet.")
	}

	return nil

}

func printValidatorDetails(validator api.MegapoolValidatorDetails, status string) {

	fmt.Printf("--------------------\n")
	fmt.Println("")

	if status == "Prelaunch" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             yes\n")
		fmt.Printf("Validator index:              \n")
	}

	if status == "Staking" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             yes\n")
		fmt.Printf("Validator index:              \n")
	}

	if status == "Initialized" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Expected pubkey:              0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Validator Queue Position:     \n")
	}

	// Main details
	if validator.ExpressUsed {
		fmt.Printf("Express Ticket Used:          yes\n")
	} else {
		fmt.Printf("Express Ticket Used:          no\n")
	}

	fmt.Println("")

}
