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

type protocolDaoProposeReplaceMemberOfSecurityCouncilContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeReplaceMemberOfSecurityCouncilContextFactory) Create(args url.Values) (*protocolDaoProposeReplaceMemberOfSecurityCouncilContext, error) {
	c := &protocolDaoProposeReplaceMemberOfSecurityCouncilContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("existing-address", args, input.ValidateAddress, &c.existingAddress),
		server.GetStringFromVars("new-id", args, &c.newID),
		server.ValidateArg("new-address", args, input.ValidateAddress, &c.newAddress),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeReplaceMemberOfSecurityCouncilContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeReplaceMemberOfSecurityCouncilContext, api.ProtocolDaoProposeReplaceMemberOfSecurityCouncilData](
		router, "security/replace", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeReplaceMemberOfSecurityCouncilContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	existingAddress common.Address
	newID           string
	newAddress      common.Address
	node            *node.Node
	pdaoMgr         *protocol.ProtocolDaoManager
	existingMember  *security.SecurityCouncilMember
	newMember       *security.SecurityCouncilMember
}

func (c *protocolDaoProposeReplaceMemberOfSecurityCouncilContext) Initialize() (types.ResponseStatus, error) {
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
	c.existingMember, err = security.NewSecurityCouncilMember(c.rp, c.existingAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding for %s: %w", c.existingAddress.Hex(), err)
	}
	c.newMember, err = security.NewSecurityCouncilMember(c.rp, c.newAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council member binding for %s: %w", c.newAddress.Hex(), err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoProposeReplaceMemberOfSecurityCouncilContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.existingMember.Exists,
		c.existingMember.ID,
		c.newMember.Exists,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeReplaceMemberOfSecurityCouncilContext) PrepareData(data *api.ProtocolDaoProposeReplaceMemberOfSecurityCouncilData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	data.NewMemberAlreadyExists = c.newMember.Exists.Get()
	data.OldMemberDoesNotExist = !c.existingMember.Exists.Get()
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.NewMemberAlreadyExists || data.OldMemberDoesNotExist || data.InsufficientRpl)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(ctx, c.rp, c.cfg, c.bc)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating pollard for proposal creation: %w", err)
		}

		message := fmt.Sprintf("replace %s (%s) on the security council with %s (%s)", c.existingMember.ID.Get(), c.existingAddress.Hex(), c.newID, c.newAddress.Hex())
		txInfo, err := c.pdaoMgr.ProposeReplaceSecurityCouncilMember(message, c.existingAddress, c.newID, c.newAddress, blockNumber, pollard, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeReplaceSecurityCouncilMember: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
