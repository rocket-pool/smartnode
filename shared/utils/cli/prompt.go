package cli

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "regexp"
)


// Prompt for user input
func Prompt(input *os.File, output *os.File, initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {

    // Read from stdin, write to stdout by default
    if input == nil { input = os.Stdin }
    if output == nil { output = os.Stdout }

    // Get initial offset
    var offset int64 = 0
    if input != os.Stdin { offset, _ = input.Seek(0, io.SeekCurrent) }

    // Print initial prompt
    fmt.Fprintln(output, initialPrompt)

    // Get valid user input, increment offset
    scanner := bufio.NewScanner(input)
    scanner.Scan()
    offset += int64(len(scanner.Bytes()) + 1)
    for !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()) {
        fmt.Fprintln(output, incorrectFormatPrompt)
        scanner.Scan()
        offset += int64(len(scanner.Bytes()) + 1)
    }

    // Seek input to offset
    if input != os.Stdin { input.Seek(offset, io.SeekStart) }

    // Return user input
    return scanner.Text()

}

