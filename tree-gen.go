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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/beacon/client"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli/v2"
)

const (
	MaxConcurrentEth1Requests = 200
)

type snapshotDetails struct {
	index                 uint64
	startTime             time.Time
	endTime               time.Time
	snapshotBeaconBlock   uint64
	snapshotElBlockHeader *types.Header
	intervalsPassed       uint64
}

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
	case 1337803:
		network = cfgtypes.Network_Zhejiang
		log.Printlnf("Beacon node is configured for Zhejiang.")
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

	// Print the network info and exit if requested
	if c.Bool("network-info") {
		printNetworkInfo(rp, cfg, bn, log, nil)
		return nil
	}

	// Run the tree generation or the rETH SP approximation
	if c.Bool("approximate-only") {
		return approximateCurrentRethSpRewards(log, rp, cfg, bn, c.String("output-dir"), c.Bool("pretty-print"), c.Uint64("ruleset"))
	} else {
		if currentIndex < 0 {
			return generateCurrentTree(log, rp, cfg, bn, c.String("output-dir"), c.Bool("pretty-print"), c.Uint64("ruleset"))
		} else {
			return generatePastTree(log, rp, cfg, bn, uint64(currentIndex), c.String("output-dir"), c.Bool("pretty-print"), c.Uint64("ruleset"))
		}
	}

}

// Generates a preview / dry run of the tree for the current interval, using the latest finalized state as the endpoint instead of whatever the actual endpoint will end up being
func generateCurrentTree(log log.ColorLogger, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bn beacon.Client, outputDir string, prettyPrint bool, ruleset uint64) error {

	// Get a snapshot of the network state
	mgr, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bn, &log)
	if err != nil {
		return fmt.Errorf("error creating network state manager: %w", err)
	}
	block, err := mgr.GetLatestFinalizedBeaconBlock()
	if err != nil {
		return fmt.Errorf("error getting latest finalized Beacon block: %w", err)
	}
	state, err := mgr.GetStateForSlot(block.Slot)
	if err != nil {
		return fmt.Errorf("error getting network state: %w", err)
	}

	details, err := getSnapshotDetails(rp, bn, log, nil)
	if err != nil {
		return fmt.Errorf("error getting snapshot details: %w", err)
	}

	log.Printlnf("Generating a dry-run tree for the current interval (%d)", details.index)

	elBlockIndex := details.snapshotElBlockHeader.Number.Uint64()

	// Log
	log.Printlnf("Snapshot Beacon block = %d, EL block = %d, running from %s to %s\n", details.snapshotBeaconBlock, elBlockIndex, details.startTime, details.endTime)

	// Generate the rewards file
	treegen, err := rprewards.NewTreeGenerator(log, "", rp, cfg, bn, details.index, details.startTime, details.endTime, details.snapshotBeaconBlock, details.snapshotElBlockHeader, details.intervalsPassed, state)
	if err != nil {
		return fmt.Errorf("error creating tree generator: %w", err)
	}
	var rewardsFile *rprewards.RewardsFile
	if ruleset == 0 {
		rewardsFile, err = treegen.GenerateTree()
	} else {
		rewardsFile, err = treegen.GenerateTreeWithRuleset(ruleset)
	}
	if err != nil {
		return fmt.Errorf("error generating Merkle tree: %w", err)
	}
	for address, network := range rewardsFile.InvalidNetworkNodes {
		log.Printlnf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.\n", address.Hex(), network)
	}

	// Get the output paths
	rewardsTreePath := filepath.Join(outputDir, fmt.Sprintf(config.RewardsTreeFilenameFormat, string(cfg.Smartnode.Network.Value.(cfgtypes.Network)), details.index))
	minipoolPerformancePath := filepath.Join(outputDir, fmt.Sprintf(config.MinipoolPerformanceFilenameFormat, string(cfg.Smartnode.Network.Value.(cfgtypes.Network)), details.index))

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
	log.Printlnf("Successfully generated rewards snapshot for interval %d.\n", details.index)

	return nil
}

// Approximates the rETH stakers' share of the Smoothing Pool's current balance
func approximateCurrentRethSpRewards(log log.ColorLogger, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bn beacon.Client, outputDir string, prettyPrint bool, ruleset uint64) error {

	// Get a snapshot of the network state
	mgr, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bn, &log)
	if err != nil {
		return fmt.Errorf("error creating network state manager: %w", err)
	}
	block, err := mgr.GetLatestFinalizedBeaconBlock()
	if err != nil {
		return fmt.Errorf("error getting latest finalized Beacon block: %w", err)
	}
	state, err := mgr.GetStateForSlot(block.Slot)
	if err != nil {
		return fmt.Errorf("error getting network state: %w", err)
	}

	details, err := getSnapshotDetails(rp, bn, log, nil)
	if err != nil {
		return fmt.Errorf("error getting snapshot details: %w", err)
	}

	elBlockIndex := details.snapshotElBlockHeader.Number.Uint64()

	// Log
	log.Printlnf("Snapshot Beacon block = %d, EL block = %d, running from %s to %s\n", details.snapshotBeaconBlock, elBlockIndex, details.startTime, details.endTime)

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", nil)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool contract: %w", err)
	}
	smoothingPoolBalance, err := rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, nil)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}

	// Approximate the balance
	treegen, err := rprewards.NewTreeGenerator(log, "", rp, cfg, bn, details.index, details.startTime, details.endTime, details.snapshotBeaconBlock, details.snapshotElBlockHeader, details.intervalsPassed, state)
	if err != nil {
		return fmt.Errorf("error creating tree generator: %w", err)
	}

	var rETHShare *big.Int
	if ruleset == 0 {
		rETHShare, err = treegen.ApproximateStakerShareOfSmoothingPool()
	} else {
		rETHShare, err = treegen.ApproximateStakerShareOfSmoothingPoolWithRuleset(ruleset)
	}
	if err != nil {
		return fmt.Errorf("error approximating rETH stakers' share of the Smoothing Pool: %w", err)
	}
	log.Printlnf("Total ETH in the Smoothing Pool: %s wei (%.6f ETH)", smoothingPoolBalance.String(), eth.WeiToEth(smoothingPoolBalance))
	log.Printlnf("rETH stakers's share:            %s wei (%.6f ETH)", rETHShare.String(), eth.WeiToEth(rETHShare))

	return nil
}

// Recreates an existing tree for a past interval
func generatePastTree(log log.ColorLogger, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bn beacon.Client, index uint64, outputDir string, prettyPrint bool, ruleset uint64) error {

	// Find the event for this interval
	rewardsEvent, err := rprewards.GetRewardSnapshotEvent(rp, cfg, index)
	if err != nil {
		return fmt.Errorf("error getting rewards submission event for interval %d: %w", index, err)
	}
	log.Printlnf("Found rewards submission event: Beacon block %s, execution block %s", rewardsEvent.ConsensusBlock.String(), rewardsEvent.ExecutionBlock.String())

	// Get a snapshot of the network state at that interval
	mgr, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bn, &log)
	if err != nil {
		return fmt.Errorf("error creating network state manager: %w", err)
	}
	state, err := mgr.GetStateForSlot(rewardsEvent.ConsensusBlock.Uint64())
	if err != nil {
		return fmt.Errorf("error getting network state: %w", err)
	}

	// Get the EL block
	elBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), rewardsEvent.ExecutionBlock)
	if err != nil {
		return fmt.Errorf("error getting execution block: %w", err)
	}

	// Generate the rewards file
	start := time.Now()
	treegen, err := rprewards.NewTreeGenerator(log, "", rp, cfg, bn, index, rewardsEvent.IntervalStartTime, rewardsEvent.IntervalEndTime, rewardsEvent.ConsensusBlock.Uint64(), elBlockHeader, rewardsEvent.IntervalsPassed.Uint64(), state)
	if err != nil {
		return fmt.Errorf("error creating tree generator: %w", err)
	}
	var rewardsFile *rprewards.RewardsFile
	if ruleset == 0 {
		rewardsFile, err = treegen.GenerateTree()
	} else {
		rewardsFile, err = treegen.GenerateTreeWithRuleset(ruleset)
	}
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

// Get the finalized slot for the given target epoch, or the latest one if there isn't a target epoch
func getFinalizedSlot(log log.ColorLogger, bn beacon.Client, targetEpoch *uint64) (uint64, uint64, time.Time, error) {

	// Get the config
	eth2Config, err := bn.GetEth2Config()
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("error getting Beacon config: %w", err)
	}
	genesisTime := time.Unix(int64(eth2Config.GenesisTime), 0)

	// Get the target epoch details

	// Get the beacon head
	beaconHead, err := bn.GetBeaconHead()
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("error getting Beacon head: %w", err)
	}

	// Get the latest finalized slot that existed
	finalizedEpoch := beaconHead.FinalizedEpoch

	// Sanity check the target epoch
	if targetEpoch != nil {
		if *targetEpoch > finalizedEpoch {
			return 0, 0, time.Time{}, fmt.Errorf("target epoch %d is not finalized yet; latest finalized epoch is %d", *targetEpoch, finalizedEpoch)
		}
	}

	// Get the target slot
	var finalizedSlot uint64
	if targetEpoch == nil {
		finalizedSlot = (finalizedEpoch+1)*eth2Config.SlotsPerEpoch - 1
	} else {
		finalizedSlot = (*targetEpoch+1)*eth2Config.SlotsPerEpoch - 1
	}

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

// Get the details of the rewards snapshot at the given block
func getSnapshotDetails(rp *rocketpool.RocketPool, bn beacon.Client, log log.ColorLogger, opts *bind.CallOpts) (snapshotDetails, error) {
	// Get the interval index
	indexBig, err := rewards.GetRewardIndex(rp, opts)
	if err != nil {
		return snapshotDetails{}, fmt.Errorf("error getting current reward index: %w", err)
	}
	index := indexBig.Uint64()

	// Get the start time for the current interval, and how long an interval is supposed to take
	startTime, err := rewards.GetClaimIntervalTimeStart(rp, opts)
	if err != nil {
		return snapshotDetails{}, fmt.Errorf("error getting claim interval start time: %w", err)
	}
	intervalTime, err := rewards.GetClaimIntervalTime(rp, opts)
	if err != nil {
		return snapshotDetails{}, fmt.Errorf("error getting claim interval time: %w", err)
	}

	// Get the block header for the target block
	var targetBlockNumber *big.Int
	if opts == nil {
		targetBlockNumber = nil
	} else {
		targetBlockNumber = opts.BlockNumber
	}
	blockHeader, err := rp.Client.HeaderByNumber(context.Background(), targetBlockNumber)
	if err != nil {
		return snapshotDetails{}, fmt.Errorf("error getting latest block header: %w", err)
	}

	// Calculate the intervals passed
	blockTime := time.Unix(int64(blockHeader.Time), 0)
	timeSinceStart := blockTime.Sub(startTime)
	intervalsPassed := timeSinceStart / intervalTime
	endTime := time.Now()

	// Get the latest finalized Beacon block
	snapshotBeaconBlock, elBlockNumber, beaconBlockTime, err := getFinalizedSlot(log, bn, nil)
	if err != nil {
		return snapshotDetails{}, fmt.Errorf("error getting latest finalized slot: %w", err)
	}

	// Get the number of the EL block matching the CL snapshot block
	var snapshotElBlockHeader *types.Header
	if elBlockNumber == 0 {
		// No EL data so the Merge hasn't happened yet, figure out the EL block based on the Epoch ending time
		snapshotElBlockHeader, err = rprewards.GetELBlockHeaderForTime(beaconBlockTime, rp)
		if err != nil {
			return snapshotDetails{}, fmt.Errorf("error getting EL block for time %s: %w", beaconBlockTime, err)
		}
	} else {
		snapshotElBlockHeader, err = rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
		if err != nil {
			return snapshotDetails{}, fmt.Errorf("error getting EL block %d: %w", elBlockNumber, err)
		}
	}

	return snapshotDetails{
		index:                 index,
		startTime:             startTime,
		endTime:               endTime,
		snapshotBeaconBlock:   snapshotBeaconBlock,
		snapshotElBlockHeader: snapshotElBlockHeader,
		intervalsPassed:       uint64(intervalsPassed),
	}, nil
}

func printNetworkInfo(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bn beacon.Client, log log.ColorLogger, opts *bind.CallOpts) error {
	details, err := getSnapshotDetails(rp, bn, log, opts)
	if err != nil {
		return fmt.Errorf("error getting network details for snapshot: %w", err)
	}

	// Get a snapshot of the network state at that interval
	mgr, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bn, &log)
	if err != nil {
		return fmt.Errorf("error creating network state manager: %w", err)
	}
	state, err := mgr.GetStateForSlot(details.snapshotBeaconBlock)
	if err != nil {
		return fmt.Errorf("error getting network state: %w", err)
	}

	generator, err := rprewards.NewTreeGenerator(log, "", rp, cfg, bn, details.index, details.startTime, details.endTime, details.snapshotBeaconBlock, details.snapshotElBlockHeader, details.intervalsPassed, state)
	if err != nil {
		return fmt.Errorf("error creating generator: %w", err)
	}

	log.Println()
	log.Println("=== Network Details ===")
	log.Printlnf("Current index:        %d", details.index)
	log.Printlnf("Start Time:           %s", details.startTime)

	// Find the event for the previous interval
	if details.index > 0 {
		rewardsEvent, err := rprewards.GetRewardSnapshotEvent(rp, cfg, details.index-1)
		if err != nil {
			return fmt.Errorf("error getting rewards submission event for previous interval (%d): %w", details.index-1, err)
		}
		log.Printlnf("Start Beacon Slot:    %d", rewardsEvent.ConsensusBlock.Uint64()+1)
		log.Printlnf("Start EL Block:       %d", rewardsEvent.ExecutionBlock.Uint64()+1)
	}

	log.Printlnf("End Time:             %s", details.endTime)
	log.Printlnf("Snapshot Beacon Slot: %d", details.snapshotBeaconBlock)
	log.Printlnf("Snapshot EL Block:    %s", details.snapshotElBlockHeader.Number.String())
	log.Printlnf("Intervals Passed:     %d", details.intervalsPassed)
	log.Printlnf("Tree Ruleset:         v%d", generator.GetGeneratorRulesetVersion())
	log.Printlnf("Approximator Ruleset: v%d", generator.GetApproximatorRulesetVersion())

	return nil
}

// Configure HTTP transport settings
func configureHTTP() {

	// The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
	// The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
	// This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}
