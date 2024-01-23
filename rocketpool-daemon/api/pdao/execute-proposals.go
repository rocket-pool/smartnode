package pdao

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
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
	server.RegisterSingleStageRoute[*protocolDaoExecuteProposalsContext, api.DataBatch[api.ProtocolDaoExecuteProposalData]](
		router, "proposal/execute", f, f.handler.serviceProvider,
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

func (c *protocolDaoExecuteProposalsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.proposals = make([]*protocol.ProtocolDaoProposal, len(c.ids))
	for i, id := range c.ids {
		c.proposals[i], err = protocol.NewProtocolDaoProposal(c.rp, id)
		if err != nil {
			return fmt.Errorf("error creating proposal binding for proposal %d: %w", id, err)
		}
	}
	return nil
}

func (c *protocolDaoExecuteProposalsContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
	)
	for _, prop := range c.proposals {
		prop.State.AddToQuery(mc)
	}
}

func (c *protocolDaoExecuteProposalsContext) PrepareData(dataBatch *api.DataBatch[api.ProtocolDaoExecuteProposalData], opts *bind.TransactOpts) error {
	dataBatch.Batch = make([]api.ProtocolDaoExecuteProposalData, len(c.ids))
	for i, prop := range c.proposals {
		// Check proposal details
		data := &dataBatch.Batch[i]
		data.DoesNotExist = (c.ids[i] > c.pdaoMgr.ProposalCount.Formatted())
		data.InvalidState = (prop.State.Formatted() != types.ProtocolDaoProposalState_Succeeded)
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
