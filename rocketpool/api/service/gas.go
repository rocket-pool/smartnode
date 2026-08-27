package service

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getGasPriceFromLatestBlock(c *cli.Command) (*api.GasPriceFromLatestBlockResponse, error) {

	// Get the execution client
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Get the gas price from the latest block
	gasPrice, err := ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return &api.GasPriceFromLatestBlockResponse{
		APIResponse: api.APIResponse{Status: "success"},
		GasPrice:    gasPrice.BaseFee,
	}, nil

}

func getGasPriceFromLatestBlockHandler(ctx snroute.Context) {
	resp, err := getGasPriceFromLatestBlock(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
