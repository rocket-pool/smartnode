package prysm

import (
    "errors"
    "regexp"
    "strconv"
    "strings"
)


// Deserialize a byte array
func deserializeBytes(value string) ([]byte, error) {

    // Check format
    if !regexp.MustCompile("^\\[(\\d+( \\d+)*)?\\]$").MatchString(value) {
        return []byte{}, errors.New("Invalid byte array format")
    }

    // Get byte strings
    byteStrings := strings.Split(value[1:], " ")

    // Get and return bytes
    bytes := []byte{}
    for _, byteString := range byteStrings {
        byteInt, err := strconv.ParseUint(byteString, 10, 8)
        if err != nil {
            return []byte{}, errors.New("Invalid byte")
        }
        bytes = append(bytes, byte(byteInt))
    }
    return bytes, nil

}

