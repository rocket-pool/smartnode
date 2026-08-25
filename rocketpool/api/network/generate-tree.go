package network

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	NormalLogger = color.FgWhite
	ErrorColor   = color.FgRed
)

func canGenerateRewardsTree(c *cli.Command, index uint64) (*api.CanNetworkGenerateRewardsTreeResponse, error) {

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
	currentIndexBig, err := rp.GetRewardIndex(nil)
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

func generateRewardsTree(c *cli.Command, index uint64) (*api.NetworkGenerateRewardsTreeResponse, error) {

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

func canGenerateRewardsTreeHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		index, err := parseUint64Param(r, "index")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canGenerateRewardsTree(c, index)
		response.WriteResponse(w, resp, err)
	}
}

func generateRewardsTreeHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		index, err := parseUint64Param(r, "index")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := generateRewardsTree(c, index)
		response.WriteResponse(w, resp, err)
	}
}
