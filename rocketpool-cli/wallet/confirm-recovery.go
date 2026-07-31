package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	promptcli "github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
)

// Returns false if the user declined, in which case the caller should stop
func confirmRecoveryOperation(yes bool, title string, effects []string) bool {
	color.YellowPrintln(title)
	fmt.Println()
	fmt.Println("This will:")
	for _, effect := range effects {
		fmt.Printf("  - %s\n", effect)
	}
	fmt.Println()
	fmt.Println("On a node with many validators this can take several minutes. The work runs in the")
	fmt.Println("node daemon, not in the CLI, so closing the CLI will not cancel it!")
	fmt.Println("It keeps going in the background and finishes on its own.")
	fmt.Println()

	if promptcli.Declined(yes, "Would you like to continue?") {
		fmt.Println("Cancelled.")
		return false
	}
	fmt.Println()
	return true
}
