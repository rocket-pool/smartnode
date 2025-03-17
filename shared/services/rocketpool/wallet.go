package rocketpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

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
func (c *Client) InitWallet(derivationPath string) (api.InitWalletResponse, error) {
	responseBytes, err := c.callAPI("wallet init --derivation-path", derivationPath)
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
	command := "wallet recover "
	if skipValidatorKeyRecovery {
		command += "--skip-validator-key-recovery "
	}
	if walletIndex != 0 {
		command += fmt.Sprintf("--wallet-index %d ", walletIndex)
	}
	command += "--derivation-path"

	responseBytes, err := c.callAPI(command, derivationPath, mnemonic)
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
	command := "wallet search-and-recover "
	if skipValidatorKeyRecovery {
		command += "--skip-validator-key-recovery "
	}

	responseBytes, err := c.callAPI(command, mnemonic, address.Hex())
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

// Recover wallet
func (c *Client) TestRecoverWallet(mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (api.RecoverWalletResponse, error) {
	command := "wallet test-recovery "
	if skipValidatorKeyRecovery {
		command += "--skip-validator-key-recovery "
	}
	if walletIndex != 0 {
		command += fmt.Sprintf("--wallet-index %d ", walletIndex)
	}
	command += "--derivation-path"

	responseBytes, err := c.callAPI(command, derivationPath, mnemonic)
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

// Search and recover wallet
func (c *Client) TestSearchAndRecoverWallet(mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (api.SearchAndRecoverWalletResponse, error) {
	command := "wallet test-search-and-recover "
	if skipValidatorKeyRecovery {
		command += "--skip-validator-key-recovery "
	}

	responseBytes, err := c.callAPI(command, mnemonic, address.Hex())
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

// Estimate the gas required to set an ENS reverse record to a name
func (c *Client) EstimateGasSetEnsName(name string) (api.SetEnsNameResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("wallet estimate-gas-set-ens-name %s", name))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("wallet set-ens-name %s", name))
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

// Set the node address to an arbitrary address
func (c *Client) Masquerade(address common.Address) (api.MasqueradeResponse, error) {
	responseBytes, err := c.callAPI("wallet masquerade", address.Hex())
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

// Set the node address back to the wallet address
func (c *Client) RestoreAddress() (api.RestoreAddressResponse, error) {
	responseBytes, err := c.callAPI("wallet restore-address")
	if err != nil {
		return api.RestoreAddressResponse{}, fmt.Errorf("Could not restore address: %w", err)
	}
	var response api.RestoreAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RestoreAddressResponse{}, fmt.Errorf("Could not decode restore address response: %w", err)
	}
	if response.Error != "" {
		return api.RestoreAddressResponse{}, fmt.Errorf("Could not restore address: %s", response.Error)
	}
	return response, nil
}
