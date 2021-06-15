package rocketpool

import (
	"encoding/json"
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get wallet status
func (c *Client) WalletStatus() (api.WalletStatusResponse, error) {
    responseBytes, err := c.callAPI("wallet status")
    if err != nil {
        return api.WalletStatusResponse{}, fmt.Errorf("Could not get wallet status: %w", err)
    }
    var response api.WalletStatusResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.WalletStatusResponse{}, fmt.Errorf("Could not decode wallet status response: %w", err)
    }
    if response.Error != "" {
        return api.WalletStatusResponse{}, fmt.Errorf("Could not get wallet status: %s", response.Error)
    }
    return response, nil
}


// Set wallet password
func (c *Client) SetPassword(password string) (api.SetPasswordResponse, error) {
    responseBytes, err := c.callAPI("wallet set-password", password)
    if err != nil {
        return api.SetPasswordResponse{}, fmt.Errorf("Could not set wallet password: %w", err)
    }
    var response api.SetPasswordResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.SetPasswordResponse{}, fmt.Errorf("Could not decode set wallet password response: %w", err)
    }
    if response.Error != "" {
        return api.SetPasswordResponse{}, fmt.Errorf("Could not set wallet password: %s", response.Error)
    }
    return response, nil
}


// Initialize wallet
func (c *Client) InitWallet() (api.InitWalletResponse, error) {
    responseBytes, err := c.callAPI("wallet init")
    if err != nil {
        return api.InitWalletResponse{}, fmt.Errorf("Could not initialize wallet: %w", err)
    }
    var response api.InitWalletResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.InitWalletResponse{}, fmt.Errorf("Could not decode initialize wallet response: %w", err)
    }
    if response.Error != "" {
        return api.InitWalletResponse{}, fmt.Errorf("Could not initialize wallet: %s", response.Error)
    }
    return response, nil
}


// Recover wallet
func (c *Client) RecoverWallet(mnemonic string) (api.RecoverWalletResponse, error) {
    responseBytes, err := c.callAPI("wallet recover", mnemonic)
    if err != nil {
        return api.RecoverWalletResponse{}, fmt.Errorf("Could not recover wallet: %w", err)
    }
    var response api.RecoverWalletResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.RecoverWalletResponse{}, fmt.Errorf("Could not decode recover wallet response: %w", err)
    }
    if response.Error != "" {
        return api.RecoverWalletResponse{}, fmt.Errorf("Could not recover wallet: %s", response.Error)
    }
    return response, nil
}


// Rebuild wallet
func (c *Client) RebuildWallet() (api.RebuildWalletResponse, error) {
    responseBytes, err := c.callAPI("wallet rebuild")
    if err != nil {
        return api.RebuildWalletResponse{}, fmt.Errorf("Could not rebuild wallet: %w", err)
    }
    var response api.RebuildWalletResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.RebuildWalletResponse{}, fmt.Errorf("Could not decode rebuild wallet response: %w", err)
    }
    if response.Error != "" {
        return api.RebuildWalletResponse{}, fmt.Errorf("Could not rebuild wallet: %s", response.Error)
    }
    return response, nil
}


// Export wallet
func (c *Client) ExportWallet() (api.ExportWalletResponse, error) {
    responseBytes, err := c.callAPI("wallet export")
    if err != nil {
        return api.ExportWalletResponse{}, fmt.Errorf("Could not export wallet: %w", err)
    }
    var response api.ExportWalletResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ExportWalletResponse{}, fmt.Errorf("Could not decode export wallet response: %w", err)
    }
    if response.Error != "" {
        return api.ExportWalletResponse{}, fmt.Errorf("Could not export wallet: %s", response.Error)
    }
    return response, nil
}

