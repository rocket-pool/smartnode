package fees

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type FeeRecipientManager interface {
	StoreFeeRecipientFile(rp *rocketpool.RocketPool, nodeAddress common.Address) (common.Address, error)
}
