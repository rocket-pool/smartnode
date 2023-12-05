package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/eth1"
	rprewards "github.com/rocket-pool/smartnode/rocketpool/common/rewards"
	"github.com/rocket-pool/smartnode/rocketpool/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const (
	networkBalanceSubmissionKey string = "network.balances.submitted.node"
)

// Submit network balances task
type SubmitNetworkBalances struct {
	sp        *services.ServiceProvider
	log       *log.ColorLogger
	errLog    *log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         *wallet.Wallet
	ec        core.ExecutionClient
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	lock      *sync.Mutex
	isRunning bool
}

// Network balance info
type networkBalances struct {
	Block                 uint64
	SlotTimestamp         uint64
	DepositPool           *big.Int
	MinipoolsTotal        *big.Int
	MinipoolsStaking      *big.Int
	DistributorShareTotal *big.Int
	SmoothingPoolShare    *big.Int
	RETHContract          *big.Int
	RETHSupply            *big.Int
	NodeCreditBalance     *big.Int
}
type minipoolBalanceDetails struct {
	IsStaking   bool
	UserBalance *big.Int
}

// Create submit network balances task
func NewSubmitNetworkBalances(sp *services.ServiceProvider, logger log.ColorLogger, errorLogger log.ColorLogger) *SubmitNetworkBalances {
	lock := &sync.Mutex{}
	return &SubmitNetworkBalances{
		sp:        sp,
		log:       &logger,
		errLog:    &errorLogger,
		lock:      lock,
		isRunning: false,
	}
}

// Submit network balances
func (t *SubmitNetworkBalances) Run(state *state.NetworkState) error {
	// Check if balance submission is enabled
	if !state.NetworkDetails.SubmitBalancesEnabled {
		return nil
	}

	// Make a new RP binding just for this portion
	rp := t.sp.GetRocketPool()

	// Check the last submission block
	lastSubmissionBlock := state.NetworkDetails.BalancesBlock.Uint64()
	networkMgr, err := network.NewNetworkManager(rp)
	if err != nil {
		return fmt.Errorf("error creating network manager binding: %w", err)
	}

	// Get the last balances updated event
	found, event, err := networkMgr.GetBalancesUpdatedEvent(lastSubmissionBlock, nil)
	if err != nil {
		return fmt.Errorf("error getting event for balances updated on block %d: %w", lastSubmissionBlock, err)
	}

	// Get the duration in seconds for the interval between submissions
	submissionIntervalDuration := time.Duration(state.NetworkDetails.BalancesSubmissionFrequency * uint64(time.Second))
	eth2Config := state.BeaconConfig

	var nextSubmissionTime time.Time
	if !found {
		// The first submission after Houston is deployed won't find an event emitted by this contract
		// The submission time will be adjusted to align with the reward time
		rewardsPool, err := rewards.NewRewardsPool(rp)
		if err != nil {
			return fmt.Errorf("error creating rewards pool binding: %w", err)
		}
		err = rp.Query(nil, nil,
			rewardsPool.IntervalStart,
			rewardsPool.IntervalDuration,
		)
		if err != nil {
			return fmt.Errorf("error getting rewards pool interval details: %w", err)
		}
		lastCheckpoint := rewardsPool.IntervalStart.Formatted()
		rewardsInterval := rewardsPool.IntervalDuration.Formatted()

		// Find the next checkpoint
		nextCheckpoint := lastCheckpoint.Add(rewardsInterval)

		// Calculate the number of submissions between now and the next checkpoint adding one so we have the first submission time that is in the past
		timeDifference := time.Until(nextCheckpoint)
		submissionsUntilNextCheckpoint := int(timeDifference/submissionIntervalDuration) + 1

		nextSubmissionTime = nextCheckpoint.Add(-time.Duration(submissionsUntilNextCheckpoint) * submissionIntervalDuration)
	} else {
		// Get the last submission reference time
		lastSubmissionTime := event.SlotTimestamp

		// Next submission adds the interval time to the last submission time
		nextSubmissionTime = lastSubmissionTime.Add(submissionIntervalDuration)
	}

	// Return if the time to submit has not arrived
	if time.Now().Before(nextSubmissionTime) {
		return nil
	}

	// Log
	t.log.Println("Checking for network balance checkpoint...")

	// Get the Beacon block corresponding to this time
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := nextSubmissionTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot
	ecBlock := beacon.Eth1Data{}

	// Search for the last existing EL block, going back one slot if the block is not found.
	for blockExists, searchSlot := false, slotNumber; !blockExists; searchSlot -= 1 {
		ecBlock, blockExists, err = t.bc.GetEth1DataForEth2Block(strconv.FormatUint(searchSlot, 10))
		if err != nil {
			return err
		}
		// If we go back more than 32 slots, error out
		if slotNumber-searchSlot > 32 {
			return fmt.Errorf("could not find EL block from slot %d", slotNumber)
		}
	}

	// Fetch the target block
	targetBlockHeader, err := t.ec.HeaderByHash(context.Background(), ecBlock.BlockHash)
	if err != nil {
		return err
	}
	blockNumber := targetBlockHeader.Number.Uint64()

	// Check if the required epoch is finalized yet
	requiredEpoch := slotNumber / eth2Config.SlotsPerEpoch
	beaconHead, err := t.bc.GetBeaconHead()
	if err != nil {
		return err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch
	if requiredEpoch > finalizedEpoch {
		t.log.Printlnf("Balances must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", blockNumber, requiredEpoch, finalizedEpoch)
		return nil
	}

	// Check if the process is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Balance report is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		// Get services
		t.cfg = t.sp.GetConfig()
		t.w = t.sp.GetWallet()
		t.rp = t.sp.GetRocketPool()
		t.ec = t.sp.GetEthClient()
		t.bc = t.sp.GetBeaconClient()
		nodeAddress, _ := t.w.GetAddress()

		logPrefix := "[Balance Report]"
		t.log.Printlnf("%s Starting balance report in a separate thread.", logPrefix)

		// Log
		t.log.Printlnf("Calculating network balances for block %d...", targetBlockHeader.Number)

		// Get network balances at block
		balances, err := t.getNetworkBalances(targetBlockHeader, targetBlockHeader.Number, slotNumber, time.Unix(int64(targetBlockHeader.Time), 0))
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}

		// Log
		t.log.Printlnf("Deposit pool balance: %s wei", balances.DepositPool.String())
		t.log.Printlnf("Node credit balance: %s wei", balances.NodeCreditBalance.String())
		t.log.Printlnf("Total minipool user balance: %s wei", balances.MinipoolsTotal.String())
		t.log.Printlnf("Staking minipool user balance: %s wei", balances.MinipoolsStaking.String())
		t.log.Printlnf("Fee distributor user balance: %s wei", balances.DistributorShareTotal.String())
		t.log.Printlnf("Smoothing pool user balance: %s wei", balances.SmoothingPoolShare.String())
		t.log.Printlnf("rETH contract balance: %s wei", balances.RETHContract.String())
		t.log.Printlnf("rETH token supply: %s wei", balances.RETHSupply.String())

		// Check if we have reported these specific values before
		balances.SlotTimestamp = uint64(nextSubmissionTime.Unix())
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockBalances(nodeAddress, blockNumber, balances)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmittedSpecific {
			t.lock.Lock()
			t.isRunning = false
			t.lock.Unlock()
			return
		}

		// We haven't submitted these values, check if we've submitted any for this block so we can log it
		hasSubmitted, err := t.hasSubmittedBlockBalances(nodeAddress, blockNumber)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmitted {
			t.log.Printlnf("Have previously submitted out-of-date balances for block %d, trying again...", blockNumber)
		}

		// Log
		t.log.Println("Submitting balances...")

		// Set the reference timestamp
		balances.SlotTimestamp = uint64(nextSubmissionTime.Unix())

		// Submit balances
		if err := t.submitBalances(balances, true); err != nil {
			t.handleError(fmt.Errorf("%s could not submit network balances: %w", logPrefix, err))
			return
		}

		// Log and return
		t.log.Printlnf("%s Balance report complete.", logPrefix)
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

}

func (t *SubmitNetworkBalances) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Balance report failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Check whether balances for a block has already been submitted by the node
func (t *SubmitNetworkBalances) hasSubmittedBlockBalances(nodeAddress common.Address, blockNumber uint64) (bool, error) {
	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	var result bool
	err := t.rp.Query(func(mc *batch.MultiCaller) error {
		t.rp.Storage.GetBool(mc, &result, crypto.Keccak256Hash([]byte(networkBalanceSubmissionKey), nodeAddress.Bytes(), blockNumberBuf))
		return nil
	}, nil)
	return result, err
}

// Check whether specific balances for a block has already been submitted by the node
func (t *SubmitNetworkBalances) hasSubmittedSpecificBlockBalances(nodeAddress common.Address, blockNumber uint64, balances networkBalances) (bool, error) {
	// Calculate total ETH balance
	totalEth := big.NewInt(0)
	totalEth.Sub(totalEth, balances.NodeCreditBalance)
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.RETHContract)
	totalEth.Add(totalEth, balances.DistributorShareTotal)
	totalEth.Add(totalEth, balances.SmoothingPoolShare)

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)

	slotTimestampBuf := make([]byte, 32)
	big.NewInt(int64(balances.SlotTimestamp)).FillBytes(slotTimestampBuf)

	totalEthBuf := make([]byte, 32)
	totalEth.FillBytes(totalEthBuf)

	stakingBuf := make([]byte, 32)
	balances.MinipoolsStaking.FillBytes(stakingBuf)

	rethSupplyBuf := make([]byte, 32)
	balances.RETHSupply.FillBytes(rethSupplyBuf)

	var result bool
	err := t.rp.Query(func(mc *batch.MultiCaller) error {
		t.rp.Storage.GetBool(mc, &result, crypto.Keccak256Hash([]byte(networkBalanceSubmissionKey), nodeAddress.Bytes(), blockNumberBuf, slotTimestampBuf, totalEthBuf, stakingBuf, rethSupplyBuf))
		return nil
	}, nil)
	return result, err
}

// Prints a message to the log
func (t *SubmitNetworkBalances) printMessage(message string) {
	t.log.Println(message)
}

// Get the network balances at a specific block
func (t *SubmitNetworkBalances) getNetworkBalances(elBlockHeader *types.Header, elBlock *big.Int, beaconBlock uint64, slotTime time.Time) (networkBalances, error) {
	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, elBlock)
	if err != nil {
		return networkBalances{}, err
	}

	// Create a new state gen manager
	mgr, err := state.NewNetworkStateManager(client, t.cfg, client.Client, t.bc, t.log)
	if err != nil {
		return networkBalances{}, fmt.Errorf("error creating network state manager for EL block %s, Beacon slot %d: %w", elBlock, beaconBlock, err)
	}

	// Create a new state for the target block
	state, err := mgr.GetStateForSlot(beaconBlock)
	if err != nil {
		return networkBalances{}, fmt.Errorf("couldn't get network state for EL block %s, Beacon slot %d: %w", elBlock, beaconBlock, err)
	}

	// Data
	var wg errgroup.Group
	var depositPoolBalance *big.Int
	var mpBalanceDetails []minipoolBalanceDetails
	var distributorShares []*big.Int
	var smoothingPoolShare *big.Int
	rethContractBalance := state.NetworkDetails.RETHBalance
	rethTotalSupply := state.NetworkDetails.TotalRETHSupply

	// Get deposit pool balance
	depositPoolBalance = state.NetworkDetails.DepositPoolUserBalance

	// Get minipool balance details
	wg.Go(func() error {
		mpBalanceDetails = make([]minipoolBalanceDetails, len(state.MinipoolDetails))
		for i, mpd := range state.MinipoolDetails {
			mpBalanceDetails[i] = t.getMinipoolBalanceDetails(&mpd, state, t.cfg)
		}
		return nil
	})

	// Get distributor balance details
	wg.Go(func() error {
		distributorShares = make([]*big.Int, len(state.NodeDetails))
		for i, node := range state.NodeDetails {
			distributorShares[i] = node.DistributorBalanceUserETH // Uses the go-lib based off-chain calculation method instead of the contract method
		}

		return nil
	})

	// Get the smoothing pool user share
	wg.Go(func() error {
		// Get the current interval
		currentIndex := state.NetworkDetails.RewardIndex

		// Get the start time for the current interval, and how long an interval is supposed to take
		startTime := state.NetworkDetails.IntervalStart
		intervalTime := state.NetworkDetails.IntervalDuration

		timeSinceStart := slotTime.Sub(startTime)
		intervalsPassed := timeSinceStart / intervalTime
		endTime := slotTime

		// Approximate the staker's share of the smoothing pool balance
		// NOTE: this will use the "vanilla" variant of treegen, without rolling records, to retain parity with other Oracle DAO nodes that aren't using rolling records
		treegen, err := rprewards.NewTreeGenerator(t.log, "[Balances]", client, t.cfg, t.bc, currentIndex, startTime, endTime, beaconBlock, elBlockHeader, uint64(intervalsPassed), state, nil)
		if err != nil {
			return fmt.Errorf("error creating merkle tree generator to approximate share of smoothing pool: %w", err)
		}
		smoothingPoolShare, err = treegen.ApproximateStakerShareOfSmoothingPool()
		if err != nil {
			return fmt.Errorf("error getting approximate share of smoothing pool: %w", err)
		}

		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return networkBalances{}, err
	}

	// Balances
	balances := networkBalances{
		Block:                 elBlockHeader.Number.Uint64(),
		DepositPool:           depositPoolBalance,
		MinipoolsTotal:        big.NewInt(0),
		MinipoolsStaking:      big.NewInt(0),
		DistributorShareTotal: big.NewInt(0),
		SmoothingPoolShare:    smoothingPoolShare,
		RETHContract:          rethContractBalance,
		RETHSupply:            rethTotalSupply,
		NodeCreditBalance:     big.NewInt(0),
	}

	// Add minipool balances
	for _, mp := range mpBalanceDetails {
		balances.MinipoolsTotal.Add(balances.MinipoolsTotal, mp.UserBalance)
		if mp.IsStaking {
			balances.MinipoolsStaking.Add(balances.MinipoolsStaking, mp.UserBalance)
		}
	}

	// Add node credits
	for _, node := range state.NodeDetails {
		balances.NodeCreditBalance.Add(balances.NodeCreditBalance, node.DepositCreditBalance)
	}

	// Add distributor shares
	for _, share := range distributorShares {
		balances.DistributorShareTotal.Add(balances.DistributorShareTotal, share)
	}

	// Return
	return balances, nil
}

// Get minipool balance details
func (t *SubmitNetworkBalances) getMinipoolBalanceDetails(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState, cfg *config.RocketPoolConfig) minipoolBalanceDetails {
	status := mpd.Status
	userDepositBalance := mpd.UserDepositBalance
	mpType := mpd.DepositType
	validator := state.ValidatorDetails[mpd.Pubkey]

	blockEpoch := state.BeaconSlotNumber / state.BeaconConfig.SlotsPerEpoch

	// Ignore vacant minipools
	if mpd.IsVacant {
		return minipoolBalanceDetails{
			UserBalance: big.NewInt(0),
		}
	}

	// Dissolved minipools don't contribute to rETH
	if status == rptypes.MinipoolStatus_Dissolved {
		return minipoolBalanceDetails{
			UserBalance: big.NewInt(0),
		}
	}

	// Use user deposit balance if initialized or prelaunch
	if status == rptypes.MinipoolStatus_Initialized || status == rptypes.MinipoolStatus_Prelaunch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
		}
	}

	// "Broken" LEBs with the Redstone delegates report their total balance minus their node deposit balance
	if mpd.DepositType == rptypes.Variable && mpd.Version == 2 {
		brokenBalance := big.NewInt(0).Set(mpd.Balance)
		brokenBalance.Add(brokenBalance, eth.GweiToWei(float64(validator.Balance)))
		brokenBalance.Sub(brokenBalance, mpd.NodeRefundBalance)
		brokenBalance.Sub(brokenBalance, mpd.NodeDepositBalance)
		return minipoolBalanceDetails{
			IsStaking:   (validator.Exists && validator.ActivationEpoch < blockEpoch && validator.ExitEpoch > blockEpoch),
			UserBalance: brokenBalance,
		}
	}

	// Use user deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
		}
	}

	// Here userBalance is CalculateUserShare(beaconBalance + minipoolBalance - refund)
	userBalance := mpd.UserShareOfBalanceIncludingBeacon
	if userDepositBalance.Cmp(big.NewInt(0)) == 0 && mpType == rptypes.Full {
		return minipoolBalanceDetails{
			IsStaking:   (validator.ExitEpoch > blockEpoch),
			UserBalance: big.NewInt(0).Sub(userBalance, eth.EthToWei(16)), // Remove 16 ETH from the user balance for full minipools in the refund queue
		}
	} else {
		return minipoolBalanceDetails{
			IsStaking:   (validator.ExitEpoch > blockEpoch),
			UserBalance: userBalance,
		}
	}
}

// Submit network balances
func (t *SubmitNetworkBalances) submitBalances(balances networkBalances, isHoustonDeployed bool) error {
	// Calculate total ETH balance
	totalEth := big.NewInt(0)
	totalEth.Sub(totalEth, balances.NodeCreditBalance)
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.RETHContract)
	totalEth.Add(totalEth, balances.DistributorShareTotal)
	totalEth.Add(totalEth, balances.SmoothingPoolShare)

	ratio := eth.WeiToEth(totalEth) / eth.WeiToEth(balances.RETHSupply)
	t.log.Printlnf("Total ETH = %s\n", totalEth)
	t.log.Printlnf("Calculated ratio = %.6f\n", ratio)

	// Log
	t.log.Printlnf("Submitting network balances for block %d...", balances.Block)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return fmt.Errorf("error getting node transactor: %w", err)
	}

	// Get the network manager
	networkMgr, err := network.NewNetworkManager(t.rp)
	if err != nil {
		return fmt.Errorf("error getting network manager: %w", err)
	}

	// Get the TX info
	txInfo, err := networkMgr.SubmitBalances(balances.Block, balances.SlotTimestamp, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts)
	if err != nil {
		if enableSubmissionAfterConsensus_Balances && strings.Contains(err.Error(), "Network balances for an equal or higher block are set") {
			// Set a gas limit which will intentionally be too low and revert
			txInfo.GasInfo = core.GasInfo{
				EstGasLimit:  utils.BalanceSubmissionForcedGas,
				SafeGasLimit: utils.BalanceSubmissionForcedGas,
			}
			t.log.Println("Network balance consensus has already been reached but submitting anyway for the health check.")
		} else {
			return fmt.Errorf("error getting TX for submitting network balances: %w", err)
		}
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating TX for submitting network balances failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.log, txInfo, opts)
	if err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}

	// Log
	t.log.Printlnf("Successfully submitted network balances for block %d.", balances.Block)

	// Return
	return nil

}
