package node

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/core"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func sendMessage(c *cli.Context, address common.Address, message []byte) (*api.TxInfoData, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.TxInfoData{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	info, err := core.NewTransactionInfoRaw(ec, address, message, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction info: %w", err)
	}
	response.TxInfo = info

	// Return response
	return &response, nil
}
