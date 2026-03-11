package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/urfave/cli/v3"
)

const TimeFormat = "2006-01-02, 15:04 -0700 MST"

func Parent(c *cli.Command) *cli.Command {
	lineage := c.Lineage()
	if len(lineage) < 2 {
		return nil
	}
	return lineage[1]
}

// Print a TX's details to the console.
func PrintTransactionHash(rp *rocketpool.Client, hash common.Hash) {

	finalMessage := "Waiting for the transaction to be included in a block... you may wait here for it, or press CTRL+C to exit and return to the terminal.\n\n"
	printTransactionHashImpl(rp, hash, finalMessage)

}

// Print a TX's details to the console, but inform the user NOT to cancel it.
func PrintTransactionHashNoCancel(rp *rocketpool.Client, hash common.Hash) {

	finalMessage := "Waiting for the transaction to be included in a block... **DO NOT EXIT!** This transaction is one of several that must be completed.\n\n"
	printTransactionHashImpl(rp, hash, finalMessage)

}

// Print a warning to the console if the user set a custom nonce, but this operation involves multiple transactions
func PrintMultiTransactionNonceWarning() {

	color.YellowPrintln("NOTE: You have specified the `nonce` flag to indicate a custom nonce for this transaction.")
	color.YellowPrintln("However, this operation requires multiple transactions.")
	color.YellowPrintln("Rocket Pool will use your custom value as a basis, and increment it for each additional transaction.")
	color.YellowPrintln("If you have multiple pending transactions, this MAY OVERRIDE more than the one that you specified.")
}

// Implementation of PrintTransactionHash and PrintTransactionHashNoCancel
func printTransactionHashImpl(rp *rocketpool.Client, hash common.Hash, finalMessage string) {

	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: couldn't read config file so the transaction URL will be unavailable (%s).\n", err)
		return
	}

	if isNew {
		fmt.Print("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
		return
	}

	txWatchUrl := cfg.Smartnode.GetTxWatchUrl()
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

// Gets the hex string of an address, or "none" if it was the 0x0 address
func GetPrettyAddress(address common.Address) string {
	addressString := address.Hex()
	if addressString == "0x0000000000000000000000000000000000000000" {
		return "<none>"
	}
	return addressString
}

// Temporary table for replacing revert messages with more useful versions until we can refactor
var errorMap = map[string]string{
	"Could not get can node deposit status: Minipool count after deposit exceeds limit based on node RPL stake": "Cannot create a new minipool: you do not have enough RPL staked to create another minipool.",
}

// Prints an error in a prettier format, removing the "stack trace" if it represents
// a contract revert message
func PrettyPrintError(err error) {
	errorMessage := err.Error()
	prettyErr := errorMessage
	if strings.Contains(errorMessage, "execution reverted:") {
		elements := strings.Split(errorMessage, ":")
		firstMessage := strings.TrimSpace(elements[0])
		secondMessage := strings.TrimSpace(elements[len(elements)-1])
		prettyErr = fmt.Sprintf("%s: %s", firstMessage, secondMessage)

		// Look for the message in the above error table and replace if appropriate
		replacementMessage, exists := errorMap[prettyErr]
		if exists {
			prettyErr = replacementMessage
		}
	}
	fmt.Println(prettyErr)
}

// Prints an error message when the Beacon client is not using the deposit contract address that Rocket Pool expects
func PrintDepositMismatchError(rpNetwork, beaconNetwork uint64, rpDepositAddress, beaconDepositAddress common.Address) {
	color.RedPrintln("***ALERT***")
	color.RedPrintln("YOUR ETH2 CLIENT IS NOT CONNECTED TO THE SAME NETWORK THAT ROCKET POOL IS USING!")
	color.RedPrintln("This is likely because your ETH2 client is using the wrong configuration.")
	color.RedPrintln("For the safety of your funds, Rocket Pool will not let you deposit your ETH until this is resolved.")
	color.RedPrintln()
	color.RedPrintln("To fix it if you are in Docker mode:")
	color.RedPrintln("\t1. Run 'rocketpool service install -d' to get the latest configuration")
	color.RedPrintln("\t2. Run 'rocketpool service stop' and 'rocketpool service start' to apply the configuration.")
	color.RedPrintln("If you are using Hybrid or Native mode, please correct the network flags in your ETH2 launch script.")
	color.RedPrintln()
	color.RedPrintln("Details:")
	color.RedPrintf("\tRocket Pool expects deposit contract %s on chain %d.\n", rpDepositAddress.Hex(), rpNetwork)
	color.RedPrintf("\tYour Beacon client is using deposit contract %s on chain %d.\n", beaconDepositAddress.Hex(), beaconNetwork)
}

// Prints what network you're currently on
func PrintNetwork(currentNetwork cfgtypes.Network, isNew bool) error {
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	var networkName string

	switch currentNetwork {
	case cfgtypes.Network_Mainnet:
		networkName = color.Green("Ethereum Mainnet")
	case cfgtypes.Network_Devnet:
		networkName = color.Yellow("Development Network")
	case cfgtypes.Network_Testnet:
		networkName = color.Yellow("Hoodi Test Network")
	default:
		color.YellowPrintf("You are on an unexpected network [%v].\n", currentNetwork)
		fmt.Println()
		return nil
	}

	fmt.Printf("Your Smart Node is currently using the %s.\n", networkName)
	fmt.Println()

	return nil
}
