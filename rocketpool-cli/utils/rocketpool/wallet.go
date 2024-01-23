package rocketpool

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type WalletRequester struct {
	client *http.Client
}

func NewWalletRequester(client *http.Client) *WalletRequester {
	return &WalletRequester{
		client: client,
	}
}

func (r *WalletRequester) GetName() string {
	return "Wallet"
}
func (r *WalletRequester) GetRoute() string {
	return "wallet"
}
func (r *WalletRequester) GetClient() *http.Client {
	return r.client
}

// Delete the wallet keystore's password from disk
func (r *WalletRequester) DeletePassword() (*api.ApiResponse[api.SuccessData], error) {
	return sendGetRequest[api.SuccessData](r, "delete-password", "DeletePassword", nil)
}

// Export wallet
func (r *WalletRequester) Export() (*api.ApiResponse[api.WalletExportData], error) {
	return sendGetRequest[api.WalletExportData](r, "export", "Export", nil)
}

// Export the wallet in encrypted ETH key format
func (r *WalletRequester) ExportEthKey() (*api.ApiResponse[api.WalletExportEthKeyData], error) {
	return sendGetRequest[api.WalletExportEthKeyData](r, "export-eth-key", "ExportEthKey", nil)
}

// Initialize the wallet with a new key
func (r *WalletRequester) Initialize(derivationPath *string, index *uint64, password *string, save *bool) (*api.ApiResponse[api.WalletInitializeData], error) {
	args := map[string]string{}
	if derivationPath != nil {
		args["derivation-path"] = *derivationPath
	}
	if index != nil {
		args["index"] = fmt.Sprint(*index)
	}
	if password != nil {
		args["password"] = *password
	}
	if save != nil {
		args["save"] = fmt.Sprint(*save)
	}
	return sendGetRequest[api.WalletInitializeData](r, "initialize", "Initialize", args)
}

// Rebuild the validator keys associated with the wallet
func (r *WalletRequester) Rebuild() (*api.ApiResponse[api.WalletRebuildData], error) {
	return sendGetRequest[api.WalletRebuildData](r, "rebuild", "Rebuild", nil)
}

// Recover wallet
func (r *WalletRequester) Recover(derivationPath *string, mnemonic *string, skipValidatorKeyRecovery *bool, index *uint64, password *string, save *bool) (*api.ApiResponse[api.WalletRecoverData], error) {
	args := map[string]string{}
	if derivationPath != nil {
		args["derivation-path"] = *derivationPath
	}
	if mnemonic != nil {
		args["mnemonic"] = *mnemonic
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = fmt.Sprint(*skipValidatorKeyRecovery)
	}
	if index != nil {
		args["index"] = fmt.Sprint(*index)
	}
	if password != nil {
		args["password"] = *password
	}
	if save != nil {
		args["save-password"] = fmt.Sprint(*save)
	}
	return sendGetRequest[api.WalletRecoverData](r, "recover", "Recover", args)
}

// Search and recover wallet
func (r *WalletRequester) SearchAndRecover(mnemonic string, address common.Address, skipValidatorKeyRecovery *bool, password []byte, save *bool) (*api.ApiResponse[api.WalletSearchAndRecoverData], error) {
	args := map[string]string{
		"mnemonic": mnemonic,
		"address":  address.Hex(),
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = fmt.Sprint(*skipValidatorKeyRecovery)
	}
	if password != nil {
		args["password"] = hex.EncodeToString(password)
	}
	if save != nil {
		args["save"] = fmt.Sprint(*save)
	}
	return sendGetRequest[api.WalletSearchAndRecoverData](r, "search-and-recover", "SearchAndRecover", args)
}

// Set an ENS reverse record to a name
func (r *WalletRequester) SetEnsName(name string) (*api.ApiResponse[api.WalletSetEnsNameData], error) {
	args := map[string]string{
		"name": name,
	}
	return sendGetRequest[api.WalletSetEnsNameData](r, "set-ens-name", "SetEnsName", args)
}

// Sets the wallet keystore's password
func (r *WalletRequester) SetPassword(password []byte, save bool) (*api.ApiResponse[api.SuccessData], error) {
	args := map[string]string{
		"password": hex.EncodeToString(password),
		"save":     fmt.Sprint(save),
	}
	return sendGetRequest[api.SuccessData](r, "set-password", "SetPassword", args)
}

// Get wallet status
func (r *WalletRequester) Status() (*api.ApiResponse[api.WalletStatusData], error) {
	return sendGetRequest[api.WalletStatusData](r, "status", "Status", nil)
}

// Search for and recover the wallet in test-mode so none of the artifacts are saved
func (r *WalletRequester) TestSearchAndRecover(mnemonic string, address common.Address, skipValidatorKeyRecovery *bool) (*api.ApiResponse[api.WalletSearchAndRecoverData], error) {
	args := map[string]string{
		"mnemonic": mnemonic,
		"address":  address.Hex(),
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = fmt.Sprint(*skipValidatorKeyRecovery)
	}
	return sendGetRequest[api.WalletSearchAndRecoverData](r, "test-search-and-recover", "TestSearchAndRecover", args)
}

// Recover wallet in test-mode so none of the artifacts are saved
func (r *WalletRequester) TestRecover(derivationPath *string, mnemonic string, skipValidatorKeyRecovery *bool, index *uint64) (*api.ApiResponse[api.WalletRecoverData], error) {
	args := map[string]string{
		"mnemonic": mnemonic,
	}
	if derivationPath != nil {
		args["derivation-path"] = *derivationPath
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = fmt.Sprint(*skipValidatorKeyRecovery)
	}
	if index != nil {
		args["index"] = fmt.Sprint(*index)
	}
	return sendGetRequest[api.WalletRecoverData](r, "test-recover", "TestRecover", args)
}

// Sends a zero-value message with a payload
func (r *WalletRequester) SendMessage(message []byte, address common.Address) (*api.ApiResponse[api.TxInfoData], error) {
	args := map[string]string{
		"message": hex.EncodeToString(message),
		"address": address.Hex(),
	}
	return sendGetRequest[api.TxInfoData](r, "send-message", "SendMessage", args)
}

// Use the node private key to sign an arbitrary message
func (r *WalletRequester) SignMessage(message []byte) (*api.ApiResponse[api.WalletSignMessageData], error) {
	args := map[string]string{
		"message": hex.EncodeToString(message),
	}
	return sendGetRequest[api.WalletSignMessageData](r, "sign-message", "SignMessage", args)
}
