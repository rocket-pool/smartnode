//go:build windows
// +build windows

package cli

// Prompt for password input
func PromptPassword(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {
	return Prompt(initialPrompt, expectedFormat, incorrectFormatPrompt)
}
