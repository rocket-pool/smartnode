package watchtower

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rprewards "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/rewards"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

// Generate rewards Merkle Tree task
type GenerateRewardsTree struct {
	ctx       context.Context
	sp        *services.ServiceProvider
	logger    *slog.Logger
	cfg       *config.SmartNodeConfig
	rp        *rocketpool.RocketPool
	ec        eth.IExecutionClient
	bc        beacon.IBeaconClient
	lock      *sync.Mutex
	isRunning bool
}

// Create generate rewards Merkle Tree task
func NewGenerateRewardsTree(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger) *GenerateRewardsTree {
	lock := &sync.Mutex{}
	return &GenerateRewardsTree{
		ctx:       ctx,
		sp:        sp,
		logger:    logger.With(slog.String(keys.RoutineKey, "Generate Rewards Tree")),
		cfg:       sp.GetConfig(),
		rp:        sp.GetRocketPool(),
		ec:        sp.GetEthClient(),
		bc:        sp.GetBeaconClient(),
		lock:      lock,
		isRunning: false,
	}
}

// Check for generation requests
func (t *GenerateRewardsTree) Run() error {
	t.logger.Info("Starting manual rewards tree generation request check.")

	// Check if rewards generation is already running
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Tree generation is already running.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Check for requests
	requestDir := t.cfg.GetWatchtowerFolder()
	files, err := os.ReadDir(requestDir)
	if os.IsNotExist(err) {
		t.logger.Info("Watchtower storage directory doesn't exist, creating...")
		err = os.Mkdir(requestDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating watchtower storage directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error enumerating files in watchtower storage directory: %w", err)
	}

	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, config.RegenerateRewardsTreeRequestSuffix) && !file.IsDir() {
			// Get the index
			indexString := strings.TrimSuffix(filename, config.RegenerateRewardsTreeRequestSuffix)
			index, err := strconv.ParseUint(indexString, 0, 64)
			if err != nil {
				return fmt.Errorf("error parsing index from [%s]: %w", filename, err)
			}

			// Delete the file
			path := filepath.Join(requestDir, filename)
			err = os.Remove(path)
			if err != nil {
				return fmt.Errorf("error removing request file [%s]: %w", path, err)
			}

			// Generate the rewards tree
			t.lock.Lock()
			t.isRunning = true
			t.lock.Unlock()
			go t.generateRewardsTree(index)

			// Return after the first request, do others at other intervals
			return nil
		}
	}

	return nil
}

func (t *GenerateRewardsTree) generateRewardsTree(index uint64) {
	// Begin generation of the tree
	logger := t.logger.With(slog.Uint64(keys.IntervalKey, index))
	logger.Info("Starting generation of Merkle rewards tree.")

	// Find the event for this interval
	rewardsEvent, err := rprewards.GetRewardSnapshotEvent(t.rp, t.cfg, index, nil)
	if err != nil {
		t.handleError(fmt.Errorf("Error getting event: %w", err), logger)
		return
	}
	logger.Info("Found snapshot event", slog.Uint64(keys.SlotKey, rewardsEvent.ConsensusBlock.Uint64()), slog.Uint64(keys.BlockKey, rewardsEvent.ExecutionBlock.Uint64()))

	// Get the EL block
	elBlockHeader, err := t.ec.HeaderByNumber(context.Background(), rewardsEvent.ExecutionBlock)
	if err != nil {
		t.handleError(fmt.Errorf("Error getting execution block: %w", err), logger)
		return
	}

	var stateManager *state.NetworkStateManager

	// Try getting the rETH address as a canary to see if the block is available
	rs := t.cfg.GetRocketPoolResources()
	client := t.rp
	opts := &bind.CallOpts{
		BlockNumber: elBlockHeader.Number,
	}
	var address common.Address
	err = client.Query(func(mc *batch.MultiCaller) error {
		client.Storage.GetAddress(mc, &address, string(rocketpool.ContractName_RocketTokenRETH))
		return nil
	}, opts)
	if err == nil {
		// Create the state manager with using the primary or fallback (not necessarily archive) EC
		stateManager, err = state.NewNetworkStateManager(t.ctx, client, t.cfg, t.rp.Client, t.bc, logger)
		if err != nil {
			t.handleError(fmt.Errorf("error creating new NetworkStateManager with Archive EC: %w", err), logger)
			return
		}
	} else {
		// Check if an Archive EC is provided, and if using it would potentially resolve the error
		errMessage := err.Error()
		logger.Warn("Error getting network state", slog.Uint64(keys.BlockKey, elBlockHeader.Number.Uint64()), slog.String(log.ErrorKey, errMessage))
		if strings.Contains(errMessage, "missing trie node") || // Geth
			strings.Contains(errMessage, "No state available for block") || // Nethermind
			strings.Contains(errMessage, "Internal error") { // Besu
			// TODO add Reth string

			// The state was missing so fall back to the archive node
			archiveEcUrl := t.cfg.ArchiveEcUrl.Value
			if archiveEcUrl != "" {
				logger.Warn("Primary EC cannot retrieve state for historical block, using archive EC", slog.Uint64(keys.BlockKey, elBlockHeader.Number.Uint64()), slog.String(keys.ArchiveEcKey, archiveEcUrl))
				ec, err := ethclient.Dial(archiveEcUrl)
				if err != nil {
					t.handleError(fmt.Errorf("error connecting to archive EC: %w", err), logger)
					return
				}
				client, err = rocketpool.NewRocketPool(ec, rs.StorageAddress, rs.MulticallAddress, rs.BalanceBatcherAddress)
				if err != nil {
					t.handleError(fmt.Errorf("error creating Rocket Pool client connected to archive EC: %w", err), logger)
					return
				}

				// Get the rETH address from the archive EC
				err = client.Query(func(mc *batch.MultiCaller) error {
					client.Storage.GetAddress(mc, &address, string(rocketpool.ContractName_RocketTokenRETH))
					return nil
				}, opts)
				if err != nil {
					t.handleError(fmt.Errorf("error verifying rETH address with Archive EC: %w", err), logger)
					return
				}
				// Create the state manager with the archive EC
				stateManager, err = state.NewNetworkStateManager(t.ctx, client, t.cfg, ec, t.bc, logger)
				if err != nil {
					t.handleError(fmt.Errorf("error creating new NetworkStateManager with Archive EC: %w", err), logger)
					return
				}
			} else {
				// No archive node specified
				t.handleError(fmt.Errorf("Primary EC cannot retrieve state for historical block %d and the Archive EC is not specified", elBlockHeader.Number.Uint64()), logger)
				return
			}

		}
	}

	// Sanity check the rETH address to make sure the client is working right
	if address != rs.RethAddress {
		t.handleError(fmt.Errorf("your Primary EC provided %s as the rETH address, but it should have been %s", address.Hex(), rs.RethAddress.Hex()), logger)
		return
	}

	// Get the state for the target slot
	state, err := stateManager.GetStateForSlot(t.ctx, rewardsEvent.ConsensusBlock.Uint64())
	if err != nil {
		t.handleError(fmt.Errorf("error getting state for beacon slot %d: %w", rewardsEvent.ConsensusBlock.Uint64(), err), logger)
		return
	}

	// Generate the tree
	t.generateRewardsTreeImpl(logger, client, index, rewardsEvent, elBlockHeader, state)
}

// Implementation for rewards tree generation using a viable EC
func (t *GenerateRewardsTree) generateRewardsTreeImpl(logger *slog.Logger, rp *rocketpool.RocketPool, index uint64, rewardsEvent rewards.RewardsEvent, elBlockHeader *types.Header, state *state.NetworkState) {
	// Generate the rewards file
	start := time.Now()
	treegen, err := rprewards.NewTreeGenerator(t.logger, rp, t.cfg, t.bc, index, rewardsEvent.IntervalStartTime, rewardsEvent.IntervalEndTime, rewardsEvent.ConsensusBlock.Uint64(), elBlockHeader, rewardsEvent.IntervalsPassed.Uint64(), state, nil)
	if err != nil {
		t.handleError(fmt.Errorf("Error creating Merkle tree generator: %w", err), logger)
		return
	}
	rewardsFile, err := treegen.GenerateTree(t.ctx)
	if err != nil {
		t.handleError(fmt.Errorf("%s Error generating Merkle tree: %w", err), logger)
		return
	}
	header := rewardsFile.GetHeader()
	for address, network := range header.InvalidNetworkNodes {
		logger.Warn("Node has invalid network assigned! Using 0 (mainnet) instead.", slog.String(keys.NodeKey, address.Hex()), slog.Uint64(keys.NetworkKey, network))
	}
	logger.Info("Finished generation.", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Validate the Merkle root
	root := common.BytesToHash(header.MerkleTree.Root())
	if root != rewardsEvent.MerkleRoot {
		logger.Warn("Your Merkle tree's root differed from the canonical Merkle tree's root. This file will not be usable for claiming rewards.", slog.String(keys.GeneratedRootKey, root.Hex()), slog.String(keys.CanonicalRootKey, rewardsEvent.MerkleRoot.Hex()))
	} else {
		logger.Info("Your Merkle tree's root matches the canonical root! You will be able to use this file for claiming rewards.")
	}

	// Create the JSON files
	rewardsFile.SetMinipoolPerformanceFileCID("---")
	logger.Info("Saving JSON files...")
	localMinipoolPerformanceFile := rprewards.NewLocalFile[sharedtypes.IMinipoolPerformanceFile](
		rewardsFile.GetMinipoolPerformanceFile(),
		t.cfg.GetMinipoolPerformancePath(index),
	)
	localRewardsFile := rprewards.NewLocalFile[sharedtypes.IRewardsFile](
		rewardsFile,
		t.cfg.GetRewardsTreePath(index),
	)

	// Write the files
	err = localMinipoolPerformanceFile.Write()
	if err != nil {
		t.handleError(fmt.Errorf("error saving minipool performance file: %w", err), logger)
		return
	}
	err = localRewardsFile.Write()
	if err != nil {
		t.handleError(fmt.Errorf("error saving rewards file: %w", err), logger)
		return
	}

	t.logger.Info("Merkle tree generation complete!")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

func (t *GenerateRewardsTree) handleError(err error, logger *slog.Logger) {
	logger.Error("*** Rewards tree generation failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
