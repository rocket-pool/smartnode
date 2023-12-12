package odao

import (
	"context"
	"errors"
	"fmt"
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

func (f *oracleDaoProposeSettingContextFactory) Create(vars map[string]string) (*oracleDaoProposeSettingContext, error) {
	c := &oracleDaoProposeSettingContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("setting", vars, &c.setting),
		server.GetStringFromVars("value", vars, &c.valueString),
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

	setting     string
	valueString string
	odaoMember  *oracle.OracleDaoMember
	oSettings   *oracle.OracleDaoSettings
	odaoMgr     *oracle.OracleDaoManager
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

	// Get the tx
	if data.CanPropose && opts != nil {
		txInfo, err := c.createProposalTx(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for proposal: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}

func (c *oracleDaoProposeSettingContext) createProposalTx(opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	valueName := "value"
	boolSettings, uintSettings := c.oSettings.GetSettings()

	// Try the bool settings
	for _, setting := range boolSettings {
		if setting.GetPath() == c.setting {
			value, err := input.ValidateBool(valueName, c.valueString)
			if err != nil {
				return nil, fmt.Errorf("error parsing value as bool: %w", err)
			}
			return setting.ProposeSet(value, opts)
		}
	}

	// Try the uint settings
	for _, setting := range uintSettings {
		if setting.GetPath() == c.setting {
			value, err := input.ValidateBigInt(valueName, c.valueString)
			if err != nil {
				return nil, fmt.Errorf("error parsing value as *big.Int: %w", err)
			}
			return setting.ProposeSet(value, opts)
		}
	}

	return nil, fmt.Errorf("[%s] is not a valid Oracle DAO setting", c.setting)
}
