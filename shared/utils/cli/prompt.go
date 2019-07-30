package cli

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
)


// Prompt for user input
func Prompt(input *os.File, initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

    // Read from stdin by default
    if input == nil { input = os.Stdin }

    // Print initial prompt
    fmt.Println(initialPrompt)

    // Get valid user input
    scanner := bufio.NewScanner(input)
    for scanner.Scan(); !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()); scanner.Scan() {
        fmt.Println(incorrectFormatPrompt)
    }

    // Return user input
    return scanner.Text()

}

