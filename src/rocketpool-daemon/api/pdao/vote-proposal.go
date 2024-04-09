package pdao

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/proposals"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

// ===============
// === Factory ===
// ===============

type protocolDaoVoteOnProposalContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoVoteOnProposalContextFactory) Create(args url.Values) (*protocolDaoVoteOnProposalContext, error) {
	c := &protocolDaoVoteOnProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.proposalID),
		server.ValidateArg("vote", args, utils.ValidateVoteDirection, &c.voteDirection),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoVoteOnProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoVoteOnProposalContext, api.ProtocolDaoVoteOnProposalData](
		router, "proposal/vote", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoVoteOnProposalContext struct {
	handler     *ProtocolDaoHandler
	cfg         *config.SmartNodeConfig
	rp          *rocketpool.RocketPool
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	proposalID      uint64
	voteDirection   rptypes.VoteDirection
	node            *node.Node
	existingVoteDir func() rptypes.VoteDirection
	pdaoMgr         *protocol.ProtocolDaoManager
	proposal        *protocol.ProtocolDaoProposal
	propMgr         *proposals.ProposalManager
}

func (c *protocolDaoVoteOnProposalContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.cfg = sp.GetConfig()
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	ctx := c.handler.ctx
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", c.nodeAddress.Hex(), err)
	}
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.proposal, err = protocol.NewProtocolDaoProposal(c.rp, c.proposalID)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
	}
	c.propMgr, err = proposals.NewProposalManager(ctx, nil, c.cfg, c.rp, c.bc)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoVoteOnProposalContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.proposal.State,
		c.proposal.TargetBlock,
	)
	c.existingVoteDir = c.proposal.GetAddressVoteDirection(mc, c.nodeAddress)
}

func (c *protocolDaoVoteOnProposalContext) PrepareData(data *api.ProtocolDaoVoteOnProposalData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	var err error
	targetBlock := c.proposal.TargetBlock.Formatted()
	totalDelegatedVP, nodeIndex, proof, err := c.propMgr.GetArtifactsForVoting(targetBlock, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting voting artifacts for node %s at block %d: %w", c.nodeAddress.Hex(), targetBlock, err)
	}

	data.VotingPower = totalDelegatedVP
	data.DoesNotExist = (c.proposalID > c.pdaoMgr.ProposalCount.Formatted())
	data.InvalidState = (c.proposal.State.Formatted() != rptypes.ProtocolDaoProposalState_ActivePhase1)
	data.AlreadyVoted = (c.existingVoteDir() != rptypes.VoteDirection_NoVote)
	data.InsufficientPower = (data.VotingPower.Cmp(common.Big0) == 0)
	data.CanVote = !(data.DoesNotExist || data.InvalidState || data.AlreadyVoted || data.InsufficientPower)

	// Get the tx
	if data.CanVote && opts != nil {
		txInfo, err := c.proposal.Vote(c.voteDirection, totalDelegatedVP, nodeIndex, proof, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Vote: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
