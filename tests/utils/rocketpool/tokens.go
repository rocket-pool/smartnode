package rocketpool

import (
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Mint RPL to an address
func MintRPL(client *ethclient.Client, cm *rocketpool.ContractManager, address common.Address, amount *big.Int) error {

    // Get owner account
    ownerPrivateKey, _, err := test.OwnerAccount()
    if err != nil { return err }

    // Mint RPL
    txor := bind.NewKeyedTransactor(ownerPrivateKey)
    if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketPoolToken"], cm.Abis["rocketPoolToken"], "mint", address, amount); err != nil { return err }

    // Return
    return nil

}

