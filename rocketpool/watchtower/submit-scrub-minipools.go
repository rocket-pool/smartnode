package watchtower

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Settings
const MinipoolPrelaunchStatusBatchSize = 20


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
	WithdrawalCredentials common.Hash
	ValidatorPubkey types.ValidatorPubkey
	Status types.MinipoolStatus
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

	// Get data
	wg.Go(func() error {
		var err error
		nodeTrusted, err = trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	// Check node trusted status
	if !(nodeTrusted) {
		return nil
	}

	// Log
	t.log.Println("Checking for scrub worthy minipools...")

	// Get minipools in prelaunch status
	prelaunchMinipools, err := t.getPrelaunchMinipools()
	if err != nil {
		return err
	}
	if len(prelaunchMinipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are in prelaunch status...", len(prelaunchMinipools))

	// Submit vote to scrub minipools
	for _, mp := range prelaunchMinipools {
		// TBD
		//if minipool.BeaconChainWithdrawalCredential != mp.Address {
		//	if mp.GetScrubVoted(mp.Address) == false {
		//		if err := t.submitVoteScrubMinipool(mp); err != nil {
		//			t.log.Println(fmt.Errorf("Could not scrub minipool %s: %w", mp.Address.Hex(), err))
		//		}
		//	}
		//}
	}

	// Return
	return nil

}


// Get all minipools in prelaunch status
func (t *submitScrubMinipools) getPrelaunchMinipools() ([]*minipool.Minipool, error) {

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
		return []*minipool.Minipool{}, err
	}

	// Create minipool contracts
	minipools := make([]*minipool.Minipool, len(addresses))
	for mi, address := range addresses {
		mp, err := minipool.NewMinipool(t.rp, address)
		if err != nil {
			return []*minipool.Minipool{}, err
		}
		minipools[mi] = mp
	}

	// Load minipool statuses in batches
	statuses := make([]minipool.StatusDetails, len(minipools))
	for bsi := 0; bsi < len(minipools); bsi += MinipoolPrelaunchStatusBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolPrelaunchStatusBatchSize
		if mei > len(minipools) { mei = len(minipools) }

		// Log
		//t.log.Printlnf("Checking minipools %d - %d of %d for prelaunch status...", msi + 1, mei, len(minipools))

		// Load statuses
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				mp := minipools[mi]
				status, err := mp.GetStatusDetails(nil)
				if err == nil { statuses[mi] = status }
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []*minipool.Minipool{}, err
		}

	}

	// Filter minipools by status
	prelaunchMinipools := []*minipool.Minipool{}
	for mi, mp := range minipools {
		if statuses[mi].Status == types.Prelaunch {
			prelaunchMinipools = append(prelaunchMinipools, mp)
		}
	}

	// Return
	return prelaunchMinipools, nil

}

// Submit minipool scrub status
func (t *submitScrubMinipools) submitVoteScrubMinipool(mp *minipool.Minipool) error {

	// Log
	t.log.Printlnf("Voting to scrub minipool %s...", mp.Address.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas estimates
	gasInfo, err := mp.EstimateVoteScrubGas(opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to voteScrub the minipool: %w", err)
	}
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log) {
		return nil
	}

	// Dissolve
	hash, err := mp.VoteScrub(opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be mined
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully voted to scrub the minipool %s.", mp.Address.Hex())

	// Return
	return nil

}

