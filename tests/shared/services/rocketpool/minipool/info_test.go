package minipool

import (
    "bytes"
    "io/ioutil"
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
    rp "github.com/rocket-pool/smartnode/tests/utils/rocketpool"
)


// Test minipool details getter
func TestGetDetails(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool details without rocketMinipool ABI; load ABI
    if _, err := minipool.GetDetails(cm, &minipoolAddress); err == nil { t.Error("GetDetails() method should return error without rocketMinipool ABI loaded") }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil { t.Fatal(err) }

    // Get minipool details
    details, err := minipool.GetDetails(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check minipool details
    expectedDepositCount := big.NewInt(0)
    if !bytes.Equal(minipoolAddress.Bytes(), details.Address.Bytes()) { t.Error("Minipool address does not match created address") }
    if details.Status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, details.Status) }
    if details.StakingDurationId != "3m" { t.Errorf("Incorrect minipool staking duration ID: expected %s, got %s", "3m", details.StakingDurationId) }
    if details.UserDepositCount.Cmp(expectedDepositCount) != 0 { t.Errorf("Incorrect minipool deposit count: expected %s, got %s", expectedDepositCount.String(), details.UserDepositCount.String()) }

    // Get details for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetDetails(cm, &address); err == nil { t.Error("GetDetails() method should return error for nonexistent minipools") }

}


// Test minipool status getter
func TestGetStatus(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool status without rocketMinipool ABI; load ABI
    if _, err := minipool.GetStatus(cm, &minipoolAddress); err == nil { t.Error("GetStatus() method should return error without rocketMinipool ABI loaded") }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil { t.Fatal(err) }

    // Get minipool status
    status, err := minipool.GetStatus(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check minipool status
    expectedStakingDuration := big.NewInt(20250)
    if status.Status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, status.Status) }
    if status.StakingDuration.Cmp(expectedStakingDuration) != 0 { t.Errorf("Incorrect minipool staking duration: expected %s, got %s", expectedStakingDuration.String(), status.StakingDuration.String()) }

    // Get status for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetStatus(cm, &address); err == nil { t.Error("GetStatus() method should return error for nonexistent minipools") }

}


// Test minipool node status getter
func TestGetNodeStatus(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool node status without rocketMinipool ABI; load ABI
    if _, err := minipool.GetNodeStatus(cm, &minipoolAddress); err == nil { t.Error("GetNodeStatus() method should return error without rocketMinipool ABI loaded") }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil { t.Fatal(err) }

    // Get minipool node status
    status, err := minipool.GetNodeStatus(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check node status
    if status.Status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, status.Status) }
    if !status.DepositExists { t.Errorf("Incorrect minipool node deposit exists status: expected %t, got %t", true, status.DepositExists) }

    // Get node status for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetNodeStatus(cm, &address); err == nil { t.Error("GetNodeStatus() method should return error for nonexistent minipools") }

}


// Test minipool status code getter
func TestGetStatusCode(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool status code without rocketMinipool ABI; load ABI
    if _, err := minipool.GetStatusCode(cm, &minipoolAddress); err == nil { t.Error("GetStatusCode() method should return error without rocketMinipool ABI loaded") }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil { t.Fatal(err) }

    // Get minipool status code
    status, err := minipool.GetStatusCode(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check status code
    if status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, status) }

    // Get status code for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetStatusCode(cm, &address); err == nil { t.Error("GetStatusCode() method should return error for nonexistent minipools") }

}


// Test active minipools by validator pubkey getter
func TestGetActiveMinipoolsByValidatorPubkey(t *testing.T) {

    // Create test app options
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }    
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise & register node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := testapp.AppRegisterNode(appOptions); err != nil { t.Fatal(err) }

    // Create group accessor
    _, accessorAddress, err := testapp.AppCreateGroupAccessor(appOptions)
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Create minipools
    minipoolAddresses1, err := testapp.AppCreateNodeMinipools(appOptions, "3m", 1)
    if err != nil { t.Fatal(err) }
    minipoolAddresses2, err := testapp.AppCreateNodeMinipools(appOptions, "6m", 1)
    if err != nil { t.Fatal(err) }
    minipoolAddresses3, err := testapp.AppCreateNodeMinipools(appOptions, "12m", 1)
    if err != nil { t.Fatal(err) }

    // Stake minipools
    if err := testapp.AppStakeAllMinipools(appOptions, "3m", accessorAddress, minipoolAddresses1); err != nil { t.Fatal(err) }
    if err := testapp.AppStakeAllMinipools(appOptions, "6m", accessorAddress, minipoolAddresses2); err != nil { t.Fatal(err) }
    if err := testapp.AppStakeAllMinipools(appOptions, "12m", accessorAddress, minipoolAddresses3); err != nil { t.Fatal(err) }

    // Get active minipools without rocketMinipool ABI; load ABI
    if _, err := minipool.GetActiveMinipoolsByValidatorPubkey(cm); err == nil { t.Error("GetActiveMinipoolsByValidatorPubkey() method should return error without rocketMinipool ABI loaded") }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil { t.Fatal(err) }

    // Get active minipools
    minipools, err := minipool.GetActiveMinipoolsByValidatorPubkey(cm)
    if err != nil { t.Fatal(err) }

    // Search for created minipools in map
    minipool1Found := false
    minipool2Found := false
    minipool3Found := false
    for _, address := range *minipools {
        if bytes.Equal(address.Bytes(), minipoolAddresses1[0].Bytes()) { minipool1Found = true }
        if bytes.Equal(address.Bytes(), minipoolAddresses2[0].Bytes()) { minipool2Found = true }
        if bytes.Equal(address.Bytes(), minipoolAddresses3[0].Bytes()) { minipool3Found = true }
    }
    if !(minipool1Found && minipool2Found && minipool3Found) { t.Error("Created minipools not found in active set") }

}

