package auction

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)

// Create an amount of slashed RPL in the auction contract
func CreateSlashedRPL(rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount, trustedNodeAccount2 *accounts.Account, userAccount *accounts.Account) error {

    // Stake a large amount of RPL against the node
    if err := nodeutils.StakeRPL(rp, ownerAccount, trustedNodeAccount, eth.EthToWei(1000000)); err != nil { return err }

    // Create unbonded minipool
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, trustedNodeAccount, big.NewInt(0))
    if err != nil { return err }

    // Deposit user ETH to minipool
    opts := userAccount.GetTransactor()
    opts.Value = eth.EthToWei(32)
    if _, err := deposit.Deposit(rp, opts); err != nil { return err }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, trustedNodeAccount); err != nil { return err }

    // Mark minipool as withdrawable with zero end balance
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil { return err }
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount2.GetTransactor()); err != nil { return err }

    // Distribute balance and destroy pool to send slashed RPL to auction contract
    if _, err := mp.DistributeBalanceAndDestroy(trustedNodeAccount.GetTransactor()); err != nil { return err }

    // Return
    return nil

}

