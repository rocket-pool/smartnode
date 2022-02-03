package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
)

type WalletStatusResponse struct {
	Status            string         `json:"status"`
	Error             string         `json:"error"`
	PasswordSet       bool           `json:"passwordSet"`
	WalletInitialized bool           `json:"walletInitialized"`
	AccountAddress    common.Address `json:"accountAddress"`
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
