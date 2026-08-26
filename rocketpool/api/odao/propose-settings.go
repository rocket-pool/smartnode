package odao

import (
	"math/big"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeSetting(c *cli.Command, w wallet.Wallet, rp *rocketpool.RocketPool) (*api.CanProposeTNDAOSettingResponse, error) {

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

func canProposeSettingMembersQuorum(c *cli.Command, quorum float64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeQuorumGas(rp, quorum, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingMembersQuorum(c *cli.Command, quorum float64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingMembersQuorumResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingMembersQuorumResponse{}

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

func canProposeSettingMembersRplBond(c *cli.Command, bondAmountWei *big.Int) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeRPLBondGas(rp, bondAmountWei, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingMembersRplBond(c *cli.Command, bondAmountWei *big.Int, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingMembersRplBondResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingMembersRplBondResponse{}

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

func canProposeSettingProposalCooldown(c *cli.Command, proposalCooldownTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeProposalCooldownTimeGas(rp, proposalCooldownTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingProposalCooldown(c *cli.Command, proposalCooldownTimespan uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingProposalCooldownResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalCooldownResponse{}

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

func canProposeSettingProposalVoteTimespan(c *cli.Command, proposalVoteTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeProposalVoteTimeGas(rp, proposalVoteTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingProposalVoteTimespan(c *cli.Command, proposalVoteTimespan uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingProposalVoteTimespanResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalVoteTimespanResponse{}

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

func canProposeSettingProposalVoteDelayTimespan(c *cli.Command, proposalDelayTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeProposalVoteDelayTimeGas(rp, proposalDelayTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingProposalVoteDelayTimespan(c *cli.Command, proposalDelayTimespan uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse{}

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

func canProposeSettingProposalExecuteTimespan(c *cli.Command, proposalExecuteTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeProposalExecuteTimeGas(rp, proposalExecuteTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingProposalExecuteTimespan(c *cli.Command, proposalExecuteTimespan uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingProposalExecuteTimespanResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalExecuteTimespanResponse{}

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

func canProposeSettingProposalActionTimespan(c *cli.Command, proposalActionTimespan uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeProposalActionTimeGas(rp, proposalActionTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingProposalActionTimespan(c *cli.Command, proposalActionTimespan uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingProposalActionTimespanResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingProposalActionTimespanResponse{}

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

func canProposeSettingScrubPeriod(c *cli.Command, scrubPeriod uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeScrubPeriodGas(rp, scrubPeriod, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingScrubPeriod(c *cli.Command, scrubPeriod uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

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

func canProposeSettingPromotionScrubPeriod(c *cli.Command, promotionScrubPeriod uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposePromotionScrubPeriodGas(rp, promotionScrubPeriod, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingPromotionScrubPeriod(c *cli.Command, promotionScrubPeriod uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingPromotionScrubPeriodResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingPromotionScrubPeriodResponse{}

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

func canProposeSettingScrubPenaltyEnabled(c *cli.Command, enabled bool) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeScrubPenaltyEnabledGas(rp, enabled, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingScrubPenaltyEnabled(c *cli.Command, enabled bool, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

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

func canProposeSettingBondReductionWindowStart(c *cli.Command, bondReductionWindowStart uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeBondReductionWindowStartGas(rp, bondReductionWindowStart, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingBondReductionWindowStart(c *cli.Command, bondReductionWindowStart uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

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

func canProposeSettingBondReductionWindowLength(c *cli.Command, bondReductionWindowLength uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

	gasLimits, err := trustednode.EstimateProposeBondReductionWindowLengthGas(rp, bondReductionWindowLength, opts)
	if err != nil {
		return nil, err
	}

	response.GasLimits = gasLimits
	return response, nil

}

func proposeSettingBondReductionWindowLength(c *cli.Command, bondReductionWindowLength uint64, t *snroute.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingScrubPeriodResponse{}

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

func canProposeMembersQuorumHandler(ctx snroute.Context) {
	quorum, err := parseFloat64(ctx.Request, "quorum")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingMembersQuorum(ctx.Command(), quorum)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeMembersQuorumHandler(ctx snroute.WriteContext) {
	quorum, err := parseFloat64(ctx.Request, "quorum")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingMembersQuorum(ctx.Command(), quorum, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeMembersRplbondHandler(ctx snroute.Context) {
	bond, err := parseBigInt(ctx.Request, "bondAmountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingMembersRplBond(ctx.Command(), bond)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeMembersRplbondHandler(ctx snroute.WriteContext) {
	bond, err := parseBigInt(ctx.Request, "bondAmountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingMembersRplBond(ctx.Command(), bond, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeProposalCooldownHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingProposalCooldown(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeProposalCooldownHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingProposalCooldown(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeProposalVoteTimespanHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingProposalVoteTimespan(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeProposalVoteTimespanHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingProposalVoteTimespan(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeProposalVoteDelayTimespanHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingProposalVoteDelayTimespan(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeProposalVoteDelayTimespanHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingProposalVoteDelayTimespan(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeProposalExecuteTimespanHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingProposalExecuteTimespan(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeProposalExecuteTimespanHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingProposalExecuteTimespan(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeProposalActionTimespanHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingProposalActionTimespan(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeProposalActionTimespanHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingProposalActionTimespan(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeScrubPeriodHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingScrubPeriod(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeScrubPeriodHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingScrubPeriod(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposePromotionScrubPeriodHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingPromotionScrubPeriod(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposePromotionScrubPeriodHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingPromotionScrubPeriod(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeScrubPenaltyEnabledHandler(ctx snroute.Context) {
	enabledStr := ctx.Request.URL.Query().Get("enabled")
	resp, err := canProposeSettingScrubPenaltyEnabled(ctx.Command(), enabledStr == "true")
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeScrubPenaltyEnabledHandler(ctx snroute.WriteContext) {
	enabledStr := ctx.Request.FormValue("enabled")
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingScrubPenaltyEnabled(ctx.Command(), enabledStr == "true", opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeBondReductionWindowStartHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingBondReductionWindowStart(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeBondReductionWindowStartHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingBondReductionWindowStart(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func canProposeBondReductionWindowLengthHandler(ctx snroute.Context) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeSettingBondReductionWindowLength(ctx.Command(), val)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeBondReductionWindowLengthHandler(ctx snroute.WriteContext) {
	val, err := parseUint64(ctx.Request, "value")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeSettingBondReductionWindowLength(ctx.Command(), val, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
