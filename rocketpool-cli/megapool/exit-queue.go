package megapool

import (
	"fmt"
	"strconv"
	"strings"

	rocketpoolapi "github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

// Exit the megapool queue
func exitQueue(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var selectedValidators []api.MegapoolValidatorDetails

	// Check if the validator id flag is set
	if c.String("validator-id") != "" {
		flagValue := c.String("validator-id")

		// Get Megapool status to resolve validator details
		status, err := rp.MegapoolStatus(false)
		if err != nil {
			return err
		}

		validatorsInQueue := []api.MegapoolValidatorDetails{}
		for _, validator := range status.Megapool.Validators {
			if validator.InQueue {
				validatorsInQueue = append(validatorsInQueue, validator)
			}
		}

		if strings.ToLower(flagValue) == "all" {
			if len(validatorsInQueue) == 0 {
				fmt.Println("No validators can exit the queue at the moment")
				return nil
			}
			selectedValidators = validatorsInQueue
		} else {
			// Parse comma-separated validator IDs
			ids := strings.Split(flagValue, ",")
			for _, idStr := range ids {
				idStr = strings.TrimSpace(idStr)
				validatorId, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					return fmt.Errorf("Invalid validator id '%s': %w", idStr, err)
				}
				found := false
				for _, v := range validatorsInQueue {
					if uint64(v.ValidatorId) == validatorId {
						selectedValidators = append(selectedValidators, v)
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Validator %d is not in the queue", validatorId)
				}
			}
		}
	} else {
		// Get Megapool status
		status, err := rp.MegapoolStatus(false)
		if err != nil {
			return err
		}

		validatorsInQueue := []api.MegapoolValidatorDetails{}
		for _, validator := range status.Megapool.Validators {
			if validator.InQueue {
				validatorsInQueue = append(validatorsInQueue, validator)
			}
		}

		if len(validatorsInQueue) == 0 {
			fmt.Println("No validators can exit the queue at the moment")
			return nil
		}

		// Display validators in queue
		fmt.Println("Validators currently in the queue:")
		for i, v := range validatorsInQueue {
			fmt.Printf("  %d: Validator ID %d | Pubkey: 0x%s | Queue Position: %s\n", i+1, v.ValidatorId, v.PubKey.String(), v.QueuePosition.String())
		}
		fmt.Println()

		// Build regex to accept "all" or comma-separated numbers within valid range
		optionNumbers := []string{}
		for i := range validatorsInQueue {
			optionNumbers = append(optionNumbers, strconv.Itoa(i+1))
		}
		numbersPattern := strings.Join(optionNumbers, "|")
		expectedFormat := fmt.Sprintf("(?i)^(all|(%s)(,(%s))*)$", numbersPattern, numbersPattern)

		// Prompt for selection
		response := prompt.Prompt(
			"Enter the numbers of the validators to exit from the queue (comma-separated) or 'all' to exit all:",
			expectedFormat,
			"Please enter a comma-separated list of validator numbers or 'all'",
		)

		if strings.ToLower(response) == "all" {
			selectedValidators = validatorsInQueue
		} else {
			parts := strings.Split(response, ",")
			seen := make(map[int]bool)
			for _, part := range parts {
				part = strings.TrimSpace(part)
				idx, _ := strconv.Atoi(part)
				if seen[idx] {
					continue
				}
				seen[idx] = true
				selectedValidators = append(selectedValidators, validatorsInQueue[idx-1])
			}
		}
	}

	if len(selectedValidators) == 0 {
		fmt.Println("No validators selected.")
		return nil
	}

	// Check whether each validator can be exited and accumulate gas estimates
	var totalGasInfo rocketpoolapi.GasInfo
	canExitResponses := make(map[uint64]*api.CanExitQueueResponse)
	for _, v := range selectedValidators {
		validatorId := uint64(v.ValidatorId)
		canExit, err := rp.CanExitQueue(uint32(validatorId))
		if err != nil {
			return fmt.Errorf("Error checking if validator %d can be exited: %w", validatorId, err)
		}
		if !canExit.CanExit {
			fmt.Printf("Validator %d cannot be exited from the megapool queue, skipping.\n", validatorId)
			continue
		}
		canExitResponses[validatorId] = &canExit
		totalGasInfo.EstGasLimit += canExit.GasInfo.EstGasLimit
		totalGasInfo.SafeGasLimit += canExit.GasInfo.SafeGasLimit
	}

	if len(canExitResponses) == 0 {
		fmt.Println("No selected validators can be exited from the queue.")
		return nil
	}

	// If a custom nonce is set and there are multiple transactions, warn the user
	if c.GlobalUint64("nonce") != 0 && len(canExitResponses) > 1 {
		cliutils.PrintMultiTransactionNonceWarning()
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(totalGasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Ask for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to exit %d validator(s) from the megapool queue?", len(canExitResponses)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Request exit from the megapool queue for each selected validator
	for _, v := range selectedValidators {
		validatorId := uint64(v.ValidatorId)
		if _, ok := canExitResponses[validatorId]; !ok {
			continue
		}

		response, err := rp.ExitQueue(uint32(validatorId))
		if err != nil {
			fmt.Printf("Could not exit validator %d from the megapool queue: %s.\n", validatorId, err.Error())
			continue
		}

		fmt.Printf("Exiting validator %d from the megapool queue...\n", validatorId)
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not exit validator %d from the megapool queue: %s.\n", validatorId, err.Error())
		} else {
			fmt.Printf("Successfully exited validator ID %d from the megapool queue.\n", validatorId)
		}

		// If a custom nonce is set, increment it for the next transaction
		if c.GlobalUint64("nonce") != 0 {
			rp.IncrementCustomNonce()
		}
	}

	fmt.Println("You have received credit for the validator deposit(s) and may withdraw it using the command `rocketpool node withdraw-credit`.")
	return nil
}
