package api

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

type TNDAOUpgradeStatusResponse struct {
	Status                 string `json:"status"`
	Error                  string `json:"error"`
	UpgradeProposalCount   uint64 `json:"upgradeProposalCount"`
	UpgradeProposalState   string `json:"upgradeProposalState"`
	UpgradeProposalEndTime uint64 `json:"upgradeProposalEndTime"`
}

type TNDAOGetUpgradeProposalsResponse struct {
	Status    string                            `json:"status"`
	Error     string                            `json:"error"`
	Proposals []upgrades.UpgradeProposalDetails `json:"proposals"`
}

type CanExecuteUpgradeProposalResponse struct {
	Status             string          `json:"status"`
	Error              string          `json:"error"`
	CanExecute         bool            `json:"canExecute"`
	DoesNotExist       bool            `json:"doesNotExist"`
	InvalidTrustedNode bool            `json:"invalidTrustedNode"`
	InvalidState       bool            `json:"invalidState"`
	GasLimits          gaslimit.Limits `json:"gasLimits"`
}
type ExecuteUpgradeProposalResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}
