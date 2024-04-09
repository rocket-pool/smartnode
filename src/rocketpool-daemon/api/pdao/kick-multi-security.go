package pdao

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	kickBatchSize int = 100
)

// ===============
// === Factory ===
// ===============

type protocolDaoProposeKickMultiFromSecurityCouncilContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeKickMultiFromSecurityCouncilContextFactory) Create(args url.Values) (*protocolDaoProposeKickMultiFromSecurityCouncilContext, error) {
	c := &protocolDaoProposeKickMultiFromSecurityCouncilContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, kickBatchSize, input.ValidateAddress, &c.addresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeKickMultiFromSecurityCouncilContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeKickMultiFromSecurityCouncilContext, api.ProtocolDaoProposeKickMultiFromSecurityCouncilData](
		router, "security/kick-multi", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeKickMultiFromSecurityCouncilContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	addresses []common.Address
	node      *node.Node
	pdaoMgr   *protocol.ProtocolDaoManager
	members   []*security.SecurityCouncilMember
}

func (c *protocolDaoProposeKickMultiFromSecurityCouncilContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.bc = sp.GetBeaconClient()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node binding: %w", err)
	}
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.members = make([]*security.SecurityCouncilMember, len(c.addresses))
	for i, address := range c.addresses {
		c.members[i], err = security.NewSecurityCouncilMember(c.rp, address)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding for node %s: %w", address.Hex(), err)
		}
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoProposeKickMultiFromSecurityCouncilContext) GetState(mc *batch.MultiCaller) {
	for _, member := range c.members {
		member.Exists.AddToQuery(mc)
	}
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeKickMultiFromSecurityCouncilContext) PrepareData(data *api.ProtocolDaoProposeKickMultiFromSecurityCouncilData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)

	for _, member := range c.members {
		if !member.Exists.Get() {
			data.NonexistingMembers = append(data.NonexistingMembers, member.Address)
		}
	}
	data.CanPropose = !(data.InsufficientRpl || len(data.NonexistingMembers) > 0)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(ctx, c.rp, c.cfg, c.bc)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating pollard for proposal creation: %w", err)
		}

		message := "kick multiple members from the security council"
		txInfo, err := c.pdaoMgr.ProposeKickMultiFromSecurityCouncil(message, c.addresses, blockNumber, pollard, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeKickMultiFromSecurityCouncil: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
