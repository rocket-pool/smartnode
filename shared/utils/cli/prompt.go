package cli

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)


// Prompt for user input
func Prompt(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

    // Print initial prompt
    fmt.Println(initialPrompt)

    // Get valid user input, increment offset
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan(); !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()); scanner.Scan() {
        fmt.Println(incorrectFormatPrompt)
    }

    // Return user input
    return scanner.Text()

}


// Prompt for user selection
func Select(initialPrompt string, options []string) string {

    // Get prompt
    prompt := initialPrompt
    for i, option := range options {
        prompt += fmt.Sprintf("\n%d: %s", (i + 1), option)
    }

    // Get expected response format
    optionNumbers := []string{}
    for i, _ := range options {
        optionNumbers = append(optionNumbers, strconv.Itoa(i + 1))
    }
    expectedFormat := fmt.Sprintf("^(%s)$", strings.Join(optionNumbers, "|"))

    // Prompt user
    response := Prompt(prompt, expectedFormat, "Please enter a number corresponding to an option")

    // Get selected option
    index, _ := strconv.Atoi(response)
    selected := options[index - 1]

    // Return
    return selected

}

