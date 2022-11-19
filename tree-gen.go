package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/beacon/client"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli/v2"
)

const (
	MaxConcurrentEth1Requests = 200
)

func GenerateTree(c *cli.Context) error {

	// Configure
	configureHTTP()

	// Initialization
	currentIndex := c.Int64("interval")
	log := log.NewColorLogger(color.FgHiWhite)

	// URL acquisiton
	ecUrl := c.String("ec-endpoint")
	if ecUrl == "" {
		return fmt.Errorf("ec-endpoint must be provided")
	}
	bnUrl := c.String("bn-endpoint")
	if ecUrl == "" {
		return fmt.Errorf("bn-endpoint must be provided")
	}

	// Create the EC and BN clients
	ec, err := ethclient.Dial(ecUrl)
	if err != nil {
		return fmt.Errorf("error connecting to the EC: %w", err)
	}
	bn := client.NewStandardHttpClient(bnUrl)

	// Check which network we're on via the BN
	depositContract, err := bn.GetEth2DepositContract()
	if err != nil {
		return fmt.Errorf("error getting deposit contract from the BN: %w", err)
	}
	var network cfgtypes.Network
	switch depositContract.ChainID {
	case 1:
		network = cfgtypes.Network_Mainnet
		log.Printlnf("Beacon node is configured for Mainnet.")
	case 5:
		network = cfgtypes.Network_Prater
		log.Printlnf("Beacon node is configured for Prater.")
	default:
		return fmt.Errorf("your Beacon node is configured for an unknown network with Chain ID [%d]", depositContract.ChainID)
	}

	// Create a new config on the proper network
	cfg := config.NewRocketPoolConfig("", true)
	cfg.Smartnode.Network.Value = network

	// Create the RP wrapper
	storageContract := cfg.Smartnode.GetStorageAddress()
	rp, err := rocketpool.NewRocketPool(ec, common.HexToAddress(storageContract))
	if err != nil {
		return fmt.Errorf("error creating Rocket Pool wrapper: %w", err)
	}

	if currentIndex < 0 {
		return generateCurrentTree(log, rp, cfg, bn, c.String("output-dir"), c.Bool("pretty-print"))
	} else {
		return generatePastTree(log, rp, cfg, bn, uint64(currentIndex), c.String("output-dir"), c.Bool("pretty-print"))
	}

}

// Generates a preview / dry run of the tree for the current interval, using the latest finalized state as the endpoint instead of whatever the actual endpoint will end up being
func generateCurrentTree(log log.ColorLogger, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bn beacon.Client, outputDir string, prettyPrint bool) error {

	// Get the current interval
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return fmt.Errorf("error getting current reward index: %w", err)
	}
	currentIndex := currentIndexBig.Uint64()

	log.Printlnf("Generating a dry-run tree for the current interval (%d)", currentIndex)

	// Get the start time for the current interval, and how long an interval is supposed to take
	startTime, err := rewards.GetClaimIntervalTimeStart(rp, nil)
	if err != nil {
		return fmt.Errorf("error getting claim interval start time: %w", err)
	}
	intervalTime, err := rewards.GetClaimIntervalTime(rp, nil)
	if err != nil {
		return fmt.Errorf("error getting claim interval time: %w", err)
	}

	// Calculate the intervals passed
	latestBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	timeSinceStart := latestBlockTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := time.Now()

	// Get the latest finalized Beacon block
	snapshotBeaconBlock, elBlockNumber, beaconBlockTime, err := getLatestFinalizedSlot(log, bn)
	if err != nil {
		return fmt.Errorf("error getting latest finalized slot: %w", err)
	}

	// Get the number of the EL block matching the CL snapshot block
	var snapshotElBlockHeader *types.Header
	if elBlockNumber == 0 {
		// No EL data so the Merge hasn't happened yet, figure out the EL block based on the Epoch ending time
		snapshotElBlockHeader, err = rprewards.GetELBlockHeaderForTime(beaconBlockTime, rp)
		if err != nil {
			return fmt.Errorf("error getting EL block for time %s: %w", beaconBlockTime, err)
		}
	} else {
		snapshotElBlockHeader, err = rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
		if err != nil {
			return fmt.Errorf("error getting EL block %d: %w", elBlockNumber, err)
		}
	}
	elBlockIndex := snapshotElBlockHeader.Number.Uint64()

	// Log
	log.Printlnf("Snapshot Beacon block = %d, EL block = %d, running from %s to %s\n", snapshotBeaconBlock, elBlockIndex, startTime, endTime)

	// Generate the rewards file
	rewardsFile := rprewards.NewRewardsFile(log, "", currentIndex, startTime, endTime, snapshotBeaconBlock, snapshotElBlockHeader, uint64(intervalsPassed))
	err = rewardsFile.GenerateTree(rp, cfg, bn)
	if err != nil {
		return fmt.Errorf("error generating Merkle tree: %w", err)
	}
	for address, network := range rewardsFile.InvalidNetworkNodes {
		log.Printlnf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.\n", address.Hex(), network)
	}

	// Get the output paths
	rewardsTreePath := filepath.Join(outputDir, fmt.Sprintf(config.RewardsTreeFilenameFormat, string(cfg.Smartnode.Network.Value.(cfgtypes.Network)), currentIndex))
	minipoolPerformancePath := filepath.Join(outputDir, fmt.Sprintf(config.MinipoolPerformanceFilenameFormat, string(cfg.Smartnode.Network.Value.(cfgtypes.Network)), currentIndex))

	// Serialize the minipool performance file
	var minipoolPerformanceBytes []byte
	if prettyPrint {
		minipoolPerformanceBytes, err = json.MarshalIndent(rewardsFile.MinipoolPerformanceFile, "", "\t")
	} else {
		minipoolPerformanceBytes, err = json.Marshal(rewardsFile.MinipoolPerformanceFile)
	}
	if err != nil {
		return fmt.Errorf("error serializing minipool performance file into JSON: %w", err)
	}

	// Write it to disk
	err = ioutil.WriteFile(minipoolPerformancePath, minipoolPerformanceBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving minipool performance file to %s: %w", minipoolPerformancePath, err)
	}

	log.Printlnf("Saved minipool performance file to %s", minipoolPerformancePath)
	rewardsFile.MinipoolPerformanceFileCID = "---"

	// Serialize the rewards tree to JSON
	var wrapperBytes []byte
	if prettyPrint {
		wrapperBytes, err = json.MarshalIndent(rewardsFile, "", "\t")
	} else {
		wrapperBytes, err = json.Marshal(rewardsFile)
	}
	if err != nil {
		return fmt.Errorf("error serializing proof wrapper into JSON: %w", err)
	}
	log.Printlnf("Generation complete! Saving tree...")

	// Write the rewards tree to disk
	err = ioutil.WriteFile(rewardsTreePath, wrapperBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving rewards tree file to %s: %w", rewardsTreePath, err)
	}

	log.Printlnf("Saved rewards snapshot file to %s", rewardsTreePath)
	log.Printlnf("Successfully generated rewards snapshot for interval %d.\n", currentIndex)

	return nil
}

// Recreates an existing tree for a past interval
func generatePastTree(log log.ColorLogger, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bn beacon.Client, index uint64, outputDir string, prettyPrint bool) error {

	// Find the event for this interval
	rewardsEvent, err := rprewards.GetRewardSnapshotEvent(rp, cfg, index)
	if err != nil {
		return fmt.Errorf("error getting rewards submission event for interval %d: %w", index, err)
	}
	log.Printlnf("Found rewards submission event: Beacon block %s, execution block %s", rewardsEvent.ConsensusBlock.String(), rewardsEvent.ExecutionBlock.String())

	// Get the EL block
	elBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), rewardsEvent.ExecutionBlock)
	if err != nil {
		return fmt.Errorf("error getting execution block: %w", err)
	}

	// Generate the rewards file
	start := time.Now()
	rewardsFile := rprewards.NewRewardsFile(log, "", index, rewardsEvent.IntervalStartTime, rewardsEvent.IntervalEndTime, rewardsEvent.ConsensusBlock.Uint64(), elBlockHeader, rewardsEvent.IntervalsPassed.Uint64())
	err = rewardsFile.GenerateTree(rp, cfg, bn)
	if err != nil {
		return fmt.Errorf("error generating Merkle tree: %w", err)
	}
	for address, network := range rewardsFile.InvalidNetworkNodes {
		log.Printlnf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.", address.Hex(), network)
	}
	log.Printlnf("Finished in %s", time.Since(start).String())

	// Validate the Merkle root
	root := common.BytesToHash(rewardsFile.MerkleTree.Root())
	if root != rewardsEvent.MerkleRoot {
		log.Printlnf("WARNING: your Merkle tree had a root of %s, but the canonical Merkle tree's root was %s. This file will not be usable for claiming rewards.", root.Hex(), rewardsEvent.MerkleRoot.Hex())
	} else {
		log.Printlnf("Your Merkle tree's root of %s matches the canonical root! You will be able to use this file for claiming rewards.", rewardsFile.MerkleRoot)
	}

	// Create the JSON files
	rewardsFile.MinipoolPerformanceFileCID = "---"
	log.Printlnf("Saving JSON files...")
	var minipoolPerformanceBytes []byte
	if prettyPrint {
		minipoolPerformanceBytes, err = json.MarshalIndent(rewardsFile.MinipoolPerformanceFile, "", "\t")
	} else {
		minipoolPerformanceBytes, err = json.Marshal(rewardsFile.MinipoolPerformanceFile)
	}
	if err != nil {
		return fmt.Errorf("error serializing minipool performance file into JSON: %w", err)
	}

	var wrapperBytes []byte
	if prettyPrint {
		wrapperBytes, err = json.MarshalIndent(rewardsFile, "", "\t")
	} else {
		wrapperBytes, err = json.Marshal(rewardsFile)
	}
	if err != nil {
		return fmt.Errorf("error serializing proof wrapper into JSON: %w", err)
	}

	// Get the output paths
	rewardsTreePath := filepath.Join(outputDir, fmt.Sprintf(config.RewardsTreeFilenameFormat, string(cfg.Smartnode.Network.Value.(cfgtypes.Network)), index))
	minipoolPerformancePath := filepath.Join(outputDir, fmt.Sprintf(config.MinipoolPerformanceFilenameFormat, string(cfg.Smartnode.Network.Value.(cfgtypes.Network)), index))

	// Write the files
	err = ioutil.WriteFile(minipoolPerformancePath, minipoolPerformanceBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving minipool performance file to %s: %w", minipoolPerformancePath, err)
	}
	log.Printlnf("Saved minipool performance file to %s", minipoolPerformancePath)

	err = ioutil.WriteFile(rewardsTreePath, wrapperBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving rewards file to %s: %w", rewardsTreePath, err)
	}
	log.Printlnf("Saved rewards snapshot file to %s", rewardsTreePath)

	log.Printlnf("Successfully generated rewards snapshot for interval %d.\n", index)

	return nil

}

// Get the latest finalized slot
func getLatestFinalizedSlot(log log.ColorLogger, bn beacon.Client) (uint64, uint64, time.Time, error) {

	// Get the config
	eth2Config, err := bn.GetEth2Config()
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("error getting Beacon config: %w", err)
	}
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)

	// Get the beacon head
	beaconHead, err := bn.GetBeaconHead()
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("error getting Beacon head: %w", err)
	}

	// Get the latest finalized slot that existed
	finalizedEpoch := beaconHead.FinalizedEpoch
	finalizedSlot := (finalizedEpoch+1)*eth2Config.SlotsPerEpoch - 1
	for {
		// Try to get the current block
		block, exists, err := bn.GetBeaconBlock(fmt.Sprint(finalizedSlot))
		if err != nil {
			return 0, 0, time.Time{}, fmt.Errorf("error getting Beacon block %d: %w", finalizedSlot, err)
		}

		// If the block was missing, try the previous one
		if !exists {
			log.Printlnf("Slot %d was missing, trying the previous one...", finalizedSlot)
			finalizedSlot--
		} else {
			blockTime := genesisTime.Add(time.Duration(finalizedSlot*eth2Config.SecondsPerSlot) * time.Second)
			return finalizedSlot, block.ExecutionBlockNumber, blockTime, nil
		}
	}

}

// Configure HTTP transport settings
func configureHTTP() {

	// The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
	// The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
	// This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}
