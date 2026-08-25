package node

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/rocket-pool/smartnode/bindings/transactions"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

func sendMessage(c *cli.Command, address common.Address, message []byte, opts *bind.TransactOpts) (*api.NodeSendMessageResponse, error) {

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

func canSendMessageHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.URL.Query().Get("address"))
		msgBytes, err := hex.DecodeString(r.URL.Query().Get("message"))
		if err != nil {
			response.WriteErrorResponse(w, fmt.Errorf("invalid message hex: %w", err))
			return
		}
		resp, err := canSendMessage(c, addr, msgBytes)
		response.WriteResponse(w, resp, err)
	}
}

func sendMessageHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.FormValue("address"))
		msgBytes, err := hex.DecodeString(r.FormValue("message"))
		if err != nil {
			response.WriteErrorResponse(w, fmt.Errorf("invalid message hex: %w", err))
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := sendMessage(c, addr, msgBytes, opts)
		response.WriteResponse(w, resp, err)
	}
}
