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

// ===============
// === Factory ===
// ===============

type oracleDaoExecuteProposalContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoExecuteProposalContextFactory) Create(args url.Values) (*oracleDaoExecuteProposalContext, error) {
	c := &oracleDaoExecuteProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.id),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoExecuteProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoExecuteProposalContext, api.OracleDaoExecuteProposalData](
		router, "proposal/execute", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoExecuteProposalContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id         uint64
	odaoMember *oracle.OracleDaoMember
	dpm        *proposals.DaoProposalManager
	prop       *proposals.OracleDaoProposal
}

func (c *oracleDaoExecuteProposalContext) Initialize() error {
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
	prop, err := c.dpm.CreateProposalFromID(c.id, nil)
	if err != nil {
		return fmt.Errorf("error creating proposal binding: %w", err)
	}
	var success bool
	c.prop, success = proposals.GetProposalAsOracle(prop)
	if !success {
		return fmt.Errorf("proposal %d is not an Oracle DAO proposal", c.id)
	}
	return nil
}

func (c *oracleDaoExecuteProposalContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.odaoMember.Exists,
		c.prop.State,
	)
}

func (c *oracleDaoExecuteProposalContext) PrepareData(data *api.OracleDaoExecuteProposalData, opts *bind.TransactOpts) error {
	// Verify oDAO status
	if !c.odaoMember.Exists.Get() {
		return errors.New("The node is not a member of the oracle DAO.")
	}

	// Check proposal details
	state := c.prop.State.Formatted()
	data.DoesNotExist = (c.id > c.dpm.ProposalCount.Formatted())
	data.InvalidState = !(state == rptypes.ProposalState_Pending || state == rptypes.ProposalState_Active)
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
