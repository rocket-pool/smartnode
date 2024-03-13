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
		router, "rewards-percentages/propose", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeRewardsPercentagesContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	nodePercent *big.Int
	odaoPercent *big.Int
	pdaoPercent *big.Int
	node        *node.Node
	pdaoMgr     *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposeRewardsPercentagesContext) Initialize() error {
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
	return nil
}

func (c *protocolDaoProposeRewardsPercentagesContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeRewardsPercentagesContext) PrepareData(data *api.ProtocolDaoGeneralProposeData, opts *bind.TransactOpts) error {
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.InsufficientRpl)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
		message := "update RPL rewards distribution"
		txInfo, err := c.pdaoMgr.ProposeSetRewardsPercentages(message, c.odaoPercent, c.pdaoPercent, c.nodePercent, blockNumber, pollard, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeSetRewardsPercentages: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
