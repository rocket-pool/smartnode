package minipool

import (
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/validator"
)


// Minipool created event
type minipoolCreated struct {
    Minipool common.Address
    Node common.Address
    Time *big.Int
}


// Create a minipool
func CreateMinipool(rp *rocketpool.RocketPool, nodeAccount *accounts.Account, depositAmount *big.Int) (*minipool.Minipool, error) {

    // Make node deposit
    opts := nodeAccount.GetTransactor()
    opts.Value = depositAmount
    txReceipt, err := node.Deposit(rp, 0, opts)
    if err != nil { return nil, err }

    // Get minipool manager contract
    rocketMinipoolManager, err := rp.GetContract("rocketMinipoolManager")
    if err != nil { return nil, err }

    // Get created minipool address
    minipoolCreatedEvents, err := rocketMinipoolManager.GetTransactionEvents(txReceipt, "MinipoolCreated", minipoolCreated{})
    if err != nil || len(minipoolCreatedEvents) == 0 {
        return nil, errors.New("Could not get minipool created event")
    }
    minipoolAddress := minipoolCreatedEvents[0].(minipoolCreated).Minipool

    // Return minipool instance
    return minipool.NewMinipool(rp, minipoolAddress)

}


// Stake a minipool
func StakeMinipool(rp *rocketpool.RocketPool, mp *minipool.Minipool, nodeAccount *accounts.Account) error {

    // Get validator & deposit data
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { return err }
    validatorSignature, err := validator.GetValidatorSignature()
    if err != nil { return err }
    depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, validator.GetWithdrawalCredentials(), validatorSignature)
    if err != nil { return err }

    // Stake minipool & return
    _, err = mp.Stake(validatorPubkey, validatorSignature, depositDataRoot, nodeAccount.GetTransactor())
    return err

}

