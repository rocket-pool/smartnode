package debug

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getRewardsEvent(c *cli.Command, interval uint64) (*api.RewardsEventResponse, error) {
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	previousRewardsPoolAddresses := cfg.Smartnode.GetPreviousRewardsPoolAddresses()
	rewardsClient := rprewards.NewRewardsExecutionClientFromConfig(rp, cfg)

	event, err := rewardsClient.GetRewardSnapshotEvent(previousRewardsPoolAddresses, interval, nil)
	if err != nil {
		return nil, err
	}

	response := api.RewardsEventResponse{
		Found: true,
	}

	response.Index = event.Index.String()
	response.ExecutionBlock = event.ExecutionBlock.String()
	response.ConsensusBlock = event.ConsensusBlock.String()
	response.MerkleRoot = event.MerkleRoot.Hex()
	response.IntervalsPassed = event.IntervalsPassed.String()
	response.TreasuryRPL = event.TreasuryRPL.String()
	response.UserETH = event.UserETH.String()
	response.IntervalStartTime = event.IntervalStartTime.Unix()
	response.IntervalEndTime = event.IntervalEndTime.Unix()
	response.SubmissionTime = event.SubmissionTime.Unix()

	response.TrustedNodeRPL = make([]string, len(event.TrustedNodeRPL))
	for i, v := range event.TrustedNodeRPL {
		response.TrustedNodeRPL[i] = v.String()
	}
	response.NodeRPL = make([]string, len(event.NodeRPL))
	for i, v := range event.NodeRPL {
		response.NodeRPL[i] = v.String()
	}
	response.NodeETH = make([]string, len(event.NodeETH))
	for i, v := range event.NodeETH {
		response.NodeETH[i] = v.String()
	}

	return &response, nil
}

func rewardsEventHandler(ctx snroute.Context) {
	raw := ctx.Request.URL.Query().Get("interval")
	if raw == "" {
		response.WriteErrorResponse(ctx.Writer, &response.BadRequestError{Err: fmt.Errorf("missing required query parameter: interval")})
		return
	}
	interval, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, &response.BadRequestError{Err: fmt.Errorf("invalid interval: %ctx.Writer", err)})
		return
	}
	resp, err := getRewardsEvent(ctx.Command(), interval)
	response.WriteResponse(ctx.Writer, resp, err)
}
