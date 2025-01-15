package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type MegapoolStatusResponse struct {
	Status   string          `json:"status"`
	Error    string          `json:"error"`
	Megapool MegapoolDetails `json:"megapoolDetails"`
}

type MegapoolDetails struct {
	Address                  common.Address `json:"address"`
	DelegateAddress          common.Address `json:"delegate"`
	EffectiveDelegateAddress common.Address `json:"effectiveDelegateAddress`
	Deployed                 bool           `json:"deployed"`
	ValidatorCount           uint16         `json:"validatorCount"`
	NodeDebt                 *big.Int       `json:"nodeDebt"`
	RefundValue              *big.Int       `json:"refundValue"`
	DelegateExpiry           uint64         `json:"delegateExpiry"`
	PendingRewards           *big.Int       `json:"pendingRewards"`
	NodeExpressTicketCount   uint64         `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool           `json:"useLatestDelegate"`
}

type MegapoolGetDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}

type MegapoolCanSetUseLatestDelegateResponse struct {
	Status  string             `json:"status"`
	Error   string             `json:"error"`
	GasInfo rocketpool.GasInfo `json:"gasInfo"`
}
type MegapoolSetUseLatestDelegateResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error"`
	TxHash common.Hash `json:"txHash"`
}

type MegapoolGetUseLatestDelegateResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Setting bool   `json:"setting"`
}

type MegapoolGetEffectiveDelegateResponse struct {
	Status  string         `json:"status"`
	Error   string         `json:"error"`
	Address common.Address `json:"address"`
}
