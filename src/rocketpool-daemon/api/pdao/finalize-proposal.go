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

type protocolDaoFinalizeProposalContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoFinalizeProposalContextFactory) Create(args url.Values) (*protocolDaoFinalizeProposalContext, error) {
	c := &protocolDaoFinalizeProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.proposalID),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoFinalizeProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoFinalizeProposalContext, api.ProtocolDaoFinalizeProposalData](
		router, "proposal/finalize", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoFinalizeProposalContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	proposalID uint64
	pdaoMgr    *protocol.ProtocolDaoManager
	proposal   *protocol.ProtocolDaoProposal
}

func (c *protocolDaoFinalizeProposalContext) Initialize() (types.ResponseStatus, error) {
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

func (c *protocolDaoFinalizeProposalContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.proposal.State,
		c.proposal.IsFinalized,
	)
}

func (c *protocolDaoFinalizeProposalContext) PrepareData(data *api.ProtocolDaoFinalizeProposalData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.DoesNotExist = (c.proposalID > c.pdaoMgr.ProposalCount.Formatted())
	data.InvalidState = (c.proposal.State.Formatted() != rptypes.ProtocolDaoProposalState_Vetoed)
	data.AlreadyFinalized = c.proposal.IsFinalized.Get()
	data.CanFinalize = !(data.DoesNotExist || data.InvalidState || data.AlreadyFinalized)

	// Get the tx
	if data.CanFinalize && opts != nil {
		txInfo, err := c.proposal.Finalize(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Finalize: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
