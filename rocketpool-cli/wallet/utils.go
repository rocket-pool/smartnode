package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/tyler-smith/go-bip39"

	"github.com/rocket-pool/smartnode/shared/services/passwords"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

var UTF8_EMPTY_SPACE string = "\xE2\x80\x8B"

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
	utf8Var, _ := os.LookupEnv("LANG")
	utf8 := strings.Contains(utf8Var, "UTF-8")
	if !utf8 {
		fmt.Println(mnemonic)
		return
	}

	fmt.Println(strings.ReplaceAll(mnemonic, " ", " "+UTF8_EMPTY_SPACE))
}

// Confirm a recovery mnemonic phrase
func confirmMnemonic(mnemonic string) {
	for {
		confirmation := cliutils.Prompt("Please enter your recorded mnemonic phrase to confirm it is correct:", "^.*$", "")
		if mnemonic == confirmation {
			return
		}

		if strings.Contains(confirmation, UTF8_EMPTY_SPACE) {
			fmt.Println("It seems like you copy pasted your mnemonic phrase. Please write it down and enter it by hand.")
		} else {
			fmt.Println("The mnemonic phrase you entered does not match your recovery phrase. Please try again.")
		}
		fmt.Println("")
	}
}
