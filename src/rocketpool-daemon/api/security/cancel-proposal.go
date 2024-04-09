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
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
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

type securityCancelProposalContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityCancelProposalContextFactory) Create(args url.Values) (*securityCancelProposalContext, error) {
	c := &securityCancelProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.id),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityCancelProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityCancelProposalContext, api.SecurityCancelProposalData](
		router, "proposal/cancel", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityCancelProposalContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id       uint64
	scMember *security.SecurityCouncilMember
	dpm      *proposals.DaoProposalManager
	prop     *proposals.SecurityCouncilProposal
}

func (c *securityCancelProposalContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.scMember, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	prop, err := c.dpm.CreateProposalFromID(c.id, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
	}
	var success bool
	c.prop, success = proposals.GetProposalAsSecurity(prop)
	if !success {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("proposal %d is not a security council proposal", c.id)
	}
	return types.ResponseStatus_Error, nil
}

func (c *securityCancelProposalContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.scMember.Exists,
		c.prop.State,
		c.prop.ProposerAddress,
	)
}

func (c *securityCancelProposalContext) PrepareData(data *api.SecurityCancelProposalData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Check proposal details
	state := c.prop.State.Formatted()
	data.DoesNotExist = (c.id > c.dpm.ProposalCount.Formatted())
	data.InvalidState = !(state == rptypes.ProposalState_Pending || state == rptypes.ProposalState_Active)
	data.InvalidProposer = (c.nodeAddress != c.prop.ProposerAddress.Get())
	data.NotOnSecurityCouncil = !c.scMember.Exists.Get()
	data.CanCancel = !(data.DoesNotExist || data.InvalidState || data.InvalidProposer || data.NotOnSecurityCouncil)

	// Get the tx
	if data.CanCancel && opts != nil {
		txInfo, err := c.prop.Cancel(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Cancel: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
