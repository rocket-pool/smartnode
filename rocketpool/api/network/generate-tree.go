package network

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
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

	// Get the current interval
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return nil, err
	}
	response.CurrentIndex = currentIndexBig.Uint64()

	// Get the path of the file to save
	filePath := cfg.Smartnode.GetRewardsTreePath(index, true, config.RewardsExtensionJSON)
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		response.TreeFileExists = false
	} else {
		response.TreeFileExists = true
	}

	return &response, nil

}

func generateRewardsTree(c *cli.Context, index uint64) (*api.NetworkGenerateRewardsTreeResponse, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkGenerateRewardsTreeResponse{}

	// Create the generation request
	requestPath := cfg.Smartnode.GetRegenerateRewardsTreeRequestPath(index, true)
	requestFile, err := os.Create(requestPath)
	if requestFile != nil {
		requestFile.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Error creating request marker: %w", err)
	}

	return &response, nil

}
