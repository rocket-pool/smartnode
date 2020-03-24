package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
)


// Export the node account
func exportNodeAccount(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get node account
    account, err := node.ExportNodeAccount(p)
    if err != nil { return err }

    // Print output & return
    fmt.Fprintln(p.Output, "Your node account password and keystore file are displayed below. These can be used to import your node account into another wallet:")
    fmt.Fprintln(p.Output, "")
    fmt.Fprintln(p.Output, "Password:", account.Password)
    fmt.Fprintln(p.Output, "Keystore file path:", account.KeystorePath)
    fmt.Fprintln(p.Output, "Keystore file contents:")
    fmt.Fprintln(p.Output, "-----------------------")
    fmt.Fprintln(p.Output, account.KeystoreFile)
    fmt.Fprintln(p.Output, "-----------------------")
    return nil

}

