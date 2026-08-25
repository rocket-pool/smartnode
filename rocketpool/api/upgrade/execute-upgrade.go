package upgrade

import (
	"strconv"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/dao/upgrades"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	response.CanExecute = !response.DoesNotExist && !response.InvalidState && !response.InvalidTrustedNode

	if response.CanExecute {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}
		gasLimits, err := upgrades.EstimateExecuteUpgradeGas(rp, upgradeProposalId, opts)
		if err != nil {
			return nil, err
		}
		response.GasLimits = gasLimits
	}
	return &response, nil

}

func executeUpgrade(c *cli.Command, upgradeProposalId uint64, t *snroute.TransactOpts) (*api.ExecuteTNDAOUpgradeResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ExecuteTNDAOUpgradeResponse{}

	// Execute upgrade
	hash, err := upgrades.ExecuteUpgrade(rp, upgradeProposalId, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canExecuteUpgradeHandler(ctx snroute.Context) {
	id, err := cliutils.ValidatePositiveUint("upgrade proposal ID", ctx.Request.URL.Query().Get("id"))
	if err != nil {
		response.WriteResponse(ctx.Writer, nil, err)
		return
	}
	resp, err := canExecuteUpgrade(ctx.Command(), id)
	response.WriteResponse(ctx.Writer, resp, err)
}

func executeUpgradeHandler(ctx snroute.WriteContext) {
	id, err := strconv.ParseUint(ctx.Request.URL.Query().Get("id"), 10, 64)
	if err != nil {
		response.WriteResponse(ctx.Writer, nil, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := executeUpgrade(ctx.Command(), id, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
