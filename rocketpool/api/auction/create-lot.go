package auction

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func createLot(c *cli.Context) (*api.CreateLotResponse, error) {
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
	response := api.CreateLotResponse{}

	// Create the bindings
	auctionMgr, err := auction.NewAuctionManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating auction manager binding: %w", err)
	}
	pSettings, err := settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	networkPrices, err := network.NewNetworkPrices(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating network prices binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		auctionMgr.GetRemainingRPLBalance(mc)
		pSettings.GetAuctionLotMinimumEthValue(mc)
		networkPrices.GetRplPrice(mc)
		pSettings.GetCreateAuctionLotEnabled(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(pSettings.Details.Auction.LotMinimumEthValue, eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, networkPrices.Details.RplPrice.RawValue)
	sufficientRemainingRplForLot := (auctionMgr.Details.RemainingRplBalance.Cmp(lotMinimumRplAmount) >= 0)

	// Check for validity
	response.InsufficientBalance = !sufficientRemainingRplForLot
	response.CreateLotDisabled = !pSettings.Details.Auction.IsCreateLotEnabled
	response.CanCreate = !(response.InsufficientBalance || response.CreateLotDisabled)

	// Get tx info
	if response.CanCreate {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
		txInfo, err := auctionMgr.CreateLot(opts)
		if err != nil {
			return nil, fmt.Errorf("error getting TX info for CreateLot: %w", err)
		}
		response.TxInfo = txInfo
	}

	return &response, nil
}
