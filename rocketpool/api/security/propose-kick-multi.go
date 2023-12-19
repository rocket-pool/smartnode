package security

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type securityProposeKickMultiContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposeKickMultiContextFactory) Create(args url.Values) (*securityProposeKickMultiContext, error) {
	c := &securityProposeKickMultiContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", args, input.ValidateAddresses, &c.addresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityProposeKickMultiContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityProposeKickMultiContext, api.SecurityProposeKickMultiData](
		router, "propose-kick-multi", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeKickMultiContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	addresses []common.Address
	scMgr     *security.SecurityCouncilManager
	members   []*security.SecurityCouncilMember
}

func (c *securityProposeKickMultiContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireOnSecurityCouncil()
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
	c.members = make([]*security.SecurityCouncilMember, len(c.addresses))
	for i, address := range c.addresses {
		c.members[i], err = security.NewSecurityCouncilMember(c.rp, address)
		if err != nil {
			return fmt.Errorf("error creating security council member binding for node %s: %w", address.Hex(), err)
		}
	}
	return nil
}

func (c *securityProposeKickMultiContext) GetState(mc *batch.MultiCaller) {
	for _, member := range c.members {
		member.Exists.AddToQuery(mc)
	}
}

func (c *securityProposeKickMultiContext) PrepareData(data *api.SecurityProposeKickMultiData, opts *bind.TransactOpts) error {
	for _, member := range c.members {
		if !member.Exists.Get() {
			data.MembersDoNotExist = append(data.MembersDoNotExist, member.Address)
		}
	}
	data.CanPropose = (len(data.MembersDoNotExist) > 0)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := "kick multiple members"
		txInfo, err := c.scMgr.ProposeKickMulti(message, c.addresses, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeKickMulti: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
