package minipool

import (
	"fmt"
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
	return createMinipoolQuery(c,
		nil,
		nil,
		nil,
		func(mc *batch.MultiCaller, mp minipool.Minipool) {
			mpCommon := mp.GetMinipoolCommon()
			mpCommon.GetNodeAddress(mc)
			mpCommon.GetNodeRefundBalance(mc)
		},
		func(rp *rocketpool.RocketPool, nodeAddress common.Address, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolRefundDetailsForNodeResponse) error {
			// Get the refund details
			details := make([]api.MinipoolRefundDetails, len(addresses))
			for i, mp := range mps {
				mpCommonDetails := mp.GetMinipoolCommon().Details

				// Validate minipool owner
				if mpCommonDetails.NodeAddress != nodeAddress {
					return fmt.Errorf("minipool %s does not belong to the node", mpCommonDetails.Address.Hex())
				}

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
	)
}

func refundMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForCommon(c, minipoolAddresses, func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpCommon.Refund(opts)
	}, "refund")
}
