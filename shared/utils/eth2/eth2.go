package eth2

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"golang.org/x/sync/errgroup"
)

// Settings
const MinipoolBalanceDetailsBatchSize = 20

// Beacon chain balance info for a minipool
type minipoolBalanceDetails struct {
	IsStaking    bool
	NodeDeposit  *big.Int
	NodeBalance  *big.Int
	TotalBalance *big.Int
}

// Get an eth2 epoch number by time
func EpochAt(config beacon.Eth2Config, time uint64) uint64 {
	return config.GenesisEpoch + (time-config.GenesisTime)/config.SecondsPerEpoch
}

// Get the balances of the minipools on the beacon chain
func GetBeaconBalances(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, beaconHead beacon.BeaconHead, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

	// Get minipool validator statuses
	validators, err := rputils.GetMinipoolValidators(rp, bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: &beaconHead.Epoch})
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

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				validator := validators[address]
				mpDetails, err := GetMinipoolBalanceDetails(rp, address, opts, validator, beaconHead.Epoch)
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

// Get the balances of the minipools on the beacon chain
func GetBeaconBalancesFromState(rp *rocketpool.RocketPool, mpds []*rpstate.NativeMinipoolDetails, state *state.NetworkState, beaconHead beacon.BeaconHead, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

	// Load details in batches
	details := make([]minipoolBalanceDetails, len(mpds))
	for bsi := 0; bsi < len(mpds); bsi += MinipoolBalanceDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolBalanceDetailsBatchSize
		if mei > len(mpds) {
			mei = len(mpds)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				mpDetails, err := GetMinipoolBalanceDetailsFromState(rp, mpds[mi], state, opts, beaconHead.Epoch)
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
func GetMinipoolBalanceDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, blockEpoch uint64) (minipoolBalanceDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}
	blockBalance := eth.GweiToWei(float64(validator.Balance))

	// Data
	var wg errgroup.Group
	var status types.MinipoolStatus
	var nodeDepositBalance *big.Int
	var finalized bool

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		nodeDepositBalance, err = mp.GetNodeDepositBalance(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		finalized, err = mp.GetFinalised(opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Deal with pools that haven't received deposits yet so their balance is still 0
	if nodeDepositBalance == nil {
		nodeDepositBalance = big.NewInt(0)
	}

	// Ignore finalized minipools
	if finalized {
		return minipoolBalanceDetails{
			NodeDeposit:  big.NewInt(0),
			NodeBalance:  big.NewInt(0),
			TotalBalance: big.NewInt(0),
		}, nil
	}

	// Use node deposit balance if initialized or prelaunch
	if status == types.Initialized || status == types.Prelaunch {
		return minipoolBalanceDetails{
			NodeDeposit:  nodeDepositBalance,
			NodeBalance:  nodeDepositBalance,
			TotalBalance: blockBalance,
		}, nil
	}

	// Use node deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			NodeDeposit:  nodeDepositBalance,
			NodeBalance:  nodeDepositBalance,
			TotalBalance: blockBalance,
		}, nil
	}

	// Get node balance at block
	nodeBalance, err := mp.CalculateNodeShare(blockBalance, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Return
	return minipoolBalanceDetails{
		IsStaking:    (validator.ExitEpoch > blockEpoch),
		NodeDeposit:  nodeDepositBalance,
		NodeBalance:  nodeBalance,
		TotalBalance: blockBalance,
	}, nil

}

// Get minipool balance details
func GetMinipoolBalanceDetailsFromState(rp *rocketpool.RocketPool, mpd *rpstate.NativeMinipoolDetails, state *state.NetworkState, opts *bind.CallOpts, blockEpoch uint64) (minipoolBalanceDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipoolFromVersion(rp, mpd.MinipoolAddress, mpd.Version, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}
	validator := state.MinipoolValidatorDetails[mpd.Pubkey]
	blockBalance := eth.GweiToWei(float64(validator.Balance))

	// Data
	status := mpd.Status
	nodeDepositBalance := mpd.NodeDepositBalance
	finalized := mpd.Finalised

	// Deal with pools that haven't received deposits yet so their balance is still 0
	if nodeDepositBalance == nil {
		nodeDepositBalance = big.NewInt(0)
	}

	// Ignore finalized minipools
	if finalized {
		return minipoolBalanceDetails{
			NodeDeposit:  big.NewInt(0),
			NodeBalance:  big.NewInt(0),
			TotalBalance: big.NewInt(0),
		}, nil
	}

	// Use node deposit balance if initialized or prelaunch
	if status == types.Initialized || status == types.Prelaunch {
		return minipoolBalanceDetails{
			NodeDeposit:  nodeDepositBalance,
			NodeBalance:  nodeDepositBalance,
			TotalBalance: blockBalance,
		}, nil
	}

	// Use node deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			NodeDeposit:  nodeDepositBalance,
			NodeBalance:  nodeDepositBalance,
			TotalBalance: blockBalance,
		}, nil
	}

	// Get node balance at block
	nodeBalance, err := mp.CalculateNodeShare(blockBalance, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Return
	return minipoolBalanceDetails{
		IsStaking:    (validator.ExitEpoch > blockEpoch),
		NodeDeposit:  nodeDepositBalance,
		NodeBalance:  nodeBalance,
		TotalBalance: blockBalance,
	}, nil

}
