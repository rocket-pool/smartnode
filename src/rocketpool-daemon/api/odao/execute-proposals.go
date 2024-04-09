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

type oracleDaoExecuteProposalsContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoExecuteProposalsContextFactory) Create(args url.Values) (*oracleDaoExecuteProposalsContext, error) {
	c := &oracleDaoExecuteProposalsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("ids", args, executeBatchSize, input.ValidatePositiveUint, &c.ids),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoExecuteProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoExecuteProposalsContext, types.DataBatch[api.OracleDaoExecuteProposalData]](
		router, "proposal/execute", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoExecuteProposalsContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	ids       []uint64
	dpm       *proposals.DaoProposalManager
	proposals []*proposals.OracleDaoProposal
}

func (c *oracleDaoExecuteProposalsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating proposal manager binding: %w", err)
	}
	c.proposals = make([]*proposals.OracleDaoProposal, len(c.ids))
	for i, id := range c.ids {
		prop, err := c.dpm.CreateProposalFromID(id, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
		}
		var success bool
		c.proposals[i], success = proposals.GetProposalAsOracle(prop)
		if !success {
			return types.ResponseStatus_InvalidChainState, fmt.Errorf("proposal %d is not an Oracle DAO proposal", id)
		}
	}
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoExecuteProposalsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
	)
	for _, prop := range c.proposals {
		prop.State.AddToQuery(mc)
	}
}

func (c *oracleDaoExecuteProposalsContext) PrepareData(dataBatch *types.DataBatch[api.OracleDaoExecuteProposalData], opts *bind.TransactOpts) (types.ResponseStatus, error) {
	dataBatch.Batch = make([]api.OracleDaoExecuteProposalData, len(c.ids))
	for i, prop := range c.proposals {

		// Check proposal details
		data := &dataBatch.Batch[i]
		state := prop.State.Formatted()
		data.DoesNotExist = (c.ids[i] > c.dpm.ProposalCount.Formatted())
		data.InvalidState = !(state == rptypes.ProposalState_Pending || state == rptypes.ProposalState_Active)
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
