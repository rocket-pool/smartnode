package minipool

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type minipoolDissolveManager struct {
}

func (m *minipoolDissolveManager) CreateBindings(rp *rocketpool.RocketPool) error {
	return nil
}

func (m *minipoolDissolveManager) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (m *minipoolDissolveManager) CheckState(node *node.Node, response *api.MinipoolDissolveDetailsResponse) bool {
	return true
}

func (m *minipoolDissolveManager) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetNodeAddress(mc)
	mpCommon.GetStatus(mc)
}

func (m *minipoolDissolveManager) PrepareResponse(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, mps []minipool.Minipool, response *api.MinipoolDissolveDetailsResponse) error {
	details := make([]api.MinipoolDissolveDetails, len(mps))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		status := mpCommonDetails.Status.Formatted()
		mpDetails := api.MinipoolDissolveDetails{
			Address:       mpCommonDetails.Address,
			InvalidStatus: !(status == types.Initialized || status == types.Prelaunch),
		}
		mpDetails.CanDissolve = !mpDetails.InvalidStatus
		details[i] = mpDetails
	}

	response.Details = details
	return nil
}

func dissolveMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.Dissolve(opts)
	}, "dissolve")
}
