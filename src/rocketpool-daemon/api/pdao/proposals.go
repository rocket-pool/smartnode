package pdao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	nodeVoteBatchSize int = 500
)

// ===============
// === Factory ===
// ===============

type protocolDaoProposalsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposalsContextFactory) Create(args url.Values) (*protocolDaoProposalsContext, error) {
	c := &protocolDaoProposalsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposalsContext, api.ProtocolDaoProposalsData](
		router, "proposals", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposalsContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	node    *node.Node
	pdaoMgr *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposalsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node binding: %w", err)
	}
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoProposalsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.node.Exists,
	)
}

func (c *protocolDaoProposalsContext) PrepareData(data *api.ProtocolDaoProposalsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the proposals
	props, err := c.pdaoMgr.GetProposals(c.pdaoMgr.ProposalCount.Formatted(), true, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Protocol DAO proposals: %w", err)
	}

	// Convert them to API proposals
	returnProps := make([]api.ProtocolDaoProposalDetails, len(props))
	for i, prop := range props {
		returnProps[i] = api.ProtocolDaoProposalDetails{
			ID:                   prop.ID,
			ProposerAddress:      prop.ProposerAddress.Get(),
			TargetBlock:          prop.TargetBlock.Formatted(),
			Message:              prop.Message.Get(),
			ChallengeWindow:      prop.ChallengeWindow.Formatted(),
			CreatedTime:          prop.CreatedTime.Formatted(),
			VotingStartTime:      prop.VotingStartTime.Formatted(),
			Phase1EndTime:        prop.Phase1EndTime.Formatted(),
			Phase2EndTime:        prop.Phase2EndTime.Formatted(),
			ExpiryTime:           prop.ExpiryTime.Formatted(),
			VotingPowerRequired:  prop.VotingPowerRequired.Raw(),
			VotingPowerFor:       prop.VotingPowerFor.Raw(),
			VotingPowerAgainst:   prop.VotingPowerAgainst.Raw(),
			VotingPowerAbstained: prop.VotingPowerAbstained.Raw(),
			VotingPowerToVeto:    prop.VotingPowerToVeto.Raw(),
			IsDestroyed:          prop.IsDestroyed.Get(),
			IsFinalized:          prop.IsFinalized.Get(),
			IsExecuted:           prop.IsExecuted.Get(),
			IsVetoed:             prop.IsVetoed.Get(),
			Payload:              prop.Payload.Get(),
			State:                prop.State.Formatted(),
			ProposalBond:         prop.ProposalBond.Get(),
			ChallengeBond:        prop.ChallengeBond.Get(),
			DefeatIndex:          prop.DefeatIndex.Formatted(),
		}
		returnProps[i].PayloadStr, err = prop.GetProposalPayloadString()
		if err != nil {
			returnProps[i].PayloadStr = "<error parsing payload string>"
		}
	}

	// Get the node's vote directions
	nodeVoteDirs := make([]func() rptypes.VoteDirection, len(props))
	if c.node.Exists.Get() {
		err = c.rp.BatchQuery(len(props), nodeVoteBatchSize, func(mc *batch.MultiCaller, i int) error {
			nodeVoteDirs[i] = props[i].GetAddressVoteDirection(mc, c.nodeAddress)
			return nil
		}, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting node %s votes: %w", c.nodeAddress.Hex(), err)
		}
	}
	for i := range returnProps {
		returnProps[i].NodeVoteDirection = nodeVoteDirs[i]()
	}

	// Return
	data.Proposals = returnProps
	return types.ResponseStatus_Success, nil
}
