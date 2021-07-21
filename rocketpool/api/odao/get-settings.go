package odao

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func getMemberSettings(c *cli.Context) (*api.GetTNDAOMemberSettingsResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.GetTNDAOMemberSettingsResponse{}

    quorum, err := trustednode.GetQuorum(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting quorum: %w", err)
    }

    rplBond, err := trustednode.GetRPLBond(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting RPL Bond: %w", err)
    }

    minipoolUnbondedMax, err := trustednode.GetMinipoolUnbondedMax(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting minipool unbonded max: %w", err)
    }

    challengeCooldown, err := trustednode.GetChallengeCooldown(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting challenge cooldown: %w", err)
    }

    challengeWindow, err := trustednode.GetChallengeWindow(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting challenge window: %w", err)
    }

    challengeCost, err := trustednode.GetChallengeCost(rp, nil)
    if(err != nil) {
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
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.GetTNDAOProposalSettingsResponse{}

    cooldown, err := trustednode.GetProposalCooldown(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting proposal cooldown time: %w", err)
    }

    voteTime, err := trustednode.GetProposalVoteTime(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting proposal vote time: %w", err)
    }

    voteDelayTime, err := trustednode.GetProposalVoteDelayTime(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting proposal vote delay time: %w", err)
    }

    executeTime, err := trustednode.GetProposalExecuteTime(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting proposal execute time: %w", err)
    }

    actionTime, err := trustednode.GetProposalActionTime(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting proposal action time: %w", err)
    }

    response.CooldownTime = cooldown
    response.VoteTime = voteTime
    response.VoteDelayTime = voteDelayTime
    response.ExecuteTime = executeTime
    response.ActionTime = actionTime
    
    // Return response
    return &response, nil
}

