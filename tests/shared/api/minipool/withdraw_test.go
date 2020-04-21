package minipool

import (
    "io/ioutil"
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test minipool withdraw methods
func TestMinipoolWithdraw(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise & register node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := testapp.AppRegisterNode(appOptions); err != nil { t.Fatal(err) }

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings", "utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool", "rocketMinipoolDelegateNode", "rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Create group accessor
    _, accessorAddress, err := testapp.AppCreateGroupAccessor(appOptions)
    if err != nil { t.Fatal(err) }

    // Create minipools
    minipoolAddresses, err := testapp.AppCreateNodeMinipools(appOptions, "12m", 3)
    if err != nil { t.Fatal(err) }

    // Stake minipools
    if err := testapp.AppStakeAllMinipools(appOptions, "12m", accessorAddress, minipoolAddresses); err != nil { t.Fatal(err) }

    // Check withdrawable minipools with no withdrawable minipools
    if withdrawable, err := minipool.GetWithdrawableMinipools(p); err != nil {
        t.Error(err)
    } else if len(withdrawable) != 0 {
        t.Error("Withdrawable minipools returned with no withdrawable minipools")
    }

    // Check nonexistent minipool cannot be withdrawn
    if canWithdraw, err := minipool.CanWithdrawMinipool(p, common.HexToAddress("0x0000000000000000000000000000000000000000")); err != nil {
        t.Error(err)
    } else if canWithdraw.Success || !canWithdraw.MinipoolDidNotExist {
        t.Error("MinipoolDidNotExist flag was not set for nonexistent minipool")
    }

    // Attempt to withdraw minipool
    if _, err := minipool.WithdrawMinipool(p, common.HexToAddress("0x0000000000000000000000000000000000000000")); err == nil {
        t.Error("WithdrawMinipool() method did not return error for nonexistent minipool")
    }

    // Check minipool with invalid status cannot be withdrawn
    if canWithdraw, err := minipool.CanWithdrawMinipool(p, minipoolAddresses[0]); err != nil {
        t.Error(err)
    } else if canWithdraw.Success || !canWithdraw.InvalidStatus {
        t.Error("InvalidStatus flag was not set for minipool with invalid status")
    }

    // Attempt to withdraw minipool
    if _, err := minipool.WithdrawMinipool(p, minipoolAddresses[0]); err == nil {
        t.Error("WithdrawMinipool() method did not return error for minipool with invalid status")
    }

    // Logout and withdraw minipools
    if err := testapp.AppWithdrawMinipools(appOptions, minipoolAddresses, eth.EthToWei(40)); err != nil { t.Fatal(err) }

    // Check withdrawable minipools
    if withdrawable, err := minipool.GetWithdrawableMinipools(p); err != nil {
        t.Error(err)
    } else if len(withdrawable) != 3 {
        t.Error("Withdrawable minipools not returned")
    }

    // Check minipool can be withdrawn
    if canWithdraw, err := minipool.CanWithdrawMinipool(p, minipoolAddresses[0]); err != nil {
        t.Error(err)
    } else if !canWithdraw.Success {
        t.Error("Minipool cannot be withdrawn")
    }

    // Withdraw from minipool
    if withdrawn, err := minipool.WithdrawMinipool(p, minipoolAddresses[0]); err != nil {
        t.Error(err)
    } else if !withdrawn.Success {
        t.Error("Minipool was not withdrawn from successfully")
    }

    // Check minipool without existing node deposit cannot be withdrawn
    if canWithdraw, err := minipool.CanWithdrawMinipool(p, minipoolAddresses[0]); err != nil {
        t.Error(err)
    } else if canWithdraw.Success || !canWithdraw.NodeDepositDidNotExist {
        t.Error("NodeDepositDidNotExist flag was not set for minipool without existing node deposit")
    }

    // Attempt to withdraw minipool
    if _, err := minipool.WithdrawMinipool(p, minipoolAddresses[0]); err == nil {
        t.Error("WithdrawMinipool() method did not return error for minipool without existing node deposit")
    }

}

