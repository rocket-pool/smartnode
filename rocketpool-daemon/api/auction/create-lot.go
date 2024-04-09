package auction

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

type auctionCreateContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionCreateContextFactory) Create(args url.Values) (*auctionCreateContext, error) {
	c := &auctionCreateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *auctionCreateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionCreateContext, api.AuctionCreateLotData](
		router, "lots/create", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionCreateContext struct {
	handler *AuctionHandler
	rp      *rocketpool.RocketPool

	auctionMgr *auction.AuctionManager
	pSettings  *protocol.ProtocolDaoSettings
	networkMgr *network.NetworkManager
}

func (c *auctionCreateContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

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

func (c *auctionCreateContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.auctionMgr.RemainingRplBalance,
		c.networkMgr.RplPrice,
		c.pSettings.Auction.LotMinimumEthValue,
		c.pSettings.Auction.IsCreateLotEnabled,
	)
}

func (c *auctionCreateContext) PrepareData(data *api.AuctionCreateLotData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(c.pSettings.Auction.LotMinimumEthValue.Get(), eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, c.networkMgr.RplPrice.Raw())
	sufficientRemainingRplForLot := (c.auctionMgr.RemainingRplBalance.Get().Cmp(lotMinimumRplAmount) >= 0)

	// Check for validity
	data.InsufficientBalance = !sufficientRemainingRplForLot
	data.CreateLotDisabled = !c.pSettings.Auction.IsCreateLotEnabled.Get()
	data.CanCreate = !(data.InsufficientBalance || data.CreateLotDisabled)

	// Get tx info
	if data.CanCreate && opts != nil {
		txInfo, err := c.auctionMgr.CreateLot(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for CreateLot: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
