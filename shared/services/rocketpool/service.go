package rocketpool

import (
	"encoding/json"
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Deletes the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from Prater to Mainnet).
func (c *Client) TerminateDataFolder() (api.TerminateDataFolderResponse, error) {
	responseBytes, err := c.callAPI("service terminate-data-folder")
	if err != nil {
		return api.TerminateDataFolderResponse{}, fmt.Errorf("Could not delete data folder: %w", err)
	}
	var response api.TerminateDataFolderResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.TerminateDataFolderResponse{}, fmt.Errorf("Could not decode terminate-data-folder response: %w", err)
	}
	if response.Error != "" {
		return api.TerminateDataFolderResponse{}, fmt.Errorf("Could not delete data folder: %s", response.Error)
	}
	return response, nil
}
