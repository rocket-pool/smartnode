package node

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// getPendingTransactions coordinates fetching and displaying pending transactions.
func getPendingTransactions() error {
	rp := rocketpool.NewClient()
	defer rp.Close()

	resp, err := rp.NodePendingTransactions()
	if err != nil {
		return err
	}

	renderPendingTransactionsHeader(&resp)

	if resp.PendingCount == 0 {
		renderNoPendingNotice(resp.LatestNonce)
		return nil
	}

	color.YellowPrintf("Found %d pending transaction(s) blocking the node queue:\n\n", resp.PendingCount)
	renderPendingTransactionsTable(os.Stdout, resp.PendingTransactions)

	fmt.Println()
	renderActionSuggestions(resp.PendingTransactions[0].Nonce, resp.PendingCount)
	return nil
}

// renderPendingTransactionsHeader outputs the summary status block for the node account.
func renderPendingTransactionsHeader(resp *api.NodePendingTransactionsResponse) {
	color.GreenPrintln("=== Node Pending Transactions ===")
	fmt.Printf("Node Account:            %s\n", color.LightBlue(resp.NodeAddress.Hex()))
	fmt.Printf("Mined Nonce (latest):    %s\n", color.Yellow(fmt.Sprintf("%d", resp.LatestNonce)))
	fmt.Printf("Mempool Nonce (pending): %s\n", color.Yellow(fmt.Sprintf("%d", resp.PendingNonce)))
	fmt.Println()
}

// renderNoPendingNotice displays the empty queue status and helpful troubleshooting tip.
func renderNoPendingNotice(latestNonce uint64) {
	color.GreenPrintln("No pending transactions found in the node's mempool.")
	fmt.Println("Note: If a transaction was recently dropped due to low fees or a node restart, you can still cancel it with:")
	fmt.Printf("  %s\n", color.LightBlue(fmt.Sprintf("rocketpool node cancel-transaction --nonce %d", latestNonce)))
}

// renderPendingTransactionsTable writes a formatted table of pending transactions to the provided writer.
func renderPendingTransactionsTable(out io.Writer, txs []api.PendingTxDetails) {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NONCE\tHASH\tTO\tVALUE (ETH)\tMAX FEE (GWEI)\tPRIO FEE (GWEI)\tSTATUS")
	fmt.Fprintln(w, "-----\t----\t--\t-----------\t--------------\t---------------\t------")

	for _, tx := range txs {
		nonce, hashStr, toStr, valEth, maxFeeGwei, prioFeeGwei, status := formatPendingTxRow(tx)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			nonce, hashStr, toStr, valEth, maxFeeGwei, prioFeeGwei, status,
		)
	}
	_ = w.Flush()
}

// formatPendingTxRow formats a single pending transaction into table column strings.
func formatPendingTxRow(tx api.PendingTxDetails) (nonce uint64, hashStr, toStr, valEth, maxFeeGwei, prioFeeGwei, status string) {
	nonce = tx.Nonce

	hashStr = "<unknown>"
	if tx.Hash != nil {
		hashStr = truncateHex(tx.Hash.Hex(), 6, 4)
	}

	toStr = "<unknown>"
	if tx.To != nil {
		toStr = truncateHex(tx.To.Hex(), 6, 4)
	}

	valEth = "0.0000"
	if tx.Value != nil {
		valEth = fmt.Sprintf("%.4f", math.WeiToEth(tx.Value))
	}

	maxFeeGwei = "-"
	if tx.MaxFeePerGas != nil {
		maxFeeGwei = fmt.Sprintf("%.2f", math.WeiToGwei(tx.MaxFeePerGas))
	}

	prioFeeGwei = "-"
	if tx.MaxPriorityFee != nil {
		prioFeeGwei = fmt.Sprintf("%.2f", math.WeiToGwei(tx.MaxPriorityFee))
	}

	status = "Pending"
	if tx.IsStuck {
		status = "Stuck"
	}

	return nonce, hashStr, toStr, valEth, maxFeeGwei, prioFeeGwei, status
}

// truncateHex abbreviates a long hex string with ellipses (e.g. 0x1234...abcd).
func truncateHex(s string, headLen, tailLen int) string {
	if len(s) > headLen+tailLen+2 {
		return s[:headLen] + "..." + s[len(s)-tailLen:]
	}
	return s
}

// renderActionSuggestions displays suggested CLI commands based on pending transaction count.
func renderActionSuggestions(lowestNonce uint64, pendingCount uint64) {
	color.YellowPrintln("Action Suggestions:")
	fmt.Printf("  • Cancel the lowest blocking transaction (nonce %d):\n", lowestNonce)
	fmt.Printf("      %s\n", color.LightBlue(fmt.Sprintf("rocketpool node cancel-transaction --nonce %d", lowestNonce)))
	if pendingCount > 1 {
		fmt.Printf("  • Cancel ALL %d pending transactions sequentially:\n", pendingCount)
		fmt.Printf("      %s\n", color.LightBlue("rocketpool node cancel-transaction --all"))
	}
}
