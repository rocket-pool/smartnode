package eth1

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
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

		// Make sure the nonce hasn't already been mined
		latestMinedNonceUint, err := ec.NonceAt(context.Background(), opts.From, nil)
		if err != nil {
			return fmt.Errorf("Could not get latest nonce: %w", err)
		}

		latestMinedNonce := big.NewInt(0).SetUint64(latestMinedNonceUint)
		if customNonce.Cmp(latestMinedNonce) == -1 {
			return fmt.Errorf("Can't use nonce %s because it has already been mined.", customNonceString)
		}

		// It points to a pending transaction, so this is a valid thing to do
		opts.Nonce = customNonce
	}
	return nil

}

// Check the status of the execution client(s) and provision the API with them
func CheckExecutionClientStatus(rp *rocketpool.Client) error {

	// Check if the primary EC is up, synced, and able to respond to requests - if not, forces the use of the fallback EC for this command
	response, err := rp.GetExecutionClientStatus()
	if err != nil {
		return err
	}

	mgrStatus := response.ManagerStatus

	// Primary EC is good
	if mgrStatus.PrimaryEcStatus.IsSynced {
		rp.SetEcStatusFlags(true, false)
	}

	// Fallback EC is good
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackEcStatus.IsSynced {
		if mgrStatus.PrimaryEcStatus.Error != "" {
			fmt.Printf("Primary execution client is unavailable (%s), using fallback execution client...\n", mgrStatus.PrimaryEcStatus.Error)
		} else {
			fmt.Printf("Primary execution client is still syncing (%.2f%%), using fallback execution client...\n", mgrStatus.PrimaryEcStatus.SyncProgress*100)
		}
		rp.SetEcStatusFlags(true, true)
		return nil
	}

	// Is the primary working and syncing?
	if mgrStatus.PrimaryEcStatus.IsWorking && mgrStatus.PrimaryEcStatus.Error == "" {
		return fmt.Errorf("fallback execution client is not configured or unavailable, and primary execution client is still syncing (%.2f%%)", mgrStatus.PrimaryEcStatus.SyncProgress)
	}

	// Is the fallback working and syncing?
	if mgrStatus.FallbackEnabled && mgrStatus.FallbackEcStatus.IsWorking && mgrStatus.FallbackEcStatus.Error == "" {
		return fmt.Errorf("primary execution client is unavailable (%s), and fallback execution client is still syncing (%.2f%%)", mgrStatus.PrimaryEcStatus.Error, mgrStatus.FallbackEcStatus.SyncProgress)
	}

	// Report if neither client is working
	if mgrStatus.FallbackEnabled {
		return fmt.Errorf("primary execution client is unavailable (%s) and fallback execution client is unavailable (%s), no execution clients are ready", mgrStatus.PrimaryEcStatus.Error, mgrStatus.FallbackEcStatus.Error)
	} else {
		return fmt.Errorf("primary execution client is unavailable (%s) and no fallback execution client is configured", mgrStatus.PrimaryEcStatus.Error)
	}

}
