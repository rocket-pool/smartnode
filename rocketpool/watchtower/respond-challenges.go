package watchtower

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Respond to challenges task
type respondChallenges struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   wallet.Wallet
	rp  *rocketpool.RocketPool
	m   *state.NetworkStateManager
}

// Create respond to challenges task
func newRespondChallenges(c *cli.Context, logger log.ColorLogger, m *state.NetworkStateManager) (*respondChallenges, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetHdWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Return task
	return &respondChallenges{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		rp:  rp,
		m:   m,
	}, nil

}

// Respond to challenges
func (t *respondChallenges) run() error {

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Log
	t.log.Println("Checking for challenges to respond to...")

	// Check for active challenges
	isChallenged, err := trustednode.GetMemberIsChallenged(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}
	if !isChallenged {
		return nil
	}

	// Log
	t.log.Printlnf("Node %s has an active challenge against it, responding...", nodeAccount.Address.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := trustednode.EstimateDecideChallengeGas(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to respond to the challenge: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = gasInfo.SafeGasLimit

	// Respond to challenge
	hash, err := trustednode.DecideChallenge(t.rp, nodeAccount.Address, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log & return
	t.log.Printlnf("Successfully responded to challenge against node %s.", nodeAccount.Address.Hex())
	return nil

}
