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
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoCancelProposalContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoCancelProposalContextFactory) Create(args url.Values) (*oracleDaoCancelProposalContext, error) {
	c := &oracleDaoCancelProposalContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, input.ValidatePositiveUint, &c.id),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoCancelProposalContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoCancelProposalContext, api.OracleDaoCancelProposalData](
		router, "proposal/execute", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoCancelProposalContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id         uint64
	odaoMember *oracle.OracleDaoMember
	dpm        *proposals.DaoProposalManager
	prop       *proposals.OracleDaoProposal
}

func (c *oracleDaoCancelProposalContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered(c.handler.context)
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

func (c *oracleDaoCancelProposalContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.dpm.ProposalCount,
		c.odaoMember.Exists,
		c.prop.State,
		c.prop.ProposerAddress,
	)
}

func (c *oracleDaoCancelProposalContext) PrepareData(data *api.OracleDaoCancelProposalData, opts *bind.TransactOpts) error {
	// Verify oDAO status
	if !c.odaoMember.Exists.Get() {
		return errors.New("The node is not a member of the oracle DAO.")
	}

	// Check proposal details
	state := c.prop.State.Formatted()
	data.DoesNotExist = (c.id > c.dpm.ProposalCount.Formatted())
	data.InvalidState = !(state == rptypes.ProposalState_Pending || state == rptypes.ProposalState_Active)
	data.InvalidProposer = !(c.nodeAddress == c.prop.ProposerAddress.Get())
	data.CanCancel = !(data.DoesNotExist || data.InvalidState || data.InvalidProposer)

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