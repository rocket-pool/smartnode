package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	minipoolBatchSize              int = 70
	minipoolCompleteShareBatchSize int = 400
	minipoolAddressBatchSize       int = 700
	minipoolVersionBatchSize       int = 400
)

// Complete details for a minipool
type NativeMinipoolDetails struct {
	// Redstone
	Exists                  bool                  `json:"exists"`
	MinipoolAddress         common.Address        `json:"minipool_address"`
	Pubkey                  types.ValidatorPubkey `json:"pubkey"`
	StatusRaw               uint8                 `json:"status_raw"`
	StatusBlock             *big.Int              `json:"status_block"`
	StatusTime              *big.Int              `json:"status_time"`
	Finalised               bool                  `json:"finalised"`
	DepositTypeRaw          uint8                 `json:"deposit_type_raw"`
	NodeFee                 *big.Int              `json:"node_fee"`
	NodeDepositBalance      *big.Int              `json:"node_deposit_balance"`
	NodeDepositAssigned     bool                  `json:"node_deposit_assigned"`
	UserDepositBalance      *big.Int              `json:"user_deposit_balance"`
	UserDepositAssigned     bool                  `json:"user_deposit_assigned"`
	UserDepositAssignedTime *big.Int              `json:"user_deposit_assigned_time"`
	UseLatestDelegate       bool                  `json:"use_latest_delegate"`
	Delegate                common.Address        `json:"delegate"`
	PreviousDelegate        common.Address        `json:"previous_delegate"`
	EffectiveDelegate       common.Address        `json:"effective_delegate"`
	PenaltyCount            *big.Int              `json:"penalty_count"`
	PenaltyRate             *big.Int              `json:"penalty_rate"`
	NodeAddress             common.Address        `json:"node_address"`
	Version                 uint8                 `json:"version"`
	Balance                 *big.Int              `json:"balance"`
	DistributableBalance    *big.Int              `json:"distributable_balance"`
	NodeShareOfBalance      *big.Int              `json:"node_share_of_balance"` // Result of calculateNodeShare(contract balance)
	UserShareOfBalance      *big.Int              `json:"user_share_of_balance"` // Result of calculateUserShare(contract balance)
	NodeRefundBalance       *big.Int              `json:"node_refund_balance"`
	WithdrawalCredentials   common.Hash           `json:"withdrawal_credentials"`
	Status                  types.MinipoolStatus  `json:"status"`
	DepositType             types.MinipoolDeposit `json:"deposit_type"`

	// Must call CalculateCompleteMinipoolShares to get these
	NodeShareOfBalanceIncludingBeacon *big.Int `json:"node_share_of_balance_including_beacon"`
	UserShareOfBalanceIncludingBeacon *big.Int `json:"user_share_of_balance_including_beacon"`
	NodeShareOfBeaconBalance          *big.Int `json:"node_share_of_beacon_balance"`
	UserShareOfBeaconBalance          *big.Int `json:"user_share_of_beacon_balance"`

	// Atlas
	UserDistributed              bool     `json:"user_distributed"`
	Slashed                      bool     `json:"slashed"`
	IsVacant                     bool     `json:"is_vacant"`
	LastBondReductionTime        *big.Int `json:"last_bond_reduction_time"`
	LastBondReductionPrevValue   *big.Int `json:"last_bond_reduction_prev_value"`
	LastBondReductionPrevNodeFee *big.Int `json:"last_bond_reduction_prev_node_fee"`
	ReduceBondTime               *big.Int `json:"reduce_bond_time"`
	ReduceBondCancelled          bool     `json:"reduce_bond_cancelled"`
	ReduceBondValue              *big.Int `json:"reduce_bond_value"`
	PreMigrationBalance          *big.Int `json:"pre_migration_balance"`
}

var sixteenEth = big.NewInt(0).Mul(big.NewInt(16), oneEth)

func (details *NativeMinipoolDetails) IsEligibleForBonuses(eligibleEnd time.Time) bool {
	// A minipool is eligible for bonuses if it was active and had a bond of less than 16 ETH during the interval
	if details.Status != types.Staking {
		return false
	}
	if details.NodeDepositBalance.Cmp(sixteenEth) >= 0 {
		return false
	}

	lastBondReductionTimestamp := details.LastBondReductionTime.Int64()
	if lastBondReductionTimestamp == 0 {
		// eligible if the bond was always under 16 eth
		return true
	}
	lastBondReductionTime := time.Unix(lastBondReductionTimestamp, 0)
	// eligible if the bond was reduced before or during the interval
	return lastBondReductionTime.Before(eligibleEnd)
}

// Gets the details for a minipool using the efficient multicall contract
func GetNativeMinipoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, minipoolAddress common.Address) (NativeMinipoolDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	details := NativeMinipoolDetails{}
	details.MinipoolAddress = minipoolAddress

	version, err := rocketpool.GetContractVersion(rp, minipoolAddress, opts)
	if err != nil {
		return NativeMinipoolDetails{}, fmt.Errorf("error getting minipool version: %w", err)
	}
	details.Version = version
	addMinipoolDetailsCalls(rp, contracts, contracts.Multicaller, &details, opts)

	_, err = contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return NativeMinipoolDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	fixupMinipoolDetails(&details)

	return details, nil
}

// Gets the minpool details for a node using the efficient multicall contract
func GetNodeNativeMinipoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, nodeAddress common.Address) ([]NativeMinipoolDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of minipool addresses for this node
	addresses, err := getNodeMinipoolAddressesFast(rp, contracts, nodeAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Get the list of minipool versions
	versions, err := getMinipoolVersionsFast(rp, contracts, addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool versions: %w", err)
	}

	// Get the minipool details
	return getBulkMinipoolDetails(rp, contracts, addresses, versions, opts)
}

// Gets all minpool details using the efficient multicall contract
func GetAllNativeMinipoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts) ([]NativeMinipoolDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of all minipool addresses
	addresses, err := getAllMinipoolAddressesFast(rp, contracts, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Get the list of minipool versions
	versions, err := getMinipoolVersionsFast(rp, contracts, addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool versions: %w", err)
	}

	// Get the minipool details
	return getBulkMinipoolDetails(rp, contracts, addresses, versions, opts)
}

// Calculate the node and user shares of the total minipool balance, including the portion on the Beacon chain
func CalculateCompleteMinipoolShares(rp *rocketpool.RocketPool, contracts *NetworkContracts, minipoolDetails []*NativeMinipoolDetails, beaconBalances []*big.Int) error {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	count := len(minipoolDetails)
	for i := 0; i < count; i += minipoolCompleteShareBatchSize {
		i := i
		max := min(i+minipoolCompleteShareBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {

				// Make the minipool contract
				details := minipoolDetails[j]
				mp, err := minipool.NewMinipoolFromVersion(rp, details.MinipoolAddress, details.Version, opts)
				if err != nil {
					return err
				}
				mpContract := mp.GetContract()

				// Calculate the Beacon shares
				beaconBalance := big.NewInt(0).Set(beaconBalances[j])
				if beaconBalance.Sign() > 0 {
					mc.AddCall(mpContract, &details.NodeShareOfBeaconBalance, "calculateNodeShare", beaconBalance)
					mc.AddCall(mpContract, &details.UserShareOfBeaconBalance, "calculateUserShare", beaconBalance)
				} else {
					details.NodeShareOfBeaconBalance = big.NewInt(0)
					details.UserShareOfBeaconBalance = big.NewInt(0)
				}

				// Calculate the total balance
				totalBalance := big.NewInt(0).Set(beaconBalances[j])      // Total balance = beacon balance
				totalBalance.Add(totalBalance, details.Balance)           // Add contract balance
				totalBalance.Sub(totalBalance, details.NodeRefundBalance) // Remove node refund

				// Calculate the node and user shares
				if totalBalance.Sign() > 0 {
					mc.AddCall(mpContract, &details.NodeShareOfBalanceIncludingBeacon, "calculateNodeShare", totalBalance)
					mc.AddCall(mpContract, &details.UserShareOfBalanceIncludingBeacon, "calculateUserShare", totalBalance)
				} else {
					details.NodeShareOfBalanceIncludingBeacon = big.NewInt(0)
					details.UserShareOfBalanceIncludingBeacon = big.NewInt(0)
				}
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("error calculating minipool shares: %w", err)
	}

	return nil
}

var oneEth = big.NewInt(1e18)

// Get the bond and node fee of a minipool for the specified time
func (details *NativeMinipoolDetails) GetMinipoolBondAndNodeFee(blockTime time.Time) (*big.Int, *big.Int) {
	currentBond := details.NodeDepositBalance
	currentFee := details.NodeFee
	previousBond := details.LastBondReductionPrevValue
	previousFee := details.LastBondReductionPrevNodeFee

	var reductionTimeBig *big.Int = details.LastBondReductionTime
	if reductionTimeBig.Cmp(common.Big0) == 0 {
		// Never reduced
		return currentBond, currentFee
	}

	reductionTime := time.Unix(reductionTimeBig.Int64(), 0)
	if reductionTime.Sub(blockTime) > 0 {
		// This block occurred before the reduction
		if previousFee.Cmp(common.Big0) == 0 {
			// Catch for minipools that were created before this call existed
			return previousBond, currentFee
		}
		return previousBond, previousFee
	}

	return currentBond, currentFee
}

// Get all minipool addresses using the multicaller
func getNodeMinipoolAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, nodeAddress common.Address, opts *bind.CallOpts) ([]common.Address, error) {
	// Get minipool count
	minipoolCount, err := minipool.GetNodeMinipoolCount(rp, nodeAddress, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	addresses := make([]common.Address, minipoolCount)

	// Run the getters in batches
	count := int(minipoolCount)
	for i := 0; i < count; i += minipoolAddressBatchSize {
		i := i
		max := min(i+minipoolAddressBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketMinipoolManager, &addresses[j], "getNodeMinipoolAt", nodeAddress, big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting minipool addresses for node %s: %w", nodeAddress.Hex(), err)
	}

	return addresses, nil
}

// Get all minipool addresses using the multicaller
func getAllMinipoolAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, opts *bind.CallOpts) ([]common.Address, error) {
	// Get minipool count
	minipoolCount, err := minipool.GetMinipoolCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	addresses := make([]common.Address, minipoolCount)

	// Run the getters in batches
	count := int(minipoolCount)
	for i := 0; i < count; i += minipoolAddressBatchSize {
		i := i
		max := min(i+minipoolAddressBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketMinipoolManager, &addresses[j], "getMinipoolAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting all minipool addresses: %w", err)
	}

	return addresses, nil
}

// Get minipool versions using the multicaller
func getMinipoolVersionsFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, addresses []common.Address, opts *bind.CallOpts) ([]uint8, error) {
	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	count := len(addresses)
	versions := make([]uint8, count)
	for i := 0; i < count; i += minipoolVersionBatchSize {
		i := i
		max := min(i+minipoolVersionBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				contract, err := rocketpool.GetRocketVersionContractForAddress(rp, addresses[j])
				if err != nil {
					return fmt.Errorf("error creating version contract for minipool %s: %w", addresses[j].Hex(), err)
				}
				mc.AddCall(contract, &versions[j], "version")
			}
			results, err := mc.FlexibleCall(false, opts) // Allow calls to fail - necessary for Prater
			for j, result := range results {
				if !result.Success {
					versions[j+i] = 1 // Anything that failed the version check didn't have the method yet so it must be v1
				}
			}
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting minipool versions: %w", err)
	}

	return versions, nil
}

// Get multiple minipool details at once
func getBulkMinipoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, addresses []common.Address, versions []uint8, opts *bind.CallOpts) ([]NativeMinipoolDetails, error) {
	minipoolDetails := make([]NativeMinipoolDetails, len(addresses))

	// Get the balances of the minipools
	balances, err := contracts.BalanceBatcher.GetEthBalances(addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool balances: %w", err)
	}
	for i := range minipoolDetails {
		minipoolDetails[i].Balance = balances[i]
	}

	// Round 1: most of the details
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	count := len(addresses)
	for i := 0; i < count; i += minipoolBatchSize {
		i := i
		max := min(i+minipoolBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {

				address := addresses[j]
				details := &minipoolDetails[j]
				details.MinipoolAddress = address
				details.Version = versions[j]

				addMinipoolDetailsCalls(rp, contracts, mc, details, opts)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting minipool details r1: %w", err)
	}

	// Round 2: NodeShare and UserShare once the refund amount has been populated
	var wg2 errgroup.Group
	wg2.SetLimit(threadLimit)
	for i := 0; i < count; i += minipoolBatchSize {
		i := i
		max := min(i+minipoolBatchSize, count)

		wg2.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				details := &minipoolDetails[j]
				details.Version = versions[j]
				addMinipoolShareCalls(rp, mc, details, opts)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			return nil
		})
	}

	if err := wg2.Wait(); err != nil {
		return nil, fmt.Errorf("error getting minipool details r2: %w", err)
	}

	// Postprocess the minipools
	for i := range minipoolDetails {
		fixupMinipoolDetails(&minipoolDetails[i])
	}

	return minipoolDetails, nil
}

// Add all of the calls for the minipool details to the multicaller
func addMinipoolDetailsCalls(rp *rocketpool.RocketPool, contracts *NetworkContracts, mc *multicall.MultiCaller, details *NativeMinipoolDetails, opts *bind.CallOpts) error {
	// Create the minipool contract binding
	address := details.MinipoolAddress
	mp, err := minipool.NewMinipoolFromVersion(rp, address, details.Version, opts)
	if err != nil {
		return err
	}
	mpContract := mp.GetContract()

	details.Version = mp.GetVersion()
	mc.AddCall(contracts.RocketMinipoolManager, &details.Exists, "getMinipoolExists", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.Pubkey, "getMinipoolPubkey", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.WithdrawalCredentials, "getMinipoolWithdrawalCredentials", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.Slashed, "getMinipoolRPLSlashed", address)
	mc.AddCall(mpContract, &details.StatusRaw, "getStatus")
	mc.AddCall(mpContract, &details.StatusBlock, "getStatusBlock")
	mc.AddCall(mpContract, &details.StatusTime, "getStatusTime")
	mc.AddCall(mpContract, &details.Finalised, "getFinalised")
	mc.AddCall(mpContract, &details.NodeFee, "getNodeFee")
	mc.AddCall(mpContract, &details.NodeDepositBalance, "getNodeDepositBalance")
	mc.AddCall(mpContract, &details.NodeDepositAssigned, "getNodeDepositAssigned")
	mc.AddCall(mpContract, &details.UserDepositBalance, "getUserDepositBalance")
	mc.AddCall(mpContract, &details.UserDepositAssigned, "getUserDepositAssigned")
	mc.AddCall(mpContract, &details.UserDepositAssignedTime, "getUserDepositAssignedTime")
	mc.AddCall(mpContract, &details.UseLatestDelegate, "getUseLatestDelegate")
	mc.AddCall(mpContract, &details.Delegate, "getDelegate")
	mc.AddCall(mpContract, &details.PreviousDelegate, "getPreviousDelegate")
	mc.AddCall(mpContract, &details.EffectiveDelegate, "getEffectiveDelegate")
	mc.AddCall(mpContract, &details.NodeAddress, "getNodeAddress")
	mc.AddCall(mpContract, &details.NodeRefundBalance, "getNodeRefundBalance")

	if details.Version < 3 {
		// These fields are all v3+ only
		details.UserDistributed = false
		details.LastBondReductionTime = big.NewInt(0)
		details.LastBondReductionPrevValue = big.NewInt(0)
		details.LastBondReductionPrevNodeFee = big.NewInt(0)
		details.IsVacant = false
		details.ReduceBondTime = big.NewInt(0)
		details.ReduceBondCancelled = false
		details.ReduceBondValue = big.NewInt(0)
		details.PreMigrationBalance = big.NewInt(0)
	} else {
		mc.AddCall(mpContract, &details.UserDistributed, "getUserDistributed")
		mc.AddCall(mpContract, &details.IsVacant, "getVacant")
		mc.AddCall(mpContract, &details.PreMigrationBalance, "getPreMigrationBalance")

		// If minipool v3 exists, RocketMinipoolBondReducer exists so this is safe
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.ReduceBondTime, "getReduceBondTime", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.ReduceBondCancelled, "getReduceBondCancelled", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.LastBondReductionTime, "getLastBondReductionTime", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.LastBondReductionPrevValue, "getLastBondReductionPrevValue", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.LastBondReductionPrevNodeFee, "getLastBondReductionPrevNodeFee", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.ReduceBondValue, "getReduceBondValue", address)
	}

	penaltyCountKey := crypto.Keccak256Hash([]byte("network.penalties.penalty"), address.Bytes())
	mc.AddCall(contracts.RocketStorage, &details.PenaltyCount, "getUint", penaltyCountKey)

	penaltyRatekey := crypto.Keccak256Hash([]byte("minipool.penalty.rate"), address.Bytes())
	mc.AddCall(contracts.RocketStorage, &details.PenaltyRate, "getUint", penaltyRatekey)

	// Query the minipool manager using the delegate-invariant function
	mc.AddCall(contracts.RocketMinipoolManager, &details.DepositTypeRaw, "getMinipoolDepositType", address)

	return nil
}

// Add the calls for the minipool node and user share to the multicaller
func addMinipoolShareCalls(rp *rocketpool.RocketPool, mc *multicall.MultiCaller, details *NativeMinipoolDetails, opts *bind.CallOpts) error {
	// Create the minipool contract binding
	address := details.MinipoolAddress
	mp, err := minipool.NewMinipoolFromVersion(rp, address, details.Version, opts)
	if err != nil {
		return err
	}
	mpContract := mp.GetContract()

	details.DistributableBalance = big.NewInt(0).Sub(details.Balance, details.NodeRefundBalance)
	if details.DistributableBalance.Sign() >= 0 {
		mc.AddCall(mpContract, &details.NodeShareOfBalance, "calculateNodeShare", details.DistributableBalance)
		mc.AddCall(mpContract, &details.UserShareOfBalance, "calculateUserShare", details.DistributableBalance)
	} else {
		details.NodeShareOfBalance = big.NewInt(0)
		details.UserShareOfBalance = big.NewInt(0)
	}

	return nil
}

// Fixes a minipool details struct with supplemental logic
func fixupMinipoolDetails(details *NativeMinipoolDetails) error {

	details.Status = types.MinipoolStatus(details.StatusRaw)
	details.DepositType = types.MinipoolDeposit(details.DepositTypeRaw)

	return nil
}
