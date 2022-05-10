package watchtower

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Number of slots to go back in time and scan for penalties if state is empty (400k is approx. 8 weeks)
const NewPenaltyScanBuffer = 400000

// Process withdrawals task
type processPenalties struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
	ec  rocketpool.ExecutionClient
	bc  beacon.Client
}

type state struct {
	LatestPenaltySlot uint64 `yaml:"latestPenaltySlot"`
}

// Create process penalties task
func newProcessPenalties(c *cli.Context, logger log.ColorLogger) (*processPenalties, error) {
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
	return &processPenalties{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		bc:  bc,
		rp:  rp,
	}, nil
}

func stateFileExists(path string) bool {
	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}

	return true
}

func (s *state) loadState(path string) (*state, error) {

	// Load file into memory
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	// Unmarshal into state object
	err = yaml.Unmarshal(yamlFile, s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (s *state) saveState(path string) error {
	// Marshal state object
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	// Write to disk
	watchtowerDir := filepath.Dir(path)
	err = os.MkdirAll(watchtowerDir, 0644)
	if err != nil {
		return fmt.Errorf("error creating watchtower directory: %w", err)
	}
	return ioutil.WriteFile(path, data, 0644)
}

// Process penalties
func (t *processPenalties) run() error {

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
	t.log.Println("Checking for illegal fee recipients...")

	// Get latest block
	head, _, err := t.bc.GetBeaconBlock("finalized")
	if err != nil {
		return fmt.Errorf("Error getting beacon block: %w", err)
	}

	currentSlot := head.Slot

	// Read state from file or create if this is the first run
	watchtowerStatePath := t.cfg.Smartnode.GetWatchtowerStatePath()
	var s state

	if stateFileExists(watchtowerStatePath) {
		_, err := s.loadState(watchtowerStatePath)
		if err != nil {
			return fmt.Errorf("Error loading watchtower state: %w", err)
		}
	} else {
		// No state file so start from NewPenaltyScanBuffer slots ago
		if currentSlot > NewPenaltyScanBuffer {
			s.LatestPenaltySlot = currentSlot - NewPenaltyScanBuffer
		}
	}

	if currentSlot <= s.LatestPenaltySlot {
		// Nothing to do
		return nil
	}

	// Loop over unprocessed slots
	slotsSinceUpdate := 0
	for i := s.LatestPenaltySlot; i < currentSlot; i++ {
		block, exists, err := t.bc.GetBeaconBlock(strconv.FormatUint(i, 10))
		if !exists {
			// Nothing to do if slot was missed
			slotsSinceUpdate++
			if slotsSinceUpdate > 10000 {
				t.log.Printlnf("\tAt block %d of %d...", block.Slot, currentSlot)
				slotsSinceUpdate = 0
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("Error getting beacon block: %w", err)
		}

		illegalFeeRecipientFound, err := t.processBlock(&block)
		if illegalFeeRecipientFound {
			s.LatestPenaltySlot = block.Slot
			saveErr := s.saveState(watchtowerStatePath)
			if saveErr != nil {
				return fmt.Errorf("Error saving watchtower state file: %w", saveErr)
			}
		}
		if err != nil {
			return err
		}

		slotsSinceUpdate++
		if slotsSinceUpdate > 10000 {
			t.log.Printlnf("\tAt block %d of %d...", block.Slot, currentSlot)
			slotsSinceUpdate = 0
		}
	}
	_, err = t.processBlock(&head)
	if err != nil {
		return err
	}

	// Update latest slot in state
	s.LatestPenaltySlot = currentSlot

	err = s.saveState(watchtowerStatePath)
	if err != nil {
		return fmt.Errorf("Error saving watchtower state file: %w", err)
	}

	// Return
	return nil

}

func (t *processPenalties) processBlock(block *beacon.BeaconBlock) (bool, error) {
	illegalFeeRecipient := false

	if !block.HasExecutionPayload {
		// Merge hasn't occurred yet so skip
		return illegalFeeRecipient, nil
	}

	status, err := t.bc.GetValidatorStatusByIndex(strconv.FormatUint(block.ProposerIndex, 10), nil)
	if err != nil {
		return illegalFeeRecipient, err
	}

	// Get the minipool address from the proposer's pubkey
	minipoolAddress, err := minipool.GetMinipoolByPubkey(t.rp, status.Pubkey, nil)
	if err != nil {
		return illegalFeeRecipient, err
	}

	// A zero result indicates this proposer is not a RocketPool node operator
	var emptyAddress [20]byte
	if bytes.Equal(emptyAddress[:], minipoolAddress[:]) {
		return illegalFeeRecipient, nil
	}

	// Retrieve the node's distributor address
	mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
	if err != nil {
		return illegalFeeRecipient, err
	}

	nodeAddress, err := mp.GetNodeAddress(nil)
	if err != nil {
		return illegalFeeRecipient, err
	}

	distributorAddress, err := node.GetDistributorAddress(t.rp, nodeAddress, nil)
	if err != nil {
		return illegalFeeRecipient, err
	}

	// Retrieve the rETH address
	rethAddress := t.cfg.Smartnode.GetRethAddress()

	// Check whether the fee recipient is set correctly
	if block.FeeRecipient != distributorAddress && block.FeeRecipient != rethAddress {
		// Penalise for non-compliance
		illegalFeeRecipient = true
		t.log.Println("=== ILLEGAL FEE RECIPIENT DETECTED ===")
		t.log.Printlnf("Beacon Block:  %d", block.Slot)
		t.log.Printlnf("Minipool:      %s", minipoolAddress.Hex())
		t.log.Printlnf("Node:          %s", nodeAddress.Hex())
		t.log.Printlnf("Distributor:   %s", distributorAddress.Hex())
		t.log.Printlnf("rETH:          %s", rethAddress.Hex())
		t.log.Printlnf("FEE RECIPIENT: %s", block.FeeRecipient.Hex())
		t.log.Println("======================================")

		hash, err := network.SubmitPenalty(t.rp, minipoolAddress, block.Slot, nil)
		if err != nil {
			return illegalFeeRecipient, fmt.Errorf("Error submitting penalty against %s for block %n: %w", minipoolAddress.Hex(), block.Slot, err)
		}

		// Wait for the TX to successfully get mined
		_, err = utils.WaitForTransaction(t.ec, hash)
		if err != nil {
			return illegalFeeRecipient, err
		}

		// Log result
		t.log.Printlnf("Submitted penalty against %s with fee recipient %s on block %n with tx %s", minipoolAddress.Hex(), block.FeeRecipient.Hex(), block.Slot, hash.Hex())
	}

	return illegalFeeRecipient, nil
}
