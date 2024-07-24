package watchtower

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	rpstate "github.com/rocket-pool/rocketpool-go/v2/utils/state"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/eth1"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	networkBalanceSubmissionKey string = "network.balances.submitted.node"
)

// Submit network balances task
type SubmitNetworkBalances struct {
	ctx              context.Context
	sp               services.ISmartNodeServiceProvider
	logger           *slog.Logger
	cfg              *config.SmartNodeConfig
	res              *config.MergedResources
	w                *wallet.Wallet
	ec               eth.IExecutionClient
	rp               *rocketpool.RocketPool
	bc               beacon.IBeaconClient
	eventLogInterval *big.Int
	lock             *sync.Mutex
	isRunning        bool
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
func NewSubmitNetworkBalances(ctx context.Context, sp services.ISmartNodeServiceProvider, logger *log.Logger) *SubmitNetworkBalances {
	lock := &sync.Mutex{}
	return &SubmitNetworkBalances{
		ctx:              ctx,
		sp:               sp,
		logger:           logger.With(slog.String(keys.TaskKey, "Balance Report")),
		cfg:              sp.GetConfig(),
		res:              sp.GetResources(),
		w:                sp.GetWallet(),
		ec:               sp.GetEthClient(),
		rp:               sp.GetRocketPool(),
		bc:               sp.GetBeaconClient(),
		eventLogInterval: big.NewInt(int64(config.EventLogInterval)),
		lock:             lock,
		isRunning:        false,
	}
}

// Submit network balances
func (t *SubmitNetworkBalances) Run(state *state.NetworkState) error {
	// Check if balance submission is enabled
	if !state.NetworkDetails.SubmitBalancesEnabled {
		return nil
	}

	// Log
	t.logger.Info("Starting network balance check.")

	// Check the last submission block
	lastSubmissionBlock := state.NetworkDetails.PricesBlock
	referenceTimestamp := t.cfg.PriceBalanceSubmissionReferenceTimestamp.Value

	// Get the duration in seconds for the interval between submissions
	submissionIntervalInSeconds := int64(state.NetworkDetails.PricesSubmissionFrequency)
	eth2Config := state.BeaconConfig

	slotNumber, nextSubmissionTime, targetBlockHeader, err := utils.FindNextSubmissionTarget(t.ctx, t.rp, eth2Config, t.bc, t.ec, lastSubmissionBlock, referenceTimestamp, submissionIntervalInSeconds)
	if err != nil {
		return err
	}

	targetBlockNumber := targetBlockHeader.Number.Uint64()
	if targetBlockNumber < lastSubmissionBlock {
		// No submission needed: target block older or equal to the last submission
		return nil
	}

	// Check if the process is already running
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Balance report is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		nodeAddress, _ := t.w.GetAddress()
		t.logger.Info("Starting balance report in a separate thread.")

		// Log
		t.logger.Info("Calculating network balances...", slog.Uint64(keys.BlockKey, targetBlockNumber))

		// Get network balances at block
		balances, err := t.getNetworkBalances(targetBlockHeader, targetBlockHeader.Number, slotNumber, time.Unix(int64(targetBlockHeader.Time), 0))
		if err != nil {
			t.handleError(err)
			return
		}

		// Log
		t.logger.Info(fmt.Sprintf("Deposit pool balance: %s wei", balances.DepositPool.String()))
		t.logger.Info(fmt.Sprintf("Node credit balance: %s wei", balances.NodeCreditBalance.String()))
		t.logger.Info(fmt.Sprintf("Total minipool user balance: %s wei", balances.MinipoolsTotal.String()))
		t.logger.Info(fmt.Sprintf("Staking minipool user balance: %s wei", balances.MinipoolsStaking.String()))
		t.logger.Info(fmt.Sprintf("Fee distributor user balance: %s wei", balances.DistributorShareTotal.String()))
		t.logger.Info(fmt.Sprintf("Smoothing pool user balance: %s wei", balances.SmoothingPoolShare.String()))
		t.logger.Info(fmt.Sprintf("rETH contract balance: %s wei", balances.RETHContract.String()))
		t.logger.Info(fmt.Sprintf("rETH token supply: %s wei", balances.RETHSupply.String()))

		// Check if we have reported these specific values before
		balances.SlotTimestamp = uint64(nextSubmissionTime.Unix())
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockBalances(nodeAddress, targetBlockNumber, balances)
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
		hasSubmitted, err := t.hasSubmittedBlockBalances(nodeAddress, targetBlockNumber)
		if err != nil {
			t.handleError(err)
			return
		}
		if hasSubmitted {
			t.logger.Info("Have previously submitted out-of-date balances, trying again...", slog.Uint64(keys.BlockKey, targetBlockNumber))
		}

		// Log
		t.logger.Info("Submitting balances...")

		// Set the reference timestamp
		balances.SlotTimestamp = uint64(nextSubmissionTime.Unix())

		// Submit balances
		if err := t.submitBalances(balances, true); err != nil {
			t.handleError(fmt.Errorf("Error submitting network balances: %w", err))
			return
		}

		// Log and return
		t.logger.Info("Balance report complete.")
		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

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

// Get the network balances at a specific block
func (t *SubmitNetworkBalances) getNetworkBalances(elBlockHeader *types.Header, elBlock *big.Int, beaconBlock uint64, slotTime time.Time) (networkBalances, error) {
	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.res, t.logger, elBlock)
	if err != nil {
		return networkBalances{}, err
	}

	// Create a new state gen manager
	mgr, err := state.NewNetworkStateManager(t.ctx, client, t.cfg, t.res, client.Client, t.bc, t.logger)
	if err != nil {
		return networkBalances{}, fmt.Errorf("error creating network state manager for EL block %s, Beacon slot %d: %w", elBlock, beaconBlock, err)
	}

	// Create a new state for the target block
	state, err := mgr.GetStateForSlot(t.ctx, beaconBlock)
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
		treegen, err := rprewards.NewTreeGenerator(t.logger, client, t.cfg, t.res, t.bc, currentIndex, startTime, endTime, beaconBlock, elBlockHeader, uint64(intervalsPassed), state, nil)
		if err != nil {
			return fmt.Errorf("error creating merkle tree generator to approximate share of smoothing pool: %w", err)
		}
		smoothingPoolShare, err = treegen.ApproximateStakerShareOfSmoothingPool(t.ctx)
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
func (t *SubmitNetworkBalances) getMinipoolBalanceDetails(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState, cfg *config.SmartNodeConfig) minipoolBalanceDetails {
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
	t.logger.Info("Calculated total ETH", slog.String(keys.AmountKey, totalEth.String()), slog.Float64(keys.RatioKey, ratio))

	// Log
	t.logger.Info("Submitting network balances...", slog.Uint64(keys.BlockKey, balances.Block))

	// Get transactor
	opts, err := t.w.GetTransactor()
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
			txInfo.SimulationResult = eth.SimulationResult{
				EstimatedGasLimit: utils.BalanceSubmissionForcedGas,
				SafeGasLimit:      utils.BalanceSubmissionForcedGas,
			}
			t.logger.Info("Network balance consensus has already been reached but submitting anyway for the health check.")
		} else {
			return fmt.Errorf("error getting TX for submitting network balances: %w", err)
		}
	}
	if txInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("simulating TX for submitting network balances failed: %s", txInfo.SimulationResult.SimulationError)
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
	err = tx.PrintAndWaitForTransaction(t.cfg, t.res, t.rp, t.logger, txInfo, opts)
	if err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}

	// Log
	t.logger.Info("Successfully submitted network balances.", slog.Uint64(keys.BlockKey, balances.Block))

	// Return
	return nil
}

func (t *SubmitNetworkBalances) handleError(err error) {
	t.logger.Error("*** Balance report failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
