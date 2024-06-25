package pdao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/proposals"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoGetStatusContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoGetStatusContextFactory) Create(args url.Values) (*protocolDaoGetStatusContext, error) {
	c := &protocolDaoGetStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoGetStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*protocolDaoGetStatusContext, api.ProtocolDAOStatusResponse](
		router, "get-status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoGetStatusContext struct {
	handler *ProtocolDaoHandler
	cfg     *config.SmartNodeConfig
	rp      *rocketpool.RocketPool
	bc      beacon.IBeaconClient

	propMgr *proposals.ProposalManager
}

func (c *protocolDaoGetStatusContext) PrepareData(data *api.ProtocolDAOStatusResponse, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	ec := sp.GetEthClient()
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	ctx := c.handler.ctx

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
	c.propMgr, err = proposals.NewProposalManager(ctx, c.handler.logger.Logger, c.cfg, c.rp, c.bc)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager: %w", err)
	}

	// Get the latest block
	blockNumber, err := ec.BlockNumber(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block number: %w", err)
	}
	data.BlockNumber = uint32(blockNumber)

	totalDelegatedVP, _, _, err := c.propMgr.GetArtifactsForVoting(uint32(blockNumber), nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting voting artifacts for node %s at block %d: %w", nodeAddress.Hex(), blockNumber, err)
	}
	data.TotalDelegatedVp = totalDelegatedVP

	votingTree, err := c.propMgr.GetNetworkTree(uint32(blockNumber), nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting network tree")
	}
	data.SumVotingPower = votingTree.Nodes[0].Sum

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
