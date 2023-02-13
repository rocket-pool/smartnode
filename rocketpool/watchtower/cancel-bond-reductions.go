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
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

type cancelBondReductions struct {
	c                *cli.Context
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               rocketpool.ExecutionClient
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
	m                *state.NetworkStateManager
	s                *state.NetworkState
}

// Create cancel bond reductions task
func newCancelBondReductions(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, m *state.NetworkStateManager) (*cancelBondReductions, error) {

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
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Bond Reduction]",
		m:                m,
	}, nil

}

// Start the bond reduction cancellation thread
func (t *cancelBondReductions) run(isAtlasDeployed bool) error {

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Check if Atlas has been deployed yet
	if !isAtlasDeployed {
		return nil
	}

	// Get the latest state
	t.s = t.m.GetLatestState()
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(t.s.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get trusted node status
	nodeTrusted, err := trustednode.GetMemberExists(t.rp, nodeAccount.Address, opts)
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

	t.printMessage(fmt.Sprintf("Checking for Beacon slot %d (EL block %d)", t.s.BeaconSlotNumber, t.s.ElBlockNumber))

	// Check if any of the minipools have bond reduction requests
	zero := big.NewInt(0)
	reductionMps := []*rpstate.NativeMinipoolDetails{}
	for i, mpd := range t.s.MinipoolDetails {
		if mpd.ReduceBondTime.Cmp(zero) == 1 {
			reductionMps = append(reductionMps, &t.s.MinipoolDetails[i])
		}
	}

	// If there aren't any, return
	if len(reductionMps) == 0 {
		t.printMessage("No minipools have requested a bond reduction.")
		return nil
	}

	// Check the status of each one
	threshold := uint64(32000000000)
	for _, mpd := range reductionMps {
		validator := t.s.ValidatorDetails[mpd.Pubkey]
		switch validator.Status {
		case beacon.ValidatorState_PendingInitialized,
			beacon.ValidatorState_PendingQueued:
			// Do nothing because this validator isn't live yet
			continue

		case beacon.ValidatorState_ActiveOngoing:
			// Check the balance
			if validator.Balance < threshold {
				// Cancel because it's under-balance
				t.cancelBondReduction(mpd.MinipoolAddress, fmt.Sprintf("minipool balance is %d (below the threshold)", validator.Balance))
			}

		case beacon.ValidatorState_ActiveExiting,
			beacon.ValidatorState_ActiveSlashed,
			beacon.ValidatorState_ExitedUnslashed,
			beacon.ValidatorState_ExitedSlashed,
			beacon.ValidatorState_WithdrawalPossible,
			beacon.ValidatorState_WithdrawalDone:
			t.cancelBondReduction(mpd.MinipoolAddress, "minipool is already slashed, exiting, or exited")

		default:
			return fmt.Errorf("unknown validator state: %v", validator.Status)
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

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := minipool.EstimateVoteCancelReductionGas(t.rp, address, opts)
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
	hash, err := minipool.VoteCancelReduction(t.rp, address, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully voted to cancel the bond reduction of minipool %s.", address.Hex())

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
