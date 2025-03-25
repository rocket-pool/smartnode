package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BeaconStateResponse struct {
	Proof []string `json:"proof"`
}

type WithdrawalProofResponse struct {
	Slot           uint64   `json:"slot"`
	WithdrawalSlot uint64   `json:"withdrawalSlot"`
	ValidatorIndex uint64   `json:"validatorIndex"`
	Amount         *big.Int `json:"amount"`
	Proof          []string `json:"proof"`

	// Contract refers to this as _withdrawalNum
	IndexInWithdrawalsArray int `json:"indexInWithdrawalsArray"`
	// Part of the Withdrawal calldata
	WithdrawalIndex   uint64         `json:"withdrawalIndex"`
	WithdrawalAddress common.Address `json:"withdrawalAddress"`
}
