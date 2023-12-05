package pdao

import (
	"errors"
	"fmt"

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

type protocolDaoClaimBondsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoClaimBondsContextFactory) Create(vars map[string]string) (*protocolDaoClaimBondsContext, error) {
	c := &protocolDaoClaimBondsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("proposal-id", vars, input.ValidatePositiveUint, &c.proposalID),
		server.ValidateArg("indices", vars, input.ValidatePositiveUints, &c.indices),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoClaimBondsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoClaimBondsContext, api.ProtocolDaoClaimBondsData](
		router, "claim-bonds", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoClaimBondsContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	proposalID uint64
	indices    []uint64
	pdaoMgr    *protocol.ProtocolDaoManager
	proposal   *protocol.ProtocolDaoProposal
}

func (c *protocolDaoClaimBondsContext) Initialize() error {
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

func (c *protocolDaoClaimBondsContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.proposal.State,
		c.proposal.ProposerAddress,
	)
}

func (c *protocolDaoClaimBondsContext) PrepareData(data *api.ProtocolDaoClaimBondsData, opts *bind.TransactOpts) error {
	// Verify the proposal's details
	state := c.proposal.State.Formatted()
	proposer := c.proposal.ProposerAddress.Get()
	data.DoesNotExist = (c.proposalID > c.pdaoMgr.ProposalCount.Formatted())
	data.IsProposer = (proposer == c.nodeAddress)
	if data.IsProposer {
		data.InvalidState = (state == types.ProtocolDaoProposalState_Defeated || state < types.ProtocolDaoProposalState_QuorumNotMet)
	} else {
		data.InvalidState = (state == types.ProtocolDaoProposalState_Pending)
	}
	data.CanClaim = !(data.DoesNotExist || data.InvalidState)

	// Get the tx
	if data.CanClaim && opts != nil {
		if data.IsProposer {
			txInfo, err := c.proposal.ClaimBondProposer(c.indices, opts)
			if err != nil {
				return fmt.Errorf("error getting TX info for ClaimBondProposer: %w", err)
			}
			data.TxInfo = txInfo
		} else {
			txInfo, err := c.proposal.ClaimBondChallenger(c.indices, opts)
			if err != nil {
				return fmt.Errorf("error getting TX info for ClaimBondChallenger: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return nil
}
