package cli

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Print a TX's details to the console.
func PrintTransactionHash(hash common.Hash) {

    hashString := hash.String()

    fmt.Printf("Transaction has been submitted with hash %s.\n", hashString)
    fmt.Printf("You may follow its progress by visiting:\n")
    fmt.Printf("https://goerli.etherscan.io/tx/%s\n\n", hashString)
    fmt.Print("Waiting for the transaction to be mined... you may wait here for it, or press CTRL+C to exit and return to the terminal.\n\n")
    
}


// Print a TX's details to the console, but inform the user NOT to cancel it.
func PrintTransactionHashNoCancel(hash common.Hash) {

    hashString := hash.String()

    fmt.Printf("Transaction has been submitted with hash %s.\n", hashString)
    fmt.Printf("You may follow its progress by visiting:\n")
    fmt.Printf("https://goerli.etherscan.io/tx/%s\n\n", hashString)
    fmt.Print("Waiting for the transaction to be mined... **DO NOT EXIT!** This transaction is one of several that must be completed.\n\n")
    
}

