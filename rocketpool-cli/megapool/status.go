package megapool

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
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

	// Get the beacon balance and node share of rewards waiting to be skimmed
	_, totalBeaconBalance, nodeShareOfCLBalance, err := getValidatorMapAndRewards(rp, status)
	if err != nil {
		return err
	}

	// Return if megapool isn't deployed
	if !status.Megapool.Deployed {
		fmt.Println("The node does not have a megapool. Please run 'rocketpool megapool deploy' and try again.")
		return nil
	}

	// Address, express tickets, validator count
	fmt.Printf("%s=== Megapool ===%s\n", colorGreen, colorReset)
	fmt.Printf("The node has a megapool deployed at %s%s%s\n", colorBlue, status.Megapool.Address, colorReset)
	if status.Megapool.DelegateExpired {
		fmt.Printf("%sThe megapool delegate is expired.%s\n", colorRed, colorReset)
		fmt.Println("Upgrade your megapool delegate using 'rocketpool megapool delegate-upgrade' to view the express ticket and validator count.")
	} else {
		fmt.Printf("The node has %d express ticket(s).\n", status.Megapool.NodeExpressTicketCount)
		fmt.Printf("The megapool has %d validators.\n", status.Megapool.ActiveValidatorCount)
	}
	fmt.Println()

	// Delegate addresses
	fmt.Printf("%s=== Megapool Delegate ===%s\n", colorGreen, colorReset)
	if status.Megapool.DelegateExpired {
		fmt.Printf("%sThe megapool delegate is expired.%s\n", colorRed, colorReset)
		fmt.Printf("The megapool is using an expired delegate at %s%s%s\n", colorBlue, status.Megapool.DelegateAddress, colorReset)
		fmt.Printf("The megapool can be upgraded to delegate %s%s%s using 'rocketpool megapool delegate-upgrade'.\n", colorBlue, status.LatestDelegate, colorReset)
	} else {
		if status.Megapool.EffectiveDelegateAddress == status.LatestDelegate {
			fmt.Println("The megapool is using the latest delegate.")
		} else {
			fmt.Printf("The megapool is using an outdated delegate at %s%s%s\n", colorBlue, status.Megapool.DelegateAddress, colorReset)
			fmt.Printf("The megapool can be upgraded to delegate %s%s%s using 'rocketpool megapool delegate-upgrade'.\n", colorBlue, status.LatestDelegate, colorReset)
		}
		fmt.Printf("The megapool's effective delegate address is %s%s%s\n", colorBlue, status.Megapool.EffectiveDelegateAddress, colorReset)
	}

	if status.Megapool.UseLatestDelegate {
		fmt.Println("The megapool is set to automatically upgrade to the latest delegate. You can toggle this setting using 'rocketpool megapool set-use-latest-delegate'.")
	} else {
		fmt.Println("The megapool has automatic delegate upgrades disabled. You can toggle this setting using 'rocketpool megapool set-use-latest-delegate'.")
		if status.Megapool.DelegateExpiry > 0 {
			fmt.Printf("Your current megapool delegate expires at %sblock %d%s.\n", colorBlue, status.Megapool.DelegateExpiry, colorReset)
		}
	}
	fmt.Println()

	// Balance and network commission
	fmt.Printf("%s=== Megapool Balance ===%s\n", colorGreen, colorReset)
	if !status.Megapool.DelegateExpired {
		totalBond := new(big.Int).Mul(status.Megapool.NodeBond, big.NewInt(8))
		rpBond := new(big.Int).Sub(totalBond, status.Megapool.NodeBond)
		fmt.Printf("The megapool has %6f node bonded ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.NodeBond), 6))
		fmt.Printf("The megapool has %6f RP ETH for a total of %6f bonded ETH.\n", math.RoundDown(eth.WeiToEth(rpBond), 6), math.RoundDown(eth.WeiToEth(totalBond), 6))
		fmt.Printf("Megapool balance (EL): %6f ETH\n", math.RoundDown(eth.WeiToEth(status.Megapool.Balances.ETH), 6))
		if status.Megapool.NodeDebt.Cmp(big.NewInt(0)) > 0 {
			fmt.Printf("The megapool debt is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.NodeDebt), 6))
		}
		if status.Megapool.RefundValue.Cmp(big.NewInt(0)) > 0 {
			fmt.Printf("The megapool refund value is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.Megapool.RefundValue), 6))
		}
		if status.Megapool.PendingRewards.Cmp(big.NewInt(0)) > 0 {
			fmt.Printf("The megapool has %.6f ETH in pending rewards to claim.\n", math.RoundDown(eth.WeiToEth(status.Megapool.PendingRewardSplit.NodeRewards), 6))
		} else {
			fmt.Println("The megapool does not have any pending rewards to claim.")
		}
		fmt.Printf("Beacon balance (CL): %6f ETH\n", math.RoundDown(eth.WeiToEth(totalBeaconBalance), 6))
		fmt.Printf("Your portion: %6f ETH\n", math.RoundDown(eth.WeiToEth(nodeShareOfCLBalance), 6))

		networkCommission := math.RoundDown(eth.WeiToEth(status.Megapool.NodeShare)*100, 6)
		effectiveNodeShare := math.RoundDown(eth.WeiToEth(status.Megapool.RevenueSplit.NodeShare)*100, 6)

		fmt.Printf("Current network commission: %.6f%%\n", networkCommission)
		if networkCommission != effectiveNodeShare {
			fmt.Printf("Effective node share: %.6f%% (time-weighted average due to universal commission changes since last distribution).\n", effectiveNodeShare)
		}
	} else {
		fmt.Print("Upgrade your megapool delegate using 'rocketpool megapool delegate-upgrade' to view the balance and commission details.\n")
	}

	return nil

}

// Get the map of validator states, the total beacon balance, and the node share of beacon balance
func getValidatorMapAndRewards(rp *rocketpool.Client, status api.MegapoolStatusResponse) (map[string][]api.MegapoolValidatorDetails, *big.Int, *big.Int, error) {

	statusValidators := map[string][]api.MegapoolValidatorDetails{
		"Staking":     {},
		"Exited":      {},
		"Initialized": {},
		"Prelaunch":   {},
		"Dissolved":   {},
		"Exiting":     {},
		"Locked":      {},
	}

	var totalBeaconBalance uint64
	var totalEffectiveBeaconBalance uint64
	// Iterate over the validators and append them based on their statuses
	for _, validator := range status.Megapool.Validators {
		if validator.Staked && !validator.Exited && !validator.Exiting {
			statusValidators["Staking"] = append(statusValidators["Staking"], validator)
			if validator.Activated {
				totalBeaconBalance += validator.BeaconStatus.Balance
				totalEffectiveBeaconBalance += validator.BeaconStatus.EffectiveBalance
			}
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
		if validator.Dissolved {
			statusValidators["Dissolved"] = append(statusValidators["Dissolved"], validator)
		}
		if validator.Exiting {
			statusValidators["Exiting"] = append(statusValidators["Exiting"], validator)
		}
		if validator.Locked {
			statusValidators["Locked"] = append(statusValidators["Locked"], validator)
		}
	}

	weiPerGwei := big.NewInt(int64(eth.WeiPerGwei))
	totalBeaconBalanceWei := new(big.Int).SetUint64(totalBeaconBalance)
	totalEffectiveBeaconBalanceWei := new(big.Int).SetUint64(totalEffectiveBeaconBalance)
	totalBeaconBalanceWei = totalBeaconBalanceWei.Mul(totalBeaconBalanceWei, weiPerGwei)
	totalEffectiveBeaconBalanceWei = totalEffectiveBeaconBalanceWei.Mul(totalEffectiveBeaconBalanceWei, weiPerGwei)

	// Get the node share of CL rewards
	nodeShareOfCLBalance := big.NewInt(0)
	if totalBeaconBalanceWei.Cmp(totalEffectiveBeaconBalanceWei) <= 0 {
		nodeShareOfCLBalance = big.NewInt(0)
	} else {
		toBeSkimmed := new(big.Int).Sub(totalBeaconBalanceWei, totalEffectiveBeaconBalanceWei)
		rewards, err := rp.CalculateRewards(toBeSkimmed)
		if err != nil {
			return statusValidators, totalBeaconBalanceWei, nodeShareOfCLBalance, fmt.Errorf("Error calculating the rewards split for amount %s: %w", toBeSkimmed.String(), err)
		}
		nodeShareOfCLBalance = nodeShareOfCLBalance.Add(rewards.RewardSplit.NodeRewards, status.Megapool.NodeBond)
	}

	return statusValidators, totalBeaconBalanceWei, nodeShareOfCLBalance, nil

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
		fmt.Println("The node does not have a megapool. Please run 'rocketpool megapool deploy' and try again.")
		return nil
	}

	// Return if delegate is expired
	if status.Megapool.DelegateExpired {
		fmt.Printf("%sThe megapool delegate is expired.%s\n", colorRed, colorReset)
		fmt.Println("Upgrade your megapool delegate using 'rocketpool megapool delegate-upgrade' to view the validator info.")
		return nil
	}

	// Get the number queue size and express queue rate
	queueDetails, err := rp.GetQueueDetails()
	if err != nil {
		return err
	}

	statusValidators, _, _, err := getValidatorMapAndRewards(rp, status)
	if err != nil {
		return fmt.Errorf("Error while creating validator map %w:", err)
	}

	fmt.Printf("There are %d validator(s) on the express queue.\n", queueDetails.ExpressLength)
	fmt.Printf("There are %d validator(s) on the standard queue.\n", queueDetails.StandardLength)
	fmt.Printf("The express queue rate is %d.\n\n", queueDetails.ExpressRate)

	statusName := []string{"Staking", "Exited", "Prelaunch", "Initialized", "Dissolved", "Exiting", "Locked"}

	// Print validators by status
	noValidators := true
	for _, status := range statusName {
		if validators := statusValidators[status]; len(validators) > 0 {
			noValidators = false
			sort.Slice(validators, func(i, j int) bool {
				return validators[i].ValidatorId < validators[j].ValidatorId
			})
			fmt.Printf("%d %s validator(s):\n", len(validators), status)
			fmt.Println()
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
	fmt.Println()

	if status == "Prelaunch" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Beacon status:                %s\n", validator.BeaconStatus.Status)
	}

	if status == "Staking" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		if validator.Activated {
			fmt.Printf("Validator active:             yes\n")
		} else {
			fmt.Printf("Validator active:             no\n")
		}
		fmt.Printf("Validator index:              %s\n", validator.BeaconStatus.Index)
		fmt.Printf("Beacon status:                %s\n", validator.BeaconStatus.Status)

	}

	if status == "Initialized" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Expected pubkey:              0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Validator Queue Position:     %d\n", validator.QueuePosition)

	}

	if status == "Dissolved" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")

	}

	if status == "Exited" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Validator index:              %s\n", validator.BeaconStatus.Index)
		fmt.Printf("Beacon status:                %s\n", validator.BeaconStatus.Status)
	}

	if status == "Exiting" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Validator index:              %s\n", validator.BeaconStatus.Index)
		fmt.Printf("Beacon status:                %s\n", validator.BeaconStatus.Status)
	}

	if status == "Locked" {
		fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
		fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
		fmt.Printf("Validator active:             no\n")
		fmt.Printf("Validator index:              %s\n", validator.BeaconStatus.Index)
		fmt.Printf("Beacon status:                %s\n", validator.BeaconStatus.Status)
	}

	// Main details
	if validator.ExpressUsed {
		fmt.Printf("Express Ticket Used:          yes\n")
	} else {
		fmt.Printf("Express Ticket Used:          no\n")
	}

	fmt.Println()

}
