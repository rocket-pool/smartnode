package auction

import (
	"fmt"
	"testing"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)

// Create an amount of slashed RPL in the auction contract
func CreateSlashedRPL(t *testing.T, rp *rocketpool.RocketPool, ownerAccount *accounts.Account, trustedNodeAccount, trustedNodeAccount2 *accounts.Account, userAccount *accounts.Account) error {

	// Stake a large amount of RPL against the node
	if err := nodeutils.StakeRPL(rp, ownerAccount, trustedNodeAccount, eth.EthToWei(1000000)); err != nil {
		return err
	}

	// Make user deposit
	depositOpts := userAccount.GetTransactor()
	depositOpts.Value = eth.EthToWei(16)
	if _, err := deposit.Deposit(rp, depositOpts); err != nil {
		return err
	}

	// Create unbonded minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, trustedNodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		return err
	}

	// Deposit user ETH to minipool
	opts := userAccount.GetTransactor()
	opts.Value = eth.EthToWei(16)
	if _, err := deposit.Deposit(rp, opts); err != nil {
		return err
	}

	// Delay for the time between depositing and staking
	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		return err
	}
	err = evm.IncreaseTime(int(scrubPeriod + 1))
	if err != nil {
		return fmt.Errorf("error increasing time: %w", err)
	}

	// Stake minipool
	if err := minipoolutils.StakeMinipool(rp, mp, trustedNodeAccount); err != nil {
		return err
	}

	// Mark minipool as withdrawable with zero end balance
	if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil {
		return err
	}
	if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount2.GetTransactor()); err != nil {
		return err
	}

	// Distribute balance and finalise pool to send slashed RPL to auction contract
	if _, err := mp.DistributeBalanceAndFinalise(trustedNodeAccount.GetTransactor()); err != nil {
		return err
	}

	// Return
	return nil

}
