package watchtower

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils"
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
	addresses, err := minipool.GetPrelaunchMinipoolAddresses(t.rp, nil)
	if err != nil {
		return err
	}
	if len(addresses) == 0 {
		return nil
	}
	t.log.Printlnf("%d minipool(s) are in prelaunch status...", len(addresses))

	// Get the starting eth1 block to search
	startBlock, err := t.getDepositContractStartBlock()
	if err != nil {
		return err
	}

	// Submit vote to scrub minipools
	for _, minipoolAddress := range addresses {

		// This is what the withdrawal credentials *should be*
		expectedCreds := utils.GetWithdrawalCredentials(minipoolAddress)
		
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


// Get the eth1 block that matches the eth1 block last seen in the latest finalized Beacon block, minus the start offset
func (t *submitScrubMinipools) getDepositContractStartBlock() (uint64, error) {

	data, err := t.bc.GetEth1DataForEth2Block("finalized")
	if err != nil {
		return 0, err
	}
	
	eth1Block, err := t.ec.BlockByHash(context.Background(), data.BlockHash)
	if err != nil {
		return 0, err
	}

	targetBlockNumber := big.NewInt(0).Sub(eth1Block.Number(), big.NewInt(BlockStartOffset))
	targetBlock, err := t.ec.BlockByNumber(context.Background(), targetBlockNumber)
	if err != nil {
		return 0, err
	}

	return targetBlock.NumberU64(), nil
}


// Get the official withdrawal address for a minipool according to the Beacon Chain
func (t *submitScrubMinipools) getCanonicalMinipoolWithdrawalAddress(minipoolAddress common.Address, eth1StartBlock uint64) (common.Hash, error) {

	// Get the validator pubkey
	pubkey, err := minipool.GetMinipoolPubkey(t.rp, minipoolAddress, nil)
	if err != nil {
		return common.Hash{}, err
	}

	// Get the status of the validator on the Beacon chain
	status, err := t.bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return common.Hash{}, err
	}

	// Get the withdrawal credentials on Beacon if this validator exists
	if status.Exists {
		return status.WithdrawalCredentials, nil
	} else {
		// TODO: Walk the deposit contract and look for events
	}
	
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

