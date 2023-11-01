package rocketpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	txRoute string = "tx"
)

// Wait for a transaction
func (r *ApiRequester) WaitForTransaction(txHash common.Hash) (*api.ApiResponse[api.SuccessData], error) {
	method := "wait"
	args := map[string]string{
		"hash": txHash.Hex(),
	}
	response, err := SendGetRequest[api.SuccessData](r, fmt.Sprintf("%s/%s", txRoute, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during WaitForTransaction request: %w", err)
	}
	return response, nil
}
