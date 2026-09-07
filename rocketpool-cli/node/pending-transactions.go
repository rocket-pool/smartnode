package node

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getPendingTransactions() error {
	rp := rocketpool.NewClient()
	defer rp.Close()

	resp, err := rp.NodePendingTransactions()
	if err != nil {
		return err
	}

	color.GreenPrintln("=== Node Pending Transactions ===")
	fmt.Printf("Node Account:            %s\n", color.LightBlue(resp.NodeAddress.Hex()))
	fmt.Printf("Mined Nonce (latest):    %s\n", color.Yellow(fmt.Sprintf("%d", resp.LatestNonce)))
	fmt.Printf("Mempool Nonce (pending): %s\n", color.Yellow(fmt.Sprintf("%d", resp.PendingNonce)))
	fmt.Println()

	if resp.PendingCount == 0 {
		color.GreenPrintln("No pending transactions found in the node's mempool.")
		fmt.Println("Note: If a transaction was recently dropped due to low fees or a node restart, you can still cancel it with:")
		fmt.Printf("  %s\n", color.LightBlue(fmt.Sprintf("rocketpool node cancel-transaction --nonce %d", resp.LatestNonce)))
		return nil
	}

	color.YellowPrintf("Found %d pending transaction(s) blocking the node queue:\n\n", resp.PendingCount)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NONCE\tHASH\tTO\tVALUE (ETH)\tMAX FEE (GWEI)\tPRIO FEE (GWEI)\tSTATUS")
	fmt.Fprintln(w, "-----\t----\t--\t-----------\t--------------\t---------------\t------")

	for _, tx := range resp.PendingTransactions {
		hashStr := "<unknown>"
		if tx.Hash != nil {
			hashStr = tx.Hash.Hex()
			if len(hashStr) > 12 {
				hashStr = hashStr[:6] + "..." + hashStr[len(hashStr)-4:]
			}
		}

		toStr := "<unknown>"
		if tx.To != nil {
			toStr = tx.To.Hex()
			if len(toStr) > 12 {
				toStr = toStr[:6] + "..." + toStr[len(toStr)-4:]
			}
		}

		valEth := "0.0000"
		if tx.Value != nil {
			valEth = fmt.Sprintf("%.4f", math.WeiToEth(tx.Value))
		}

		maxFeeGwei := "-"
		if tx.MaxFeePerGas != nil {
			maxFeeGwei = fmt.Sprintf("%.2f", math.WeiToGwei(tx.MaxFeePerGas))
		}

		prioFeeGwei := "-"
		if tx.MaxPriorityFee != nil {
			prioFeeGwei = fmt.Sprintf("%.2f", math.WeiToGwei(tx.MaxPriorityFee))
		}

		status := "Pending"
		if tx.IsStuck {
			status = "Stuck"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			tx.Nonce,
			hashStr,
			toStr,
			valEth,
			maxFeeGwei,
			prioFeeGwei,
			status,
		)
	}
	_ = w.Flush()

	fmt.Println()
	color.YellowPrintln("Action Suggestions:")
	fmt.Printf("  • Cancel the lowest blocking transaction (nonce %d):\n", resp.PendingTransactions[0].Nonce)
	fmt.Printf("      %s\n", color.LightBlue(fmt.Sprintf("rocketpool node cancel-transaction --nonce %d", resp.PendingTransactions[0].Nonce)))
	if resp.PendingCount > 1 {
		fmt.Printf("  • Cancel ALL %d pending transactions sequentially:\n", resp.PendingCount)
		fmt.Printf("      %s\n", color.LightBlue("rocketpool node cancel-transaction --all"))
	}
	return nil
}
