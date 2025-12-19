package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-json"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
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

// Details about the snapshot block / timestamp for a treegen target
type snapshotDetails struct {
	index                 uint64
	startTime             time.Time
	endTime               time.Time
	startSlot             uint64
	snapshotBeaconBlock   uint64
	snapshotElBlockHeader *types.Header
	intervalsPassed       uint64
}

// targets holds information about the span we wish to gather data for.
// It could be an entire interval, or it could be a portion of one.
type targets struct {
	// If generating a whole interval, we use the rewardsEvent
	rewardsEvent *rewards.RewardsEvent

	// For preview related functions, we use snapshotDetails
	snapshotDetails *snapshotDetails

	// Cached beacon block - not necessarily the last block in the last epoch,
	// as it may not have been proposed
	block *beacon.BeaconBlock
}

// Arguments that will be passed down to the actual treegen routine
type treegenArguments struct {
	// Header of the starting EL block
	elBlockHeader *types.Header

	// Number of intervals elapsed
	intervalsPassed uint64

	// End time for the rewards tree (may not align on a full interval)
	endTime time.Time

	// Start time for the rewards tree
	startTime time.Time

	// Index of the rewards period
	index uint64

	// The first slot in the period
	startSlot uint64

	// Consensus end block
	block *beacon.BeaconBlock

	// Network State at end EL block
	state *state.NetworkState
}

// Treegen holder for the requested execution metadata and necessary artifacts
type treeGenerator struct {
	log                 *log.ColorLogger
	errLog              *log.ColorLogger
	rp                  rprewards.RewardsExecutionClient
	rpNative            *rocketpool.RocketPool
	cfg                 *config.RocketPoolConfig
	mgr                 *state.NetworkStateManager
	bn                  beacon.Client
	beaconConfig        beacon.Eth2Config
	targets             targets
	outputDir           string
	prettyPrint         bool
	ruleset             uint64
	generateVotingPower bool
}

// Generates a new rewards tree based on the command line flags
func GenerateTree(c *cli.Context) error {
	// Configure
	configureHTTP()

	// Initialization
	interval := c.Int64("interval")
	targetEpoch := c.Uint64("target-epoch")
	logger := log.NewColorLogger(color.FgHiWhite)
	errLogger := log.NewColorLogger(color.FgRed)

	// URL acquisition
	ecUrl := c.String("ec-endpoint")
	if ecUrl == "" {
		return fmt.Errorf("ec-endpoint must be provided")
	}
	bnUrl := c.String("bn-endpoint")
	if ecUrl == "" {
		return fmt.Errorf("bn-endpoint must be provided")
	}

	// Create the EC and BN clients
	ec, err := services.NewEthClient(ecUrl)
	if err != nil {
		return fmt.Errorf("error connecting to the EC: %w", err)
	}
	bn := client.NewStandardHttpClient(bnUrl)
	beaconConfig, err := bn.GetEth2Config()
	if err != nil {
		return fmt.Errorf("error getting beacon config from the BN at %s - %w", bnUrl, err)
	}

	// Check which network we're on via the BN
	depositContract, err := bn.GetEth2DepositContract()
	if err != nil {
		return fmt.Errorf("error getting deposit contract from the BN: %w", err)
	}
	var network cfgtypes.Network
	switch depositContract.ChainID {
	case 1:
		network = cfgtypes.Network_Mainnet
		logger.Printlnf("Beacon node is configured for Mainnet.")
	case 560048:
		network = cfgtypes.Network_Testnet
		logger.Printlnf("Beacon node is configured for Testnet.")
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

	// Create the NetworkStateManager
	mgr := state.NewNetworkStateManager(rp, cfg.Smartnode.GetStateManagerContracts(), bn, &logger)

	// Create the generator
	generator := treeGenerator{
		log:                 &logger,
		errLog:              &errLogger,
		rp:                  rprewards.NewRewardsExecutionClient(rp),
		rpNative:            rp,
		cfg:                 cfg,
		bn:                  bn,
		mgr:                 mgr,
		beaconConfig:        beaconConfig,
		outputDir:           c.String("output-dir"),
		prettyPrint:         c.Bool("pretty-print"),
		ruleset:             c.Uint64("ruleset"),
		generateVotingPower: c.Bool("generate-voting-power"),
	}

	// initialize the generator targets
	if err := generator.setTargets(interval, targetEpoch); err != nil {
		return fmt.Errorf("error setting the targeted consensus epoch and block: %w", err)
	}

	// Run the tree generation or the rETH SP approximation
	if c.Bool("approximate-only") {
		return generator.approximateRethSpRewards()
	}

	// Print the network info and exit if requested
	if c.Bool("network-info") {
		return generator.printNetworkInfo()
	}

	return generator.generateTree()
}

func (g *treeGenerator) getTreegenArgs() (*treegenArguments, error) {

	// Cache the network state at the time of the targeted epoch for later use
	state, err := g.mgr.GetStateForSlot(g.targets.block.Slot)
	if err != nil {
		return nil, fmt.Errorf("unable to get state at slot %d: %w", g.targets.block.Slot, err)
	}

	// If we have a rewardsEvent, we're generating a full interval
	if g.targets.rewardsEvent != nil {
		index := g.targets.rewardsEvent.Index.Uint64()
		startSlot := uint64(0)
		if index > 0 {
			// Get the start slot for this interval
			previousRewardsEvent, err := g.rp.GetRewardSnapshotEvent(
				g.cfg.Smartnode.GetPreviousRewardsPoolAddresses(),
				uint64(index-1),
				nil,
			)
			if err != nil {
				return nil, fmt.Errorf("error getting event for interval %d: %w", index-1, err)
			}
			startSlot, err = getStartSlotForInterval(previousRewardsEvent, g.bn, g.beaconConfig)
			if err != nil {
				return nil, fmt.Errorf("error getting start slot for interval %d: %w", index, err)
			}
		}

		elBlockHeader, err := g.rp.HeaderByNumber(context.Background(), g.targets.rewardsEvent.ExecutionBlock)
		if err != nil {
			return nil, fmt.Errorf("error getting el block header %d: %w", g.targets.rewardsEvent.ExecutionBlock.Uint64(), err)
		}
		return &treegenArguments{
			startTime:       g.targets.rewardsEvent.IntervalStartTime,
			endTime:         g.targets.rewardsEvent.IntervalEndTime,
			index:           index,
			intervalsPassed: g.targets.rewardsEvent.IntervalsPassed.Uint64(),
			startSlot:       startSlot,
			block:           g.targets.block,
			elBlockHeader:   elBlockHeader,
			state:           state,
		}, nil
	}

	// Partial interval
	return &treegenArguments{
		startTime:       g.targets.snapshotDetails.startTime,
		endTime:         g.targets.snapshotDetails.endTime,
		index:           g.targets.snapshotDetails.index,
		intervalsPassed: g.targets.snapshotDetails.intervalsPassed,
		startSlot:       g.targets.snapshotDetails.startSlot,
		block:           g.targets.block,
		elBlockHeader:   g.targets.snapshotDetails.snapshotElBlockHeader,
		state:           state,
	}, nil
}

func (d *snapshotDetails) log(l *log.ColorLogger) {
	l.Printlnf("Snapshot Beacon block = %d, EL block = %d, running from %s to %s\n",
		d.snapshotBeaconBlock,
		d.snapshotElBlockHeader.Number.Uint64(),
		d.startTime,
		d.endTime)
}

func (g *treeGenerator) lastBlockInEpoch(epoch uint64) (*beacon.BeaconBlock, error) {

	// Get the last block proposed in the targeted epoch.
	// If the targeted epoch has no proposals, return nil, nil
	end := epoch * g.beaconConfig.SlotsPerEpoch
	start := end + g.beaconConfig.SlotsPerEpoch - 1
	for slot := start; slot >= end; slot-- {
		block, exists, err := g.bn.GetBeaconBlock(fmt.Sprint(slot))
		if err != nil {
			return nil, err
		}

		if exists {
			// We found it, so set it and exit
			return &block, nil
		}
	}

	return nil, nil
}

func (g *treeGenerator) setTargets(interval int64, targetEpoch uint64) error {
	var err error

	// Validate that the target epoch is finalized
	if targetEpoch > 0 {
		beaconHead, err := g.bn.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("unable to query beacon head: %w", err)
		}

		if targetEpoch > beaconHead.FinalizedEpoch {
			return fmt.Errorf("targeted epoch has not yet been finalized")
		}
	}

	// If interval isn't set, we're generating a preview of the current interval
	if interval < 0 {
		var block *beacon.BeaconBlock
		if targetEpoch == 0 {
			// No targetEpoch was passed, so set it to the latest finalized epoch
			b, err := g.mgr.GetLatestFinalizedBeaconBlock()
			if err != nil {
				return err
			}

			block = &b
		} else {
			// A target epoch was passed, so find its last block
			block, err = g.lastBlockInEpoch(targetEpoch)
			if err != nil {
				return err
			}
			if block == nil {
				return fmt.Errorf("Unable to find any valid blocks in epoch %d. Was your BN checkpoint synced against a slot that occurred after this epoch?", targetEpoch)
			}
		}

		g.targets.block = block
		g.targets.snapshotDetails, err = g.getSnapshotDetails()
		if err != nil {
			return err
		}

		// Ensure the target block is in the current interval
		if g.slotToTime(block.Slot).Before(g.targets.snapshotDetails.startTime) {
			return fmt.Errorf("selected epoch precedes current interval. use -i to generate previous intervals")
		}

		// Inform the user of the range they're querying
		g.log.Printlnf("Targeting a portion of the current interval (%d)", g.targets.snapshotDetails.index)
		g.targets.snapshotDetails.log(g.log)

		return nil
	}

	// We're generating a previous interval (full or partial)
	// Get the corresponding rewards event for that interval
	rewardsEvent, err := g.rp.GetRewardSnapshotEvent(
		g.cfg.Smartnode.GetPreviousRewardsPoolAddresses(),
		uint64(interval),
		nil,
	)
	if err != nil {
		return err
	}

	// If targetEpoch isn't set, we're generating a full interval
	if targetEpoch == 0 {
		g.log.Printlnf("Targeting full interval %d", interval)
		g.targets.rewardsEvent = &rewardsEvent

		// Cache the last block of the rewards period
		g.targets.block, err = g.lastBlockInEpoch(rewardsEvent.ConsensusBlock.Uint64() / g.beaconConfig.SlotsPerEpoch)
		if err != nil {
			return err
		}
		if g.targets.block == nil {
			return fmt.Errorf("unable to find any valid blocks in epoch %d. Was your BN checkpoint synced against a slot that occurred after this epoch?", targetEpoch)
		}

		return nil
	}

	// We're generating a partial interval
	// Ensure the target slot happens *before* the end of the interval
	eventBlock := rewardsEvent.ConsensusBlock.Uint64()
	finalEpochOfInterval := eventBlock / g.beaconConfig.SlotsPerEpoch
	if targetEpoch == finalEpochOfInterval {
		return fmt.Errorf("target epoch %d was the end of the targeted interval %d.\nRerun without -t", targetEpoch, interval)
	}
	if targetEpoch > finalEpochOfInterval {
		return fmt.Errorf("target epoch %d was after targeted interval %d", targetEpoch, interval)
	}

	// Ensure the target epoch started *after* the start of the interval, which should land on the start of an epoch boundary
	epochStartTime := g.slotToTime(targetEpoch * g.beaconConfig.SlotsPerEpoch)
	if epochStartTime.Before(rewardsEvent.IntervalStartTime) {
		return fmt.Errorf("target epoch %d was before targeted interval %d", targetEpoch, interval)
	}

	// Cache the target block for later use
	block, found, err := g.bn.GetBeaconBlock(fmt.Sprint(eventBlock))
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Unable to find the ending block for interval %d (slot %d). Was your BN checkpoint synced against a slot that occurred after this epoch?", interval, eventBlock)
	}
	g.targets.block = &block

	g.targets.snapshotDetails, err = g.getSnapshotDetails()
	if err != nil {
		return err
	}

	// Inform the user of the range they're querying
	g.log.Printlnf("Targeting a portion of a previous interval (%d)", g.targets.snapshotDetails.index)
	g.targets.snapshotDetails.log(g.log)

	return nil
}

// Gets the timestamp for a Beacon slot
func (g *treeGenerator) slotToTime(slot uint64) time.Time {
	genesisTime := time.Unix(int64(g.beaconConfig.GenesisTime), 0)
	secondsForSlot := time.Duration(slot*g.beaconConfig.SecondsPerSlot) * time.Second
	return genesisTime.Add(secondsForSlot)
}

// Generates the rewards file for the given generator
func (g *treeGenerator) generateRewardsFile(treegen *rprewards.TreeGenerator) (*rprewards.GenerateTreeResult, error) {
	if g.ruleset == 0 {
		return treegen.GenerateTree()
	}

	return treegen.GenerateTreeWithRuleset(g.ruleset)
}

func (g *treeGenerator) serializeVotingPower(votingPowerFile *VotingPowerFile) ([]byte, error) {
	if g.prettyPrint {
		return json.MarshalIndent(votingPowerFile, "", "\t")
	}

	return json.Marshal(votingPowerFile)
}

// Serializes the minipool performance file into JSON
func (g *treeGenerator) serializeMinipoolPerformance(result *rprewards.GenerateTreeResult) ([]byte, error) {
	perfFile := result.MinipoolPerformanceFile

	if g.prettyPrint {
		return perfFile.SerializeHuman()
	}

	return perfFile.Serialize()
}

// Serializes the rewards tree file in to JSON
func (g *treeGenerator) serializeRewardsTree(rewardsFile rprewards.IRewardsFile) ([]byte, error) {
	if g.prettyPrint {
		return json.MarshalIndent(rewardsFile, "", "\t")
	}

	return json.Marshal(rewardsFile)
}

// Writes both the performance file and the rewards file to disk
func (g *treeGenerator) writeFiles(result *rprewards.GenerateTreeResult, votingPowerFile *VotingPowerFile) error {
	g.log.Printlnf("Saving JSON files...")
	rewardsFile := result.RewardsFile
	index := rewardsFile.GetIndex()

	// Get the output paths
	rewardsTreePath := filepath.Join(
		g.outputDir,
		g.cfg.Smartnode.GetRewardsTreeFilename(index, config.RewardsExtensionJSON),
	)
	var performancePath string
	if g.ruleset < 11 {
		performancePath = filepath.Join(
			g.outputDir,
			g.cfg.Smartnode.GetMinipoolPerformanceFilename(index),
		)
	} else {
		performancePath = filepath.Join(
			g.outputDir,
			g.cfg.Smartnode.GetPerformanceFilename(index),
		)
	}

	// Serialize the minipool performance file
	performanceBytes, err := g.serializeMinipoolPerformance(result)
	if err != nil {
		return fmt.Errorf("error serializing minipool performance file into JSON: %w", err)
	}

	// Write it to disk
	err = os.WriteFile(performancePath, performanceBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving minipool performance file to %s: %w", performancePath, err)
	}

	g.log.Printlnf("Saved minipool performance file to %s", performancePath)
	rewardsFile.SetMinipoolPerformanceFileCID("---")

	// Serialize the rewards tree to JSON
	wrapperBytes, err := g.serializeRewardsTree(rewardsFile)
	if err != nil {
		return fmt.Errorf("error serializing proof wrapper into JSON: %w", err)
	}
	g.log.Printlnf("Generation complete! Saving tree...")

	// Write the rewards tree to disk
	err = os.WriteFile(rewardsTreePath, wrapperBytes, 0644)
	if err != nil {
		return fmt.Errorf("error saving rewards tree file to %s: %w", rewardsTreePath, err)
	}

	g.log.Printlnf("Saved rewards snapshot file to %s", rewardsTreePath)

	// Write the voting power file to disk
	if votingPowerFile != nil {
		votingPowerFileBytes, err := g.serializeVotingPower(votingPowerFile)
		if err != nil {
			return fmt.Errorf("error serializing voting power file into JSON: %w", err)
		}

		votingFilePath := filepath.Join(g.outputDir, fmt.Sprintf("rp-voting-power-%s-%d.json", string(g.cfg.Smartnode.Network.Value.(cfgtypes.Network)), votingPowerFile.ConsensusSlot))
		err = os.WriteFile(votingFilePath, votingPowerFileBytes, 0644)
		if err != nil {
			return fmt.Errorf("error saving voting power file to %s: %w", votingFilePath, err)
		}
		g.log.Printlnf("Saved voting power file to %s", votingFilePath)
	}
	g.log.Printlnf("Successfully generated rewards snapshot for interval %d", index)

	return nil
}

// Creates a tree generator using the provided arguments
func (g *treeGenerator) getGenerator(args *treegenArguments) (*rprewards.TreeGenerator, error) {

	// Create the tree generator
	out, err := rprewards.NewTreeGenerator(
		g.log, "", g.rp, g.cfg, g.bn, args.index,
		args.startTime, args.endTime,
		&rprewards.SnapshotEnd{
			Slot:           args.state.BeaconConfig.FirstSlotAtLeast(args.endTime.Unix()),
			ConsensusBlock: args.block.Slot,
			ExecutionBlock: args.elBlockHeader.Number.Uint64(),
		}, args.elBlockHeader,
		args.intervalsPassed, args.state)
	if err != nil {
		return nil, fmt.Errorf("error creating tree generator: %w", err)
	}

	return out, nil
}

// Approximates the rETH stakers' share of the Smoothing Pool's current balance
func (g *treeGenerator) approximateRethSpRewards() error {
	args, err := g.getTreegenArgs()
	if err != nil {
		return fmt.Errorf("error compiling treegen arguments: %w", err)
	}

	opts := &bind.CallOpts{
		BlockNumber: args.elBlockHeader.Number,
	}

	// Log
	g.log.Printlnf("Approximating rETH rewards for the current interval (%d)", args.index)
	g.log.Printlnf("Snapshot Beacon block = %d, EL block = %d, running from %s to %s\n",
		args.block.Slot, opts.BlockNumber.Uint64(), args.startTime, args.endTime)

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := g.rpNative.GetContract("rocketSmoothingPool", opts)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool contract: %w", err)
	}
	smoothingPoolBalance, err := g.rpNative.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, opts.BlockNumber)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}

	// Create the tree generator
	treegen, err := g.getGenerator(args)
	if err != nil {
		return err
	}

	// Approximate the balance
	var rETHShare *big.Int
	if g.ruleset == 0 {
		rETHShare, err = treegen.ApproximateStakerShareOfSmoothingPool()
	} else {
		rETHShare, err = treegen.ApproximateStakerShareOfSmoothingPoolWithRuleset(g.ruleset)
	}
	if err != nil {
		return fmt.Errorf("error approximating rETH stakers' share of the Smoothing Pool: %w", err)
	}
	g.log.Printlnf("Total ETH in the Smoothing Pool: %s wei (%.6f ETH)", smoothingPoolBalance.String(), eth.WeiToEth(smoothingPoolBalance))
	g.log.Printlnf("rETH stakers's share:            %s wei (%.6f ETH)", rETHShare.String(), eth.WeiToEth(rETHShare))

	return nil
}

// Generate a complete rewards tree
func (g *treeGenerator) generateTree() error {
	var votingPowerFile *VotingPowerFile
	args, err := g.getTreegenArgs()
	if err != nil {
		return fmt.Errorf("error compiling treegen arguments: %w", err)
	}

	// Create the tree generator
	treegen, err := g.getGenerator(args)
	if err != nil {
		return err
	}

	// If a voting power file was requested, generate it now.
	if g.generateVotingPower {
		votingPowerFile = g.GenerateVotingPower(args.state)
		votingPowerFile.Time = args.endTime
	}

	// Generate the rewards file
	start := time.Now()
	result, err := g.generateRewardsFile(treegen)
	if err != nil {
		return fmt.Errorf("error generating Merkle tree: %w", err)
	}

	for address, network := range result.InvalidNetworkNodes {
		g.log.Printlnf("WARNING: Node %s has invalid network %d assigned! Using 0 (mainnet) instead.", address.Hex(), network)
	}
	g.log.Printlnf("Finished in %s", time.Since(start).String())

	// Validate the Merkle root
	if g.targets.rewardsEvent != nil {
		rootString := result.RewardsFile.GetMerkleRoot()
		root := common.HexToHash(rootString)
		if !bytes.Equal(root[:], g.targets.rewardsEvent.MerkleRoot[:]) {
			g.log.Printlnf("WARNING: your Merkle tree had a root of %s, but the canonical Merkle tree's root was %s. This file will not be usable for claiming rewards.", root.Hex(), g.targets.rewardsEvent.MerkleRoot.Hex())
		} else {
			g.log.Printlnf("Your Merkle tree's root of %s matches the canonical root! You will be able to use this file for claiming rewards.", rootString)
		}
	}

	err = g.writeFiles(result, votingPowerFile)
	if err != nil {
		return err
	}

	return nil

}

// Create a rewards snapshot at the target block
func (g *treeGenerator) getSnapshotDetails() (*snapshotDetails, error) {
	var err error
	var opts bind.CallOpts

	endTime := g.slotToTime(g.targets.block.Slot)

	// Get the number of the EL block matching the CL snapshot block
	var snapshotElBlockHeader *types.Header
	if g.targets.block.ExecutionBlockNumber == 0 {
		return nil, fmt.Errorf("slot %d was pre-merge", g.targets.block.Slot)
	}
	opts.BlockNumber = big.NewInt(0).SetUint64(g.targets.block.ExecutionBlockNumber)
	snapshotElBlockHeader, err = g.rp.HeaderByNumber(context.Background(), opts.BlockNumber)
	if err != nil {
		return nil, fmt.Errorf("error getting EL block %d: %w", opts.BlockNumber.Uint64(), err)
	}

	// Get the interval index
	indexBig, err := g.rpNative.GetRewardIndex(&opts)
	if err != nil {
		return nil, fmt.Errorf("error getting current reward index: %w", err)
	}
	index := indexBig.Uint64()

	// Get the start slot
	startSlot := uint64(0)
	if index > 0 {
		// Get the start slot for this interval
		previousRewardsEvent, err := g.rp.GetRewardSnapshotEvent(
			g.cfg.Smartnode.GetPreviousRewardsPoolAddresses(),
			uint64(index-1),
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting event for interval %d: %w", index-1, err)
		}
		startSlot, err = getStartSlotForInterval(previousRewardsEvent, g.bn, g.beaconConfig)
		if err != nil {
			return nil, fmt.Errorf("error getting start slot for interval %d: %w", index, err)
		}
	}

	// Get the start time for the interval, and how long an interval is supposed to take
	startTime, err := rewards.GetClaimIntervalTimeStart(g.rpNative, &opts)
	if err != nil {
		return nil, fmt.Errorf("error getting claim interval start time: %w", err)
	}
	intervalTime, err := rewards.GetClaimIntervalTime(g.rpNative, &opts)
	if err != nil {
		return nil, fmt.Errorf("error getting claim interval time: %w", err)
	}

	// Calculate the intervals passed
	blockTime := time.Unix(int64(snapshotElBlockHeader.Time), 0)
	timeSinceStart := blockTime.Sub(startTime)
	intervalsPassed := uint64(timeSinceStart / intervalTime)

	return &snapshotDetails{
		index:                 index,
		startTime:             startTime,
		endTime:               endTime,
		startSlot:             startSlot,
		snapshotBeaconBlock:   g.targets.block.Slot,
		snapshotElBlockHeader: snapshotElBlockHeader,
		intervalsPassed:       intervalsPassed,
	}, nil
}

// Gets the start slot for the given interval
func getStartSlotForInterval(previousIntervalEvent rewards.RewardsEvent, bc beacon.Client, beaconConfig beacon.Eth2Config) (uint64, error) {
	// Sanity check to confirm the BN can access the block from the previous interval
	_, exists, err := bc.GetBeaconBlock(previousIntervalEvent.ConsensusBlock.String())
	if err != nil {
		return 0, fmt.Errorf("error verifying block from previous interval: %w", err)
	}
	if !exists {
		return 0, fmt.Errorf("couldn't retrieve CL block from previous interval (slot %d); this likely means you checkpoint sync'd your Beacon Node and it has not backfilled to the previous interval yet so it cannot be used for tree generation", previousIntervalEvent.ConsensusBlock.Uint64())
	}

	previousEpoch := previousIntervalEvent.ConsensusBlock.Uint64() / beaconConfig.SlotsPerEpoch
	nextEpoch := previousEpoch + 1
	consensusStartBlock := nextEpoch * beaconConfig.SlotsPerEpoch

	// Get the first block that isn't missing
	for {
		_, exists, err := bc.GetBeaconBlock(fmt.Sprint(consensusStartBlock))
		if err != nil {
			return 0, fmt.Errorf("error getting EL data for BC slot %d: %w", consensusStartBlock, err)
		}
		if !exists {
			consensusStartBlock++
		} else {
			break
		}
	}

	return consensusStartBlock, nil
}

// Print information about the current network and interval info
func (g *treeGenerator) printNetworkInfo() error {
	args, err := g.getTreegenArgs()
	if err != nil {
		return fmt.Errorf("error compiling treegen arguments: %w", err)
	}

	// Generate the rewards file
	generator, err := g.getGenerator(args)
	if err != nil {
		return err
	}

	g.log.Println()
	g.log.Println("=== Network Details ===")
	g.log.Printlnf("Current index:        %d", args.index)
	g.log.Printlnf("Start Time:           %s", args.startTime)

	// Find the event for the previous interval
	if args.index > 0 {
		rewardsEvent, err := g.rp.GetRewardSnapshotEvent(
			g.cfg.Smartnode.GetPreviousRewardsPoolAddresses(),
			args.index-1,
			nil,
		)
		if err != nil {
			return fmt.Errorf("error getting rewards submission event for previous interval (%d): %w", args.index-1, err)
		}
		g.log.Printlnf("Start Beacon Slot:    %d", rewardsEvent.ConsensusBlock.Uint64()+1)
		g.log.Printlnf("Start EL Block:       %d", rewardsEvent.ExecutionBlock.Uint64()+1)
	}

	g.log.Printlnf("End Time:             %s", args.endTime)
	g.log.Printlnf("Snapshot Beacon Slot: %d", args.block.Slot)
	g.log.Printlnf("Snapshot EL Block:    %s", args.elBlockHeader.Number.String())
	g.log.Printlnf("Intervals Passed:     %d", args.intervalsPassed)
	g.log.Printlnf("Tree Ruleset:         v%d", generator.GetGeneratorRulesetVersion())
	g.log.Printlnf("Approximator Ruleset: v%d", generator.GetApproximatorRulesetVersion())

	return nil
}

// Configure HTTP transport settings
func configureHTTP() {

	// The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
	// The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
	// This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}
