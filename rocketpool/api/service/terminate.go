package service

import (
	"os"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

const dataFolder string = "/.rocketpool/data"

// Deletes the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from Prater to Mainnet).
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

	// Remove it
	err = os.RemoveAll(dataFolder)
	if err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
