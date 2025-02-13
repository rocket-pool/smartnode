package rocketpool

import (
	"fmt"
	"math/big"

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
	if response.DepositPoolBalance == nil {
		response.DepositPoolBalance = big.NewInt(0)
	}
	if response.MinipoolQueueCapacity == nil {
		response.MinipoolQueueCapacity = big.NewInt(0)
	}
	return response, nil
}

// Check whether the queue can be processed
func (c *Client) CanProcessQueue(max uint32) (api.CanProcessQueueResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("queue can-process %d", max))
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
func (c *Client) ProcessQueue(max uint32) (api.ProcessQueueResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("queue process %d", max))
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

func (c *Client) GetTotalQueueLength() (api.GetQueueLengthResponse, error) {
	responseBytes, err := c.callAPI("queue get-total-length")
	if err != nil {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not get total queue length: %w", err)
	}
	var response api.GetQueueLengthResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not decode get total queue length response: %w", err)
	}
	if response.Error != "" {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not get total queue length: %s", response.Error)
	}
	return response, nil
}

func (c *Client) GetExpressQueueLength() (api.GetQueueLengthResponse, error) {
	responseBytes, err := c.callAPI("queue get-express-length")
	if err != nil {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not get express queue length: %w", err)
	}
	var response api.GetQueueLengthResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not decode get express queue length response: %w", err)
	}
	if response.Error != "" {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not get express queue length: %s", response.Error)
	}
	return response, nil
}

func (c *Client) GetStandardQueueLength() (api.GetQueueLengthResponse, error) {
	responseBytes, err := c.callAPI("queue get-standard-length")
	if err != nil {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not get standard queue length: %w", err)
	}
	var response api.GetQueueLengthResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not decode get standard queue length response: %w", err)
	}
	if response.Error != "" {
		return api.GetQueueLengthResponse{}, fmt.Errorf("Could not get standard queue length: %s", response.Error)
	}
	return response, nil
}
