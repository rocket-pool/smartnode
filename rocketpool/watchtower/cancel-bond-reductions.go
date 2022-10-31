package watchtower

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

type cancelBondReductions struct {
	c                *cli.Context
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               rocketpool.ExecutionClient
	bc               beacon.Client
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
}

// Create cancel bond reductions task
func newCancelBondReductions(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger) (*cancelBondReductions, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	lock := &sync.Mutex{}
	return &cancelBondReductions{
		c:                c,
		log:              logger,
		errLog:           errorLogger,
		cfg:              cfg,
		w:                w,
		rp:               rp,
		ec:               ec,
		bc:               bc,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Bond Reduction]",
	}, nil

}

// Start the bond reduction cancellation thread
func (t *cancelBondReductions) run() error {

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

	// Get trusted node status
	nodeTrusted, err := trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}
	if !(nodeTrusted) {
		return nil
	}

	// Log
	t.log.Println("Checking for bond reductions to cancel...")

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Bond reduction cancel check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		t.printMessage("Starting bond reduction cancel check in a separate thread.")

		err := t.checkBondReductions()
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", t.generationPrefix, err))
			return
		}

		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

}

// Check for bond reductions to cancel
func (t *cancelBondReductions) checkBondReductions() error {

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var eth2Config beacon.Eth2Config
	var beaconHead beacon.BeaconHead

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = t.bc.GetEth2Config()
		if err != nil {
			return fmt.Errorf("error getting Beacon config: %w", err)
		}
		return nil
	})

	// Get beacon head
	wg1.Go(func() error {
		var err error
		beaconHead, err = t.bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("error getting Beacon head: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return err
	}

	// Get the latest finalized slot that exists, and the EL block for it
	finalizedEpoch := beaconHead.FinalizedEpoch
	lastSlot := (finalizedEpoch+1)*eth2Config.SlotsPerEpoch - 1
	var elBlock uint64
	for lastSlot > 0 {
		block, exists, err := t.bc.GetBeaconBlock(fmt.Sprint(lastSlot))
		if err != nil {
			return fmt.Errorf("error getting Beacon block %d: %w", lastSlot, err)
		}
		if !exists {
			lastSlot--
		} else {
			elBlock = block.ExecutionBlockNumber
			break
		}
	}

	t.printMessage(fmt.Sprintf("Latest finalized epoch is %d, checking for Beacon slot %d (EL block %d)", finalizedEpoch, lastSlot, elBlock))
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlock),
	}

	// Get minipool addresses
	addresses, err := minipool.GetMinipoolAddresses(t.rp, opts)
	if err != nil {
		return fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Get the bond reduction time of each MP
	reductionMpTimes := make([]int64, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolBalanceDetailsBatchSize {
		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolBalanceDetailsBatchSize
		if mei > len(addresses) {
			mei = len(addresses)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				mp, err := minipool.NewMinipool(t.rp, address)
				if err != nil {
					return fmt.Errorf("error creating binding for minipool %s: %w", address.Hex(), err)
				}

				rawTime, err := mp.GetReduceBondTime(opts)
				if err != nil {
					return fmt.Errorf("error getting bond reduction time for minipool %s: %w", address.Hex(), err)
				}
				reductionMpTimes[mi] = rawTime.Int64()
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return err
		}
	}

	// Check if any of them have bond reduction requests
	reductionMps := []common.Address{}
	for i := 0; i < len(addresses); i++ {
		if reductionMpTimes[i] > 0 {
			reductionMps = append(reductionMps, addresses[i])
		}
	}

	// If there aren't any, return
	if len(reductionMps) == 0 {
		t.printMessage("No minipools have requested a bond reduction.")
		return nil
	}

	// Get the statuses of minipools with bond reductions
	validators, err := rp.GetMinipoolValidators(t.rp, t.bc, reductionMps, opts, &beacon.ValidatorStatusOptions{Slot: &lastSlot})
	if err != nil {
		return fmt.Errorf("error getting bond reduction validator statuses: %w", err)
	}

	// Check the status of each one
	threshold := uint64(32000000000)
	for address, status := range validators {
		switch status.Status {
		case beacon.ValidatorState_PendingInitialized,
			beacon.ValidatorState_PendingQueued:
			// Do nothing because this validator isn't live yet
			continue

		case beacon.ValidatorState_ActiveOngoing:
			// Check the balance
			if status.Balance < threshold {
				// Cancel because it's under-balance
				t.cancelBondReduction(address, fmt.Sprintf("minipool balance is %d (below the threshold)", status.Balance))
			}

		case beacon.ValidatorState_ActiveExiting,
			beacon.ValidatorState_ActiveSlashed,
			beacon.ValidatorState_ExitedUnslashed,
			beacon.ValidatorState_ExitedSlashed,
			beacon.ValidatorState_WithdrawalPossible,
			beacon.ValidatorState_WithdrawalDone:
			t.cancelBondReduction(address, "minipool is already slashed, exiting, or exited")

		default:
			return fmt.Errorf("unknown validator state: %v", status.Status)
		}
	}

	return nil

}

// Cancel a bond reduction
func (t *cancelBondReductions) cancelBondReduction(address common.Address, reason string) error {

	// Log
	t.printMessage("=== CANCELLING BOND REDUCTION ===")
	t.printMessage(fmt.Sprintf("Minipool: %s", address.Hex()))
	t.printMessage(fmt.Sprintf("Reason:   %s", reason))
	t.printMessage("=================================")

	// Make the binding
	mp, err := minipool.NewMinipool(t.rp, address)
	if err != nil {
		return fmt.Errorf("error creating binding for minipool %s: %w", address.Hex(), err)
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := mp.EstimateVoteCancelReductionGas(opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to voteCancelReduction the minipool: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(WatchtowerMaxFee)
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(WatchtowerMaxPriorityFee)
	opts.GasLimit = gasInfo.SafeGasLimit

	// Cancel the reduction
	hash, err := mp.VoteCancelReduction(opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully voted to cancel the bond reduction of minipool %s.", mp.Address.Hex())

	// Return
	return nil

}

func (t *cancelBondReductions) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Bond reduction cancel check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Print a message from the tree generation goroutine
func (t *cancelBondReductions) printMessage(message string) {
	t.log.Printlnf("%s %s", t.generationPrefix, message)
}
