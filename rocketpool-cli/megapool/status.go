package megapool

import (
	"fmt"
	"math/big"

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
		fmt.Println("The node does not have a megapool.")
		return nil
	}

	fmt.Printf("%s=== Megapool ===%s\n", colorGreen, colorReset)
	fmt.Printf("The node has a megapool deployed at %s%s%s\n", colorBlue, status.Megapool.Address.Hex(), colorReset)
	fmt.Printf("The node has %d express ticket(s).\n", status.Megapool.NodeExpressTicketCount)
	fmt.Printf("The megapool has %d validators.\n", status.Megapool.ValidatorCount)
	fmt.Println("")

	fmt.Printf("%s=== Megapool Delegate ===%s\n", colorGreen, colorReset)
	if status.Megapool.EffectiveDelegateAddress == status.LatestDelegate {
		fmt.Printf("The megapool is using the latest delegate at %s%s%s\n", colorBlue, status.Megapool.DelegateAddress, colorReset)
	} else {
		fmt.Printf("The megapool can be upgraded to delegate %s!%s%s\n", colorBlue, status.LatestDelegate.Hex(), colorReset)
	}
	fmt.Printf("The megapool's effective delegate address is %s%s%s\n", colorBlue, status.Megapool.EffectiveDelegateAddress, colorReset)

	if status.Megapool.UseLatestDelegate {
		fmt.Println("The megapool is set to automatically upgrade to the latest delegate. You can toggle this setting using 'rocketpool megapool set-use-latest-delegate'.")
	} else {
		fmt.Println("The megapool has automatic delegate upgrades disabled. You can toggle this setting using 'rocketpool megapool set-use-latest-delegate'.")
		if status.Megapool.DelegateExpiry > 0 {
			fmt.Printf("Your current megapool delegate expires at %s block %d%s. Afterwards, it can be permissionlessly upgraded by other users.\n", colorBlue, status.Megapool.DelegateExpiry, colorReset)
		}
	}
	fmt.Println("")

	beaconBalance := new(big.Int)
	beaconBalance.Add(status.Megapool.UserCapital, status.Megapool.NodeCapital)

	fmt.Printf("%s=== Megapool Balance ===%s\n", colorGreen, colorReset)
	fmt.Printf("The megapool has %6f node-bonded ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.NodeBond), 6))
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
		fmt.Println("The node does not have a megapool.")
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
	for _, status := range statusName {
		validators, ok := statusValidators[status]
		if !ok || len(validators) == 0 {
			continue
		}

		fmt.Printf("%d %s validator(s):\n", len(validators), status)
		fmt.Println("")

		for _, validator := range validators {
			printValidatorDetails(validator, status)
		}

	}

	return nil

}

func printValidatorDetails(validator api.MegapoolValidatorDetails, status string) {

	fmt.Printf("--------------------\n")
	fmt.Println("")

	if status == "Prelaunch" {
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator active:             yes\n")
		fmt.Printf("Validator index:              \n")
		fmt.Printf("RP ETH assignment time:       %s\n", validator.LastAssignmentTime.Format(TimeFormat))
		fmt.Printf("Node deposit:                 %d ETH\n", validator.LastRequestedBond/1000)
		fmt.Printf("RP deposit:                   %d ETH\n", (validator.LastRequestedValue-validator.LastRequestedBond)/1000)
		fmt.Printf("Validator active:             yes\n")
	}

	if status == "Staking" {
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator active:             yes\n")
		fmt.Printf("Validator index:              \n")
		fmt.Printf("Validator active:             yes\n")
	}

	// Main details
	if validator.ExpressUsed {
		fmt.Printf("Express Ticket Used:          yes\n")
	} else {
		fmt.Printf("Express Ticket Used:          no\n")
	}

	if status == "Initialized" {
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Expected pubkey:              0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator Queue Position:     \n")
		fmt.Printf("RP ETH requested:             %d ETH\n", (validator.LastRequestedValue-validator.LastRequestedBond)/1000)
		fmt.Printf("Node deposit:                 %d ETH\n", validator.LastRequestedBond/1000)
	}
	fmt.Println("")

}
