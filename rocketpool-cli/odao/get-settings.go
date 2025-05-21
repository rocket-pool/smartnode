package odao

import (
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getMemberSettings(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get oracle DAO member settings
	response, err := rp.GetTNDAOMemberSettings()
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("ODAO Voting Quorum Threshold: %f%%\n", response.Quorum*100)
	fmt.Printf("Required Member RPL Bond: %f RPL\n", eth.WeiToEth(response.RPLBond))
	fmt.Printf("Max Number of Unbonded Minipools: %d\n", response.MinipoolUnbondedMax)
	fmt.Printf("Consecutive Challenge Cooldown: %d Blocks\n", response.ChallengeCooldown)
	fmt.Printf("Challenge Meeting Window: %d Blocks\n", response.ChallengeWindow)
	fmt.Printf("Cost for Non-members to Challenge Members: %f ETH\n", eth.WeiToEth(response.ChallengeCost))
	return nil

}

func getProposalSettings(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get oracle DAO proposal settings
	response, err := rp.GetTNDAOProposalSettings()
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Cooldown Between Proposals: %s\n", time.Duration(response.Cooldown*1000000000))
	fmt.Printf("Proposal Voting Window: %s\n", time.Duration(response.VoteTime*1000000000))
	fmt.Printf("Delay Before Voting on a Proposal is Allowed: %s\n", time.Duration(response.VoteDelayTime*1000000000))
	fmt.Printf("Window to Execute an Accepted Proposal: %s\n", time.Duration(response.ExecuteTime*1000000000))
	fmt.Printf("Window to Act on an Executed Proposal: %s\n", time.Duration(response.ActionTime*1000000000))
	return nil

}

func getMinipoolSettings(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get oracle DAO proposal settings
	response, err := rp.GetTNDAOMinipoolSettings()
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Scrub Period: %s\n", time.Duration(response.ScrubPeriod*1000000000))
	fmt.Printf("Promotion Scrub Period: %s\n", time.Duration(response.PromotionScrubPeriod*1000000000))
	fmt.Printf("Scrub Penalty Enabled: %t\n", response.ScrubPenaltyEnabled)
	fmt.Printf("Bond Reduction Window Start: %s\n", time.Duration(response.BondReductionWindowStart*1000000000))
	fmt.Printf("Bond Reduction Window Length: %s\n", time.Duration(response.BondReductionWindowLength*1000000000))
	return nil

}
