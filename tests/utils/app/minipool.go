package app

import (
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// RocketDepositQueue DepositChunkFragmentAssign event
type DepositChunkFragmentAssign struct {
    MinipoolAddress common.Address
    DepositID [32]byte
    UserID common.Address
    GroupID common.Address
    Value *big.Int
    Created *big.Int
}


// Progress all minipools to staking from app options
func AppStakeAllMinipools(options AppOptions, durationId string, depositorAddress common.Address) error {

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketDepositQueue", "rocketDepositSettings"}); err != nil { return err }
    if err := cm.LoadABIs([]string{"rocketGroupAccessorContract"}); err != nil { return err }

    // Get deposit size
    var depositSize big.Int
    chunkSize := new(*big.Int)
    if err := cm.Contracts["rocketDepositSettings"].Call(nil, chunkSize, "getDepositChunkSize"); err != nil { return err }
    chunkAssignMax := new(*big.Int)
    if err := cm.Contracts["rocketDepositSettings"].Call(nil, chunkAssignMax, "getChunkAssignMax"); err != nil { return err }
    depositSize.Mul(*chunkSize, *chunkAssignMax)

    // Get owner account
    ownerPrivateKey, _, err := test.OwnerAccount()
    if err != nil { return err }

    // Deposit until no assignments are made
    for {

        // Deposit
        txor := bind.NewKeyedTransactor(ownerPrivateKey)
        txor.Value = &depositSize
        txor.GasLimit = 8000000
        txReceipt, err := eth.ExecuteContractTransaction(client, txor, &depositorAddress, cm.Abis["rocketGroupAccessorContract"], "deposit", durationId)
        if err != nil { return err }

        // Stop if no assignments made
        if chunkAssignEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketDepositQueue"], cm.Abis["rocketDepositQueue"], "DepositChunkFragmentAssign", DepositChunkFragmentAssign{}); err != nil {
            return err
        } else if len(chunkAssignEvents) == 0 {
            return nil
        }

    }

}


// Withdraw minipools from app options
// Requires app node to be trusted
func AppWithdrawMinipools(options AppOptions, minipoolAddresses []common.Address, balance *big.Int) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketNodeWatchtower"}); err != nil { return err }

    // Logout & withdraw minipools
    for _, address := range minipoolAddresses {
        txor, err := am.GetNodeAccountTransactor()
        if err != nil { return err }
        if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketNodeWatchtower"], cm.Abis["rocketNodeWatchtower"], "logoutMinipool", address); err != nil { return err }
        txor, err = am.GetNodeAccountTransactor()
        if err != nil { return err }
        if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketNodeWatchtower"], cm.Abis["rocketNodeWatchtower"], "withdrawMinipool", address, balance); err != nil { return err }
    }

    // Return
    return nil

}

