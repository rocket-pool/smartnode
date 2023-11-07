package rocketpool

import (
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type QueueRequester struct {
	client *http.Client
}

func NewQueueRequester(client *http.Client) *QueueRequester {
	return &QueueRequester{
		client: client,
	}
}

func (r *QueueRequester) GetName() string {
	return "Queue"
}
func (r *QueueRequester) GetRoute() string {
	return "queue"
}
func (r *QueueRequester) GetClient() *http.Client {
	return r.client
}

// Process the queue
func (r *QueueRequester) Process() (*api.ApiResponse[api.QueueProcessData], error) {
	return sendGetRequest[api.QueueProcessData](r, "process", "Process", nil)
}

// Get queue status
func (r *QueueRequester) Status() (*api.ApiResponse[api.QueueStatusData], error) {
	return sendGetRequest[api.QueueStatusData](r, "status", "Status", nil)
}
