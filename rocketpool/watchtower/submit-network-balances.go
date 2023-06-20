package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"golang.org/x/sync/errgroup"
)

// Submit network balances task
type submitNetworkBalances struct {
	log       *log.ColorLogger
	errLog    *log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	logPrefix string
}

// Network balance info
type networkBalances struct {
	Block                 uint64
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
func newSubmitNetworkBalances(logger *log.ColorLogger, errorLogger *log.ColorLogger, cfg *config.RocketPoolConfig, w *wallet.Wallet, rp *rocketpool.RocketPool, bc beacon.Client) *submitNetworkBalances {
	return &submitNetworkBalances{
		log:       logger,
		errLog:    errorLogger,
		cfg:       cfg,
		w:         w,
		rp:        rp,
		bc:        bc,
		logPrefix: "[Network Balances]",
	}
}

// Submit network balances
func (t *submitNetworkBalances) run(state *state.NetworkState) error {
	balances, err := t.getNetworkBalances(state)
	if err != nil {
		return fmt.Errorf("%s %w", t.logPrefix, err)
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

	// Save this in case we end up needing specific balance submission info later
	/*
		// Check if we have reported these specific values before
		hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockBalances(nodeAccount.Address, blockNumber, balances)
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
		hasSubmitted, err := t.hasSubmittedBlockBalances(nodeAccount.Address, blockNumber)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", logPrefix, err))
			return
		}
		if hasSubmitted {
			t.log.Printlnf("Have previously submitted out-of-date balances for block %d, trying again...", blockNumber)
		}
	*/

	// Log
	t.log.Println("Submitting balances...")

	// Submit balances
	if err := t.submitBalances(balances); err != nil {
		return fmt.Errorf("%s could not submit network balances: %w", t.logPrefix, err)
	}

	return nil
}

// Get the network balances at a specific block
func (t *submitNetworkBalances) getNetworkBalances(state *state.NetworkState) (networkBalances, error) {

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

		// Get the time of the block
		elBlockHeader, err := t.rp.Client.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(state.ElBlockNumber))
		if err != nil {
			return err
		}
		slotTime := time.Unix(int64(elBlockHeader.Time), 0)

		// Get the current interval
		currentIndex := state.NetworkDetails.RewardIndex

		// Get the start time for the current interval, and how long an interval is supposed to take
		startTime := state.NetworkDetails.IntervalStart
		intervalTime := state.NetworkDetails.IntervalDuration

		timeSinceStart := slotTime.Sub(startTime)
		intervalsPassed := timeSinceStart / intervalTime
		endTime := slotTime

		// Approximate the staker's share of the smoothing pool balance
		treegen, err := rprewards.NewTreeGenerator(t.log, "[Balances]", t.rp, t.cfg, t.bc, currentIndex, startTime, endTime, state.BeaconSlotNumber, elBlockHeader, uint64(intervalsPassed), state, nil)
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
		Block:                 state.ElBlockNumber,
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
func (t *submitNetworkBalances) getMinipoolBalanceDetails(mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState, cfg *config.RocketPoolConfig) minipoolBalanceDetails {

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
	if status == rptypes.Dissolved {
		return minipoolBalanceDetails{
			UserBalance: big.NewInt(0),
		}
	}

	// Use user deposit balance if initialized or prelaunch
	if status == rptypes.Initialized || status == rptypes.Prelaunch {
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
func (t *submitNetworkBalances) submitBalances(balances networkBalances) error {

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

	// Get the gas limit
	gasInfo, err := network.EstimateSubmitBalancesGas(t.rp, balances.Block, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit network balances: %w", err)
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

	// Submit balances
	hash, err := network.SubmitBalances(t.rp, balances.Block, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts)
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
