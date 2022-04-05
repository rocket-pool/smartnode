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

// Creates the fee recipient file for the validator container.
func (c *Client) CreateFeeRecipientFile() (api.CreateFeeRecipientFileResponse, error) {
	responseBytes, err := c.callAPI("service create-fee-recipient-file")
	if err != nil {
		return api.CreateFeeRecipientFileResponse{}, fmt.Errorf("Could not create fee recipient file: %w", err)
	}
	var response api.CreateFeeRecipientFileResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CreateFeeRecipientFileResponse{}, fmt.Errorf("Could not decode create-fee-recipient-file response: %w", err)
	}
	if response.Error != "" {
		return api.CreateFeeRecipientFileResponse{}, fmt.Errorf("Could not create fee recipient file: %s", response.Error)
	}
	return response, nil
}
