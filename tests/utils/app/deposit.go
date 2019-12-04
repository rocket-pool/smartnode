package app

import (
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Make a deposit to a queue
func AppDeposit(options AppOptions, durationId string, depositAmount *big.Int, depositorAddress common.Address) error {

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load ABIs
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadABIs([]string{"rocketGroupAccessorContract"}); err != nil { return err }

    // Get owner account
    ownerPrivateKey, _, err := test.OwnerAccount()
    if err != nil { return err }

    // Deposit
    txor := bind.NewKeyedTransactor(ownerPrivateKey)
    txor.Value = depositAmount
    txor.GasLimit = 8000000
    if _, err := eth.ExecuteContractTransaction(client, txor, &depositorAddress, cm.Abis["rocketGroupAccessorContract"], "deposit", durationId); err != nil { return err }

    // Return
    return nil

}

