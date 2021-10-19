package watchtower

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prysmaticlabs/prysm/v2/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v2/contracts/deposit"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils"
	rputils "github.com/rocket-pool/rocketpool-go/utils"
	"github.com/urfave/cli"

	ethpb "github.com/prysmaticlabs/prysm/v2/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Settings
const MinipoolPrelaunchStatusBatchSize = 20
const BlockStartOffset = 100000


// Submit scrub minipools task
type submitScrubMinipools struct {
    c *cli.Context
    log log.ColorLogger
    cfg config.RocketPoolConfig
    w *wallet.Wallet
    rp *rocketpool.RocketPool
    ec *ethclient.Client 
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
    ec, err := services.GetEthClient(c)
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
        ec: ec,
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

    // Get minipools in prelaunch status
    minipoolAddresses, err := minipool.GetPrelaunchMinipoolAddresses(t.rp, nil)
    if err != nil {
        return err
    }
    totalMinipools := len(minipoolAddresses)
    if totalMinipools == 0 {
        t.log.Println("No minipools in prelaunch.")
        return nil
    }

    // Get the correct withdrawal credentials for each minipool
    expectedWithdrawalCredentials, err := t.getExpectedWithdrawalCredentials(minipoolAddresses)
    if err != nil {
        return err
    }

    // Get the validator pubkeys for the minipools
    pubkeys, pubkeyMap, err := t.getPubkeys(minipoolAddresses)
    if err != nil {
        return err
    }

    // Get the withdrawal addresses that are defined on the Beacon Chain
    beaconWithdrawalCredentials, err := t.getCanonicalMinipoolWithdrawalAddresses(minipoolAddresses, pubkeys)
    if err != nil {
        return err
    }

    // Step 1: Verify the Beacon credentials if they exist
    minipoolsToVerifyPrestakeFor, beaconMinipoolsToScrub, goodOnBeaconCount, badOnBeaconCount := 
        t.checkBeaconCredentials(expectedWithdrawalCredentials, beaconWithdrawalCredentials)

    // Get the starting eth1 block to search
    startBlock, err := t.getDepositContractStartBlock()
    if err != nil {
        return err
    }

    // Check the prestake event from the minipool and validate its signature
    eventLogInterval, err := api.GetEventLogInterval(t.cfg)
    if err != nil {
        return err
    }

    // Put together the signature validation data
    eth2Config, err := t.bc.GetEth2Config()
    if err != nil {
        return err
    }
    depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
    if err != nil {
        return err
    }

    // Step 2: Validate the MinipoolPrestaked events
    minipoolsToCheckDepositsFor, prestakeMinipoolsToScrub, goodPrestakeCount, badPrestakeCount, err := 
        t.checkPrestakeEvents(minipoolsToVerifyPrestakeFor, eventLogInterval, depositDomain)
    if err != nil {
        return err
    }

    // Make a map of minipools to pubkeys
    pubkeysForDepositCheckMinipools := make([]types.ValidatorPubkey, len(minipoolsToCheckDepositsFor))
    for i, minipool := range minipoolsToCheckDepositsFor {
        pubkeysForDepositCheckMinipools[i] = pubkeyMap[minipool.Address]
    }

    // Make a map of minipools to deposit contract events
    depositMap, err := rputils.GetMinipoolDeposits(t.rp, minipoolsToCheckDepositsFor, pubkeysForDepositCheckMinipools,
        startBlock, eventLogInterval, nil)
    if err != nil {
        return err
    }

    // Step 3: Validate the deposit data of the remaining minipools
    depositMinipoolsToScrub, goodOnDepositContract, badOnDepositContract, unknownDeposits :=
        t.checkDeposits(depositMap, depositDomain, expectedWithdrawalCredentials)

    // Scrub beacon minipools
    for _, minipoolAddress := range beaconMinipoolsToScrub {
        mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
        if err != nil {
            return err
        }

        err = t.submitVoteScrubMinipool(mp)
        if err != nil {
            return err
        }
    }

    // Scrub prestake minipools
    for _, mp := range prestakeMinipoolsToScrub {
        err = t.submitVoteScrubMinipool(mp)
        if err != nil {
            return err
        }
    }

    // Scrub deposit minipools
    for _, mp := range depositMinipoolsToScrub {
        err = t.submitVoteScrubMinipool(mp)
        if err != nil {
            return err
        }
    }

    // Print final tally
    t.log.Println("Scrub check complete.")
    t.log.Printlnf("\tTotal prelaunch minipools: %d", totalMinipools)
    t.log.Printlnf("\tBeacon Chain scrubs: %d/%d", badOnBeaconCount, (badOnBeaconCount + goodOnBeaconCount))
    t.log.Printlnf("\tPrestake scrubs: %d/%d", badPrestakeCount, (badPrestakeCount + goodPrestakeCount))
    t.log.Printlnf("\tDeposit Contract scrubs: %d/%d", badOnDepositContract, (badOnDepositContract + goodOnDepositContract))
    t.log.Printlnf("\tPools without deposits: %d", unknownDeposits)

    // Return
    return nil

}


// Create a map of what the withdrawal credentials *should be*
func (t *submitScrubMinipools) getExpectedWithdrawalCredentials(addresses []common.Address) (map[common.Address]common.Hash, error) {
    
    expectedWithdrawalCredentials := make(map[common.Address]common.Hash)
    for _, minipoolAddress := range addresses {
        expectedCreds, err := minipool.GetMinipoolWithdrawalCredentials(t.rp, minipoolAddress, nil)
        if err != nil {
            return nil, err
        }
        expectedWithdrawalCredentials[minipoolAddress] = expectedCreds
    }

    return expectedWithdrawalCredentials, nil

}

// Get the eth1 block that matches the eth1 block last seen in the latest finalized Beacon block, minus the start offset
func (t *submitScrubMinipools) getDepositContractStartBlock() (*big.Int, error) {

    /*
    data, err := t.bc.GetEth1DataForEth2Block("finalized")
    if err != nil {
        return nil, err
    }
    
    eth1Block, err := t.ec.BlockByHash(context.Background(), data.BlockHash)
    if err != nil {
        return nil, err
    }
    */
    eth1Block, err := t.ec.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return nil, err
    }
    targetBlockNumber := big.NewInt(0).Sub(eth1Block.Number, big.NewInt(BlockStartOffset))
    targetBlock, err := t.ec.BlockByNumber(context.Background(), targetBlockNumber)
    if err != nil {
        return nil, err
    }

    return targetBlock.Number(), nil
}


// Get all of the pubkeys for the provided minipools
func (t *submitScrubMinipools) getPubkeys(minipoolAddresses []common.Address) ([]types.ValidatorPubkey, map[common.Address]types.ValidatorPubkey, error) {
    pubkeys := make([]types.ValidatorPubkey, len(minipoolAddresses))
    pubkeyMap := map[common.Address]types.ValidatorPubkey{}

    for i, minipoolAddress := range minipoolAddresses {
        pubkey, err := minipool.GetMinipoolPubkey(t.rp, minipoolAddress, nil)
        if err != nil {
            return nil, nil, err
        }
        pubkeys[i] = pubkey
        pubkeyMap[minipoolAddress] = pubkey
    }

    return pubkeys, pubkeyMap, nil
}


// Get the official withdrawal addresses for a set of minipools according to the Beacon Chain
func (t *submitScrubMinipools) getCanonicalMinipoolWithdrawalAddresses(minipoolAddresses []common.Address, pubkeys []types.ValidatorPubkey) (map[common.Address]common.Hash, error) {

    // Get the status of the validators on the Beacon chain
    statuses, err := t.bc.GetValidatorStatuses(pubkeys, nil)
    if err != nil {
        return nil, err
    }

    // Get the withdrawal credentials on Beacon for each validator if they exist
    withdrawalCredsMap := map[common.Address]common.Hash{}
    for i, pubkey := range pubkeys {
        status := statuses[pubkey]
        if status.Exists {
            minipoolAddress := minipoolAddresses[i]
            withdrawalCredsMap[minipoolAddress] = status.WithdrawalCredentials
        }
    }
    
    return withdrawalCredsMap, nil
}


// Step 1: Verify the Beacon Chain credentials for a minipool if they're present
func (t *submitScrubMinipools) checkBeaconCredentials(expectedWithdrawalCredentials map[common.Address]common.Hash, beaconWithdrawalCredentials map[common.Address]common.Hash) ([]common.Address, []common.Address, int, int) {

    badOnBeaconCount := 0
    goodOnBeaconCount := 0
    minipoolsToVerifyPrestakeFor := []common.Address{}
    minipoolsToScrub := []common.Address{}

    for minipoolAddress, expectedCreds := range expectedWithdrawalCredentials {
        beaconCreds, exists := beaconWithdrawalCredentials[minipoolAddress]
        if exists {
            // This minipool's deposit has been seen on the Beacon Chain
            if beaconCreds != expectedCreds {
                t.log.Println("=== SCRUB DETECTED ON BEACON CHAIN ===")
                t.log.Printlnf("\tMinipool: %s", minipoolAddress.Hex())
                t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
                t.log.Printlnf("\tActual creds: %s", (beaconCreds).Hex())
                t.log.Println("======================================")
                minipoolsToScrub = append(minipoolsToScrub, minipoolAddress)
                badOnBeaconCount++
            } else {
                // This minipool's credentials match, so it's good to go.
                goodOnBeaconCount++
            }
        } else {
            // This isn't seen on the Beacon Chain yet so we have to look at the deposit contract
            minipoolsToVerifyPrestakeFor = append(minipoolsToVerifyPrestakeFor, minipoolAddress)
        }
    }

    return minipoolsToVerifyPrestakeFor, minipoolsToScrub, goodOnBeaconCount, badOnBeaconCount
    
}


// Step 2: Validate the MinipoolPrestaked event
func (t *submitScrubMinipools) checkPrestakeEvents(minipoolsToVerifyPrestakeFor []common.Address, intervalSize *big.Int, depositDomain []byte) ([]*minipool.Minipool, []*minipool.Minipool, int, int, error) {

    badPrestakeCount := 0
    goodPrestakeCount := 0
    minipoolsToCheckDepositsFor := []*minipool.Minipool{}
    minipoolsToScrub := []*minipool.Minipool{}

    for _, minipoolAddress := range minipoolsToVerifyPrestakeFor {
        // Create a minipool contract wrapper for the given address
        mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
        if err != nil {
            return []*minipool.Minipool{}, []*minipool.Minipool{}, 0, 0, err
        }

        // Get the MinipoolPrestaked event
        prestakeData, err := mp.GetPrestakeEvent(intervalSize, nil)
        if err != nil {
            return []*minipool.Minipool{}, []*minipool.Minipool{}, 0, 0, err
        }

        // Convert it into Prysm's deposit data struct
        depositData := new(ethpb.Deposit_Data)
        depositData.Amount = prestakeData.Amount.Uint64()
        depositData.PublicKey = prestakeData.Pubkey.Bytes()
        depositData.WithdrawalCredentials = prestakeData.WithdrawalCredentials.Bytes()
        depositData.Signature = prestakeData.Signature.Bytes()
        
        // Validate the signature
        err = prdeposit.VerifyDepositSignature(depositData, depositDomain)
        if err != nil {
            // The signature is illegal
            t.log.Println("=== SCRUB DETECTED ON PRESTAKE EVENT ===")
            t.log.Printlnf("Invalid prestake data for minipool %s:", minipoolAddress.Hex())
            t.log.Printlnf("\tError: %s", err.Error())
            t.log.Println("========================================")
            minipoolsToScrub = append(minipoolsToScrub, mp)
            badPrestakeCount++
        } else {
            // The signature is good
            minipoolsToCheckDepositsFor = append(minipoolsToCheckDepositsFor, mp)
            goodPrestakeCount++
        }
    }

    return minipoolsToCheckDepositsFor, minipoolsToScrub, goodPrestakeCount, badPrestakeCount, nil 

}


// Step 3: Verify minipools by their deposits
func (t *submitScrubMinipools) checkDeposits(depositMap map[*minipool.Minipool][]utils.DepositData, depositDomain []byte, expectedWithdrawalCredentials map[common.Address]common.Hash) ([]*minipool.Minipool, int, int, int) {

    unknownDeposits := 0
    goodOnDepositContract := 0
    badOnDepositContract := 0
    minipoolsToScrub := []*minipool.Minipool{}

    // Check each minipool's deposit data
    for minipool, deposits := range depositMap {

        if len(deposits) == 0 {
            unknownDeposits++
        } else {
            // Go through each deposit for this minipool and find the first one that's valid
            for depositIndex, deposit := range deposits {
                depositData := new(ethpb.Deposit_Data)
                depositData.Amount = deposit.Amount
                depositData.PublicKey = deposit.Pubkey.Bytes()
                depositData.WithdrawalCredentials = deposit.WithdrawalCredentials.Bytes()
                depositData.Signature = deposit.Signature.Bytes()
                
                err := prdeposit.VerifyDepositSignature(depositData, depositDomain)
                if err != nil {
                    t.log.Printlnf("Invalid deposit for minipool %s:", minipool.Address.Hex())
                    t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
                    t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
                    t.log.Printlnf("\tError: %s", err.Error())
                } else {
                    // This is a valid deposit
                    expectedCreds := expectedWithdrawalCredentials[minipool.Address]
                    if deposit.WithdrawalCredentials != expectedCreds {
                        t.log.Println("=== SCRUB DETECTED ON DEPOSIT CONTRACT ===")
                        t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
                        t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
                        t.log.Printlnf("\tMinipool: %s", minipool.Address.Hex())
                        t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
                        t.log.Printlnf("\tActual creds: %s", deposit.WithdrawalCredentials.Hex())
                        t.log.Println("==========================================")
                        minipoolsToScrub = append(minipoolsToScrub, minipool)
                        badOnDepositContract++
                    } else {
                        goodOnDepositContract++
                    }
                    break
                }
            }
        }
    }

    return minipoolsToScrub, goodOnDepositContract, badOnDepositContract, unknownDeposits
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

