package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Encrypted validator keystore following the EIP-2335 standard
// (https://eips.ethereum.org/EIPS/eip-2335)
type ValidatorKeystore struct {
	Crypto  map[string]interface{} `json:"crypto"`
	Version uint                   `json:"version"`
	UUID    uuid.UUID              `json:"uuid"`
	Path    string                 `json:"path"`
	Pubkey  beacon.ValidatorPubkey `json:"pubkey"`
}

type WalletStatusData struct {
	WalletStatus wallet.WalletStatus `json:"walletStatus"`
}

type WalletInitializeData struct {
	Mnemonic       string         `json:"mnemonic"`
	AccountAddress common.Address `json:"accountAddress"`
}

type WalletRecoverData struct {
	AccountAddress common.Address           `json:"accountAddress"`
	ValidatorKeys  []beacon.ValidatorPubkey `json:"validatorKeys"`
}

type WalletSearchAndRecoverData struct {
	FoundWallet    bool                     `json:"foundWallet"`
	AccountAddress common.Address           `json:"accountAddress"`
	DerivationPath string                   `json:"derivationPath"`
	Index          uint                     `json:"index"`
	ValidatorKeys  []beacon.ValidatorPubkey `json:"validatorKeys"`
}

type WalletRebuildData struct {
	ValidatorKeys []beacon.ValidatorPubkey `json:"validatorKeys"`
}

type WalletExportData struct {
	Password          string `json:"password"`
	Wallet            string `json:"wallet"`
	AccountPrivateKey []byte `json:"accountPrivateKey"`
}

type WalletSetEnsNameData struct {
	Address common.Address       `json:"address"`
	EnsName string               `json:"ensName"`
	TxInfo  *eth.TransactionInfo `json:"txInfo"`
}

type WalletTestMnemonicData struct {
	CurrentAddress   common.Address `json:"currentAddress"`
	RecoveredAddress common.Address `json:"recoveredAddress"`
}

type WalletSignMessageData struct {
	SignedMessage []byte `json:"signedMessage"`
}

type WalletSignTxData struct {
	SignedTx []byte `json:"signedTx"`
}

type WalletExportEthKeyData struct {
	EthKeyJson []byte `json:"ethKeyJson"`
	Password   string `json:"password"`
}
