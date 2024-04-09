package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoClaimBondsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoClaimBondsContextFactory) Create(body api.ProtocolDaoClaimBondsBody) (*protocolDaoClaimBondsContext, error) {
	c := &protocolDaoClaimBondsContext{
		handler: f.handler,
		body:    body,
	}
	// Validate the submission
	if body.Claims == nil {
		return nil, fmt.Errorf("claims must be set")
	}
	for _, claim := range body.Claims {
		if claim.Indices == nil {
			return nil, fmt.Errorf("indices for proposal %d must be set", claim.ProposalID)
		}
	}
	return c, nil
}

func (f *protocolDaoClaimBondsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStagePost[*protocolDaoClaimBondsContext, api.ProtocolDaoClaimBondsBody, types.DataBatch[api.ProtocolDaoClaimBondsData]](
		router, "claim-bonds", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoClaimBondsContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	body      api.ProtocolDaoClaimBondsBody
	pdaoMgr   *protocol.ProtocolDaoManager
	proposals []*protocol.ProtocolDaoProposal
}

func (c *protocolDaoClaimBondsContext) Initialize() (types.ResponseStatus, error) {
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
	c.proposals = make([]*protocol.ProtocolDaoProposal, len(c.body.Claims))
	for i, claim := range c.body.Claims {
		c.proposals[i], err = protocol.NewProtocolDaoProposal(c.rp, claim.ProposalID)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating proposal binding: %w", err)
		}
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoClaimBondsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
	)
	for _, prop := range c.proposals {
		eth.AddQueryablesToMulticall(mc,
			prop.State,
			prop.ProposerAddress,
		)
	}
}

func (c *protocolDaoClaimBondsContext) PrepareData(dataBatch *types.DataBatch[api.ProtocolDaoClaimBondsData], opts *bind.TransactOpts) (types.ResponseStatus, error) {
	dataBatch.Batch = make([]api.ProtocolDaoClaimBondsData, len(c.body.Claims))
	for i, claim := range c.body.Claims {
		proposal := c.proposals[i]
		data := &dataBatch.Batch[i]

		// Verify the proposal's details
		state := proposal.State.Formatted()
		proposer := proposal.ProposerAddress.Get()
		data.DoesNotExist = (claim.ProposalID > c.pdaoMgr.ProposalCount.Formatted())
		data.IsProposer = (proposer == c.nodeAddress)
		if data.IsProposer {
			data.InvalidState = (state < rptypes.ProtocolDaoProposalState_QuorumNotMet)
		} else {
			data.InvalidState = (state == rptypes.ProtocolDaoProposalState_Pending)
		}
		data.CanClaim = !(data.DoesNotExist || data.InvalidState)

		// Get the tx
		if data.CanClaim && opts != nil {
			if data.IsProposer {
				txInfo, err := proposal.ClaimBondProposer(claim.Indices, opts)
				if err != nil {
					return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ClaimBondProposer: %w", err)
				}
				data.TxInfo = txInfo
			} else {
				txInfo, err := proposal.ClaimBondChallenger(claim.Indices, opts)
				if err != nil {
					return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ClaimBondChallenger: %w", err)
				}
				data.TxInfo = txInfo
			}
		}
	}
	return types.ResponseStatus_Success, nil
}
