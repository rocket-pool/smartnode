package prompt

// Creating this package to avoid import cycles that often happen when using the cli package

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

// Prompt for user input
func Prompt(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

	// Print initial prompt
	fmt.Println(initialPrompt)

	// Get valid user input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan(); !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()); scanner.Scan() {
		fmt.Println("")
		fmt.Println(incorrectFormatPrompt)
	}
	fmt.Println("")

	// Return user input
	return scanner.Text()

}

// Prompt for confirmation
func Confirm(fmtStr string, args ...any) bool {
	initialPrompt := fmt.Sprintf(fmtStr, args...)
	response := Prompt(fmt.Sprintf("%s [y/n]", initialPrompt), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
	return (strings.ToLower(response[:1]) == "y")
}

// Declined returns true if the user did not pre-approve with the `yes` flag
// and declined the interactive confirmation prompt.
func Declined(yes bool, fmtStr string, args ...any) bool {
	return !yes && !Confirm(fmtStr, args...)
}

func confirmColor(colorFunc func(string) string, fmtStr string, args ...any) bool {
	initialPrompt := fmt.Sprintf(fmtStr, args...)
	response := Prompt(fmt.Sprintf("%s [y/n]", colorFunc(initialPrompt)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
	return (strings.ToLower(response[:1]) == "y")
}

func ConfirmRed(fmtStr string, args ...any) bool {
	return confirmColor(color.Red, fmtStr, args...)
}

func ConfirmYellow(fmtStr string, args ...any) bool {
	return confirmColor(color.Yellow, fmtStr, args...)
}

// Prompt for 'I agree' confirmation (used on important questions to avoid a quick 'y' response from the user)
func ConfirmWithIAgree(fmtStr string, args ...any) bool {
	initialPrompt := fmt.Sprintf(fmtStr, args...)
	response := Prompt(fmt.Sprintf("%s [Type 'I agree' or 'n']", initialPrompt), "(?i)^(i agree|n|no)$", "Please answer 'I agree' or 'n'")
	return (len(response) == 7 && strings.ToLower(response[:7]) == "i agree")
}

// Prompt for user selection
func Select(initialPrompt string, options []string) (int, string) {

	// Get prompt
	prompt := initialPrompt
	for i, option := range options {
		prompt += fmt.Sprintf("\n%d: %s", (i + 1), option)
	}

	// Get expected response format
	optionNumbers := []string{}
	for i := range options {
		optionNumbers = append(optionNumbers, strconv.Itoa(i+1))
	}
	expectedFormat := fmt.Sprintf("^(%s)$", strings.Join(optionNumbers, "|"))

	// Prompt user
	response := Prompt(prompt, expectedFormat, "Please enter a number corresponding to an option")

	// Get selected option
	index, _ := strconv.Atoi(response)
	selectedIndex := index - 1
	selectedOption := options[selectedIndex]

	// Return
	return selectedIndex, selectedOption

}

// Prompts the user to verify that there is nobody looking over their shoulder before printing sensitive information.
func ConfirmSecureSession(warning string) bool {
	if !Confirm("%s\nAre you sure you want to continue?", color.Yellow(warning)) {
		fmt.Println("Cancelled.")
		return false
	}

	return true
}
