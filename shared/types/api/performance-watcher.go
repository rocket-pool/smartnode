package api

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/performancewatcher"
)

type PerformanceWatcherStatusResponse struct {
	APIResponse
	Watcher performancewatcher.Status `json:"watcher"`
}
type PerformanceWatcherCandidatesResponse struct {
	APIResponse
	Reason     string                         `json:"reason,omitempty"`
	Total      int                            `json:"total"`
	Candidates []performancewatcher.Candidate `json:"candidates"`
}
type PerformanceWatcherCandidateResponse struct {
	APIResponse
	Candidate *performancewatcher.Candidate `json:"candidate"`
}
type PerformanceWatcherValidatorResponse struct {
	APIResponse
	Validator *performancewatcher.ValidatorPerformance `json:"validator"`
}

type PerformanceWatcherNodeValidatorsResponse struct {
	APIResponse
	NodeAddress common.Address                            `json:"nodeAddress"`
	Validators  []performancewatcher.ValidatorPerformance `json:"validators"`
}
