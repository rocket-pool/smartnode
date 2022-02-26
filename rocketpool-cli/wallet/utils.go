package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/tyler-smith/go-bip39"

	"github.com/rocket-pool/smartnode/shared/services/passwords"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const UTF8_EMPTY_SPACE string = "\xE2\x80\x8B"
const colorYellow string = "\033[33m"
const colorReset string = "\033[0m"

// Prompt for a wallet password
func promptPassword() string {
	for {
		password := cliutils.PromptPassword(
			"Please enter a password to secure your wallet with:",
			fmt.Sprintf("^.{%d,}$", passwords.MinPasswordLength),
			fmt.Sprintf("Your password must be at least %d characters long. Please try again:", passwords.MinPasswordLength),
		)
		confirmation := cliutils.PromptPassword("Please confirm your password:", "^.*$", "")
		if password == confirmation {
			return password
		}

		fmt.Println("Password confirmation does not match.")
		fmt.Println("")
	}
}

// Prompt for a recovery mnemonic phrase
func promptMnemonic() string {
	for {
		mnemonic := cliutils.PromptPassword("Please enter your recovery mnemonic phrase:", "^.*$", "")
		// If the user copy-pasted their mnemonic, don't make them panic by failing on the UTF8 guardians
		mnemonic = strings.ReplaceAll(mnemonic, UTF8_EMPTY_SPACE, "")
		if bip39.IsMnemonicValid(mnemonic) {
			return mnemonic
		}

		fmt.Println("Invalid mnemonic phrase.")
		fmt.Println("")
	}
}

// Print a mnemonic
func printMnemonic(mnemonic string) {
	utf8, _ := os.LookupEnv("LANG")

	fmt.Println(colorYellow + "Your mnemonic phrase to recover your wallet is printed below. It can be used to recover your node account and validator keys if they are lost." + colorReset)
	fmt.Println(colorYellow + "Record this phrase somewhere secure and private. Do not share it with anyone as it will give them control of your node account and validators." + colorReset)
	fmt.Println("==============================================================================================================================================")
	fmt.Println("")
	if strings.Contains(utf8, "UTF-8") {
		fmt.Println(strings.ReplaceAll(mnemonic, " ", " "+UTF8_EMPTY_SPACE))
	} else {
		fmt.Println(mnemonic)
	}
	fmt.Println("")
	fmt.Println("==============================================================================================================================================")
	fmt.Println("")
}

// Confirm a recovery mnemonic phrase
func confirmMnemonic(mnemonic string) {
	for {
		confirmation := cliutils.Prompt("Please enter your recorded mnemonic phrase to confirm it is correct:", "^.*$", "")
		if mnemonic == confirmation {
			return
		}

		if strings.Contains(confirmation, UTF8_EMPTY_SPACE) {
			fmt.Println(colorYellow + "It seems like you copy pasted your mnemonic phrase. Please write it down and enter it by hand." + colorReset)
		} else {
			fmt.Println("The mnemonic phrase you entered does not match your recovery phrase. Please try again.")
		}
		fmt.Println("")
	}
}
