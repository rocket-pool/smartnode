package rocketpool

import (
	"fmt"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorYellow string = "\033[33m"
)

// Print a warning about the gas estimate for operations that have multiple transactions
func PrintMultiTxWarning() {

	fmt.Printf("%sNOTE: This operation requires multiple transactions.\n%s",
		colorYellow,
		colorReset)

}
