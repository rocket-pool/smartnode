package odao

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func proposeSettingMembersQuorum(c *cli.Context, quorumPercent float64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingMembersQuorum(quorumPercent / 100)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a members.quorum setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingMembersRplBond(c *cli.Context, bondAmountEth float64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingMembersRplBond(eth.EthToWei(bondAmountEth))
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a members.rplbond setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingMinipoolUnbondedMax(c *cli.Context, unbondedMinipoolMax uint64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingMinipoolUnbondedMax(unbondedMinipoolMax)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a members.minipool.unbonded.max setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingProposalCooldown(c *cli.Context, proposalCooldownBlocks uint64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingProposalCooldown(proposalCooldownBlocks)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a proposal.cooldown setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingProposalVoteBlocks(c *cli.Context, proposalVoteBlocks uint64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingProposalVoteBlocks(proposalVoteBlocks)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a proposal.vote.blocks setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingProposalVoteDelayBlocks(c *cli.Context, proposalDelayBlocks uint64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingProposalVoteDelayBlocks(proposalDelayBlocks)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a proposal.vote.delay.blocks setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingProposalExecuteBlocks(c *cli.Context, proposalExecuteBlocks uint64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingProposalExecuteBlocks(proposalExecuteBlocks)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a proposal.execute.blocks setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}


func proposeSettingProposalActionBlocks(c *cli.Context, proposalActionBlocks uint64) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeTNDAOSetting()
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

    // Submit proposal
    response, err := rp.ProposeTNDAOSettingProposalActionBlocks(proposalActionBlocks)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a proposal.action.blocks setting update proposal with ID %d.\n", response.ProposalId)
    return nil

}

