package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func setStakeRplForAllowed(c *cli.Context, addressOrEns string, allowed bool) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	var address common.Address
	var addressString string
	if strings.Contains(addressOrEns, ".") {
		response, err := rp.Api.Node.ResolveEns(common.Address{}, addressOrEns)
		if err != nil {
			return err
		}
		address = response.Data.Address
		addressString = fmt.Sprintf("%s (%s)", addressOrEns, address.Hex())
	} else {
		address, err = input.ValidateAddress("address", addressOrEns)
		if err != nil {
			return err
		}
		addressString = address.Hex()
	}

	// Build the TX
	response, err := rp.Api.Node.SetStakeRplForAllowed(address, allowed)
	if err != nil {
		return err
	}

	// Run the TX
	var confirmMsg string
	var identifierMsg string
	var submitMsg string
	if allowed {
		confirmMsg = fmt.Sprintf("Are you sure you want to allow %s to stake RPL for your node?", addressString)
		identifierMsg = "adding address to RPL stake whitelist"
		submitMsg = "Adding address to RPL stake whitelist..."
	} else {
		confirmMsg = fmt.Sprintf("Are you sure you want to remove %s from your RPL staking whitelist?", addressString)
		identifierMsg = "removing address from RPL stake whitelist"
		submitMsg = "Removing address from RPL stake whitelist..."
	}
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		confirmMsg,
		identifierMsg,
		submitMsg,
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	if allowed {
		fmt.Printf("Successfully added %s to your node's RPL staking whitelist.\n", addressString)
	} else {
		fmt.Printf("Successfully removed %s from your node's RPL staking whitelist.\n", addressString)
	}
	return nil
}
