package debug

import (
	"github.com/urfave/cli/v3"

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
	rewardsClient := rprewards.NewRewardsExecutionClient(rp)

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
