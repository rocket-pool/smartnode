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
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		router, "proposal/defeat", f, f.handler.serviceProvider,
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
	challengeState types.ChallengeState
	pdaoMgr        *protocol.ProtocolDaoManager
	proposal       *protocol.ProtocolDaoProposal
}

func (c *protocolDaoDefeatProposalContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.proposal, err = protocol.NewProtocolDaoProposal(c.rp, c.proposalID)
	if err != nil {
		return fmt.Errorf("error creating proposal binding: %w", err)
	}
	return nil
}

func (c *protocolDaoDefeatProposalContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.proposal.State,
		c.proposal.CreatedTime,
		c.proposal.ChallengeWindow,
	)
	c.proposal.GetChallengeState(mc, &c.challengeState, c.index)
}

func (c *protocolDaoDefeatProposalContext) PrepareData(data *api.ProtocolDaoDefeatProposalData, opts *bind.TransactOpts) error {
	defeatStart := c.proposal.CreatedTime.Formatted().Add(c.proposal.ChallengeWindow.Formatted())
	data.StillInChallengeWindow = (time.Until(defeatStart) > 0)
	data.DoesNotExist = (c.proposalID > c.pdaoMgr.ProposalCount.Formatted())
	data.AlreadyDefeated = (c.proposal.State.Formatted() == types.ProtocolDaoProposalState_Defeated)
	data.InvalidChallengeState = (c.challengeState != types.ChallengeState_Challenged)
	data.CanDefeat = !(data.DoesNotExist || data.StillInChallengeWindow || data.AlreadyDefeated || data.InvalidChallengeState)

	// Get the tx
	if data.CanDefeat && opts != nil {
		txInfo, err := c.proposal.Defeat(c.index, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Defeat: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
