package auction

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func claimFromLot(c *cli.Context, lotIndex uint64) (*api.ClaimFromLotResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.ClaimFromLotResponse{}

	// Sync
	var addressBidAmount *big.Int

	// Create the bindings
	lot, err := auction.NewAuctionLot(rp, lotIndex)
	if err != nil {
		return nil, fmt.Errorf("error creating lot %d binding: %w", lotIndex, err)
	}

	// Get contract state
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}
	err = rp.Query(func(mc *batch.MultiCaller) error {
		lot.GetLotExists(mc)
		lot.GetLotAddressBidAmount(mc, &addressBidAmount, nodeAccount.Address)
		lot.GetLotIsCleared(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Check for validity
	response.DoesNotExist = !lot.Details.Exists
	response.NoBidFromAddress = (addressBidAmount.Cmp(big.NewInt(0)) == 0)
	response.NotCleared = !lot.Details.IsCleared
	response.CanClaim = !(response.DoesNotExist || response.NoBidFromAddress || response.NotCleared)

	// Get tx info
	if response.CanClaim {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
		txInfo, err := lot.ClaimBid(opts)
		if err != nil {
			return nil, fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		response.TxInfo = txInfo
	}

	return &response, nil
}
