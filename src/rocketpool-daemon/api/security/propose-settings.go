package security

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/dao/security"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type securityProposeSettingContextFactory struct {
	handler *SecurityCouncilHandler
}

func (f *securityProposeSettingContextFactory) Create(args url.Values) (*securityProposeSettingContext, error) {
	c := &securityProposeSettingContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("contract", args, &c.contractNameString),
		server.GetStringFromVars("setting", args, &c.setting),
		server.GetStringFromVars("value", args, &c.valueString),
	}
	return c, errors.Join(inputErrs...)
}

func (f *securityProposeSettingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*securityProposeSettingContext, api.SecurityProposeSettingData](
		router, "setting/propose", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type securityProposeSettingContext struct {
	handler *SecurityCouncilHandler

	contractNameString string
	setting            string
	valueString        string
}

func (c *securityProposeSettingContext) PrepareData(data *api.SecurityProposeSettingData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireOnSecurityCouncil(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pdaoMgr, err := protocol.NewProtocolDaoManager(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating protocol DAO manager binding: %w", err)
	}
	pSettings := pdaoMgr.Settings
	scMgr, err := security.NewSecurityCouncilManager(rp, pSettings)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating security council manager binding: %w", err)
	}

	// Make sure the setting exists
	settings := scMgr.Settings.GetSettings()
	category, exists := settings[rocketpool.ContractName(c.contractNameString)]
	if !exists {
		data.UnknownSetting = true
	}
	data.CanPropose = !(data.UnknownSetting)

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

func (c *securityProposeSettingContext) createProposalTx(category security.SettingsCategory, opts *bind.TransactOpts) (bool, *eth.TransactionInfo, error, error) {
	valueName := "value"

	// Try the bool settings
	for _, setting := range category.BoolSettings {
		if string(setting.GetProtocolDaoSetting().GetSettingName()) == c.setting {
			value, err := input.ValidateBool(valueName, c.valueString)
			if err != nil {
				return false, nil, fmt.Errorf("error parsing value '%s' as bool: %w", c.valueString, err), nil
			}
			txInfo, err := setting.ProposeSet(value, opts)
			return true, txInfo, nil, err
		}
	}

	return false, nil, nil, nil
}
