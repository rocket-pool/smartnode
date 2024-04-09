package minipool

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type minipoolCloseContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolCloseContextFactory) Create(args url.Values) (*minipoolCloseContext, error) {
	c := &minipoolCloseContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolCloseContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolCloseContext, types.BatchTxInfoData](
		router, "close", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolCloseContext struct {
	handler *MinipoolHandler

	minipoolAddresses []common.Address
}

func (c *minipoolCloseContext) Initialize() (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}

func (c *minipoolCloseContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolCloseContext) CheckState(node *node.Node, response *types.BatchTxInfoData) bool {
	return true
}

func (c *minipoolCloseContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mp.Common().Status.AddToQuery(mc)
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		mpv3.HasUserDistributed.AddToQuery(mc)
	}
}

func (c *minipoolCloseContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *types.BatchTxInfoData) (types.ResponseStatus, error) {
	return prepareMinipoolBatchTxData(c.handler.ctx, c.handler.serviceProvider, addresses, data, c.CreateTx, "close")
}

func (c *minipoolCloseContext) CreateTx(mp minipool.IMinipool, opts *bind.TransactOpts) (types.ResponseStatus, *eth.TransactionInfo, error) {
	mpCommon := mp.Common()
	minipoolAddress := mpCommon.Address
	mpv3, isMpv3 := minipool.GetMinipoolAsV3(mp)

	// If it's dissolved, just close it
	if mpCommon.Status.Formatted() == rptypes.MinipoolStatus_Dissolved {
		// Get gas estimate
		txInfo, err := mpCommon.Close(opts)
		if err != nil {
			return types.ResponseStatus_Error, nil, fmt.Errorf("error simulating close for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return types.ResponseStatus_Success, txInfo, nil
	}

	// Check if it's an upgraded Atlas-era minipool
	if isMpv3 {
		if mpv3.HasUserDistributed.Get() {
			// It's already been distributed so just finalize it
			txInfo, err := mpv3.Finalise(opts)
			if err != nil {
				return types.ResponseStatus_Error, nil, fmt.Errorf("error simulating finalise for minipool %s: %w", minipoolAddress.Hex(), err)
			}
			return types.ResponseStatus_Success, txInfo, nil
		}

		// Do a distribution, which will finalize it
		txInfo, err := mpv3.DistributeBalance(opts, false)
		if err != nil {
			return types.ResponseStatus_Error, nil, fmt.Errorf("error simulating distribute balance for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return types.ResponseStatus_Success, txInfo, nil
	}

	// Handle old minipools
	return types.ResponseStatus_InvalidChainState, nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", minipoolAddress.Hex(), mpCommon.Version)
}
