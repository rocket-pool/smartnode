package odao

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
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

type oracleDaoVoteContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoVoteContextFactory) Create(args url.Values) (*oracleDaoVoteContext, error) {
	c := &oracleDaoVoteContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidateUint, &c.id),
		server.ValidateArg("support", args, input.ValidateBool, &c.support),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoVoteContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoVoteContext, api.OracleDaoVoteOnProposalData](
		router, "proposal/vote", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoVoteContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id         uint64
	support    bool
	odaoMember *oracle.OracleDaoMember
	dpm        *proposals.DaoProposalManager
	prop       *proposals.OracleDaoProposal
	hasVoted   bool
}

func (c *oracleDaoVoteContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireOnOracleDao(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.odaoMember, err = oracle.NewOracleDaoMember(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oracle DAO member binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager binding: %w", err)
	}
	prop, err := c.dpm.CreateProposalFromID(c.id, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
	}
	var success bool
	c.prop, success = proposals.GetProposalAsOracle(prop)
	if !success {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("proposal %d is not an Oracle DAO proposal", c.id)
	}
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoVoteContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.prop.State,
		c.odaoMember.JoinedTime,
		c.prop.CreatedTime,
	)
	c.prop.GetMemberHasVoted(mc, &c.hasVoted, c.nodeAddress)
}

func (c *oracleDaoVoteContext) PrepareData(data *api.OracleDaoVoteOnProposalData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.DoesNotExist = (c.prop.ID > c.dpm.ProposalCount.Formatted())
	data.InvalidState = (c.prop.State.Formatted() != rptypes.ProposalState_Active)
	data.AlreadyVoted = c.hasVoted
	data.JoinedAfterCreated = (c.odaoMember.JoinedTime.Formatted().Sub(c.prop.CreatedTime.Formatted()) >= 0)
	data.CanVote = !(data.DoesNotExist || data.InvalidState || data.JoinedAfterCreated || data.AlreadyVoted)

	// Get the tx
	if data.CanVote && opts != nil {
		txInfo, err := c.prop.VoteOn(c.support, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for VoteOn: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
