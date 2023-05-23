package rocketpool

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get queue status
func (c *Client) QueueStatus() (api.QueueStatusResponse, error) {
	responseBytes, err := c.callAPI("queue status")
	if err != nil {
		return api.QueueStatusResponse{}, fmt.Errorf("Could not get queue status: %w", err)
	}
	var response api.QueueStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.QueueStatusResponse{}, fmt.Errorf("Could not decode queue status response: %w", err)
	}
	if response.Error != "" {
		return api.QueueStatusResponse{}, fmt.Errorf("Could not get queue status: %s", response.Error)
	}
	return response, nil
}

// Check whether the queue can be processed
func (c *Client) CanProcessQueue() (api.CanProcessQueueResponse, error) {
	responseBytes, err := c.callAPI("queue can-process")
	if err != nil {
		return api.CanProcessQueueResponse{}, fmt.Errorf("Could not get can process queue status: %w", err)
	}
	var response api.CanProcessQueueResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProcessQueueResponse{}, fmt.Errorf("Could not decode can process queue response: %w", err)
	}
	if response.Error != "" {
		return api.CanProcessQueueResponse{}, fmt.Errorf("Could not get can process queue status: %s", response.Error)
	}
	return response, nil
}

// Process the queue
func (c *Client) ProcessQueue() (api.ProcessQueueResponse, error) {
	responseBytes, err := c.callAPI("queue process")
	if err != nil {
		return api.ProcessQueueResponse{}, fmt.Errorf("Could not process queue: %w", err)
	}
	var response api.ProcessQueueResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProcessQueueResponse{}, fmt.Errorf("Could not decode process queue response: %w", err)
	}
	if response.Error != "" {
		return api.ProcessQueueResponse{}, fmt.Errorf("Could not process queue: %s", response.Error)
	}
	return response, nil
}
