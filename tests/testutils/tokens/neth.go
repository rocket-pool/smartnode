package tokens

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
)


// Mint an amount of nETH to an account
func MintNETH(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount *accounts.Account, toAccount *accounts.Account, amount *big.Int) error {

    // Register node if not registered
    if nodeExists, err := node.GetNodeExists(rp, toAccount.Address, nil); err != nil {
        return err
    } else if !nodeExists {
        if _, err := node.RegisterNode(rp, "Australia/Brisbane", toAccount.GetTransactor()); err != nil { return err }
    }

    // Create & stake minipool
    mp, err := minipoolutils.CreateMinipool(rp, toAccount, eth.EthToWei(32))
    if err != nil { return err }
    if err := minipoolutils.StakeMinipool(rp, mp, toAccount); err != nil { return err }

    // Disable minipool withdrawal delay
    withdrawalDelay, err := protocol.GetMinipoolWithdrawalDelay(rp, nil)
    if err != nil { return err }
    if _, err := protocol.SetMinipoolWithdrawalDelay(rp, 0, ownerAccount.GetTransactor()); err != nil { return err }

    // Mark minipool as withdrawable and withdraw
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, eth.EthToWei(32), amount, trustedNodeAccount.GetTransactor()); err != nil { return err }
    if _, err := mp.Withdraw(toAccount.GetTransactor()); err != nil { return err }

    // Re-enable minipool withdrawal delay
    if _, err := protocol.SetMinipoolWithdrawalDelay(rp, withdrawalDelay, ownerAccount.GetTransactor()); err != nil { return err }

    // Return
    return nil

}

