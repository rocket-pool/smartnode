package watchtower

import (
	"fmt"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

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

func (s *state) loadState() (*state, error) {
	// Load file into memory
	yamlFile, err := ioutil.ReadFile("~/.rocketpool/watchtowerState/state.yml")
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

func (s *state) saveState() error {
	// Marshal state object
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	// Write to disc
	return ioutil.WriteFile("~/.rocketpool/watchtowerState/state.yml", data, 0)
}

// Process withdrawals
func (t *processPenalties) run() error {

	var s state
	_, err := s.loadState()
	if err != nil {
		return fmt.Errorf("Error loading watchtower state: %q", err)
	}

	// Get latest block
	head, err := t.bc.GetBeaconBlock("finalized")
	if err != nil {
		return fmt.Errorf("Error getting beacon block: %q", err)
	}

	currentSlot := head.Slot

	if currentSlot <= s.LatestPenaltySlot {
		// Nothing to do
		return nil
	}

	// Loop over unprocessed blocks
	for i := s.LatestPenaltySlot; i < currentSlot; i++ {
		block, err := t.bc.GetBeaconBlock(strconv.FormatUint(i, 10))
		if err != nil {
			return fmt.Errorf("Error getting beacon block: %q", err)
		}

		err = t.processSlot(&block)
		if err != nil {
			return err
		}
	}
	err = t.processSlot(&head)
	if err != nil {
		return err
	}

	err = s.saveState()
	if err != nil {
		return fmt.Errorf("Error saving watchtower state: %q", err)
	}

	// Return
	return nil

}

func (t *processPenalties) processSlot(block *beacon.BeaconBlock) error {

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
	if minipoolAddress.Hex() == "0x0000000000000000000000000000000000000000" {
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
	if block.FeeRecipient.Hex() != distributorAddress.Hex() {
		// Penalise for non-compliance
		hash, err := network.SubmitPenalty(t.rp, minipoolAddress, block.Slot, nil)
		if err != nil {
			return fmt.Errorf("Error submitting penalty against %s for slot %n: %q", minipoolAddress.Hex(), block.Slot, err)
		}

		// Log result
		t.log.Printf("Submitted penalty against %s with fee recipient %s on slot %n with tx %s\n", minipoolAddress.Hex(), block.FeeRecipient.Hex(), block.Slot, hash.Hex())
	}

	return nil
}
