package app

import (
    "bytes"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/services/validators"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    rp "github.com/rocket-pool/smartnode/tests/utils/rocketpool"
)


// Register a node from app options
func AppRegisterNode(options AppOptions) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { return err }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { return err }

    // Register node
    _, _, err = rp.RegisterNode(client, cm, am)
    return err

}


// Seed a node account from app options
func AppSeedNodeAccount(options AppOptions, ethAmount *big.Int, rplAmount *big.Int) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketPoolToken"}); err != nil { return err }

    // Seed node account
    if ethAmount != nil && ethAmount.Cmp(big.NewInt(0)) > 0 {
        if err := test.SeedAccount(client, nodeAccount.Address, ethAmount); err != nil { return err }
    }
    if rplAmount != nil && rplAmount.Cmp(big.NewInt(0)) > 0 {
        if err := rp.MintRPL(client, cm, nodeAccount.Address, rplAmount); err != nil { return err }
    }

    // Return
    return nil

}


// Seed a node contract from app options
func AppSeedNodeContract(options AppOptions, ethAmount *big.Int, rplAmount *big.Int) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPoolToken"}); err != nil { return err }

    // Get node contract address
    nodeContractAddress := new(common.Address)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return errors.New("Node is not registered with Rocket Pool")
    }

    // Seed node contract
    if ethAmount != nil && ethAmount.Cmp(big.NewInt(0)) > 0 {
        if err := test.SeedAccount(client, *nodeContractAddress, ethAmount); err != nil { return err }
    }
    if rplAmount != nil && rplAmount.Cmp(big.NewInt(0)) > 0 {
        if err := rp.MintRPL(client, cm, *nodeContractAddress, rplAmount); err != nil { return err }
    }

    // Return
    return nil

}


// Get a node's required deposit balances from app options
func AppGetNodeRequiredBalances(options AppOptions) (*node.Balances, error) {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return nil, err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return nil, err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return nil, err }
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { return nil, err }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { return nil, err }

    // Get node contract address
    nodeContractAddress := new(common.Address)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return nil, errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return nil, errors.New("Node is not registered with Rocket Pool")
    }

    // Get node contract
    nodeContract, err := cm.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil { return nil, err }

    // Get and return required deposit balances
    return node.GetRequiredBalances(nodeContract)

}


// Create minipools under a node from app options
func AppCreateNodeMinipools(options AppOptions, durationId string, minipoolCount int) ([]common.Address, error) {

    // Create password manager, account manager & key manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)
    km := validators.NewKeyManager(options.KeychainBeacon, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return []common.Address{}, err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return []common.Address{}, err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return []common.Address{}, err }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPool", "rocketPoolToken"}); err != nil { return []common.Address{}, err }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { return []common.Address{}, err }

    // Get node contract address
    nodeContractAddress := new(common.Address)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return []common.Address{}, errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return []common.Address{}, errors.New("Node is not registered with Rocket Pool")
    }

    // Get node contract
    nodeContract, err := cm.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil { return []common.Address{}, err }

    // Create minipools
    minipoolAddresses := []common.Address{}
    for mi := 0; mi < minipoolCount; mi++ {
        if address, err := rp.CreateNodeMinipool(client, cm, am, km, nodeContract, *nodeContractAddress, durationId); err != nil {
            return []common.Address{}, err
        } else {
            minipoolAddresses = append(minipoolAddresses, address)
        }
    }

    // Return
    return minipoolAddresses, nil

}


// Make a node trusted from app options
func AppSetNodeTrusted(options AppOptions) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketAdmin"}); err != nil { return err }

    // Get owner account
    ownerPrivateKey, _, err := test.OwnerAccount()
    if err != nil { return err }

    // Set node trusted status
    txor := bind.NewKeyedTransactor(ownerPrivateKey)
    if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketAdmin"], cm.Abis["rocketAdmin"], "setNodeTrusted", nodeAccount.Address, true); err != nil { return err }

    // Return
    return nil

}

