package rewards

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
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
	MainnetV8Interval uint64 = 18

	// Devnet intervals

	// Holesky intervals
	HoleskyV8Interval uint64 = 93
)

type TreeGenerator struct {
	rewardsIntervalInfos map[uint64]rewardsIntervalInfo
	logger               *log.ColorLogger
	logPrefix            string
	rp                   *rocketpool.RocketPool
	cfg                  *config.RocketPoolConfig
	bc                   beacon.Client
	index                uint64
	startTime            time.Time
	endTime              time.Time
	consensusBlock       uint64
	elSnapshotHeader     *types.Header
	intervalsPassed      uint64
	generatorImpl        treeGeneratorImpl
	approximatorImpl     treeGeneratorImpl
}

type treeGeneratorImpl interface {
	generateTree(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (IRewardsFile, error)
	approximateStakerShareOfSmoothingPool(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*big.Int, error)
	getRulesetVersion() uint64
}

func NewTreeGenerator(logger *log.ColorLogger, logPrefix string, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client, index uint64, startTime time.Time, endTime time.Time, consensusBlock uint64, elSnapshotHeader *types.Header, intervalsPassed uint64, state *state.NetworkState, rollingRecord *RollingRecord) (*TreeGenerator, error) {
	t := &TreeGenerator{
		logger:           logger,
		logPrefix:        logPrefix,
		rp:               rp,
		cfg:              cfg,
		bc:               bc,
		index:            index,
		startTime:        startTime,
		endTime:          endTime,
		consensusBlock:   consensusBlock,
		elSnapshotHeader: elSnapshotHeader,
		intervalsPassed:  intervalsPassed,
	}

	// v8
	var v8_generator treeGeneratorImpl
	if rollingRecord == nil {
		v8_generator = newTreeGeneratorImpl_v8(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed, state)
	} else {
		v8_generator = newTreeGeneratorImpl_v8_rolling(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed, state, rollingRecord)
	}

	// Create the interval wrappers
	rewardsIntervalInfos := []rewardsIntervalInfo{
		{
			rewardsRulesetVersion: 8,
			mainnetStartInterval:  MainnetV8Interval,
			holeskyStartInterval:  HoleskyV8Interval,
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

	// Determine which actual rulesets to use based on the current interval number, checking in descending order from the latest
	// to interval 2.
	// Do not default- require intervals to be explicit
	foundGenerator := false
	foundApproximator := false

	// Start iterating at the end
	slices.Reverse(rewardsIntervalInfos)
	for _, info := range rewardsIntervalInfos {
		startInterval, err := info.GetStartInterval(network)
		if err != nil {
			return nil, fmt.Errorf("error getting start interval for rewards period %d: %w", t.index, err)
		}
		if !foundGenerator && t.index >= startInterval {
			t.generatorImpl = info.generator
			foundGenerator = true
		}
		if !foundApproximator && t.index > startInterval {
			t.approximatorImpl = info.generator
			foundApproximator = true
		}

		if foundGenerator && foundApproximator {
			break
		}
	}

	if !foundGenerator || !foundApproximator {
		return nil, fmt.Errorf("No treegen implementation could be found for interval %d", t.index)
	}

	return t, nil
}

func (t *TreeGenerator) GenerateTree() (IRewardsFile, error) {
	return t.generatorImpl.generateTree(t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPool() (*big.Int, error) {
	return t.approximatorImpl.approximateStakerShareOfSmoothingPool(t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) GetGeneratorRulesetVersion() uint64 {
	return t.generatorImpl.getRulesetVersion()
}

func (t *TreeGenerator) GetApproximatorRulesetVersion() uint64 {
	return t.approximatorImpl.getRulesetVersion()
}

func (t *TreeGenerator) GenerateTreeWithRuleset(ruleset uint64) (IRewardsFile, error) {
	info, exists := t.rewardsIntervalInfos[ruleset]
	if !exists {
		return nil, fmt.Errorf("ruleset v%d does not exist", ruleset)
	}

	return info.generator.generateTree(t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPoolWithRuleset(ruleset uint64) (*big.Int, error) {
	info, exists := t.rewardsIntervalInfos[ruleset]
	if !exists {
		return nil, fmt.Errorf("ruleset v%d does not exist", ruleset)
	}

	return info.generator.approximateStakerShareOfSmoothingPool(t.rp, t.cfg, t.bc)
}
