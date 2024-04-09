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

const (
	executeBatchSize int = 100
)

// ===============
// === Factory ===
// ===============

type protocolDaoExecuteProposalsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoExecuteProposalsContextFactory) Create(args url.Values) (*protocolDaoExecuteProposalsContext, error) {
	c := &protocolDaoExecuteProposalsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("ids", args, executeBatchSize, input.ValidatePositiveUint, &c.ids),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoExecuteProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoExecuteProposalsContext, types.DataBatch[api.ProtocolDaoExecuteProposalData]](
		router, "proposal/execute", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoExecuteProposalsContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	ids       []uint64
	pdaoMgr   *protocol.ProtocolDaoManager
	proposals []*protocol.ProtocolDaoProposal
}

func (c *protocolDaoExecuteProposalsContext) Initialize() (types.ResponseStatus, error) {
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
	c.proposals = make([]*protocol.ProtocolDaoProposal, len(c.ids))
	for i, id := range c.ids {
		c.proposals[i], err = protocol.NewProtocolDaoProposal(c.rp, id)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding for proposal %d: %w", id, err)
		}
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoExecuteProposalsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
	)
	for _, prop := range c.proposals {
		prop.State.AddToQuery(mc)
	}
}

func (c *protocolDaoExecuteProposalsContext) PrepareData(dataBatch *types.DataBatch[api.ProtocolDaoExecuteProposalData], opts *bind.TransactOpts) (types.ResponseStatus, error) {
	dataBatch.Batch = make([]api.ProtocolDaoExecuteProposalData, len(c.ids))
	for i, prop := range c.proposals {
		// Check proposal details
		data := &dataBatch.Batch[i]
		data.DoesNotExist = (c.ids[i] > c.pdaoMgr.ProposalCount.Formatted())
		data.InvalidState = (prop.State.Formatted() != rptypes.ProtocolDaoProposalState_Succeeded)
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
