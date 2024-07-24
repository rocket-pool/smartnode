package utils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	utilsstrings "github.com/rocket-pool/rocketpool-go/v2/utils/strings"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

// Print a TX's details to the console.
func PrintTransactionHash(rp *client.Client, hash common.Hash) {
	finalMessage := "Waiting for the transaction to be included in a block... you may wait here for it, or press CTRL+C to exit and return to the terminal.\n\n"
	printTransactionHashImpl(rp, hash, finalMessage)
}

// Print a TX's details to the console, but inform the user NOT to cancel it.
func PrintTransactionHashNoCancel(rp *client.Client, hash common.Hash) {
	finalMessage := "Waiting for the transaction to be included in a block... **DO NOT EXIT!** This transaction is one of several that must be completed.\n\n"
	printTransactionHashImpl(rp, hash, finalMessage)
}

// Print a batch of transaction hashes to the console.
func PrintTransactionBatchHashes(rp *client.Client, hashes []common.Hash) {
	finalMessage := "Waiting for the transactions to be included in one or more blocks... you may wait here for them, or press CTRL+C to exit and return to the terminal.\n\n"

	// Print the hashes
	fmt.Println("Transactions have been submitted with the following hashes:")
	hashStrings := make([]string, len(hashes))
	for i, hash := range hashes {
		hashString := hash.String()
		hashStrings[i] = hashString
		fmt.Println(hashString)
	}
	fmt.Println()

	txWatchUrl := getTxWatchUrl(rp)
	if txWatchUrl != "" {
		fmt.Println("You may follow their progress by visiting the following URLs in sequence:")
		for _, hash := range hashStrings {
			fmt.Printf("%s/%s\n", txWatchUrl, hash)
		}
	}
	fmt.Println()

	fmt.Print(finalMessage)
}

// Print a warning to the console if the user set a custom nonce, but this operation involves multiple transactions
func PrintMultiTransactionNonceWarning() {
	fmt.Printf("%sNOTE: You have specified the `nonce` flag to indicate a custom nonce for this transaction.\n"+
		"However, this operation requires multiple transactions.\n"+
		"Rocket Pool will use your custom value as a basis, and increment it for each additional transaction.\n"+
		"If you have multiple pending transactions, this MAY OVERRIDE more than the one that you specified.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
}

// Implementation of PrintTransactionHash and PrintTransactionHashNoCancel
func printTransactionHashImpl(rp *client.Client, hash common.Hash, finalMessage string) {

	_, isNew, err := rp.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: couldn't read config file so the transaction URL will be unavailable (%s).\n", err)
		return
	}

	if isNew {
		fmt.Print("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
		return
	}

	txWatchUrl := getTxWatchUrl(rp)
	hashString := hash.String()
	fmt.Printf("Transaction has been submitted with hash %s.\n", hashString)
	if txWatchUrl != "" {
		fmt.Printf("You may follow its progress by visiting:\n")
		fmt.Printf("%s/%s\n\n", txWatchUrl, hashString)
	}
	fmt.Print(finalMessage)
}

// Get the URL for watching the transaction in a block explorer
func getTxWatchUrl(rp *client.Client) string {
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: couldn't read config file so the transaction URL will be unavailable (%s).\n", err.Error())
		return ""
	}

	if isNew {
		fmt.Print("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
		return ""
	}

	res, err := cfg.GetResources()
	if err != nil {
		fmt.Printf("Warning: couldn't read resources from config file so the transaction URL will be unavailable (%s).\n", err.Error())
		return ""
	}
	return res.TxWatchUrl
}

// Convert a Unix datetime to a string, or `---` if it's zero
func GetDateTimeString(dateTime uint64) string {
	return GetDateTimeStringOfTime(time.Unix(int64(dateTime), 0))
}

// Convert a Unix datetime to a string, or `---` if it's zero
func GetDateTimeStringOfTime(dateTime time.Time) string {
	timeString := dateTime.Format(time.RFC822)
	if dateTime == time.Unix(0, 0) {
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

// Prints what network you're currently on
func PrintNetwork(currentNetwork config.Network, isNew bool) error {
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	switch currentNetwork {
	case config.Network_Mainnet:
		fmt.Printf("Your Smart Node is currently using the %sEthereum Mainnet.%s\n\n", terminal.ColorGreen, terminal.ColorReset)
	case snCfg.Network_Devnet:
		fmt.Printf("Your Smart Node is currently using the %sHolesky Development Network.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	case config.Network_Holesky:
		fmt.Printf("Your Smart Node is currently using the %sHolesky Test Network.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	default:
		fmt.Printf("%sYou are on an unexpected network [%v].%s\n\n", terminal.ColorYellow, currentNetwork, terminal.ColorReset)
	}

	return nil
}

// Parses a string representing either a floating point value or a raw wei amount into a *big.Int
func ParseFloat(c *cli.Context, name string, value string, isFraction bool) (*big.Int, error) {
	var floatValue float64
	if c.Bool(RawFlag.Name) {
		val, err := input.ValidateBigInt(name, value)
		if err != nil {
			return nil, err
		}
		return val, nil
	} else if isFraction {
		val, err := input.ValidateFraction(name, value)
		if err != nil {
			return nil, err
		}
		floatValue = val
	} else {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		floatValue = val
	}

	trueVal := eth.EthToWei(floatValue)
	fmt.Printf("Your value will be multiplied by 10^18 to be used in the contracts, which results in:\n\n\t[%s]\n\n", trueVal.String())
	if !(c.Bool("yes") || Confirm("Please make sure this is what you want and does not have any floating-point errors.\n\nIs this result correct?")) {
		fmt.Printf("Cancelled. Please try again with the '--%s' flag and provide an explicit value instead.\n", RawFlag.Name)
		return nil, nil
	}
	return trueVal, nil
}

func TruncateAndSanitize(str string, size int) string {
	if len(str) > size {
		str = fmt.Sprintf("%s...", str[:size])
	}
	return utilsstrings.Sanitize(str)
}
