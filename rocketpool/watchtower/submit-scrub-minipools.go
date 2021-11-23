package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prysmaticlabs/prysm/v2/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v2/contracts/deposit"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	ethpb "github.com/prysmaticlabs/prysm/v2/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
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
    c *cli.Context
    log log.ColorLogger
    cfg config.RocketPoolConfig
    w *wallet.Wallet
    rp *rocketpool.RocketPool
    ec *ethclient.Client 
    bc beacon.Client
    it *iterationData
    coll *collectors.ScrubCollector
    maxFee *big.Int
    maxPriorityFee *big.Int
    gasLimit uint64
}


type iterationData struct {
    // Counters
    totalMinipools int
    goodOnBeaconCount int
    badOnBeaconCount int
    goodPrestakeCount int
    badPrestakeCount int
    goodOnDepositContract int
    badOnDepositContract int
    unknownMinipools int
    safetyScrubs int

    // Minipool info
    minipools map[*minipool.Minipool]*minipoolDetails

    // ETH1 search artifacts
    startBlock *big.Int
    eventLogInterval *big.Int
    depositDomain []byte
    latestBlockTime time.Time
}


type minipoolDetails struct {
    pubkey types.ValidatorPubkey
    expectedWithdrawalCredentials common.Hash
}


// Create submit scrub minipools task
func newSubmitScrubMinipools(c *cli.Context, logger log.ColorLogger, coll *collectors.ScrubCollector) (*submitScrubMinipools, error) {

    // Get services
    cfg, err := services.GetConfig(c)
    if err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Get the user-requested max fee
    maxFee, err := cfg.GetMaxFee()
    if err != nil {
        return nil, fmt.Errorf("Error getting max fee in configuration: %w", err)
    }

    // Get the user-requested max fee
    maxPriorityFee, err := cfg.GetMaxPriorityFee()
    if err != nil {
        return nil, fmt.Errorf("Error getting max priority fee in configuration: %w", err)
    }
    if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
        logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.");
        maxPriorityFee = big.NewInt(2)
    }

    // Get the user-requested gas limit
    gasLimit, err := cfg.GetGasLimit()
    if err != nil {
        return nil, fmt.Errorf("Error getting gas limit in configuration: %w", err)
    }

    // Return task
    return &submitScrubMinipools{
        c: c,
        log: logger,
        cfg: cfg,
        w: w,
        rp: rp,
        ec: ec,
        bc: bc,
        coll: coll,
        maxFee: maxFee,
        maxPriorityFee: maxPriorityFee,
        gasLimit: gasLimit,
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

    // Get trusted node status
    nodeTrusted, err := trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
    if err != nil {
        return err
    }
    if !(nodeTrusted) {
        return nil
    }

    // Log
    t.log.Println("Checking for minipools to scrub...")
    t.it = new(iterationData)

    // Get minipools in prelaunch status
    minipoolAddresses, err := minipool.GetPrelaunchMinipoolAddresses(t.rp, nil)
    if err != nil {
        return err
    }
    t.it.totalMinipools = len(minipoolAddresses)
    if t.it.totalMinipools == 0 {
        t.log.Println("No minipools in prelaunch.")
        return nil
    }
    t.it.minipools = make(map[*minipool.Minipool]*minipoolDetails, t.it.totalMinipools)

    // Get the correct withdrawal credentials and validator pubkeys for each minipool
    pubkeys := t.initializeMinipoolDetails(minipoolAddresses)

    // Step 1: Verify the Beacon credentials if they exist
    err = t.verifyBeaconWithdrawalCredentials(pubkeys)
    if err != nil {
        return err
    }

    // If there aren't any minipools left to check, print the final tally and exit
    if len(t.it.minipools) == 0 {
        t.printFinalTally()
        return nil
    }

    // Get various elements needed to do eth1 prestake and deposit contract searches
    err = t.getEth1SearchArtifacts()
    if err != nil {
        return err
    }

    // Step 2: Verify the MinipoolPrestaked events
    t.verifyPrestakeEvents()

    // If there aren't any minipools left to check, print the final tally and exit
    if len(t.it.minipools) == 0 {
        t.printFinalTally()
        return nil
    }
    
    // Step 3: Verify the deposit data of the remaining minipools
    err = t.verifyDeposits()
    if err != nil {
        return err
    }

    // If there aren't any minipools left to check, print the final tally and exit
    if len(t.it.minipools) == 0 {
        t.printFinalTally()
        return nil
    }

    // Step 4: Scrub all of the undeposited minipools after half the scrub period for safety
    err = t.checkSafetyScrub()
    if err != nil {
        return err
    }

    // Log and return
    t.printFinalTally()
    t.it = nil
    return nil

}


// Get the correct withdrawal credentials and pubkeys for each minipool
func (t *submitScrubMinipools) initializeMinipoolDetails(minipoolAddresses []common.Address) ([]types.ValidatorPubkey) {
    
    pubkeys := []types.ValidatorPubkey{}

    for _, minipoolAddress := range minipoolAddresses {
        // Create a minipool contract wrapper for the given address
        mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
        if err != nil {
            t.log.Printf("Error creating minipool wrapper for %s: %s", minipoolAddress.Hex(), err.Error())
            continue
        }

        // Get the correct withdrawal credentials
        expectedCreds, err := minipool.GetMinipoolWithdrawalCredentials(t.rp, minipoolAddress, nil)
        if err != nil {
            t.log.Printf("Error getting expected withdrawal creds for minipool %s: %s", minipoolAddress.Hex(), err.Error())
            continue
        }

        // Get the validator pubkey
        pubkey, err := minipool.GetMinipoolPubkey(t.rp, minipoolAddress, nil)
        if err != nil {
            t.log.Printf("Error getting validator pubkey for minipool %s: %s", minipoolAddress.Hex(), err.Error())
            continue
        }
        pubkeys = append(pubkeys, pubkey)

        // Create a new details entry for this minipool
        t.it.minipools[mp] = &minipoolDetails{
            expectedWithdrawalCredentials: expectedCreds,
            pubkey: pubkey,
        }
    }

    return pubkeys

}


// Step 1: Verify the Beacon Chain credentials for a minipool if they're present
func (t *submitScrubMinipools) verifyBeaconWithdrawalCredentials(pubkeys []types.ValidatorPubkey) (error) {

    minipoolsToScrub := []*minipool.Minipool{}

    // Get the status of the validators on the Beacon chain
    statuses, err := t.bc.GetValidatorStatuses(pubkeys, nil)
    if err != nil {
        return err
    }

    // Get the withdrawal credentials on Beacon for each validator if they exist
    for minipool, details := range t.it.minipools {
        pubkey := details.pubkey

        status := statuses[pubkey]
        if status.Exists {
            // This minipool's deposit has been seen on the Beacon Chain
            expectedCreds := details.expectedWithdrawalCredentials
            beaconCreds := status.WithdrawalCredentials
            if beaconCreds != expectedCreds {
                t.log.Println("=== SCRUB DETECTED ON BEACON CHAIN ===")
                t.log.Printlnf("\tMinipool: %s", minipool.Address.Hex())
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
        err = t.submitVoteScrubMinipool(minipool)
        if err != nil {
            t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Address.Hex(), err.Error())
        }
    }

    return nil

}


// Get various elements needed to do eth1 prestake and deposit contract searches
func (t *submitScrubMinipools) getEth1SearchArtifacts() (error) {

    // Get the starting eth1 block to search from
    /*
    data, err := t.bc.GetEth1DataForEth2Block("finalized")
    if err != nil {
        return nil, err
    }
    
    latestEth1Block, err := t.ec.BlockByHash(context.Background(), data.BlockHash)
    if err != nil {
        return nil, err
    }
    */
    latestEth1Block, err := t.ec.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return err
    }
    t.it.latestBlockTime = time.Unix(int64(latestEth1Block.Time), 0)
    targetBlockNumber := big.NewInt(0).Sub(latestEth1Block.Number, big.NewInt(BlockStartOffset))
    targetBlock, err := t.ec.HeaderByNumber(context.Background(), targetBlockNumber)
    if err != nil {
        return err
    }
    t.it.startBlock = targetBlock.Number

    // Check the prestake event from the minipool and validate its signature
    eventLogInterval, err := api.GetEventLogInterval(t.cfg)
    if err != nil {
        return err
    }
    t.it.eventLogInterval = eventLogInterval

    // Put together the signature validation data
    eth2Config, err := t.bc.GetEth2Config()
    if err != nil {
        return err
    }
    depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
    if err != nil {
        return err
    }
    t.it.depositDomain = depositDomain

    return nil

}


// Step 2: Verify the MinipoolPrestaked event of each minipool
func (t *submitScrubMinipools) verifyPrestakeEvents() () {

    minipoolsToScrub := []*minipool.Minipool{}

    weiPerGwei := big.NewInt(int64(eth.WeiPerGwei))
    for minipool := range t.it.minipools {
        // Get the MinipoolPrestaked event
        prestakeData, err := minipool.GetPrestakeEvent(t.it.eventLogInterval, nil)
        if err != nil {
            t.log.Printlnf("Error getting prestake event for minipool %s: %s", minipool.Address.Hex(), err.Error())
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
            t.log.Printlnf("Invalid prestake data for minipool %s:", minipool.Address.Hex())
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
            t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Address.Hex(), err.Error())
        }
    }

    return

}


// Step 3: Verify minipools by their deposits
func (t *submitScrubMinipools) verifyDeposits() (error) {

    minipoolsToScrub := []*minipool.Minipool{}

    // Create a "hashset" of the remaining pubkeys
    pubkeys := make(map[types.ValidatorPubkey]bool, len(t.it.minipools))
    for _, details := range t.it.minipools {
        pubkeys[details.pubkey] = true
    }

    // Get the deposits from the deposit contract
    depositMap, err := utils.GetDeposits(t.rp, pubkeys, t.it.startBlock, t.it.eventLogInterval, nil)
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
                t.log.Printlnf("Invalid deposit for minipool %s:", minipool.Address.Hex())
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
                    t.log.Printlnf("\tMinipool: %s", minipool.Address.Hex())
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
            t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Address.Hex(), err.Error())
        }
    }

    return nil

}


// Step 4: Catch-all safety mechanism that scrubs minipools without valid deposits after a certain period of time
// This should never be used, it's simply here as a redundant check
func (t *submitScrubMinipools) checkSafetyScrub() (error) {

    minipoolsToScrub := []*minipool.Minipool{}

    // Warn if there are any remaining minipools - this should never happen
    remainingMinipools := len(t.it.minipools)
    if remainingMinipools > 0 {
        t.log.Printlnf("WARNING: %d minipools did not have deposit information", remainingMinipools)
    } else {
        return nil
    }

    // Get the scrub period
    scrubPeriodUint, err := tnsettings.GetScrubPeriod(t.rp, nil)
    if err != nil {
        return err
    }
    scrubPeriod := time.Duration(scrubPeriodUint) * time.Second
    
    // Get the safety period where minipools can be scrubbed without a valid deposit
    safetyPeriod := scrubPeriod / ScrubSafetyDivider
    if safetyPeriod < MinScrubSafetyTime {
        safetyPeriod = MinScrubSafetyTime
    }

    for minipool := range t.it.minipools {
        // Get the minipool's status
        statusDetails, err := minipool.GetStatusDetails(nil)
        if err != nil {
            t.log.Printlnf("Error getting status for minipool %s: %s", minipool.Address.Hex(), err.Error())
            continue
        }

        // Verify this is actually a prelaunch minipool
        if statusDetails.Status != types.Prelaunch {
            t.log.Printlnf("\tMinipool %s is under review but is in %d status?", minipool.Address.Hex(), types.MinipoolDepositTypes[statusDetails.Status])
            continue
        }

        // Check the time it entered prelaunch against the safety period
        if (t.it.latestBlockTime.Sub(statusDetails.StatusTime)) > safetyPeriod {
            t.log.Println("=== SAFETY SCRUB DETECTED ===")
            t.log.Printlnf("\tMinipool: %s", minipool.Address.Hex())
            t.log.Printlnf("\tTime since prelaunch: %s", time.Since(statusDetails.StatusTime))
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
            t.log.Printlnf("ALERT: Couldn't scrub minipool %s: %s", minipool.Address.Hex(), err.Error())
        }
    }

    return nil

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

    // Get the gas limit
    gasInfo, err := mp.EstimateVoteScrubGas(opts)
    if err != nil {
        return fmt.Errorf("Could not estimate the gas required to voteScrub the minipool: %w", err)
    }
    var gas *big.Int 
    if t.gasLimit != 0 {
        gas = new(big.Int).SetUint64(t.gasLimit)
    } else {
        gas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
    }

    // Get the max fee
    maxFee := t.maxFee
    if maxFee == nil || maxFee.Uint64() == 0 {
        maxFee, err = rpgas.GetHeadlessMaxFeeWei()
        if err != nil {
            return err
        }
    }

    // Print the gas info
    if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, t.gasLimit) {
        return nil
    }

    opts.GasFeeCap = maxFee
    opts.GasTipCap = t.maxPriorityFee
    opts.GasLimit = gas.Uint64()

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


// Prints the final tally of minipool counts
func (t *submitScrubMinipools) printFinalTally() {

    t.log.Println("Scrub check complete.")
    t.log.Printlnf("\tTotal prelaunch minipools: %d", t.it.totalMinipools)
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
        t.coll.LatestBlockTime = float64(t.it.latestBlockTime.Unix())
    }
}

