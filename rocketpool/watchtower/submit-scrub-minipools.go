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
	addresses, err := minipool.GetPrelaunchMinipoolAddresses(t.rp, nil)
	if err != nil {
		return err
	}
	if len(addresses) == 0 {
		return nil
	}

	// Get the starting eth1 block to search
	startBlock, err := t.getDepositContractStartBlock()
	if err != nil {
		return err
	}

	// Some counters
	totalMinipools := len(addresses)
	goodOnBeaconCount := 0
	badOnBeaconCount := 0
	goodOnDepositContract := 0
	badOnDepositContract := 0
	unknownDeposits := 0

	// A map of what the withdrawal credentials *should be*
	expectedWithdrawalCredentials := make(map[common.Address]common.Hash)

	// A list of minipools that haven't been registered on the Beacon Chain yet and
	// need to be validated via the deposit contract
	minipoolsToCheckViaDepositContract := []common.Address{}

	// A list of minipools that need to be scrubbed
	minipoolsToScrub := []common.Address{}

	for _, minipoolAddress := range addresses {
		// This is what the withdrawal credentials *should be*
		expectedCreds, err := minipool.GetMinipoolWithdrawalCredentials(t.rp, minipoolAddress, nil)
		if err != nil {
			return err
		}
		expectedWithdrawalCredentials[minipoolAddress] = expectedCreds

		// Query the Beacon Chain for the canonical credentials if present
		beaconCreds, err := t.getCanonicalMinipoolWithdrawalAddress(minipoolAddress)
		if err != nil {
			return err
		}

		if beaconCreds != nil {
			// This minipool's deposit has been seen on the Beacon Chain
			if *beaconCreds != expectedCreds {
				t.log.Println("=== SCRUB DETECTED ON BEACON CHAIN ===")
				t.log.Printlnf("\tMinipool: %s", minipoolAddress.Hex())
				t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
				t.log.Printlnf("\tActual creds: %s", (*beaconCreds).Hex())
				t.log.Println("======================================")
				minipoolsToScrub = append(minipoolsToScrub, minipoolAddress)
				badOnBeaconCount++
			} else {
				goodOnBeaconCount++
			}
		} else {
			minipoolsToCheckViaDepositContract = append(minipoolsToCheckViaDepositContract, minipoolAddress)
		}
	}

	// Get the map of minipools to deposit contract events
	eventLogInterval, err := api.GetEventLogInterval(t.cfg)
	if err != nil {
		return err
	}
	depositMap, err := rputils.GetMinipoolDeposits(t.rp, minipoolsToCheckViaDepositContract, 
		startBlock, eventLogInterval, nil)
	if err != nil {
		return err
	}

	// Put together the signature validation data
    eth2Config, err := t.bc.GetEth2Config()
    if err != nil {
        return err
    }
	domain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return err
	}

	// Check each minipool's deposit data
	for minipoolAddress, deposits := range depositMap {

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
				
				err = prdeposit.VerifyDepositSignature(depositData, domain)
				if err != nil {
					t.log.Printlnf("Invalid deposit for minipool %s:", minipoolAddress)
					t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
					t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
					t.log.Printlnf("\tError: %s", err.Error())
				} else {
					// This is a valid deposit
					expectedCreds := expectedWithdrawalCredentials[minipoolAddress]
					if deposit.WithdrawalCredentials != expectedCreds {
						t.log.Println("=== SCRUB DETECTED ON DEPOSIT CONTRACT ===")
						t.log.Printlnf("\tTX Hash: %s", deposit.TxHash.Hex())
						t.log.Printlnf("\tBlock: %d, TX Index: %d, Deposit Index: %d", deposit.BlockNumber, deposit.TxIndex, depositIndex)
						t.log.Printlnf("\tMinipool: %s", minipoolAddress.Hex())
						t.log.Printlnf("\tExpected creds: %s", expectedCreds.Hex())
						t.log.Printlnf("\tActual creds: %s", deposit.WithdrawalCredentials.Hex())
						t.log.Println("==========================================")
						minipoolsToScrub = append(minipoolsToScrub, minipoolAddress)
						badOnDepositContract++
					} else {
						goodOnDepositContract++
					}
					break
				}
			}
		}
	}

	// Scrub minipools
	for _, minipoolAddress := range minipoolsToScrub {
		mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
		if err != nil {
			return err
		}

		err = t.submitVoteScrubMinipool(mp)
		if err != nil {
			return err
		}
	}

	// Print final tally
	t.log.Println("Scrub check complete.")
	t.log.Printlnf("\tTotal prelaunch minipools: %d", totalMinipools)
	t.log.Printlnf("\tBeacon Chain scrubs: %d/%d", badOnBeaconCount, (badOnBeaconCount + goodOnBeaconCount))
	t.log.Printlnf("\tDeposit Contract scrubs: %d/%d", badOnDepositContract, (badOnDepositContract + goodOnDepositContract))
	t.log.Printlnf("\tPools without deposits: %d", unknownDeposits)

	// Return
	return nil

}


// Get the eth1 block that matches the eth1 block last seen in the latest finalized Beacon block, minus the start offset
func (t *submitScrubMinipools) getDepositContractStartBlock() (*big.Int, error) {

	data, err := t.bc.GetEth1DataForEth2Block("finalized")
	if err != nil {
		return nil, err
	}
	
	eth1Block, err := t.ec.BlockByHash(context.Background(), data.BlockHash)
	if err != nil {
		return nil, err
	}

	targetBlockNumber := big.NewInt(0).Sub(eth1Block.Number(), big.NewInt(BlockStartOffset))
	targetBlock, err := t.ec.BlockByNumber(context.Background(), targetBlockNumber)
	if err != nil {
		return nil, err
	}

	return targetBlock.Number(), nil
}


// Get the official withdrawal address for a minipool according to the Beacon Chain
func (t *submitScrubMinipools) getCanonicalMinipoolWithdrawalAddress(minipoolAddress common.Address) (*common.Hash, error) {

	// Get the validator pubkey
	pubkey, err := minipool.GetMinipoolPubkey(t.rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the status of the validator on the Beacon chain
	status, err := t.bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return nil, err
	}

	// Get the withdrawal credentials on Beacon if this validator exists
	if status.Exists {
		return &status.WithdrawalCredentials, nil
	}
	
	return nil, nil
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

