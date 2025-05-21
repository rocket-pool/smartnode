package odao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canProposeSetting(c *cli.Context, w *wallet.Wallet, rp *rocketpool.RocketPool) (*api.CanProposeTNDAOSettingResponse, error) {

	// Response
	response := api.CanProposeTNDAOSettingResponse{}

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

func canProposeSettingMembersQuorum(c *cli.Context, quorum float64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingMembersRplBond(c *cli.Context, bondAmountWei *big.Int) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingMinipoolUnbondedMax(c *cli.Context, unbondedMinipoolMax uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingProposalCooldown(c *cli.Context, proposalCooldownTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingProposalVoteTimespan(c *cli.Context, proposalVoteTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingProposalVoteDelayTimespan(c *cli.Context, proposalDelayTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingProposalExecuteTimespan(c *cli.Context, proposalExecuteTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingProposalActionTimespan(c *cli.Context, proposalActionTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingScrubPeriod(c *cli.Context, scrubPeriod uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingPromotionScrubPeriod(c *cli.Context, promotionScrubPeriod uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingScrubPenaltyEnabled(c *cli.Context, enabled bool) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingBondReductionWindowStart(c *cli.Context, bondReductionWindowStart uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingBondReductionWindowLength(c *cli.Context, bondReductionWindowLength uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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
