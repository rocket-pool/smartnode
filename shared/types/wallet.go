package types

type WalletStatus int

const (
	WalletStatus_Unknown          WalletStatus = iota
	WalletStatus_Ready            WalletStatus = iota
	WalletStatus_NoAddress        WalletStatus = iota
	WalletStatus_NoKeystore       WalletStatus = iota
	WalletStatus_NoPassword       WalletStatus = iota
	WalletStatus_KeystoreMismatch WalletStatus = iota
)

type DerivationPath string

const (
	DerivationPath_Default    DerivationPath = ""
	DerivationPath_LedgerLive DerivationPath = "ledger-live"
	DerivationPath_Mew        DerivationPath = "mew"
)
