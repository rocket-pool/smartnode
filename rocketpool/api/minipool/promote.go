package minipool

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMinipoolPromoteDetailsForNode(c *cli.Context) (*api.GetMinipoolPromoteDetailsForNodeResponse, error) {
	var oSettings *settings.OracleDaoSettings

	return runMinipoolQuery(c, MinipoolQuerier[api.GetMinipoolPromoteDetailsForNodeResponse]{
		CreateBindings: func(rp *rocketpool.RocketPool) error {
			var err error
			oSettings, err = settings.NewOracleDaoSettings(rp)
			if err != nil {
				return fmt.Errorf("error creating oDAO settings binding: %w", err)
			}
			return nil
		},
		GetState: func(node *node.Node, mc *batch.MultiCaller) {
			oSettings.GetPromotionScrubPeriod(mc)
		},
		CheckState: nil,
		GetMinipoolDetails: func(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
			mpv3, success := minipool.GetMinipoolAsV3(mp)
			if success {
				mpv3.GetNodeAddress(mc)
				mpv3.GetStatusTime(mc)
				mpv3.GetVacant(mc)
			}
		},
		PrepareResponse: func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolPromoteDetailsForNodeResponse) error {
			// Get the time of the latest block
			latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				return fmt.Errorf("Can't get the latest block time: %w", err)
			}
			latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

			// Get the promotion details
			details := make([]api.MinipoolPromoteDetails, len(addresses))
			for i, mp := range mps {
				mpCommon := mp.GetMinipoolCommon()
				mpDetails := api.MinipoolPromoteDetails{
					Address:    mpCommon.Details.Address,
					CanPromote: false,
				}

				// Check its eligibility
				mpv3, success := minipool.GetMinipoolAsV3(mps[i])
				if success && mpv3.Details.IsVacant {
					creationTime := mpCommon.Details.StatusTime.Formatted()
					remainingTime := creationTime.Add(oSettings.Details.Minipools.ScrubPeriod.Formatted()).Sub(latestBlockTime)
					if remainingTime < 0 {
						mpDetails.CanPromote = true
					}
				}

				details[i] = mpDetails
			}

			response.Details = details
			return nil
		},
	})
}

func promoteMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForV3(c, minipoolAddresses, func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpv3.Promote(opts)
	}, "promote")
}
