package minipool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMinipoolRefundDetailsForNode(c *cli.Context) (*api.GetMinipoolRefundDetailsForNodeResponse, error) {
	return runMinipoolQuery(c, MinipoolQuerier[api.GetMinipoolRefundDetailsForNodeResponse]{
		CreateBindings: nil,
		GetState:       nil,
		CheckState:     nil,
		GetMinipoolDetails: func(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
			mpCommon := mp.GetMinipoolCommon()
			mpCommon.GetNodeAddress(mc)
			mpCommon.GetNodeRefundBalance(mc)
		},
		PrepareResponse: func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolRefundDetailsForNodeResponse) error {
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
		},
	})
}

func refundMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.Refund(opts)
	}, "refund")
}
