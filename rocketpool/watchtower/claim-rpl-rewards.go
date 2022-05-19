package watchtower

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/legacy/v1.0.0/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Claim RPL rewards task
type claimRplRewards struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
}

// Create claim RPL rewards task
func newClaimRplRewards(c *cli.Context, logger log.ColorLogger) (*claimRplRewards, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &claimRplRewards{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		rp:  rp,
	}, nil

}

// Claim RPL rewards
func (t *claimRplRewards) run() (bool, error) {

	legacyClaimTrustedNodeAddress := t.cfg.Smartnode.GetLegacyClaimTrustedNodeAddress()

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return false, err
	}

	// Check if the contract upgrade has happened yet
	isMergeUpdateDeployed, err := rp.IsMergeUpdateDeployed(t.rp)
	if err != nil {
		return false, fmt.Errorf("error checking if merge update has been deployed: %w", err)
	}
	if isMergeUpdateDeployed {
		t.log.Println("The merge update contracts have been deployed! Auto-claiming is no longer necessary. Enjoy the new rewards system!")
		return true, nil
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return false, err
	}

	// Check node trusted status
	nodeTrusted, err := trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return false, err
	}
	if !nodeTrusted {
		return false, nil
	}

	// Log
	t.log.Println("Checking for RPL rewards to claim...")

	// Check for rewards
	rewardsAmountWei, err := rewards.GetTrustedNodeClaimRewardsAmount(t.rp, nodeAccount.Address, nil, &legacyClaimTrustedNodeAddress)
	if err != nil {
		return false, err
	}
	if rewardsAmountWei.Cmp(big.NewInt(0)) == 0 {
		return false, nil
	}

	// Log
	t.log.Printlnf("%.6f RPL is available to claim...", math.RoundDown(eth.WeiToEth(rewardsAmountWei), 6))

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	gasInfo, err := rewards.EstimateClaimTrustedNodeRewardsGas(t.rp, opts, &legacyClaimTrustedNodeAddress)
	if err != nil {
		return false, fmt.Errorf("Could not estimate the gas required to claim RPL: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(WatchtowerMaxFee)
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return false, nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(WatchtowerMaxPriorityFee)
	opts.GasLimit = gasInfo.SafeGasLimit

	// Claim rewards
	hash, err := rewards.ClaimTrustedNodeRewards(t.rp, opts, &legacyClaimTrustedNodeAddress)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be mined
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return false, err
	}

	// Log & return
	t.log.Printlnf("Successfully claimed %.6f RPL in rewards.", math.RoundDown(eth.WeiToEth(rewardsAmountWei), 6))
	return false, nil

}
