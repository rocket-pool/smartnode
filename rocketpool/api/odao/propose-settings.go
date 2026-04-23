package odao

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"

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

	gasInfo, err := trustednode.EstimateProposeQuorumGas(rp, quorum, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeRPLBondGas(rp, bondAmountWei, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

func canProposeSettingMinipoolUnbondedMax(c *cli.Command, unbondedMinipoolMax uint64) (*api.CanProposeTNDAOSettingResponse, error) {

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

func proposeSettingMinipoolUnbondedMax(c *cli.Command, unbondedMinipoolMax uint64, opts *bind.TransactOpts) (*api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse{}

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

	gasInfo, err := trustednode.EstimateProposeProposalCooldownTimeGas(rp, proposalCooldownTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeProposalVoteTimeGas(rp, proposalVoteTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeProposalVoteDelayTimeGas(rp, proposalDelayTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeProposalExecuteTimeGas(rp, proposalExecuteTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeProposalActionTimeGas(rp, proposalActionTimespan, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeScrubPeriodGas(rp, scrubPeriod, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposePromotionScrubPeriodGas(rp, promotionScrubPeriod, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeScrubPenaltyEnabledGas(rp, enabled, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeBondReductionWindowStartGas(rp, bondReductionWindowStart, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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

	gasInfo, err := trustednode.EstimateProposeBondReductionWindowLengthGas(rp, bondReductionWindowLength, opts)
	if err != nil {
		return nil, err
	}

	response.GasInfo = gasInfo
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
