package minipool

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Settings
const MinipoolDetailsBatchSize = 10

// Create a scaffolded generic minipool query, with caller-specific functionality where applicable
func createMinipoolQuery[responseType any](
	c *cli.Context,
	createBindings func(rp *rocketpool.RocketPool) error,
	getState func(node *node.Node, mc *batch.MultiCaller),
	checkState func(node *node.Node, response *responseType) bool,
	getMinipoolDetails func(mc *batch.MultiCaller, mp minipool.Minipool),
	prepareResponse func(rp *rocketpool.RocketPool, addresses []common.Address, mps []minipool.Minipool, response *responseType) error,
) (*responseType, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, fmt.Errorf("error checking if node is registered: %w", err)
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, fmt.Errorf("error getting wallet: %w", err)
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Rocket Pool binding: %w", err)
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := new(responseType)

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}
	if createBindings != nil {
		// Supplemental function-specific bindings
		err = createBindings(rp)
		if err != nil {
			return nil, err
		}
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		if getState != nil {
			// Supplemental function-specific state
			getState(node, mc)
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific check to see if minipool processing should continue
	if checkState != nil {
		if !checkState(node, response) {
			return response, nil
		}
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	if getMinipoolDetails != nil {
		err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
			getMinipoolDetails(mc, mps[i]) // Supplemental function-specific minipool details
			return nil
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting minipool details: %w", err)
		}
	}

	// Supplemental function-specific response construction
	if prepareResponse != nil {
		err = prepareResponse(rp, addresses, mps, response)
		if err != nil {
			return nil, err
		}
	}

	// Return
	return response, nil
}

// Get transaction info for an operation on all of the provided minipools, using the common minipool API (for version-agnostic functions)
func createBatchTxResponseForCommon(c *cli.Context, minipoolAddresses []common.Address, txCreator func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error), txName string) (*api.BatchTxResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		txInfo, err := txCreator(mpCommon, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, mpCommon.Details.Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}

// Get transaction info for an operation on all of the provided minipools, using the v3 minipool API (for Atlas-specific functions)
func createBatchTxResponseForV3(c *cli.Context, minipoolAddresses []common.Address, txCreator func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error), txName string) (*api.BatchTxResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			return nil, fmt.Errorf("minipool %s is too old (current version: %d); please upgrade the delegate for it first", minipoolAddress.Hex(), mpCommon.Details.Version)
		}
		txInfo, err := txCreator(mpv3, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, minipoolAddress.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}

// Get all node minipool details
func getNodeMinipoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, legacyMinipoolQueueAddress *common.Address) ([]api.MinipoolDetails, error) {

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
