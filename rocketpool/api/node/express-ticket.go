package node

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

func getExpressTicketCount(c *cli.Command) (*api.GetExpressTicketCountResponse, error) {

	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response := api.GetExpressTicketCountResponse{}

	ticketCount, err := node.GetExpressTicketCount(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.Count = ticketCount

	return &response, nil
}

func getExpressTicketsProvisioned(c *cli.Command) (*api.GetExpressTicketsProvisionedResponse, error) {

	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response := api.GetExpressTicketsProvisionedResponse{}

	provisioned, err := node.GetExpressTicketsProvisioned(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.Provisioned = provisioned

	return &response, nil
}

func canProvisionExpressTickets(c *cli.Command) (*api.CanProvisionExpressTicketsResponse, error) {

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
	response := api.CanProvisionExpressTicketsResponse{}

	// Check node is not already provisioned
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	provisioned, err := node.GetExpressTicketsProvisioned(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.AlreadyProvisioned = provisioned

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateProvisionExpressTicketsGas(rp, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Check data
	response.CanProvision = !(response.AlreadyProvisioned)

	return &response, nil

}

func provisionExpressTickets(c *cli.Command, opts *bind.TransactOpts) (*api.ProvisionExpressTicketsResponse, error) {

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

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProvisionExpressTicketsResponse{}

	// Provision express tickets
	hash, err := node.ProvisionExpressTickets(rp, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
