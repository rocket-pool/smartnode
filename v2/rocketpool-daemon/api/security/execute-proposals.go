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
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
	server.RegisterSingleStageRoute[*securityExecuteProposalsContext, types.DataBatch[api.SecurityExecuteProposalData]](
		router, "proposal/execute", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *securityExecuteProposalsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	c.proposals = make([]*proposals.SecurityCouncilProposal, len(c.ids))
	for i, id := range c.ids {
		prop, err := c.dpm.CreateProposalFromID(id, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
		}
		var success bool
		c.proposals[i], success = proposals.GetProposalAsSecurity(prop)
		if !success {
			return types.ResponseStatus_InvalidChainState, fmt.Errorf("proposal %d is not a Security Council proposal", id)
		}
	}
	return types.ResponseStatus_Success, nil
}

func (c *securityExecuteProposalsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
	)
	for _, prop := range c.proposals {
		prop.State.AddToQuery(mc)
	}
}

func (c *securityExecuteProposalsContext) PrepareData(dataBatch *types.DataBatch[api.SecurityExecuteProposalData], opts *bind.TransactOpts) (types.ResponseStatus, error) {
	dataBatch.Batch = make([]api.SecurityExecuteProposalData, len(c.ids))
	for i, prop := range c.proposals {

		// Check proposal details
		data := &dataBatch.Batch[i]
		state := prop.State.Formatted()
		data.DoesNotExist = (c.ids[i] > c.dpm.ProposalCount.Formatted())
		data.InvalidState = !(state == rptypes.ProposalState_Succeeded)
		data.CanExecute = !(data.DoesNotExist || data.InvalidState)

		// Get the tx
		if data.CanExecute && opts != nil {
			txInfo, err := prop.Execute(opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Execute: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return types.ResponseStatus_Success, nil
}
