package pdao

import (
	"fmt"

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

	// Get all PDAO settings
	response, err := rp.Api.PDao.Settings()
	if err != nil {
		return err
	}

	// Auction
	fmt.Println("== Auction Settings ==")
	fmt.Printf("\tCreating New Lot Enabled: %t\n", response.Data.Auction.IsCreateLotEnabled)
	fmt.Printf("\tBidding on Lots Enabled:  %t\n", response.Data.Auction.IsBidOnLotEnabled)
	fmt.Printf("\tMin ETH per Lot:          %.6f ETH\n", eth.WeiToEth(response.Data.Auction.LotMinimumEthValue))
	fmt.Printf("\tMax ETH per Lot:          %.6f ETH\n", eth.WeiToEth(response.Data.Auction.LotMaximumEthValue))
	fmt.Printf("\tLot Duration:             %s\n", response.Data.Auction.LotDuration)
	fmt.Printf("\tStarting Price Ratio:     %.2f%%\n", eth.WeiToEth(response.Data.Auction.LotStartingPriceRatio)*100)
	fmt.Printf("\tReserve Price Ratio:      %.2f%%\n", eth.WeiToEth(response.Data.Auction.LotReservePriceRatio)*100)
	fmt.Println()

	// Deposit
	fmt.Println("== Deposit Settings ==")
	fmt.Printf("\tPool Deposits Enabled:              %t\n", response.Data.Deposit.IsDepositingEnabled)
	fmt.Printf("\tDeposit Assignments Enabled:        %t\n", response.Data.Deposit.AreDepositAssignmentsEnabled)
	fmt.Printf("\tMin Pool Deposit:                   %.6f ETH\n", eth.WeiToEth(response.Data.Deposit.MinimumDeposit))
	fmt.Printf("\tMax Deposit Pool Size:              %.6f ETH\n", eth.WeiToEth(response.Data.Deposit.MaximumDepositPoolSize))
	fmt.Printf("\tMax Total Assigns Per Deposit:      %d\n", response.Data.Deposit.MaximumAssignmentsPerDeposit)
	fmt.Printf("\tMax Socialized Assigns Per Deposit: %d\n", response.Data.Deposit.MaximumSocialisedAssignmentsPerDeposit)
	fmt.Printf("\tDeposit Fee:                        %.2f%%\n", eth.WeiToEth(response.Data.Deposit.DepositFee)*100)
	fmt.Println()

	// Inflation
	fmt.Println("== Inflation Settings ==")
	fmt.Printf("\tInterval Rate:  %.6f\n", eth.WeiToEth(response.Data.Inflation.IntervalRate))
	fmt.Printf("\tInterval Start: %s\n", response.Data.Inflation.StartTime)
	fmt.Println()

	// Minipool
	fmt.Println("== Minipool Settings ==")
	fmt.Printf("\tMark as Withdrawable Enabled: %t\n", response.Data.Minipool.IsSubmitWithdrawableEnabled)
	fmt.Printf("\tStaking Launch Timeout:       %s\n", response.Data.Minipool.LaunchTimeout)
	fmt.Printf("\tBond Reduction Enabled:       %t\n", response.Data.Minipool.IsBondReductionEnabled)
	fmt.Printf("\tMax Number of Minipools:      %d\n", response.Data.Minipool.MaximumCount)
	fmt.Printf("\tUser Distribute Start Wait:   %s\n", response.Data.Minipool.UserDistributeWindowStart)
	fmt.Printf("\tUser Distribute Window:       %s\n", response.Data.Minipool.UserDistributeWindowLength)
	fmt.Println()

	// Network
	fmt.Println("== Network Settings ==")
	fmt.Printf("\toDAO Consensus Quorum:      %.2f%%\n", eth.WeiToEth(response.Data.Network.OracleDaoConsensusThreshold)*100)
	fmt.Printf("\tNode Penalty Quorum:        %.2f%%\n", eth.WeiToEth(response.Data.Network.NodePenaltyThreshold)*100)
	fmt.Printf("\tPenalty Size:               %.2f%%\n", eth.WeiToEth(response.Data.Network.PerPenaltyRate)*100)
	fmt.Printf("\tBalance Submission Enabled: %t\n", response.Data.Network.IsSubmitBalancesEnabled)
	fmt.Printf("\tBalance Submission Freq:    %s\n", response.Data.Network.SubmitBalancesFrequency)
	fmt.Printf("\tPrice Submission Enabled:   %t\n", response.Data.Network.IsSubmitPricesEnabled)
	fmt.Printf("\tPrice Submission Freq:      %s\n", response.Data.Network.SubmitPricesFrequency)
	fmt.Printf("\tMin Commission:             %.2f%%\n", eth.WeiToEth(response.Data.Network.MinimumNodeFee)*100)
	fmt.Printf("\tTarget Commission:          %.2f%%\n", eth.WeiToEth(response.Data.Network.TargetNodeFee)*100)
	fmt.Printf("\tMax Commission:             %.2f%%\n", eth.WeiToEth(response.Data.Network.MaximumNodeFee)*100)
	fmt.Printf("\tCommission Demand Range:    %.6f ETH\n", eth.WeiToEth(response.Data.Network.NodeFeeDemandRange))
	fmt.Printf("\trETH Collateral Target:     %.2f%%\n", eth.WeiToEth(response.Data.Network.TargetRethCollateralRate)*100)
	fmt.Printf("\tRewards Submission Enabled: %t\n", response.Data.Network.IsSubmitRewardsEnabled)
	fmt.Println()

	// Node
	fmt.Println("== Node Settings ==")
	fmt.Printf("\tRegistration Enabled:          %t\n", response.Data.Node.IsRegistrationEnabled)
	fmt.Printf("\tSmoothing Pool Opt-In Enabled: %t\n", response.Data.Node.IsSmoothingPoolRegistrationEnabled)
	fmt.Printf("\tNode Deposits Enabled:         %t\n", response.Data.Node.IsDepositingEnabled)
	fmt.Printf("\tVacant Minipools Enabled:      %t\n", response.Data.Node.AreVacantMinipoolsEnabled)
	fmt.Printf("\tMin Stake per Minipool:        %.2f%%\n", eth.WeiToEth(response.Data.Node.MinimumPerMinipoolStake)*100)
	fmt.Printf("\tMax Stake per Minipool:        %.2f%%\n", eth.WeiToEth(response.Data.Node.MaximumPerMinipoolStake)*100)
	fmt.Println()

	// Proposals
	fmt.Println("== Proposal Settings ==")
	fmt.Printf("\tVoting Window (Phase 1): %s\n", response.Data.Proposals.VotePhase1Time)
	fmt.Printf("\tVoting Window (Phase 2): %s\n", response.Data.Proposals.VotePhase2Time)
	fmt.Printf("\tVoting Start Delay:      %s\n", response.Data.Proposals.VoteDelayTime)
	fmt.Printf("\tExecute Window:          %s\n", response.Data.Proposals.ExecuteTime)
	fmt.Printf("\tBond per Proposal:       %.6f RPL\n", eth.WeiToEth(response.Data.Proposals.ProposalBond))
	fmt.Printf("\tBond per Challenge:      %.6f RPL\n", eth.WeiToEth(response.Data.Proposals.ChallengeBond))
	fmt.Printf("\tChallenge Response Time: %s\n", response.Data.Proposals.ChallengePeriod)
	fmt.Printf("\tQuorum:                  %.2f%%\n", eth.WeiToEth(response.Data.Proposals.Quorum)*100)
	fmt.Printf("\tVeto Quorum:             %.2f%%\n", eth.WeiToEth(response.Data.Proposals.VetoQuorum)*100)
	fmt.Printf("\tTarget Block Age Limit:  %d Blocks\n", response.Data.Proposals.MaxBlockAge)
	fmt.Println()

	// Rewards
	fmt.Println("== Rewards Settings ==")
	fmt.Printf("\tInterval Length: %s\n", response.Data.Rewards.IntervalTime)
	fmt.Println()

	// Security
	fmt.Println("== Security Settings ==")
	fmt.Printf("\tMember Quorum:         %.2f%%\n", eth.WeiToEth(response.Data.Security.MembersQuorum)*100)
	fmt.Printf("\tMember Leave Time:     %s\n", response.Data.Security.MembersLeaveTime)
	fmt.Printf("\tProposal Vote Time:    %s\n", response.Data.Security.ProposalVoteTime)
	fmt.Printf("\tProposal Execute Time: %s\n", response.Data.Security.ProposalExecuteTime)
	fmt.Printf("\tProposal Action Time:  %s\n", response.Data.Security.ProposalActionTime)

	return nil
}
