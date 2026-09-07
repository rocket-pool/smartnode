package node

import (
	"errors"
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func cancelTransaction(nonce uint64, nonceSet bool, cancelAll bool, yes bool) error {
	rp := rocketpool.NewClient()
	defer rp.Close()

	if cancelAll {
		return cancelAllTransactions(rp, yes)
	}

	// If nonce was not explicitly specified via flag, detect lowest pending nonce
	if !nonceSet {
		pendingResp, err := rp.NodePendingTransactions()
		if err != nil {
			return fmt.Errorf("error detecting pending transactions: %w", err)
		}
		if pendingResp.PendingCount == 0 {
			nonce = pendingResp.LatestNonce
			color.YellowPrintf("No pending transactions reported by local mempool; defaulting to current on-chain nonce %d.\n", nonce)
		} else {
			nonce = pendingResp.PendingTransactions[0].Nonce
		}
	}

	return cancelSingleTransaction(rp, nonce, yes)
}

func cancelSingleTransaction(rp *rocketpool.Client, nonce uint64, yes bool) error {
	canCancel, err := rp.CanCancelNodeTransaction(nonce)
	if err != nil {
		return err
	}
	if !canCancel.CanCancel {
		return errors.New("cannot cancel transaction at this nonce")
	}

	color.YellowPrintf("Preparing to cancel transaction at nonce %d with a 0-ETH replacement.\n", nonce)
	fmt.Printf("  Gas Limit:          %d\n", canCancel.GasLimit)
	fmt.Printf("  Max Priority Fee:   %.2f gwei\n", canCancel.MinPriorityFeeGwei)
	fmt.Printf("  Suggested Max Fee:  %.2f gwei\n", canCancel.SuggestedMaxFeeGwei)
	maxCostEth := (canCancel.SuggestedMaxFeeGwei / 1e9) * float64(canCancel.GasLimit)
	fmt.Printf("  Estimated Max Cost: %.6f ETH\n", maxCostEth)

	if !yes {
		if !prompt.Confirm("Are you sure you want to cancel the pending transaction at nonce %d?", nonce) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	resp, err := rp.CancelNodeTransaction(nonce)
	if err != nil {
		return fmt.Errorf("error submitting cancel transaction: %w", err)
	}

	color.GreenPrintf("Cancellation transaction submitted with hash %s\n", resp.TxHash.Hex())
	fmt.Println("Waiting for the transaction to be included in a block...")

	if _, err := rp.WaitForTransaction(resp.TxHash); err != nil {
		return fmt.Errorf("error waiting for cancellation transaction: %w", err)
	}

	color.GreenPrintf("Transaction successfully mined! Nonce %d is now unblocked.\n", nonce)
	return nil
}

func cancelAllTransactions(rp *rocketpool.Client, yes bool) error {
	pendingResp, err := rp.NodePendingTransactions()
	if err != nil {
		return fmt.Errorf("error querying pending transactions: %w", err)
	}

	if pendingResp.PendingCount == 0 {
		color.GreenPrintln("No pending transactions to cancel in the local mempool.")
		return nil
	}

	color.YellowPrintf("Found %d pending transaction(s) to cancel sequentially.\n", pendingResp.PendingCount)

	if !yes {
		if !prompt.Confirm("Are you sure you want to cancel all %d pending transactions?", pendingResp.PendingCount) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	for _, tx := range pendingResp.PendingTransactions {
		color.YellowPrintf("\n--- Cancelling transaction at nonce %d ---\n", tx.Nonce)
		if err := cancelSingleTransaction(rp, tx.Nonce, true); err != nil {
			return fmt.Errorf("error cancelling nonce %d: %w", tx.Nonce, err)
		}
	}

	color.GreenPrintln("\nAll pending transactions have been cancelled successfully.")
	return nil
}
