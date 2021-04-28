package cli

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

// Print a TX's details to the console.
func PrintTransactionHash(rp *rocketpool.Client, hash common.Hash) {

    finalMessage := "Waiting for the transaction to be mined... you may wait here for it, or press CTRL+C to exit and return to the terminal.\n\n"
    printTransactionHashImpl(rp, hash, finalMessage)
    
}


// Print a TX's details to the console, but inform the user NOT to cancel it.
func PrintTransactionHashNoCancel(rp *rocketpool.Client, hash common.Hash) {

    finalMessage := "Waiting for the transaction to be mined... **DO NOT EXIT!** This transaction is one of several that must be completed.\n\n"
    printTransactionHashImpl(rp, hash, finalMessage)
    
}


// Implementation of PrintTransactionHash and PrintTransactionHashNoCancel
func printTransactionHashImpl(rp *rocketpool.Client, hash common.Hash, finalMessage string) {

    txWatchUrl := ""

    config, err := rp.LoadGlobalConfig()
    if err != nil {
        fmt.Printf("Warning: couldn't read config file so the transaction URL will be unavailable (%s).\n", err)
    } else {
        txWatchUrl = config.Smartnode.TxWatchUrl
    } 

    hashString := hash.String()

    fmt.Printf("Transaction has been submitted with hash %s.\n", hashString)
    if txWatchUrl != "" {
        fmt.Printf("You may follow its progress by visiting:\n")
        fmt.Printf("%s/%s\n\n", txWatchUrl, hashString)
    }
    fmt.Print(finalMessage)
    
}

