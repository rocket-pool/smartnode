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
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoProposeRewardsPercentagesContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeRewardsPercentagesContextFactory) Create(args url.Values) (*protocolDaoProposeRewardsPercentagesContext, error) {
	c := &protocolDaoProposeRewardsPercentagesContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("node", args, input.ValidateBigInt, &c.nodePercent),
		server.ValidateArg("odao", args, input.ValidateBigInt, &c.odaoPercent),
		server.ValidateArg("pdao", args, input.ValidateBigInt, &c.pdaoPercent),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeRewardsPercentagesContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeRewardsPercentagesContext, api.ProtocolDaoGeneralProposeData](
		router, "rewards-percentages/propose", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeRewardsPercentagesContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	res         *config.MergedResources
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	nodePercent *big.Int
	odaoPercent *big.Int
	pdaoPercent *big.Int
	node        *node.Node
	pdaoMgr     *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposeRewardsPercentagesContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.res = sp.GetResources()
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
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoProposeRewardsPercentagesContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
		c.node.IsRplLockingAllowed,
	)
}

func (c *protocolDaoProposeRewardsPercentagesContext) PrepareData(data *api.ProtocolDaoGeneralProposeData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Validate sum of percentages == 100%
	one := eth.EthToWei(1)
	sum := big.NewInt(0).Set(c.nodePercent)
	sum.Add(sum, c.odaoPercent)
	sum.Add(sum, c.pdaoPercent)
	if sum.Cmp(one) != 0 {
		return types.ResponseStatus_InvalidArguments, fmt.Errorf("values don't add up to 100%%")
	}

	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.IsRplLockingDisallowed = !c.node.IsRplLockingAllowed.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.InsufficientRpl || data.IsRplLockingDisallowed)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(ctx, c.handler.logger.Logger, c.rp, c.cfg, c.res, c.bc)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating pollard for proposal creation: %w", err)
		}

		message := "update RPL rewards distribution"
		txInfo, err := c.pdaoMgr.ProposeSetRewardsPercentages(message, c.odaoPercent, c.pdaoPercent, c.nodePercent, blockNumber, pollard, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeSetRewardsPercentages: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
