package wallet

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/core"
)

// NYI
type Wallet interface {
	Load() error
	GetAddress() error
	SignTx(*core.TransactionInfo) (*types.Transaction, error)
	SignMessage(to common.Address, message []byte) (*types.Transaction, error)
}
