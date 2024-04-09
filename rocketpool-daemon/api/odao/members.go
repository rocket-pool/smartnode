package odao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type oracleDaoMembersContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoMembersContextFactory) Create(args url.Values) (*oracleDaoMembersContext, error) {
	c := &oracleDaoMembersContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *oracleDaoMembersContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoMembersContext, api.OracleDaoMembersData](
		router, "members", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoMembersContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	odaoMgr *oracle.OracleDaoManager
}

func (c *oracleDaoMembersContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *oracleDaoMembersContext) GetState(mc *batch.MultiCaller) {
	c.odaoMgr.MemberCount.AddToQuery(mc)
}

func (c *oracleDaoMembersContext) PrepareData(data *api.OracleDaoMembersData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the member addresses
	addresses, err := c.odaoMgr.GetMemberAddresses(c.odaoMgr.MemberCount.Formatted(), nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Oracle DAO member addresses: %w", err)
	}

	// Get the member bindings
	members, err := c.odaoMgr.CreateMembersFromAddresses(addresses, true, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating Oracle DAO member bindings: %w", err)
	}

	for _, member := range members {
		memberDetails := api.OracleDaoMemberDetails{
			Address:          member.Address,
			Exists:           member.Exists.Get(),
			ID:               member.ID.Get(),
			Url:              member.Url.Get(),
			JoinedTime:       member.JoinedTime.Formatted(),
			LastProposalTime: member.LastProposalTime.Formatted(),
			RplBondAmount:    member.RplBondAmount.Get(),
		}
		data.Members = append(data.Members, memberDetails)
	}
	return types.ResponseStatus_Success, nil
}
