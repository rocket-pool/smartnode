package node

import (
	"encoding/hex"
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canSendMessage(c *cli.Command, address common.Address, message []byte) (*api.CanNodeSendMessageResponse, error) {

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
	response := api.CanNodeSendMessageResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	gasLimits, err := transactions.EstimateSendTransactionGas(ec, address, message, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating gas to send message: %w", err)
	}

	response.GasLimits = gasLimits

	return &response, nil

}

func sendMessage(c *cli.Command, address common.Address, message []byte, t *snroute.TransactOpts) (*api.NodeSendMessageResponse, error) {
	opts := t.Opts()

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
	response := api.NodeSendMessageResponse{}

	// Send the message
	hash, err := transactions.SendTransaction(ec, address, w.GetChainID(), message, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canSendMessageHandler(ctx snroute.Context) {
	addr := common.HexToAddress(ctx.Request.URL.Query().Get("address"))
	msgBytes, err := hex.DecodeString(ctx.Request.URL.Query().Get("message"))
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, fmt.Errorf("invalid message hex: %ctx.Writer", err))
		return
	}
	resp, err := canSendMessage(ctx.Command(), addr, msgBytes)
	response.WriteResponse(ctx.Writer, resp, err)
}

func sendMessageHandler(ctx snroute.WriteContext) {
	addr := common.HexToAddress(ctx.Request.FormValue("address"))
	msgBytes, err := hex.DecodeString(ctx.Request.FormValue("message"))
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, fmt.Errorf("invalid message hex: %ctx.Writer", err))
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := sendMessage(ctx.Command(), addr, msgBytes, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
