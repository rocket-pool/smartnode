//lint:file-ignore U1000 Ignore all unused code, penalty system TBD

package watchtower

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Number of slots to go back in time and scan for penalties if state is empty (400k is approx. 8 weeks)
const NewPenaltyScanBuffer = 400000

// Process withdrawals task
type processPenalties struct {
	c              *cli.Context
	log            log.ColorLogger
	errLog         log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	ec             rocketpool.ExecutionClient
	bc             beacon.Client
	lock           *sync.Mutex
	isRunning      bool
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
	beaconConfig   beacon.Eth2Config
	m              *state.NetworkStateManager
	s              *state.NetworkState
}

type penaltyState struct {
	LatestPenaltySlot uint64 `yaml:"latestPenaltySlot"`
}

// Create process penalties task
func newProcessPenalties(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, m *state.NetworkStateManager) (*processPenalties, error) {
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

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Get the Beacon config
	beaconConfig, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Return task
	lock := &sync.Mutex{}
	return &processPenalties{
		c:              c,
		log:            logger,
		errLog:         errorLogger,
		cfg:            cfg,
		w:              w,
		ec:             ec,
		bc:             bc,
		rp:             rp,
		lock:           lock,
		isRunning:      false,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
		beaconConfig:   beaconConfig,
		m:              m,
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

func (s *penaltyState) loadState(path string) (*penaltyState, error) {

	// Load file into memory
	yamlFile, err := os.ReadFile(path)
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

func (s *penaltyState) saveState(path string) error {
	// Marshal state object
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	// Write to disk
	watchtowerDir := filepath.Dir(path)
	err = os.MkdirAll(watchtowerDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating watchtower directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
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

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Fee recipient check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		checkPrefix := "[Fee Recipients]"

		// Get the Smoothing Pool address
		smoothingPoolContract, err := t.rp.GetContract("rocketSmoothingPool", nil)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error getting smoothing pool contract: %w", checkPrefix, err))
			return
		}
		smoothingPoolAddress := *smoothingPoolContract.Address

		// Get latest block
		head, headExists, err := t.bc.GetBeaconBlock("finalized")
		if err != nil {
			t.handleError(fmt.Errorf("%s Error getting beacon block: %w", checkPrefix, err))
			return
		}

		currentSlot := head.Slot

		// Read state from file or create if this is the first run
		watchtowerStatePath := t.cfg.Smartnode.GetWatchtowerStatePath()
		var s penaltyState

		if stateFileExists(watchtowerStatePath) {
			_, err := s.loadState(watchtowerStatePath)
			if err != nil {
				t.handleError(fmt.Errorf("%s Error loading watchtower state: %w", checkPrefix, err))
				return
			}
		} else {
			// No state file so start from NewPenaltyScanBuffer slots ago
			if currentSlot > NewPenaltyScanBuffer {
				s.LatestPenaltySlot = currentSlot - NewPenaltyScanBuffer
			}
		}

		if currentSlot <= s.LatestPenaltySlot {
			// Nothing to do
			t.log.Printlnf("%s Finished checking for illegal fee recipients.", checkPrefix)
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		t.log.Printlnf("Starting check in a separate thread at block %d", s.LatestPenaltySlot)

		// Loop over unprocessed slots
		slotsSinceUpdate := 0
		for i := s.LatestPenaltySlot; i < currentSlot; i++ {
			block, exists, err := t.bc.GetBeaconBlock(strconv.FormatUint(i, 10))
			if err != nil {
				t.handleError(fmt.Errorf("%s Error getting beacon block: %w", checkPrefix, err))
				return
			}
			if exists {
				illegalFeeRecipientFound, err := t.processBlock(&block, smoothingPoolAddress)
				if illegalFeeRecipientFound {
					s.LatestPenaltySlot = block.Slot
					saveErr := s.saveState(watchtowerStatePath)
					if saveErr != nil {
						t.handleError(fmt.Errorf("%s Error saving watchtower state file: %w", checkPrefix, saveErr))
						return
					}
				}
				if err != nil {
					t.handleError(fmt.Errorf("%s %w", checkPrefix, err))
					return
				}
			}

			slotsSinceUpdate++
			if slotsSinceUpdate >= 10000 {
				t.log.Printlnf("\t%s At block %d of %d...", checkPrefix, i, currentSlot)
				slotsSinceUpdate = 0
				s.LatestPenaltySlot = block.Slot
				err = s.saveState(watchtowerStatePath)
				if err != nil {
					t.handleError(fmt.Errorf("%s Error saving watchtower state file: %w", checkPrefix, err))
					return
				}
			}
		}

		if headExists {
			_, err = t.processBlock(&head, smoothingPoolAddress)
			if err != nil {
				t.handleError(fmt.Errorf("%s %w", checkPrefix, err))
				return
			}
		}

		// Update latest slot in state
		s.LatestPenaltySlot = currentSlot
		err = s.saveState(watchtowerStatePath)
		if err != nil {
			t.handleError(fmt.Errorf("%s Error saving watchtower state file: %w", checkPrefix, err))
			return
		}

		t.log.Printlnf("%s Finished checking for illegal fee recipients.", checkPrefix)
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

}

func (t *processPenalties) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Illegal fee recipient check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

func (t *processPenalties) processBlock(block *beacon.BeaconBlock, smoothingPoolAddress common.Address) (bool, error) {

	isIllegalFeeRecipient := false

	if !block.HasExecutionPayload {
		// Merge hasn't occurred yet so skip
		return isIllegalFeeRecipient, nil
	}

	status, err := t.bc.GetValidatorStatusByIndex(block.ProposerIndex, nil)
	if err != nil {
		return isIllegalFeeRecipient, err
	}

	// Get the minipool address from the proposer's pubkey
	minipoolAddress, err := minipool.GetMinipoolByPubkey(t.rp, status.Pubkey, nil)
	if err != nil {
		return isIllegalFeeRecipient, err
	}

	// A zero result indicates this proposer is not a RocketPool node operator
	var emptyAddress [20]byte
	if bytes.Equal(emptyAddress[:], minipoolAddress[:]) {
		return isIllegalFeeRecipient, nil
	}

	// Retrieve the node's distributor address
	mp, err := minipool.NewMinipool(t.rp, minipoolAddress, nil)
	if err != nil {
		return isIllegalFeeRecipient, err
	}

	nodeAddress, err := mp.GetNodeAddress(nil)
	if err != nil {
		return isIllegalFeeRecipient, err
	}

	distributorAddress, err := node.GetDistributorAddress(t.rp, nodeAddress, nil)
	if err != nil {
		return isIllegalFeeRecipient, err
	}

	// Retrieve the rETH address
	rethAddress := t.cfg.Smartnode.GetRethAddress()

	// Ignore blocks that were sent to the smoothing pool
	if smoothingPoolAddress != emptyAddress && block.FeeRecipient == smoothingPoolAddress {
		return isIllegalFeeRecipient, nil
	}

	// Ignore blocks that were sent to the rETH address
	if block.FeeRecipient == rethAddress {
		return isIllegalFeeRecipient, nil
	}

	// Check if the user was opted into the smoothing pool for this block
	opts := bind.CallOpts{
		BlockNumber: big.NewInt(int64(block.ExecutionBlockNumber)),
	}
	isOptedIn, err := node.GetSmoothingPoolRegistrationState(t.rp, nodeAddress, &opts)
	if err != nil {
		t.log.Printlnf("*** WARNING: Couldn't check if node %s was opted into the smoothing pool for slot %d (execution block %d), skipping check... error: %w\n***", nodeAddress.Hex(), block.Slot, block.ExecutionBlockNumber, err)
		isOptedIn = false
	}

	// Check for smoothing pool theft
	if isOptedIn && block.FeeRecipient != smoothingPoolAddress {
		t.log.Println("=== SMOOTHING POOL THEFT DETECTED ===")
		t.log.Printlnf("Beacon Block:  %d", block.Slot)
		t.log.Printlnf("Minipool:      %s", minipoolAddress.Hex())
		t.log.Printlnf("Node:          %s", nodeAddress.Hex())
		t.log.Printlnf("FEE RECIPIENT: %s", block.FeeRecipient.Hex())
		t.log.Println("=====================================")

		isIllegalFeeRecipient = true
		err = t.submitPenalty(minipoolAddress, block)
		return isIllegalFeeRecipient, err
	}

	// Make sure they didn't opt out in order to steal a block
	if !isOptedIn {
		// Get the opt out time
		optOutTime, err := node.GetSmoothingPoolRegistrationChanged(t.rp, nodeAddress, &opts)
		if err != nil {
			t.log.Printlnf("*** WARNING: Couldn't check when node %s opted out of the smoothing pool for slot %d (execution block %d), skipping check... error: %w\n***", nodeAddress.Hex(), block.Slot, block.ExecutionBlockNumber, err)
		} else if optOutTime != time.Unix(0, 0) {
			// Get the time of the epoch before this one
			blockEpoch := block.Slot / t.beaconConfig.SlotsPerEpoch
			previousEpoch := blockEpoch - 1
			genesisTime := time.Unix(int64(t.beaconConfig.GenesisTime), 0)
			epochStartTime := genesisTime.Add(time.Second * time.Duration(t.beaconConfig.SecondsPerEpoch*previousEpoch))

			// If they opted out after the start of the previous epoch, they cheated
			if optOutTime.Sub(epochStartTime) > 0 {
				t.log.Println("=== SMOOTHING POOL THEFT DETECTED ===")
				t.log.Printlnf("Beacon Block:         %d", block.Slot)
				t.log.Printlnf("Safe Opt Out Time:    %s", epochStartTime)
				t.log.Printlnf("ACTUAL OPT OUT TIME:  %s", optOutTime)
				t.log.Printlnf("Minipool:             %s", minipoolAddress.Hex())
				t.log.Printlnf("Node:                 %s", nodeAddress.Hex())
				t.log.Printlnf("FEE RECIPIENT:        %s", block.FeeRecipient.Hex())
				t.log.Println("=====================================")

				isIllegalFeeRecipient = true
				err = t.submitPenalty(minipoolAddress, block)
				return isIllegalFeeRecipient, err
			}
		}
	}

	// Check for distributor address theft
	if !isOptedIn && block.FeeRecipient != distributorAddress {
		t.log.Println("=== ILLEGAL FEE RECIPIENT DETECTED ===")
		t.log.Printlnf("Beacon Block:  %d", block.Slot)
		t.log.Printlnf("Minipool:      %s", minipoolAddress.Hex())
		t.log.Printlnf("Node:          %s", nodeAddress.Hex())
		t.log.Printlnf("Distributor:   %s", distributorAddress.Hex())
		t.log.Printlnf("FEE RECIPIENT: %s", block.FeeRecipient.Hex())
		t.log.Println("======================================")

		isIllegalFeeRecipient = true
		err = t.submitPenalty(minipoolAddress, block)
		return isIllegalFeeRecipient, err
	}

	// No cheating detected
	return isIllegalFeeRecipient, nil

}

func (t *processPenalties) submitPenalty(minipoolAddress common.Address, block *beacon.BeaconBlock) error {

	// Check if this penalty has already been applied
	blockNumberBuf := make([]byte, 32)
	slotBig := big.NewInt(int64(block.Slot))
	slotBig.FillBytes(blockNumberBuf)
	penaltyExecuted, err := t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.penalties.executed"), minipoolAddress.Bytes(), blockNumberBuf))
	if err != nil {
		return fmt.Errorf("Could not check if penality has already been applied for block %d, minipool %s: %w", block.Slot, minipoolAddress.Hex(), err)
	}
	if penaltyExecuted {
		t.log.Printlnf("NOTE: Minipool %s was already penalized on block %d, skipping...", minipoolAddress.Hex(), block.Slot)
		return nil
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimateSubmitPenaltyGas(t.rp, minipoolAddress, slotBig, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit penalty: %w", err)
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
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gas.Uint64()

	hash, err := network.SubmitPenalty(t.rp, minipoolAddress, slotBig, opts)
	if err != nil {
		return fmt.Errorf("Error submitting penalty against %s for block %d: %w", minipoolAddress.Hex(), block.Slot, err)
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return err
	}

	// Log result
	t.log.Printlnf("Submitted penalty against %s with fee recipient %s on block %d with tx %s", minipoolAddress.Hex(), block.FeeRecipient.Hex(), block.Slot, hash.Hex())

	return nil

}
