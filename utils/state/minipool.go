package state

import (
	"context"
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
	legacyMinipoolBatchSize  int = 20
	minipoolAddressBatchSize int = 1000
)

// Gets the details for a minipool using the efficient multicall contract
func GetNativeMinipoolDetails_Legacy(rp *rocketpool.RocketPool, minipoolAddress common.Address, multicallerAddress common.Address, opts *bind.CallOpts) (minipool.NativeMinipoolDetails, error) {
	contracts, err := NewNetworkContracts(rp, opts)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, err
	}

	details := minipool.NativeMinipoolDetails{}
	details.MinipoolAddress = minipoolAddress
	mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, err
	}

	addMinipoolDetailsCalls(rp, contracts, mc, &details, minipoolAddress, opts)

	_, err = mc.FlexibleCall(true)
	if err != nil {
		return minipool.NativeMinipoolDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	fixupMinipoolDetails(rp, &details, opts)

	return details, nil
}

// Gets the minpool details for a node using the efficient multicall contract
func GetNodeNativeMinipoolDetails_Legacy(rp *rocketpool.RocketPool, nodeAddress common.Address, multicallerAddress common.Address, opts *bind.CallOpts) ([]minipool.NativeMinipoolDetails, error) {
	contracts, err := NewNetworkContracts(rp, opts)
	if err != nil {
		return nil, err
	}

	// Get the list of minipool addresses for this node
	addresses, err := getNodeMinipoolAddressesFast(rp, contracts, nodeAddress, multicallerAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}
	return getBulkMinipoolDetails(rp, contracts, multicallerAddress, addresses, opts)
}

// Gets all minpool details using the efficient multicall contract
func GetAllNativeMinipoolDetails_Legacy(rp *rocketpool.RocketPool, nodeAddress common.Address, multicallerAddress common.Address, opts *bind.CallOpts) ([]minipool.NativeMinipoolDetails, error) {
	contracts, err := NewNetworkContracts(rp, opts)
	if err != nil {
		return nil, err
	}

	// Get the list of all minipool addresses
	addresses, err := getAllMinipoolAddressesFast(rp, contracts, multicallerAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}
	return getBulkMinipoolDetails(rp, contracts, multicallerAddress, addresses, opts)
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
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
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
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	return addresses, nil
}

// Get multiple minipool details at once
func getBulkMinipoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, multicallerAddress common.Address, addresses []common.Address, opts *bind.CallOpts) ([]minipool.NativeMinipoolDetails, error) {
	minipoolDetails := make([]minipool.NativeMinipoolDetails, len(addresses))

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

				addMinipoolDetailsCalls(rp, contracts, mc, details, address, opts)
			}
			_, err = mc.FlexibleCall(true)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			for j := i; j < max; j++ {
				fixupMinipoolDetails(rp, &minipoolDetails[j], opts)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	return minipoolDetails, nil
}

// Add all of the calls for the minipool details to the multicaller
func addMinipoolDetailsCalls(rp *rocketpool.RocketPool, contracts *NetworkContracts, mc *multicall.MultiCaller, details *minipool.NativeMinipoolDetails, address common.Address, opts *bind.CallOpts) error {
	// Create the minipool contract binding
	mp, err := minipool.NewMinipool(rp, address, opts)
	if err != nil {
		return err
	}
	mpContract := mp.GetContract()

	details.Version = mp.GetVersion()
	mc.AddCall(contracts.RocketMinipoolManager, &details.Exists, "getMinipoolExists", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.Pubkey, "getMinipoolPubkey", address)
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

	penaltyCountKey := crypto.Keccak256Hash([]byte("network.penalties.penalty"), address.Bytes())
	mc.AddCall(contracts.RocketStorage, &details.PenaltyCount, "getUint", penaltyCountKey)

	penaltyRatekey := crypto.Keccak256Hash([]byte("minipool.penalty.rate"), address.Bytes())
	mc.AddCall(contracts.RocketStorage, &details.PenaltyRate, "getUint", penaltyRatekey)

	// UserDistributed is v3+ only
	// Slashed is v3+ only

	mc.AddCall(mpContract, &details.NodeAddress, "getNodeAddress")

	// LastBondReductionTime is v3+ only
	// LastBondReductionPrevValue is v3+ only
	// IsVacant is v3+ only
	// NodeShareOfBalance is v3+ only
	// ReduceBondTime is v3+ only
	// ReduceBondCancelled is v3+ only
	// ReduceBondValue is v3+ only

	mc.AddCall(contracts.RocketMinipoolManager, &details.WithdrawalCredentials, "getMinipoolWithdrawalCredentials", address)

	return nil
}

// Fixes a legacy minipool details struct with supplemental logic
func fixupMinipoolDetails(rp *rocketpool.RocketPool, details *minipool.NativeMinipoolDetails, opts *bind.CallOpts) error {
	address := details.MinipoolAddress

	var err error

	// Get the minipool's ETH balance
	details.Balance, err = rp.Client.BalanceAt(context.Background(), address, opts.BlockNumber)
	if err != nil {
		return err
	}

	details.Status = types.MinipoolStatus(details.StatusRaw)
	details.DepositType = types.MinipoolDeposit(details.DepositTypeRaw)

	return nil
}
