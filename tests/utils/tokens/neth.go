package tokens

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/utils/minipool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Mint an amount of nETH to an account
func MintNETH(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, nodeAccount *accounts.Account, toAccount *accounts.Account, amount *big.Int) error {

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", toAccount.GetTransactor()); err != nil { return err }
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { return err }
    if _, err := node.SetNodeTrusted(rp, nodeAccount.Address, true, ownerAccount.GetTransactor()); err != nil { return err }

    // Create & stake minipool
    mp, err := minipoolutils.CreateMinipool(rp, toAccount, eth.EthToWei(32))
    if err != nil { return err }
    if err := minipoolutils.StakeMinipool(rp, mp, toAccount); err != nil { return err }

    // Disable minipool withdrawal delay
    withdrawalDelay, err := settings.GetMinipoolWithdrawalDelay(rp, nil)
    if err != nil { return err }
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, 0, ownerAccount.GetTransactor()); err != nil { return err }

    // Mark minipool as withdrawable and withdraw
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, eth.EthToWei(32), amount, nodeAccount.GetTransactor()); err != nil { return err }
    if _, err := mp.Withdraw(toAccount.GetTransactor()); err != nil { return err }

    // Re-enable minipool withdrawal delay
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, withdrawalDelay, ownerAccount.GetTransactor()); err != nil { return err }

    // Return
    return nil

}

