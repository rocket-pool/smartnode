package pdao

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
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
	ec      eth.IExecutionClient
	bc      beacon.IBeaconClient

	node             *node.Node
	propMgr          *proposals.ProposalManager
	totalDelegatedVP *big.Int
	blockNumber      uint64
}

func (c *protocolDaoGetStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	c.propMgr, err = proposals.NewProposalManager(c.handler.ctx, c.handler.logger.Logger, c.cfg, c.rp, c.bc)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager: %w", err)
	}
	// Get the latest block
	c.blockNumber, err = c.ec.BlockNumber(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block number: %w", err)
	}

	c.totalDelegatedVP, _, _, err = c.propMgr.GetArtifactsForVoting(uint32(c.blockNumber), nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting voting artifacts for node %s at block %d: %w", nodeAddress.Hex(), c.blockNumber, err)
	}

	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoGetStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		// Node
		c.node.Exists,
		c.node.IsVotingInitialized,
	)
}

func (c *protocolDaoGetStatusContext) PrepareData(data *api.ProtocolDAOStatusResponse, opts *bind.TransactOpts) (types.ResponseStatus, error) {

	data.BlockNumber = uint32(c.blockNumber)
	data.TotalDelegatedVp = c.totalDelegatedVP

	votingTree, err := c.propMgr.GetNetworkTree(uint32(c.blockNumber), nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting network tree")
	}
	data.SumVotingPower = votingTree.Nodes[0].Sum

	// Get the voting power and delegate at that block
	err = c.rp.Query(func(mc *batch.MultiCaller) error {
		c.node.GetVotingPowerAtBlock(mc, &data.VotingPower, data.BlockNumber)
		c.node.GetVotingDelegateAtBlock(mc, &data.OnchainVotingDelegate, data.BlockNumber)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting voting info for block %d: %w", c.blockNumber, err)
	}
	data.OnchainVotingDelegateFormatted = utils.GetFormattedAddress(c.ec, data.OnchainVotingDelegate)
	return types.ResponseStatus_Success, nil
}
