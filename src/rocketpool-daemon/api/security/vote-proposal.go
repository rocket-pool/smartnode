package security

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type securityVoteOnProposalContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityVoteOnProposalContextFactory) Create(args url.Values) (*securityVoteOnProposalContext, error) {
	c := &securityVoteOnProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.id),
		server.ValidateArg("support", args, input.ValidateBool, &c.support),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityVoteOnProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityVoteOnProposalContext, api.SecurityVoteOnProposalData](
		router, "proposal/vote", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityVoteOnProposalContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id       uint64
	support  bool
	hasVoted bool
	dpm      *proposals.DaoProposalManager
	prop     *proposals.SecurityCouncilProposal
	member   *security.SecurityCouncilMember
}

func (c *securityVoteOnProposalContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireOnSecurityCouncil(c.handler.context)
	if err != nil {
		return err
	}

	// Bindings
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	prop, err := c.dpm.CreateProposalFromID(c.id, nil)
	if err != nil {
		return fmt.Errorf("error creating proposal binding: %w", err)
	}
	var success bool
	c.prop, success = proposals.GetProposalAsSecurity(prop)
	if !success {
		return fmt.Errorf("proposal %d is not a security council proposal", c.id)
	}
	c.member, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating security council member binding: %w", err)
	}
	return nil
}

func (c *securityVoteOnProposalContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.prop.State,
		c.prop.CreatedTime,
		c.member.JoinedTime,
	)
	c.prop.GetMemberHasVoted(mc, &c.hasVoted, c.nodeAddress)
}

func (c *securityVoteOnProposalContext) PrepareData(data *api.SecurityVoteOnProposalData, opts *bind.TransactOpts) error {
	// Check proposal details
	state := c.prop.State.Formatted()
	data.DoesNotExist = (c.id > c.dpm.ProposalCount.Formatted())
	data.InvalidState = !(state == types.ProposalState_Active)
	data.AlreadyVoted = c.hasVoted
	data.JoinedAfterCreated = (c.prop.CreatedTime.Formatted().After(c.member.JoinedTime.Formatted()))
	data.CanVote = !(data.DoesNotExist || data.InvalidState || data.AlreadyVoted || data.JoinedAfterCreated)

	// Get the tx
	if data.CanVote && opts != nil {
		txInfo, err := c.prop.VoteOn(c.support, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for VoteOn: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
