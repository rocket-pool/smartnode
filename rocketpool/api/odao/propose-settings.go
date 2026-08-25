package odao

import (
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

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

func proposeSettingMembersQuorum(c *cli.Command, quorum float64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingMembersQuorumResponse, error) {

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

func proposeSettingMembersRplBond(c *cli.Command, bondAmountWei *big.Int, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingMembersRplBondResponse, error) {

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

func proposeSettingProposalCooldown(c *cli.Command, proposalCooldownTimespan uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingProposalCooldownResponse, error) {

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

func proposeSettingProposalVoteTimespan(c *cli.Command, proposalVoteTimespan uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingProposalVoteTimespanResponse, error) {

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

func proposeSettingProposalVoteDelayTimespan(c *cli.Command, proposalDelayTimespan uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse, error) {

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

func proposeSettingProposalExecuteTimespan(c *cli.Command, proposalExecuteTimespan uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingProposalExecuteTimespanResponse, error) {

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

func proposeSettingProposalActionTimespan(c *cli.Command, proposalActionTimespan uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingProposalActionTimespanResponse, error) {

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

func proposeSettingScrubPeriod(c *cli.Command, scrubPeriod uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

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

func proposeSettingPromotionScrubPeriod(c *cli.Command, promotionScrubPeriod uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingPromotionScrubPeriodResponse, error) {

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

func proposeSettingScrubPenaltyEnabled(c *cli.Command, enabled bool, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

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

func proposeSettingBondReductionWindowStart(c *cli.Command, bondReductionWindowStart uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

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

func proposeSettingBondReductionWindowLength(c *cli.Command, bondReductionWindowLength uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingScrubPeriodResponse, error) {

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

func canProposeMembersQuorumHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quorum, err := parseFloat64(r, "quorum")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingMembersQuorum(c, quorum)
		response.WriteResponse(w, resp, err)
	}
}

func proposeMembersQuorumHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quorum, err := parseFloat64(r, "quorum")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingMembersQuorum(c, quorum, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeMembersRplbondHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bond, err := parseBigInt(r, "bondAmountWei")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingMembersRplBond(c, bond)
		response.WriteResponse(w, resp, err)
	}
}

func proposeMembersRplbondHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bond, err := parseBigInt(r, "bondAmountWei")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingMembersRplBond(c, bond, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeProposalCooldownHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalCooldown(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeProposalCooldownHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalCooldown(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeProposalVoteTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalVoteTimespan(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeProposalVoteTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalVoteTimespan(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeProposalVoteDelayTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalVoteDelayTimespan(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeProposalVoteDelayTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalVoteDelayTimespan(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeProposalExecuteTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalExecuteTimespan(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeProposalExecuteTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalExecuteTimespan(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeProposalActionTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingProposalActionTimespan(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeProposalActionTimespanHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingProposalActionTimespan(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeScrubPeriodHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingScrubPeriod(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeScrubPeriodHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingScrubPeriod(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposePromotionScrubPeriodHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingPromotionScrubPeriod(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposePromotionScrubPeriodHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingPromotionScrubPeriod(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeScrubPenaltyEnabledHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enabledStr := r.URL.Query().Get("enabled")
		resp, err := canProposeSettingScrubPenaltyEnabled(c, enabledStr == "true")
		response.WriteResponse(w, resp, err)
	}
}

func proposeScrubPenaltyEnabledHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enabledStr := r.FormValue("enabled")
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingScrubPenaltyEnabled(c, enabledStr == "true", opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeBondReductionWindowStartHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingBondReductionWindowStart(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeBondReductionWindowStartHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingBondReductionWindowStart(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}

func canProposeBondReductionWindowLengthHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeSettingBondReductionWindowLength(c, val)
		response.WriteResponse(w, resp, err)
	}
}

func proposeBondReductionWindowLengthHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := parseUint64(r, "value")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeSettingBondReductionWindowLength(c, val, opts)
		response.WriteResponse(w, resp, err)
	}
}
