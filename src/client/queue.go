package client

import (
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type QueueRequester struct {
	context *client.RequesterContext
}

func NewQueueRequester(context *client.RequesterContext) *QueueRequester {
	return &QueueRequester{
		context: context,
	}
}

func (r *QueueRequester) GetName() string {
	return "Queue"
}
func (r *QueueRequester) GetRoute() string {
	return "queue"
}
func (r *QueueRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Process the queue
func (r *QueueRequester) Process() (*types.ApiResponse[api.QueueProcessData], error) {
	return client.SendGetRequest[api.QueueProcessData](r, "process", "Process", nil)
}

// Get queue status
func (r *QueueRequester) Status() (*types.ApiResponse[api.QueueStatusData], error) {
	return client.SendGetRequest[api.QueueStatusData](r, "status", "Status", nil)
}
