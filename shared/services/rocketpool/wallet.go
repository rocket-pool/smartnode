package rocketpool

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get wallet status
func (c *Client) WalletStatus() (api.WalletStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/wallet/status", nil)
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
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/set-password", url.Values{"password": {password}})
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
func (c *Client) InitWallet(derivationPath string) (api.InitWalletResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/init", url.Values{"derivationPath": {derivationPath}})
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
func (c *Client) RecoverWallet(mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (api.RecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/recover", url.Values{
		"mnemonic":                 {mnemonic},
		"skipValidatorKeyRecovery": {skipStr},
		"derivationPath":           {derivationPath},
		"walletIndex":              {fmt.Sprintf("%d", walletIndex)},
	})
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

// Search and recover wallet
func (c *Client) SearchAndRecoverWallet(mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (api.SearchAndRecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/search-and-recover", url.Values{
		"mnemonic":                 {mnemonic},
		"address":                  {address.Hex()},
		"skipValidatorKeyRecovery": {skipStr},
	})
	if err != nil {
		return api.SearchAndRecoverWalletResponse{}, fmt.Errorf("Could not search and recover wallet: %w", err)
	}
	var response api.SearchAndRecoverWalletResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SearchAndRecoverWalletResponse{}, fmt.Errorf("Could not decode search-and-recover wallet response: %w", err)
	}
	if response.Error != "" {
		return api.SearchAndRecoverWalletResponse{}, fmt.Errorf("Could not search and recover wallet: %s", response.Error)
	}
	return response, nil
}

// Recover wallet (test, no save)
func (c *Client) TestRecoverWallet(mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (api.RecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/test-recover", url.Values{
		"mnemonic":                 {mnemonic},
		"skipValidatorKeyRecovery": {skipStr},
		"derivationPath":           {derivationPath},
		"walletIndex":              {fmt.Sprintf("%d", walletIndex)},
	})
	if err != nil {
		return api.RecoverWalletResponse{}, fmt.Errorf("Could not test recover wallet: %w", err)
	}
	var response api.RecoverWalletResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RecoverWalletResponse{}, fmt.Errorf("Could not decode test recover wallet response: %w", err)
	}
	if response.Error != "" {
		return api.RecoverWalletResponse{}, fmt.Errorf("Could not test recover wallet: %s", response.Error)
	}
	return response, nil
}

// Search and recover wallet (test, no save)
func (c *Client) TestSearchAndRecoverWallet(mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (api.SearchAndRecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/test-search-and-recover", url.Values{
		"mnemonic":                 {mnemonic},
		"address":                  {address.Hex()},
		"skipValidatorKeyRecovery": {skipStr},
	})
	if err != nil {
		return api.SearchAndRecoverWalletResponse{}, fmt.Errorf("Could not test search and recover wallet: %w", err)
	}
	var response api.SearchAndRecoverWalletResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SearchAndRecoverWalletResponse{}, fmt.Errorf("Could not decode test-search-and-recover wallet response: %w", err)
	}
	if response.Error != "" {
		return api.SearchAndRecoverWalletResponse{}, fmt.Errorf("Could not test search and recover wallet: %s", response.Error)
	}
	return response, nil
}

// Rebuild wallet
func (c *Client) RebuildWallet() (api.RebuildWalletResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/rebuild", nil)
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

// Estimate the gas required to set an ENS reverse record to a name
func (c *Client) EstimateGasSetEnsName(name string) (api.SetEnsNameResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/wallet/estimate-gas-set-ens-name", url.Values{"name": {name}})
	if err != nil {
		return api.SetEnsNameResponse{}, fmt.Errorf("Could not get estimate-gas-set-ens-name response: %w", err)
	}
	var response api.SetEnsNameResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetEnsNameResponse{}, fmt.Errorf("Could not decode estimate-gas-set-ens-name response: %w", err)
	}
	if response.Error != "" {
		return api.SetEnsNameResponse{}, fmt.Errorf("Could not get estimate-gas-set-ens-name response: %s", response.Error)
	}
	return response, nil
}

// Set an ENS reverse record to a name
func (c *Client) SetEnsName(name string) (api.SetEnsNameResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/set-ens-name", url.Values{"name": {name}})
	if err != nil {
		return api.SetEnsNameResponse{}, fmt.Errorf("Could not update ENS record: %w", err)
	}
	var response api.SetEnsNameResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetEnsNameResponse{}, fmt.Errorf("Could not decode set-ens-name response: %w", err)
	}
	if response.Error != "" {
		return api.SetEnsNameResponse{}, fmt.Errorf("Could not update ENS record: %s", response.Error)
	}
	return response, nil
}

// Export wallet
func (c *Client) ExportWallet() (api.ExportWalletResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/wallet/export", nil)
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

// Set the node address to an arbitrary address
func (c *Client) Masquerade(address common.Address) (api.MasqueradeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/masquerade", url.Values{"address": {address.Hex()}})
	if err != nil {
		return api.MasqueradeResponse{}, fmt.Errorf("Could not masquerade wallet: %w", err)
	}
	var response api.MasqueradeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MasqueradeResponse{}, fmt.Errorf("Could not decode masquerade wallet response: %w", err)
	}
	if response.Error != "" {
		return api.MasqueradeResponse{}, fmt.Errorf("Could not masquerade wallet: %s", response.Error)
	}
	return response, nil
}

// Delete the address file, ending a masquerade
func (c *Client) EndMasquerade() (api.EndMasqueradeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/wallet/end-masquerade", nil)
	if err != nil {
		return api.EndMasqueradeResponse{}, fmt.Errorf("Could not end masquerade: %w", err)
	}
	var response api.EndMasqueradeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.EndMasqueradeResponse{}, fmt.Errorf("Could not decode end masquerade response: %w", err)
	}
	if response.Error != "" {
		return api.EndMasqueradeResponse{}, fmt.Errorf("Could not end masquerade: %s", response.Error)
	}
	return response, nil
}
