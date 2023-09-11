package minipool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type minipoolRefundManager struct {
}

func (m *minipoolRefundManager) CreateBindings(rp *rocketpool.RocketPool) error {
	return nil
}

func (m *minipoolRefundManager) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (m *minipoolRefundManager) CheckState(node *node.Node, response *api.GetMinipoolRefundDetailsForNodeResponse) bool {
	return true
}

func (m *minipoolRefundManager) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetNodeAddress(mc)
	mpCommon.GetNodeRefundBalance(mc)
}

func (m *minipoolRefundManager) PrepareResponse(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolRefundDetailsForNodeResponse) error {
	// Get the refund details
	details := make([]api.MinipoolRefundDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		mpDetails := api.MinipoolRefundDetails{
			Address:                   mpCommonDetails.Address,
			InsufficientRefundBalance: (mpCommonDetails.NodeRefundBalance.Cmp(big.NewInt(0)) == 0),
		}
		mpDetails.CanRefund = !mpDetails.InsufficientRefundBalance
		details[i] = mpDetails
	}

	response.Details = details
	return nil
}

func refundMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.Refund(opts)
	}, "refund")
}
