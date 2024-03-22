package node

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/collateral"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeCheckCollateralContextFactory struct {
	handler *NodeHandler
}

func (f *nodeCheckCollateralContextFactory) Create(args url.Values) (*nodeCheckCollateralContext, error) {
	c := &nodeCheckCollateralContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeCheckCollateralContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeCheckCollateralContext, api.NodeCheckCollateralData](
		router, "check-collateral", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeCheckCollateralContext struct {
	handler *NodeHandler
}

func (c *nodeCheckCollateralContext) PrepareData(data *api.NodeCheckCollateralData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Check collateral
	collateral, err := collateral.CheckCollateral(rp, nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("error checking node collateral: %w", err)
	}
	data.EthMatched = collateral.EthMatched
	data.EthMatchedLimit = collateral.EthMatchedLimit
	data.PendingMatchAmount = collateral.PendingMatchAmount

	// Check if there's sufficient collateral including pending bond reductions
	remainingMatch := big.NewInt(0).Sub(data.EthMatchedLimit, data.EthMatched)
	remainingMatch.Sub(remainingMatch, data.PendingMatchAmount)
	data.InsufficientCollateral = (remainingMatch.Cmp(big.NewInt(0)) < 0)
	return nil
}
