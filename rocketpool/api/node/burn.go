package node

import (
	"math/big"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canNodeBurn(c *cli.Command, amountWei *big.Int, token string) (*api.CanNodeBurnResponse, error) {

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
	response := api.CanNodeBurnResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Sync
	var wg errgroup.Group

	// Check node balance
	wg.Go(func() error {
		switch token {
		case "reth":

			// Check node rETH balance
			rethBalanceWei, err := tokens.GetRETHBalance(rp, nodeAccount.Address, nil)
			if err != nil {
				return err
			}
			response.InsufficientBalance = (amountWei.Cmp(rethBalanceWei) > 0)

		}
		return nil
	})

	// Check token contract collateral
	wg.Go(func() error {
		switch token {
		case "reth":

			// Check rETH collateral
			rethTotalCollateral, err := tokens.GetRETHTotalCollateral(rp, nil)
			if err != nil {
				return err
			}
			response.InsufficientCollateral = (amountWei.Cmp(rethTotalCollateral) > 0)

		}
		return nil
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		switch token {
		case "reth":
			gasLimits, err := tokens.EstimateBurnRETHGas(rp, amountWei, opts)
			if err == nil {
				response.GasLimits = gasLimits
			}
			return err
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanBurn = !response.InsufficientBalance && !response.InsufficientCollateral
	return &response, nil

}

func nodeBurn(c *cli.Command, amountWei *big.Int, token string, t *snroute.TransactOpts) (*api.NodeBurnResponse, error) {
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
	response := api.NodeBurnResponse{}

	// Handle token type
	switch token {
	case "reth":

		// Burn rETH
		hash, err := tokens.BurnRETH(rp, amountWei, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash

	}

	// Return response
	return &response, nil

}

func canBurnHandler(ctx snroute.Context) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	token := ctx.Request.URL.Query().Get("token")
	resp, err := canNodeBurn(ctx.Command(), amountWei, token)
	response.WriteResponse(ctx.Writer, resp, err)
}

func burnHandler(ctx snroute.WriteContext) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	token := ctx.Request.FormValue("token")
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := nodeBurn(ctx.Command(), amountWei, token, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
