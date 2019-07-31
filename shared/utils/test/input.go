package test

import (
    "io"
    "io/ioutil"
    "os"
)


// Create a temporary input file
func NewInputFile(contents string) (f *os.File, err error) {

    // Create temporary input file
    input, err := ioutil.TempFile("", "")
    if err != nil { return nil, err }

    // Write contents to file
    if _, err := io.WriteString(input, contents); err != nil { return nil, err }

    // Seek to start of file
    if _, err := input.Seek(0, io.SeekStart); err != nil { return nil, err }

    // Return
    return input, nil

}

