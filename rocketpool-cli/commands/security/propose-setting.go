package security

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

// Master general proposal function
func proposeSetting[ValueType utils.SettingType](c *cli.Context, contract rocketpool.ContractName, setting protocol.SettingName, value ValueType) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

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
	response, err := rp.Api.Security.ProposeSetting(contract, setting, valueString)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("You cannot currently submit this proposal:")
		if response.Data.UnknownSetting {
			fmt.Printf("Unknown setting '%s' on contract '%s'.\n", setting, contract)
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to submit this proposal?",
		"setting update",
		"Proposing Protocol DAO setting update...",
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