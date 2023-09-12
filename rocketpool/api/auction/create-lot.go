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

type auctionCreateHandler struct {
	auctionMgr    *auction.AuctionManager
	pSettings     *settings.ProtocolDaoSettings
	networkPrices *network.NetworkPrices
}

func (h *auctionCreateHandler) CreateBindings(rp *rocketpool.RocketPool) error {
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

func (h *auctionCreateHandler) GetState(nodeAddress common.Address, mc *batch.MultiCaller) {
	h.auctionMgr.GetRemainingRPLBalance(mc)
	h.pSettings.GetAuctionLotMinimumEthValue(mc)
	h.networkPrices.GetRplPrice(mc)
	h.pSettings.GetCreateAuctionLotEnabled(mc)
}

func (h *auctionCreateHandler) PrepareResponse(rp *rocketpool.RocketPool, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.CreateLotResponse) error {
	// Check the balance requirement
	lotMinimumRplAmount := big.NewInt(0).Mul(h.pSettings.Details.Auction.LotMinimumEthValue, eth.EthToWei(1))
	lotMinimumRplAmount.Quo(lotMinimumRplAmount, h.networkPrices.Details.RplPrice.RawValue)
	sufficientRemainingRplForLot := (h.auctionMgr.Details.RemainingRplBalance.Cmp(lotMinimumRplAmount) >= 0)

	// Check for validity
	response.InsufficientBalance = !sufficientRemainingRplForLot
	response.CreateLotDisabled = !h.pSettings.Details.Auction.IsCreateLotEnabled
	response.CanCreate = !(response.InsufficientBalance || response.CreateLotDisabled)

	// Get tx info
	if response.CanCreate {
		txInfo, err := h.auctionMgr.CreateLot(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for CreateLot: %w", err)
		}
		response.TxInfo = txInfo
	}
	return nil
}
