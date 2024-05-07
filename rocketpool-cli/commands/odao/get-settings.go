package odao

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
)

func getSettings(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Get oracle DAO settings
	response, err := rp.Api.ODao.Settings()
	if err != nil {
		return err
	}

	// Member settings
	fmt.Println("=== Member Settings ===")
	fmt.Printf("ODAO Voting Quorum Threshold: %s%%\n", strconv.FormatFloat(response.Data.Member.Quorum*100, 'f', -1, 64))
	fmt.Printf("Required Member RPL Bond: %s RPL\n", strconv.FormatFloat(eth.WeiToEth(response.Data.Member.RplBond), 'f', -1, 64))
	fmt.Printf("Consecutive Challenge Cooldown: %s\n", response.Data.Member.ChallengeCooldown)
	fmt.Printf("Challenge Meeting Window: %s\n", response.Data.Member.ChallengeWindow)
	fmt.Printf("Cost for Non-members to Challenge Members: %s ETH\n", strconv.FormatFloat(eth.WeiToEth(response.Data.Member.ChallengeCost), 'f', -1, 64))
	fmt.Println()

	// Proposal settings
	fmt.Println("=== Proposal Settings ===")
	fmt.Printf("Cooldown Between Proposals: %s\n", response.Data.Proposal.Cooldown)
	fmt.Printf("Proposal Voting Window: %s\n", response.Data.Proposal.VoteTime)
	fmt.Printf("Delay Before Voting on a Proposal is Allowed: %s\n", response.Data.Proposal.VoteDelayTime)
	fmt.Printf("Window to Execute an Accepted Proposal: %s\n", response.Data.Proposal.ExecuteTime)
	fmt.Printf("Window to Act on an Executed Proposal: %s\n", response.Data.Proposal.ActionTime)
	fmt.Println()

	// Minipool settings
	fmt.Println("=== Minipool Settings ===")
	fmt.Printf("Scrub Period: %s\n", response.Data.Minipool.ScrubPeriod)
	fmt.Printf("Promotion Scrub Period: %s\n", response.Data.Minipool.PromotionScrubPeriod)
	fmt.Printf("Scrub Penalty Enabled: %t\n", response.Data.Minipool.IsScrubPenaltyEnabled)
	fmt.Printf("Bond Reduction Window Start: %s\n", response.Data.Minipool.BondReductionWindowStart)
	fmt.Printf("Bond Reduction Window Length: %s\n", response.Data.Minipool.BondReductionWindowLength)
	fmt.Println()

	return nil
}
