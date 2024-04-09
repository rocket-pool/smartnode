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

type protocolDaoProposeSettingContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeSettingContextFactory) Create(args url.Values) (*protocolDaoProposeSettingContext, error) {
	c := &protocolDaoProposeSettingContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("contract", args, &c.contractNameString),
		server.GetStringFromVars("setting", args, &c.setting),
		server.GetStringFromVars("value", args, &c.valueString),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeSettingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeSettingContext, api.ProtocolDaoProposeSettingData](
		router, "setting/propose", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeSettingContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.SmartNodeConfig
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	contractNameString string
	setting            string
	valueString        string
	node               *node.Node
	pdaoMgr            *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposeSettingContext) Initialize() (types.ResponseStatus, error) {
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

func (c *protocolDaoProposeSettingContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeSettingContext) PrepareData(data *api.ProtocolDaoProposeSettingData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.StakedRpl = c.node.RplStake.Get()
	data.LockedRpl = c.node.RplLocked.Get()
	data.ProposalBond = c.pdaoMgr.Settings.Proposals.ProposalBond.Get()
	unlockedRpl := big.NewInt(0).Sub(data.StakedRpl, data.LockedRpl)
	data.InsufficientRpl = (unlockedRpl.Cmp(data.ProposalBond) < 0)

	// Make sure the setting exists
	settings := c.pdaoMgr.Settings.GetSettings()
	category, exists := settings[rocketpool.ContractName(c.contractNameString)]
	if !exists {
		data.UnknownSetting = true
	}
	data.CanPropose = !(data.InsufficientRpl || data.UnknownSetting)

	// Get the tx
	if data.CanPropose && opts != nil {
		validSetting, txInfo, parseErr, createErr := c.createProposalTx(category, opts)
		if parseErr != nil {
			return types.ResponseStatus_InvalidArguments, parseErr
		}
		if createErr != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ProposeSet: %w", createErr)
		}
		if !validSetting {
			data.UnknownSetting = true
			data.CanPropose = false
		} else {
			data.TxInfo = txInfo
		}
	}
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoProposeSettingContext) createProposalTx(category protocol.SettingsCategory, opts *bind.TransactOpts) (bool, *eth.TransactionInfo, error, error) {
	ctx := c.handler.ctx
	valueName := "value"

	// Try the bool settings
	for _, setting := range category.BoolSettings {
		if string(setting.GetSettingName()) == c.setting {
			value, err := input.ValidateBool(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value '%s' as bool: %w", c.valueString, err), nil
			}
			blockNumber, pollard, err := createPollard(ctx, c.rp, c.cfg, c.bc)
			if err != nil {
				return false, nil, fmt.Errorf("error creating pollard for proposal creation: %w", err), nil
			}

			txInfo, err := setting.ProposeSet(value, blockNumber, pollard, opts)
			return true, txInfo, nil, err
		}
	}

	// Try the uint settings
	for _, setting := range category.UintSettings {
		if string(setting.GetSettingName()) == c.setting {
			value, err := input.ValidateBigInt(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value '%s' as *big.Int: %w", c.valueString, err), nil
			}
			blockNumber, pollard, err := createPollard(ctx, c.rp, c.cfg, c.bc)
			if err != nil {
				return false, nil, fmt.Errorf("error creating pollard for proposal creation: %w", err), nil
			}

			txInfo, err := setting.ProposeSet(value, blockNumber, pollard, opts)
			return true, txInfo, nil, err
		}
	}

	return false, nil, nil, nil
}
