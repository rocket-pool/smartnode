package watchtower

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Settings
const MinipoolPrelaunchDetailsBatchSize = 20


// Submit scrub minipools task
type submitScrubMinipools struct {
	c *cli.Context
	log log.ColorLogger
	cfg config.RocketPoolConfig
	w *wallet.Wallet
	rp *rocketpool.RocketPool
	bc beacon.Client
}


// Prelaunch minipool info
type minipoolPrelaunchDetails struct {
	Address common.Address
	WithdrawalCredentials []byte	`ssz-size:"32"`
	validatorPubkey types.ValidatorPubkey
}


// Create submit scrub minipools task
func newSubmitScrubMinipools(c *cli.Context, logger log.ColorLogger) (*submitScrubMinipools, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil { return nil, err }
	w, err := services.GetWallet(c)
	if err != nil { return nil, err }
	rp, err := services.GetRocketPool(c)
	if err != nil { return nil, err }
	bc, err := services.GetBeaconClient(c)
	if err != nil { return nil, err }

	// Return task
	return &submitScrubMinipools{
		c: c,
		log: logger,
		cfg: cfg,
		w: w,
		rp: rp,
		bc: bc,
	}, nil

}


// Submit scrub minipools
func (t *submitScrubMinipools) run() error {

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Data
	var wg errgroup.Group
	var nodeTrusted bool
	var submitScrubEnabled bool

	// Get data
	wg.Go(func() error {
		var err error
		nodeTrusted, err = trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		submitScrubEnabled, err = protocol.GetMinipoolSubmitScrubEnabled(t.rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	// Check node trusted status & settings
	if !(nodeTrusted && submitScrubEnabled) {
		return nil
	}

	// Log
	t.log.Println("Checking for scrub worthy minipools...")

	// Get minipool prelaunch details
	minipools, err := t.getNetworkMinipoolPrelaunchDetails(nodeAccount.Address)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are scrub worthy...", len(minipools))

	// Submit minipools scrub status
	for _, details := range minipools {
		if err := t.submitScrubMinipool(details); err != nil {
			t.log.Println(fmt.Errorf("Could not scrub minipool %s: %w", details.Address.Hex(), err))
		}
	}

	// Return
	return nil

}


// Get all minipool prelaunch details
func (t *submitScrubMinipools) getNetworkMinipoolPrelaunchDetails(nodeAddress common.Address) ([]minipoolPrelaunchDetails, error) {

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetMinipoolAddresses(t.rp, nil)
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return []minipoolPrelaunchDetails{}, err
	}

	// Get minipool validator statuses
	validators, err := rp.GetMinipoolValidators(t.rp, t.bc, addresses, nil, nil)
	if err != nil {
		return []minipoolPrelaunchDetails{}, err
	}

	// Load details in batches
	minipools := make([]minipoolPrelaunchDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolPrelaunchDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolPrelaunchDetailsBatchSize
		if mei > len(addresses) { mei = len(addresses) }

		// Log
		//t.log.Printlnf("Checking minipools %d - %d of %d for prelaunch status...", msi + 1, mei, len(addresses))

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				mpDetails, err := t.getMinipoolPrelaunchDetails(nodeAddress, address)
				if err == nil { minipools[mi] = mpDetails }
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []minipoolPrelaunchDetails{}, err
		}

	}

	// Filter by prelaunch status
	prelaunchMinipools := []minipoolPrelaunchDetails{}
	for _, details := range minipools {
		if details.Prelaunch {
			prelaunchMinipools = append(prelaunchMinipools, details)
		}
	}

	// Return
	return prelaunchMinipools, nil

}


// Get minipool prelaunch details
func (t *submitScrubMinipools) getMinipoolPrelaunchDetails(nodeAddress common.Address, minipoolAddress common.Address) (minipoolPrelaunchDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
	if err != nil {
		return minipoolPrelaunchDetails{}, err
	}

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		return err
	})

	wg.Go(func() error {
		var err error
		withdrawalCredentials, err = mp.GetWithdrawalCredentials(nodeAddress)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolPrelaunchDetails{}, err
	}

	// Check minipool status
	if status == types.Prelaunch {
		// Return
		return minipoolPrelaunchDetails{
			Address: minipoolAddress,
			WithdrawalCredentials: withdrawalCredentials,
			validatorPubkey: validatorPubkey
		}, nil
	}

	return minipoolPrelaunchDetails{}, nil
}


// Submit minipool scrub status
func (t *submitScrubMinipools) submitScrubMinipool(details minipoolPrelauchDetails) error {

	// Log
	t.log.Printlnf("Submitting minipool %s scrub status...", details.Address.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas estimates
	gasInfo, err := minipool.EstimateSubmitMinipoolScrubGas(t.rp, details.Address, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit minipool scrub status: %w", err)
	}
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log) {
		return nil
	}

	// Scrub
	hash, err := minipool.SubmitMinipoolScrub(t.rp, details.Address, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be mined
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted minipool %s scrub status.", details.Address.Hex())

	// Return
	return nil

}

