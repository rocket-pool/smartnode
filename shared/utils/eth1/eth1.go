package eth1

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/urfave/cli"
)

// Sets the nonce of the provided transaction options to the latest nonce if requested
func CheckForNonceOverride(c *cli.Context, opts *bind.TransactOpts) error {

	customNonceString := c.GlobalString("nonce")
	if customNonceString != "" {
		customNonce, success := big.NewInt(0).SetString(customNonceString, 0)
		if !success {
			return fmt.Errorf("Invalid nonce: %s", customNonceString)
		}

		// Do a sanity check to make sure the provided nonce is for a pending transaction
		// otherwise the user is burning gas for no reason
		ec, err := services.GetEthClient(c)
		if err != nil {
			return fmt.Errorf("Could not retrieve ETH1 client: %w", err)
		}

		// Make sure it's not higher than the next available nonce
		nextNonceUint, err := ec.PendingNonceAt(context.Background(), opts.From)
		if err != nil {
			return fmt.Errorf("Could not get next available nonce: %w", err)
		}

		nextNonce := big.NewInt(0).SetUint64(nextNonceUint)
		if customNonce.Cmp(nextNonce) == 1 {
			return fmt.Errorf("Can't use nonce %s because it's greater than the next available nonce (%d).", customNonceString, nextNonceUint)
		}

		// Make sure the nonce hasn't already been included in a block
		latestProposedNonceUint, err := ec.NonceAt(context.Background(), opts.From, nil)
		if err != nil {
			return fmt.Errorf("Could not get latest nonce: %w", err)
		}

		latestProposedNonce := big.NewInt(0).SetUint64(latestProposedNonceUint)
		if customNonce.Cmp(latestProposedNonce) == -1 {
			return fmt.Errorf("Can't use nonce %s because it has already been included in a block.", customNonceString)
		}

		// It points to a pending transaction, so this is a valid thing to do
		opts.Nonce = customNonce
	}
	return nil

}

// Determines if the primary EC can be used for historical queries, or if the Archive EC is required
func GetBestApiClient(primary *rocketpool.RocketPool, cfg *config.RocketPoolConfig, printMessage func(string), blockNumber *big.Int) (*rocketpool.RocketPool, error) {

	client := primary

	// Try getting the rETH address as a canary to see if the block is available
	opts := &bind.CallOpts{
		BlockNumber: blockNumber,
	}
	address, err := client.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.addressrocketTokenRETH")))
	if err != nil {
		errMessage := err.Error()
		printMessage(fmt.Sprintf("Error getting state for block %d: %s", blockNumber.Uint64(), errMessage))
		// The state was missing so fall back to the archive node
		archiveEcUrl := cfg.Smartnode.ArchiveECUrl.Value.(string)
		if archiveEcUrl != "" {
			printMessage(fmt.Sprintf("Primary EC cannot retrieve state for historical block %d, using archive EC [%s]", blockNumber.Uint64(), archiveEcUrl))
			ec, err := services.NewEthClient(archiveEcUrl)
			if err != nil {
				return nil, fmt.Errorf("Error connecting to archive EC: %w", err)
			}
			client, err = rocketpool.NewRocketPool(ec, common.HexToAddress(cfg.Smartnode.GetStorageAddress()))
			if err != nil {
				return nil, fmt.Errorf("Error creating Rocket Pool client connected to archive EC: %w", err)
			}

			// Get the rETH address from the archive EC
			address, err = client.RocketStorage.GetAddress(opts, crypto.Keccak256Hash([]byte("contract.addressrocketTokenRETH")))
			if err != nil {
				return nil, fmt.Errorf("Error verifying rETH address with Archive EC: %w", err)
			}
		} else {
			// No archive node specified
			return nil, fmt.Errorf("***ERROR*** Primary EC cannot retrieve state for historical block %d and the Archive EC is not specified.", blockNumber.Uint64())
		}
	}

	// Sanity check the rETH address to make sure the client is working right
	if address != cfg.Smartnode.GetRethAddress() {
		return nil, fmt.Errorf("***ERROR*** Your Primary EC provided %s as the rETH address, but it should have been %s!", address.Hex(), cfg.Smartnode.GetRethAddress().Hex())
	}

	return client, nil

}
