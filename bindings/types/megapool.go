package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type DepositData struct {
	BondAmount         *big.Int
	UseExpressTicket   bool
	ValidatorPubkey    []byte
	ValidatorSignature []byte
	DepositDataRoot    common.Hash
}
