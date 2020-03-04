package node

import (
    "bytes"
    "testing"

    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    rp "github.com/rocket-pool/smartnode/tests/utils/rocketpool"
)


// Test node account balances getter
func TestGetAccountBalances(t *testing.T) {

    // Create account manager & get account
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }
    account, err := am.GetNodeAccount()
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketPoolToken"}); err != nil { t.Fatal(err) }

    // Amounts to seed to account
    ethAmount := eth.EthToWei(3)
    rethAmount := eth.EthToWei(0)
    rplAmount := eth.EthToWei(2)

    // Seed account
    if err := test.SeedAccount(client, account.Address, ethAmount); err != nil { t.Fatal(err) }
    if err := rp.MintRPL(client, cm, account.Address, rplAmount); err != nil { t.Fatal(err) }

    // Get account balances without required contracts; load contracts
    if _, err := node.GetAccountBalances(account.Address, client, cm); err == nil { t.Error("GetAccountBalances() method should return error without contracts loaded") }
    if err := cm.LoadContracts([]string{"rocketETHToken"}); err != nil { t.Fatal(err) }

    // Get account balances
    balances, err := node.GetAccountBalances(account.Address, client, cm)
    if err != nil { t.Fatal(err) }

    // Check account balances
    if balances.EtherWei.String() != ethAmount.String() { t.Errorf("Incorrect balance ETH value: expected %s, got %s", ethAmount.String(), balances.EtherWei.String()) }
    if balances.RethWei.String() != rethAmount.String() { t.Errorf("Incorrect balance rETH value: expected %s, got %s", rethAmount.String(), balances.RethWei.String()) }
    if balances.RplWei.String() != rplAmount.String() { t.Errorf("Incorrect balance RPL value: expected %s, got %s", rplAmount.String(), balances.RplWei.String()) }

}


// Test node contract balances getter
func TestGetBalances(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Amounts to seed to contract
    ethAmount := eth.EthToWei(5)
    rplAmount := eth.EthToWei(4)

    // Seed contract
    if err := test.SeedAccount(client, nodeContractAddress, ethAmount); err != nil { t.Fatal(err) }
    if err := rp.MintRPL(client, cm, nodeContractAddress, rplAmount); err != nil { t.Fatal(err) }

    // Get contract balances
    balances, err := node.GetBalances(nodeContract)
    if err != nil { t.Fatal(err) }

    // Check contract balances
    if balances.EtherWei.String() != ethAmount.String() { t.Errorf("Incorrect balance ETH value: expected %s, got %s", ethAmount.String(), balances.EtherWei.String()) }
    if balances.RplWei.String() != rplAmount.String() { t.Errorf("Incorrect balance RPL value: expected %s, got %s", rplAmount.String(), balances.RplWei.String()) }

}


// Test node deposit reservation required balances getter
func TestGetRequiredBalances(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Reserve node deposit
    if err := rp.ReserveNodeDeposit(client, cm, am, nodeContractAddress, "3m"); err != nil { t.Fatal(err) }

    // Get required balances
    balances, err := node.GetRequiredBalances(nodeContract)
    if err != nil { t.Fatal(err) }

    // Check required balances
    expectedEth := eth.EthToWei(16)
    if balances.EtherWei.String() != expectedEth.String() { t.Errorf("Incorrect required ETH value: expected %s, got %s", expectedEth.String(), balances.EtherWei.String()) }

}


// Test node deposit reservation details getter
func TestGetReservationDetails(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Get reservation details without required contracts; load contracts
    if _, err := node.GetReservationDetails(nodeContract, cm); err == nil { t.Error("GetReservationDetails() method should return error without contracts loaded") }
    if err := cm.LoadContracts([]string{"rocketNodeSettings"}); err != nil { t.Fatal(err) }

    // Get reservation details before deposit reservation
    details, err := node.GetReservationDetails(nodeContract, cm)
    if err != nil { t.Fatal(err) }

    // Check reservation details
    if details.Exists { t.Errorf("Incorrect deposit exists value: expected %t, got %t", false, details.Exists) }
    if details.StakingDurationID != "" { t.Error("Staking duration ID should be undefined") }
    if details.EtherRequiredWei != nil { t.Error("Required ETH value should be undefined") }
    if details.RplRequiredWei != nil { t.Error("Required ETH value should be undefined") }

    // Reserve node deposit
    if err := rp.ReserveNodeDeposit(client, cm, am, nodeContractAddress, "3m"); err != nil { t.Fatal(err) }

    // Get reservation details
    details, err = node.GetReservationDetails(nodeContract, cm)
    if err != nil { t.Fatal(err) }

    // Check reservation details
    expectedEth := eth.EthToWei(16)
    if !details.Exists { t.Errorf("Incorrect deposit exists value: expected %t, got %t", true, details.Exists) }
    if details.StakingDurationID != "3m" { t.Errorf("Incorrect staking duration ID: expected %s, got %s", "3m", details.StakingDurationID) }
    if details.EtherRequiredWei.String() != expectedEth.String() { t.Errorf("Incorrect required ETH value: expected %s, got %s", expectedEth.String(), details.EtherRequiredWei.String()) }

}


// Test minipool addresses getter
func TestGetMinipoolAddresses(t *testing.T) {

    // Create account manager & get account
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }
    account, err := am.GetNodeAccount()
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Get minipool addresses without required contracts; load contracts
    if _, err := node.GetMinipoolAddresses(account.Address, cm); err == nil { t.Error("GetMinipoolAddresses() method should return error without contracts loaded") }
    if err := cm.LoadContracts([]string{"utilAddressSetStorage"}); err != nil { t.Fatal(err) }

    // Get minipool addresses for nonexistent node
    minipoolAddresses, err := node.GetMinipoolAddresses(account.Address, cm)
    if err != nil { t.Fatal(err) }

    // Check minipool addresses
    if len(minipoolAddresses) > 0 { t.Error("Minipool address list should be empty for new node") }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Get minipool addresses before minipool creation
    minipoolAddresses, err = node.GetMinipoolAddresses(account.Address, cm)
    if err != nil { t.Fatal(err) }

    // Check minipool addresses
    if len(minipoolAddresses) > 0 { t.Error("Minipool address list should be empty for new node") }

    // Create minipools
    minipool1Address, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "3m")
    if err != nil { t.Fatal(err) }
    minipool2Address, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "6m")
    if err != nil { t.Fatal(err) }
    minipool3Address, err := rp.CreateNodeMinipool(client, cm, am, nodeContract, nodeContractAddress, "12m")
    if err != nil { t.Fatal(err) }

    // Get minipool addresses
    minipoolAddresses, err = node.GetMinipoolAddresses(account.Address, cm)
    if err != nil { t.Fatal(err) }

    // Check minipool addresses
    if !bytes.Equal(minipoolAddresses[0].Bytes(), minipool1Address.Bytes()) { t.Error("Minipool address 1 does not match created address") }
    if !bytes.Equal(minipoolAddresses[1].Bytes(), minipool2Address.Bytes()) { t.Error("Minipool address 2 does not match created address") }
    if !bytes.Equal(minipoolAddresses[2].Bytes(), minipool3Address.Bytes()) { t.Error("Minipool address 3 does not match created address") }

}

