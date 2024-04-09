package odao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	proposalBatchSize int = 100
)

// ===============
// === Factory ===
// ===============

type oracleDaoProposalsContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoProposalsContextFactory) Create(args url.Values) (*oracleDaoProposalsContext, error) {
	c := &oracleDaoProposalsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposalsContext, api.OracleDaoProposalsData](
		router, "proposals", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoProposalsContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address
	hasAddress  bool

	odaoMgr *oracle.OracleDaoManager
	dpm     *proposals.DaoProposalManager
}

func (c *oracleDaoProposalsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, c.hasAddress = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoProposalsContext) GetState(mc *batch.MultiCaller) {
	c.dpm.ProposalCount.AddToQuery(mc)
}

func (c *oracleDaoProposalsContext) PrepareData(data *api.OracleDaoProposalsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	odaoProps, _, err := c.dpm.GetProposals(c.dpm.ProposalCount.Formatted(), true, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting proposals: %w", err)
	}

	// Get the basic details
	for _, odaoProp := range odaoProps {
		prop := api.OracleDaoProposalDetails{
			ID:              odaoProp.ID,
			ProposerAddress: odaoProp.ProposerAddress.Get(),
			Message:         odaoProp.Message.Get(),
			CreatedTime:     odaoProp.CreatedTime.Formatted(),
			StartTime:       odaoProp.StartTime.Formatted(),
			EndTime:         odaoProp.EndTime.Formatted(),
			ExpiryTime:      odaoProp.ExpiryTime.Formatted(),
			VotesRequired:   odaoProp.VotesRequired.Formatted(),
			VotesFor:        odaoProp.VotesFor.Formatted(),
			VotesAgainst:    odaoProp.VotesAgainst.Formatted(),
			IsCancelled:     odaoProp.IsCancelled.Get(),
			IsExecuted:      odaoProp.IsExecuted.Get(),
			Payload:         odaoProp.Payload.Get(),
		}
		prop.PayloadStr, err = odaoProp.GetPayloadAsString()
		if err != nil {
			prop.PayloadStr = fmt.Sprintf("<error decoding payload: %s>", err.Error())
		}
		data.Proposals = append(data.Proposals, prop)
	}

	// Get the node-specific details
	if c.hasAddress {
		err = c.rp.BatchQuery(len(data.Proposals), proposalBatchSize, func(mc *batch.MultiCaller, i int) error {
			odaoProp := odaoProps[i]
			odaoProp.GetMemberHasVoted(mc, &data.Proposals[i].MemberVoted, c.nodeAddress)
			odaoProp.GetMemberSupported(mc, &data.Proposals[i].MemberSupported, c.nodeAddress)
			return nil
		}, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting node vote status on proposals: %w", err)
		}
	}
	return types.ResponseStatus_Success, nil
}
