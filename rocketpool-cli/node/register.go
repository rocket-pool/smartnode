package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func registerNode(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check node can be registered
    canRegister, err := rp.CanRegisterNode()
    if err != nil {
        return err
    }
    if !canRegister.CanRegister {
        fmt.Println("The node cannot be registered:")
        if canRegister.AlreadyRegistered {
            fmt.Println("The node is already registered with Rocket Pool.")
        }
        if canRegister.RegistrationDisabled {
            fmt.Println("Node registrations are currently disabled.")
        }
        return nil
    }

    // Prompt for timezone location
    var timezoneLocation string
    if c.String("timezone") != "" {
        timezoneLocation = c.String("timezone")
    } else {
        timezoneLocation = promptTimezone()
    }

    // Register node
    if _, err := rp.RegisterNode(timezoneLocation); err != nil {
        return err
    }

    // Log & return
    fmt.Println("The node was successfully registered with Rocket Pool.")
    return nil

}

