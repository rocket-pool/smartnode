package prysm

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Get a byte slice value from a config map
func getConfigBytes(cfg map[string]string, key string) ([]byte, error) {
	valueStr, err := getConfigString(cfg, key)
	if err != nil {
		return []byte{}, err
	}
	value, err := deserializeBytes(valueStr)
	if err != nil {
		return []byte{}, fmt.Errorf("Could not parse config option '%s': %w", key, err)
	}
	return value, nil
}

// Get an unsigned integer value from a config map
func getConfigUint(cfg map[string]string, key string) (uint64, error) {
	valueStr, err := getConfigString(cfg, key)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(valueStr, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("Could not parse config option '%s': %w", key, err)
	}
	return value, nil
}

// Get a string value from a config map
func getConfigString(cfg map[string]string, key string) (string, error) {
	value, ok := cfg[key]
	if !ok {
		return "", fmt.Errorf("Config option '%s' not found", key)
	}
	return value, nil
}

// Deserialize a byte slice
func deserializeBytes(value string) ([]byte, error) {

	// Check format
	if !regexp.MustCompile("^\\[(\\d+( \\d+)*)?\\]$").MatchString(value) {
		return []byte{}, errors.New("Invalid byte slice format")
	}

	// Get byte strings
	byteStrings := strings.Fields(value[1 : len(value)-1])

	// Get and return bytes
	bytes := make([]byte, len(byteStrings))
	for bi, byteStr := range byteStrings {
		byteInt, err := strconv.ParseUint(byteStr, 10, 8)
		if err != nil {
			return []byte{}, errors.New("Invalid byte")
		}
		bytes[bi] = byte(byteInt)
	}
	return bytes, nil

}
