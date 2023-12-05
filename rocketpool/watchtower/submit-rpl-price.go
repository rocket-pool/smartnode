package watchtower

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/contracts"
	"github.com/rocket-pool/smartnode/rocketpool/common/eth1"
	"github.com/rocket-pool/smartnode/rocketpool/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	mathutils "github.com/rocket-pool/smartnode/shared/utils/math"
)

// Settings
const (
	SubmissionKey string = "network.prices.submitted.node.key"
	BlocksPerTurn uint64 = 75 // Approx. 15 minutes

	twapNumberOfSeconds uint32 = 60 * 60 * 12 // 12 hours
)

type poolObserveResponse struct {
	TickCumulatives                    []*big.Int `abi:"tickCumulatives"`
	SecondsPerLiquidityCumulativeX128s []*big.Int `abi:"secondsPerLiquidityCumulativeX128s"`
}

// Submit RPL price task
type SubmitRplPrice struct {
	sp        *services.ServiceProvider
	log       log.ColorLogger
	errLog    log.ColorLogger
	cfg       *config.RocketPoolConfig
	ec        core.ExecutionClient
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	lock      *sync.Mutex
	isRunning bool
}

// Create submit RPL price task
func NewSubmitRplPrice(sp *services.ServiceProvider, logger log.ColorLogger, errorLogger log.ColorLogger) *SubmitRplPrice {
	lock := &sync.Mutex{}
	return &SubmitRplPrice{
		sp:     sp,
		log:    logger,
		errLog: errorLogger,
		lock:   lock,
	}
}

// Submit RPL price
func (t *SubmitRplPrice) Run(state *state.NetworkState) error {
	// Check if submission is enabled
	if !state.NetworkDetails.SubmitPricesEnabled {
		return nil
	}

	// Check if L2 rates are stale and update if necessary
	err := t.updateL2Prices(state)
	if err != nil {
		// Error is not fatal for this task so print and continue
		t.log.Printlnf("Error updating L2 prices: %s", err.Error())
	}

	// Make a new RP binding just for this portion
	rp := t.sp.GetRocketPool()

	// Check the last submission block
	lastSubmissionBlock := state.NetworkDetails.BalancesBlock.Uint64()
	networkMgr, err := network.NewNetworkManager(rp)
	if err != nil {
		return fmt.Errorf("error creating network manager binding: %w", err)
	}

	// Get the last prices updated event
	found, event, err := networkMgr.GetPriceUpdatedEvent(lastSubmissionBlock, nil)
	if err != nil {
		return fmt.Errorf("error getting event for price updated on block %d: %w", lastSubmissionBlock, err)
	}

	// Get the duration in seconds for the interval between submissions
	submissionIntervalDuration := time.Duration(state.NetworkDetails.PricesSubmissionFrequency * uint64(time.Second))
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
	t.log.Println("Checking for RPL price checkpoint...")

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
		t.log.Printlnf("Prices must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", blockNumber, requiredEpoch, finalizedEpoch)
		return nil
	}

	// Check if the process is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Prices report is already running in the background.")
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

		logPrefix := "[Price Report]"
		t.log.Printlnf("%s Starting price report in a separate thread.", logPrefix)

		// Log
		t.log.Printlnf("Getting RPL price for block %d...", blockNumber)

		// Get RPL price at block
		rplPrice, err := t.getRplTwap(blockNumber)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}

		// Log
		t.log.Printlnf("RPL price: %.6f ETH", mathutils.RoundDown(eth.WeiToEth(rplPrice), 6))

		// Check if we have reported these specific values before
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockPrices(nodeAddress, blockNumber, uint64(nextSubmissionTime.Unix()), rplPrice, true)
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
		hasSubmitted, err := t.hasSubmittedBlockPrices(nodeAddress, blockNumber, uint64(nextSubmissionTime.Unix()), true)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmitted {
			t.log.Printlnf("Have previously submitted out-of-date prices for block %d, trying again...", blockNumber)
		}

		// Log
		t.log.Println("Submitting RPL price...")

		// Submit RPL price
		if err := t.submitRplPrice(blockNumber, uint64(nextSubmissionTime.Unix()), rplPrice, true); err != nil {
			t.handleError(fmt.Errorf("%s could not submit RPL price: %w", logPrefix, err))
			return
		}

		// Log and return
		t.log.Printlnf("%s Price report complete.", logPrefix)
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil
}

func (t *SubmitRplPrice) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Price report failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Check whether prices for a block has already been submitted by the node
func (t *SubmitRplPrice) hasSubmittedBlockPrices(nodeAddress common.Address, blockNumber uint64, slotTimestamp uint64, isHoustonDeployed bool) (bool, error) {
	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	var result bool
	err := t.rp.Query(func(mc *batch.MultiCaller) error {
		t.rp.Storage.GetBool(mc, &result, crypto.Keccak256Hash([]byte(SubmissionKey), nodeAddress.Bytes(), blockNumberBuf))
		return nil
	}, nil)
	return result, err
}

// Check whether specific prices for a block has already been submitted by the node
func (t *SubmitRplPrice) hasSubmittedSpecificBlockPrices(nodeAddress common.Address, blockNumber uint64, slotTimestamp uint64, rplPrice *big.Int, isHoustonDeployed bool) (bool, error) {
	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)

	slotTimestampBuf := make([]byte, 32)
	big.NewInt(int64(slotTimestamp)).FillBytes(slotTimestampBuf)

	rplPriceBuf := make([]byte, 32)
	rplPrice.FillBytes(rplPriceBuf)

	var result bool
	err := t.rp.Query(func(mc *batch.MultiCaller) error {
		t.rp.Storage.GetBool(mc, &result, crypto.Keccak256Hash([]byte(SubmissionKey), nodeAddress.Bytes(), blockNumberBuf, slotTimestampBuf, rplPriceBuf))
		return nil
	}, nil)
	return result, err
}

// Get RPL price via TWAP at block
func (t *SubmitRplPrice) getRplTwap(blockNumber uint64) (*big.Int, error) {
	// Initialize call options
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(blockNumber)),
	}

	poolAddressStr := t.cfg.Smartnode.GetRplTwapPoolAddress()
	if poolAddressStr == "" {
		return nil, fmt.Errorf("RPL TWAP pool contract not deployed on this network")
	}
	poolAddress := common.HexToAddress(poolAddressStr)

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, opts.BlockNumber)
	if err != nil {
		return nil, err
	}
	twap, err := contracts.NewRplTwapPool(poolAddress, client.Client)
	if err != nil {
		return nil, fmt.Errorf("error creating TWAP pool binding: %w", err)
	}

	// Get RPL price
	response := contracts.PoolObserveResponse{}
	interval := twapNumberOfSeconds
	args := []uint32{interval, 0}

	err = client.Query(func(mc *batch.MultiCaller) error {
		twap.Observe(mc, &response, args)
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("could not get RPL price at block %d: %w", blockNumber, err)
	}
	if len(response.TickCumulatives) < 2 {
		return nil, fmt.Errorf("TWAP contract didn't have enough tick cumulatives for block %d (raw: %v)", blockNumber, response.TickCumulatives)
	}

	tick := big.NewInt(0).Sub(response.TickCumulatives[1], response.TickCumulatives[0])
	tick.Div(tick, big.NewInt(int64(interval))) // tick = (cumulative[1] - cumulative[0]) / interval

	base := eth.EthToWei(1.0001) // 1.0001e18
	one := eth.EthToWei(1)       // 1e18

	numerator := big.NewInt(0).Exp(base, tick, nil) // 1.0001e18 ^ tick
	numerator.Mul(numerator, one)

	denominator := big.NewInt(0).Exp(one, tick, nil) // 1e18 ^ tick
	denominator.Div(numerator, denominator)          // denominator = (1.0001e18^tick / 1e18^tick)

	numerator.Mul(one, one)                               // 1e18 ^ 2
	rplPrice := big.NewInt(0).Div(numerator, denominator) // 1e18 ^ 2 / (1.0001e18^tick * 1e18 / 1e18^tick)

	// Return
	return rplPrice, nil
}

func (t *SubmitRplPrice) printMessage(message string) {
	t.log.Println(message)
}

// Submit RPL price and total effective RPL stake
func (t *SubmitRplPrice) submitRplPrice(blockNumber uint64, slotTimestamp uint64, rplPrice *big.Int, isHoustonDeployed bool) error {

	// Log
	t.log.Printlnf("Submitting RPL price for block %d...", blockNumber)

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Create the network manager
	networkMgr, err := network.NewNetworkManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating network manager binding: %w", err)
	}

	// Get the tx info
	txInfo, err := networkMgr.SubmitPrices(blockNumber, slotTimestamp, rplPrice, opts)
	if err != nil {
		return fmt.Errorf("error getting the TX for submitting RPL price: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating SubmitPrices tx failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted RPL price for block %d.", blockNumber)

	// Return
	return nil
}

// Checks if L2 rates are stale and submits any that are
func (t *SubmitRplPrice) updateL2Prices(state *state.NetworkState) error {
	// Get services
	cfg := t.sp.GetConfig()
	rp := t.sp.GetRocketPool()
	ec := t.sp.GetEthClient()

	// Create bindings
	errs := []error{}
	optimismMessengerAddress := cfg.Smartnode.GetOptimismMessengerAddress()
	var optimismMessenger *contracts.OptimismMessenger
	if optimismMessengerAddress != "" {
		var err error
		optimismMessenger, err = contracts.NewOptimismMessenger(common.HexToAddress(optimismMessengerAddress), ec)
		errs = append(errs, err)
	}
	polygonMessengerAddress := cfg.Smartnode.GetPolygonMessengerAddress()
	var polygonMessenger *contracts.PolygonMessenger
	if polygonMessengerAddress != "" {
		var err error
		polygonMessenger, err = contracts.NewPolygonMessenger(common.HexToAddress(polygonMessengerAddress), ec)
		errs = append(errs, err)
	}
	arbitrumMessengerAddress := cfg.Smartnode.GetArbitrumMessengerAddress()
	var arbitrumMessenger *contracts.ArbitrumMessenger
	if arbitrumMessengerAddress != "" {
		var err error
		arbitrumMessenger, err = contracts.NewArbitrumMessenger(common.HexToAddress(arbitrumMessengerAddress), ec)
		errs = append(errs, err)
	}
	zksyncEraMessengerAddress := cfg.Smartnode.GetZkSyncEraMessengerAddress()
	var zkSyncEraMessenger *contracts.ZkSyncEraMessenger
	if zksyncEraMessengerAddress != "" {
		var err error
		zkSyncEraMessenger, err = contracts.NewZkSyncEraMessenger(common.HexToAddress(zksyncEraMessengerAddress), ec)
		errs = append(errs, err)
	}
	baseMessengerAddress := cfg.Smartnode.GetBaseMessengerAddress()
	var baseMessenger *contracts.OptimismMessenger // Base uses the same contract as Optimism
	if baseMessengerAddress != "" {
		var err error
		baseMessenger, err = contracts.NewOptimismMessenger(common.HexToAddress(baseMessengerAddress), ec)
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	// Exit if there aren't any L2 messengers
	if optimismMessenger == nil &&
		polygonMessenger == nil &&
		arbitrumMessenger == nil &&
		zkSyncEraMessenger == nil &&
		baseMessenger == nil {
		return nil
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return fmt.Errorf("Failed getting transactor: %q", err)
	}

	// Check if any rates are stale
	callOpts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}
	var optimismStale bool
	var polygonStale bool
	var arbitrumStale bool
	var zkSyncEraStale bool
	var baseStale bool
	err = rp.Query(func(mc *batch.MultiCaller) error {
		if optimismMessenger != nil {
			optimismMessenger.IsRateStale(mc, &optimismStale)
		}
		if polygonMessenger != nil {
			polygonMessenger.IsRateStale(mc, &polygonStale)
		}
		if arbitrumMessenger != nil {
			arbitrumMessenger.IsRateStale(mc, &arbitrumStale)
		}
		if zkSyncEraMessenger != nil {
			zkSyncEraMessenger.IsRateStale(mc, &zkSyncEraStale)
		}
		if baseMessenger != nil {
			baseMessenger.IsRateStale(mc, &baseStale)
		}
		return nil
	}, callOpts)
	if err != nil {
		return fmt.Errorf("error checking if rates are stale: %w", err)
	}
	if !(optimismStale || polygonStale || arbitrumStale || zkSyncEraStale || baseStale) {
		return nil
	}

	// Find out which oDAO index we are
	count := uint64(len(state.OracleDaoMemberDetails))
	var index = uint64(0)
	for i := uint64(0); i < count; i++ {
		if state.OracleDaoMemberDetails[i].Address == opts.From {
			index = i
			break
		}
	}

	// Submit if it's our turn
	blockNumber := state.ElBlockNumber
	indexToSubmit := (blockNumber / BlocksPerTurn) % count
	if index == indexToSubmit {
		errs := []error{}
		if optimismStale {
			errs = append(errs, t.updateOptimism(cfg, rp, optimismMessenger, blockNumber, opts))
		}
		if polygonStale {
			errs = append(errs, t.updatePolygon(cfg, rp, polygonMessenger, blockNumber, opts))
		}
		if arbitrumStale {
			errs = append(errs, t.updateArbitrum(cfg, rp, arbitrumMessenger, blockNumber, opts))
		}
		if zkSyncEraStale {
			errs = append(errs, t.updateZkSyncEra(cfg, rp, zkSyncEraMessenger, blockNumber, opts))
		}
		if baseStale {
			errs = append(errs, t.updateBase(cfg, rp, baseMessenger, blockNumber, opts))
		}
		return errors.Join(errs...)
	}
	return nil
}

// Submit a price update to Optimism
func (t *SubmitRplPrice) updateOptimism(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, optimismMessenger *contracts.OptimismMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.log.Println("Submitting rate to Optimism...")
	txInfo, err := optimismMessenger.SubmitRate(opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Optimism price update: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating tx for Optimism price update failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted Optimism price for block %d.", blockNumber)
	return nil
}

// Submit a price update to Polygon
func (t *SubmitRplPrice) updatePolygon(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, polygonmMessenger *contracts.PolygonMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.log.Println("Submitting rate to Polygon...")
	txInfo, err := polygonmMessenger.SubmitRate(opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Polygon price update: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating tx for Polygon price update failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted Polygon price for block %d.", blockNumber)
	return nil
}

// Submit a price update to Arbitrum
func (t *SubmitRplPrice) updateArbitrum(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, arbitrumMessenger *contracts.ArbitrumMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.log.Println("Submitting rate to Arbitrum...")

	// Get the current network recommended max fee
	suggestedMaxFee, err := gas.GetMaxFeeWeiForDaemon(&t.log)
	if err != nil {
		return fmt.Errorf("error getting recommended base fee from the network for Arbitrum price submission: %w", err)
	}

	// Constants for Arbitrum
	bufferMultiplier := big.NewInt(4)
	dataLength := big.NewInt(36)
	arbitrumGasLimit := big.NewInt(40000)
	arbitrumMaxFeePerGas := eth.GweiToWei(0.1)

	// Gas limit calculation on Arbitrum
	maxSubmissionCost := big.NewInt(6)
	maxSubmissionCost.Mul(maxSubmissionCost, dataLength)
	maxSubmissionCost.Add(maxSubmissionCost, big.NewInt(1400))
	maxSubmissionCost.Mul(maxSubmissionCost, suggestedMaxFee)  // (1400 + 6 * dataLength) * baseFee
	maxSubmissionCost.Mul(maxSubmissionCost, bufferMultiplier) // Multiply by the buffer constant for safety

	// Provide enough ETH for the L2 and roundtrip TX's
	value := big.NewInt(0)
	value.Mul(arbitrumGasLimit, arbitrumMaxFeePerGas)
	value.Add(value, maxSubmissionCost)
	opts.Value = value

	// Get the TX info
	txInfo, err := arbitrumMessenger.SubmitRate(maxSubmissionCost, arbitrumGasLimit, arbitrumMaxFeePerGas, opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Arbitrum price update: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating tx for Arbitrum price update failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted Arbitrum price for block %d.", blockNumber)
	return nil
}

// Submit a price update to zkSync Era
func (t *SubmitRplPrice) updateZkSyncEra(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, zkSyncEraMessenger *contracts.ZkSyncEraMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.log.Println("Submitting rate to zkSync Era...")
	// Constants for zkSync Era
	l1GasPerPubdataByte := big.NewInt(17)
	fairL2GasPrice := eth.GweiToWei(0.5)
	l2GasLimit := big.NewInt(750000)
	gasPerPubdataByte := big.NewInt(800)
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))

	// Value calculation on zkSync Era
	pubdataPrice := big.NewInt(0).Mul(l1GasPerPubdataByte, maxFee)
	minL2GasPrice := big.NewInt(0).Add(pubdataPrice, gasPerPubdataByte)
	minL2GasPrice.Sub(minL2GasPrice, big.NewInt(1))
	minL2GasPrice.Div(minL2GasPrice, gasPerPubdataByte)
	gasPrice := big.NewInt(0).Set(fairL2GasPrice)
	if minL2GasPrice.Cmp(gasPrice) > 0 {
		gasPrice.Set(minL2GasPrice)
	}
	txValue := big.NewInt(0).Mul(l2GasLimit, gasPrice)
	opts.Value = txValue

	// Get the TX info
	txInfo, err := zkSyncEraMessenger.SubmitRate(l2GasLimit, gasPerPubdataByte, opts)
	if err != nil {
		return fmt.Errorf("error getting tx for zkSync Era price update: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating tx for zkSync Era price update failed: %s", txInfo.SimError)
	}

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted zkSync Era price for block %d.", blockNumber)
	return nil
}

// Submit a price update to Base
func (t *SubmitRplPrice) updateBase(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, baseMessenger *contracts.OptimismMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.log.Println("Submitting rate to Base...")
	txInfo, err := baseMessenger.SubmitRate(opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Base price update: %w", err)
	}
	if txInfo.SimError != "" {
		return fmt.Errorf("simulating tx for Base price update failed: %s", txInfo.SimError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.GasInfo, false, 0, &t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.GasInfo.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, &t.log, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted Base price for block %d.", blockNumber)
	return nil
}
