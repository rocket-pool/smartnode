package client

import (
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type ServiceRequester struct {
	context *client.RequesterContext
}

func NewServiceRequester(context *client.RequesterContext) *ServiceRequester {
	return &ServiceRequester{
		context: context,
	}
}

func (r *ServiceRequester) GetName() string {
	return "Service"
}
func (r *ServiceRequester) GetRoute() string {
	return "service"
}
func (r *ServiceRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Gets the status of the configured Execution and Beacon clients
func (r *ServiceRequester) ClientStatus() (*types.ApiResponse[api.ServiceClientStatusData], error) {
	return client.SendGetRequest[api.ServiceClientStatusData](r, "client-status", "ClientStatus", nil)
}

// Restarts the Validator client
func (r *ServiceRequester) RestartVc() (*types.ApiResponse[types.SuccessData], error) {
	return client.SendGetRequest[types.SuccessData](r, "restart-vc", "RestartVc", nil)
}

// Deletes the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from a Testnet to Mainnet).
func (r *ServiceRequester) TerminateDataFolder() (*types.ApiResponse[api.ServiceTerminateDataFolderData], error) {
	return client.SendGetRequest[api.ServiceTerminateDataFolderData](r, "terminate-data-folder", "TerminateDataFolder", nil)
}

// Gets the version of the daemon
func (r *ServiceRequester) Version() (*types.ApiResponse[api.ServiceVersionData], error) {
	return client.SendGetRequest[api.ServiceVersionData](r, "version", "Version", nil)
}
