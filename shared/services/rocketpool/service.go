package rocketpool

import (
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Deletes the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from a Testnet to Mainnet).
func (c *Client) TerminateDataFolder() (api.TerminateDataFolderResponse, error) {
	return c.callAPI[api.TerminateDataFolderResponse]("POST", "/api/service/terminate-data-folder", nil, "Could not delete data folder")
}

// Gets the status of the configured Execution and Beacon clients
func (c *Client) GetClientStatus() (api.ClientStatusResponse, error) {
	return c.callAPI[api.ClientStatusResponse]("GET", "/api/service/get-client-status", nil, "Could not get client status")
}

// Restarts the Validator client
func (c *Client) RestartVc() (api.RestartVcResponse, error) {
	return c.callAPI[api.RestartVcResponse]("POST", "/api/service/restart-vc", nil, "Could not get restart-vc status")
}
