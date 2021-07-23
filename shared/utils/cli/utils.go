package cli

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

const colorReset string = "\033[0m"
const colorYellow string = "\033[33m"


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


// Print a warning to the console if the user set a custom nonce, but this operation involves multiple transactions
func PrintMultiTransactionNonceWarning() {

    fmt.Printf("%sNOTE: You have specified the `nonce` flag to indicate a custom nonce for this transaction.\n" +
        "However, this operation requires multiple transactions.\n" +
        "Rocket Pool will use your custom value as a basis, and increment it for each additional transaction.\n" +
        "If you have multiple pending transactions, this MAY OVERRIDE more than the one that you specified.%s\n\n", colorYellow, colorReset)

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


// Convert a Unix datetime to a string, or `---` if it's zero
func GetDateTimeString(dateTime uint64) string {
    timeString := time.Unix(int64(dateTime), 0).Format(time.RFC822)
    if dateTime == 0 {
        timeString = "---"
    }
    return timeString
}
