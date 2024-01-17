package odao

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
	server.RegisterSingleStageRoute[*oracleDaoExecuteProposalsContext, api.DataBatch[api.OracleDaoExecuteProposalsData]](
		router, "proposal/execute", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoExecuteProposalsContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	ids        []uint64
	odaoMember *oracle.OracleDaoMember
	dpm        *proposals.DaoProposalManager
	proposals  []*proposals.OracleDaoProposal
}

func (c *oracleDaoExecuteProposalsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.odaoMember, err = oracle.NewOracleDaoMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating oracle DAO member binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating proposal manager binding: %w", err)
	}
	c.proposals = make([]*proposals.OracleDaoProposal, len(c.ids))
	for i, id := range c.ids {
		prop, err := c.dpm.CreateProposalFromID(id, nil)
		if err != nil {
			return fmt.Errorf("error creating proposal binding: %w", err)
		}
		var success bool
		c.proposals[i], success = proposals.GetProposalAsOracle(prop)
		if !success {
			return fmt.Errorf("proposal %d is not an Oracle DAO proposal", id)
		}
	}
	return nil
}

func (c *oracleDaoExecuteProposalsContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.odaoMember.Exists,
	)
	for _, prop := range c.proposals {
		prop.State.AddToQuery(mc)
	}
}

func (c *oracleDaoExecuteProposalsContext) PrepareData(dataBatch *api.DataBatch[api.OracleDaoExecuteProposalsData], opts *bind.TransactOpts) error {
	// Verify oDAO status
	if !c.odaoMember.Exists.Get() {
		return errors.New("The node is not a member of the oracle DAO.")
	}

	dataBatch.Batch = make([]api.OracleDaoExecuteProposalsData, len(c.ids))
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
				return fmt.Errorf("error getting TX info for Execute: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return nil
}
