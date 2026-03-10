package rocketpool

import "github.com/rocket-pool/smartnode/shared/utils/cli/color"

// Print a warning about the gas estimate for operations that have multiple transactions
func (rp *Client) PrintMultiTxWarning() {

	color.YellowPrintln("NOTE: This operation requires multiple transactions.")

}
