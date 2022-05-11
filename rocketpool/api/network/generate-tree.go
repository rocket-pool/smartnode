package network

import (
	"os"

	"github.com/fatih/color"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/urfave/cli"
)

const (
	NormalLogger = color.FgWhite
	ErrorColor   = color.FgRed
)

func canGenerateRewardsTree(c *cli.Context, index uint64) (*api.CanNetworkGenerateRewardsTreeResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNetworkGenerateRewardsTreeResponse{}

	// Check if the contracts have been upgraded yet
	isUpdated, err := rputils.IsMergeUpdateDeployed(rp)
	if err != nil {
		return nil, err
	}
	response.IsUpgraded = isUpdated
	if !isUpdated {
		return &response, nil
	}

	// Get the current interval
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return nil, err
	}
	response.CurrentIndex = currentIndexBig.Uint64()

	// Get the path of the file to save
	filePath := cfg.Smartnode.GetRewardsTreePath(index, true)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		response.TreeFileExists = false
	} else {
		response.TreeFileExists = true
	}

	return &response, nil

}
