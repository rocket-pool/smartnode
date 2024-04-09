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

type protocolDaoProposeRecurringSpendUpdateContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeRecurringSpendUpdateContextFactory) Create(args url.Values) (*protocolDaoProposeRecurringSpendUpdateContext, error) {
	c := &protocolDaoProposeRecurringSpendUpdateContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("contract-name", args, &c.contractName),
		server.ValidateArg("recipient", args, input.ValidateAddress, &c.recipient),
		server.ValidateArg("amount-per-period", args, input.ValidateBigInt, &c.amountPerPeriod),
		server.ValidateArg("period-length", args, input.ValidateDuration, &c.periodLength),
		server.ValidateArg("num-periods", args, input.ValidatePositiveUint, &c.numberOfPeriods),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeRecurringSpendUpdateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeRecurringSpendUpdateContext, api.ProtocolDaoProposeRecurringSpendUpdateData](
		router, "recurring-spend-update", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeRecurringSpendUpdateContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	contractName    string
	recipient       common.Address
	amountPerPeriod *big.Int
	periodLength    time.Duration
	numberOfPeriods uint64
	node            *node.Node
	pdaoMgr         *protocol.ProtocolDaoManager
	contractExists  bool
}

func (c *protocolDaoProposeRecurringSpendUpdateContext) Initialize() (types.ResponseStatus, error) {
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
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoProposeRecurringSpendUpdateContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
	c.pdaoMgr.GetContractExists(mc, &c.contractExists, c.contractName)
}

func (c *protocolDaoProposeRecurringSpendUpdateContext) PrepareData(data *api.ProtocolDaoProposeRecurringSpendUpdateData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	data.DoesNotExist = !c.contractExists
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)
	data.CanPropose = !(data.InsufficientRpl || data.DoesNotExist)

	// Get the tx
	if data.CanPropose && opts != nil {
		blockNumber, pollard, err := createPollard(ctx, c.rp, c.cfg, c.bc)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating pollard for proposal creation: %w", err)
		}

		message := fmt.Sprintf("recurring payment to %s", c.contractName)
		txInfo, err := c.pdaoMgr.ProposeRecurringTreasurySpendUpdate(message, c.contractName, c.recipient, c.amountPerPeriod, c.periodLength, c.numberOfPeriods, blockNumber, pollard, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeRecurringTreasurySpendUpdate: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
