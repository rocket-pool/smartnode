package security

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/dao/proposals"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
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

type securityProposalsContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposalsContextFactory) Create(args url.Values) (*securityProposalsContext, error) {
	c := &securityProposalsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *securityProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityProposalsContext, api.SecurityProposalsData](
		router, "proposals", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposalsContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address
	hasAddress  bool

	scMgr *security.SecurityCouncilManager
	dpm   *proposals.DaoProposalManager
}

func (c *securityProposalsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, c.hasAddress = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Protocol DAO manager binding: %w", err)
	}
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, pdaoMgr.Settings)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council manager binding: %w", err)
	}
	c.dpm, err = proposals.NewDaoProposalManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating DAO proposal manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *securityProposalsContext) GetState(mc *batch.MultiCaller) {
	c.dpm.ProposalCount.AddToQuery(mc)
}

func (c *securityProposalsContext) PrepareData(data *api.SecurityProposalsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	_, scProps, err := c.dpm.GetProposals(c.dpm.ProposalCount.Formatted(), true, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting proposals: %w", err)
	}

	// Get the basic details
	for _, scProp := range scProps {
		prop := api.SecurityProposalDetails{
			ID:              scProp.ID,
			ProposerAddress: scProp.ProposerAddress.Get(),
			Message:         scProp.Message.Get(),
			CreatedTime:     scProp.CreatedTime.Formatted(),
			StartTime:       scProp.StartTime.Formatted(),
			EndTime:         scProp.EndTime.Formatted(),
			ExpiryTime:      scProp.ExpiryTime.Formatted(),
			VotesRequired:   scProp.VotesRequired.Formatted(),
			VotesFor:        scProp.VotesFor.Formatted(),
			VotesAgainst:    scProp.VotesAgainst.Formatted(),
			IsCancelled:     scProp.IsCancelled.Get(),
			IsExecuted:      scProp.IsExecuted.Get(),
			Payload:         scProp.Payload.Get(),
		}
		prop.PayloadStr, err = scProp.GetPayloadAsString()
		if err != nil {
			prop.PayloadStr = fmt.Sprintf("<error decoding payload: %s>", err.Error())
		}
		data.Proposals = append(data.Proposals, prop)
	}

	// Get the node-specific details
	if c.hasAddress {
		err = c.rp.BatchQuery(len(data.Proposals), proposalBatchSize, func(mc *batch.MultiCaller, i int) error {
			odaoProp := scProps[i]
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
