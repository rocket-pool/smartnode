package validator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/rocket-pool/node-manager-core/node/wallet"
)

// Load the next account file from disk
func loadNextAccount(nextAccountPath string) (uint64, error) {
	_, err := os.Stat(nextAccountPath)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("error checking for next account data at [%s]: %w", nextAccountPath, err)
	}

	// Read the file
	bytes, err := os.ReadFile(nextAccountPath)
	if err != nil {
		return 0, fmt.Errorf("error reading next account data at [%s]: %w", nextAccountPath, err)
	}

	// Parse the account
	nextAccountString := string(bytes)
	nextAccountString = strings.TrimSpace(nextAccountString)
	nextAccount, err := strconv.ParseUint(nextAccountString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing next account data at [%s]: %w", nextAccountPath, err)
	}

	return nextAccount, nil
}

// Save the next account file
func saveNextAccount(nextAccount uint64, nextAccountPath string) error {
	nextAccountString := strconv.FormatUint(nextAccount, 10)
	bytes := []byte(nextAccountString)

	// Write the file
	err := os.WriteFile(nextAccountPath, bytes, wallet.FileMode)
	if err != nil {
		return fmt.Errorf("error writing next account data to [%s]: %w", nextAccountPath, err)
	}
	return nil
}
