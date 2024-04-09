package client

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type WalletRequester struct {
	context *client.RequesterContext
}

func NewWalletRequester(context *client.RequesterContext) *WalletRequester {
	return &WalletRequester{
		context: context,
	}
}

func (r *WalletRequester) GetName() string {
	return "Wallet"
}
func (r *WalletRequester) GetRoute() string {
	return "wallet"
}
func (r *WalletRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Delete the wallet keystore's password from disk
func (r *WalletRequester) DeletePassword() (*types.ApiResponse[types.SuccessData], error) {
	return client.SendGetRequest[types.SuccessData](r, "delete-password", "DeletePassword", nil)
}

// Export wallet
func (r *WalletRequester) Export() (*types.ApiResponse[api.WalletExportData], error) {
	return client.SendGetRequest[api.WalletExportData](r, "export", "Export", nil)
}

// Export the wallet in encrypted ETH key format
func (r *WalletRequester) ExportEthKey() (*types.ApiResponse[api.WalletExportEthKeyData], error) {
	return client.SendGetRequest[api.WalletExportEthKeyData](r, "export-eth-key", "ExportEthKey", nil)
}

// Initialize the wallet with a new key
func (r *WalletRequester) Initialize(derivationPath *string, index *uint64, saveWallet bool, password string, savePassword bool) (*types.ApiResponse[api.WalletInitializeData], error) {
	args := map[string]string{
		"password":      password,
		"save-wallet":   strconv.FormatBool(saveWallet),
		"save-password": strconv.FormatBool(savePassword),
	}
	if derivationPath != nil {
		args["derivation-path"] = *derivationPath
	}
	if index != nil {
		args["index"] = fmt.Sprint(*index)
	}
	return client.SendGetRequest[api.WalletInitializeData](r, "initialize", "Initialize", args)
}

// Set the node address to an arbitrary address
func (r *WalletRequester) Masquerade(address common.Address) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"address": address.Hex(),
	}
	return client.SendGetRequest[types.SuccessData](r, "masquerade", "Masquerade", args)
}

// Rebuild the validator keys associated with the wallet
func (r *WalletRequester) Rebuild() (*types.ApiResponse[api.WalletRebuildData], error) {
	return client.SendGetRequest[api.WalletRebuildData](r, "rebuild", "Rebuild", nil)
}

// Recover wallet
func (r *WalletRequester) Recover(derivationPath *string, mnemonic string, skipValidatorKeyRecovery *bool, index *uint64, password string, save bool) (*types.ApiResponse[api.WalletRecoverData], error) {
	args := map[string]string{
		"password":      password,
		"save-password": strconv.FormatBool(save),
		"mnemonic":      mnemonic,
	}
	if derivationPath != nil {
		args["derivation-path"] = *derivationPath
	}
	if index != nil {
		args["index"] = strconv.FormatUint(*index, 10)
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = strconv.FormatBool(*skipValidatorKeyRecovery)
	}
	return client.SendGetRequest[api.WalletRecoverData](r, "recover", "Recover", args)
}

// Set the node address back to the wallet address
func (r *WalletRequester) RestoreAddress() (*types.ApiResponse[types.SuccessData], error) {
	return client.SendGetRequest[types.SuccessData](r, "restore-address", "RestoreAddress", nil)
}

// Search and recover wallet
func (r *WalletRequester) SearchAndRecover(mnemonic string, address common.Address, skipValidatorKeyRecovery *bool, password string, save bool) (*types.ApiResponse[api.WalletSearchAndRecoverData], error) {
	args := map[string]string{
		"mnemonic":      mnemonic,
		"address":       address.Hex(),
		"password":      password,
		"save-password": strconv.FormatBool(save),
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = fmt.Sprint(*skipValidatorKeyRecovery)
	}
	return client.SendGetRequest[api.WalletSearchAndRecoverData](r, "search-and-recover", "SearchAndRecover", args)
}

// Set an ENS reverse record to a name
func (r *WalletRequester) SetEnsName(name string) (*types.ApiResponse[api.WalletSetEnsNameData], error) {
	args := map[string]string{
		"name": name,
	}
	return client.SendGetRequest[api.WalletSetEnsNameData](r, "set-ens-name", "SetEnsName", args)
}

// Sets the wallet keystore's password
func (r *WalletRequester) SetPassword(password string, save bool) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"password": password,
		"save":     fmt.Sprint(save),
	}
	return client.SendGetRequest[types.SuccessData](r, "set-password", "SetPassword", args)
}

// Get wallet status
func (r *WalletRequester) Status() (*types.ApiResponse[api.WalletStatusData], error) {
	return client.SendGetRequest[api.WalletStatusData](r, "status", "Status", nil)
}

// Recover wallet in test-mode so none of the artifacts are saved
func (r *WalletRequester) TestRecover(derivationPath *string, mnemonic string, skipValidatorKeyRecovery *bool, index *uint64) (*types.ApiResponse[api.WalletRecoverData], error) {
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
	return client.SendGetRequest[api.WalletRecoverData](r, "test-recover", "TestRecover", args)
}

// Search for and recover the wallet in test-mode so none of the artifacts are saved
func (r *WalletRequester) TestSearchAndRecover(mnemonic string, address common.Address, skipValidatorKeyRecovery *bool) (*types.ApiResponse[api.WalletSearchAndRecoverData], error) {
	args := map[string]string{
		"mnemonic": mnemonic,
		"address":  address.Hex(),
	}
	if skipValidatorKeyRecovery != nil {
		args["skip-validator-key-recovery"] = fmt.Sprint(*skipValidatorKeyRecovery)
	}
	return client.SendGetRequest[api.WalletSearchAndRecoverData](r, "test-search-and-recover", "TestSearchAndRecover", args)
}

// Sends a zero-value message with a payload
func (r *WalletRequester) SendMessage(message []byte, address common.Address) (*types.ApiResponse[types.TxInfoData], error) {
	args := map[string]string{
		"message": hex.EncodeToString(message),
		"address": address.Hex(),
	}
	return client.SendGetRequest[types.TxInfoData](r, "send-message", "SendMessage", args)
}

// Use the node private key to sign an arbitrary message
func (r *WalletRequester) SignMessage(message []byte) (*types.ApiResponse[api.WalletSignMessageData], error) {
	args := map[string]string{
		"message": hex.EncodeToString(message),
	}
	return client.SendGetRequest[api.WalletSignMessageData](r, "sign-message", "SignMessage", args)
}

// Use the node private key to sign a transaction
func (r *WalletRequester) SignTx(message []byte) (*types.ApiResponse[api.WalletSignTxData], error) {
	args := map[string]string{
		"tx": hex.EncodeToString(message),
	}
	return client.SendGetRequest[api.WalletSignTxData](r, "sign-tx", "SignTx", args)
}
