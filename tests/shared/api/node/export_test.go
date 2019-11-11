package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node export methods
func TestNodeExport(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Export node account
    if _, err := node.ExportNodeAccount(p); err != nil {
        t.Error(err)
    }

}

