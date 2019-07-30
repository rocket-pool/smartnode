package cli

import (
    "io"
    "io/ioutil"
    "testing"
)


// Test user input prompt function
func TestPrompt(t *testing.T) {

    // Create temporary input file
    input, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    defer input.Close()

    // Write input to file
    io.WriteString(input,
        "foobar" + "\n" +
        "" + "\n" +
        "Y" + "\n" +
        "YES" + "\n" +
        "N" + "\n" +
        "NO" + "\n")
    input.Seek(0, io.SeekStart)

    // Test prompts
    input1 := Prompt(input, "Run test 'Y' [y/n]",   "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    input2 := Prompt(input, "Run test 'YES' [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    input3 := Prompt(input, "Run test 'N' [y/n]",   "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    input4 := Prompt(input, "Run test 'NO' [y/n]",  "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")

    // Check input values
    if input1 != "Y"   { t.Errorf("Incorrect input value: expected %s, got %s", "Y", input1) }
    if input2 != "YES" { t.Errorf("Incorrect input value: expected %s, got %s", "YES", input2) }
    if input3 != "N"   { t.Errorf("Incorrect input value: expected %s, got %s", "N", input3) }
    if input4 != "NO"  { t.Errorf("Incorrect input value: expected %s, got %s", "NO", input4) }

}

