package node

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeBurnContextFactory struct {
	handler *NodeHandler
}

func (f *nodeBurnContextFactory) Create(args url.Values) (*nodeBurnContext, error) {
	c := &nodeBurnContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidatePositiveWeiAmount, &c.amountWei),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeBurnContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeBurnContext, api.NodeBurnData](
		router, "burn", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeBurnContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	amountWei *big.Int
	balance   *big.Int
	reth      *tokens.TokenReth
}

func (c *nodeBurnContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.reth, err = tokens.NewTokenReth(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating reth binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *nodeBurnContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.reth.TotalCollateral,
	)
	c.reth.BalanceOf(mc, &c.balance, c.nodeAddress)
}

func (c *nodeBurnContext) PrepareData(data *api.NodeBurnData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Check for validity
	data.InsufficientBalance = (c.amountWei.Cmp(c.balance) > 0)
	data.InsufficientCollateral = (c.amountWei.Cmp(c.reth.TotalCollateral.Get()) > 0)
	data.CanBurn = !(data.InsufficientBalance || data.InsufficientCollateral)

	// Get tx info
	if data.CanBurn && opts != nil {
		txInfo, err := c.reth.Burn(c.amountWei, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Burn: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
