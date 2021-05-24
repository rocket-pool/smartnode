package eth1

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/urfave/cli"
)

// Sets the nonce of the provided transaction options to the latest nonce if requested
func CheckForNonceOverride(c *cli.Context, opts *bind.TransactOpts) error {

    customNonce := c.GlobalUint64("nonce")
    if customNonce != 0 {
        // Do a sanity check to make sure the provided nonce is for a pending transaction
        // otherwise the user is burning gas for no reason
        ec, err := services.GetEthClient(c)
        if err != nil {
            return fmt.Errorf("Could not retrieve ETH1 client: %w", err)
        }

        // Make sure it's not higher than the next available nonce
        nextNonce, err := ec.PendingNonceAt(context.Background(), opts.From)
        if err != nil {
            return fmt.Errorf("Could not get next available nonce: %w", err)
        }
        if customNonce > nextNonce {
            return fmt.Errorf("Can't use nonce %d because it's greater than the next available nonce (%d).", customNonce, nextNonce)
        }

        // Make sure the nonce hasn't already been mined
        latestMinedNonce, err := ec.NonceAt(context.Background(), opts.From, nil)
        if err != nil {
            return fmt.Errorf("Could not get latest nonce: %w", err)
        }
        if customNonce <= latestMinedNonce {
            return fmt.Errorf("Can't use nonce %d because it has already been mined.", customNonce)
        }

        // It points to a pending transaction, so this is a valid thing to do
        opts.Nonce = new(big.Int).SetUint64(customNonce)
    }
    return nil

}

