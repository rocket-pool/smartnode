package watchtower

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/client"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Settings
const MinipoolBalanceDetailsBatchSize = 20
const SubmitFollowDistanceBalances = 2
const ConfirmDistanceBalances = 30

// Submit network balances task
type submitNetworkBalances struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            config.RocketPoolConfig
	w              *wallet.Wallet
	ec             *client.EthClientProxy
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Network balance info
type networkBalances struct {
	Block            uint64
	DepositPool      *big.Int
	MinipoolsTotal   *big.Int
	MinipoolsStaking *big.Int
	RETHContract     *big.Int
	RETHSupply       *big.Int
}
type minipoolBalanceDetails struct {
	IsStaking   bool
	UserBalance *big.Int
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

	// Get the user-requested max fee
	maxFee, err := cfg.GetMaxFee()
	if err != nil {
		return nil, fmt.Errorf("Error getting max fee in configuration: %w", err)
	}

	// Get the user-requested max fee
	maxPriorityFee, err := cfg.GetMaxPriorityFee()
	if err != nil {
		return nil, fmt.Errorf("Error getting max priority fee in configuration: %w", err)
	}
	if maxPriorityFee == nil || maxPriorityFee.Uint64() == 0 {
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		maxPriorityFee = big.NewInt(2)
	}

	// Get the user-requested gas limit
	gasLimit, err := cfg.GetGasLimit()
	if err != nil {
		return nil, fmt.Errorf("Error getting gas limit in configuration: %w", err)
	}

	// Return task
	return &submitNetworkBalances{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		ec:             ec,
		rp:             rp,
		bc:             bc,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
		gasLimit:       gasLimit,
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

	// Allow some blocks to pass in case of a short reorg
	currentBlockNumber, err := t.ec.BlockNumber(context.Background())
	if err != nil {
		return err
	}
	if blockNumber+SubmitFollowDistanceBalances > currentBlockNumber {
		return nil
	}

	// Check if a submission needs to be made
	balancesBlock, err := network.GetBalancesBlock(t.rp, nil)
	if err != nil {
		return err
	}
	if blockNumber <= balancesBlock {
		return nil
	}

	// If confirm distance has passed, we just want to ensure we have submitted and then early exit
	if blockNumber+ConfirmDistanceBalances <= currentBlockNumber {
		hasSubmitted, err := t.hasSubmittedBlockBalances(nodeAccount.Address, blockNumber)
		if err != nil {
			return err
		}
		if hasSubmitted {
			return nil
		}
	}

	// Log
	t.log.Printlnf("Calculating network balances for block %d...", blockNumber)

	// Get network balances at block
	balances, err := t.getNetworkBalances(blockNumber)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Deposit pool balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.DepositPool), 6))
	t.log.Printlnf("Total minipool user balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.MinipoolsTotal), 6))
	t.log.Printlnf("Staking minipool user balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.MinipoolsStaking), 6))
	t.log.Printlnf("rETH contract balance: %.6f ETH", math.RoundDown(eth.WeiToEth(balances.RETHContract), 6))
	t.log.Printlnf("rETH token supply: %.6f rETH", math.RoundDown(eth.WeiToEth(balances.RETHSupply), 6))

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

// Get the network balances at a specific block
func (t *submitNetworkBalances) getNetworkBalances(blockNumber uint64) (networkBalances, error) {

	// Initialize call options
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(blockNumber)),
	}

	// Data
	var wg errgroup.Group
	var depositPoolBalance *big.Int
	var minipoolBalanceDetails []minipoolBalanceDetails
	var rethContractBalance *big.Int
	var rethTotalSupply *big.Int

	// Get deposit pool balance
	wg.Go(func() error {
		var err error
		depositPoolBalance, err = deposit.GetBalance(t.rp, opts)
		return err
	})

	// Get minipool balance details
	wg.Go(func() error {
		var err error
		minipoolBalanceDetails, err = t.getNetworkMinipoolBalanceDetails(opts)
		return err
	})

	// Get rETH contract balance
	wg.Go(func() error {
		rethContractAddress, err := t.rp.GetAddress("rocketTokenRETH")
		if err != nil {
			return err
		}
		rethContractBalance, err = t.ec.BalanceAt(context.Background(), *rethContractAddress, opts.BlockNumber)
		return err
	})

	// Get rETH token supply
	wg.Go(func() error {
		var err error
		rethTotalSupply, err = tokens.GetRETHTotalSupply(t.rp, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return networkBalances{}, err
	}

	// Balances
	balances := networkBalances{
		Block:            blockNumber,
		DepositPool:      depositPoolBalance,
		MinipoolsTotal:   big.NewInt(0),
		MinipoolsStaking: big.NewInt(0),
		RETHContract:     rethContractBalance,
		RETHSupply:       rethTotalSupply,
	}

	// Add minipool balances
	for _, mp := range minipoolBalanceDetails {
		balances.MinipoolsTotal.Add(balances.MinipoolsTotal, mp.UserBalance)
		if mp.IsStaking {
			balances.MinipoolsStaking.Add(balances.MinipoolsStaking, mp.UserBalance)
		}
	}

	// Return
	return balances, nil

}

// Get all minipool balance details
func (t *submitNetworkBalances) getNetworkMinipoolBalanceDetails(opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var eth2Config beacon.Eth2Config
	var beaconHead beacon.BeaconHead
	var blockTime uint64

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetMinipoolAddresses(t.rp, opts)
		return err
	})

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = t.bc.GetEth2Config()
		return err
	})

	// Get beacon head
	wg1.Go(func() error {
		var err error
		beaconHead, err = t.bc.GetBeaconHead()
		return err
	})

	// Get block time
	wg1.Go(func() error {
		header, err := t.ec.HeaderByNumber(context.Background(), opts.BlockNumber)
		if err == nil {
			blockTime = header.Time
		}
		return err
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
	validators, err := rp.GetMinipoolValidators(t.rp, t.bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: blockEpoch})
	if err != nil {
		return []minipoolBalanceDetails{}, err
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
				mpDetails, err := t.getMinipoolBalanceDetails(address, opts, validator, eth2Config, blockEpoch)
				if err == nil {
					details[mi] = mpDetails
				}
				return err
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
func (t *submitNetworkBalances) getMinipoolBalanceDetails(minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, blockEpoch uint64) (minipoolBalanceDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Data
	var wg errgroup.Group
	var status types.MinipoolStatus
	var userDepositBalance *big.Int
	var userDepositTime uint64

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		userDepositBalance, err = mp.GetUserDepositBalance(opts)
		return err
	})
	wg.Go(func() error {
		userDepositAssignedTime, err := mp.GetUserDepositAssignedTime(opts)
		if err == nil {
			userDepositTime = uint64(userDepositAssignedTime.Unix())
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolBalanceDetails{}, err
	}

	// No balance if no user deposit assigned
	if userDepositBalance.Cmp(big.NewInt(0)) == 0 {
		return minipoolBalanceDetails{
			UserBalance: big.NewInt(0),
		}, nil
	}

	// Use user deposit balance if initialized or prelaunch
	if status == types.Initialized || status == types.Prelaunch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
		}, nil
	}

	// Use user deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			UserBalance: userDepositBalance,
		}, nil
	}

	// Get start epoch for node balance calculation
	startEpoch := eth2.EpochAt(eth2Config, userDepositTime)
	if startEpoch < validator.ActivationEpoch {
		startEpoch = validator.ActivationEpoch
	} else if startEpoch > blockEpoch {
		startEpoch = blockEpoch
	}

	// Get user balance at block
	blockBalance := eth.GweiToWei(float64(validator.Balance))
	userBalance, err := mp.CalculateUserShare(blockBalance, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}
	/*
	   nodeBalance, err := mp.CalculateNodeShare(blockBalance, opts)
	   if err != nil {
	       return minipoolBalanceDetails{}, err
	   }

	   // Log debug details
	   finalised, err := mp.GetFinalised(opts)
	   if err != nil {
	       return minipoolBalanceDetails{}, err
	   }
	   t.log.Printlnf("%s %s %s %d %s %s %s %t",
	       minipoolAddress.Hex(),
	       validator.Pubkey.Hex(),
	       blockBalance.String(),
	       blockEpoch,
	       nodeBalance.String(),
	       userBalance.String(),
	       types.MinipoolStatuses[status],
	       finalised,
	   )
	*/

	// Return
	return minipoolBalanceDetails{
		IsStaking:   (validator.ExitEpoch > blockEpoch),
		UserBalance: userBalance,
	}, nil

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

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := network.EstimateSubmitBalancesGas(t.rp, balances.Block, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit network balances: %w", err)
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
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, t.gasLimit) {
		return nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gas.Uint64()

	// Submit balances
	hash, err := network.SubmitBalances(t.rp, balances.Block, totalEth, balances.MinipoolsStaking, balances.RETHSupply, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be mined
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted network balances for block %d.", balances.Block)

	// Return
	return nil

}
