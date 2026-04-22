package node

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"

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

	gasInfo, err := eth.EstimateSendTransactionGas(ec, address, message, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error estimating gas to send message: %w", err)
	}

	response.GasInfo = gasInfo

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
	hash, err := eth.SendTransaction(ec, address, w.GetChainID(), message, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
