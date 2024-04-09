package odao

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func proposeSetting[ValueType utils.SettingType](c *cli.Context, contract rocketpool.ContractName, setting oracle.SettingName, value ValueType) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Serialize the value
	var valueString string
	switch trueValue := any(value).(type) {
	case *big.Int:
		valueString = trueValue.String()
	case time.Duration:
		valueString = strconv.FormatUint(uint64(trueValue.Seconds()), 10)
	case bool:
		valueString = strconv.FormatBool(trueValue)
	case uint64:
		valueString = strconv.FormatUint(trueValue, 10)
	default:
		panic("unknown setting type")
	}

	// Build the TX
	response, err := rp.Api.ODao.ProposeSetting(contract, setting, valueString)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if response.Data.UnknownSetting {
			fmt.Printf("Unknown setting '%s' on contract '%s'.\n", setting, contract)
		}
		if response.Data.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to submit this proposal?",
		"setting update",
		"Proposing Oracle DAO setting update...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully proposed setting '%s.%s' to '%v'.\n", contract, setting, value)
	return nil
}
