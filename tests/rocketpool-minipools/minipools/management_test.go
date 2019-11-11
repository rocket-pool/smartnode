package minipools

import (
    "io/ioutil"
    "os"
    "testing"
    "time"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool-minipools/minipools"
    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test minipool management functionality
func TestMinipoolManagement(t *testing.T) {

    // Create temporary output file
    output, err := ioutil.TempFile("", "")
    if err != nil { t.Fatal(err) }
    output.Close()

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Get test app args & options
    args := testapp.GetAppArgs(dataPath, "", output.Name())
    appOptions := testapp.GetAppOptions(dataPath)

    // Create & configure test app
    app := cli.NewApp()
    cliutils.Configure(app)
    app.Action = func(c *cli.Context) error {

        // Initialise and register node
        if err := testapp.AppInitNode(appOptions); err != nil { return err }
        if err := testapp.AppRegisterNode(appOptions); err != nil { return err }

        // Create and launch minipools under node
        minipoolAddresses, err := testapp.AppCreateNodeMinipools(appOptions, "3m", 1)
        if err != nil { return err }
        if _, accessorAddress, err := testapp.AppCreateGroupAccessor(appOptions); err != nil {
            return err
        } else if err := testapp.AppStakeAllMinipools(appOptions, "3m", accessorAddress); err != nil {
            return err
        }

        // Initialise services
        p, err := services.NewProvider(c, services.ProviderOpts{
            AM: true,
            CM: true,
            Docker: true,
            LoadContracts: []string{"utilAddressSetStorage"},
            LoadAbis: []string{"rocketMinipool"},
        })
        if err != nil { return err }

        // Start minipools management process
        go minipools.StartManagementProcess(p, os.Getenv("HOME") + "/.rptest", "rocketpool/smartnode-minipool:" + test.IMAGE_VERSION, "rocketpool_minipool_", "none")

        // Allow time to launch minipool containers
        time.Sleep(5 * time.Second)

        // Check output
        prefix := "(?i)^\\d{4}/\\d{2}/\\d{2} \\d{2}:\\d{2}:\\d{2} "
        if messages, err := testapp.CheckOutput(output.Name(), []string{prefix + "Creating minipool container", prefix + "Starting minipool container"}, map[int][]string{
            1: []string{prefix + "Created minipool container rocketpool_minipool_" + minipoolAddresses[0].Hex() + " successfully$", "Container created message incorrect"},
            2: []string{prefix + "Started minipool container rocketpool_minipool_" + minipoolAddresses[0].Hex() + " successfully$", "Container started message incorrect"},
        }); err != nil {
            return err
        } else {
            for _, msg := range messages { t.Error(msg) }
        }

        // Return
        return nil

    }

    // Run test app
    if err := app.Run(args); err != nil { t.Fatal(err) }

}

