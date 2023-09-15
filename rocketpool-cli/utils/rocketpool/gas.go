package rocketpool

import (
	"fmt"
)

// Print a warning about the gas estimate for operations that have multiple transactions
func (rp *Client) PrintMultiTxWarning() {

	fmt.Printf("%sNOTE: This operation requires multiple transactions.\n%s",
		colorYellow,
		colorReset)

}
