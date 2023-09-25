package odao

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
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
		server.GetStringFromVars("value", vars, &c.value),
	}
	return c, errors.Join(inputErrs...)
}

func (f *oracleDaoProposeSettingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*oracleDaoProposeSettingContext, api.OracleDaoProposeSettingData](
		router, "propose-setting", f, f.handler.serviceProvider,
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
	candidate   *oracle.OracleDaoMember
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
	c.candidate, err = oracle.NewOracleDaoMember(c.rp, c.address)
	if err != nil {
		return fmt.Errorf("error creating candidate oracle DAO member binding: %w", err)
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
		c.candidate.Exists,
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
	settingNames := api.GetOracleDaoSettingsNames()
	valueName := "value"

	// Prepare the TX based on the setting
	switch c.setting {
	case settingNames.Member.Quorum:
		value, err := input.ValidateBigInt(valueName, c.valueString)
		if err != nil {
			return nil, err
		}
		return c.odaoMgr.Settings.Member.Quorum.ProposeSet(value, opts)

	case settingNames.Member.RplBond:
		value, err := input.ValidateBigInt(valueName, c.valueString)
		if err != nil {
			return nil, err
		}
		return c.odaoMgr.Settings.Member.RplBond.ProposeSet(value, opts)

	default:
		return nil, fmt.Errorf("[%s] is not a valid Oracle DAO setting", c.setting)
	}
}

func canProposeSetting(c *cli.Context, w *wallet.LocalWallet, rp *rocketpool.RocketPool) (*api.OracleDaoProposeSettingData, error) {

	// Response
	response := api.OracleDaoProposeSettingData{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check if proposal cooldown is active
	proposalCooldownActive, err := getProposalCooldownActive(rp, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	response.ProposalCooldownActive = proposalCooldownActive

	// Update & return response
	response.CanPropose = !response.ProposalCooldownActive
	return &response, nil

}

func canProposeSettingMembersQuorum(c *cli.Context, quorum float64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeQuorumGas(rp, quorum, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingMembersQuorum(c *cli.Context, quorum float64) (*api.ProposeTNDAOSettingMembersQuorumResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingMembersQuorumResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeQuorum(rp, quorum, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingMembersRplBond(c *cli.Context, bondAmountWei *big.Int) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeRPLBondGas(rp, bondAmountWei, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingMembersRplBond(c *cli.Context, bondAmountWei *big.Int) (*api.ProposeTNDAOSettingMembersRplBondResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingMembersRplBondResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeRPLBond(rp, bondAmountWei, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingMinipoolUnbondedMax(c *cli.Context, unbondedMinipoolMax uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeMinipoolUnbondedMaxGas(rp, unbondedMinipoolMax, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingMinipoolUnbondedMax(c *cli.Context, unbondedMinipoolMax uint64) (*api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeMinipoolUnbondedMax(rp, unbondedMinipoolMax, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingProposalCooldown(c *cli.Context, proposalCooldownTimespan uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeProposalCooldownTimeGas(rp, proposalCooldownTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingProposalCooldown(c *cli.Context, proposalCooldownTimespan uint64) (*api.ProposeTNDAOSettingProposalCooldownResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalCooldownResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeProposalCooldownTime(rp, proposalCooldownTimespan, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingProposalVoteTimespan(c *cli.Context, proposalVoteTimespan uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeProposalVoteTimeGas(rp, proposalVoteTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingProposalVoteTimespan(c *cli.Context, proposalVoteTimespan uint64) (*api.ProposeTNDAOSettingProposalVoteTimespanResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalVoteTimespanResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeProposalVoteTime(rp, proposalVoteTimespan, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingProposalVoteDelayTimespan(c *cli.Context, proposalDelayTimespan uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeProposalVoteDelayTimeGas(rp, proposalDelayTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingProposalVoteDelayTimespan(c *cli.Context, proposalDelayTimespan uint64) (*api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeProposalVoteDelayTime(rp, proposalDelayTimespan, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingProposalExecuteTimespan(c *cli.Context, proposalExecuteTimespan uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeProposalExecuteTimeGas(rp, proposalExecuteTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingProposalExecuteTimespan(c *cli.Context, proposalExecuteTimespan uint64) (*api.ProposeTNDAOSettingProposalExecuteTimespanResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalExecuteTimespanResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeProposalExecuteTime(rp, proposalExecuteTimespan, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingProposalActionTimespan(c *cli.Context, proposalActionTimespan uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeProposalActionTimeGas(rp, proposalActionTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingProposalActionTimespan(c *cli.Context, proposalActionTimespan uint64) (*api.ProposeTNDAOSettingProposalActionTimespanResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalActionTimespanResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeProposalActionTime(rp, proposalActionTimespan, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingScrubPeriod(c *cli.Context, scrubPeriod uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeScrubPeriodGas(rp, scrubPeriod, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingScrubPeriod(c *cli.Context, scrubPeriod uint64) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeScrubPeriod(rp, scrubPeriod, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingPromotionScrubPeriod(c *cli.Context, promotionScrubPeriod uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposePromotionScrubPeriodGas(rp, promotionScrubPeriod, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingPromotionScrubPeriod(c *cli.Context, promotionScrubPeriod uint64) (*api.ProposeTNDAOSettingPromotionScrubPeriodResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingPromotionScrubPeriodResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposePromotionScrubPeriod(rp, promotionScrubPeriod, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingScrubPenaltyEnabled(c *cli.Context, enabled bool) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeScrubPenaltyEnabledGas(rp, enabled, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingScrubPenaltyEnabled(c *cli.Context, enabled bool) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeScrubPenaltyEnabled(rp, enabled, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingBondReductionWindowStart(c *cli.Context, bondReductionWindowStart uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeBondReductionWindowStartGas(rp, bondReductionWindowStart, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingBondReductionWindowStart(c *cli.Context, bondReductionWindowStart uint64) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeBondReductionWindowStart(rp, bondReductionWindowStart, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeSettingBondReductionWindowLength(c *cli.Context, bondReductionWindowLength uint64) (*api.OracleDaoProposeSettingData, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	response, err := canProposeSetting(c, w, rp)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := trustednode.EstimateProposeBondReductionWindowLengthGas(rp, bondReductionWindowLength, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
	return response, nil

}

func proposeSettingBondReductionWindowLength(c *cli.Context, bondReductionWindowLength uint64) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	proposalId, hash, err := trustednode.ProposeBondReductionWindowLength(rp, bondReductionWindowLength, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}
