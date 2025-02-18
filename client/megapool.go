package client

import (
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type MegapoolRequester struct {
	context client.IRequesterContext
}

func NewMegapoolRequester(context client.IRequesterContext) *MegapoolRequester {
	return &MegapoolRequester{
		context: context,
	}
}

func (r *MegapoolRequester) GetName() string {
	return "Megapool"
}
func (r *MegapoolRequester) GetRoute() string {
	return "megapool"
}
func (r *MegapoolRequester) GetContext() client.IRequesterContext {
	return r.context
}

// Get node status
func (r *MegapoolRequester) Status() (*types.ApiResponse[api.MegapoolStatusData], error) {
	return client.SendGetRequest[api.MegapoolStatusData](r, "status", "Status", nil)
}
