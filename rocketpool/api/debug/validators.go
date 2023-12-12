package debug

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

const MinipoolBalanceDetailsBatchSize = 20

// Get all minipool balance details
func ExportValidators(c *cli.Context) error {

	opts := &bind.CallOpts{}

	// Get services
	ec, err := services.GetEthClient(c)
	if err != nil {
		return err
	}
	rpl, err := services.GetRocketPool(c)
	if err != nil {
		return err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return err
	}

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var eth2Config beacon.Eth2Config
	var beaconHead beacon.BeaconHead
	var blockTime uint64

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetMinipoolAddresses(rpl, opts)
		return err
	})

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = bc.GetEth2Config()
		return err
	})

	// Get beacon head
	wg1.Go(func() error {
		var err error
		beaconHead, err = bc.GetBeaconHead()
		return err
	})

	// Get block time
	wg1.Go(func() error {
		header, err := ec.HeaderByNumber(context.Background(), opts.BlockNumber)
		if err == nil {
			blockTime = header.Time
		}
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return err
	}

	// Get & check epoch at block
	blockEpoch := eth2.EpochAt(eth2Config, blockTime)
	if blockEpoch > beaconHead.Epoch {
		return fmt.Errorf("Epoch %d at block %s is higher than current epoch %d", blockEpoch, opts.BlockNumber.String(), beaconHead.Epoch)
	}

	// Get minipool validator statuses
	validators, err := rp.GetMinipoolValidators(rpl, bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: &blockEpoch})
	if err != nil {
		return err
	}

	fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"Minipool Address",
		"Validator Pub Key",
		"Activation Epoch",
		"Node Fee",
		"Validator Balance",
		"Node Balance",
		"User Balance",
		"Status",
		"Finalised",
		"Active",
		"Pending",
	)

	// Load details in batches
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
				err := getMinipoolBalanceDetails(rpl, address, opts, validator, eth2Config, blockEpoch)
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return err
		}

	}

	// Return
	return nil
}

// Get minipool balance details
func getMinipoolBalanceDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, blockEpoch uint64) error {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, opts)
	if err != nil {
		return err
	}

	// Data
	var wg errgroup.Group
	var status types.MinipoolStatus
	var userDepositBalance *big.Int
	var nodeFee float64

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
		nodeFee, err = mp.GetNodeFee(opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return err
	}

	// Get user balance at block
	blockBalance := eth.GweiToWei(float64(validator.Balance))
	userBalance, err := mp.CalculateUserShare(blockBalance, opts)
	if err != nil {
		return err
	}
	nodeBalance, err := mp.CalculateNodeShare(blockBalance, opts)
	if err != nil {
		return err
	}

	// Log debug details
	finalised, err := mp.GetFinalised(opts)
	if err != nil {
		return err
	}

	if status == types.Initialized || status == types.Prelaunch {
		// Use user deposit balance if initialized or prelaunch
		userBalance = userDepositBalance
		blockBalance = eth.EthToWei(32)
		nodeBalance.Sub(blockBalance, userBalance)
	} else if status == types.Dissolved {
		userBalance = big.NewInt(0)
		blockBalance = big.NewInt(0)
		nodeBalance = big.NewInt(0)
	} else if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		// Use user deposit balance if validator not yet active on beacon chain at block
		userBalance = userDepositBalance
		blockBalance = eth.EthToWei(32)
		nodeBalance.Sub(blockBalance, userBalance)
	}

	fmt.Printf("%s\t%s\t%d\t%.10f\t%.10f\t%.10f\t%.10f\t%s\t%t\t%t\t%t\n",
		minipoolAddress.Hex(),
		validator.Pubkey.Hex(),
		validator.ActivationEpoch,
		nodeFee,
		eth.WeiToEth(blockBalance),
		eth.WeiToEth(nodeBalance),
		eth.WeiToEth(userBalance),
		types.MinipoolStatuses[status],
		finalised,
		validator.ExitEpoch > blockEpoch,
		validator.ActivationEpoch >= blockEpoch,
	)

	// Return
	return nil
}
