package odao

import (
	"fmt"
	"time"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeSettingMembersQuorum(c *cli.Context, quorumPercent float64) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingMembersQuorum(quorumPercent / 100)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingMembersQuorum(quorumPercent / 100)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a members.quorum setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingMembersRplBond(c *cli.Context, bondAmountEth float64) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingMembersRplBond(eth.EthToWei(bondAmountEth))
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingMembersRplBond(eth.EthToWei(bondAmountEth))
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a members.rplbond setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingMinipoolUnbondedMax(c *cli.Context, unbondedMinipoolMax uint64) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingMinipoolUnbondedMax(unbondedMinipoolMax)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingMinipoolUnbondedMax(unbondedMinipoolMax)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a members.minipool.unbonded.max setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingProposalCooldown(c *cli.Context, proposalCooldownTimespan string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(proposalCooldownTimespan)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingProposalCooldown(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingProposalCooldown(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a proposal.cooldown setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingProposalVoteTimespan(c *cli.Context, proposalVoteTimespan string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(proposalVoteTimespan)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingProposalVoteTimespan(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingProposalVoteTimespan(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a proposal.vote.time setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingProposalVoteDelayTimespan(c *cli.Context, proposalDelayTimespan string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(proposalDelayTimespan)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingProposalVoteDelayTimespan(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingProposalVoteDelayTimespan(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a proposal.vote.delay.time setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingProposalExecuteTimespan(c *cli.Context, proposalExecuteTimespan string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(proposalExecuteTimespan)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingProposalExecuteTimespan(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingProposalExecuteTimespan(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a proposal.execute.time setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingProposalActionTimespan(c *cli.Context, proposalActionTimespan string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(proposalActionTimespan)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingProposalActionTimespan(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingProposalActionTimespan(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a proposal.action.time setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingScrubPeriod(c *cli.Context, scrubPeriod string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(scrubPeriod)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingScrubPeriod(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingScrubPeriod(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a minipool.scrub.period setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingPromotionScrubPeriod(c *cli.Context, scrubPeriod string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(scrubPeriod)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingPromotionScrubPeriod(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingPromotionScrubPeriod(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a minipool.promotion.scrub.period setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingScrubPenaltyEnabled(c *cli.Context, enabled bool) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingScrubPenaltyEnabled(enabled)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingScrubPenaltyEnabled(enabled)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a minipool.scrub.penalty.enabled setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingBondReductionWindowStart(c *cli.Context, windowStart string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(windowStart)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingBondReductionWindowStart(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingBondReductionWindowStart(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a minipool.bond.reduction.window.start setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}

func proposeSettingBondReductionWindowLength(c *cli.Context, windowLength string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Parse the timespan
	timespan, err := time.ParseDuration(windowLength)
	if err != nil {
		return fmt.Errorf("Error parsing time: %w\n", err)
	}
	seconds := uint64(timespan.Seconds())

	// Check if proposal can be made
	canPropose, err := rp.CanProposeTNDAOSettingBondReductionWindowLength(seconds)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose setting update:")
		if canPropose.ProposalCooldownActive {
			fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.ProposeTNDAOSettingBondReductionWindowLength(seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted a minipool.bond.reduction.window.length setting update proposal with ID %d.\n", response.ProposalId)
	return nil

}
