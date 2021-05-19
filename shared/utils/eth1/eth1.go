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

    if c.GlobalBool("override-nonce") {
        ec, err := services.GetEthClient(c)
        if err != nil {
            return fmt.Errorf("Could not retrieve ETH1 client: %w", err)
        }
        // Get the latest nonce
        lastNonce, err := ec.NonceAt(context.Background(), opts.From, nil)
        if err != nil {
            return fmt.Errorf("Could not get latest nonce: %w", err)
        }
        // Set the nonce of this TX to the same as the previous one
        opts.Nonce = new(big.Int).SetUint64(lastNonce)
    }
    return nil

}

