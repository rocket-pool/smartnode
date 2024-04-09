package odao

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
)

func getSettings(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get oracle DAO settings
	response, err := rp.Api.ODao.Settings()
	if err != nil {
		return err
	}

	// Member settings
	fmt.Printf("ODAO Voting Quorum Threshold: %f%%\n", response.Data.Member.Quorum*100)
	fmt.Printf("Required Member RPL Bond: %f RPL\n", eth.WeiToEth(response.Data.Member.RplBond))
	fmt.Printf("Consecutive Challenge Cooldown: %d Blocks\n", response.Data.Member.ChallengeCooldown)
	fmt.Printf("Challenge Meeting Window: %d Blocks\n", response.Data.Member.ChallengeWindow)
	fmt.Printf("Cost for Non-members to Challenge Members: %f ETH\n", eth.WeiToEth(response.Data.Member.ChallengeCost))

	// Proposal settings
	fmt.Printf("Cooldown Between Proposals: %s\n", time.Duration(response.Data.Proposal.Cooldown))
	fmt.Printf("Proposal Voting Window: %s\n", time.Duration(response.Data.Proposal.VoteTime)*time.Second)
	fmt.Printf("Delay Before Voting on a Proposal is Allowed: %s\n", time.Duration(response.Data.Proposal.VoteDelayTime)*time.Second)
	fmt.Printf("Window to Execute an Accepted Proposal: %s\n", time.Duration(response.Data.Proposal.ExecuteTime)*time.Second)
	fmt.Printf("Window to Act on an Executed Proposal: %s\n", time.Duration(response.Data.Proposal.ActionTime)*time.Second)

	// Minipool settings
	fmt.Printf("Scrub Period: %s\n", time.Duration(response.Data.Minipool.ScrubPeriod)*time.Second)
	fmt.Printf("Promotion Scrub Period: %s\n", time.Duration(response.Data.Minipool.PromotionScrubPeriod)*time.Second)
	fmt.Printf("Scrub Penalty Enabled: %t\n", response.Data.Minipool.IsScrubPenaltyEnabled)
	fmt.Printf("Bond Reduction Window Start: %s\n", time.Duration(response.Data.Minipool.BondReductionWindowStart)*time.Second)
	fmt.Printf("Bond Reduction Window Length: %s\n", time.Duration(response.Data.Minipool.BondReductionWindowLength)*time.Second)

	return nil
}
