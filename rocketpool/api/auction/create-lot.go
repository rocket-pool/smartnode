package auction

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type auctionCreateHandler struct {
	auctionMgr    *auction.AuctionManager
	pSettings     *settings.ProtocolDaoSettings
	networkPrices *network.NetworkPrices
}

func NewAuctionCreateHandler(vars map[string]string) (*auctionCreateHandler, error) {
	h := &auctionCreateHandler{}
	return h, nil
}

func (h *auctionCreateHandler) CreateBindings(ctx *callContext) error {
	var err error
	rp := ctx.rp

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

func (h *auctionCreateHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	h.auctionMgr.GetRemainingRPLBalance(mc)
	h.pSettings.GetAuctionLotMinimumEthValue(mc)
	h.networkPrices.GetRplPrice(mc)
	h.pSettings.GetCreateAuctionLotEnabled(mc)
}

func (h *auctionCreateHandler) PrepareData(ctx *callContext, data *api.CreateLotData) error {
	opts := ctx.opts

	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(h.pSettings.Details.Auction.LotMinimumEthValue, eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, h.networkPrices.Details.RplPrice.RawValue)
	sufficientRemainingRplForLot := (h.auctionMgr.Details.RemainingRplBalance.Cmp(lotMinimumRplAmount) >= 0)

	// Check for validity
	data.InsufficientBalance = !sufficientRemainingRplForLot
	data.CreateLotDisabled = !h.pSettings.Details.Auction.IsCreateLotEnabled
	data.CanCreate = !(data.InsufficientBalance || data.CreateLotDisabled)

	// Get tx info
	if data.CanCreate && opts != nil {
		txInfo, err := h.auctionMgr.CreateLot(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for CreateLot: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
