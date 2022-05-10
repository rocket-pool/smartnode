package watchtower

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strconv"

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
	w   *wallet.Wallet
	rp  *rocketpool.RocketPool
	ec  *client.EthClientProxy
	bc  beacon.Client
}

type state struct {
	LatestPenaltySlot uint64 `yaml:"latestPenaltySlot"`
}

// Create process penalties task
func newProcessPenalties(c *cli.Context, logger log.ColorLogger) (*processPenalties, error) {
	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClientProxy(c)
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
		w:   w,
		ec:  ec,
		bc:  bc,
		rp:  rp,
	}, nil
}

func stateFileExists(watchtowerFolder string) bool {
	// Check if file exists
	_, err := os.Stat(path.Join(watchtowerFolder, "state.yml"))
	return err != nil || !errors.Is(err, os.ErrNotExist)
}

func (s *state) loadState(watchtowerFolder string) (*state, error) {

	// Load file into memory
	yamlFile, err := ioutil.ReadFile(path.Join(watchtowerFolder, "state.yml"))
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

func (s *state) saveState(watchtowerFolder string) error {
	// Marshal state object
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	// Write to disc
	return ioutil.WriteFile(path.Join(watchtowerFolder, "state.yml"), data, 0)
}

// Process withdrawals
func (t *processPenalties) run() error {

	// Get latest block
	head, _, err := t.bc.GetBeaconBlock("finalized")
	if err != nil {
		return fmt.Errorf("Error getting beacon block: %q", err)
	}

	currentSlot := head.Slot

	// Read state from file or create if this is the first run
	watchtowerFolder := t.c.GlobalString("watchtowerFolder")
	var s state

	if stateFileExists(watchtowerFolder) {
		_, err := s.loadState(watchtowerFolder)
		if err != nil {
			return fmt.Errorf("Error loading watchtower state: %q", err)
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
	for i := s.LatestPenaltySlot; i < currentSlot; i++ {
		block, exists, err := t.bc.GetBeaconBlock(strconv.FormatUint(i, 10))
		if !exists {
			// Nothing to do if slot was missed
			continue
		}
		if err != nil {
			return fmt.Errorf("Error getting beacon block: %q", err)
		}

		err = t.processBlock(&block)
		if err != nil {
			return err
		}
	}
	err = t.processBlock(&head)
	if err != nil {
		return err
	}

	// Update latest slot in state
	s.LatestPenaltySlot = currentSlot

	err = s.saveState(watchtowerFolder)
	if err != nil {
		return fmt.Errorf("Error saving watchtower state: %q", err)
	}

	// Return
	return nil

}

func (t *processPenalties) processBlock(block *beacon.BeaconBlock) error {
	if !block.HasExecutionPayload {
		// Merge hasn't occurred yet so skip
		return nil
	}

	status, err := t.bc.GetValidatorStatusByIndex(strconv.FormatUint(block.ProposerIndex, 10), nil)
	if err != nil {
		return err
	}

	// Get the minipool address from the proposer's pubkey
	minipoolAddress, err := minipool.GetMinipoolByPubkey(t.rp, status.Pubkey, nil)
	if err != nil {
		return err
	}

	// A zero result indicates this proposer is not a RocketPool node operator
	var emptyAddress [20]byte
	if bytes.Compare(emptyAddress[:], minipoolAddress[:]) == 0 {
		return nil
	}

	// Retrieve the node's distributor address
	mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
	if err != nil {
		return err
	}

	nodeAddress, err := mp.GetNodeAddress(nil)
	if err != nil {
		return err
	}

	distributorAddress, err := node.GetDistributorAddress(t.rp, nodeAddress, nil)
	if err != nil {
		return err
	}

	// Check whether the fee recipient is set correctly
	var expectedFeeRecipient common.Hash
	copy(expectedFeeRecipient[32-20:], distributorAddress[:])

	if bytes.Compare(expectedFeeRecipient[:], block.FeeRecipient[:]) != 0 {
		// Penalise for non-compliance
		hash, err := network.SubmitPenalty(t.rp, minipoolAddress, block.Slot, nil)
		if err != nil {
			return fmt.Errorf("Error submitting penalty against %s for block %n: %q", minipoolAddress.Hex(), block.Slot, err)
		}

		// Log result
		t.log.Printf("Submitted penalty against %s with fee recipient %s on block %n with tx %s\n", minipoolAddress.Hex(), block.FeeRecipient.Hex(), block.Slot, hash.Hex())
	}

	return nil
}
