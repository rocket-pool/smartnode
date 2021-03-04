package tndao

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func join(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if node can join the trusted node DAO
    canJoin, err := rp.CanJoinTNDAO()
    if err != nil {
        return err
    }
    if !canJoin.CanJoin {
        fmt.Println("Cannot join the trusted node DAO:")
        if canJoin.ProposalExpired {
            fmt.Println("The proposal for you to join the trusted node DAO does not exist or has expired.")
        }
        if canJoin.AlreadyMember {
            fmt.Println("The node is already a member of the trusted node DAO.")
        }
        if canJoin.InsufficientRplBalance {
            fmt.Println("The node does not have enough RPL to pay the trusted node RPL bond.")
        }
        return nil
    }

    // Join the trusted node DAO
    if _, err := rp.JoinTNDAO(); err != nil {
        return err
    }

    // Log & return
    fmt.Println("Successfully joined the trusted node DAO.")
    return nil

}

