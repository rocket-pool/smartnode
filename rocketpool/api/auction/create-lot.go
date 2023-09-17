package auction

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type auctionCreateContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionCreateContextFactory) Create(vars map[string]string) (*auctionCreateContext, error) {
	c := &auctionCreateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *auctionCreateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionCreateContext, api.CreateLotData](
		router, "create-lot", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionCreateContext struct {
	handler *AuctionHandler
	rp      *rocketpool.RocketPool

	auctionMgr    *auction.AuctionManager
	pSettings     *settings.ProtocolDaoSettings
	networkPrices *network.NetworkPrices
}

func (c *auctionCreateContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.auctionMgr, err = auction.NewAuctionManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating auction manager binding: %w", err)
	}
	c.pSettings, err = settings.NewProtocolDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	c.networkPrices, err = network.NewNetworkPrices(c.rp)
	if err != nil {
		return fmt.Errorf("error creating network prices binding: %w", err)
	}
	return nil
}

func (c *auctionCreateContext) GetState(mc *batch.MultiCaller) {
	c.auctionMgr.GetRemainingRPLBalance(mc)
	c.pSettings.GetAuctionLotMinimumEthValue(mc)
	c.networkPrices.GetRplPrice(mc)
	c.pSettings.GetCreateAuctionLotEnabled(mc)
}

func (c *auctionCreateContext) PrepareData(data *api.CreateLotData, opts *bind.TransactOpts) error {
	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(c.pSettings.Details.Auction.LotMinimumEthValue, eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, c.networkPrices.Details.RplPrice.RawValue)
	sufficientRemainingRplForLot := (c.auctionMgr.Details.RemainingRplBalance.Cmp(lotMinimumRplAmount) >= 0)

	// Check for validity
	data.InsufficientBalance = !sufficientRemainingRplForLot
	data.CreateLotDisabled = !c.pSettings.Details.Auction.IsCreateLotEnabled
	data.CanCreate = !(data.InsufficientBalance || data.CreateLotDisabled)

	// Get tx info
	if data.CanCreate && opts != nil {
		txInfo, err := c.auctionMgr.CreateLot(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for CreateLot: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
