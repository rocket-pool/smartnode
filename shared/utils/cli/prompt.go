package cli

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "regexp"
)


// Prompt for user input
func Prompt(input *os.File, initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

    // Read from stdin by default
    if input == nil { input = os.Stdin }

    // Get initial offset
    var offset int64 = 0
    if input != os.Stdin { offset, _ = input.Seek(0, io.SeekCurrent) }

    // Print initial prompt
    fmt.Println(initialPrompt)

    // Get valid user input, increment offset
    scanner := bufio.NewScanner(input)
    scanner.Scan()
    offset += int64(len(scanner.Bytes()) + 1)
    for !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()) {
        fmt.Println(incorrectFormatPrompt)
        scanner.Scan()
        offset += int64(len(scanner.Bytes()) + 1)
    }

    // Seek input to offset
    if input != os.Stdin { input.Seek(offset, io.SeekStart) }

    // Return user input
    return scanner.Text()

}

