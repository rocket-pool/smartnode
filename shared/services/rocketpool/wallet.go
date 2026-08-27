package rocketpool

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get wallet status
func (c *Client) WalletStatus() (api.WalletStatusResponse, error) {
	return c.callAPI[api.WalletStatusResponse]("GET", "/api/wallet/status", nil, "Could not get wallet status")
}

// Set wallet password
func (c *Client) SetPassword(password string) (api.SetPasswordResponse, error) {
	return c.callAPI[api.SetPasswordResponse]("POST", "/api/wallet/set-password", url.Values{"password": {password}}, "Could not set wallet password")
}

// Initialize wallet
func (c *Client) InitWallet(derivationPath string) (api.InitWalletResponse, error) {
	return c.callAPI[api.InitWalletResponse]("POST", "/api/wallet/init", url.Values{"derivationPath": {derivationPath}}, "Could not initialize wallet")
}

// Recover wallet
func (c *Client) RecoverWallet(mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (api.RecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	return c.callAPI[api.RecoverWalletResponse]("POST", "/api/wallet/recover", url.Values{
		"mnemonic":                 {mnemonic},
		"skipValidatorKeyRecovery": {skipStr},
		"derivationPath":           {derivationPath},
		"walletIndex":              {fmt.Sprintf("%d", walletIndex)},
	}, "Could not recover wallet")
}

// Search and recover wallet
func (c *Client) SearchAndRecoverWallet(mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (api.SearchAndRecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	return c.callAPICtx[api.SearchAndRecoverWalletResponse](context.Background(), "POST", "/api/wallet/search-and-recover", url.Values{
		"mnemonic":                 {mnemonic},
		"address":                  {address.Hex()},
		"skipValidatorKeyRecovery": {skipStr},
	}, "Could not search and recover wallet")
}

// Recover wallet (test, no save)
func (c *Client) TestRecoverWallet(mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (api.RecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	return c.callAPI[api.RecoverWalletResponse]("POST", "/api/wallet/test-recover", url.Values{
		"mnemonic":                 {mnemonic},
		"skipValidatorKeyRecovery": {skipStr},
		"derivationPath":           {derivationPath},
		"walletIndex":              {fmt.Sprintf("%d", walletIndex)},
	}, "Could not test recover wallet")
}

// Search and recover wallet (test, no save)
func (c *Client) TestSearchAndRecoverWallet(mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (api.SearchAndRecoverWalletResponse, error) {
	skipStr := "false"
	if skipValidatorKeyRecovery {
		skipStr = "true"
	}
	return c.callAPI[api.SearchAndRecoverWalletResponse]("POST", "/api/wallet/test-search-and-recover", url.Values{
		"mnemonic":                 {mnemonic},
		"address":                  {address.Hex()},
		"skipValidatorKeyRecovery": {skipStr},
	}, "Could not test search and recover wallet")
}

// Rebuild wallet
func (c *Client) RebuildWallet() (api.RebuildWalletResponse, error) {
	// removed timeout as large nodes were exceeding it
	return c.callAPICtx[api.RebuildWalletResponse](context.Background(), "POST", "/api/wallet/rebuild", nil, "Could not rebuild wallet")
}

// Get the status of any validator key recovery currently running
func (c *Client) GetKeyRecoveryStatus() (api.KeyRecoveryStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.callAPICtx[api.KeyRecoveryStatusResponse](ctx, "GET", "/api/wallet/recovery-status", nil, "Could not get key recovery status")
}

// Estimate the gas required to set an ENS reverse record to a name
func (c *Client) EstimateGasSetEnsName(name string) (api.SetEnsNameResponse, error) {
	return c.callAPI[api.SetEnsNameResponse]("GET", "/api/wallet/estimate-gas-set-ens-name", url.Values{"name": {name}}, "Could not get estimate-gas-set-ens-name response")
}

// Set an ENS reverse record to a name
func (c *Client) SetEnsName(name string) (api.SetEnsNameResponse, error) {
	return c.callAPI[api.SetEnsNameResponse]("POST", "/api/wallet/set-ens-name", url.Values{"name": {name}}, "Could not update ENS record")
}

// Export wallet
func (c *Client) ExportWallet() (api.ExportWalletResponse, error) {
	return c.callAPI[api.ExportWalletResponse]("GET", "/api/wallet/export", nil, "Could not export wallet")
}

// Set the node address to an arbitrary address
func (c *Client) Masquerade(address common.Address, observe bool) (api.MasqueradeResponse, error) {
	observeStr := "false"
	if observe {
		observeStr = "true"
	}
	return c.callAPI[api.MasqueradeResponse]("POST", "/api/wallet/masquerade", url.Values{"address": {address.Hex()}, "observe": {observeStr}}, "Could not masquerade wallet")
}

// Delete the address file, ending a masquerade
func (c *Client) EndMasquerade() (api.EndMasqueradeResponse, error) {
	return c.callAPI[api.EndMasqueradeResponse]("POST", "/api/wallet/end-masquerade", nil, "Could not end masquerade")
}
