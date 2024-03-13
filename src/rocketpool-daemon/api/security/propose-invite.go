package security

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils"
)

// ===============
// === Factory ===
// ===============

type securityProposeInviteContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposeInviteContextFactory) Create(args url.Values) (*securityProposeInviteContext, error) {
	c := &securityProposeInviteContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("id", args, utils.ValidateDAOMemberID, &c.id),
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityProposeInviteContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityProposeInviteContext, api.SecurityProposeInviteData](
		router, "propose-invite", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeInviteContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	id       string
	address  common.Address
	scMgr    *security.SecurityCouncilManager
	scMember *security.SecurityCouncilMember
}

func (c *securityProposeInviteContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireOnSecurityCouncil(c.handler.context)
	if err != nil {
		return err
	}

	// Bindings
	c.scMember, err = security.NewSecurityCouncilMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating security council member binding: %w", err)
	}
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	pSettings := pdaoMgr.Settings
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, pSettings)
	if err != nil {
		return fmt.Errorf("error creating security council manager binding: %w", err)
	}
	return nil
}

func (c *securityProposeInviteContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.scMember.Exists,
	)
}

func (c *securityProposeInviteContext) PrepareData(data *api.SecurityProposeInviteData, opts *bind.TransactOpts) error {
	data.MemberAlreadyExists = c.scMember.Exists.Get()
	data.CanPropose = !(data.MemberAlreadyExists)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := fmt.Sprintf("invite %s (%s)", c.id, c.address.Hex())
		txInfo, err := c.scMgr.ProposeInvite(message, c.id, c.address, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeInvite: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
