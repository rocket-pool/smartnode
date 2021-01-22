package tokens

import (
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/utils/validator"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Minipool created event
type minipoolCreated struct {
    Minipool common.Address
    Node common.Address
    Time *big.Int
}


// Mint an amount of nETH to an account
func MintNETH(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, nodeAccount *accounts.Account, toAccount *accounts.Account, amount *big.Int) error {

    // Register trusted node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", toAccount.GetTransactor()); err != nil { return err }
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { return err }
    if _, err := node.SetNodeTrusted(rp, nodeAccount.Address, true, ownerAccount.GetTransactor()); err != nil { return err }

    // Make node deposit
    opts := toAccount.GetTransactor()
    opts.Value = eth.EthToWei(32)
    txReceipt, err := node.Deposit(rp, 0, opts)
    if err != nil { return err }

    // Get created minipool address
    minipoolManagerAddress, minipoolManagerABI, err := contract.GetDetails(rp, "rocketMinipoolManager")
    if err != nil { return err }
    minipoolCreatedEvents, err := contract.GetTransactionEvents(rp.Client, minipoolManagerAddress, minipoolManagerABI, txReceipt, "MinipoolCreated", minipoolCreated{})
    if err != nil || len(minipoolCreatedEvents) == 0 {
        return errors.New("Could not get minipool created event")
    }
    minipoolAddress := minipoolCreatedEvents[0].(minipoolCreated).Minipool

    // Stake minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil { return err }
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { return err }
    validatorSignature, err := validator.GetValidatorSignature()
    if err != nil { return err }
    depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, validator.GetWithdrawalCredentials(), validatorSignature)
    if err != nil { return err }
    if _, err := mp.Stake(validatorPubkey, validatorSignature, depositDataRoot, toAccount.GetTransactor()); err != nil { return err }

    // Disable minipool withdrawal delay
    withdrawalDelay, err := settings.GetMinipoolWithdrawalDelay(rp, nil)
    if err != nil { return err }
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, 0, ownerAccount.GetTransactor()); err != nil { return err }

    // Mark minipool as withdrawable and withdraw
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, minipoolAddress, eth.EthToWei(32), amount, nodeAccount.GetTransactor()); err != nil { return err }
    if _, err := mp.Withdraw(toAccount.GetTransactor()); err != nil { return err }

    // Re-enable minipool withdrawal delay
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, withdrawalDelay, ownerAccount.GetTransactor()); err != nil { return err }

    // Return
    return nil

}

