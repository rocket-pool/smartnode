package pdao

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
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

type protocolDaoProposeRecurringSpendContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeRecurringSpendContextFactory) Create(args url.Values) (*protocolDaoProposeRecurringSpendContext, error) {
	c := &protocolDaoProposeRecurringSpendContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("contract-name", args, &c.contractName),
		server.ValidateArg("recipient", args, input.ValidateAddress, &c.recipient),
		server.ValidateArg("amount-per-period", args, input.ValidateBigInt, &c.amountPerPeriod),
		server.ValidateArg("period-length", args, input.ValidateDuration, &c.periodLength),
		server.ValidateArg("start-time", args, input.ValidateTime, &c.startTime),
		server.ValidateArg("num-periods", args, input.ValidatePositiveUint, &c.numberOfPeriods),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeRecurringSpendContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeRecurringSpendContext, api.ProtocolDaoGeneralProposeData](
		router, "recurring-spend", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeRecurringSpendContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	bc          beacon.Client
	nodeAddress common.Address

	contractName    string
	recipient       common.Address
	amountPerPeriod *big.Int
	periodLength    time.Duration
	startTime       time.Time
	numberOfPeriods uint64
	node            *node.Node
	pdaoMgr         *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposeRecurringSpendContext) Initialize() error {
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

func (c *protocolDaoProposeRecurringSpendContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeRecurringSpendContext) PrepareData(data *api.ProtocolDaoGeneralProposeData, opts *bind.TransactOpts) error {
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.InsufficientRpl)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
		message := fmt.Sprintf("recurring payment to %s", c.contractName)
		txInfo, err := c.pdaoMgr.ProposeRecurringTreasurySpend(message, c.contractName, c.recipient, c.amountPerPeriod, c.periodLength, c.startTime, c.numberOfPeriods, blockNumber, pollard, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ProposeRecurringTreasurySpend: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
