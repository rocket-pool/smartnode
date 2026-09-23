package minipool

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"

	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Settings
const MinipoolDetailsBatchSize = 10
const MinipoolPubkeyBatchSize = 50
const MinipoolQueuePositionBatchSize = 700

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

// Get all node minipool details, batched via multicall
func GetNodeMinipoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, legacyMinipoolQueueAddress *common.Address, multicallerAddress common.Address, balanceBatcherAddress common.Address) ([]api.MinipoolDetails, error) {

	// Pin every multicall batch below to the same EL block
	latestBlockNumber, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error getting latest block number: %w", err)
	}
	opts := &bind.CallOpts{BlockNumber: big.NewInt(0).SetUint64(latestBlockNumber)}

	// Build the minimal set of contracts the batched fetch needs
	mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return nil, fmt.Errorf("Error creating multicaller: %w", err)
	}
	bb, err := multicall.NewBalanceBatcher(rp.Client, balanceBatcherAddress)
	if err != nil {
		return nil, fmt.Errorf("Error creating balance batcher: %w", err)
	}
	minipoolManager, err := rp.GetContract("rocketMinipoolManager", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting the minipool manager contract: %w", err)
	}
	minipoolBondReducer, err := rp.GetContract("rocketMinipoolBondReducer", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting the minipool bond reducer contract: %w", err)
	}
	minipoolQueue, err := rp.GetContract("rocketMinipoolQueue", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting the minipool queue contract: %w", err)
	}
	rethContract, err := rp.GetContract("rocketTokenRETH", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting the rETH token contract: %w", err)
	}
	rplContract, err := rp.GetContract("rocketTokenRPL", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting the RPL token contract: %w", err)
	}
	fixedSupplyRplContract, err := rp.GetContract("rocketTokenRPLFixedSupply", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting the fixed-supply RPL token contract: %w", err)
	}
	contracts := &rpstate.NetworkContracts{
		Multicaller:               mc,
		BalanceBatcher:            bb,
		RocketMinipoolManager:     minipoolManager,
		RocketMinipoolBondReducer: minipoolBondReducer,
		RocketStorage:             rp.RocketStorageContract,
		ElBlockNumber:             opts.BlockNumber,
	}

	// Data
	var wg1 errgroup.Group
	var nativeMinipoolDetails []rpstate.NativeMinipoolDetails
	var currentEpoch uint64

	wg1.Go(func() error {
		var err error
		nativeMinipoolDetails, err = rpstate.GetNodeNativeMinipoolDetails(rp, contracts, nodeAddress)
		return err
	})

	wg1.Go(func() error {
		head, err := bc.GetBeaconHead()
		if err == nil {
			currentEpoch = head.Epoch
		}
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return []api.MinipoolDetails{}, err
	}

	addresses := make([]common.Address, len(nativeMinipoolDetails))
	pubkeys := make([]types.ValidatorPubkey, len(nativeMinipoolDetails))
	for i, nmd := range nativeMinipoolDetails {
		addresses[i] = nmd.MinipoolAddress
		pubkeys[i] = nmd.Pubkey
	}

	// Get Beacon validator statuses and the EL-side queue positions/token balances
	var wg2 errgroup.Group
	var validators map[common.Address]beacon.ValidatorStatus
	var queuePositions []int64
	var rethBalances, rplBalances, fixedSupplyRplBalances []*big.Int

	wg2.Go(func() error {
		var err error
		validators, err = buildMinipoolValidatorMap(addresses, pubkeys, bc, nil)
		return err
	})

	wg2.Go(func() error {
		var err error
		queuePositions, err = getMinipoolQueuePositions(minipoolQueue, mc, addresses, opts)
		if err != nil {
			return err
		}
		rethBalances, rplBalances, fixedSupplyRplBalances, err = getMinipoolTokenBalances(mc, rethContract, rplContract, fixedSupplyRplContract, addresses, opts)
		return err
	})

	if err := wg2.Wait(); err != nil {
		return []api.MinipoolDetails{}, err
	}

	// Compute the Beacon-chain share of balance for staking minipools with an activated validator
	beaconBalances := make([]*big.Int, len(nativeMinipoolDetails))
	minipoolPtrs := make([]*rpstate.NativeMinipoolDetails, len(nativeMinipoolDetails))
	for i := range nativeMinipoolDetails {
		nmd := &nativeMinipoolDetails[i]
		minipoolPtrs[i] = nmd
		beaconBalances[i] = big.NewInt(0)
		if nmd.Status == types.Staking || (nmd.Status == types.Dissolved && !nmd.Finalised) {
			validator := validators[nmd.MinipoolAddress]
			if validator.Exists && validator.ActivationEpoch < currentEpoch {
				beaconBalances[i] = math.GweiToWei(float64(validator.Balance))
			}
		}
	}
	if err := rpstate.CalculateCompleteMinipoolShares(rp, contracts, minipoolPtrs, beaconBalances); err != nil {
		return []api.MinipoolDetails{}, fmt.Errorf("error calculating minipool shares: %w", err)
	}

	// Build the response
	details := make([]api.MinipoolDetails, len(nativeMinipoolDetails))
	for i := range nativeMinipoolDetails {
		nmd := &nativeMinipoolDetails[i]
		details[i] = buildMinipoolDetails(nmd, validators[nmd.MinipoolAddress], queuePositions[i], currentEpoch, rethBalances[i], rplBalances[i], fixedSupplyRplBalances[i])
	}

	// Get the scrub period
	scrubPeriodSeconds, err := trustednode.GetScrubPeriod(rp, opts)
	if err != nil {
		return nil, err
	}
	scrubPeriod := time.Duration(scrubPeriodSeconds) * time.Second

	// Get the dissolve timeout
	timeout, err := protocol.GetMinipoolLaunchTimeout(rp, opts)
	if err != nil {
		return nil, err
	}

	// Get the time of the pinned block
	pinnedEth1Block, err := rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("Can't get the latest block time: %w", err)
	}
	latestBlockTime := time.Unix(int64(pinnedEth1Block.Time), 0)

	// Check the stake status of each minipool
	for i, mpDetails := range details {
		if mpDetails.Status.Status == types.Prelaunch {
			details[i].DissolveTimeout = timeout
			creationTime := mpDetails.Status.StatusTime
			dissolveTime := creationTime.Add(timeout)
			remainingTime := creationTime.Add(scrubPeriod).Sub(latestBlockTime)
			if remainingTime < 0 {
				details[i].CanStake = true
			}
			details[i].TimeUntilDissolve = time.Until(dissolveTime)
		}
	}

	// Get the promotion scrub period
	promotionScrubPeriodSeconds, err := trustednode.GetPromotionScrubPeriod(rp, opts)
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

// Build an api.MinipoolDetails from a multicall-batched NativeMinipoolDetails record, its Beacon
// validator status, and its deposit-queue position
func buildMinipoolDetails(nmd *rpstate.NativeMinipoolDetails, validator beacon.ValidatorStatus, queuePosition int64, currentEpoch uint64, rethBalance *big.Int, rplBalance *big.Int, fixedSupplyRplBalance *big.Int) api.MinipoolDetails {

	details := api.MinipoolDetails{
		Address:           nmd.MinipoolAddress,
		ValidatorPubkey:   nmd.Pubkey,
		DepositType:       nmd.DepositType,
		Finalised:         nmd.Finalised,
		UseLatestDelegate: nmd.UseLatestDelegate,
		Delegate:          nmd.Delegate,
		PreviousDelegate:  nmd.PreviousDelegate,
		EffectiveDelegate: nmd.EffectiveDelegate,
		Penalties:         nmd.PenaltyCount.Uint64(),
		Queue:             minipool.QueueDetails{Position: queuePosition},
		Status: minipool.StatusDetails{
			Status:      nmd.Status,
			StatusBlock: nmd.StatusBlock.Uint64(),
			StatusTime:  time.Unix(nmd.StatusTime.Int64(), 0),
			IsVacant:    nmd.IsVacant,
		},
		Node: minipool.NodeDetails{
			Address:         nmd.NodeAddress,
			Fee:             math.WeiToEth(nmd.NodeFee),
			DepositBalance:  nmd.NodeDepositBalance,
			RefundBalance:   nmd.NodeRefundBalance,
			DepositAssigned: nmd.NodeDepositAssigned,
		},
		User: minipool.UserDetails{
			DepositBalance:      nmd.UserDepositBalance,
			DepositAssigned:     nmd.UserDepositAssigned,
			DepositAssignedTime: time.Unix(nmd.UserDepositAssignedTime.Int64(), 0),
		},
		Balances: tokens.Balances{
			ETH:            nmd.Balance,
			RETH:           rethBalance,
			RPL:            rplBalance,
			FixedSupplyRPL: fixedSupplyRplBalance,
		},
		NodeShareOfETHBalance: nmd.NodeShareOfBalance,
	}

	details.RefundAvailable = (details.Node.RefundBalance.Cmp(big.NewInt(0)) > 0) && (details.Balances.ETH.Cmp(details.Node.RefundBalance) >= 0)
	details.CloseAvailable = (details.Status.Status == types.Dissolved)
	if details.Status.Status == types.Withdrawable {
		details.WithdrawalAvailable = true
	}

	// Get validator details if staking
	if details.Status.Status == types.Staking || (details.Status.Status == types.Dissolved && !details.Finalised) {
		details.Validator = buildMinipoolValidatorDetails(nmd, validator, currentEpoch)
	}

	return details
}

// Build a minipool's validator details from its NativeMinipoolDetails and Beacon status
func buildMinipoolValidatorDetails(nmd *rpstate.NativeMinipoolDetails, validator beacon.ValidatorStatus, currentEpoch uint64) api.ValidatorDetails {

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
		details.Balance.Add(nmd.NodeDepositBalance, nmd.UserDepositBalance)
		details.NodeBalance = new(big.Int)
		details.NodeBalance.Set(nmd.NodeDepositBalance)
		return details
	}

	// Set validator balance
	details.Balance = math.GweiToWei(float64(validator.Balance))

	// Node's share of the Beacon balance, already computed via multicall (CalculateCompleteMinipoolShares)
	details.NodeBalance = nmd.NodeShareOfBeaconBalance

	// Return
	return details

}

// Get every minipool's position in the deposit queue
func getMinipoolQueuePositions(minipoolQueue *rocketpool.Contract, mc *multicall.MultiCaller, addresses []common.Address, opts *bind.CallOpts) ([]int64, error) {

	positions := make([]int64, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolQueuePositionBatchSize {
		msi := bsi
		mei := min(bsi+MinipoolQueuePositionBatchSize, len(addresses))
		rawPositions := make([]*big.Int, mei-msi)
		for i := msi; i < mei; i++ {
			if err := mc.AddCall(minipoolQueue, &rawPositions[i-msi], "getMinipoolPosition", addresses[i]); err != nil {
				return nil, fmt.Errorf("error adding queue position call for minipool %s: %w", addresses[i].Hex(), err)
			}
		}
		if _, err := mc.FlexibleCall(true, opts); err != nil {
			return nil, fmt.Errorf("error executing multicall: %w", err)
		}
		for i := msi; i < mei; i++ {
			positions[i] = rawPositions[i-msi].Int64() + 1
		}
	}
	return positions, nil

}

// Get every minipool's rETH/RPL/fixed-supply-RPL token balances
func getMinipoolTokenBalances(mc *multicall.MultiCaller, rethContract *rocketpool.Contract, rplContract *rocketpool.Contract, fixedSupplyRplContract *rocketpool.Contract, addresses []common.Address, opts *bind.CallOpts) ([]*big.Int, []*big.Int, []*big.Int, error) {

	rethBalances := make([]*big.Int, len(addresses))
	rplBalances := make([]*big.Int, len(addresses))
	fixedSupplyRplBalances := make([]*big.Int, len(addresses))

	for bsi := 0; bsi < len(addresses); bsi += MinipoolQueuePositionBatchSize {
		msi := bsi
		mei := min(bsi+MinipoolQueuePositionBatchSize, len(addresses))
		for i := msi; i < mei; i++ {
			if err := mc.AddCall(rethContract, &rethBalances[i], "balanceOf", addresses[i]); err != nil {
				return nil, nil, nil, fmt.Errorf("error adding rETH balance call for minipool %s: %w", addresses[i].Hex(), err)
			}
			if err := mc.AddCall(rplContract, &rplBalances[i], "balanceOf", addresses[i]); err != nil {
				return nil, nil, nil, fmt.Errorf("error adding RPL balance call for minipool %s: %w", addresses[i].Hex(), err)
			}
			if err := mc.AddCall(fixedSupplyRplContract, &fixedSupplyRplBalances[i], "balanceOf", addresses[i]); err != nil {
				return nil, nil, nil, fmt.Errorf("error adding fixed-supply RPL balance call for minipool %s: %w", addresses[i].Hex(), err)
			}
		}
		if _, err := mc.FlexibleCall(true, opts); err != nil {
			return nil, nil, nil, fmt.Errorf("error executing multicall: %w", err)
		}
	}

	return rethBalances, rplBalances, fixedSupplyRplBalances, nil

}

// Build a map of minipool address -> Beacon validator status from a matching list of pubkeys
// (deduplicating/filtering null pubkeys before the batched Beacon call)
func buildMinipoolValidatorMap(addresses []common.Address, pubkeys []types.ValidatorPubkey, bc beacon.Client, validatorStatusOpts *beacon.ValidatorStatusOptions) (map[common.Address]beacon.ValidatorStatus, error) {

	// Filter out null and duplicate pubkeys
	filteredPubkeys := []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if bytes.Equal(pubkey.Bytes(), types.ValidatorPubkey{}.Bytes()) {
			continue
		}
		isDuplicate := false
		for _, pk := range filteredPubkeys {
			if bytes.Equal(pubkey.Bytes(), pk.Bytes()) {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}
		filteredPubkeys = append(filteredPubkeys, pubkey)
	}

	// Get validator statuses
	statuses, err := bc.GetValidatorStatuses(filteredPubkeys, validatorStatusOpts)
	if err != nil {
		return map[common.Address]beacon.ValidatorStatus{}, err
	}

	// Build validator map
	validators := make(map[common.Address]beacon.ValidatorStatus)
	for mi := 0; mi < len(addresses); mi++ {
		address := addresses[mi]
		pubkey := pubkeys[mi]
		status, ok := statuses[pubkey]
		if !ok {
			status = beacon.ValidatorStatus{}
		}
		validators[address] = status
	}

	// Return
	return validators, nil

}

func GetMinipoolValidators(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, callOpts *bind.CallOpts, validatorStatusOpts *beacon.ValidatorStatusOptions) (map[common.Address]beacon.ValidatorStatus, error) {

	// Load minipool validator pubkeys in batches
	pubkeys := make([]types.ValidatorPubkey, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolPubkeyBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := min(bsi+MinipoolPubkeyBatchSize, len(addresses))

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				pubkey, err := minipool.GetMinipoolPubkey(rp, address, callOpts)
				if err == nil {
					pubkeys[mi] = pubkey
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return map[common.Address]beacon.ValidatorStatus{}, err
		}

	}

	// Return
	return buildMinipoolValidatorMap(addresses, pubkeys, bc, validatorStatusOpts)

}
