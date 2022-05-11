package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
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

func generateRewardsTree(c *cli.Context, index uint64) (*api.NetworkGenerateRewardsTreeResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Handle custom EC URLs for archive nodes
	var ec rocketpool.ExecutionClient
	customEcUrl := c.String("execution-client-url")
	if customEcUrl == "" {
		ec, err = services.GetEthClient(c)
		if err != nil {
			return nil, err
		}
	} else {
		ec, err = ethclient.Dial(customEcUrl)
		if err != nil {
			return nil, fmt.Errorf("Error connecting to custom EC: %w", err)
		}
	}

	// Response
	response := api.NetworkGenerateRewardsTreeResponse{}

	go func() {
		// Create the logger
		logger := log.NewColorLogger(NormalLogger)

		logger.Printlnf("Starting generation of Merkle rewards tree for interval %d.", index)

		// Get the current interval
		currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
		if err != nil {
			printError(err)
			return
		}
		currentIndex := currentIndexBig.Uint64()
		logger.Printlnf("Active interval is %d", currentIndex)

		// Get the interval time
		intervalTime, err := rewards.GetClaimIntervalTime(rp, nil)
		if err != nil {
			printError(fmt.Errorf("Error getting claim interval time: %w", err))
			return
		}
		logger.Printlnf("Interval time is %s", intervalTime)

		// Get the event log interval
		eventLogInterval, err := apiutils.GetEventLogInterval(cfg)
		if err != nil {
			printError(fmt.Errorf("Error getting event log interval: %w", err))
			return
		}

		// Find the event for it
		rewardsEvent, err := rewards.GetRewardSnapshotEvent(rp, index, eventLogInterval, nil)
		if err != nil {
			printError(fmt.Errorf("Error getting event for interval %d: %w", index, err))
			return
		}
		logger.Printlnf("Found snapshot event: consensus block %s", rewardsEvent.Block.String())

		// Figure out the timestamp for the block
		eth2Config, err := bc.GetEth2Config()
		if err != nil {
			printError(fmt.Errorf("Error getting Beacon config: %w", err))
			return
		}
		genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)
		blockTime := genesisTime.Add(time.Duration(rewardsEvent.Block.Uint64()*eth2Config.SecondsPerSlot) * time.Second)
		logger.Printlnf("Block time is %s", blockTime)

		// Get the matching EL block
		elBlockHeader, err := rprewards.GetELBlockHeaderForTime(blockTime, ec)
		if err != nil {
			printError(fmt.Errorf("Error getting matching EL block: %w", err))
			return
		}
		logger.Printlnf("Found matching EL block: %s", elBlockHeader.Number.String())

		// Get the total pending rewards and respective distribution percentages
		logger.Println("Calculating RPL rewards...")
		start := time.Now()
		nodeRewardsMap, networkRewardsMap, invalidNodeNetworks, err := rprewards.CalculateRplRewards(rp, elBlockHeader, intervalTime)
		if err != nil {
			printError(fmt.Errorf("Error calculating node operator rewards: %w", err))
			return
		}
		for address, network := range invalidNodeNetworks {
			logger.Printlnf("WARNING: Node %s has invalid network %d assigned!\n", address.Hex(), network)
		}
		logger.Printlnf("Finished in %s", time.Since(start).String())

		// Generate the Merkle tree
		logger.Printlnf("Generating Merkle tree...")
		start = time.Now()
		tree, err := rprewards.GenerateMerkleTree(nodeRewardsMap)
		if err != nil {
			printError(fmt.Errorf("Error generating Merkle tree: %w", err))
			return
		}
		logger.Printlnf("Finished in %s", time.Since(start).String())

		// Validate the Merkle root
		root := common.BytesToHash(tree.Root())
		if root != rewardsEvent.MerkleRoot {
			logger.Printlnf("WARNING: your Merkle tree had a root of %s, but the canonical Merkle tree's root was %s. This file will not be usable for claiming rewards.", root.Hex(), rewardsEvent.MerkleRoot.Hex())
		} else {
			logger.Printlnf("Your Merkle tree's root of %s matches the canonical root! You will be able to use this file for claiming rewards.", hexutil.Encode(tree.Root()))
		}

		// Create the JSON proof wrapper and encode it
		logger.Printlnf("Saving JSON file...")
		proofWrapper := rprewards.GenerateTreeJson(tree.Root(), nodeRewardsMap, networkRewardsMap)
		wrapperBytes, err := json.Marshal(proofWrapper)
		if err != nil {
			printError(fmt.Errorf("Error serializing proof wrapper into JSON: %w", err))
			return
		}

		// Write the file
		path := cfg.Smartnode.GetRewardsTreePath(index, true)
		err = ioutil.WriteFile(path, wrapperBytes, 0755)
		if err != nil {
			printError(fmt.Errorf("Error saving file to %s: %w", path, err))
			return
		}
		logger.Printlnf("Merkle tree generation complete!")

	}()

	return &response, nil

}

func printError(err error) {
	errorLogger := log.NewColorLogger(ErrorColor)
	errorLogger.Println(err)
	errorLogger.Println("*** Generating snapshot failed. ***")
}
