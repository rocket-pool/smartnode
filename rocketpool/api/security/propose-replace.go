package security

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
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

type securityProposeReplaceContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposeReplaceContextFactory) Create(args url.Values) (*securityProposeReplaceContext, error) {
	c := &securityProposeReplaceContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("existing-address", args, input.ValidateAddress, &c.existingAddress),
		server.GetStringFromVars("new-id", args, &c.newID),
		server.ValidateArg("new-address", args, input.ValidateAddress, &c.newAddress),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityProposeReplaceContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*securityProposeReplaceContext, api.SecurityProposeReplaceData](
		router, "propose-replace", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeReplaceContext struct {
	handler     *SecurityCouncilHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	existingAddress common.Address
	newID           string
	newAddress      common.Address
	scMgr           *security.SecurityCouncilManager
	existingMember  *security.SecurityCouncilMember
	newMember       *security.SecurityCouncilMember
}

func (c *securityProposeReplaceContext) Initialize() error {
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
	c.existingMember, err = security.NewSecurityCouncilMember(c.rp, c.existingAddress)
	if err != nil {
		return fmt.Errorf("error creating security council member binding for %s: %w", c.existingAddress.Hex(), err)
	}
	c.newMember, err = security.NewSecurityCouncilMember(c.rp, c.newAddress)
	if err != nil {
		return fmt.Errorf("error creating security council member binding for %s: %w", c.newAddress.Hex(), err)
	}
	return nil
}

func (c *securityProposeReplaceContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.existingMember.Exists,
		c.existingMember.ID,
		c.newMember.Exists,
	)
}

func (c *securityProposeReplaceContext) PrepareData(data *api.SecurityProposeReplaceData, opts *bind.TransactOpts) error {
	data.NewMemberAlreadyExists = c.newMember.Exists.Get()
	data.OldMemberDoesNotExist = !c.newMember.Exists.Get()
	data.CanPropose = !(data.NewMemberAlreadyExists || data.OldMemberDoesNotExist)

	// Get the tx
	if data.CanPropose && opts != nil {
		message := fmt.Sprintf("replace %s (%s) on the security council with %s (%s)", c.existingMember.ID.Get(), c.existingAddress.Hex(), c.newID, c.newAddress.Hex())
		txInfo, err := c.scMgr.ProposeReplace(message, c.existingAddress, c.newID, c.newAddress, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeReplace: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
