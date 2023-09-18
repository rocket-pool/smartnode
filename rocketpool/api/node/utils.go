package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

// Wrapper for callbacks used by runNodeCallWithTx; this implements the caller-specific functionality
type NodeCallHandler[responseType any] interface {
	// Used to create supplemental contract bindings
	CreateBindings(ctx *callContext) error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(ctx *callContext, mc *batch.MultiCaller)

	// Prepare the response object using all of the provided artifacts
	PrepareResponse(ctx *callContext, response *responseType) error
}

// Context with services and common bindings for calls
type callContext struct {
	// Services
	w    *wallet.LocalWallet
	rp   *rocketpool.RocketPool
	cfg  *config.RocketPoolConfig
	bc   beacon.Client
	opts *bind.TransactOpts

	// Common Bindings
	node *node.Node
}

// Create a scaffolded generic call handler, with caller-specific functionality where applicable
func runNodeCall[responseType any](c *cli.Context, requireRegistered bool, h NodeCallHandler[responseType]) (*responseType, error) {
	// Get services
	if requireRegistered {
		if err := services.RequireNodeRegistered(c); err != nil {
			return nil, fmt.Errorf("error checking if node is registered: %w", err)
		}
	} else {
		if err := services.RequireNodeWallet(c); err != nil {
			return nil, fmt.Errorf("error getting node wallet: %w", err)
		}
		if err := services.RequireEthClientSynced(c); err != nil {
			return nil, fmt.Errorf("error getting Execution client sync status: %w", err)
		}
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, fmt.Errorf("error getting wallet: %w", err)
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Rocket Pool binding: %w", err)
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Smartnode config: %w", err)
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, fmt.Errorf("error getting node account transactor: %w", err)
	}

	// Response
	response := new(responseType)

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}

	// Create the context
	context := &callContext{
		w:    w,
		rp:   rp,
		cfg:  cfg,
		bc:   bc,
		opts: opts,
		node: node,
	}

	// Supplemental function-specific bindings
	err = h.CreateBindings(context)
	if err != nil {
		return nil, err
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		h.GetState(context, mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific response construction
	err = h.PrepareResponse(context, response)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}

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
