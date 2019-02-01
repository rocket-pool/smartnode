package commands

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
)


// Prompt for user input
func Prompt(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

    // Print initial prompt
    fmt.Println(initialPrompt)

    // Get valid user input
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan(); !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()); scanner.Scan() {
        fmt.Println(incorrectFormatPrompt)
    }

    // Return user input
    return scanner.Text()

}

