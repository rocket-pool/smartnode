package watchtower

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/eth1"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

// Settings
const (
	SubmissionKey string = "network.prices.submitted.node.key"
	BlocksPerTurn uint64 = 75 // Approx. 15 minutes

	twapNumberOfSeconds uint32 = 60 * 60 * 12 // 12 hours
)

// Submit RPL price task
type SubmitRplPrice struct {
	ctx       context.Context
	sp        *services.ServiceProvider
	logger    *slog.Logger
	cfg       *config.SmartNodeConfig
	ec        eth.IExecutionClient
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	bc        beacon.IBeaconClient
	lock      *sync.Mutex
	isRunning bool
}

// Create submit RPL price task
func NewSubmitRplPrice(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger) *SubmitRplPrice {
	lock := &sync.Mutex{}
	return &SubmitRplPrice{
		ctx:    ctx,
		sp:     sp,
		logger: logger.With(slog.String(keys.RoutineKey, "Price Report")),
		cfg:    sp.GetConfig(),
		ec:     sp.GetEthClient(),
		w:      sp.GetWallet(),
		rp:     sp.GetRocketPool(),
		bc:     sp.GetBeaconClient(),
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
		t.logger.Error("Error updating L2 prices: %s", err.Error())
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
	t.logger.Info("Starting RPL price report check.")

	// Get the Beacon block corresponding to this time
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := nextSubmissionTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot

	// Search for the last existing EL block, going back up to 32 slots if the block is not found.
	ecBlock, err := utils.FindLastExistingELBlockFromSlot(t.ctx, t.bc, slotNumber)
	if err != nil {
		return err
	}

	// Fetch the target block
	targetBlockHeader, err := t.ec.HeaderByHash(context.Background(), ecBlock.BlockHash)
	if err != nil {
		return err
	}
	blockNumber := targetBlockHeader.Number.Uint64()

	// Check if the required epoch is finalized yet
	requiredEpoch := slotNumber / eth2Config.SlotsPerEpoch
	beaconHead, err := t.bc.GetBeaconHead(t.ctx)
	if err != nil {
		return err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch
	if requiredEpoch > finalizedEpoch {
		t.logger.Info("Prices must be reported, waiting until target Epoch is finalized.", slog.Uint64(keys.BlockKey, blockNumber), slog.Uint64(keys.TargetEpochKey, requiredEpoch), slog.Uint64(keys.FinalizedEpochKey, finalizedEpoch))
		return nil
	}

	// Check if the process is already running
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Prices report is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		nodeAddress, _ := t.w.GetAddress()
		t.logger.Info("Starting price report in a separate thread.")

		// Get RPL price at block
		rplPrice, err := t.getRplTwap(blockNumber)
		if err != nil {
			t.handleError(err)
			return
		}

		// Log
		t.logger.Info("Retrieved RPL price. %.6f ETH", slog.Uint64(keys.BlockKey, blockNumber), slog.Float64(keys.PriceKey, math.RoundDown(eth.WeiToEth(rplPrice), 6)))

		// Check if we have reported these specific values before
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockPrices(nodeAddress, blockNumber, uint64(nextSubmissionTime.Unix()), rplPrice, true)
		if err != nil {
			t.handleError(err)
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
			t.handleError(err)
			return
		}
		if hasSubmitted {
			t.logger.Info("Have previously submitted out-of-date prices, trying again...", slog.Uint64(keys.BlockKey, blockNumber))
		}

		// Log
		t.logger.Info("Submitting RPL price...")

		// Submit RPL price
		if err := t.submitRplPrice(blockNumber, uint64(nextSubmissionTime.Unix()), rplPrice, true); err != nil {
			t.handleError(fmt.Errorf("error submitting RPL price: %w", err))
			return
		}

		// Log and return
		t.logger.Info("Price report complete.")
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil
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

	rs := t.cfg.GetRocketPoolResources()
	poolAddress := rs.RplTwapPoolAddress
	if poolAddress == nil {
		return nil, fmt.Errorf("RPL TWAP pool contract not deployed on this network")
	}

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.logger, opts.BlockNumber)
	if err != nil {
		return nil, err
	}
	twap, err := contracts.NewRplTwapPool(*poolAddress, client.Client)
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

// Submit RPL price and total effective RPL stake
func (t *SubmitRplPrice) submitRplPrice(blockNumber uint64, slotTimestamp uint64, rplPrice *big.Int, isHoustonDeployed bool) error {
	// Log
	t.logger.Info("Submitting RPL price...", slog.Uint64(keys.BlockKey, blockNumber))

	// Get transactor
	opts, err := t.w.GetTransactor()
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
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating SubmitPrices tx failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted RPL price.", slog.Uint64(keys.BlockKey, blockNumber))

	// Return
	return nil
}

// Checks if L2 rates are stale and submits any that are
func (t *SubmitRplPrice) updateL2Prices(state *state.NetworkState) error {
	// Get services
	cfg := t.sp.GetConfig()
	rp := t.sp.GetRocketPool()
	ec := t.sp.GetEthClient()
	txMgr := t.sp.GetTransactionManager()
	rs := cfg.GetRocketPoolResources()

	// Create bindings
	errs := []error{}
	optimismMessengerAddress := rs.OptimismPriceMessengerAddress
	var optimismMessenger *contracts.OptimismMessenger
	if optimismMessengerAddress != nil {
		var err error
		optimismMessenger, err = contracts.NewOptimismMessenger(*optimismMessengerAddress, ec, txMgr)
		errs = append(errs, err)
	}
	polygonMessengerAddress := rs.PolygonPriceMessengerAddress
	var polygonMessenger *contracts.PolygonMessenger
	if polygonMessengerAddress != nil {
		var err error
		polygonMessenger, err = contracts.NewPolygonMessenger(*polygonMessengerAddress, ec, txMgr)
		errs = append(errs, err)
	}
	arbitrumMessengerAddress := rs.ArbitrumPriceMessengerAddress
	var arbitrumMessenger *contracts.ArbitrumMessenger
	if arbitrumMessengerAddress != nil {
		var err error
		arbitrumMessenger, err = contracts.NewArbitrumMessenger(*arbitrumMessengerAddress, ec, txMgr)
		errs = append(errs, err)
	}
	arbitrumMessengerV2Address := rs.ArbitrumPriceMessengerAddressV2
	var arbitrumMessengerV2 *contracts.ArbitrumMessenger
	if arbitrumMessengerV2Address != nil {
		var err error
		arbitrumMessengerV2, err = contracts.NewArbitrumMessenger(*arbitrumMessengerV2Address, ec, txMgr)
		errs = append(errs, err)
	}
	zksyncEraMessengerAddress := rs.ZkSyncEraPriceMessengerAddress
	var zkSyncEraMessenger *contracts.ZkSyncEraMessenger
	if zksyncEraMessengerAddress != nil {
		var err error
		zkSyncEraMessenger, err = contracts.NewZkSyncEraMessenger(*zksyncEraMessengerAddress, ec, txMgr)
		errs = append(errs, err)
	}
	baseMessengerAddress := rs.BasePriceMessengerAddress
	var baseMessenger *contracts.OptimismMessenger // Base uses the same contract as Optimism
	if baseMessengerAddress != nil {
		var err error
		baseMessenger, err = contracts.NewOptimismMessenger(*baseMessengerAddress, ec, txMgr)
		errs = append(errs, err)
	}
	scrollMessengerAddress := rs.ScrollPriceMessengerAddress
	var scrollMessenger *contracts.ScrollMessenger
	var scrollEstimator *contracts.ScrollFeeEstimator
	if scrollMessengerAddress != nil {
		var err error
		scrollMessenger, err = contracts.NewScrollMessenger(*scrollMessengerAddress, ec, txMgr)
		errs = append(errs, err)

		scrollEstimator, err = contracts.NewScrollFeeEstimator(*rs.ScrollFeeEstimatorAddress, ec)
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}

	// Exit if there aren't any L2 messengers
	if optimismMessenger == nil &&
		polygonMessenger == nil &&
		arbitrumMessenger == nil &&
		arbitrumMessengerV2 == nil &&
		zkSyncEraMessenger == nil &&
		baseMessenger == nil &&
		scrollMessenger == nil {
		return nil
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		return fmt.Errorf("failed getting transactor: %q", err)
	}

	// Check if any rates are stale
	callOpts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}
	var optimismStale bool
	var polygonStale bool
	var arbitrumStale bool
	var arbitrumV2Stale bool
	var zkSyncEraStale bool
	var baseStale bool
	var scrollStale bool
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
		if arbitrumMessengerV2 != nil {
			arbitrumMessengerV2.IsRateStale(mc, &arbitrumV2Stale)
		}
		if zkSyncEraMessenger != nil {
			zkSyncEraMessenger.IsRateStale(mc, &zkSyncEraStale)
		}
		if baseMessenger != nil {
			baseMessenger.IsRateStale(mc, &baseStale)
		}
		if scrollMessenger != nil {
			scrollMessenger.IsRateStale(mc, &scrollStale)
		}
		return nil
	}, callOpts)
	if err != nil {
		return fmt.Errorf("error checking if rates are stale: %w", err)
	}
	if !(optimismStale || polygonStale || arbitrumStale || arbitrumV2Stale || zkSyncEraStale || baseStale || scrollStale) {
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
			errs = append(errs, t.updateArbitrum(cfg, rp, arbitrumMessenger, "V1", blockNumber, opts))
		}
		if arbitrumV2Stale {
			errs = append(errs, t.updateArbitrum(cfg, rp, arbitrumMessengerV2, "V2", blockNumber, opts))
		}
		if zkSyncEraStale {
			errs = append(errs, t.updateZkSyncEra(cfg, rp, zkSyncEraMessenger, blockNumber, opts))
		}
		if baseStale {
			errs = append(errs, t.updateBase(cfg, rp, baseMessenger, blockNumber, opts))
		}
		if scrollStale {
			errs = append(errs, t.updateScroll(cfg, rp, ec, scrollMessenger, scrollEstimator, blockNumber, opts))
		}
		return errors.Join(errs...)
	}
	return nil
}

// Submit a price update to Optimism
func (t *SubmitRplPrice) updateOptimism(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, optimismMessenger *contracts.OptimismMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.logger.Info("Submitting rate to Optimism...")
	txInfo, err := optimismMessenger.SubmitRate(opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Optimism price update: %w", err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating tx for Optimism price update failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted Optimism price.", slog.Uint64(keys.BlockKey, blockNumber))
	return nil
}

// Submit a price update to Polygon
func (t *SubmitRplPrice) updatePolygon(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, polygonmMessenger *contracts.PolygonMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.logger.Info("Submitting rate to Polygon...")
	txInfo, err := polygonmMessenger.SubmitRate(opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Polygon price update: %w", err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating tx for Polygon price update failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted Polygon price.", slog.Uint64(keys.BlockKey, blockNumber))
	return nil
}

// Submit a price update to Arbitrum
func (t *SubmitRplPrice) updateArbitrum(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, arbitrumMessenger *contracts.ArbitrumMessenger, version string, blockNumber uint64, opts *bind.TransactOpts) error {
	t.logger.Info("Submitting rate to Arbitrum...")

	// Get the current network recommended max fee
	suggestedMaxFee, err := gas.GetMaxFeeWeiForDaemon(t.logger)
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
		return fmt.Errorf("error getting tx for Arbitrum %s price update: %w", version, err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating tx for Arbitrum %s price update failed: %s", version, txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info(fmt.Sprintf("Successfully submitted Arbitrum %s price.", version), slog.Uint64(keys.BlockKey, blockNumber))
	return nil
}

// Submit a price update to zkSync Era
func (t *SubmitRplPrice) updateZkSyncEra(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, zkSyncEraMessenger *contracts.ZkSyncEraMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.logger.Info("Submitting rate to zkSync Era...")
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
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating tx for zkSync Era price update failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted zkSync Era price.", slog.Uint64(keys.BlockKey, blockNumber))
	return nil
}

// Submit a price update to Base
func (t *SubmitRplPrice) updateBase(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, baseMessenger *contracts.OptimismMessenger, blockNumber uint64, opts *bind.TransactOpts) error {
	t.logger.Info("Submitting rate to Base...")
	txInfo, err := baseMessenger.SubmitRate(opts)
	if err != nil {
		return fmt.Errorf("error getting tx for Base price update: %w", err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating tx for Base price update failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted Base price.", slog.Uint64(keys.BlockKey, blockNumber))
	return nil
}

// Submit a price update to Scroll
func (t *SubmitRplPrice) updateScroll(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, ec eth.IExecutionClient, scrollMessenger *contracts.ScrollMessenger, scrollEstimator *contracts.ScrollFeeEstimator, blockNumber uint64, opts *bind.TransactOpts) error {
	t.logger.Info("Submitting rate to Scroll...")
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(cfg))

	// Set a fixed gas limit a bit above the estimated 85,283
	l2GasLimit := big.NewInt(90000)

	// Query the L2 message fee
	var messageFee *big.Int
	err := rp.Query(func(mc *batch.MultiCaller) error {
		scrollEstimator.EstimateCrossDomainMessageFee(mc, &messageFee, l2GasLimit)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting cross domain message fee for Scroll: %w", err)
	}
	opts.Value = messageFee

	// Get the TX info
	txInfo, err := scrollMessenger.SubmitRate(l2GasLimit, opts)
	if err != nil {
		return fmt.Errorf("error getting tx for scroll price update: %w", err)
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating tx for scroll price update failed: %s", txInfo.SimulationResult.SimulationError)
	}

	// Print the gas info
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(cfg, rp, t.logger, txInfo, opts)
	if err != nil {
		return err
	}

	// Log
	t.logger.Info("Successfully submitted Scroll price.", slog.Uint64(keys.BlockKey, blockNumber))
	return nil
}

func (t *SubmitRplPrice) handleError(err error) {
	t.logger.Error("*** Price report failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
