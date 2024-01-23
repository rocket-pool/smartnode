package security

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

const (
	executeBatchSize int = 100
)

// ===============
// === Factory ===
// ===============

type securityExecuteProposalsContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityExecuteProposalsContextFactory) Create(args url.Values) (*securityExecuteProposalsContext, error) {
	c := &securityExecuteProposalsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("ids", args, executeBatchSize, input.ValidatePositiveUint, &c.ids),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityExecuteProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityExecuteProposalsContext, api.DataBatch[api.SecurityExecuteProposalData]](
		router, "proposal/execute", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityExecuteProposalsContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	ids       []uint64
	dpm       *proposals.DaoProposalManager
	proposals []*proposals.SecurityCouncilProposal
}

func (c *securityExecuteProposalsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Bindings
	var err error
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	c.proposals = make([]*proposals.SecurityCouncilProposal, len(c.ids))
	for i, id := range c.ids {
		prop, err := c.dpm.CreateProposalFromID(id, nil)
		if err != nil {
			return fmt.Errorf("error creating proposal binding: %w", err)
		}
		var success bool
		c.proposals[i], success = proposals.GetProposalAsSecurity(prop)
		if !success {
			return fmt.Errorf("proposal %d is not a Security Council proposal", id)
		}
	}
	return nil
}

func (c *securityExecuteProposalsContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
	)
	for _, prop := range c.proposals {
		prop.State.AddToQuery(mc)
	}
}

func (c *securityExecuteProposalsContext) PrepareData(dataBatch *api.DataBatch[api.SecurityExecuteProposalData], opts *bind.TransactOpts) error {
	dataBatch.Batch = make([]api.SecurityExecuteProposalData, len(c.ids))
	for i, prop := range c.proposals {

		// Check proposal details
		data := &dataBatch.Batch[i]
		state := prop.State.Formatted()
		data.DoesNotExist = (c.ids[i] > c.dpm.ProposalCount.Formatted())
		data.InvalidState = !(state == types.ProposalState_Succeeded)
		data.CanExecute = !(data.DoesNotExist || data.InvalidState)

		// Get the tx
		if data.CanExecute && opts != nil {
			txInfo, err := prop.Execute(opts)
			if err != nil {
				return fmt.Errorf("error getting TX info for Execute: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return nil
}
