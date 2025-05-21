package node

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canSendMessage(c *cli.Context, address common.Address, message []byte) (*api.CanNodeSendMessageResponse, error) {

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

func sendMessage(c *cli.Context, address common.Address, message []byte) (*api.NodeSendMessageResponse, error) {

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

	// Send the message
	hash, err := eth.SendTransaction(ec, address, w.GetChainID(), message, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
