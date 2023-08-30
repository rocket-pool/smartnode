package auction

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Settings
const (
	lotCountDetailsBatchSize uint64 = 500
)

// Lot count details
type lotCountDetails struct {
	AddressHasBid   bool
	Cleared         bool
	HasRemainingRpl bool
	RplRecovered    bool
}

func getStatus(c *cli.Context) (*api.AuctionStatusResponse, error) {

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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.AuctionStatusResponse{}

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
		auctionMgr.GetTotalRPLBalance(mc)
		auctionMgr.GetAllottedRPLBalance(mc)
		auctionMgr.GetRemainingRPLBalance(mc)
		auctionMgr.GetLotCount(mc)
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

	// Get lot counts
	lotCountDetails, err := getAllLotCountDetails(rp, nodeAccount.Address, auctionMgr.Details.LotCount.Formatted())
	if err != nil {
		return nil, fmt.Errorf("error getting auction lot count details: %w", err)
	}
	for _, details := range lotCountDetails {
		if details.AddressHasBid && details.Cleared {
			response.LotCounts.ClaimAvailable++
		}
		if !details.Cleared && details.HasRemainingRpl {
			response.LotCounts.BiddingAvailable++
		}
		if details.Cleared && details.HasRemainingRpl && !details.RplRecovered {
			response.LotCounts.RPLRecoveryAvailable++
		}
	}

	// Set response details
	response.TotalRPLBalance = auctionMgr.Details.TotalRplBalance
	response.AllottedRPLBalance = auctionMgr.Details.AllottedRplBalance
	response.RemainingRPLBalance = auctionMgr.Details.RemainingRplBalance
	response.CanCreateLot = sufficientRemainingRplForLot

	// Return response
	return &response, nil

}

// Get all lot count details
func getAllLotCountDetails(rp *rocketpool.RocketPool, bidderAddress common.Address, lotCount uint64) ([]lotCountDetails, error) {

	details := make([]lotCountDetails, lotCount)
	lots := make([]*auction.AuctionLot, lotCount)
	addressBids := make([]*big.Int, lotCount)

	// Load details
	err := rp.BatchQuery(int(lotCount), int(lotCountDetailsBatchSize), func(mc *batch.MultiCaller, i int) error {
		lot, err := auction.NewAuctionLot(rp, uint64(i))
		if err != nil {
			return fmt.Errorf("error creating lot %d binding: %w", i, err)
		}
		lots[i] = lot

		lot.GetLotAddressBidAmount(mc, &addressBids[i], bidderAddress)
		lot.GetLotIsCleared(mc)
		lot.GetLotRemainingRplAmount(mc)
		lot.GetLotRplRecovered(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting lot count details: %w", err)
	}

	for i := 0; i < int(lotCount); i++ {
		details[i].AddressHasBid = (addressBids[i].Cmp(big.NewInt(0)) > 0)
		details[i].Cleared = lots[i].Details.IsCleared
		details[i].HasRemainingRpl = (lots[i].Details.RemainingRplAmount.Cmp(big.NewInt(0)) > 0)
		details[i].RplRecovered = lots[i].Details.RplRecovered
	}

	// Return
	return details, nil
}
