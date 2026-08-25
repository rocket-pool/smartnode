package service

import (
	"context"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
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
		Status:   "success",
		GasPrice: gasPrice.BaseFee,
		Error:    "",
	}, nil

}

func getGasPriceFromLatestBlockHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getGasPriceFromLatestBlock(c)
		response.WriteResponse(w, resp, err)
	}
}
