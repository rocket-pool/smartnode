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
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		server.ValidateArg("addresses", args, input.ValidateAddresses, &c.addresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeKickMultiFromSecurityCouncilContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeKickMultiFromSecurityCouncilContext, api.ProtocolDaoProposeKickMultiFromSecurityCouncilData](
		router, "security/kick-multi", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeKickMultiFromSecurityCouncilContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	bc          beacon.Client
	nodeAddress common.Address

	addresses []common.Address
	node      *node.Node
	pdaoMgr   *protocol.ProtocolDaoManager
	members   []*security.SecurityCouncilMember
}

func (c *protocolDaoProposeKickMultiFromSecurityCouncilContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node binding: %w", err)
	}
	c.pdaoMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	c.members = make([]*security.SecurityCouncilMember, len(c.addresses))
	for i, address := range c.addresses {
		c.members[i], err = security.NewSecurityCouncilMember(c.rp, address)
		if err != nil {
			return fmt.Errorf("error creating security council member binding for node %s: %w", address.Hex(), err)
		}
	}
	return nil
}

func (c *protocolDaoProposeKickMultiFromSecurityCouncilContext) GetState(mc *batch.MultiCaller) {
	for _, member := range c.members {
		member.Exists.AddToQuery(mc)
	}
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeKickMultiFromSecurityCouncilContext) PrepareData(data *api.ProtocolDaoProposeKickMultiFromSecurityCouncilData, opts *bind.TransactOpts) error {
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
		blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
		message := "kick multiple members from the security council"
		txInfo, err := c.pdaoMgr.ProposeKickMultiFromSecurityCouncil(message, c.addresses, blockNumber, pollard, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeKickMultiFromSecurityCouncil: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
