package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
)

// Encrypted validator keystore following the EIP-2335 standard
// (https://eips.ethereum.org/EIPS/eip-2335)
type ValidatorKeystore struct {
	Crypto  map[string]interface{} `json:"crypto"`
	Version uint                   `json:"version"`
	UUID    uuid.UUID              `json:"uuid"`
	Path    string                 `json:"path"`
	Pubkey  types.ValidatorPubkey  `json:"pubkey"`
}

type WalletStatusResponse struct {
	Status            string `json:"status"`
	Error             string `json:"error"`
	PasswordSet       bool   `json:"passwordSet"`
	WalletInitialized bool   `json:"walletInitialized"`
	// When masquerading, AccountAddress represents the masqueraded address.
	// When using a normal wallet, AccountAddress represents the address derived from the wallet stored on disk
	AccountAddress common.Address `json:"accountAddress"`
	// NodeAddress always represents the address drived from the wallet stored on disk
	NodeAddress    common.Address `json:"nodeAddress"`
	IsMasquerading bool           `json:"isMasquerading"`
}

type SetPasswordResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type InitWalletResponse struct {
	Status         string         `json:"status"`
	Error          string         `json:"error"`
	Mnemonic       string         `json:"mnemonic"`
	AccountAddress common.Address `json:"accountAddress"`
}

type RecoverWalletResponse struct {
	Status         string                  `json:"status"`
	Error          string                  `json:"error"`
	AccountAddress common.Address          `json:"accountAddress"`
	ValidatorKeys  []types.ValidatorPubkey `json:"validatorKeys"`
}

type SearchAndRecoverWalletResponse struct {
	Status         string                  `json:"status"`
	Error          string                  `json:"error"`
	FoundWallet    bool                    `json:"foundWallet"`
	AccountAddress common.Address          `json:"accountAddress"`
	DerivationPath string                  `json:"derivationPath"`
	Index          uint                    `json:"index"`
	ValidatorKeys  []types.ValidatorPubkey `json:"validatorKeys"`
}

type RebuildWalletResponse struct {
	Status        string                  `json:"status"`
	Error         string                  `json:"error"`
	ValidatorKeys []types.ValidatorPubkey `json:"validatorKeys"`
}

type ExportWalletResponse struct {
	Status            string `json:"status"`
	Error             string `json:"error"`
	Password          string `json:"password"`
	Wallet            string `json:"wallet"`
	AccountPrivateKey string `json:"accountPrivateKey"`
}

type SetEnsNameResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	Address common.Address     `json:"address"`
	EnsName string             `json:"ensName"`
	TxHash  common.Hash        `json:"txHash"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}

type TestMnemonicResponse struct {
	Status           string         `json:"status"`
	Error            string         `json:"error"`
	CurrentAddress   common.Address `json:"currentAddress"`
	RecoveredAddress common.Address `json:"recoveredAddress"`
}

type PurgeResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type MasqueradeResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type EndMasqueradeResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
