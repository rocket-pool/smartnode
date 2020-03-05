package node

import (
    "io/ioutil"
    "testing"
    "time"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/rocketpool-node/node"
    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node checkin functionality
func TestNodeCheckin(t *testing.T) {

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

        // Initialise services
        p, err := services.NewProvider(c, services.ProviderOpts{
            DB: true,
            AM: true,
            CM: true,
            NodeContract: true,
            LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
            LoadAbis: []string{"rocketNodeContract"},
        })
        if err != nil { return err }

        // Start node checkin process
        go node.StartCheckinProcess(p)

        // Give node time to checkin
        time.Sleep(10 * time.Second)

        // Check output
        prefix := "(?i)^\\d{4}/\\d{2}/\\d{2} \\d{2}:\\d{2}:\\d{2} "
        if messages, err := testapp.CheckOutput(output.Name(), []string{prefix + "Checking in"}, map[int][]string{
            1: []string{prefix + "Checked in successfully with an average load of \\d+\\.\\d+% and a node fee vote of '.+'$", "Checkin message incorrect"},
            2: []string{prefix + "Time until next checkin:", "Next checkin message incorrect"},
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

