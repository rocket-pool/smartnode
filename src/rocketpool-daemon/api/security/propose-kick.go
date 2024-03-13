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
)

// ===============
// === Factory ===
// ===============

type securityProposeKickContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposeKickContextFactory) Create(args url.Values) (*securityProposeKickContext, error) {
	c := &securityProposeKickContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityProposeKickContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityProposeKickContext, api.SecurityProposeKickData](
		router, "propose-kick", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeKickContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	address common.Address
	scMgr   *security.SecurityCouncilManager
	member  *security.SecurityCouncilMember
}

func (c *securityProposeKickContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireOnSecurityCouncil(c.handler.context)
	if err != nil {
		return err
	}

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	pSettings := pdaoMgr.Settings
	c.scMgr, err = security.NewSecurityCouncilManager(c.rp, pSettings)
	if err != nil {
		return fmt.Errorf("error creating security council manager binding: %w", err)
	}
	c.member, err = security.NewSecurityCouncilMember(c.rp, c.address)
	if err != nil {
		return fmt.Errorf("error creating security council member binding for node %s: %w", c.address.Hex(), err)
	}
	return nil
}

func (c *securityProposeKickContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.member.Exists,
		c.member.ID,
	)
}

func (c *securityProposeKickContext) PrepareData(data *api.SecurityProposeKickData, opts *bind.TransactOpts) error {
	data.MemberDoesNotExist = !c.member.Exists.Get()
	data.CanPropose = !(data.MemberDoesNotExist)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := fmt.Sprintf("kick %s (%s)", c.member.ID.Get(), c.address.Hex())
		txInfo, err := c.scMgr.ProposeKick(message, c.address, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeKick: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
