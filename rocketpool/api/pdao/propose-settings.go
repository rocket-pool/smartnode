package pdao

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/node"
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

type protocolDaoProposeSettingContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoProposeSettingContextFactory) Create(vars map[string]string) (*protocolDaoProposeSettingContext, error) {
	c := &protocolDaoProposeSettingContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("contract", vars, &c.contractNameString),
		server.GetStringFromVars("setting", vars, &c.setting),
		server.GetStringFromVars("value", vars, &c.valueString),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoProposeSettingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoProposeSettingContext, api.ProtocolDaoProposeSettingData](
		router, "setting/propose", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoProposeSettingContext struct {
	handler     *ProtocolDaoHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	bc          beacon.Client
	nodeAddress common.Address

	contractNameString string
	setting            string
	valueString        string
	node               *node.Node
	pdaoMgr            *protocol.ProtocolDaoManager
}

func (c *protocolDaoProposeSettingContext) Initialize() error {
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

func (c *protocolDaoProposeSettingContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.pdaoMgr.Settings.Proposals.ProposalBond,
		c.node.RplLocked,
		c.node.RplStake,
	)
}

func (c *protocolDaoProposeSettingContext) PrepareData(data *api.ProtocolDaoProposeSettingData, opts *bind.TransactOpts) error {
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
			return parseErr
		}
		if createErr != nil {
			return fmt.Errorf("error getting TX info for ProposeSet: %w", createErr)
		}
		if !validSetting {
			data.UnknownSetting = true
			data.CanPropose = false
		} else {
			data.TxInfo = txInfo
		}
	}
	return nil
}

func (c *protocolDaoProposeSettingContext) createProposalTx(category protocol.SettingsCategory, opts *bind.TransactOpts) (bool, *core.TransactionInfo, error, error) {
	valueName := "value"

	// Try the bool settings
	for _, setting := range category.BoolSettings {
		if setting.GetPath() == c.setting {
			value, err := input.ValidateBool(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value '%s' as bool: %w", c.valueString, err), nil
			}
			blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
			txInfo, err := setting.ProposeSet(value, blockNumber, pollard, opts)
			return true, txInfo, nil, err
		}
	}

	// Try the uint settings
	for _, setting := range category.UintSettings {
		if setting.GetPath() == c.setting {
			value, err := input.ValidateBigInt(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value '%s' as *big.Int: %w", c.valueString, err), nil
			}
			blockNumber, pollard, err := createPollard(c.rp, c.cfg, c.bc)
			txInfo, err := setting.ProposeSet(value, blockNumber, pollard, opts)
			return true, txInfo, nil, err
		}
	}

	return false, nil, nil, nil
}
