package pdao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	nodeVoteBatchSize int = 500
)

// ===============
// === Factory ===
// ===============

type protocolDaoProposalsContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposalsContextFactory) Create(args url.Values) (*protocolDaoProposalsContext, error) {
	c := &protocolDaoProposalsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoProposalsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposalsContext, api.ProtocolDaoProposalsData](
		router, "proposals", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposalsContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	bc          beacon.Client
	nodeAddress common.Address

	id      string
	address common.Address
	node    *node.Node
	pdaoMgr *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposalsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Bindings
	var err error
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node binding: %w", err)
	}
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	return nil
}

func (c *protocolDaoProposalsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.ProposalCount,
		c.node.Exists,
	)
}

func (c *protocolDaoProposalsContext) PrepareData(data *api.ProtocolDaoProposalsData, opts *bind.TransactOpts) error {
	// Get the proposals
	props, err := c.pdaoMgr.GetProposals(c.pdaoMgr.ProposalCount.Formatted(), true, nil)
	if err != nil {
		return fmt.Errorf("error getting Protocol DAO proposals: %w", err)
	}

	// Convert them to API proposals
	returnProps := make([]api.ProtocolDaoProposalDetails, len(props))
	for i, prop := range props {
		returnProps[i] = api.ProtocolDaoProposalDetails{
			ID:                   prop.ID,
			ProposerAddress:      prop.ProposerAddress.Get(),
			TargetBlock:          prop.TargetBlock.Formatted(),
			Message:              prop.Message.Get(),
			ChallengeWindow:      prop.ChallengeWindow.Formatted(),
			CreatedTime:          prop.CreatedTime.Formatted(),
			VotingStartTime:      prop.VotingStartTime.Formatted(),
			Phase1EndTime:        prop.Phase1EndTime.Formatted(),
			Phase2EndTime:        prop.Phase2EndTime.Formatted(),
			ExpiryTime:           prop.ExpiryTime.Formatted(),
			VotingPowerRequired:  prop.VotingPowerRequired.Raw(),
			VotingPowerFor:       prop.VotingPowerFor.Raw(),
			VotingPowerAgainst:   prop.VotingPowerAgainst.Raw(),
			VotingPowerAbstained: prop.VotingPowerAbstained.Raw(),
			VotingPowerToVeto:    prop.VotingPowerToVeto.Raw(),
			IsDestroyed:          prop.IsDestroyed.Get(),
			IsFinalized:          prop.IsFinalized.Get(),
			IsExecuted:           prop.IsExecuted.Get(),
			IsVetoed:             prop.IsVetoed.Get(),
			Payload:              prop.Payload.Get(),
			State:                prop.State.Formatted(),
			ProposalBond:         prop.ProposalBond.Get(),
			ChallengeBond:        prop.ChallengeBond.Get(),
			DefeatIndex:          prop.DefeatIndex.Formatted(),
		}
		returnProps[i].PayloadStr, err = prop.GetProposalPayloadString()
		if err != nil {
			returnProps[i].PayloadStr = "<error parsing payload string>"
		}
	}

	// Get the node's vote directions
	nodeVoteDirs := make([]func() types.VoteDirection, len(props))
	if c.node.Exists.Get() {
		err = c.rp.BatchQuery(len(props), nodeVoteBatchSize, func(mc *batch.MultiCaller, i int) error {
			nodeVoteDirs[i] = props[i].GetAddressVoteDirection(mc, c.nodeAddress)
			return nil
		}, nil)
		if err != nil {
			return fmt.Errorf("error getting node %s votes: %w", c.nodeAddress.Hex(), err)
		}
	}
	for i := range returnProps {
		returnProps[i].NodeVoteDirection = nodeVoteDirs[i]()
	}

	// Return
	data.Proposals = returnProps
	return nil
}
