package upgrade

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canExecuteUpgrade(c *cli.Command, upgradeProposalId uint64) (*api.CanExecuteTNDAOUpgradeResponse, error) {

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

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanExecuteTNDAOUpgradeResponse{}

	// Sync
	var wg errgroup.Group

	// Check upgrade proposal exists
	wg.Go(func() error {
		upgradeProposalCount, err := upgrades.GetTotalUpgradeProposals(rp, nil)
		if err == nil {
			response.DoesNotExist = (upgradeProposalId > upgradeProposalCount)
		}
		return err
	})

	// Check proposal state
	wg.Go(func() error {
		upgradeProposalState, err := upgrades.GetUpgradeProposalState(rp, upgradeProposalId, nil)
		if err == nil {
			response.InvalidState = (upgradeProposalState != rptypes.UpgradeProposalState_Succeeded)
		}
		return err
	})

	// Check trusted node exists
	wg.Go(func() error {
		var err error
		memberExists, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
		if err == nil {
			response.InvalidTrustedNode = !memberExists
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanExecute = !(response.DoesNotExist || response.InvalidState || response.InvalidTrustedNode)

	if response.CanExecute {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}
		gasInfo, err := upgrades.EstimateExecuteUpgradeGas(rp, upgradeProposalId, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	}
	return &response, nil

}

func executeUpgrade(c *cli.Command, upgradeProposalId uint64) (*api.ExecuteTNDAOUpgradeResponse, error) {

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
	response := api.ExecuteTNDAOUpgradeResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Execute upgrade
	hash, err := upgrades.ExecuteUpgrade(rp, upgradeProposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
