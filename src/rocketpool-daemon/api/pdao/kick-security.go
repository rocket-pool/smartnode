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
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type protocolDaoProposeKickFromSecurityCouncilContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeKickFromSecurityCouncilContextFactory) Create(args url.Values) (*protocolDaoProposeKickFromSecurityCouncilContext, error) {
	c := &protocolDaoProposeKickFromSecurityCouncilContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeKickFromSecurityCouncilContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeKickFromSecurityCouncilContext, api.ProtocolDaoProposeKickFromSecurityCouncilData](
		router, "security/kick", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeKickFromSecurityCouncilContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	bc          beacon.Client
	nodeAddress common.Address

	address common.Address
	node    *node.Node
	pdaoMgr *protocol.ProtocolDaoManager
	member  *security.SecurityCouncilMember
}

func (c *protocolDaoProposeKickFromSecurityCouncilContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

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
	c.member, err = security.NewSecurityCouncilMember(c.rp, c.address)
	if err != nil {
		return fmt.Errorf("error creating security council member binding: %w", err)
	}
	return nil
}

func (c *protocolDaoProposeKickFromSecurityCouncilContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.member.Exists,
		c.member.ID,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeKickFromSecurityCouncilContext) PrepareData(data *api.ProtocolDaoProposeKickFromSecurityCouncilData, opts *bind.TransactOpts) error {
	data.MemberDoesNotExist = !c.member.Exists.Get()
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.MemberDoesNotExist || data.InsufficientRpl)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
		message := fmt.Sprintf("kick %s (%s) from the security council", c.member.ID.Get(), c.address.Hex())
		txInfo, err := c.pdaoMgr.ProposeKickFromSecurityCouncil(message, c.address, blockNumber, pollard, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeKickFromSecurityCouncil: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
