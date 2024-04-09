package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	rputils "github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/urfave/cli"

	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Settings
const MinipoolBatchSize = 20
const BlockStartOffset = 100000
const ScrubSafetyDivider = 2
const MinScrubSafetyTime = time.Duration(0) * time.Hour

// Submit scrub minipools task
type submitScrubMinipools struct {
	c         *cli.Context
	log       log.ColorLogger
	errLog    log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	ec        rocketpool.ExecutionClient
	bc        beacon.Client
	it        *iterationData
	coll      *collectors.ScrubCollector
	lock      *sync.Mutex
	isRunning bool
}

type iterationData struct {
	// Counters
	totalMinipools        int
	vacantMinipools       int
	goodOnBeaconCount     int
	badOnBeaconCount      int
	goodPrestakeCount     int
	badPrestakeCount      int
	goodOnDepositContract int
	badOnDepositContract  int
	unknownMinipools      int
	safetyScrubs          int

	// Minipool info
	minipools map[minipool.Minipool]*minipoolDetails

	// ETH1 search artifacts
	startBlock       *big.Int
	eventLogInterval *big.Int
	depositDomain    []byte
	stateBlockTime   time.Time
}

type minipoolDetails struct {
	pubkey                        types.ValidatorPubkey
	expectedWithdrawalCredentials common.Hash
}

// Create submit scrub minipools task
func newSubmitScrubMinipools(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, coll *collectors.ScrubCollector) (*submitScrubMinipools, error) {

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
	return &submitScrubMinipools{
		c:         c,
		log:       logger,
		errLog:    errorLogger,
		cfg:       cfg,
		w:         w,
		rp:        rp,
		ec:        ec,
		bc:        bc,
		coll:      coll,
		lock:      lock,
		isRunning: false,
	}, nil

}

// Submit scrub minipools
func (t *submitScrubMinipools) run(state *state.NetworkState) error {

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Log
	t.log.Println("Checking for minipools to scrub...")

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Scrub check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		checkPrefix := "[Minipool Scrub]"
		t.log.Printlnf("%s Starting scrub check in a separate thread.", checkPrefix)

		t.it = new(iterationData)

		// Get minipools in prelaunch status
		prelaunchMinipools := []rpstate.NativeMinipoolDetails{}
		for _, mpd := range state.MinipoolDetails {
			if mpd.Status == types.Prelaunch {
				prelaunchMinipools = append(prelaunchMinipools, mpd)
			}
		}

		t.it.totalMinipools = len(prelaunchMinipools)
		if t.it.totalMinipools == 0 {
			t.log.Printlnf("%s No minipools in prelaunch.", checkPrefix)
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		t.it.minipools = make(map[minipool.Minipool]*minipoolDetails, t.it.totalMinipools)

		// Get the correct withdrawal credentials and validator pubkeys for each minipool
		opts := &bind.CallOpts{
			BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
		}
		t.initializeMinipoolDetails(prelaunchMinipools, opts)

		// Step 1: Verify the Beacon credentials if they exist
		t.verifyBeaconWithdrawalCredentials(state)

		// If there aren't any minipools left to check, print the final tally and exit
		if len(t.it.minipools) == 0 {
			t.printFinalTally(checkPrefix)
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		// Get various elements needed to do eth1 prestake and deposit contract searches
		err := t.getEth1SearchArtifacts(state)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", checkPrefix, err))
			return
		}

		// Step 2: Verify the MinipoolPrestaked events
		t.verifyPrestakeEvents()

		// If there aren't any minipools left to check, print the final tally and exit
		if len(t.it.minipools) == 0 {
			t.printFinalTally(checkPrefix)
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		// Step 3: Verify the deposit data of the remaining minipools
		err = t.verifyDeposits()
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", checkPrefix, err))
			return
		}

		// If there aren't any minipools left to check, print the final tally and exit
		if len(t.it.minipools) == 0 {
			t.printFinalTally(checkPrefix)
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		// Step 4: Scrub all of the undeposited minipools after half the scrub period for safety
		err = t.checkSafetyScrub(state)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", checkPrefix, err))
			return
		}

		// Log and return
		t.printFinalTally(checkPrefix)
		t.it = nil
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

}

func (t *submitScrubMinipools) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Minipool scrub check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Get the correct withdrawal credentials and pubkeys for each minipool
func (t *submitScrubMinipools) initializeMinipoolDetails(minipools []rpstate.NativeMinipoolDetails, opts *bind.CallOpts) {
	for _, mpd := range minipools {
		// Ignore vacant minipools - they have the wrong withdrawal creds (temporarily) by design
		if mpd.IsVacant {
			t.it.vacantMinipools++
			continue
		}

		// Create a minipool contract wrapper for the given address
		mp, err := minipool.NewMinipoolFromVersion(t.rp, mpd.MinipoolAddress, mpd.Version, opts)
		if err != nil {
			t.log.Printf("Error creating minipool wrapper for %s: %s", mpd.MinipoolAddress.Hex(), err.Error())
			continue
		}

		// Create a new details entry for this minipool
		t.it.minipools[mp] = &minipoolDetails{
			expectedWithdrawalCredentials: mpd.WithdrawalCredentials,
			pubkey:                        mpd.Pubkey,
		}
	}
}

// Step 1: Verify the Beacon Chain credentials for a minipool if they're present
func (t *submitScrubMinipools) verifyBeaconWithdrawalCredentials(state *state.NetworkState) error {
	minipoolsToScrub := []minipool.Minipool{}

	// Get the withdrawal credentials on Beacon for each validator if they exist
	for minipool, details := range t.it.minipools {
		pubkey := details.pubkey

		status := state.ValidatorDetails[pubkey]
		if status.Exists {
			// This minipool's deposit has been seen on the Beacon Chain
			expectedCreds := details.expectedWithdrawalCredentials
			beaconCreds := status.WithdrawalCredentials
			if beaconCreds != expectedCreds {
				t.log.Println("=== SCRUB DETECTED ON BEACON CHAIN ===")
				t.log.Printlnf("\tMinipool: %s", minipool.GetAddress().Hex())
				t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
				t.log.Printlnf("\tActual creds: %s", beaconCreds.Hex())
				t.log.Println("======================================")
				minipoolsToScrub = append(minipoolsToScrub, minipool)
				t.it.badOnBeaconCount++
			} else {
				// This minipool's credentials match, it's clean.
				t.it.goodOnBeaconCount++
			}

			// If it was seen on Beacon we can remove it from the list of things to check on eth1.
			// Otherwise we have to keep it in the map.
			delete(t.it.minipools, minipool)
		}
	}

	// Scrub the offending minipools
	for _, minipool := range minipoolsToScrub {
		err := t.submitVoteScrubMinipool(minipool)
		if err != nil {
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.GetAddress().Hex(), err.Error())
		}
	}

	return nil
}

// Get various elements needed to do eth1 prestake and deposit contract searches
func (t *submitScrubMinipools) getEth1SearchArtifacts(state *state.NetworkState) error {

	// Get the time of the state's EL block
	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
	t.it.stateBlockTime = genesisTime.Add(secondsSinceGenesis)

	// Get the block to start searching the deposit contract from
	stateBlockNumber := big.NewInt(0).SetUint64(state.ElBlockNumber)
	offset := big.NewInt(BlockStartOffset)
	if stateBlockNumber.Cmp(offset) < 0 {
		offset = stateBlockNumber // Deal with chains that are younger than the look-behind interval
	}
	targetBlockNumber := big.NewInt(0).Sub(stateBlockNumber, offset)
	targetBlock, err := t.ec.HeaderByNumber(context.Background(), targetBlockNumber)
	if err != nil {
		return fmt.Errorf("error getting header for EL block %d: %w", targetBlockNumber, err)
	}
	t.it.startBlock = targetBlock.Number

	// Check the prestake event from the minipool and validate its signature
	eventLogInterval, err := t.cfg.GetEventLogInterval()
	if err != nil {
		return fmt.Errorf("error getting event log interval %w", err)
	}
	t.it.eventLogInterval = big.NewInt(int64(eventLogInterval))

	// Put together the signature validation data
	eth2Config := state.BeaconConfig
	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return fmt.Errorf("error computing deposit domain: %w", err)
	}
	t.it.depositDomain = depositDomain

	return nil

}

// Step 2: Verify the MinipoolPrestaked event of each minipool
func (t *submitScrubMinipools) verifyPrestakeEvents() {

	minipoolsToScrub := []minipool.Minipool{}

	weiPerGwei := big.NewInt(int64(eth.WeiPerGwei))
	for minipool := range t.it.minipools {
		// Get the MinipoolPrestaked event
		prestakeData, err := minipool.GetPrestakeEvent(t.it.eventLogInterval, nil)
		if err != nil {
			t.log.Printlnf("Error getting prestake event for minipool %s: %s", minipool.GetAddress().Hex(), err.Error())
			continue
		}

		// Convert the amount to gwei
		prestakeData.Amount.Div(prestakeData.Amount, weiPerGwei)

		// Convert it into Prysm's deposit data struct
		depositData := new(ethpb.Deposit_Data)
		depositData.Amount = prestakeData.Amount.Uint64()
		depositData.PublicKey = prestakeData.Pubkey.Bytes()
		depositData.WithdrawalCredentials = prestakeData.WithdrawalCredentials.Bytes()
		depositData.Signature = prestakeData.Signature.Bytes()

		// Validate the signature
		err = prdeposit.VerifyDepositSignature(depositData, t.it.depositDomain)
		if err != nil {
			// The signature is illegal
			t.log.Println("=== SCRUB DETECTED ON PRESTAKE EVENT ===")
			t.log.Printlnf("Invalid prestake data for minipool %s:", minipool.GetAddress().Hex())
			t.log.Printlnf("\tError: %s", err.Error())
			t.log.Println("========================================")

			// Remove this minipool from the list of things to process in the next step
			minipoolsToScrub = append(minipoolsToScrub, minipool)
			t.it.badPrestakeCount++
			delete(t.it.minipools, minipool)
		} else {
			// The signature is good, it can proceed to the next step
			t.it.goodPrestakeCount++
		}
	}

	// Scrub the offending minipools
	for _, minipool := range minipoolsToScrub {
		err := t.submitVoteScrubMinipool(minipool)
		if err != nil {
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.GetAddress().Hex(), err.Error())
		}
	}

}

// Step 3: Verify minipools by their deposits
func (t *submitScrubMinipools) verifyDeposits() error {

	minipoolsToScrub := []minipool.Minipool{}

	// Create a "hashset" of the remaining pubkeys
	pubkeys := make(map[types.ValidatorPubkey]bool, len(t.it.minipools))
	for _, details := range t.it.minipools {
		pubkeys[details.pubkey] = true
	}

	// Get the deposits from the deposit contract
	depositMap, err := rputils.GetDeposits(t.rp, pubkeys, t.it.startBlock, t.it.eventLogInterval, nil)
	if err != nil {
		return err
	}

	// Check each minipool's deposit data
	for minipool, details := range t.it.minipools {

		// Get the deposit list for this minipool
		deposits, exists := depositMap[details.pubkey]
		if !exists || len(deposits) == 0 {
			// Somehow this minipool doesn't have a deposit?
			t.it.unknownMinipools++
			continue
		}

		// Go through each deposit for this minipool and find the first one that's valid
		for depositIndex, deposit := range deposits {
			depositData := new(ethpb.Deposit_Data)
			depositData.Amount = deposit.Amount
			depositData.PublicKey = deposit.Pubkey.Bytes()
			depositData.WithdrawalCredentials = deposit.WithdrawalCredentials.Bytes()
			depositData.Signature = deposit.Signature.Bytes()

			err := prdeposit.VerifyDepositSignature(depositData, t.it.depositDomain)
			if err != nil {
				// This isn't a valid deposit, so ignore it
				t.log.Printlnf("Invalid deposit for minipool %s:", minipool.GetAddress().Hex())
				t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
				t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
				t.log.Printlnf("\tError: %s", err.Error())
			} else {
				// This is a valid deposit
				expectedCreds := details.expectedWithdrawalCredentials
				actualCreds := deposit.WithdrawalCredentials
				if actualCreds != expectedCreds {
					t.log.Println("=== SCRUB DETECTED ON DEPOSIT CONTRACT ===")
					t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
					t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
					t.log.Printlnf("\tMinipool: %s", minipool.GetAddress().Hex())
					t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
					t.log.Printlnf("\tActual creds: %s", actualCreds.Hex())
					t.log.Println("==========================================")
					minipoolsToScrub = append(minipoolsToScrub, minipool)
					t.it.badOnDepositContract++
				} else {
					t.it.goodOnDepositContract++
				}

				// Remove this minipool from the list of things to process in the next step
				delete(t.it.minipools, minipool)
				break
			}
		}
	}

	// Scrub the offending minipools
	for _, minipool := range minipoolsToScrub {
		err := t.submitVoteScrubMinipool(minipool)
		if err != nil {
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.GetAddress().Hex(), err.Error())
		}
	}

	return nil

}

// Step 4: Catch-all safety mechanism that scrubs minipools without valid deposits after a certain period of time
// This should never be used, it's simply here as a redundant check
func (t *submitScrubMinipools) checkSafetyScrub(state *state.NetworkState) error {

	minipoolsToScrub := []minipool.Minipool{}

	// Warn if there are any remaining minipools - this should never happen
	remainingMinipools := len(t.it.minipools)
	if remainingMinipools > 0 {
		t.log.Printlnf("WARNING: %d minipools did not have deposit information", remainingMinipools)
	} else {
		return nil
	}

	// Get the scrub period
	scrubPeriod := state.NetworkDetails.ScrubPeriod

	// Get the safety period where minipools can be scrubbed without a valid deposit
	safetyPeriod := scrubPeriod / ScrubSafetyDivider
	if safetyPeriod < MinScrubSafetyTime {
		safetyPeriod = MinScrubSafetyTime
	}

	for minipool := range t.it.minipools {
		// Get the minipool's status
		mpd := state.MinipoolDetailsByAddress[minipool.GetAddress()]

		// Verify this is actually a prelaunch minipool
		if mpd.Status != types.Prelaunch {
			t.log.Printlnf("\tMinipool %s is under review but is in %d status?", minipool.GetAddress().Hex(), types.MinipoolDepositTypes[mpd.Status])
			continue
		}

		// Check the time it entered prelaunch against the safety period
		statusTime := time.Unix(mpd.StatusTime.Int64(), 0)
		if t.it.stateBlockTime.Sub(statusTime) > safetyPeriod {
			t.log.Println("=== SAFETY SCRUB DETECTED ===")
			t.log.Printlnf("\tMinipool: %s", minipool.GetAddress().Hex())
			t.log.Printlnf("\tTime since prelaunch: %s", time.Since(statusTime))
			t.log.Printlnf("\tSafety scrub period: %s", safetyPeriod)
			t.log.Println("=============================")
			minipoolsToScrub = append(minipoolsToScrub, minipool)
			t.it.safetyScrubs++
			// Remove this minipool from the list of things to process in the next step
			delete(t.it.minipools, minipool)
		}
	}

	// Scrub the offending minipools
	for _, minipool := range minipoolsToScrub {
		err := t.submitVoteScrubMinipool(minipool)
		if err != nil {
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.GetAddress().Hex(), err.Error())
		}
	}

	return nil

}

// Submit minipool scrub status
func (t *submitScrubMinipools) submitVoteScrubMinipool(mp minipool.Minipool) error {

	// Log
	t.log.Printlnf("Voting to scrub minipool %s...", mp.GetAddress().Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := mp.EstimateVoteScrubGas(opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to voteScrub the minipool: %w", err)
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

	// Dissolve
	hash, err := mp.VoteScrub(opts)
	if err != nil {
		return fmt.Errorf("error voting to scrub minipool %s: %w", mp.GetAddress().Hex(), err)
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully voted to scrub the minipool %s.", mp.GetAddress().Hex())

	// Return
	return nil

}

// Prints the final tally of minipool counts
func (t *submitScrubMinipools) printFinalTally(prefix string) {

	t.log.Printlnf("%s Scrub check complete.", prefix)
	t.log.Printlnf("\tTotal prelaunch minipools: %d", t.it.totalMinipools)
	t.log.Printlnf("\tVacant minipools: %d", t.it.vacantMinipools)
	t.log.Printlnf("\tBeacon Chain scrubs: %d/%d", t.it.badOnBeaconCount, (t.it.badOnBeaconCount + t.it.goodOnBeaconCount))
	t.log.Printlnf("\tPrestake scrubs: %d/%d", t.it.badPrestakeCount, (t.it.badPrestakeCount + t.it.goodPrestakeCount))
	t.log.Printlnf("\tDeposit Contract scrubs: %d/%d", t.it.badOnDepositContract, (t.it.badOnDepositContract + t.it.goodOnDepositContract))
	t.log.Printlnf("\tPools without deposits: %d", t.it.unknownMinipools)
	t.log.Printlnf("\tRemaining uncovered minipools: %d", len(t.it.minipools))

	// Update the metrics collector
	if t.coll != nil {
		t.coll.UpdateLock.Lock()
		defer t.coll.UpdateLock.Unlock()

		t.coll.TotalMinipools = float64(t.it.totalMinipools)
		t.coll.GoodOnBeaconCount = float64(t.it.goodOnBeaconCount)
		t.coll.BadOnBeaconCount = float64(t.it.badOnBeaconCount)
		t.coll.GoodPrestakeCount = float64(t.it.goodPrestakeCount)
		t.coll.BadPrestakeCount = float64(t.it.badPrestakeCount)
		t.coll.GoodOnDepositContract = float64(t.it.goodOnDepositContract)
		t.coll.BadOnDepositContract = float64(t.it.badOnDepositContract)
		t.coll.DepositlessMinipools = float64(t.it.unknownMinipools)
		t.coll.UncoveredMinipools = float64(len(t.it.minipools))
		t.coll.LatestBlockTime = float64(t.it.stateBlockTime.Unix())
	}
}
