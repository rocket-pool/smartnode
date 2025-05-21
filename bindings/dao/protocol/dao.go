package protocol

import (
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Get contracts
var rocketDAOProtocolLock sync.Mutex

func getRocketDAOProtocol(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolLock.Lock()
	defer rocketDAOProtocolLock.Unlock()
	return rp.GetContract("rocketDAOProtocol", opts)
}
