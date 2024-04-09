package pdao

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoDefeatProposalContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoDefeatProposalContextFactory) Create(args url.Values) (*protocolDaoDefeatProposalContext, error) {
	c := &protocolDaoDefeatProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.proposalID),
		server.ValidateArg("index", args, input.ValidatePositiveUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoDefeatProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoDefeatProposalContext, api.ProtocolDaoDefeatProposalData](
		router, "proposal/defeat", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoDefeatProposalContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	proposalID     uint64
	index          uint64
	challengeState func() rptypes.ChallengeState
	pdaoMgr        *protocol.ProtocolDaoManager
	proposal       *protocol.ProtocolDaoProposal
}

func (c *protocolDaoDefeatProposalContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.proposal, err = protocol.NewProtocolDaoProposal(c.rp, c.proposalID)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoDefeatProposalContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.proposal.State,
		c.proposal.CreatedTime,
		c.proposal.ChallengeWindow,
	)
	c.challengeState = c.proposal.GetChallengeState(mc, c.index)
}

func (c *protocolDaoDefeatProposalContext) PrepareData(data *api.ProtocolDaoDefeatProposalData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	defeatStart := c.proposal.CreatedTime.Formatted().Add(c.proposal.ChallengeWindow.Formatted())
	data.StillInChallengeWindow = (time.Until(defeatStart) > 0)
	data.DoesNotExist = (c.proposalID > c.pdaoMgr.ProposalCount.Formatted())
	data.AlreadyDefeated = (c.proposal.State.Formatted() == rptypes.ProtocolDaoProposalState_Destroyed)
	data.InvalidChallengeState = (c.challengeState() != rptypes.ChallengeState_Challenged)
	data.CanDefeat = !(data.DoesNotExist || data.StillInChallengeWindow || data.AlreadyDefeated || data.InvalidChallengeState)

	// Get the tx
	if data.CanDefeat && opts != nil {
		txInfo, err := c.proposal.Defeat(c.index, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Defeat: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
