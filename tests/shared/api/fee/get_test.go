package fee

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/fee"
    "github.com/rocket-pool/smartnode/shared/services"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test fee get methods
func TestFeeGet(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context
    c := testapp.GetAppContext(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
        CM: true,
        LoadContracts: []string{"rocketNodeSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get user fee
    if userFee, err := fee.GetUserFee(p); err != nil {
        t.Error(err)
    } else if userFee.TargetUserFeePerc != -1 {
        t.Error("TargetUserFeePerc was not unset without a target user fee set")
    }

    // Set user fee
    if _, err := fee.SetTargetUserFee(p, 10); err != nil { t.Fatal(err) }

    // Get updated user fee
    if userFee, err := fee.GetUserFee(p); err != nil {
        t.Error(err)
    } else if userFee.TargetUserFeePerc != 10 {
        t.Error("TargetUserFeePerc was not set to correct target user fee")
    }

}

