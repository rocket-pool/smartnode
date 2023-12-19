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

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type protocolDaoExecuteProposalContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoExecuteProposalContextFactory) Create(args url.Values) (*protocolDaoExecuteProposalContext, error) {
	c := &protocolDaoExecuteProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.proposalID),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoExecuteProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoExecuteProposalContext, api.ProtocolDaoExecuteProposalData](
		router, "proposal/execute", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoExecuteProposalContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	proposalID uint64
	pdaoMgr    *protocol.ProtocolDaoManager
	proposal   *protocol.ProtocolDaoProposal
}

func (c *protocolDaoExecuteProposalContext) Initialize() error {
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
	c.proposal, err = protocol.NewProtocolDaoProposal(c.rp, c.proposalID)
	if err != nil {
		return fmt.Errorf("error creating proposal binding: %w", err)
	}
	return nil
}

func (c *protocolDaoExecuteProposalContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.proposal.State,
	)
}

func (c *protocolDaoExecuteProposalContext) PrepareData(data *api.ProtocolDaoExecuteProposalData, opts *bind.TransactOpts) error {
	data.DoesNotExist = (c.proposalID > c.pdaoMgr.ProposalCount.Formatted())
	data.InvalidState = (c.proposal.State.Formatted() != types.ProtocolDaoProposalState_Succeeded)
	data.CanExecute = !(data.DoesNotExist || data.InvalidState)

	// Get the tx
	if data.CanExecute && opts != nil {
		txInfo, err := c.proposal.Execute(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for Execute: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
