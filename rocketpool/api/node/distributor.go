package node

import (
	"context"
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func isFeeDistributorInitialized(c *cli.Context) (*api.NodeIsFeeDistributorInitializedResponse, error) {
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

	// Response
	response := api.NodeIsFeeDistributorInitializedResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the fee distributor status
	isInitialized, err := node.GetFeeDistributorInitialized(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.IsInitialized = isInitialized

	return &response, nil
}

func getInitializeFeeDistributorGas(c *cli.Context) (*api.NodeInitializeFeeDistributorGasResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.NodeInitializeFeeDistributorGasResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get fee distributor address
	distributor, err := node.GetDistributorAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.Distributor = distributor

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateInitializeFeeDistributorGas(rp, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil

}

func initializeFeeDistributor(c *cli.Context) (*api.NodeInitializeFeeDistributorResponse, error) {

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

	// Response
	response := api.NodeInitializeFeeDistributorResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Initialize the fee distributor
	hash, err := node.InitializeFeeDistributor(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canDistribute(c *cli.Context) (*api.NodeCanDistributeResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.NodeCanDistributeResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the fee distributor
	distributorAddress, err := node.GetDistributorAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	distributor, err := node.NewDistributor(rp, distributorAddress, nil)
	if err != nil {
		return nil, err
	}

	// Sync
	var wg errgroup.Group

	// Get the contract's balance
	wg.Go(func() error {
		var err error
		response.Balance, err = rp.Client.BalanceAt(context.Background(), distributorAddress, nil)
		return err
	})

	// Get the node share of the balance
	wg.Go(func() error {
		nodeShareRaw, err := distributor.GetNodeShare(nil)
		if err != nil {
			return fmt.Errorf("error getting node share for distributor %s: %w", distributorAddress.Hex(), err)
		}
		response.NodeShare = eth.WeiToEth(nodeShareRaw)
		return nil
	})

	// Get gas estimates
	wg.Go(func() error {
		var err error
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := distributor.EstimateDistributeGas(opts)
		response.GasInfo = gasInfo
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func distribute(c *cli.Context) (*api.NodeDistributeResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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

	// Response
	response := api.NodeDistributeResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get fee distributor address
	distributorAddress, err := node.GetDistributorAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Create the distributor
	distributor, err := node.NewDistributor(rp, distributorAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	hash, err := distributor.Distribute(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
