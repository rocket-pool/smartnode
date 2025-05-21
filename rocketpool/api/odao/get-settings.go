package odao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMemberSettings(c *cli.Context) (*api.GetTNDAOMemberSettingsResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetTNDAOMemberSettingsResponse{}

	quorum, err := trustednode.GetQuorum(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting quorum: %w", err)
	}

	rplBond, err := trustednode.GetRPLBond(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting RPL Bond: %w", err)
	}

	minipoolUnbondedMax, err := trustednode.GetMinipoolUnbondedMax(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting minipool unbonded max: %w", err)
	}

	challengeCooldown, err := trustednode.GetChallengeCooldown(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting challenge cooldown: %w", err)
	}

	challengeWindow, err := trustednode.GetChallengeWindow(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting challenge window: %w", err)
	}

	challengeCost, err := trustednode.GetChallengeCost(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting challenge cost: %w", err)
	}

	response.Quorum = quorum
	response.RPLBond = rplBond
	response.MinipoolUnbondedMax = minipoolUnbondedMax
	response.ChallengeCooldown = challengeCooldown
	response.ChallengeWindow = challengeWindow
	response.ChallengeCost = challengeCost

	// Return response
	return &response, nil
}

func getProposalSettings(c *cli.Context) (*api.GetTNDAOProposalSettingsResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetTNDAOProposalSettingsResponse{}

	cooldown, err := trustednode.GetProposalCooldownTime(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting proposal cooldown: %w", err)
	}

	voteTime, err := trustednode.GetProposalVoteTime(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting proposal vote time: %w", err)
	}

	voteDelayTime, err := trustednode.GetProposalVoteDelayTime(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting proposal vote delay time: %w", err)
	}

	executeTime, err := trustednode.GetProposalExecuteTime(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting proposal execute time: %w", err)
	}

	actionTime, err := trustednode.GetProposalActionTime(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting proposal action time: %w", err)
	}

	response.Cooldown = cooldown
	response.VoteTime = voteTime
	response.VoteDelayTime = voteDelayTime
	response.ExecuteTime = executeTime
	response.ActionTime = actionTime

	// Return response
	return &response, nil
}

func getMinipoolSettings(c *cli.Context) (*api.GetTNDAOMinipoolSettingsResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetTNDAOMinipoolSettingsResponse{}

	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting scrub period: %w", err)
	}
	promotionScrubPeriod, err := trustednode.GetPromotionScrubPeriod(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting promotion scrub period: %w", err)
	}
	scrubPenaltyEnabled, err := trustednode.GetScrubPenaltyEnabled(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting scrub penalty flag: %w", err)
	}
	bondReductionWindowStart, err := trustednode.GetBondReductionWindowStart(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting bond reduction window start: %w", err)
	}
	bondReductionWindowLength, err := trustednode.GetBondReductionWindowLength(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting bond reduction window length: %w", err)
	}

	response.ScrubPeriod = scrubPeriod
	response.PromotionScrubPeriod = promotionScrubPeriod
	response.ScrubPenaltyEnabled = scrubPenaltyEnabled
	response.BondReductionWindowStart = bondReductionWindowStart
	response.BondReductionWindowLength = bondReductionWindowLength

	// Return response
	return &response, nil

}
