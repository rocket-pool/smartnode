package minipool

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Settings
const MinipoolDetailsBatchSize = 10

// Validate that a minipool belongs to a node
func validateMinipoolOwner(mp minipool.Minipool, nodeAddress common.Address) error {
	owner, err := mp.GetNodeAddress(nil)
	if err != nil {
		return err
	}
	if !bytes.Equal(owner.Bytes(), nodeAddress.Bytes()) {
		return fmt.Errorf("Minipool %s does not belong to the node", mp.GetAddress().Hex())
	}
	return nil
}

// Get all node minipool details
func GetNodeMinipoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, legacyMinipoolQueueAddress *common.Address) ([]api.MinipoolDetails, error) {

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var eth2Config beacon.Eth2Config
	var currentEpoch uint64
	var currentBlock uint64

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
		return err
	})

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = bc.GetEth2Config()
		return err
	})

	// Get current epoch
	wg1.Go(func() error {
		head, err := bc.GetBeaconHead()
		if err == nil {
			currentEpoch = head.Epoch
		}
		return err
	})

	// Get current block
	wg1.Go(func() error {
		header, err := rp.Client.HeaderByNumber(context.Background(), nil)
		if err == nil {
			currentBlock = header.Number.Uint64()
		}
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return []api.MinipoolDetails{}, err
	}

	// Get minipool validator statuses
	validators, err := rputils.GetMinipoolValidators(rp, bc, addresses, nil, nil)
	if err != nil {
		return []api.MinipoolDetails{}, err
	}

	// Load details in batches
	details := make([]api.MinipoolDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolDetailsBatchSize
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
				mpDetails, err := getMinipoolDetails(rp, address, validator, eth2Config, currentEpoch, currentBlock, legacyMinipoolQueueAddress)
				if err == nil {
					details[mi] = mpDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []api.MinipoolDetails{}, err
		}

	}

	// Get the scrub period
	scrubPeriodSeconds, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		return nil, err
	}
	scrubPeriod := time.Duration(scrubPeriodSeconds) * time.Second

	// Get the dissolve timeout
	timeout, err := protocol.GetMinipoolLaunchTimeout(rp, nil)
	if err != nil {
		return nil, err
	}

	// Get the time of the latest block
	latestEth1Block, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("Can't get the latest block time: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Check the stake status of each minipool
	for i, mpDetails := range details {
		if mpDetails.Status.Status == types.Prelaunch {
			creationTime := mpDetails.Status.StatusTime
			dissolveTime := creationTime.Add(timeout)
			remainingTime := creationTime.Add(scrubPeriod).Sub(latestBlockTime)
			if remainingTime < 0 {
				details[i].CanStake = true
				details[i].TimeUntilDissolve = time.Until(dissolveTime)
			}
		}
	}

	// Get the promotion scrub period
	promotionScrubPeriodSeconds, err := trustednode.GetPromotionScrubPeriod(rp, nil)
	if err != nil {
		return nil, err
	}
	promotionScrubPeriod := time.Duration(promotionScrubPeriodSeconds) * time.Second

	// Check the promotion status of each minipool
	for i, mpDetails := range details {
		if mpDetails.Status.IsVacant {
			creationTime := mpDetails.Status.StatusTime
			dissolveTime := creationTime.Add(timeout)
			remainingTime := creationTime.Add(promotionScrubPeriod).Sub(latestBlockTime)
			if remainingTime < 0 {
				details[i].CanPromote = true
				details[i].TimeUntilDissolve = time.Until(dissolveTime)
			}
		}
	}

	// Return
	return details, nil

}

// Get a minipool's details
func getMinipoolDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, currentEpoch, currentBlock uint64, legacyMinipoolQueueAddress *common.Address) (api.MinipoolDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return api.MinipoolDetails{}, err
	}

	// Data
	var wg errgroup.Group
	details := api.MinipoolDetails{Address: minipoolAddress}

	// Load data
	wg.Go(func() error {
		var err error
		details.ValidatorPubkey, err = minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Status, err = mp.GetStatusDetails(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.DepositType, err = minipool.GetMinipoolDepositType(rp, minipoolAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Node, err = mp.GetNodeDetails(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.User, err = mp.GetUserDetails(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Balances, err = tokens.GetBalances(rp, minipoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting minipool %s balances: %w", minipoolAddress.Hex(), err)
		}
		return err
	})
	wg.Go(func() error {
		var err error
		details.UseLatestDelegate, err = mp.GetUseLatestDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Delegate, err = mp.GetDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.PreviousDelegate, err = mp.GetPreviousDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.EffectiveDelegate, err = mp.GetEffectiveDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Finalised, err = mp.GetFinalised(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Penalties, err = minipool.GetMinipoolPenaltyCount(rp, minipoolAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Queue, err = minipool.GetQueueDetails(rp, mp.GetAddress(), nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ReduceBondTime, err = minipool.GetReduceBondTime(rp, minipoolAddress, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return api.MinipoolDetails{}, err
	}

	// Get node share of balance
	if details.Balances.ETH.Cmp(details.Node.RefundBalance) == -1 {
		details.NodeShareOfETHBalance = big.NewInt(0)
	} else {
		effectiveBalance := big.NewInt(0).Sub(details.Balances.ETH, details.Node.RefundBalance)
		details.NodeShareOfETHBalance, err = mp.CalculateNodeShare(effectiveBalance, nil)
		if err != nil {
			return api.MinipoolDetails{}, fmt.Errorf("error calculating node share: %w", err)
		}
	}

	// Get validator details if staking
	if details.Status.Status == types.Staking || (details.Status.Status == types.Dissolved && !details.Finalised) {
		validatorDetails, err := getMinipoolValidatorDetails(rp, details, validator, eth2Config, currentEpoch)
		if err != nil {
			return api.MinipoolDetails{}, err
		}
		details.Validator = validatorDetails
	}

	// Update & return
	details.RefundAvailable = (details.Node.RefundBalance.Cmp(big.NewInt(0)) > 0) && (details.Balances.ETH.Cmp(details.Node.RefundBalance) >= 0)
	details.CloseAvailable = (details.Status.Status == types.Dissolved)
	if details.Status.Status == types.Withdrawable {
		details.WithdrawalAvailable = true
	}
	return details, nil

}

// Get a minipool's validator details
func getMinipoolValidatorDetails(rp *rocketpool.RocketPool, minipoolDetails api.MinipoolDetails, validator beacon.ValidatorStatus, eth2Config beacon.Eth2Config, currentEpoch uint64) (api.ValidatorDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolDetails.Address, nil)
	if err != nil {
		return api.ValidatorDetails{}, err
	}

	// Validator details
	details := api.ValidatorDetails{}

	// Set validator status details
	validatorActivated := false
	if validator.Exists {
		details.Exists = true
		details.Active = (validator.ActivationEpoch < currentEpoch && validator.ExitEpoch > currentEpoch)
		details.Index = validator.Index
		validatorActivated = (validator.ActivationEpoch < currentEpoch)
	}

	// use deposit balances if validator not activated
	if !validatorActivated {
		details.Balance = new(big.Int)
		details.Balance.Add(minipoolDetails.Node.DepositBalance, minipoolDetails.User.DepositBalance)
		details.NodeBalance = new(big.Int)
		details.NodeBalance.Set(minipoolDetails.Node.DepositBalance)
		return details, nil
	}

	// Set validator balance
	details.Balance = eth.GweiToWei(float64(validator.Balance))

	// Get expected node balance
	blockBalance := eth.GweiToWei(float64(validator.Balance))
	nodeBalance, err := mp.CalculateNodeShare(blockBalance, nil)
	if err != nil {
		return api.ValidatorDetails{}, err
	}
	details.NodeBalance = nodeBalance

	// Return
	return details, nil

}
