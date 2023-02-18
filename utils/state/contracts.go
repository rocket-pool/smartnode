package state

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
)

// Container for network contracts
type NetworkContracts struct {
	// Non-RP Utility
	BalanceBatcher *multicall.BalanceBatcher
	Multicaller    *multicall.MultiCaller
	ElBlockNumber  *big.Int

	// Redstone
	RocketDAOProtocolSettingsMinipool    *rocketpool.Contract
	RocketDAOProtocolSettingsNetwork     *rocketpool.Contract
	RocketDAOProtocolSettingsNode        *rocketpool.Contract
	RocketDAONodeTrustedSettingsMinipool *rocketpool.Contract
	RocketDepositPool                    *rocketpool.Contract
	RocketMinipoolManager                *rocketpool.Contract
	RocketMinipoolQueue                  *rocketpool.Contract
	RocketNetworkBalances                *rocketpool.Contract
	RocketNetworkFees                    *rocketpool.Contract
	RocketNetworkPrices                  *rocketpool.Contract
	RocketNodeDeposit                    *rocketpool.Contract
	RocketNodeDistributorFactory         *rocketpool.Contract
	RocketNodeManager                    *rocketpool.Contract
	RocketNodeStaking                    *rocketpool.Contract
	RocketRewardsPool                    *rocketpool.Contract
	RocketSmoothingPool                  *rocketpool.Contract
	RocketStorage                        *rocketpool.Contract
	RocketTokenRETH                      *rocketpool.Contract
	RocketTokenRPL                       *rocketpool.Contract
	RocketTokenRPLFixedSupply            *rocketpool.Contract

	// Atlas
	RocketMinipoolBondReducer *rocketpool.Contract
}

type contractArtifacts struct {
	name       string
	address    common.Address
	abiEncoded string
	contract   **rocketpool.Contract
}

// Get a new network contracts container
func NewNetworkContracts(rp *rocketpool.RocketPool, multicallerAddress common.Address, balanceBatcherAddress common.Address, isAtlasDeployed bool, opts *bind.CallOpts) (*NetworkContracts, error) {
	// Create the contract binding
	contracts := &NetworkContracts{
		RocketStorage: rp.RocketStorageContract,
		ElBlockNumber: opts.BlockNumber,
	}

	// Create the multicaller
	var err error
	contracts.Multicaller, err = multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return nil, err
	}

	// Create the balance batcher
	contracts.BalanceBatcher, err = multicall.NewBalanceBatcher(rp.Client, balanceBatcherAddress)
	if err != nil {
		return nil, err
	}

	// Create the contract wrappers for Redstone
	wrappers := []contractArtifacts{
		{
			name:     "rocketDAOProtocolSettingsNode",
			contract: &contracts.RocketDAOProtocolSettingsNode,
		}, {
			name:     "rocketDAOProtocolSettingsNode",
			contract: &contracts.RocketDAOProtocolSettingsNode,
		}, {
			name:     "rocketRewardsPool",
			contract: &contracts.RocketRewardsPool,
		}, {
			name:     "rocketNetworkPrices",
			contract: &contracts.RocketNetworkPrices,
		}, {
			name:     "rocketNodeManager",
			contract: &contracts.RocketNodeManager,
		}, {
			name:     "rocketNodeStaking",
			contract: &contracts.RocketNodeStaking,
		}, {
			name:     "rocketMinipoolManager",
			contract: &contracts.RocketMinipoolManager,
		}, {
			name:     "rocketNodeDistributorFactory",
			contract: &contracts.RocketNodeDistributorFactory,
		}, {
			name:     "rocketTokenRETH",
			contract: &contracts.RocketTokenRETH,
		}, {
			name:     "rocketTokenRPL",
			contract: &contracts.RocketTokenRPL,
		}, {
			name:     "rocketTokenRPLFixedSupply",
			contract: &contracts.RocketTokenRPLFixedSupply,
		}, {
			name:     "rocketNodeDeposit",
			contract: &contracts.RocketNodeDeposit,
		}, {
			name:     "rocketDAONodeTrustedSettingsMinipool",
			contract: &contracts.RocketDAONodeTrustedSettingsMinipool,
		}, {
			name:     "rocketSmoothingPool",
			contract: &contracts.RocketSmoothingPool,
		}, {
			name:     "rocketDepositPool",
			contract: &contracts.RocketDepositPool,
		}, {
			name:     "rocketMinipoolQueue",
			contract: &contracts.RocketMinipoolQueue,
		}, {
			name:     "rocketNetworkBalances",
			contract: &contracts.RocketNetworkBalances,
		}, {
			name:     "rocketNetworkFees",
			contract: &contracts.RocketNetworkFees,
		}, {
			name:     "rocketDAOProtocolSettingsNetwork",
			contract: &contracts.RocketDAOProtocolSettingsNetwork,
		}, {
			name:     "rocketDAOProtocolSettingsMinipool",
			contract: &contracts.RocketDAOProtocolSettingsMinipool,
		},
	}

	// Atlas wrappers
	if isAtlasDeployed {
		wrappers = append(wrappers, contractArtifacts{
			name:     "rocketMinipoolBondReducer",
			contract: &contracts.RocketMinipoolBondReducer,
		})
	}

	// Add the address and ABI getters to multicall
	for _, wrapper := range wrappers {
		// Add the address getter
		contracts.Multicaller.AddCall(contracts.RocketStorage, &wrapper.address, "getAddress", crypto.Keccak256Hash([]byte("contract.address"), []byte(wrapper.name)))

		// Add the ABI getter
		contracts.Multicaller.AddCall(contracts.RocketStorage, &wrapper.abiEncoded, "getString", crypto.Keccak256Hash([]byte("contract.abi"), []byte(wrapper.name)))
	}

	// Run the multi-getter
	_, err = contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return nil, fmt.Errorf("error executing multicall for contract retrieval: %w", err)
	}

	// Postprocess the contracts
	for _, wrapper := range wrappers {
		// Decode the ABI
		abi, err := rocketpool.DecodeAbi(wrapper.abiEncoded)
		if err != nil {
			return nil, fmt.Errorf("error decoding ABI for %s: %w", wrapper.name, err)
		}

		// Create the contract binding
		contract := &rocketpool.Contract{
			Contract: bind.NewBoundContract(wrapper.address, *abi, rp.Client, rp.Client, rp.Client),
			Address:  &wrapper.address,
			ABI:      abi,
			Client:   rp.Client,
		}

		// Set the contract in the main wrapper object
		*wrapper.contract = contract
	}

	return contracts, nil
}
