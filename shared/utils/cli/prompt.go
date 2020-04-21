package cli

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "regexp"
    "strconv"
    "strings"
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


// Prompt for user selection
func PromptSelect(input *os.File, output *os.File, initialPrompt string, options []string) string {

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
    response := Prompt(input, output, prompt, expectedFormat, "Please enter a number corresponding to an option")

    // Get selected option
    index, _ := strconv.Atoi(response)
    selected := options[index - 1]

    // Return
    return selected

}

