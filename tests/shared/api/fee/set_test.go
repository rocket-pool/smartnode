package fee

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/fee"
    "github.com/rocket-pool/smartnode/shared/services"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test fee set methods
func TestFeeSet(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context
    c := testapp.GetAppContext(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Set target user fee
    if feeSet, err := fee.SetTargetUserFee(p, 10); err != nil {
        t.Error(err)
    } else if !feeSet.Success {
        t.Error("Target user fee was not set successfully")
    }

}

