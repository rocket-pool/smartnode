package auction

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/auction"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type auctionStatusContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionStatusContextFactory) Create(args url.Values) (*auctionStatusContext, error) {
	c := &auctionStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *auctionStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionStatusContext, api.AuctionStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

// Lot count details
type lotCountDetails struct {
	AddressHasBid   bool
	Cleared         bool
	HasRemainingRpl bool
	RplRecovered    bool
}

type auctionStatusContext struct {
	handler     *AuctionHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	auctionMgr *auction.AuctionManager
	pSettings  *protocol.ProtocolDaoSettings
	networkMgr *network.NetworkManager
}

func (c *auctionStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.auctionMgr, err = auction.NewAuctionManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating auction manager binding: %w", err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating network prices binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *auctionStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.auctionMgr.TotalRplBalance,
		c.auctionMgr.AllottedRplBalance,
		c.auctionMgr.RemainingRplBalance,
		c.auctionMgr.LotCount,
		c.pSettings.Auction.LotMinimumEthValue,
		c.networkMgr.RplPrice,
		c.pSettings.Auction.IsCreateLotEnabled,
	)
}

func (c *auctionStatusContext) PrepareData(data *api.AuctionStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(c.pSettings.Auction.LotMinimumEthValue.Get(), eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, c.networkMgr.RplPrice.Raw())
	sufficientRemainingRplForLot := (c.auctionMgr.RemainingRplBalance.Get().Cmp(lotMinimumRplAmount) >= 0)

	// Get lot counts
	lotCountDetails, err := c.getAllLotCountDetails(c.auctionMgr.LotCount.Formatted())
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting auction lot count details: %w", err)
	}
	for _, details := range lotCountDetails {
		if details.AddressHasBid && details.Cleared {
			data.LotCounts.ClaimAvailable++
		}
		if !details.Cleared && details.HasRemainingRpl {
			data.LotCounts.BiddingAvailable++
		}
		if details.Cleared && details.HasRemainingRpl && !details.RplRecovered {
			data.LotCounts.RplRecoveryAvailable++
		}
	}

	// Set response details
	data.TotalRplBalance = c.auctionMgr.TotalRplBalance.Get()
	data.AllottedRplBalance = c.auctionMgr.AllottedRplBalance.Get()
	data.RemainingRplBalance = c.auctionMgr.RemainingRplBalance.Get()
	data.CanCreateLot = sufficientRemainingRplForLot
	return types.ResponseStatus_Success, nil
}

// Get all lot count details
func (c *auctionStatusContext) getAllLotCountDetails(lotCount uint64) ([]lotCountDetails, error) {
	details := make([]lotCountDetails, lotCount)
	lots := make([]*auction.AuctionLot, lotCount)
	addressBids := make([]*big.Int, lotCount)

	// Load details
	err := c.rp.BatchQuery(int(lotCount), int(lotCountDetailsBatchSize), func(mc *batch.MultiCaller, i int) error {
		lot, err := auction.NewAuctionLot(c.rp, uint64(i))
		if err != nil {
			return fmt.Errorf("error creating lot %d binding: %w", i, err)
		}
		lots[i] = lot

		lot.GetLotAddressBidAmount(mc, &addressBids[i], c.nodeAddress)
		eth.AddQueryablesToMulticall(mc,
			lot.IsCleared,
			lot.RemainingRplAmount,
			lot.RplRecovered,
		)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting lot count details: %w", err)
	}

	for i := 0; i < int(lotCount); i++ {
		details[i].AddressHasBid = (addressBids[i].Cmp(big.NewInt(0)) > 0)
		details[i].Cleared = lots[i].IsCleared.Get()
		details[i].HasRemainingRpl = (lots[i].RemainingRplAmount.Get().Cmp(big.NewInt(0)) > 0)
		details[i].RplRecovered = lots[i].RplRecovered.Get()
	}

	// Return
	return details, nil
}
