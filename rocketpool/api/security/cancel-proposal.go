package security

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type securityCancelProposalContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityCancelProposalContextFactory) Create(vars map[string]string) (*securityCancelProposalContext, error) {
	c := &securityCancelProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", vars, input.ValidatePositiveUint, &c.id),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityCancelProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityCancelProposalContext, api.SecurityCancelProposalData](
		router, "proposal/cancel", f, f.handler.serviceProvider,
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

func (c *securityCancelProposalContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.scMember, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating security council member binding: %w", err)
	}
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
	return nil
}

func (c *securityCancelProposalContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.scMember.Exists,
		c.prop.State,
		c.prop.ProposerAddress,
	)
}

func (c *securityCancelProposalContext) PrepareData(data *api.SecurityCancelProposalData, opts *bind.TransactOpts) error {
	// Check proposal details
	state := c.prop.State.Formatted()
	data.DoesNotExist = (c.id > c.dpm.ProposalCount.Formatted())
	data.InvalidState = !(state == types.ProposalState_Pending || state == types.ProposalState_Active)
	data.InvalidProposer = (c.nodeAddress != c.prop.ProposerAddress.Get())
	data.NotOnSecurityCouncil = !c.scMember.Exists.Get()
	data.CanCancel = !(data.DoesNotExist || data.InvalidState || data.InvalidProposer || data.NotOnSecurityCouncil)

	// Get the tx
	if data.CanCancel && opts != nil {
		txInfo, err := c.prop.Cancel(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Cancel: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
