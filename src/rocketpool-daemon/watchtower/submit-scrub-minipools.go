package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v4/contracts/deposit"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	rputils "github.com/rocket-pool/rocketpool-go/utils"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"

	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/collectors"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/config"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Settings
const MinipoolBatchSize = 20
const BlockStartOffset = 100000
const ScrubSafetyDivider = 2
const MinScrubSafetyTime = time.Duration(0) * time.Hour

// Submit scrub minipools task
type SubmitScrubMinipools struct {
	sp        *services.ServiceProvider
	log       log.ColorLogger
	errLog    log.ColorLogger
	cfg       *config.SmartNodeConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	ec        eth.IExecutionClient
	bc        beacon.IBeaconClient
	mpMgr     *minipool.MinipoolManager
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
	minipools map[minipool.IMinipool]*minipoolDetails

	// ETH1 search artifacts
	startBlock       *big.Int
	eventLogInterval *big.Int
	depositDomain    []byte
	stateBlockTime   time.Time
}

type minipoolDetails struct {
	pubkey                        beacon.ValidatorPubkey
	expectedWithdrawalCredentials common.Hash
}

// Create submit scrub minipools task
func NewSubmitScrubMinipools(sp *services.ServiceProvider, logger log.ColorLogger, errorLogger log.ColorLogger, coll *collectors.ScrubCollector) *SubmitScrubMinipools {
	lock := &sync.Mutex{}
	return &SubmitScrubMinipools{
		sp:        sp,
		log:       logger,
		errLog:    errorLogger,
		coll:      coll,
		lock:      lock,
		isRunning: false,
	}
}

// Submit scrub minipools
func (t *SubmitScrubMinipools) Run(state *state.NetworkState) error {
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

		// Get services
		t.cfg = t.sp.GetConfig()
		t.w = t.sp.GetWallet()
		t.rp = t.sp.GetRocketPool()
		t.ec = t.sp.GetEthClient()
		t.bc = t.sp.GetBeaconClient()
		var err error
		t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
		if err != nil {
			t.handleError(fmt.Errorf("error creating minipool manager: %w", err))
			return
		}
		t.it = new(iterationData)

		// Get minipools in prelaunch status
		prelaunchMinipools := []rpstate.NativeMinipoolDetails{}
		for _, mpd := range state.MinipoolDetails {
			if mpd.Status == types.MinipoolStatus_Prelaunch {
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

		t.it.minipools = make(map[minipool.IMinipool]*minipoolDetails, t.it.totalMinipools)

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
		err = t.getEth1SearchArtifacts(state)
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

func (t *SubmitScrubMinipools) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Minipool scrub check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Get the correct withdrawal credentials and pubkeys for each minipool
func (t *SubmitScrubMinipools) initializeMinipoolDetails(minipools []rpstate.NativeMinipoolDetails, opts *bind.CallOpts) {
	for _, mpd := range minipools {
		// Ignore vacant minipools - they have the wrong withdrawal creds (temporarily) by design
		if mpd.IsVacant {
			t.it.vacantMinipools++
			continue
		}

		// Create a minipool contract wrapper for the given address
		mp, err := t.mpMgr.NewMinipoolFromVersion(mpd.MinipoolAddress, mpd.Version)
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
func (t *SubmitScrubMinipools) verifyBeaconWithdrawalCredentials(state *state.NetworkState) error {
	minipoolsToScrub := []minipool.IMinipool{}

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
				t.log.Printlnf("\tMinipool: %s", minipool.Common().Address.Hex())
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
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Common().Address.Hex(), err.Error())
		}
	}

	return nil
}

// Get various elements needed to do eth1 prestake and deposit contract searches
func (t *SubmitScrubMinipools) getEth1SearchArtifacts(state *state.NetworkState) error {
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
	t.it.eventLogInterval = big.NewInt(int64(config.EventLogInterval))

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
func (t *SubmitScrubMinipools) verifyPrestakeEvents() {
	minipoolsToScrub := []minipool.IMinipool{}

	weiPerGwei := big.NewInt(int64(eth.WeiPerGwei))
	for minipool := range t.it.minipools {
		// Get the MinipoolPrestaked event
		prestakeData, err := minipool.Common().GetPrestakeEvent(t.it.eventLogInterval, nil)
		if err != nil {
			t.log.Printlnf("Error getting prestake event for minipool %s: %s", minipool.Common().Address.Hex(), err.Error())
			continue
		}

		// Convert the amount to gwei
		prestakeData.Amount.Div(prestakeData.Amount, weiPerGwei)

		// Convert it into Prysm's deposit data struct
		depositData := new(ethpb.Deposit_Data)
		depositData.Amount = prestakeData.Amount.Uint64()
		depositData.PublicKey = prestakeData.Pubkey[:]
		depositData.WithdrawalCredentials = prestakeData.WithdrawalCredentials.Bytes()
		depositData.Signature = prestakeData.Signature[:]

		// Validate the signature
		err = prdeposit.VerifyDepositSignature(depositData, t.it.depositDomain)
		if err != nil {
			// The signature is illegal
			t.log.Println("=== SCRUB DETECTED ON PRESTAKE EVENT ===")
			t.log.Printlnf("Invalid prestake data for minipool %s:", minipool.Common().Address.Hex())
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
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Common().Address.Hex(), err.Error())
		}
	}
}

// Step 3: Verify minipools by their deposits
func (t *SubmitScrubMinipools) verifyDeposits() error {
	minipoolsToScrub := []minipool.IMinipool{}

	// Create a "hashset" of the remaining pubkeys
	pubkeys := make(map[beacon.ValidatorPubkey]bool, len(t.it.minipools))
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
			depositData.PublicKey = deposit.Pubkey[:]
			depositData.WithdrawalCredentials = deposit.WithdrawalCredentials.Bytes()
			depositData.Signature = deposit.Signature[:]

			err := prdeposit.VerifyDepositSignature(depositData, t.it.depositDomain)
			if err != nil {
				// This isn't a valid deposit, so ignore it
				t.log.Printlnf("Invalid deposit for minipool %s:", minipool.Common().Address.Hex())
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
					t.log.Printlnf("\tMinipool: %s", minipool.Common().Address.Hex())
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
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Common().Address.Hex(), err.Error())
		}
	}

	return nil
}

// Step 4: Catch-all safety mechanism that scrubs minipools without valid deposits after a certain period of time
// This should never be used, it's simply here as a redundant check
func (t *SubmitScrubMinipools) checkSafetyScrub(state *state.NetworkState) error {
	minipoolsToScrub := []minipool.IMinipool{}

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
		mpd := state.MinipoolDetailsByAddress[minipool.Common().Address]

		// Verify this is actually a prelaunch minipool
		if mpd.Status != types.MinipoolStatus_Prelaunch {
			t.log.Printlnf("\tMinipool %s is under review but is in %d status?", minipool.Common().Address.Hex(), types.MinipoolDepositTypes[mpd.Status])
			continue
		}

		// Check the time it entered prelaunch against the safety period
		statusTime := time.Unix(mpd.StatusTime.Int64(), 0)
		if t.it.stateBlockTime.Sub(statusTime) > safetyPeriod {
			t.log.Println("=== SAFETY SCRUB DETECTED ===")
			t.log.Printlnf("\tMinipool: %s", minipool.Common().Address.Hex())
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
			t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Common().Address.Hex(), err.Error())
		}
	}

	return nil
}

// Submit minipool scrub status
func (t *SubmitScrubMinipools) submitVoteScrubMinipool(mp minipool.IMinipool) error {
	// Log
	t.log.Printlnf("Voting to scrub minipool %s...", mp.Common().Address.Hex())

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return err
	}

	// Get the TX info
	txInfo, err := mp.Common().VoteScrub(opts)
	if err != nil {
		return fmt.Errorf("error getting voteScrub TX: %w", err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating voteScrub tx for minipool %s failed: %s", mp.Common().Address.Hex(), txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully voted to scrub the minipool %s.", mp.Common().Address.Hex())
	return nil
}

// Prints the final tally of minipool counts
func (t *SubmitScrubMinipools) printFinalTally(prefix string) {
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
