package odao

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type oracleDaoProposeSettingContextFactory struct {
	handler *OracleDaoHandler
}

func (f *oracleDaoProposeSettingContextFactory) Create(args url.Values) (*oracleDaoProposeSettingContext, error) {
	c := &oracleDaoProposeSettingContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("contract", args, &c.contractNameString),
		server.GetStringFromVars("setting", args, &c.setting),
		server.GetStringFromVars("value", args, &c.valueString),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoProposeSettingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposeSettingContext, api.OracleDaoProposeSettingData](
		router, "setting/propose", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type oracleDaoProposeSettingContext struct {
	handler     *OracleDaoHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	contractNameString string
	setting            string
	valueString        string
	odaoMember         *oracle.OracleDaoMember
	oSettings          *oracle.OracleDaoSettings
	odaoMgr            *oracle.OracleDaoManager
}

func (c *oracleDaoProposeSettingContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireOnOracleDao()
	if err != nil {
		return err
	}

	// Bindings
	c.odaoMember, err = oracle.NewOracleDaoMember(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating oracle DAO member binding: %w", err)
	}
	c.odaoMgr, err = oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating Oracle DAO manager binding: %w", err)
	}
	c.oSettings = c.odaoMgr.Settings
	return nil
}

func (c *oracleDaoProposeSettingContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.odaoMember.LastProposalTime,
		c.oSettings.Proposal.CooldownTime,
	)
}

func (c *oracleDaoProposeSettingContext) PrepareData(data *api.OracleDaoProposeSettingData, opts *bind.TransactOpts) error {
	// Get the timestamp of the latest block
	latestHeader, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(latestHeader.Time), 0)
	cooldownTime := c.oSettings.Proposal.CooldownTime.Formatted()

	// Check proposal details
	data.ProposalCooldownActive = isProposalCooldownActive(cooldownTime, c.odaoMember.LastProposalTime.Formatted(), currentTime)
	data.CanPropose = !(data.ProposalCooldownActive)

	// Make sure the setting exists
	settings := c.odaoMgr.Settings.GetSettings()
	category, exists := settings[rocketpool.ContractName(c.contractNameString)]
	if !exists {
		data.UnknownSetting = true
	}
	data.CanPropose = !(data.ProposalCooldownActive || data.UnknownSetting)

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

func (c *oracleDaoProposeSettingContext) createProposalTx(category oracle.SettingsCategory, opts *bind.TransactOpts) (bool, *core.TransactionInfo, error, error) {
	valueName := "value"

	// Try the bool settings
	for _, setting := range category.BoolSettings {
		if setting.GetPath() == c.setting {
			value, err := input.ValidateBool(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value as bool: %w", err), nil
			}
			txInfo, err := setting.ProposeSet(value, opts)
			return true, txInfo, nil, err
		}
	}

	// Try the uint settings
	for _, setting := range category.UintSettings {
		if setting.GetPath() == c.setting {
			value, err := input.ValidateBigInt(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value as *big.Int: %w", err), nil
			}
			txInfo, err := setting.ProposeSet(value, opts)
			return true, txInfo, nil, err
		}
	}

	return false, nil, nil, nil
}
