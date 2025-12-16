package watchtower

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const (
	networkBalanceSubmissionKey string  = "network.balances.submitted.node"
	saturnBondInEth             float64 = 4
)

// Submit network balances task
type submitNetworkBalances struct {
	c         *cli.Context
	log       *log.ColorLogger
	errLog    *log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         wallet.Wallet
	ec        rocketpool.ExecutionClient
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	lock      *sync.Mutex
	isRunning bool
}

// Network balance info
type networkBalances struct {
	Block                   uint64
	SlotTimestamp           uint64
	DepositPool             *big.Int
	MinipoolsTotal          *big.Int
	MinipoolsStaking        *big.Int
	MegapoolStaking         *big.Int
	MegapoolsUserShareTotal *big.Int
	DistributorShareTotal   *big.Int
	SmoothingPoolShare      *big.Int
	RETHContract            *big.Int
	RETHSupply              *big.Int
	NodeCreditBalance       *big.Int
}
type validatorBalanceDetails struct {
	IsStaking   bool
	UserBalance *big.Int
}

type megapoolBalanceDetail struct {
	BeaconBalanceTotal *big.Int
	UserCapital        *big.Int
	ContractBalance    *big.Int
	RethRewards        *big.Int
	StakingBalance     *big.Int
}

// Create submit network balances task
func newSubmitNetworkBalances(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger) (*submitNetworkBalances, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetHdWallet(c)
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
	lock := &sync.Mutex{}
	return &submitNetworkBalances{
		c:         c,
		log:       &logger,
		errLog:    &errorLogger,
		cfg:       cfg,
		w:         w,
		ec:        ec,
		rp:        rp,
		bc:        bc,
		lock:      lock,
		isRunning: false,
	}, nil

}

// Submit network balances
func (t *submitNetworkBalances) run(state *state.NetworkState) error {

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

	// Check if balance submission is enabled
	if !state.NetworkDetails.SubmitBalancesEnabled {
		return nil
	}

	// Check the last submission block
	lastSubmissionBlock := state.NetworkDetails.BalancesBlock

	referenceTimestamp := t.cfg.Smartnode.PriceBalanceSubmissionReferenceTimestamp.Value.(int64)
	// Get the duration in seconds for the interval between submissions
	submissionIntervalInSeconds := int64(state.NetworkDetails.BalancesSubmissionFrequency)
	eth2Config := state.BeaconConfig

	// Log
	t.log.Println("Checking for network balance checkpoint...")
	slotNumber, nextSubmissionTime, targetBlockHeader, err := utils.FindNextSubmissionTarget(t.rp, eth2Config, t.bc, t.ec, lastSubmissionBlock, referenceTimestamp, submissionIntervalInSeconds)
	if err != nil {
		return err
	}
	targetBlockNumber := targetBlockHeader.Number.Uint64()

	if targetBlockNumber > state.ElBlockNumber || targetBlockNumber == lastSubmissionBlock {
		if targetBlockNumber > state.ElBlockNumber {
			// No submission needed: Target block in the future
			t.log.Println("not enough time has passed for the next price/balances submission")
			return nil
		}
		if targetBlockNumber == lastSubmissionBlock {
			// No submission needed: Already submitted for this block
			t.log.Println("balances have already been submitted for this block")
		}
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
		logPrefix := "[Balance Report]"
		t.log.Printlnf("%s Starting balance report in a separate thread.", logPrefix)

		// Log
		t.log.Printlnf("Calculating network balances for block %d...", targetBlockNumber)

		// Get network balances at block
		balances, err := t.getNetworkBalances(targetBlockHeader, big.NewInt(int64(targetBlockNumber)), slotNumber, time.Unix(int64(targetBlockHeader.Time), 0))
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
		t.log.Printlnf("Total megapool user balance: %s wei", balances.MegapoolsUserShareTotal.String())
		t.log.Printlnf("Staking megapool user balance: %s wei", balances.MegapoolStaking.String())
		t.log.Printlnf("Smoothing pool user balance: %s wei", balances.SmoothingPoolShare.String())
		t.log.Printlnf("rETH contract balance: %s wei", balances.RETHContract.String())
		t.log.Printlnf("rETH token supply: %s wei", balances.RETHSupply.String())

		// Check if we have reported these specific values before
		balances.SlotTimestamp = uint64(nextSubmissionTime.Unix())
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockBalances(nodeAccount.Address, targetBlockNumber, balances)
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
		hasSubmitted, err := t.hasSubmittedBlockBalances(nodeAccount.Address, targetBlockNumber)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmitted {
			t.log.Printlnf("Have previously submitted out-of-date balances for block %d, trying again...", targetBlockNumber)
		}

		// Log
		t.log.Println("Submitting balances...")

		// Set the reference timestamp
		balances.SlotTimestamp = uint64(nextSubmissionTime.Unix())

		// Submit balances
		if err := t.submitBalances(balances); err != nil {
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

func (t *submitNetworkBalances) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Balance report failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Check whether balances for a block has already been submitted by the node
func (t *submitNetworkBalances) hasSubmittedBlockBalances(nodeAddress common.Address, blockNumber uint64) (bool, error) {

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte(networkBalanceSubmissionKey), nodeAddress.Bytes(), blockNumberBuf))

}

// Check whether specific balances for a block has already been submitted by the node
func (t *submitNetworkBalances) hasSubmittedSpecificBlockBalances(nodeAddress common.Address, blockNumber uint64, balances networkBalances) (bool, error) {

	// Calculate total ETH balance
	totalEth := big.NewInt(0)
	totalEth.Sub(totalEth, balances.NodeCreditBalance)
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.MegapoolsUserShareTotal)
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

	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte(networkBalanceSubmissionKey), nodeAddress.Bytes(), blockNumberBuf, slotTimestampBuf, totalEthBuf, stakingBuf, rethSupplyBuf))

}

// Prints a message to the log
func (t *submitNetworkBalances) printMessage(message string) {
	t.log.Println(message)
}

// Get the network balances at a specific block
func (t *submitNetworkBalances) getNetworkBalances(elBlockHeader *types.Header, elBlock *big.Int, beaconBlock uint64, slotTime time.Time) (networkBalances, error) {

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, elBlock)
	if err != nil {
		return networkBalances{}, err
	}

	// Create a new state gen manager
	mgr := state.NewNetworkStateManager(client, t.cfg.Smartnode.GetStateManagerContracts(), t.bc, t.log)

	// Create a new state for the target block
	state, err := mgr.GetStateForSlot(beaconBlock)
	if err != nil {
		return networkBalances{}, fmt.Errorf("couldn't get network state for EL block %s, Beacon slot %d: %w", elBlock, beaconBlock, err)
	}

	// Data
	var wg errgroup.Group
	var depositPoolBalance *big.Int
	var mpBalanceDetails []validatorBalanceDetails
	var megapoolBalanceDetails []megapoolBalanceDetail
	var distributorShares []*big.Int
	var smoothingPoolShare *big.Int
	rethContractBalance := state.NetworkDetails.RETHBalance
	rethTotalSupply := state.NetworkDetails.TotalRETHSupply

	// Get deposit pool balance
	depositPoolBalance = state.NetworkDetails.DepositPoolUserBalance

	// Get minipool balance details
	wg.Go(func() error {
		mpBalanceDetails = make([]validatorBalanceDetails, len(state.MinipoolDetails))
		for i, mpd := range state.MinipoolDetails {
			mpBalanceDetails[i] = t.getMinipoolBalanceDetails(&mpd, state, t.cfg)
		}
		return nil
	})

	// Get megapool balance details
	wg.Go(func() error {
		megapoolBalanceDetails = make([]megapoolBalanceDetail, len(state.MegapoolDetails))
		i := 0
		for megapoolAddress, megapoolDetails := range state.MegapoolDetails {
			megapoolBalanceDetails[i], err = t.getMegapoolBalanceDetails(megapoolAddress, state, megapoolDetails)
			if err != nil {
				return fmt.Errorf("error getting megapool balance details: %w", err)
			}
			i += 1
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
		// Since we aren't generating an actual tree, just use beaconBlock as the snapshotEnd
		snapshotEnd := &rprewards.SnapshotEnd{
			Slot:           beaconBlock,
			ConsensusBlock: beaconBlock,
			ExecutionBlock: state.ElBlockNumber,
		}

		// Approximate the staker's share of the smoothing pool balance
		// NOTE: this will use the "vanilla" variant of treegen, without rolling records, to retain parity with other Oracle DAO nodes that aren't using rolling records
		treegen, err := rprewards.NewTreeGenerator(t.log, "[Balances]", rprewards.NewRewardsExecutionClient(client), t.cfg, t.bc, currentIndex, startTime, endTime, snapshotEnd, elBlockHeader, uint64(intervalsPassed), state)
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
		Block:                   elBlockHeader.Number.Uint64(),
		DepositPool:             depositPoolBalance,
		MinipoolsTotal:          big.NewInt(0),
		MinipoolsStaking:        big.NewInt(0),
		DistributorShareTotal:   big.NewInt(0),
		MegapoolsUserShareTotal: big.NewInt(0),
		MegapoolStaking:         big.NewInt(0),
		SmoothingPoolShare:      smoothingPoolShare,
		RETHContract:            rethContractBalance,
		RETHSupply:              rethTotalSupply,
		NodeCreditBalance:       big.NewInt(0),
	}

	// Add minipool balances
	for _, mp := range mpBalanceDetails {
		balances.MinipoolsTotal.Add(balances.MinipoolsTotal, mp.UserBalance)
		if mp.IsStaking {
			balances.MinipoolsStaking.Add(balances.MinipoolsStaking, mp.UserBalance)
		}
	}

	// Add megapool balances
	for _, mega := range megapoolBalanceDetails {
		balances.MegapoolsUserShareTotal.Add(balances.MegapoolsUserShareTotal, mega.UserCapital)
		balances.MegapoolsUserShareTotal.Add(balances.MegapoolsUserShareTotal, mega.RethRewards)
		balances.MegapoolStaking.Add(balances.MegapoolStaking, mega.StakingBalance)
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

func (t *submitNetworkBalances) getMegapoolBalanceDetails(megapoolAddress common.Address, state *state.NetworkState, megapoolDetails rpstate.NativeMegapoolDetails) (megapoolBalanceDetail, error) {
	megapoolBalanceDetails := megapoolBalanceDetail{}
	megapoolValidators := state.MegapoolToPubkeysMap[megapoolAddress]
	// iterate the megapoolValidators array
	megapoolBeaconBalanceTotal := big.NewInt(0)
	megapoolStakingBalance := big.NewInt(0)
	blockEpoch := state.BeaconSlotNumber / state.BeaconConfig.SlotsPerEpoch

	for _, megapoolValidatorKey := range megapoolValidators {
		// Grab the validator details from the pubkey
		megapoolValidatorDetails := state.MegapoolValidatorDetails[megapoolValidatorKey]
		if megapoolValidatorDetails.Exists {
			megapoolBeaconBalanceTotal.Add(megapoolBeaconBalanceTotal, eth.GweiToWei(float64(megapoolValidatorDetails.Balance)))
			if megapoolValidatorDetails.ActivationEpoch < blockEpoch && megapoolValidatorDetails.ExitEpoch > blockEpoch {
				megapoolStakingBalance.Add(megapoolStakingBalance, eth.GweiToWei(float64(megapoolValidatorDetails.Balance)))
				megapoolStakingBalance.Sub(megapoolStakingBalance, eth.EthToWei(saturnBondInEth))
			}
		}
	}
	megapoolBalanceDetails.BeaconBalanceTotal = megapoolBeaconBalanceTotal
	megapoolBalanceDetails.StakingBalance = megapoolStakingBalance
	megapoolBalanceDetails.UserCapital = megapoolDetails.UserCapital
	megapoolBalanceDetails.ContractBalance = megapoolDetails.EthBalance
	capitalTotal := megapoolDetails.UserCapital
	balanceTotal := megapoolBeaconBalanceTotal.Add(megapoolBeaconBalanceTotal, megapoolDetails.EthBalance)
	rewards := balanceTotal.Sub(balanceTotal, capitalTotal)
	// Load the megapool
	megapoolContract, err := megapool.NewMegaPoolV1(t.rp, megapoolAddress, nil)
	if err != nil {
		return megapoolBalanceDetail{}, fmt.Errorf("error loading megapool contract: %w", err)
	}
	rewardsSplit := megapool.RewardSplit{
		NodeRewards:  big.NewInt(0),
		VoterRewards: big.NewInt(0),
		RethRewards:  big.NewInt(0),
	}
	if rewards.Cmp(big.NewInt(0)) > 0 {
		rewardsSplit, err = megapoolContract.CalculateRewards(rewards, nil)

		if err != nil {
			return megapoolBalanceDetail{}, fmt.Errorf("error calculating rewards split: %w", err)
		}
	}
	megapoolBalanceDetails.RethRewards = rewardsSplit.RethRewards
	return megapoolBalanceDetails, nil
}

// Get minipool balance details
func (t *submitNetworkBalances) getMinipoolBalanceDetails(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState, cfg *config.RocketPoolConfig) validatorBalanceDetails {

	status := mpd.Status
	userDepositBalance := mpd.UserDepositBalance
	mpType := mpd.DepositType
	validator := state.MinipoolValidatorDetails[mpd.Pubkey]

	blockEpoch := state.BeaconSlotNumber / state.BeaconConfig.SlotsPerEpoch

	// Ignore vacant minipools
	if mpd.IsVacant {
		return validatorBalanceDetails{
			UserBalance: big.NewInt(0),
		}
	}

	// Dissolved minipools don't contribute to rETH
	if status == rptypes.Dissolved {
		return validatorBalanceDetails{
			UserBalance: big.NewInt(0),
		}
	}

	// Use user deposit balance if initialized or prelaunch
	if status == rptypes.Initialized || status == rptypes.Prelaunch {
		return validatorBalanceDetails{
			UserBalance: userDepositBalance,
		}
	}

	// "Broken" LEBs with the Redstone delegates report their total balance minus their node deposit balance
	if mpd.DepositType == rptypes.Variable && mpd.Version == 2 {
		brokenBalance := big.NewInt(0).Set(mpd.Balance)
		brokenBalance.Add(brokenBalance, eth.GweiToWei(float64(validator.Balance)))
		brokenBalance.Sub(brokenBalance, mpd.NodeRefundBalance)
		brokenBalance.Sub(brokenBalance, mpd.NodeDepositBalance)
		return validatorBalanceDetails{
			IsStaking:   (validator.Exists && validator.ActivationEpoch < blockEpoch && validator.ExitEpoch > blockEpoch),
			UserBalance: brokenBalance,
		}
	}

	// Use user deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return validatorBalanceDetails{
			UserBalance: userDepositBalance,
		}
	}

	// Here userBalance is CalculateUserShare(beaconBalance + minipoolBalance - refund)
	userBalance := mpd.UserShareOfBalanceIncludingBeacon
	if userDepositBalance.Cmp(big.NewInt(0)) == 0 && mpType == rptypes.Full {
		return validatorBalanceDetails{
			IsStaking:   (validator.ExitEpoch > blockEpoch),
			UserBalance: big.NewInt(0).Sub(userBalance, eth.EthToWei(16)), // Remove 16 ETH from the user balance for full minipools in the refund queue
		}
	} else {
		return validatorBalanceDetails{
			IsStaking:   (validator.ExitEpoch > blockEpoch),
			UserBalance: userBalance,
		}
	}

}

// Submit network balances
func (t *submitNetworkBalances) submitBalances(balances networkBalances) error {

	// Calculate total ETH balance
	totalEth := big.NewInt(0)
	totalEth.Sub(totalEth, balances.NodeCreditBalance)
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.MegapoolsUserShareTotal)
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
	totalStaking := balances.MinipoolsStaking.Add(balances.MinipoolsStaking, balances.MegapoolStaking)

	// Get the gas limit
	var gasInfo rocketpool.GasInfo
	gasInfo, err = network.EstimateSubmitBalancesGas(t.rp, balances.Block, balances.SlotTimestamp, totalEth, totalStaking, balances.RETHSupply, opts)

	if err != nil {
		if enableSubmissionAfterConsensus_Balances && strings.Contains(err.Error(), "Network balances for an equal or higher block are set") {
			// Set a gas limit which will intentionally be too low and revert
			gasInfo = rocketpool.GasInfo{
				EstGasLimit:  utils.BalanceSubmissionForcedGas,
				SafeGasLimit: utils.BalanceSubmissionForcedGas,
			}
			t.log.Println("Network balance consensus has already been reached but submitting anyway for the health check.")
		} else {
			return fmt.Errorf("Could not estimate the gas required to submit network balances: %w", err)
		}
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = gasInfo.SafeGasLimit
	var hash common.Hash
	// Submit balances
	hash, err = network.SubmitBalances(t.rp, balances.Block, balances.SlotTimestamp, totalEth, totalStaking, balances.RETHSupply, opts)
	if err != nil {
		return fmt.Errorf("error submitting balances: %w", err)
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return fmt.Errorf("error waiting for transaction: %w", err)
	}

	// Log
	t.log.Printlnf("Successfully submitted network balances for block %d.", balances.Block)

	// Return
	return nil

}
