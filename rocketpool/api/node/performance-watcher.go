package node

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/performancewatcher"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func performanceWatcherStatusHandler(ctx snroute.Context) {
	result := &api.PerformanceWatcherStatusResponse{}
	cfg, err := services.GetConfig(ctx.Command())
	if err != nil {
		response.WriteResponse(ctx.Writer, result, err)
		return
	}
	enabled := cfg.Smartnode.RunPerformanceWatcher.Value == true
	result.Watcher = performancewatcher.DisabledStatus(enabled, "disabled")
	result.Watcher.BackfillEnabled = cfg.Smartnode.PerformanceWatcherBackfill.Value == true
	if enabled {
		result.Watcher.Health = "starting"
	}
	if services.IsStaticStateMode(ctx.Command()) {
		result.Watcher.Health = "unavailable in static network-state mode"
	}
	if w := services.GetPerformanceWatcher(); w != nil {
		result.Watcher = w.Status()
	}
	response.WriteResponse(ctx.Writer, result, nil)
}
func watcherPage(value string, defaultValue, maxValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(value)
	if err != nil || v < 0 || v > maxValue {
		return 0, &response.BadRequestError{Err: fmt.Errorf("invalid pagination value %q", value)}
	}
	return v, nil
}
func performanceWatcherCandidatesHandler(ctx snroute.Context) {
	result := &api.PerformanceWatcherCandidatesResponse{Candidates: []performancewatcher.Candidate{}}
	offset, err := watcherPage(ctx.Request.FormValue("offset"), 0, int(^uint(0)>>1))
	if err != nil {
		response.WriteResponse(ctx.Writer, result, err)
		return
	}
	limit, err := watcherPage(ctx.Request.FormValue("limit"), 50, 100)
	if err != nil || limit == 0 {
		if err == nil {
			err = &response.BadRequestError{Err: fmt.Errorf("limit must be positive")}
		}
		response.WriteResponse(ctx.Writer, result, err)
		return
	}
	if w := services.GetPerformanceWatcher(); w != nil {
		result.Candidates, result.Total, result.Reason = w.List(offset, limit)
	} else {
		result.Reason = "Performance watcher is disabled or unavailable."
	}
	response.WriteResponse(ctx.Writer, result, nil)
}
func performanceWatcherCandidateHandler(ctx snroute.Context) {
	result := &api.PerformanceWatcherCandidateResponse{}
	index, err := strconv.ParseUint(ctx.Request.FormValue("validatorIndex"), 10, 64)
	if err != nil {
		response.WriteResponse(ctx.Writer, result, &response.BadRequestError{Err: fmt.Errorf("invalid validator index")})
		return
	}
	w := services.GetPerformanceWatcher()
	if w == nil {
		response.WriteResponse(ctx.Writer, result, fmt.Errorf("performance watcher is disabled or unavailable"))
		return
	}
	result.Candidate, err = w.Candidate(ctx.Request.Context(), index)
	response.WriteResponse(ctx.Writer, result, err)
}

func performanceWatcherValidatorHandler(ctx snroute.Context) {
	result := &api.PerformanceWatcherValidatorResponse{}
	index, err := strconv.ParseUint(ctx.Request.FormValue("validatorIndex"), 10, 64)
	if err != nil {
		response.WriteResponse(ctx.Writer, result, &response.BadRequestError{Err: fmt.Errorf("invalid validator index")})
		return
	}
	w := services.GetPerformanceWatcher()
	if w == nil {
		response.WriteResponse(ctx.Writer, result, fmt.Errorf("performance watcher is disabled or unavailable"))
		return
	}
	result.Validator, err = w.Validator(ctx.Request.Context(), index)
	if errors.Is(err, performancewatcher.ErrValidatorNotTracked) {
		err = &response.NotFoundError{Path: fmt.Sprintf("performance watcher validator %d", index)}
	}
	response.WriteResponse(ctx.Writer, result, err)
}

func performanceWatcherNodeValidatorsHandler(ctx snroute.Context) {
	result := &api.PerformanceWatcherNodeValidatorsResponse{Validators: []performancewatcher.ValidatorPerformance{}}
	w := services.GetPerformanceWatcher()
	if w == nil {
		response.WriteResponse(ctx.Writer, result, fmt.Errorf("performance watcher is disabled or unavailable"))
		return
	}
	wallet, err := services.GetWallet(ctx.Command())
	if err != nil {
		response.WriteResponse(ctx.Writer, result, err)
		return
	}
	account, err := wallet.GetNodeAccount()
	if err != nil {
		response.WriteResponse(ctx.Writer, result, err)
		return
	}
	result.NodeAddress = account.Address
	result.Validators, err = w.NodeValidators(ctx.Request.Context(), account.Address.Hex())
	response.WriteResponse(ctx.Writer, result, err)
}
