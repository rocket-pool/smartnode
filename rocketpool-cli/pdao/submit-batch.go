package pdao

import (
	"fmt"
	"math/big"
	"strings"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func printSubmitBatchHelp() {
	fmt.Print(`submit-batch creates one Protocol DAO proposal that changes multiple settings.

Build a JSON file with --to-json on the usual setting propose commands (creates the file or appends; the same contract+setting is replaced):

  rocketpool pdao propose setting auction is-create-lot-enabled true --to-json settings.json
  rocketpool pdao propose setting deposit minimum-deposit 1.0 --to-json settings.json

Submit that file as a single proposal:

  rocketpool pdao propose submit-batch --file settings.json

The file is a JSON array. Each object is one setting:

[
  {
    "contract": "rocketDAOProtocolSettingsAuction",
    "setting": "auction.lot.create.enabled",
    "type": "bool",
    "value": "true"
  },
  {
    "contract": "rocketDAOProtocolSettingsDeposit",
    "setting": "deposit.minimum",
    "type": "uint256",
    "value": "1000000000000000000"
  }
]
`)
}

func submitBatch(file string, message string, yes bool) error {
	if file == "" {
		printSubmitBatchHelp()
		file = prompt.Prompt("Please enter the path to the JSON file:", "^.+$", "Invalid file path")
	}

	settings, err := readBatchSettingsFile(file)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return fmt.Errorf("JSON file %s does not contain any settings", file)
	}

	fmt.Println("The following settings will be submitted as a single proposal:")
	for i, setting := range settings {
		fmt.Printf("  %d. %s / %s = %s\n", i+1, setting.Contract, setting.Setting, setting.Value)
	}
	fmt.Println()

	if message == "" {
		message = prompt.Prompt("Please enter a custom message for this multi-setting proposal (no blank spaces):", "^\\S*$", "Invalid message")
	}
	if message == "" {
		paths := make([]string, len(settings))
		for i, setting := range settings {
			paths[i] = setting.Setting
		}
		message = "set-" + strings.Join(paths, ",")
	}

	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	canPropose, err := rp.PDAOCanProposeSettingMulti(settings, message)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.InsufficientRpl {
			fmt.Printf("You do not have enough RPL staked but unlocked to make another proposal (unlocked: %.6f RPL, required: %.6f RPL).\n",
				math.WeiToEth(big.NewInt(0).Sub(canPropose.StakedRpl, canPropose.LockedRpl)), math.WeiToEth(canPropose.ProposalBond),
			)
		}
		if canPropose.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	err = gas.AssignMaxFeeAndLimit(canPropose.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	if prompt.Declined(yes, "Are you sure you want to submit this proposal with %d setting(s)?", len(settings)) {
		fmt.Println("Cancelled.")
		return nil
	}

	response, err := rp.PDAOProposeSettingMulti(settings, message, canPropose.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting multi-setting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("Successfully submitted a multi-setting proposal with %d setting(s).\n", len(settings))
	return nil
}
