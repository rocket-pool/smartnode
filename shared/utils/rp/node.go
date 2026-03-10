package rp

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

func GetNodeValidatorIndices(rp *rocketpool.RocketPool, ec rocketpool.ExecutionClient, bc beacon.Client, nodeAddress common.Address) ([]string, error) {
	// Get current block number so all subsequent queries are done at same point in time
	blockNumber, err := ec.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error getting block number: %w", err)
	}

	// Setup call opts
	blockNumberBig := big.NewInt(0).SetUint64(blockNumber)
	callOpts := bind.CallOpts{BlockNumber: blockNumberBig}

	// Get list of pubkeys for this given node
	pubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, nodeAddress, &callOpts)
	if err != nil {
		return nil, err
	}

	// Remove zero pubkeys
	zeroPubkey := types.ValidatorPubkey{}
	filteredPubkeys := []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if !bytes.Equal(pubkey[:], zeroPubkey[:]) {
			filteredPubkeys = append(filteredPubkeys, pubkey)
		}
	}
	pubkeys = filteredPubkeys

	// Get validator statuses by pubkeys
	statuses, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator statuses: %w", err)
	}

	// Enumerate validators statuses and fill indices array
	validatorIndices := make([]string, len(statuses)+1)

	i := 0
	for _, status := range statuses {
		validatorIndices[i] = status.Index
		i++
	}

	return validatorIndices, nil
}

// Checks the given node's current borrowed ETH, its limit on borrowed ETH, and how much ETH is preparing to be borrowed by pending bond reductions
func CheckCollateral(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (ethBorrowed *big.Int, ethBorrowedLimit *big.Int, pendingBorrowAmount *big.Int, err error) {
	ethBorrowed, err = node.GetNodeETHBorrowed(rp, nodeAddress, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting node's borrowed ETH amount: %w", err)
	}

	return
}
