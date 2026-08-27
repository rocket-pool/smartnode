package rocketpool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get queue status
func (c *Client) QueueStatus() (api.QueueStatusResponse, error) {
	response, err := c.callAPI[api.QueueStatusResponse]("GET", "/api/queue/status", nil, "Could not get queue status")
	if err != nil {
		return response, err
	}
	if response.DepositPoolBalance == nil {
		response.DepositPoolBalance = big.NewInt(0)
	}
	if response.MinipoolQueueCapacity == nil {
		response.MinipoolQueueCapacity = big.NewInt(0)
	}
	return response, nil
}

// Check whether the queue can be processed
func (c *Client) CanProcessQueue(m uint32) (api.CanProcessQueueResponse, error) {
	return c.callAPI[api.CanProcessQueueResponse]("GET", "/api/queue/can-process", url.Values{"max": {fmt.Sprintf("%d", m)}}, "Could not get can process queue status")
}

// Process the queue
func (c *Client) ProcessQueue(m uint32) (api.ProcessQueueResponse, error) {
	return c.callAPI[api.ProcessQueueResponse]("POST", "/api/queue/process", url.Values{"max": {fmt.Sprintf("%d", m)}}, "Could not process queue")
}

// Check whether deposits can be assigned
func (c *Client) CanAssignDeposits(m uint32) (api.CanAssignDepositsResponse, error) {
	return c.callAPI[api.CanAssignDepositsResponse]("GET", "/api/queue/can-assign-deposits", url.Values{"max": {fmt.Sprintf("%d", m)}}, "Could not get can assign deposits status")
}

// Assign deposits to queued validators
func (c *Client) AssignDeposits(m uint32) (api.AssignDepositsResponse, error) {
	return c.callAPI[api.AssignDepositsResponse]("POST", "/api/queue/assign-deposits", url.Values{"max": {fmt.Sprintf("%d", m)}}, "Could not assign deposits")
}

func (c *Client) GetQueueDetails() (api.GetQueueDetailsResponse, error) {
	return c.callAPI[api.GetQueueDetailsResponse]("GET", "/api/queue/get-queue-details", nil, "Could not get total queue length")
}
