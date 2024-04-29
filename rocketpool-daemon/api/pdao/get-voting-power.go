package pdao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoGetVotingPowerContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoGetVotingPowerContextFactory) Create(args url.Values) (*protocolDaoGetVotingPowerContext, error) {
	c := &protocolDaoGetVotingPowerContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoGetVotingPowerContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*protocolDaoGetVotingPowerContext, api.ProtocolDaoGetVotingPowerData](
		router, "get-voting-power", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoGetVotingPowerContext struct {
	handler *ProtocolDaoHandler
}

func (c *protocolDaoGetVotingPowerContext) PrepareData(data *api.ProtocolDaoGetVotingPowerData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()

	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}

	// Get the latest block
	blockNumber, err := ec.BlockNumber(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block number: %w", err)
	}
	data.BlockNumber = uint32(blockNumber)

	// Get the voting power and delegate at that block
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetVotingPowerAtBlock(mc, &data.VotingPower, data.BlockNumber)
		node.GetVotingDelegateAtBlock(mc, &data.OnchainVotingDelegate, data.BlockNumber)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting voting info for block %d: %w", blockNumber, err)
	}
	data.OnchainVotingDelegateFormatted = utils.GetFormattedAddress(ec, data.OnchainVotingDelegate)
	return types.ResponseStatus_Success, nil
}
