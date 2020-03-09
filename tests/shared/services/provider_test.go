package services

import (
    "io/ioutil"
    "testing"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test service provider functionality
func TestServiceProvider(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get test app args & options
    args := testapp.GetAppArgs(dataPath, "", "")
    appOptions := testapp.GetAppOptions(dataPath)

    // Create & configure test app
    app := cli.NewApp()
    cliutils.Configure(app)
    app.Action = func(c *cli.Context) error {

        // Provider options (node password / account optional)
        optsLight := services.ProviderOpts{
            DB: true,
            PM: true,
            AM: true,
            Client: true,
            CM: true,
            Publisher: true,
            Beacon: true,
            Docker: true,
            PasswordOptional: true,
            NodeAccountOptional: true,
        }

        // Create provider with node password / account optional
        if provider, err := services.NewProvider(c, optsLight); err != nil {
            t.Error(err)
        } else {
            provider.Cleanup()
        }

        // Provider options (all services)
        opts := services.ProviderOpts{
            DB: true,
            PM: true,
            AM: true,
            Client: true,
            CM: true,
            NodeContractAddress: true,
            NodeContract: true,
            Publisher: true,
            Beacon: true,
            Docker: true,
        }

        // Attempt to create provider without node password set
        if _, err := services.NewProvider(c, opts); err == nil { t.Error("NewProvider() should return error without node password") }

        // Set node password
        pm := passwords.NewPasswordManager(appOptions.Password)
        if err := pm.SetPassword("foobarbaz"); err != nil { return err }

        // Attempt to create provider without node account set
        if _, err := services.NewProvider(c, opts); err == nil { t.Error("NewProvider() should return error without node account") }

        // Set node account
        am := accounts.NewAccountManager(appOptions.KeychainPow, pm)
        if _, err := am.CreateNodeAccount(); err != nil { return err }

        // Attempt to create provider without RocketNodeAPI contract loaded; load contract
        if _, err := services.NewProvider(c, opts); err == nil { t.Error("NewProvider() should return error without RocketNodeAPI contract loaded") }
        opts.LoadContracts = []string{"rocketNodeAPI"}

        // Attempt to create provider for unregistered node
        if _, err := services.NewProvider(c, opts); err == nil { t.Error("NewProvider() should return error for unregistered node") }

        // Register node
        if err := testapp.AppRegisterNode(appOptions); err != nil { return err }

        // Attempt to create provider without RocketNodeContract ABI loaded; load ABI
        if _, err := services.NewProvider(c, opts); err == nil { t.Error("NewProvider() should return error without RocketNodeContract ABI loaded") }
        opts.LoadAbis = []string{"rocketNodeContract"}

        // Create provider
        if provider, err := services.NewProvider(c, opts); err != nil {
            t.Error(err)
        } else {
            provider.Cleanup()
        }

        // Return
        return nil

    }

    // Run test app
    if err := app.Run(args); err != nil { t.Fatal(err) }

}

