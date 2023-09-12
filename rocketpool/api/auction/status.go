package auction

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Lot count details
type lotCountDetails struct {
	AddressHasBid   bool
	Cleared         bool
	HasRemainingRpl bool
	RplRecovered    bool
}

type auctionStatusHandler struct {
	auctionMgr    *auction.AuctionManager
	pSettings     *settings.ProtocolDaoSettings
	networkPrices *network.NetworkPrices
}

func (h *auctionStatusHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	h.auctionMgr, err = auction.NewAuctionManager(rp)
	if err != nil {
		return fmt.Errorf("error creating auction manager binding: %w", err)
	}
	h.pSettings, err = settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	h.networkPrices, err = network.NewNetworkPrices(rp)
	if err != nil {
		return fmt.Errorf("error creating network prices binding: %w", err)
	}
	return nil
}

func (h *auctionStatusHandler) GetState(nodeAddress common.Address, mc *batch.MultiCaller) {
	h.auctionMgr.GetTotalRPLBalance(mc)
	h.auctionMgr.GetAllottedRPLBalance(mc)
	h.auctionMgr.GetRemainingRPLBalance(mc)
	h.auctionMgr.GetLotCount(mc)
	h.pSettings.GetAuctionLotMinimumEthValue(mc)
	h.networkPrices.GetRplPrice(mc)
	h.pSettings.GetCreateAuctionLotEnabled(mc)
}

func (h *auctionStatusHandler) PrepareResponse(rp *rocketpool.RocketPool, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.AuctionStatusResponse) error {
	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(h.pSettings.Details.Auction.LotMinimumEthValue, eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, h.networkPrices.Details.RplPrice.RawValue)
	sufficientRemainingRplForLot := (h.auctionMgr.Details.RemainingRplBalance.Cmp(lotMinimumRplAmount) >= 0)

	// Get lot counts
	lotCountDetails, err := getAllLotCountDetails(rp, nodeAccount.Address, h.auctionMgr.Details.LotCount.Formatted())
	if err != nil {
		return fmt.Errorf("error getting auction lot count details: %w", err)
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
	response.TotalRPLBalance = h.auctionMgr.Details.TotalRplBalance
	response.AllottedRPLBalance = h.auctionMgr.Details.AllottedRplBalance
	response.RemainingRPLBalance = h.auctionMgr.Details.RemainingRplBalance
	response.CanCreateLot = sufficientRemainingRplForLot
	return nil
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
