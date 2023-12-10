package pdao

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
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

type protocolDaoProposeInviteToSecurityCouncilContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeInviteToSecurityCouncilContextFactory) Create(vars map[string]string) (*protocolDaoProposeInviteToSecurityCouncilContext, error) {
	c := &protocolDaoProposeInviteToSecurityCouncilContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("id", vars, &c.id),
		server.ValidateArg("address", vars, input.ValidateAddress, &c.address),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeInviteToSecurityCouncilContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeInviteToSecurityCouncilContext, api.ProtocolDaoProposeInviteToSecurityCouncilData](
		router, "security/invite", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeInviteToSecurityCouncilContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool
	cfg     *config.RocketPoolConfig
	bc      beacon.Client

	id      string
	address common.Address
	pdaoMgr *protocol.ProtocolDaoManager
	member  *security.SecurityCouncilMember
}

func (c *protocolDaoProposeInviteToSecurityCouncilContext) Initialize() error {
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

func (c *protocolDaoProposeInviteToSecurityCouncilContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.member.Exists,
	)
}

func (c *protocolDaoProposeInviteToSecurityCouncilContext) PrepareData(data *api.ProtocolDaoProposeInviteToSecurityCouncilData, opts *bind.TransactOpts) error {
	data.MemberAlreadyExists = c.member.Exists.Get()
	data.CanPropose = !(data.MemberAlreadyExists)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
		message := fmt.Sprintf("invite %s (%s) to the security council", c.id, c.address.Hex())
		txInfo, err := c.pdaoMgr.ProposeInviteToSecurityCouncil(message, c.id, c.address, blockNumber, pollard, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeInviteToSecurityCouncil: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
