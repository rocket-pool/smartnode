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
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type securityExecuteProposalContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityExecuteProposalContextFactory) Create(vars map[string]string) (*securityExecuteProposalContext, error) {
	c := &securityExecuteProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", vars, input.ValidatePositiveUint, &c.id),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityExecuteProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityExecuteProposalContext, api.SecurityExecuteProposalData](
		router, "proposal/execute", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityExecuteProposalContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id   uint64
	dpm  *proposals.DaoProposalManager
	prop *proposals.SecurityCouncilProposal
}

func (c *securityExecuteProposalContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Bindings
	var err error
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
		return fmt.Errorf("proposal %d is not an security council proposal", c.id)
	}
	return nil
}

func (c *securityExecuteProposalContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.prop.State,
	)
}

func (c *securityExecuteProposalContext) PrepareData(data *api.SecurityExecuteProposalData, opts *bind.TransactOpts) error {
	// Check proposal details
	state := c.prop.State.Formatted()
	data.DoesNotExist = (c.id > c.dpm.ProposalCount.Formatted())
	data.InvalidState = !(state == types.ProposalState_Succeeded)
	data.CanExecute = !(data.DoesNotExist || data.InvalidState)

	// Get the tx
	if data.CanExecute && opts != nil {
		txInfo, err := c.prop.Execute(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Execute: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
