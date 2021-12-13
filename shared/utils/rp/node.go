package rp

import (
	"context"
	"fmt"
    "math/big"
    "sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"golang.org/x/sync/errgroup"
)

func GetNodeValidatorIndices(rp *rocketpool.RocketPool, ec *ethclient.Client, bc beacon.Client, nodeAddress common.Address) ([]uint64, error) {
    // Get current block number so all subsequent queries are done at same point in time
    blockNumber, err := ec.BlockNumber(context.Background())
    if err != nil {
        return nil, fmt.Errorf("Error getting block number: %w", err)
    }

    // Setup call opts
    blockNumberBig := big.NewInt(0).SetUint64(blockNumber)
    callOpts := bind.CallOpts{BlockNumber: blockNumberBig}

    // Get node's minipool count
    minipoolCount, err := minipool.GetNodeMinipoolCount(rp, nodeAddress, &callOpts)
    if err != nil {
        return nil, fmt.Errorf("Error getting node minipool count: %w", err)
    }

    fmt.Printf("Minipool count: %d\n", minipoolCount)

    // Enumerate node's minipools and grab each of their pubkey
    var wg errgroup.Group
    var lock = sync.RWMutex{}
    pubkeys := make([]types.ValidatorPubkey, minipoolCount)

    for i := uint64(0); i < minipoolCount; i++ {
        func(index uint64) {
            wg.Go(func() error {
                minipoolAddress, err := minipool.GetNodeMinipoolAt(rp, nodeAddress, index, &callOpts)
                if err != nil {
                    return fmt.Errorf("Error getting minipool: %w", err)
                }

                pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
                if err != nil {
                    return fmt.Errorf("Error getting minipool pubkey: %w", err)
                }

                fmt.Printf("%s = %s\n", minipoolAddress.Hex(), pubkey.Hex())

                lock.Lock()
                pubkeys[index] = pubkey
                lock.Unlock()
                return nil
            })
        }(i)
    }

    // Check for errors
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Get validator statuses by pubkeys
    statuses, err := bc.GetValidatorStatuses(pubkeys, nil)
    if err != nil {
        return nil, fmt.Errorf("Error getting validator statuses: %w", err)
    }

    // Enumerate validators statuses and fill indices array
    validatorIndices := make([]uint64, len(statuses))

    i := 0
    for _, status := range statuses {
        validatorIndices[i] = status.Index
        i++
    }

    return validatorIndices, nil
}
