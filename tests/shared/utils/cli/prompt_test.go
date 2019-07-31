package cli

import (
    "testing"

    "github.com/rocket-pool/smartnode/shared/utils/cli"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test user input prompt function
func TestPrompt(t *testing.T) {

    // Create temporary input file
    input, err := test.NewInputFile(
        "foobar" + "\n" +
        "" + "\n" +
        "Y" + "\n" +
        "YES" + "\n" +
        "N" + "\n" +
        "NO" + "\n")
    if err != nil { t.Fatal(err) }
    defer input.Close()

    // Test prompts
    if input := cli.Prompt(input, "Run test 'Y' [y/n]",   "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'"); input != "Y" {
        t.Errorf("Incorrect input value: expected %s, got %s", "Y", input)
    }
    if input := cli.Prompt(input, "Run test 'YES' [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'"); input != "YES" {
        t.Errorf("Incorrect input value: expected %s, got %s", "YES", input)
    }
    if input := cli.Prompt(input, "Run test 'N' [y/n]",   "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'"); input != "N" {
        t.Errorf("Incorrect input value: expected %s, got %s", "N", input)
    }
    if input := cli.Prompt(input, "Run test 'NO' [y/n]",  "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'"); input != "NO" {
        t.Errorf("Incorrect input value: expected %s, got %s", "NO", input)
    }

}

