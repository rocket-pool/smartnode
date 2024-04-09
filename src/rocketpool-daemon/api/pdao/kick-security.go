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
		router, "security/kick", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeKickFromSecurityCouncilContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	address common.Address
	node    *node.Node
	pdaoMgr *protocol.ProtocolDaoManager
	member  *security.SecurityCouncilMember
}

func (c *protocolDaoProposeKickFromSecurityCouncilContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

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
	c.member, err = security.NewSecurityCouncilMember(c.rp, c.address)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
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

func (c *protocolDaoProposeKickFromSecurityCouncilContext) PrepareData(data *api.ProtocolDaoProposeKickFromSecurityCouncilData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	data.MemberDoesNotExist = !c.member.Exists.Get()
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.MemberDoesNotExist || data.InsufficientRpl)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(ctx, c.rp, c.cfg, c.bc)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating pollard for proposal creation: %w", err)
		}

		message := fmt.Sprintf("kick %s (%s) from the security council", c.member.ID.Get(), c.address.Hex())
		txInfo, err := c.pdaoMgr.ProposeKickFromSecurityCouncil(message, c.address, blockNumber, pollard, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeKickFromSecurityCouncil: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
