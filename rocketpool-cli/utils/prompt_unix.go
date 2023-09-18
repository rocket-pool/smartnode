//go:build !windows
// +build !windows

package utils

import (
	"fmt"
	"regexp"
	"syscall"

	"golang.org/x/term"
)

// Prompt for password input
func PromptPassword(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

	// Print initial prompt
	fmt.Println(initialPrompt)

	// Get valid user input
	var input string
	var init bool
	for !init || !regexp.MustCompile(expectedFormat).MatchString(input) {

		// Incorrect format
		if init {
			fmt.Println("")
			fmt.Println(incorrectFormatPrompt)
		} else {
			init = true
		}

		// Read password
		if bytes, err := term.ReadPassword(syscall.Stdin); err != nil {
			fmt.Println(fmt.Errorf("Could not read password: %w", err))
		} else {
			input = string(bytes)
		}

	}
	fmt.Println("")

	// Return user input
	return input

}
