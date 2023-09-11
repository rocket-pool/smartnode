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

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type minipoolPromoteManager struct {
	oSettings *settings.OracleDaoSettings
}

func (m *minipoolPromoteManager) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	m.oSettings, err = settings.NewOracleDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error creating oDAO settings binding: %w", err)
	}
	return nil
}

func (m *minipoolPromoteManager) GetState(node *node.Node, mc *batch.MultiCaller) {
	m.oSettings.GetPromotionScrubPeriod(mc)
}

func (m *minipoolPromoteManager) CheckState(node *node.Node, response *api.GetMinipoolPromoteDetailsForNodeResponse) bool {
	return true
}

func (m *minipoolPromoteManager) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		mpv3.GetNodeAddress(mc)
		mpv3.GetStatusTime(mc)
		mpv3.GetVacant(mc)
	}
}

func (m *minipoolPromoteManager) PrepareResponse(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolPromoteDetailsForNodeResponse) error {
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
			remainingTime := creationTime.Add(m.oSettings.Details.Minipools.ScrubPeriod.Formatted()).Sub(latestBlockTime)
			if remainingTime < 0 {
				mpDetails.CanPromote = true
			}
		}

		details[i] = mpDetails
	}

	response.Details = details
	return nil
}

func promoteMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForV3(c, minipoolAddresses, func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpv3.Promote(opts)
	}, "promote")
}
