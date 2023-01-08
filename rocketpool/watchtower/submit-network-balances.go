package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
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
	NodeAddress common.Address
	NodeFee     *big.Int
}

// Create submit network balances task
func newSubmitNetworkBalances(c *cli.Context, logger log.ColorLogger) (*submitNetworkBalances, error) {

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
	}, nil

}

// Submit network balances
func (t *submitNetworkBalances) run() error {

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

	// Data
	var wg errgroup.Group
	var nodeTrusted bool
	var submitBalancesEnabled bool

	// Get data
	wg.Go(func() error {
		var err error
		nodeTrusted, err = trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		submitBalancesEnabled, err = protocol.GetSubmitBalancesEnabled(t.rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	// Check node trusted status & settings
	if !(nodeTrusted && submitBalancesEnabled) {
		return nil
	}

	// Log
	t.log.Println("Checking for network balance checkpoint...")

	// Get block to submit balances for
	blockNumber, err := t.getLatestReportableBlock()
	if err != nil {
		return err
	}

	// Check if a submission needs to be made
	balancesBlock, err := network.GetBalancesBlock(t.rp, nil)
	if err != nil {
		return err
	}
	if blockNumber <= balancesBlock {
		return nil
	}

	// Get the time of the block
	header, err := t.ec.HeaderByNumber(context.Background(), big.NewInt(0).SetUint64(blockNumber))
	if err != nil {
		return err
	}
	blockTime := time.Unix(int64(header.Time), 0)

	// Get the Beacon block corresponding to this time
	eth2Config, err := t.bc.GetEth2Config()
	if err != nil {
		return err
	}
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
	timeSinceGenesis := blockTime.Sub(genesisTime)
	slotNumber := uint64(timeSinceGenesis.Seconds()) / eth2Config.SecondsPerSlot

	// Check if the epoch is finalized yet
	epoch := slotNumber / eth2Config.SlotsPerEpoch
	beaconHead, err := t.bc.GetBeaconHead()
	if err != nil {
		return err
	}
	finalizedEpoch := beaconHead.FinalizedEpoch
	if epoch > finalizedEpoch {
		t.log.Printlnf("Balances must be reported for EL block %d, waiting until Epoch %d is finalized (currently %d)", blockNumber, epoch, finalizedEpoch)
		return nil
	}

	// Log
	t.log.Printlnf("Calculating network balances for block %d...", blockNumber)

	// Get network balances at block
	balances, err := t.getNetworkBalances(header, slotNumber)
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
		t.log.Printlnf("Have previously submitted out-of-date balances for block $d, trying again...", blockNumber)
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

// Get the latest block number to report balances for
func (t *submitNetworkBalances) getLatestReportableBlock() (uint64, error) {

	// Require eth client synced
	if err := services.RequireEthClientSynced(t.c); err != nil {
		return 0, err
	}

	latestBlock, err := network.GetLatestReportableBalancesBlock(t.rp, nil)
	if err != nil {
		return 0, fmt.Errorf("Error getting latest reportable block: %w", err)
	}
	return latestBlock.Uint64(), nil

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
func (t *submitNetworkBalances) getNetworkBalances(elBlockHeader *types.Header, beaconBlock uint64) (networkBalances, error) {

	// Initialize call options
	opts := &bind.CallOpts{
		BlockNumber: elBlockHeader.Number,
	}

	// Get a client with the block number available
	client, err := eth1.GetBestApiClient(t.rp, t.cfg, t.printMessage, opts.BlockNumber)
	if err != nil {
		return networkBalances{}, err
	}

	// Data
	var wg errgroup.Group
	var depositPoolBalance *big.Int
	var minipoolBalanceDetails []minipoolBalanceDetails
	var distributorShares []*big.Int
	var smoothingPoolShare *big.Int
	var rethContractBalance *big.Int
	var rethTotalSupply *big.Int

	// Get deposit pool balance
	wg.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(client, opts)
		if err != nil {
			return fmt.Errorf("error getting deposit pool balance: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		// Get minipool balance details
		var err error
		minipoolBalanceDetails, err = t.getNetworkMinipoolBalanceDetails(client, opts)
		if err != nil {
			return fmt.Errorf("error getting minipool balance details: %w", err)
		}

		// Calculate average node fee
		minipoolFees := map[common.Address][]*big.Int{}
		for _, details := range minipoolBalanceDetails {
			fees, exists := minipoolFees[details.NodeAddress]
			if !exists {
				fees = []*big.Int{}
			}
			fees = append(fees, details.NodeFee)
			minipoolFees[details.NodeAddress] = fees
		}

		avgNodeFees := map[common.Address]*big.Int{}
		for nodeAddress, fees := range minipoolFees {
			feeCount := len(fees)
			if feeCount == 0 {
				// Shouldn't happen but just in case to prevent divide by zeros
				continue
			}

			// Get the average fee
			sum := big.NewInt(0)
			for _, fee := range fees {
				sum.Add(sum, fee)
			}
			sum.Div(sum, big.NewInt(int64(feeCount)))
			avgNodeFees[nodeAddress] = sum
		}

		// Get distributor balance details
		distributorShares, err = t.getFeeDistributorBalances(client, opts, avgNodeFees)
		if err != nil {
			return fmt.Errorf("error getting fee distributor balances: %w", err)
		}

		return nil
	})

	// Get the smoothing pool user share
	wg.Go(func() error {

		// Get the current interval
		currentIndexBig, err := rewards.GetRewardIndex(client, opts)
		if err != nil {
			return fmt.Errorf("error getting current reward index: %w", err)
		}
		currentIndex := currentIndexBig.Uint64()

		// Get the start time for the current interval, and how long an interval is supposed to take
		startTime, err := rewards.GetClaimIntervalTimeStart(client, opts)
		if err != nil {
			return fmt.Errorf("error getting claim interval start time: %w", err)
		}
		intervalTime, err := rewards.GetClaimIntervalTime(client, opts)
		if err != nil {
			return fmt.Errorf("error getting claim interval time: %w", err)
		}

		// Calculate the intervals passed
		blockHeader, err := client.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
		if err != nil {
			return fmt.Errorf("error getting latest block header: %w", err)
		}
		latestBlockTime := time.Unix(int64(blockHeader.Time), 0)
		timeSinceStart := latestBlockTime.Sub(startTime)
		intervalsPassed := timeSinceStart / intervalTime
		endTime := time.Now()

		// Approximate the staker's share of the smoothing pool balance
		treegen, err := rprewards.NewTreeGenerator(t.log, "[Balances]", client, t.cfg, t.bc, currentIndex, startTime, endTime, beaconBlock, elBlockHeader, uint64(intervalsPassed))
		if err != nil {
			return fmt.Errorf("error creating merkle tree generator to approximate share of smoothing pool: %w", err)
		}
		smoothingPoolShare, err = treegen.ApproximateStakerShareOfSmoothingPool()
		if err != nil {
			return fmt.Errorf("error getting approximate share of smoothing pool: %w", err)
		}

		return nil

	})

	// Get rETH contract balance
	wg.Go(func() error {
		rethContractAddress, err := client.GetAddress("rocketTokenRETH", opts)
		if err != nil {
			return fmt.Errorf("error getting rETH contract address: %w", err)
		}
		rethContractBalance, err = client.Client.BalanceAt(context.Background(), *rethContractAddress, opts.BlockNumber)
		if err != nil {
			return fmt.Errorf("error getting rETH contract balance: %w", err)
		}
		return nil
	})

	// Get rETH token supply
	wg.Go(func() error {
		var err error
		rethTotalSupply, err = tokens.GetRETHTotalSupply(client, opts)
		if err != nil {
			return fmt.Errorf("error getting total rETH supply: %w", err)
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
	for _, mp := range minipoolBalanceDetails {
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

// Get all minipool balance details
func (t *submitNetworkBalances) getNetworkMinipoolBalanceDetails(client *rocketpool.RocketPool, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var eth2Config beacon.Eth2Config
	var beaconHead beacon.BeaconHead
	var blockTime uint64

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetMinipoolAddresses(client, opts)
		if err != nil {
			return fmt.Errorf("error getting minipool addresses: %w", err)
		}
		return nil
	})

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = t.bc.GetEth2Config()
		if err != nil {
			return fmt.Errorf("error getting Beacon config: %w", err)
		}
		return nil
	})

	// Get beacon head
	wg1.Go(func() error {
		var err error
		beaconHead, err = t.bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("error getting Beacon head: %w", err)
		}
		return nil
	})

	// Get block time
	wg1.Go(func() error {
		header, err := client.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
		if err != nil {
			return fmt.Errorf("error getting block header for block %s: %w", opts.BlockNumber.String(), err)
		}
		blockTime = header.Time
		return nil
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return []minipoolBalanceDetails{}, err
	}

	// Get & check epoch at block
	blockEpoch := eth2.EpochAt(eth2Config, blockTime)
	if blockEpoch > beaconHead.Epoch {
		return []minipoolBalanceDetails{}, fmt.Errorf("Epoch %d at block %s is higher than current epoch %d", blockEpoch, opts.BlockNumber.String(), beaconHead.Epoch)
	}

	// Get minipool validator statuses
	validators, err := rp.GetMinipoolValidators(client, t.bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: &blockEpoch})
	if err != nil {
		return []minipoolBalanceDetails{}, fmt.Errorf("error getting minipool validators: %w", err)
	}

	// Load details in batches
	details := make([]minipoolBalanceDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolBalanceDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolBalanceDetailsBatchSize
		if mei > len(addresses) {
			mei = len(addresses)
		}

		// Log
		//t.log.Printlnf("Calculating balances for minipools %d - %d of %d...", msi + 1, mei, len(addresses))

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				validator := validators[address]
				mpDetails, err := t.getMinipoolBalanceDetails(client, address, opts, validator, eth2Config, blockEpoch)
				if err != nil {
					return fmt.Errorf("error getting balance details for minipool %s: %w", address.Hex(), err)
				}
				details[mi] = mpDetails
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return []minipoolBalanceDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get minipool balance details
func (t *submitNetworkBalances) getMinipoolBalanceDetails(client *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, blockEpoch uint64) (minipoolBalanceDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(client, minipoolAddress, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Data
	var wg errgroup.Group
	var status rptypes.MinipoolStatus
	var userDepositBalance *big.Int
	var mpType rptypes.MinipoolDeposit
	var nodeFee *big.Int
	var nodeAddress common.Address

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(opts)
		if err != nil {
			return fmt.Errorf("error getting minipool %s status: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg.Go(func() error {
		var err error
		userDepositBalance, err = mp.GetUserDepositBalance(opts)
		if err != nil {
			return fmt.Errorf("error getting user deposit balance for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg.Go(func() error {
		var err error
		mpType, err = mp.GetDepositType(opts)
		if err != nil {
			return fmt.Errorf("error getting user deposit type for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg.Go(func() error {
		var err error
		nodeFee, err = mp.GetNodeFeeRaw(opts)
		if err != nil {
			return fmt.Errorf("error getting node fee for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})
	wg.Go(func() error {
		var err error
		nodeAddress, err = mp.GetNodeAddress(opts)
		if err != nil {
			return fmt.Errorf("error getting node address for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Use user deposit balance if initialized or prelaunch
	if status == rptypes.Initialized || status == rptypes.Prelaunch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
			NodeAddress: nodeAddress,
			NodeFee:     nodeFee,
		}, nil
	}

	// Use user deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
			NodeAddress: nodeAddress,
			NodeFee:     nodeFee,
		}, nil
	}

	// Get user balance at block
	blockBalance := eth.GweiToWei(float64(validator.Balance))
	userBalance, err := mp.CalculateUserShare(blockBalance, opts)
	if err != nil {
		return minipoolBalanceDetails{}, fmt.Errorf("error calculating user share for minipool %s: %w", minipoolAddress.Hex(), err)
	}

	// Return
	if userDepositBalance.Cmp(big.NewInt(0)) == 0 && mpType == rptypes.Full {
		return minipoolBalanceDetails{
			IsStaking:   (validator.ExitEpoch > blockEpoch),
			UserBalance: big.NewInt(0).Sub(userBalance, eth.EthToWei(16)), // Remove 16 ETH from the user balance for full minipools in the refund queue
			NodeAddress: nodeAddress,
			NodeFee:     nodeFee,
		}, nil
	} else {
		return minipoolBalanceDetails{
			IsStaking:   (validator.ExitEpoch > blockEpoch),
			UserBalance: userBalance,
			NodeAddress: nodeAddress,
			NodeFee:     nodeFee,
		}, nil
	}

}

// Get the fee distributor balances
func (t *submitNetworkBalances) getFeeDistributorBalances(client *rocketpool.RocketPool, opts *bind.CallOpts, avgNodeFees map[common.Address]*big.Int) ([]*big.Int, error) {

	// Get all of the nodes
	nodeAddresses, err := node.GetNodeAddresses(client, opts)
	if err != nil {
		return []*big.Int{}, fmt.Errorf("error getting node addresses: %w", err)
	}

	// Load balances in batches
	balances := make([]*big.Int, len(nodeAddresses))
	for bsi := 0; bsi < len(nodeAddresses); bsi += MinipoolBalanceDetailsBatchSize {
		// Get batch start & end index
		nsi := bsi
		nei := bsi + MinipoolBalanceDetailsBatchSize
		if nei > len(nodeAddresses) {
			nei = len(nodeAddresses)
		}

		// Load details
		var wg errgroup.Group
		for ni := nsi; ni < nei; ni++ {
			ni := ni
			wg.Go(func() error {
				// Get the fee distributor's balance
				address := nodeAddresses[ni]
				distributor, err := node.GetDistributorAddress(client, address, opts)
				if err != nil {
					return fmt.Errorf("error getting distributor for node %s: %w", address.Hex(), err)
				}
				distributorBalance, err := client.Client.BalanceAt(context.Background(), distributor, opts.BlockNumber)
				if err != nil {
					return fmt.Errorf("error getting distributor balance for distributor %s, node %s: %w", distributor.Hex(), address.Hex(), err)
				}

				// Get the node's average fee
				// TODO: fix after update, manual calculation for now
				/*
					averageFee, err := node.GetNodeAverageFeeRaw(t.rp, address, opts)
					if err != nil {
						return fmt.Errorf("error getting average fee for node %s: %w", address.Hex(), err)
					}
				*/

				// Calculate the rETH share of the balance
				if distributorBalance.Cmp(big.NewInt(0)) > 0 {
					avgFee, exists := avgNodeFees[address]
					if !exists {
						// If a node doesn't have any minipools, there's no fee; it's split 50/50
						avgFee = eth.EthToWei(0.5)
					}

					// avgFee describes a node operator's average commission, so we need to take it out of the rEth holder's half
					one := big.NewInt(1e18)
					two := big.NewInt(2e18)
					avgFee.Sub(one, avgFee)                            // avgFee = 1 - avgFee
					distributorBalance.Mul(distributorBalance, avgFee) // balance *= avgFee
					distributorBalance.Div(distributorBalance, two)    // balance /= 2
				}

				balances[ni] = distributorBalance
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return []*big.Int{}, err
		}
	}

	return balances, nil

}

// Submit network balances
func (t *submitNetworkBalances) submitBalances(balances networkBalances) error {

	// Log
	t.log.Printlnf("Submitting network balances for block %d...", balances.Block)

	// Calculate total ETH balance
	totalEth := big.NewInt(0)
	totalEth.Add(totalEth, balances.DepositPool)
	totalEth.Add(totalEth, balances.MinipoolsTotal)
	totalEth.Add(totalEth, balances.RETHContract)
	totalEth.Add(totalEth, balances.DistributorShareTotal)
	totalEth.Add(totalEth, balances.SmoothingPoolShare)

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
