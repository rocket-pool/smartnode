package minipool

import (
    "bytes"
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"

    test "github.com/rocket-pool/smartnode/tests/utils"
    rp "github.com/rocket-pool/smartnode/tests/utils/rocketpool"
)


// Test minipool details getter
func TestGetDetails(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Create key manager
    km, err := test.NewInitKeyManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketMinipool", "rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, km, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool details
    details, err := minipool.GetDetails(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check minipool details
    expectedDepositCount := big.NewInt(0)
    if !bytes.Equal(minipoolAddress.Bytes(), details.Address.Bytes()) { t.Error("Minipool address does not match created address") }
    if details.Status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, details.Status) }
    if details.StakingDurationId != "3m" { t.Errorf("Incorrect minipool staking duration ID: expected %s, got %s", "3m", details.StakingDurationId) }
    if details.DepositCount.Cmp(expectedDepositCount) != 0 { t.Errorf("Incorrect minipool deposit count: expected %s, got %s", expectedDepositCount.String(), details.DepositCount.String()) }

    // Get details for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetDetails(cm, &address); err == nil { t.Error("GetDetails() method should return error for nonexistent minipools") }

}


// Test minipool status getter
func TestGetStatus(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Create key manager
    km, err := test.NewInitKeyManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketMinipool", "rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, km, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool status
    status, err := minipool.GetStatus(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check minipool status
    expectedStakingDuration := big.NewInt(526000)
    if status.Status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, status.Status) }
    if status.StakingDuration.Cmp(expectedStakingDuration) != 0 { t.Errorf("Incorrect minipool staking duration: expected %s, got %s", expectedStakingDuration.String(), status.StakingDuration.String()) }
    if _, err := km.GetValidatorKey(status.ValidatorPubkey); err != nil { t.Error("Minipool validator pubkey does not match local validator key") }

    // Get status for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetStatus(cm, &address); err == nil { t.Error("GetStatus() method should return error for nonexistent minipools") }

}


// Test minipool node status getter
func TestGetNodeStatus(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Create key manager
    km, err := test.NewInitKeyManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketMinipool", "rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, km, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

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

    // Create key manager
    km, err := test.NewInitKeyManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketMinipool", "rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Create minipool
    minipoolAddress, err := rp.CreateNodeMinipool(client, cm, am, km, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }

    // Get minipool status code
    status, err := minipool.GetStatusCode(cm, &minipoolAddress)
    if err != nil { t.Fatal(err) }

    // Check status code
    if status != minipool.INITIALIZED { t.Errorf("Incorrect minipool status: expected %d, got %d", minipool.INITIALIZED, status) }

    // Get status code for nonexistent minipool
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")
    if _, err := minipool.GetStatusCode(cm, &address); err == nil { t.Error("GetStatusCode() method should return error for nonexistent minipools") }

}

