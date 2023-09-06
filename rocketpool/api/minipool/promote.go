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
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMinipoolPromoteDetailsForNode(c *cli.Context) (*api.GetMinipoolPromoteDetailsForNodeResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.GetMinipoolPromoteDetailsForNodeResponse{}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}
	oSettings, err := settings.NewOracleDaoSettings(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating oDAO settings binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		oSettings.GetPromotionScrubPeriod(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].GetMinipoolCommon()
		mpCommon.GetNodeAddress(mc)
		mpCommon.GetStatusTime(mc)
		mpv3, success := minipool.GetMinipoolAsV3(mps[i])
		if success {
			mpv3.GetVacant(mc)
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the time of the latest block
	latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("Can't get the latest block time: %w", err)
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

		// Validate minipool owner
		if mpCommon.Details.NodeAddress != nodeAccount.Address {
			return nil, fmt.Errorf("minipool %s does not belong to the node", mpCommon.Details.Address.Hex())
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
	return &response, nil
}

func promoteMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	return createBatchTxResponseForV3(c, minipoolAddresses, func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpv3.Promote(opts)
	}, "promote")
}
