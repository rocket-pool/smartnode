package pdao

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func setVotingDelegate(c *cli.Context, nameOrAddress string) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the address
	var address common.Address
	var addressString string
	if strings.Contains(nameOrAddress, ".") {
		response, err := rp.Api.Node.ResolveEns(common.Address{}, nameOrAddress)
		if err != nil {
			return err
		}
		address = response.Data.Address
		addressString = fmt.Sprintf("%s (%s)", nameOrAddress, address.Hex())
	} else {
		address, err = input.ValidateAddress("delegate", nameOrAddress)
		if err != nil {
			return err
		}
		addressString = address.Hex()
	}

	// Get the TX
	response, err := rp.Api.PDao.SetVotingDelegate(address)
	if err != nil {
		return err
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want %s to represent your node in Rocket Pool on-chain governance proposals?", addressString),
		"setting voting delegate",
		"Setting voting delegate...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("The node's voting delegate was successfuly set to %s.\n", addressString)
	return nil

}
