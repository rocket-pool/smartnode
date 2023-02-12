package state

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	legacyMinipoolBatchSize  int = 200
	minipoolAddressBatchSize int = 2000
	minipoolVersionBatchSize int = 500
)

// Gets the details for a minipool using the efficient multicall contract
func GetNativeMinipoolDetails_Legacy(rp *rocketpool.RocketPool, minipoolAddress common.Address, multicallerAddress common.Address, isAtlasDeployed bool, opts *bind.CallOpts) (minipool.NativeMinipoolDetails, error) {
	contracts, err := NewNetworkContracts(rp, isAtlasDeployed, opts)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, err
	}

	details := minipool.NativeMinipoolDetails{}
	details.MinipoolAddress = minipoolAddress
	mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, err
	}

	version, err := rocketpool.GetContractVersion(rp, minipoolAddress, opts)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, fmt.Errorf("error getting minipool version: %w", err)
	}
	addMinipoolDetailsCalls(rp, contracts, mc, &details, version, opts)

	_, err = mc.FlexibleCall(true)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	fixupMinipoolDetails(rp, &details, opts)

	return details, nil
}

// Gets the minpool details for a node using the efficient multicall contract
func GetNodeNativeMinipoolDetails_Legacy(rp *rocketpool.RocketPool, nodeAddress common.Address, multicallerAddress common.Address, balanceBatcherAddress common.Address, isAtlasDeployed bool, opts *bind.CallOpts) ([]minipool.NativeMinipoolDetails, error) {
	contracts, err := NewNetworkContracts(rp, isAtlasDeployed, opts)
	if err != nil {
		return nil, err
	}

	balanceBatcher, err := multicall.NewBalanceBatcher(rp.Client, balanceBatcherAddress)
	if err != nil {
		return nil, err
	}

	// Get the list of minipool addresses for this node
	addresses, err := getNodeMinipoolAddressesFast(rp, contracts, nodeAddress, multicallerAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Get the list of minipool versions
	versions, err := getMinipoolVersionsFast(rp, contracts, multicallerAddress, addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool versions: %w", err)
	}

	// Get the minipool details
	return getBulkMinipoolDetails(rp, contracts, multicallerAddress, addresses, versions, balanceBatcher, opts)
}

// Gets all minpool details using the efficient multicall contract
func GetAllNativeMinipoolDetails_Legacy(rp *rocketpool.RocketPool, multicallerAddress common.Address, balanceBatcherAddress common.Address, isAtlasDeployed bool, opts *bind.CallOpts) ([]minipool.NativeMinipoolDetails, error) {
	contracts, err := NewNetworkContracts(rp, isAtlasDeployed, opts)
	if err != nil {
		return nil, err
	}

	balanceBatcher, err := multicall.NewBalanceBatcher(rp.Client, balanceBatcherAddress)
	if err != nil {
		return nil, err
	}

	// Get the list of all minipool addresses
	addresses, err := getAllMinipoolAddressesFast(rp, contracts, multicallerAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Get the list of minipool versions
	versions, err := getMinipoolVersionsFast(rp, contracts, multicallerAddress, addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool versions: %w", err)
	}

	// Get the minipool details
	return getBulkMinipoolDetails(rp, contracts, multicallerAddress, addresses, versions, balanceBatcher, opts)
}

// Get all minipool addresses using the multicaller
func getNodeMinipoolAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, nodeAddress common.Address, multicallerAddress common.Address, opts *bind.CallOpts) ([]common.Address, error) {
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
		max := i + minipoolAddressBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketMinipoolManager, &addresses[j], "getNodeMinipoolAt", nodeAddress, big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true)
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
func getAllMinipoolAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, multicallerAddress common.Address, opts *bind.CallOpts) ([]common.Address, error) {
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
		max := i + minipoolAddressBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketMinipoolManager, &addresses[j], "getMinipoolAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true)
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
func getMinipoolVersionsFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, multicallerAddress common.Address, addresses []common.Address, opts *bind.CallOpts) ([]uint8, error) {
	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	count := len(addresses)
	versions := make([]uint8, count)
	for i := 0; i < count; i += minipoolVersionBatchSize {
		i := i
		max := i + minipoolVersionBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
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
			results, err := mc.FlexibleCall(false) // Allow calls to fail - necessary for Prater
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
func getBulkMinipoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, multicallerAddress common.Address, addresses []common.Address, versions []uint8, balanceBatcher *multicall.BalanceBatcher, opts *bind.CallOpts) ([]minipool.NativeMinipoolDetails, error) {
	minipoolDetails := make([]minipool.NativeMinipoolDetails, len(addresses))

	// Get the balances of the minipools
	balances, err := balanceBatcher.GetEthBalances(addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool balances: %w", err)
	}
	for i := range minipoolDetails {
		minipoolDetails[i].Balance = balances[i]
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	count := len(addresses)
	for i := 0; i < count; i += legacyMinipoolBatchSize {
		i := i
		max := i + legacyMinipoolBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {

				address := addresses[j]
				details := &minipoolDetails[j]
				details.MinipoolAddress = address

				addMinipoolDetailsCalls(rp, contracts, mc, details, versions[j], opts)
			}
			_, err = mc.FlexibleCall(true)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Postprocess the minipools
	for i := range minipoolDetails {
		fixupMinipoolDetails(rp, &minipoolDetails[i], opts)
	}

	return minipoolDetails, nil
}

// Add all of the calls for the minipool details to the multicaller
func addMinipoolDetailsCalls(rp *rocketpool.RocketPool, contracts *NetworkContracts, mc *multicall.MultiCaller, details *minipool.NativeMinipoolDetails, version uint8, opts *bind.CallOpts) error {
	// Create the minipool contract binding
	address := details.MinipoolAddress
	mp, err := minipool.NewMinipoolFromVersion(rp, address, version, opts)
	if err != nil {
		return err
	}
	mpContract := mp.GetContract()

	details.Version = mp.GetVersion()
	mc.AddCall(contracts.RocketMinipoolManager, &details.Exists, "getMinipoolExists", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.Pubkey, "getMinipoolPubkey", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.WithdrawalCredentials, "getMinipoolWithdrawalCredentials", address)
	mc.AddCall(mpContract, &details.StatusRaw, "getStatus")
	mc.AddCall(mpContract, &details.StatusBlock, "getStatusBlock")
	mc.AddCall(mpContract, &details.StatusTime, "getStatusTime")
	mc.AddCall(mpContract, &details.Finalised, "getFinalised")
	mc.AddCall(mpContract, &details.DepositTypeRaw, "getDepositType")
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

	if version < 3 {
		// These fields are all v3+ only
		details.UserDistributed = false
		details.Slashed = false
		details.LastBondReductionTime = big.NewInt(0)
		details.LastBondReductionPrevValue = big.NewInt(0)
		details.IsVacant = false
		details.NodeShareOfBalance = big.NewInt(0)
		details.ReduceBondTime = big.NewInt(0)
		details.ReduceBondCancelled = false
		details.ReduceBondValue = big.NewInt(0)
		details.PreMigrationBalance = big.NewInt(0)
	} else {
		mc.AddCall(mpContract, &details.UserDistributed, "getUserDistributed")
		mc.AddCall(mpContract, &details.Slashed, "getSlashed")
		mc.AddCall(mpContract, &details.IsVacant, "getVacant")
		mc.AddCall(mpContract, &details.NodeShareOfBalance, "calculateNodeShare", details.Balance)
		mc.AddCall(mpContract, &details.PreMigrationBalance, "getPreMigrationBalance")

		// If v3 exists, RocketMinipoolBondReducer exists so this is safe
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.ReduceBondTime, "getReduceBondTime", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.ReduceBondCancelled, "getReduceBondCancelled", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.LastBondReductionTime, "getLastBondReductionTime", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.LastBondReductionPrevValue, "getLastBondReductionPrevValue", address)
		mc.AddCall(contracts.RocketMinipoolBondReducer, &details.ReduceBondValue, "getReduceBondValue", address)
	}

	penaltyCountKey := crypto.Keccak256Hash([]byte("network.penalties.penalty"), address.Bytes())
	mc.AddCall(contracts.RocketStorage, &details.PenaltyCount, "getUint", penaltyCountKey)

	penaltyRatekey := crypto.Keccak256Hash([]byte("minipool.penalty.rate"), address.Bytes())
	mc.AddCall(contracts.RocketStorage, &details.PenaltyRate, "getUint", penaltyRatekey)

	return nil
}

// Fixes a legacy minipool details struct with supplemental logic
func fixupMinipoolDetails(rp *rocketpool.RocketPool, details *minipool.NativeMinipoolDetails, opts *bind.CallOpts) error {

	details.Status = types.MinipoolStatus(details.StatusRaw)
	details.DepositType = types.MinipoolDeposit(details.DepositTypeRaw)

	return nil
}
