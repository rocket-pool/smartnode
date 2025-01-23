package megapool

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/types/api"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

const TimeFormat = "2006-01-02, 15:04 -0700 MST"

func getStatus(c *cli.Context) error {

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

	fmt.Printf("Megapool Address: %s\n", status.Megapool.Address)
	fmt.Printf("Megapool Delegate Address: %s\n", status.Megapool.DelegateAddress)
	fmt.Printf("Megapool Effective Delegate Address: %s\n", status.Megapool.EffectiveDelegateAddress)
	fmt.Printf("Megapool Latest Delegate Address %s\n", status.LatestDelegate)
	fmt.Printf("Megapool Delegate Expiry Block: %d\n", status.Megapool.DelegateExpiry)
	fmt.Printf("Megapool Deployed: %t\n", status.Megapool.Deployed)
	fmt.Printf("Megapool UseLatestDelegate: %t\n", status.Megapool.UseLatestDelegate)
	fmt.Printf("Megapool Refund Value: %.6f ETH. \n", math.RoundDown(eth.WeiToEth(status.Megapool.RefundValue), 6))
	fmt.Printf("Megapool Pending Rewards: %.6f ETH. \n", math.RoundDown(eth.WeiToEth(status.Megapool.PendingRewards), 6))
	fmt.Printf("Megapool Validator Count: %d \n", status.Megapool.ValidatorCount)
	fmt.Printf("Node Express Ticket Count: %d\n", status.Megapool.NodeExpressTicketCount)
	// fmt.Printf("Validator Details %+v\n", status.Megapool.Validators)

	fmt.Println()
	fmt.Println("============================")
	fmt.Println()

	statusValidators := map[string][]api.MegapoolValidatorDetails{
		"active":     {},
		"exited":     {},
		"inQueue":    {},
		"inPrestake": {},
	}

	statusName := []string{"active", "exited", "inPrestake", "inQueue"}

	// Iterate over the validators and append them based on their statuses
	for _, validator := range status.Megapool.Validators {
		if validator.Active {
			statusValidators["active"] = append(statusValidators["active"], validator)
		}
		if validator.Exited {
			statusValidators["exited"] = append(statusValidators["exited"], validator)
		}
		if validator.InQueue {
			statusValidators["inQueue"] = append(statusValidators["inQueue"], validator)
		}
		if validator.InPrestake {
			statusValidators["inPrestake"] = append(statusValidators["inPrestake"], validator)
		}
	}

	// Print validators by status
	for _, status := range statusName {
		validators, ok := statusValidators[status]
		if !ok {
			continue
		}
		fmt.Printf("%d %s validator(s):\n", len(validators), status)
		fmt.Println("")

		for _, validator := range validators {
			printValidatorDetails(validator)
		}

	}

	return nil

}

func printValidatorDetails(validator api.MegapoolValidatorDetails) {

	fmt.Printf("--------------------\n")
	fmt.Printf("\n")

	// Main details
	fmt.Printf("Validator pubkey:             0x%s\n", string(validator.PubKey.String()))
	fmt.Printf("Megapool Validator ID:        %d\n", validator.ValidatorId)
	if validator.Active {
		fmt.Printf("Validator active:             yes\n")
	} else {
		fmt.Printf("Validator active:             no\n")
	}
	if validator.ExpressUsed {
		fmt.Printf("Express Ticket Used:          yes\n")
	} else {
		fmt.Printf("Express Ticket Used:          no\n")
	}
	fmt.Printf("Validator LastRequestedBond:  %d\n", validator.LastRequestedBond)
	fmt.Printf("Validator LastRequestedValue: %d\n", validator.LastRequestedValue)
	fmt.Printf("Validator LastAssignmentTime: %s\n", validator.LastAssignmentTime.Format(TimeFormat))
	fmt.Printf("Validator index: \n")
	fmt.Printf("Validator fee: \n")
	fmt.Printf("Total EL rewards: \n")
	fmt.Printf("Validator Balance (EL): \n")
	fmt.Printf("Beacon balance (CL): \n")
	fmt.Printf("Validator fee: \n")
	fmt.Printf("Your Portion: \n")

	fmt.Printf("\n")

}
