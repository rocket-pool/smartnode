package rewards

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipfs/go-cid"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Settings
const (
	SmoothingPoolDetailsBatchSize uint64 = 8

	// Obsoleted Treegen intervals (for documentation only)
	// HoleskyV2Interval uint64 = 0
	// HoleskyV3Interval uint64 = 0
	// HoleskyV4Interval uint64 = 0
	// HoleskyV5Interval uint64 = 0
	// HoleskyV6Interval uint64 = 0
	// HoleskyV7Interval uint64 = 0
	// MainnetV2Interval uint64 = 4
	// MainnetV3Interval uint64 = 5
	// MainnetV4Interval uint64 = 6
	// MainnetV5Interval uint64 = 8
	// MainnetV6Interval uint64 = 12
	// MainnetV7Interval uint64 = 15
	// DevnetV2Interval uint64 = 0
	// DevnetV3Interval uint64 = 0
	// DevnetV4Interval uint64 = 0
	// DevnetV5Interval uint64 = 0
	// DevnetV6Interval uint64 = 0
	// DevnetV7Interval uint64 = 0
	// HoleskyV2Interval uint64 = 0
	// HoleskyV3Interval uint64 = 0
	// HoleskyV4Interval uint64 = 0
	// HoleskyV5Interval uint64 = 0
	// HoleskyV6Interval uint64 = 0
	// HoleskyV7Interval uint64 = 0

	// Mainnet intervals
	MainnetV8Interval  uint64 = 18
	MainnetV9Interval  uint64 = 29
	MainnetV10Interval uint64 = 30
	// Devnet intervals

	// Testnet intervals
	TestnetV10Interval uint64 = 0
)

func GetMainnetRulesetVersion(interval uint64) uint64 {
	if interval >= MainnetV10Interval {
		return 10
	}
	if interval >= MainnetV9Interval {
		return 9
	}
	return 8
}

func GetRulesetVersion(network cfgtypes.Network, interval uint64) uint64 {
	switch network {
	case cfgtypes.Network_Mainnet:
		return GetMainnetRulesetVersion(interval)
	case cfgtypes.Network_Testnet:
		return 10
	case cfgtypes.Network_Devnet:
		return 10
	default:
		return 10
	}
}

type TreeGenerator struct {
	rewardsIntervalInfos map[uint64]rewardsIntervalInfo
	logger               *log.ColorLogger
	logPrefix            string
	rp                   RewardsExecutionClient
	cfg                  *config.RocketPoolConfig
	bc                   beacon.Client
	index                uint64
	startTime            time.Time
	endTime              time.Time
	snapshotEnd          *SnapshotEnd
	elSnapshotHeader     *types.Header
	intervalsPassed      uint64
	generatorImpl        treeGeneratorImpl
	approximatorImpl     treeGeneratorImpl
}

type SnapshotEnd struct {
	// Slot is the last slot of the interval
	Slot uint64
	// ConsensusBlock is the last non-missed slot of the interval
	ConsensusBlock uint64
	// ExecutionBlock is the EL block number of ConsensusBlock
	ExecutionBlock uint64
}

type treeGeneratorImpl interface {
	generateTree(rp RewardsExecutionClient, networkName string, previousRewardsPoolAddresses []common.Address, bc RewardsBeaconClient) (*GenerateTreeResult, error)
	approximateStakerShareOfSmoothingPool(rp RewardsExecutionClient, networkName string, bc RewardsBeaconClient) (*big.Int, error)
	getRulesetVersion() uint64
	// Returns the primary artifact cid for consensus, all cids of all files in a map, and any potential errors
	saveFiles(smartnode *config.SmartnodeConfig, treeResult *GenerateTreeResult, nodeTrusted bool) (cid.Cid, map[string]cid.Cid, error)
}

func NewTreeGenerator(logger *log.ColorLogger, logPrefix string, rp RewardsExecutionClient, cfg *config.RocketPoolConfig, bc beacon.Client, index uint64, startTime time.Time, endTime time.Time, snapshotEnd *SnapshotEnd, elSnapshotHeader *types.Header, intervalsPassed uint64, state *state.NetworkState) (*TreeGenerator, error) {
	t := &TreeGenerator{
		logger:           logger,
		logPrefix:        logPrefix,
		rp:               rp,
		cfg:              cfg,
		bc:               bc,
		index:            index,
		startTime:        startTime,
		endTime:          endTime,
		snapshotEnd:      snapshotEnd,
		elSnapshotHeader: elSnapshotHeader,
		intervalsPassed:  intervalsPassed,
	}

	// v10
	v10_generator := newTreeGeneratorImpl_v9_v10(10, t.logger, t.logPrefix, t.index, t.snapshotEnd, t.elSnapshotHeader, t.intervalsPassed, state)

	// v9
	v9_generator := newTreeGeneratorImpl_v9_v10(9, t.logger, t.logPrefix, t.index, t.snapshotEnd, t.elSnapshotHeader, t.intervalsPassed, state)

	// v8
	v8_generator := newTreeGeneratorImpl_v8(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.snapshotEnd.ConsensusBlock, t.elSnapshotHeader, t.intervalsPassed, state)

	// Create the interval wrappers
	rewardsIntervalInfos := []rewardsIntervalInfo{
		{
			rewardsRulesetVersion: 10,
			mainnetStartInterval:  MainnetV10Interval,
			testnetStartInterval:  TestnetV10Interval,
			generator:             v10_generator,
		},
		{
			rewardsRulesetVersion: 9,
			mainnetStartInterval:  MainnetV9Interval,
			testnetStartInterval:  0,
			generator:             v9_generator,
		},
		{
			rewardsRulesetVersion: 8,
			mainnetStartInterval:  MainnetV8Interval,
			testnetStartInterval:  0,
			generator:             v8_generator,
		},
	}

	// Create the map of versions to infos
	t.rewardsIntervalInfos = map[uint64]rewardsIntervalInfo{}
	for _, info := range rewardsIntervalInfos {
		// Sanity check to make sure there aren't multiple infos with the same version
		_, exists := t.rewardsIntervalInfos[info.rewardsRulesetVersion]
		if exists {
			return nil, fmt.Errorf("multiple ruleset interval infos with ruleset v%d", info.rewardsRulesetVersion)
		}

		t.rewardsIntervalInfos[info.rewardsRulesetVersion] = info
	}

	// Get the current network
	network := t.cfg.Smartnode.Network.Value.(cfgtypes.Network)

	// Determine which actual rulesets to use based on the current interval number, checking in descending order.
	foundGenerator := false
	foundApproximator := false

	// Sort by version number, reversed. That way we will pick the highest version number whose startInterval
	// is eligible
	slices.SortFunc(rewardsIntervalInfos, func(a, b rewardsIntervalInfo) int {
		// b - a sorts high to low
		return int(b.rewardsRulesetVersion) - int(a.rewardsRulesetVersion)
	})

	// The first ruleset whose startInterval is at most t.index is the one to use
	for _, info := range rewardsIntervalInfos {

		startInterval, err := info.GetStartInterval(network)
		if err != nil {
			return nil, fmt.Errorf("error getting start interval for rewards period %d: %w", t.index, err)
		}
		if !foundGenerator && startInterval <= t.index {
			t.generatorImpl = info.generator
			foundGenerator = true
		}
		if !foundApproximator && startInterval <= t.index {
			t.approximatorImpl = info.generator
			foundApproximator = true
		}

		if foundGenerator && foundApproximator {
			break
		}
	}

	if !foundGenerator || !foundApproximator {
		// Do not default- require intervals to be explicit
		return nil, fmt.Errorf("No treegen implementation could be found for interval %d", t.index)
	}

	return t, nil
}

type GenerateTreeResult struct {
	RewardsFile             IRewardsFile
	MinipoolPerformanceFile IMinipoolPerformanceFile
	InvalidNetworkNodes     map[common.Address]uint64
}

func (t *TreeGenerator) GenerateTree() (*GenerateTreeResult, error) {
	return t.generatorImpl.generateTree(t.rp, fmt.Sprint(t.cfg.Smartnode.Network.Value), t.cfg.Smartnode.GetPreviousRewardsPoolAddresses(), t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPool() (*big.Int, error) {
	return t.approximatorImpl.approximateStakerShareOfSmoothingPool(t.rp, fmt.Sprint(t.cfg.Smartnode.Network.Value), t.bc)
}

func (t *TreeGenerator) GetGeneratorRulesetVersion() uint64 {
	return t.generatorImpl.getRulesetVersion()
}

func (t *TreeGenerator) GetApproximatorRulesetVersion() uint64 {
	return t.approximatorImpl.getRulesetVersion()
}

func (t *TreeGenerator) GenerateTreeWithRuleset(ruleset uint64) (*GenerateTreeResult, error) {
	info, exists := t.rewardsIntervalInfos[ruleset]
	if !exists {
		return nil, fmt.Errorf("ruleset v%d does not exist", ruleset)
	}

	return info.generator.generateTree(
		t.rp,
		fmt.Sprint(t.cfg.Smartnode.Network.Value),
		t.cfg.Smartnode.GetPreviousRewardsPoolAddresses(),
		t.bc,
	)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPoolWithRuleset(ruleset uint64) (*big.Int, error) {
	info, exists := t.rewardsIntervalInfos[ruleset]
	if !exists {
		return nil, fmt.Errorf("ruleset v%d does not exist", ruleset)
	}

	return info.generator.approximateStakerShareOfSmoothingPool(t.rp, fmt.Sprint(t.cfg.Smartnode.Network.Value), t.bc)
}

func (t *TreeGenerator) SaveFiles(treeResult *GenerateTreeResult, nodeTrusted bool) (cid.Cid, map[string]cid.Cid, error) {
	return t.generatorImpl.saveFiles(t.cfg.Smartnode, treeResult, nodeTrusted)
}
