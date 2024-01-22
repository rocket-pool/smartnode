package types

import "github.com/ethereum/go-ethereum/common"

type WalletStatus struct {
	NodeAddress     common.Address `json:"nodeAddress"`
	KeystoreAddress common.Address `json:"keystoreAddress"`
	HasAddress      bool           `json:"hasAddress"`
	HasPassword     bool           `json:"hasPassword"`
	HasKeystore     bool           `json:"hasKeystore"`
	IsPasswordSaved bool           `json:"isPasswordSaved"`
}

type DerivationPath string

const (
	DerivationPath_Default    DerivationPath = ""
	DerivationPath_LedgerLive DerivationPath = "ledger-live"
	DerivationPath_Mew        DerivationPath = "mew"
)
