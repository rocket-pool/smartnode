package voting

import (
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

/*
func Propose(rp *rocketpool.RocketPool, message string, blockNumber uint32, treeNodes []VotingTreeNode) (common.Hash, error) {
	getRocketDAOProtocolProposals
}

func ProposeSettingUint()
*/
// Get contracts
var rocketDAOProtocolProposalsInterface sync.Mutex

func getRocketDAOProtocolProposalsInterface(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProtocolProposalsInterface.Lock()
	defer rocketDAOProtocolProposalsInterface.Unlock()
	return rp.GetContract("rocketDAOProtocolProposalsInterface", opts)
}
