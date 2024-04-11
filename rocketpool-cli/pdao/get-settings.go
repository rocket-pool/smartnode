package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getSettings(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Get all PDAO settings
	response, err := rp.PDAOGetSettings()
	if err != nil {
		return err
	}

	// Auction
	fmt.Println("== Auction Settings ==")
	fmt.Printf("\tCreating New Lot Enabled: %t\n", response.Auction.IsCreateLotEnabled)
	fmt.Printf("\tBidding on Lots Enabled:  %t\n", response.Auction.IsBidOnLotEnabled)
	fmt.Printf("\tMin ETH per Lot:          %.6f ETH\n", eth.WeiToEth(response.Auction.LotMinimumEthValue))
	fmt.Printf("\tMax ETH per Lot:          %.6f ETH\n", eth.WeiToEth(response.Auction.LotMaximumEthValue))
	fmt.Printf("\tLot Duration:             %s\n", response.Auction.LotDuration)
	fmt.Printf("\tStarting Price Ratio:     %.2f%%\n", eth.WeiToEth(response.Auction.LotStartingPriceRatio)*100)
	fmt.Printf("\tReserve Price Ratio:      %.2f%%\n", eth.WeiToEth(response.Auction.LotReservePriceRatio)*100)
	fmt.Println()

	// Deposit
	fmt.Println("== Deposit Settings ==")
	fmt.Printf("\tPool Deposits Enabled:              %t\n", response.Deposit.IsDepositingEnabled)
	fmt.Printf("\tDeposit Assignments Enabled:        %t\n", response.Deposit.AreDepositAssignmentsEnabled)
	fmt.Printf("\tMin Pool Deposit:                   %.6f ETH\n", eth.WeiToEth(response.Deposit.MinimumDeposit))
	fmt.Printf("\tMax Deposit Pool Size:              %.6f ETH\n", eth.WeiToEth(response.Deposit.MaximumDepositPoolSize))
	fmt.Printf("\tMax Total Assigns Per Deposit:      %d\n", response.Deposit.MaximumAssignmentsPerDeposit)
	fmt.Printf("\tMax Socialized Assigns Per Deposit: %d\n", response.Deposit.MaximumSocialisedAssignmentsPerDeposit)
	fmt.Printf("\tDeposit Fee:                        %.2f%%\n", eth.WeiToEth(response.Deposit.DepositFee)*100)
	fmt.Println()

	// Inflation
	fmt.Println("== Inflation Settings ==")
	fmt.Printf("\tInterval Rate:  %.6f\n", eth.WeiToEth(response.Inflation.IntervalRate))
	fmt.Printf("\tInterval Start: %s\n", response.Inflation.StartTime)
	fmt.Println()

	// Minipool
	fmt.Println("== Minipool Settings ==")
	fmt.Printf("\tMark as Withdrawable Enabled: %t\n", response.Minipool.IsSubmitWithdrawableEnabled)
	fmt.Printf("\tStaking Launch Timeout:       %s\n", response.Minipool.LaunchTimeout)
	fmt.Printf("\tBond Reduction Enabled:       %t\n", response.Minipool.IsBondReductionEnabled)
	fmt.Printf("\tMax Number of Minipools:      %d\n", response.Minipool.MaximumCount)
	fmt.Printf("\tUser Distribute Start Wait:   %s\n", response.Minipool.UserDistributeWindowStart)
	fmt.Printf("\tUser Distribute Window:       %s\n", response.Minipool.UserDistributeWindowLength)
	fmt.Println()

	// Network
	fmt.Println("== Network Settings ==")
	fmt.Printf("\toDAO Consensus Quorum:      %.2f%%\n", eth.WeiToEth(response.Network.OracleDaoConsensusThreshold)*100)
	fmt.Printf("\tNode Penalty Quorum:        %.2f%%\n", eth.WeiToEth(response.Network.NodePenaltyThreshold)*100)
	fmt.Printf("\tPenalty Size:               %.2f%%\n", eth.WeiToEth(response.Network.PerPenaltyRate)*100)
	fmt.Printf("\tBalance Submission Enabled: %t\n", response.Network.IsSubmitBalancesEnabled)
	fmt.Printf("\tBalance Submission Freq:    %s\n", response.Network.SubmitBalancesFrequency)
	fmt.Printf("\tPrice Submission Enabled:   %t\n", response.Network.IsSubmitPricesEnabled)
	fmt.Printf("\tPrice Submission Freq:      %s\n", response.Network.SubmitPricesFrequency)
	fmt.Printf("\tMin Commission:             %.2f%%\n", eth.WeiToEth(response.Network.MinimumNodeFee)*100)
	fmt.Printf("\tTarget Commission:          %.2f%%\n", eth.WeiToEth(response.Network.TargetNodeFee)*100)
	fmt.Printf("\tMax Commission:             %.2f%%\n", eth.WeiToEth(response.Network.MaximumNodeFee)*100)
	fmt.Printf("\tCommission Demand Range:    %.6f ETH\n", eth.WeiToEth(response.Network.NodeFeeDemandRange))
	fmt.Printf("\trETH Collateral Target:     %.2f%%\n", eth.WeiToEth(response.Network.TargetRethCollateralRate)*100)
	fmt.Printf("\tRewards Submission Enabled: %t\n", response.Network.IsSubmitRewardsEnabled)
	fmt.Println()

	// Node
	fmt.Println("== Node Settings ==")
	fmt.Printf("\tRegistration Enabled:          %t\n", response.Node.IsRegistrationEnabled)
	fmt.Printf("\tSmoothing Pool Opt-In Enabled: %t\n", response.Node.IsSmoothingPoolRegistrationEnabled)
	fmt.Printf("\tNode Deposits Enabled:         %t\n", response.Node.IsDepositingEnabled)
	fmt.Printf("\tVacant Minipools Enabled:      %t\n", response.Node.AreVacantMinipoolsEnabled)
	fmt.Printf("\tMin Stake per Minipool:        %.2f%%\n", eth.WeiToEth(response.Node.MinimumPerMinipoolStake)*100)
	fmt.Printf("\tMax Stake per Minipool:        %.2f%%\n", eth.WeiToEth(response.Node.MaximumPerMinipoolStake)*100)
	fmt.Println()

	// Proposals
	fmt.Println("== Proposal Settings ==")
	fmt.Printf("\tVoting Window (Phase 1): %s\n", response.Proposals.VotePhase1Time)
	fmt.Printf("\tVoting Window (Phase 2): %s\n", response.Proposals.VotePhase2Time)
	fmt.Printf("\tVoting Start Delay:      %s\n", response.Proposals.VoteDelayTime)
	fmt.Printf("\tExecute Window:          %s\n", response.Proposals.ExecuteTime)
	fmt.Printf("\tBond per Proposal:       %.6f RPL\n", eth.WeiToEth(response.Proposals.ProposalBond))
	fmt.Printf("\tBond per Challenge:      %.6f RPL\n", eth.WeiToEth(response.Proposals.ChallengeBond))
	fmt.Printf("\tChallenge Response Time: %s\n", response.Proposals.ChallengePeriod)
	fmt.Printf("\tQuorum:                  %.2f%%\n", eth.WeiToEth(response.Proposals.Quorum)*100)
	fmt.Printf("\tVeto Quorum:             %.2f%%\n", eth.WeiToEth(response.Proposals.VetoQuorum)*100)
	fmt.Printf("\tTarget Block Age Limit:  %d Blocks\n", response.Proposals.MaxBlockAge)
	fmt.Println()

	// Rewards
	fmt.Println("== Rewards Settings ==")
	fmt.Printf("\tInterval Length: %s\n", response.Rewards.IntervalTime)
	fmt.Println()

	// Security
	fmt.Println("== Security Settings ==")
	fmt.Printf("\tMember Quorum:         %.2f%%\n", eth.WeiToEth(response.Security.MembersQuorum)*100)
	fmt.Printf("\tMember Leave Time:     %s\n", response.Security.MembersLeaveTime)
	fmt.Printf("\tProposal Vote Time:    %s\n", response.Security.ProposalVoteTime)
	fmt.Printf("\tProposal Execute Time: %s\n", response.Security.ProposalExecuteTime)
	fmt.Printf("\tProposal Action Time:  %s\n", response.Security.ProposalActionTime)

	return nil
}
