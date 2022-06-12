package wallet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet/bip39"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const bold string = "\033[1m"
const unbold string = "\033[0m"

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
		} else {
			fmt.Println("Password confirmation does not match.")
			fmt.Println("")
		}
	}
}

// Prompt for a recovery mnemonic phrase
func promptMnemonic() string {
	for {
		lengthInput := cliutils.Prompt(
			"Please enter the "+bold+"number"+unbold+" of words in your mnemonic phrase (24 by default):",
			"^[1-9][0-9]*$",
			"Please enter a valid number.")

		length, err := strconv.Atoi(lengthInput)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		mv := bip39.Create(length)
		if mv == nil {
			fmt.Println("Please enter a valid mnemonic length.")
			continue
		}

		i := 0
		for mv.Filled() == false {
			prompt := fmt.Sprintf("Enter %sWord Number %d%s of your mnemonic:", bold, i+1, unbold)
			word := cliutils.PromptPassword(prompt, "^[a-zA-Z]+$", "Please enter a single word only.")

			if err := mv.AddWord(strings.ToLower(word)); err != nil {
				fmt.Println("Inputted word not valid, please retry.")
				continue
			}

			i++
		}

		mnemonic, err := mv.Finalize()
		if err != nil {
			fmt.Printf("Error validating mnemonic: %s\n", err)
			fmt.Println("Please try again.")
			fmt.Println("")
			continue
		}

		return mnemonic
	}
}

// Confirm a recovery mnemonic phrase
func confirmMnemonic(mnemonic string) {
	for {
		fmt.Println("Please enter your mnemonic phrase to confirm.")
		confirmation := promptMnemonic()
		if mnemonic == confirmation {
			return
		} else {
			fmt.Println("The mnemonic phrase you entered does not match your recovery phrase. Please try again.")
			fmt.Println("")
		}
	}
}
