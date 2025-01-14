package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type MegapoolStatusResponse struct {
	Status   string          `json:"status"`
	Error    string          `json:"error"`
	Megapool MegapoolDetails `json:"megapoolDetails"`
}

type MegapoolDetails struct {
	Address                common.Address `json:"address"`
	DelegateAddress        common.Address `json:"delegate"`
	Deployed               bool           `json:"deployed"`
	ValidatorCount         uint16         `json:"validatorCount"`
	NodeDebt               *big.Int       `json:"nodeDebt"`
	RefundValue            *big.Int       `json:"refundValue"`
	DelegateExpiry         uint64         `json:"delegateExpiry"`
	PendingRewards         *big.Int       `json:"pendingRewards"`
	NodeExpressTicketCount uint64         `json:"nodeExpressTicketCount"`
	UseLatestDelegate      bool           `json:"useLatestDelegate"`
}
