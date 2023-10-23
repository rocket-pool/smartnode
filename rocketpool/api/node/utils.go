package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

// Get the node's Beacon client and make sure it's synced
func getSyncedBeaconClient(c *cli.Context) (beacon.Client, error) {
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon client: %w", err)
	}
	err = services.RequireBeaconClientSynced(c)
	if err != nil {
		return nil, fmt.Errorf("error checking Beacon client's sync status: %w", err)
	}
	return bc, nil
}

// Settings
const MinipoolCountDetailsBatchSize = 10

// Minipool count details
type minipoolCountDetails struct {
	Address             common.Address
	Status              types.MinipoolStatus
	RefundAvailable     bool
	WithdrawalAvailable bool
	CloseAvailable      bool
	Finalised           bool
	Penalties           uint64
}

// Get all node minipool count details
func getNodeMinipoolCountDetails(rp *rocketpool.RocketPool, node *node.Node) ([]minipoolCountDetails, error) {

	node.GetMinipoolAddresses(node.mini)

	// Data
	var wg1 errgroup.Group
	var addresses []common.Address
	var currentBlock uint64

	// Get minipool addresses
	wg1.Go(func() error {
		var err error
		addresses, err = minipool.GetNodeMinipoolAddresses(rp, nodeAddress, nil)
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
		return []minipoolCountDetails{}, err
	}

	// Load details in batches
	details := make([]minipoolCountDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolCountDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MinipoolCountDetailsBatchSize
		if mei > len(addresses) {
			mei = len(addresses)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				mpDetails, err := getMinipoolCountDetails(rp, address, currentBlock)
				if err == nil {
					details[mi] = mpDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []minipoolCountDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get a minipool's count details
func getMinipoolCountDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, currentBlock uint64) (minipoolCountDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return minipoolCountDetails{}, err
	}

	// Data
	var wg errgroup.Group
	var status types.MinipoolStatus
	var refundBalance *big.Int
	var finalised bool
	var penaltyCount uint64

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		refundBalance, err = mp.GetNodeRefundBalance(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		finalised, err = mp.GetFinalised(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		penaltyCount, err = minipool.GetMinipoolPenaltyCount(rp, minipoolAddress, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolCountDetails{}, err
	}

	// Return
	return minipoolCountDetails{
		Address:             minipoolAddress,
		Status:              status,
		RefundAvailable:     (refundBalance.Cmp(big.NewInt(0)) > 0),
		WithdrawalAvailable: (status == types.Withdrawable),
		CloseAvailable:      (status == types.Dissolved),
		Finalised:           finalised,
		Penalties:           penaltyCount,
	}, nil

}
