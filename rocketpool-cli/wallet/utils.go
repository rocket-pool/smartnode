package wallet

import (
    "fmt"

    "github.com/tyler-smith/go-bip39"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Prompt for a wallet password
func promptPassword() string {
    for {
        password := cliutils.PromptPassword(
            "Please enter a password to secure your wallet with:",
            fmt.Sprintf("^.{%d,}$", passwords.MinPasswordLength),
            fmt.Sprintf("Your password must be at least %d characters long", passwords.MinPasswordLength),
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
        mnemonic := cliutils.PromptPassword("Please enter your recovery mnemonic phrase:", "^.*$", "")
        if bip39.IsMnemonicValid(mnemonic) {
            return mnemonic
        } else {
            fmt.Println("Invalid mnemonic phrase.")
            fmt.Println("")
        }
    }
}


// Confirm a recovery mnemonic phrase
func confirmMnemonic(mnemonic string) {
    for {
        confirmation := cliutils.Prompt("Please enter your recorded mnemonic phrase to confirm it is correct:", "^.*$", "")
        if mnemonic == confirmation {
            return
        } else {
            fmt.Println("The mnemonic phrase you entered does not match your recovery phrase. Please try again.")
            fmt.Println("")
        }
    }
}

