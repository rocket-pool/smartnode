package megapool

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func notifyFinalBalance(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	var validatorId uint64
	var validatorIndex uint64

	if c.IsSet("validator-id") {
		validatorId = c.Uint64("validator-id")
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus()
		if err != nil {
			return err
		}

		exitingValidators := []api.MegapoolValidatorDetails{}

		for _, validator := range status.Megapool.Validators {
			if validator.Exiting && validator.BeaconStatus.Status == beacon.ValidatorState_WithdrawalDone {
				exitingValidators = append(exitingValidators, validator)
			}
		}
		if len(exitingValidators) > 0 {
			sort.Sort(ByIndex(exitingValidators))
			options := make([]string, len(exitingValidators))
			for vi, v := range exitingValidators {
				options[vi] = fmt.Sprintf("ID: %d - Index: %d - Pubkey: 0x%s", v.ValidatorId, v.ValidatorIndex, v.PubKey.String())
			}
			selected, _ := prompt.Select("Please select a validator to notify the final balance:", options)

			// Get validators
			validatorId = uint64(exitingValidators[selected].ValidatorId)
			validatorIndex = uint64(exitingValidators[selected].ValidatorIndex)

		} else {
			fmt.Println("No validators at the state where the full withdrawal can be proved")
			return nil
		}
	}
	slot := uint64(0)

	if c.IsSet("slot") {
		fmt.Println("Using withdrawal slot: ", c.Uint64("slot"))
	} else {
		fmt.Println("The Smart Node needs to find the slot containing the validator withdrawal. This may take a while. You can speed up the final balance proof generation by submitting the withdrawal slot for your validator.")
		fmt.Println()

		beaconChainUrl := getBeaconChainURL(validatorIndex, cfg)
		fmt.Printf("The withdrawal slot for validator ID: %d can be found under the 'Consensus Layer' tab on this page: %s\n", validatorId, beaconChainUrl)
		fmt.Println()

		if prompt.Confirm("Would you like to manually input the withdrawal slot?") {
			slotString := prompt.Prompt("Please enter the withdrawal slot:", "^\\d+$", "Invalid slot. Please provide a slot number.")
			slot, err = strconv.ParseUint(slotString, 0, 64)
			if err != nil {
				return fmt.Errorf("'%s' is not a valid slot: %w.\n", slotString, err)
			}
		}
	}

	fmt.Printf("%sFetching the beacon state to craft a final balance proof. This process can take several minutes and is CPU and memory intensive.%s", colorYellow, colorReset)
	fmt.Println()

	response, err := rp.CanNotifyFinalBalance(validatorId, slot)
	if err != nil {
		return err
	}

	if !response.CanExit {
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(response.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to notify the final balance for validator id %d exit?", validatorId))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Exit the validator
	resp, err := rp.NotifyFinalBalance(validatorId, slot)
	if err != nil {
		return err
	}

	fmt.Printf("Notifying validator final balance...\n")
	cliutils.PrintTransactionHash(rp, resp.TxHash)
	if _, err = rp.WaitForTransaction(resp.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully notified final balance for validator id %d.\n", validatorId)
	return nil

}

// returns the Beaconcha.in withdrawals URL for a validator index.
func getBeaconChainURL(index uint64, cfg *config.RocketPoolConfig) string {
	network := cfg.GetNetwork()

	var baseURL string
	switch network {
	case cfgtypes.Network_Mainnet:
		baseURL = "https://beaconcha.in"
	case cfgtypes.Network_Devnet, cfgtypes.Network_Testnet:
		baseURL = "https://hoodi.beaconcha.in"
	default:
		return ""
	}

	return fmt.Sprintf("%s/validator/%d#withdrawals", baseURL, index)
}
