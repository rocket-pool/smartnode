package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

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

// Settings
const MinipoolBalanceDetailsBatchSize = 8

// Submit network balances task
type submitNetworkBalances struct {
	c   *cli.Context
	log log.ColorLogger
	cfg *config.RocketPoolConfig
	w   *wallet.Wallet
	ec  rocketpool.ExecutionClient
	rp  *rocketpool.RocketPool
	bc  beacon.Client
	m   *state.NetworkStateManager
	s   *state.NetworkState
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
}
type minipoolBalanceDetails struct {
	IsStaking   bool
	UserBalance *big.Int
}

// Create submit network balances task
func newSubmitNetworkBalances(c *cli.Context, logger log.ColorLogger, m *state.NetworkStateManager) (*submitNetworkBalances, error) {

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
	return &submitNetworkBalances{
		c:   c,
		log: logger,
		cfg: cfg,
		w:   w,
		ec:  ec,
		rp:  rp,
		bc:  bc,
		m:   m,
	}, nil

}

// Submit network balances
func (t *submitNetworkBalances) run(isAtlasDeployed bool) error {

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Get the latest state
	t.s = t.m.GetLatestState()

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Check balance submission
	if !t.s.NetworkDetails.SubmitBalancesEnabled {
		return nil
	}

	// Log
	t.log.Println("Checking for network balance checkpoint...")

	// Get block to submit balances for
	blockNumberBig := t.s.NetworkDetails.LatestReportableBalancesBlock
	blockNumber := blockNumberBig.Uint64()

	// Check if a submission needs to be made
	if blockNumber <= t.s.NetworkDetails.BalancesBlock.Uint64() {
		return nil
	}

	// Get the time of the block
	header, err := t.ec.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return err
	}
	blockTime := time.Unix(int64(header.Time), 0)

	// Get the Beacon block corresponding to this time
	eth2Config := t.s.BeaconConfig
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := blockTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot
	requiredEpoch := slotNumber / eth2Config.SlotsPerEpoch

	// Check if the epoch in the snapshot is finalized yet
	stateEpoch := t.s.BeaconSlotNumber / t.s.BeaconConfig.SlotsPerEpoch
	if requiredEpoch > stateEpoch {
		t.log.Printlnf("Balances must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", blockNumber, requiredEpoch, stateEpoch)
		return nil
	}

	// Log
	t.log.Printlnf("Calculating network balances for block %d...", blockNumber)

	// Get network balances at block
	balances, err := t.getNetworkBalances(header, blockNumberBig, slotNumber, blockTime, isAtlasDeployed)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Deposit pool balance: %s wei", balances.DepositPool.String())
	t.log.Printlnf("Total minipool user balance: %s wei", balances.MinipoolsTotal.String())
	t.log.Printlnf("Staking minipool user balance: %s wei", balances.MinipoolsStaking.String())
	t.log.Printlnf("Fee distributor user balance: %s wei", balances.DistributorShareTotal.String())
	t.log.Printlnf("Smoothing pool user balance: %s wei", balances.SmoothingPoolShare.String())
	t.log.Printlnf("rETH contract balance: %s wei", balances.RETHContract.String())
	t.log.Printlnf("rETH token supply: %s wei", balances.RETHSupply.String())

	// Check if we have reported these specific values before
	hasSubmittedSpecific, err := t.hasSubmittedSpecificBlockBalances(nodeAccount.Address, blockNumber, balances)
	if err != nil {
		return err
	}
	if hasSubmittedSpecific {
		return nil
	}

	// We haven't submitted these values, check if we've submitted any for this block so we can log it
	hasSubmitted, err := t.hasSubmittedBlockBalances(nodeAccount.Address, blockNumber)
	if err != nil {
		return err
	}
	if hasSubmitted {
		t.log.Printlnf("Have previously submitted out-of-date balances for block %d, trying again...", blockNumber)
	}

	// Log
	t.log.Println("Submitting balances...")

	// Submit balances
	if err := t.submitBalances(balances); err != nil {
		return fmt.Errorf("Could not submit network balances: %w", err)
	}

	// Return
	return nil

}

// Check whether balances for a block has already been submitted by the node
func (t *submitNetworkBalances) hasSubmittedBlockBalances(nodeAddress common.Address, blockNumber uint64) (bool, error) {

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)
	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), nodeAddress.Bytes(), blockNumberBuf))

}

// Check whether specific balances for a block has already been submitted by the node
func (t *submitNetworkBalances) hasSubmittedSpecificBlockBalances(nodeAddress common.Address, blockNumber uint64, balances networkBalances) (bool, error) {

	// Calculate total ETH balance
	totalEth := big.NewInt(0)
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.RETHContract)
	totalEth.Add(totalEth, balances.DistributorShareTotal)
	totalEth.Add(totalEth, balances.SmoothingPoolShare)

	blockNumberBuf := make([]byte, 32)
	big.NewInt(int64(blockNumber)).FillBytes(blockNumberBuf)

	totalEthBuf := make([]byte, 32)
	totalEth.FillBytes(totalEthBuf)

	stakingBuf := make([]byte, 32)
	balances.MinipoolsStaking.FillBytes(stakingBuf)

	rethSupplyBuf := make([]byte, 32)
	balances.RETHSupply.FillBytes(rethSupplyBuf)

	return t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("network.balances.submitted.node"), nodeAddress.Bytes(), blockNumberBuf, totalEthBuf, stakingBuf, rethSupplyBuf))

}

// Prints a message to the log
func (t *submitNetworkBalances) printMessage(message string) {
	t.log.Println(message)
}

// Get the network balances at a specific block
func (t *submitNetworkBalances) getNetworkBalances(elBlockHeader *types.Header, elblock *big.Int, beaconBlock uint64, slotTime time.Time, isAtlasDeployed bool) (networkBalances, error) {

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, elblock)
	if err != nil {
		return networkBalances{}, err
	}

	// Data
	var wg errgroup.Group
	var depositPoolBalance *big.Int
	var mpBalanceDetails []minipoolBalanceDetails
	var distributorShares []*big.Int
	var smoothingPoolShare *big.Int
	rethContractBalance := t.s.NetworkDetails.RETHBalance
	rethTotalSupply := t.s.NetworkDetails.TotalRETHSupply

	// Get deposit pool balance
	if isAtlasDeployed {
		depositPoolBalance = t.s.NetworkDetails.DepositPoolUserBalance
	} else {
		depositPoolBalance = t.s.NetworkDetails.DepositPoolBalance
	}

	// Get minipool balance details
	wg.Go(func() error {
		mpBalanceDetails = make([]minipoolBalanceDetails, len(t.s.MinipoolDetails))
		blockEpoch := t.s.BeaconSlotNumber / t.s.BeaconConfig.SlotsPerEpoch
		for i, mpd := range t.s.MinipoolDetails {
			mpBalanceDetails[i] = t.getMinipoolBalanceDetails(&mpd, blockEpoch)
		}
		return nil
	})

	// Get distributor balance details
	wg.Go(func() error {
		distributorShares = make([]*big.Int, len(t.s.NodeDetails))
		for i, node := range t.s.NodeDetails {
			distributorShares[i] = node.DistributorBalanceUserETH
		}

		return nil
	})

	// Get the smoothing pool user share
	wg.Go(func() error {

		// Get the current interval
		currentIndex := t.s.NetworkDetails.RewardIndex

		// Get the start time for the current interval, and how long an interval is supposed to take
		startTime := t.s.NetworkDetails.IntervalStart
		intervalTime := t.s.NetworkDetails.IntervalDuration

		timeSinceStart := slotTime.Sub(startTime)
		intervalsPassed := timeSinceStart / intervalTime
		endTime := time.Now()

		// Approximate the staker's share of the smoothing pool balance
		treegen, err := rprewards.NewTreeGenerator(t.log, "[Balances]", client, t.cfg, t.bc, currentIndex, startTime, endTime, beaconBlock, elBlockHeader, uint64(intervalsPassed), t.s)
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
	}

	// Add minipool balances
	for _, mp := range mpBalanceDetails {
		balances.MinipoolsTotal.Add(balances.MinipoolsTotal, mp.UserBalance)
		if mp.IsStaking {
			balances.MinipoolsStaking.Add(balances.MinipoolsStaking, mp.UserBalance)
		}
	}

	// Add distributor shares
	for _, share := range distributorShares {
		balances.DistributorShareTotal.Add(balances.DistributorShareTotal, share)
	}

	// Return
	return balances, nil

}

// Get minipool balance details
func (t *submitNetworkBalances) getMinipoolBalanceDetails(mpd *rpstate.NativeMinipoolDetails, blockEpoch uint64) minipoolBalanceDetails {

	status := mpd.Status
	userDepositBalance := mpd.UserDepositBalance
	mpType := mpd.DepositType
	validator := t.s.ValidatorDetails[mpd.Pubkey]

	// Ignore vacant minipools
	if mpd.IsVacant {
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

	// Use user deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
		}
	}

	userBalance := mpd.UserShareOfBalanceIncludingBeacon
	// Return
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
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.RETHContract)
	totalEth.Add(totalEth, balances.DistributorShareTotal)
	totalEth.Add(totalEth, balances.SmoothingPoolShare)

	ratio := eth.WeiToEth(balances.RETHSupply) / eth.WeiToEth(totalEth)
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
	maxFee := eth.GweiToWei(WatchtowerMaxFee)
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(WatchtowerMaxPriorityFee)
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
