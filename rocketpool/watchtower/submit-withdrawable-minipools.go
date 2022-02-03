package watchtower

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/types"
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
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Settings
const MinipoolWithdrawableDetailsBatchSize = 20

// Submit withdrawable minipools task
type submitWithdrawableMinipools struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Withdrawable minipool info
type minipoolWithdrawableDetails struct {
	Address      common.Address
	StartBalance *big.Int
	EndBalance   *big.Int
	Withdrawable bool
}

// Create submit withdrawable minipools task
func newSubmitWithdrawableMinipools(c *cli.Context, logger log.ColorLogger) (*submitWithdrawableMinipools, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
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
	return &submitWithdrawableMinipools{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		maxFee:         maxFee,
		maxPriorityFee: maxPriorityFee,
		gasLimit:       gasLimit,
	}, nil

}

// Submit withdrawable minipools
func (t *submitWithdrawableMinipools) run() error {

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
	var submitWithdrawableEnabled bool

	// Get data
	wg.Go(func() error {
		var err error
		nodeTrusted, err = trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		submitWithdrawableEnabled, err = protocol.GetMinipoolSubmitWithdrawableEnabled(t.rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	// Check node trusted status & settings
	if !(nodeTrusted && submitWithdrawableEnabled) {
		return nil
	}

	// Log
	t.log.Println("Checking for withdrawable minipools...")

	// Get minipool withdrawable details
	minipools, err := t.getNetworkMinipoolWithdrawableDetails(nodeAccount.Address)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are withdrawable...", len(minipools))

	// Submit minipools withdrawable status
	for _, details := range minipools {
		if err := t.submitWithdrawableMinipool(details); err != nil {
			t.log.Println(fmt.Errorf("Could not submit minipool %s withdrawable status: %w", details.Address.Hex(), err))
		}
	}

	// Return
	return nil

}

// Get all minipool withdrawable details
func (t *submitWithdrawableMinipools) getNetworkMinipoolWithdrawableDetails(nodeAddress common.Address) ([]minipoolWithdrawableDetails, error) {

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var eth2Config beacon.Eth2Config
	var beaconHead beacon.BeaconHead

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetMinipoolAddresses(t.rp, nil)
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

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return []minipoolWithdrawableDetails{}, err
	}

	// Get minipool validator statuses
	validators, err := rp.GetMinipoolValidators(t.rp, t.bc, addresses, nil, nil)
	if err != nil {
		return []minipoolWithdrawableDetails{}, err
	}

	// Load details in batches
	minipools := make([]minipoolWithdrawableDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolWithdrawableDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolWithdrawableDetailsBatchSize
		if mei > len(addresses) {
			mei = len(addresses)
		}

		// Log
		//t.log.Printlnf("Checking minipools %d - %d of %d for withdrawable status...", msi + 1, mei, len(addresses))

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				validator := validators[address]
				mpDetails, err := t.getMinipoolWithdrawableDetails(nodeAddress, address, validator, eth2Config, beaconHead)
				if err == nil {
					minipools[mi] = mpDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []minipoolWithdrawableDetails{}, err
		}

	}

	// Filter by withdrawable status
	withdrawableMinipools := []minipoolWithdrawableDetails{}
	for _, details := range minipools {
		if details.Withdrawable {
			withdrawableMinipools = append(withdrawableMinipools, details)
		}
	}

	// Return
	return withdrawableMinipools, nil

}

// Get minipool withdrawable details
func (t *submitWithdrawableMinipools) getMinipoolWithdrawableDetails(nodeAddress common.Address, minipoolAddress common.Address, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, beaconHead beacon.BeaconHead) (minipoolWithdrawableDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(t.rp, minipoolAddress)
	if err != nil {
		return minipoolWithdrawableDetails{}, err
	}

	// Data
	var wg errgroup.Group
	var status types.MinipoolStatus
	var nodeDepositBalance *big.Int
	var userDepositBalance *big.Int
	var userDepositTime uint64

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		nodeDepositBalance, err = mp.GetNodeDepositBalance(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		userDepositBalance, err = mp.GetUserDepositBalance(nil)
		return err
	})
	wg.Go(func() error {
		userDepositAssignedTime, err := mp.GetUserDepositAssignedTime(nil)
		if err == nil {
			userDepositTime = uint64(userDepositAssignedTime.Unix())
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolWithdrawableDetails{}, err
	}

	// Check minipool status
	if status != types.Staking {
		return minipoolWithdrawableDetails{}, nil
	}

	// Check validator status
	if !validator.Exists || validator.WithdrawableEpoch >= beaconHead.FinalizedEpoch {
		return minipoolWithdrawableDetails{}, nil
	}

	// Get start epoch for node balance calculation
	startEpoch := eth2.EpochAt(eth2Config, userDepositTime)
	if startEpoch < validator.ActivationEpoch {
		startEpoch = validator.ActivationEpoch
	} else if startEpoch > beaconHead.FinalizedEpoch {
		startEpoch = beaconHead.FinalizedEpoch
	}

	// Get validator activation balance
	activationBalanceWei := new(big.Int)
	activationBalanceWei.Add(nodeDepositBalance, userDepositBalance)
	activationBalance := eth.WeiToGwei(activationBalanceWei)

	// Calculate approximate validator balance at start epoch & validator balance at current epoch
	startBalance := eth.GweiToWei(activationBalance + (float64(validator.Balance)-activationBalance)*float64(startEpoch-validator.ActivationEpoch)/float64(beaconHead.FinalizedEpoch-validator.ActivationEpoch))
	endBalance := eth.GweiToWei(float64(validator.Balance))

	// Check for existing node submission
	nodeSubmittedMinipool, err := t.rp.RocketStorage.GetBool(nil, crypto.Keccak256Hash([]byte("minipool.withdrawable.submitted.node"), nodeAddress.Bytes(), minipoolAddress.Bytes()))
	if err != nil {
		return minipoolWithdrawableDetails{}, err
	}
	if nodeSubmittedMinipool {
		return minipoolWithdrawableDetails{}, nil
	}

	// Get the current ETH balance
	ethBalance, err := t.rp.Client.BalanceAt(context.Background(), minipoolAddress, nil)
	if err != nil {
		return minipoolWithdrawableDetails{}, err
	}

	// Get the refund balance
	refundBalance, err := mp.GetNodeRefundBalance(nil)
	if err != nil {
		return minipoolWithdrawableDetails{}, err
	}

	// Check if there's enough ETH to assume a successful withdrawal)
	remainingBalance := big.NewInt(0)
	remainingBalance.Sub(ethBalance, refundBalance)
	if remainingBalance.Cmp(endBalance) == -1 {
		return minipoolWithdrawableDetails{}, nil
	}

	// Return
	return minipoolWithdrawableDetails{
		Address:      minipoolAddress,
		StartBalance: startBalance,
		EndBalance:   endBalance,
		Withdrawable: true,
	}, nil

}

// Submit minipool withdrawable status
func (t *submitWithdrawableMinipools) submitWithdrawableMinipool(details minipoolWithdrawableDetails) error {

	// Log
	t.log.Printlnf("Submitting minipool %s withdrawable status...", details.Address.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := minipool.EstimateSubmitMinipoolWithdrawableGas(t.rp, details.Address, opts)
	if err != nil {
		return fmt.Errorf("Could not estimate the gas required to submit minipool withdrawable status: %w", err)
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

	// Dissolve
	hash, err := minipool.SubmitMinipoolWithdrawable(t.rp, details.Address, opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be mined
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully submitted minipool %s withdrawable status.", details.Address.Hex())

	// Return
	return nil

}
