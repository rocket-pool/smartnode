package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

const dataFolder string = "/.rocketpool/data"

// Deletes the contents of the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from a Testnet to Mainnet).
func terminateDataFolder(c *cli.Context) (*api.TerminateDataFolderResponse, error) {

	// Response
	response := api.TerminateDataFolderResponse{}

	// Check if it exists
	_, err := os.Stat(dataFolder)
	if os.IsNotExist(err) {
		response.FolderExisted = false
		return &response, nil
	}
	response.FolderExisted = true

	// Traverse it
	files, err := os.ReadDir(dataFolder)
	if err != nil {
		return nil, fmt.Errorf("error enumerating files: %w", err)
	}

	// Delete the children
	for _, file := range files {
		// Skip the validators directory - that get special treatment
		if file.Name() != "validators" && !file.IsDir() {
			fullPath := filepath.Join(dataFolder, file.Name())
			if file.IsDir() {
				err = os.RemoveAll(fullPath)
			} else {
				err = os.Remove(fullPath)
			}
			if err != nil {
				return nil, fmt.Errorf("error removing [%s]: %w", file.Name(), err)
			}
		}
	}

	// Traverse the validators dir
	validatorsDir := filepath.Join(dataFolder, "validators")
	files, err = os.ReadDir(validatorsDir)
	if err != nil {
		return nil, fmt.Errorf("error enumerating validator files: %w", err)
	}

	// Delete the children
	for _, file := range files {
		fullPath := filepath.Join(validatorsDir, file.Name())
		if file.IsDir() {
			err = os.RemoveAll(fullPath)
		} else {
			err = os.Remove(fullPath)
		}
		if err != nil {
			return nil, fmt.Errorf("error removing [%s]: %w", file.Name(), err)
		}
	}

	// Return response
	return &response, nil

}
