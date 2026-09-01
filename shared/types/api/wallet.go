package api

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"

	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
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
	APIResponse
	PasswordSet       bool `json:"passwordSet"`
	WalletInitialized bool `json:"walletInitialized"`
	// When masquerading, AccountAddress represents the masqueraded address.
	// When using a normal wallet, AccountAddress represents the address derived from the wallet stored on disk
	AccountAddress common.Address `json:"accountAddress"`
	// NodeAddress always represents the address derived from the wallet stored on disk
	NodeAddress    common.Address `json:"nodeAddress"`
	IsMasquerading bool           `json:"isMasquerading"`
	IsObserve      bool           `json:"isObserve"`
}

type SetPasswordResponse struct {
	APIResponse
}

type InitWalletResponse struct {
	APIResponse
	Mnemonic       string         `json:"mnemonic"`
	AccountAddress common.Address `json:"accountAddress"`
}

type RecoverWalletResponse struct {
	APIResponse
	AccountAddress common.Address          `json:"accountAddress"`
	ValidatorKeys  []types.ValidatorPubkey `json:"validatorKeys"`
}

type SearchAndRecoverWalletResponse struct {
	APIResponse
	FoundWallet    bool                    `json:"foundWallet"`
	AccountAddress common.Address          `json:"accountAddress"`
	DerivationPath string                  `json:"derivationPath"`
	Index          uint                    `json:"index"`
	ValidatorKeys  []types.ValidatorPubkey `json:"validatorKeys"`
}

type RebuildWalletResponse struct {
	APIResponse
	ValidatorKeys []types.ValidatorPubkey `json:"validatorKeys"`
}

type KeyRecoveryStatus struct {
	Running        bool      `json:"running"`
	Operation      string    `json:"operation"`
	StartedAt      time.Time `json:"startedAt"`
	ElapsedSeconds float64   `json:"elapsedSeconds"`
	KeysFound      int       `json:"keysFound"`
	KeysTotal      int       `json:"keysTotal"`
	// False until the daemon has finished working out how many keys this node has
	TotalKnown bool `json:"totalKnown"`
}

type KeyRecoveryStatusResponse struct {
	APIResponse
	Recovery KeyRecoveryStatus `json:"recovery"`
}

type ExportWalletResponse struct {
	APIResponse
	Password          string `json:"password"`
	Wallet            string `json:"wallet"`
	AccountPrivateKey string `json:"accountPrivateKey"`
}

type SetEnsNameResponse struct {
	APIResponse
	Address   common.Address  `json:"address"`
	EnsName   string          `json:"ensName"`
	TxHash    common.Hash     `json:"txHash"`
	GasLimits gaslimit.Limits `json:"gasLimits"`
}

type TestMnemonicResponse struct {
	APIResponse
	CurrentAddress   common.Address `json:"currentAddress"`
	RecoveredAddress common.Address `json:"recoveredAddress"`
}

type PurgeResponse struct {
	APIResponse
}

type MasqueradeResponse struct {
	APIResponse
}

type EndMasqueradeResponse struct {
	APIResponse
}
