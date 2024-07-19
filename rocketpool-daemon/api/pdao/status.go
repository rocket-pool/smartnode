package pdao

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/proposals"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/voting"
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
	server.RegisterSingleStageRoute[*protocolDaoGetStatusContext, api.ProtocolDaoStatusResponse](
		router, "get-status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoGetStatusContext struct {
	handler  *ProtocolDaoHandler
	cfg      *config.SmartNodeConfig
	rp       *rocketpool.RocketPool
	ec       eth.IExecutionClient
	bc       beacon.IBeaconClient
	registry *contracts.RocketSignerRegistry

	node              *node.Node
	nodeAddress       common.Address
	propMgr           *proposals.ProposalManager
	blockNumber       uint64
	signallingAddress common.Address
	votingTree        *proposals.NetworkVotingTree
}

func (c *protocolDaoGetStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()
	network := c.cfg.GetNetworkResources().Network

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	c.registry = sp.GetRocketSignerRegistry()
	if c.registry == nil {
		return types.ResponseStatus_Error, fmt.Errorf("Network [%v] does not have a signer registry contract.", network)
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", c.nodeAddress.Hex(), err)
	}
	c.propMgr, err = proposals.NewProposalManager(c.handler.ctx, c.handler.logger.Logger, c.cfg, c.rp, c.bc)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager: %w", err)
	}
	c.blockNumber, err = c.ec.BlockNumber(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block number: %w", err)
	}
	c.votingTree, err = c.propMgr.GetNetworkTree(uint32(c.blockNumber), nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting network tree")
	}

	return types.ResponseStatus_Success, nil

}

func (c *protocolDaoGetStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		// Node
		c.node.Exists,
		c.node.IsVotingInitialized,
		c.node.IsRplLockingAllowed,
		c.node.RplLocked,
	)
	// Snapshot Registry
	c.registry.NodeToSigner(mc, &c.signallingAddress, c.node.Address)
}

func (c *protocolDaoGetStatusContext) PrepareData(data *api.ProtocolDaoStatusResponse, opts *bind.TransactOpts) (types.ResponseStatus, error) {

	var err error

	data.IsVotingInitialized = c.node.IsVotingInitialized.Get()

	if !data.IsVotingInitialized {
		data.TotalDelegatedVp = big.NewInt(0)
	} else {
		data.TotalDelegatedVp, _, _, err = c.propMgr.GetArtifactsForVoting(uint32(c.blockNumber), c.nodeAddress)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting voting artifacts for node %s at block %d: %w", c.nodeAddress.Hex(), c.blockNumber, err)
		}
	}

	data.IsNodeRegistered = c.node.Exists.Get()
	data.BlockNumber = uint32(c.blockNumber)
	data.AccountAddress = c.node.Address
	data.AccountAddressFormatted = utils.GetFormattedAddress(c.ec, data.AccountAddress)
	data.IsVotingInitialized = c.node.IsVotingInitialized.Get()
	data.SumVotingPower = c.votingTree.Nodes[0].Sum
	data.IsRPLLockingAllowed = c.node.IsRplLockingAllowed.Get()
	data.NodeRPLLocked = c.node.RplLocked.Get()
	data.VerifyEnabled = c.cfg.VerifyProposals.Value

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

	// Get the signalling address and active snapshot proposals
	emptyAddress := common.Address{}
	data.SignallingAddress = c.signallingAddress
	if data.SignallingAddress != emptyAddress {
		data.SignallingAddressFormatted = utils.GetFormattedAddress(c.ec, c.signallingAddress)
	}
	props, err := voting.GetSnapshotProposals(c.cfg, c.node.Address, c.signallingAddress, true)
	if err != nil {
		data.SnapshotResponse.Error = fmt.Sprintf("error getting snapshot proposals: %s", err.Error())
	} else {
		data.SnapshotResponse.ActiveSnapshotProposals = props
	}

	return types.ResponseStatus_Success, nil
}
