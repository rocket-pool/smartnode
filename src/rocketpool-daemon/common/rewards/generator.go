package rewards

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

// Settings
const (
	SmoothingPoolDetailsBatchSize uint64 = 8
	TestingInterval               uint64 = 1000000000 // A large number that won't ever actually be hit

	// Mainnet intervals
	MainnetV2Interval uint64 = 4
	MainnetV3Interval uint64 = 5
	MainnetV4Interval uint64 = 6
	MainnetV5Interval uint64 = 8
	MainnetV6Interval uint64 = 12
	MainnetV7Interval uint64 = 15
	MainnetV8Interval uint64 = 18

	// Devnet intervals
	DevnetV2Interval uint64 = 0
	DevnetV3Interval uint64 = 0
	DevnetV4Interval uint64 = 0
	DevnetV5Interval uint64 = 0
	DevnetV6Interval uint64 = 0
	DevnetV7Interval uint64 = 0

	// Holesky intervals
	HoleskyV2Interval uint64 = 0
	HoleskyV3Interval uint64 = 0
	HoleskyV4Interval uint64 = 0
	HoleskyV5Interval uint64 = 0
	HoleskyV6Interval uint64 = 0
	HoleskyV7Interval uint64 = 0
	HoleskyV8Interval uint64 = 93
)

type TreeGenerator struct {
	rewardsIntervalInfos map[uint64]rewardsIntervalInfo
	logger               *slog.Logger
	rp                   *rocketpool.RocketPool
	cfg                  *config.SmartNodeConfig
	bc                   beacon.IBeaconClient
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
	generateTree(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient) (sharedtypes.IRewardsFile, error)
	approximateStakerShareOfSmoothingPool(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient) (*big.Int, error)
	getRulesetVersion() uint64
}

func NewTreeGenerator(logger *slog.Logger, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient, index uint64, startTime time.Time, endTime time.Time, consensusBlock uint64, elSnapshotHeader *types.Header, intervalsPassed uint64, state *state.NetworkState, rollingRecord *RollingRecord) (*TreeGenerator, error) {
	t := &TreeGenerator{
		logger:           logger,
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
		v8_generator = newTreeGeneratorImpl_v8(t.logger, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed, state)
	} else {
		v8_generator = newTreeGeneratorImpl_v8_rolling(t.logger, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed, state, rollingRecord)
	}

	// Create the interval wrappers
	rewardsIntervalInfos := []rewardsIntervalInfo{
		{
			rewardsRulesetVersion: 8,
			mainnetStartInterval:  MainnetV8Interval,
			holeskyStartInterval:  HoleskyV8Interval,
			generator:             v8_generator,
		},
		{
			rewardsRulesetVersion: 7,
			mainnetStartInterval:  MainnetV7Interval,
			devnetStartInterval:   DevnetV7Interval,
			holeskyStartInterval:  HoleskyV7Interval,
			generator:             nil,
		}, {
			rewardsRulesetVersion: 6,
			mainnetStartInterval:  MainnetV6Interval,
			devnetStartInterval:   DevnetV6Interval,
			holeskyStartInterval:  HoleskyV6Interval,
			generator:             nil,
		}, {
			rewardsRulesetVersion: 5,
			mainnetStartInterval:  MainnetV5Interval,
			devnetStartInterval:   DevnetV5Interval,
			holeskyStartInterval:  HoleskyV5Interval,
			generator:             nil,
		}, {
			rewardsRulesetVersion: 4,
			mainnetStartInterval:  MainnetV4Interval,
			devnetStartInterval:   DevnetV4Interval,
			holeskyStartInterval:  HoleskyV4Interval,
			generator:             nil,
		}, {
			rewardsRulesetVersion: 3,
			mainnetStartInterval:  MainnetV3Interval,
			devnetStartInterval:   DevnetV3Interval,
			holeskyStartInterval:  HoleskyV3Interval,
			generator:             nil,
		}, {
			rewardsRulesetVersion: 2,
			mainnetStartInterval:  MainnetV2Interval,
			devnetStartInterval:   DevnetV2Interval,
			holeskyStartInterval:  HoleskyV2Interval,
			generator:             nil,
		}, {
			rewardsRulesetVersion: 1,
			mainnetStartInterval:  0,
			devnetStartInterval:   0,
			holeskyStartInterval:  0,
			generator:             nil,
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
	network := t.cfg.Network.Value

	// Determine which actual rulesets to use based on the current interval number, checking in descending order from the latest
	// to interval 2 since interval 1 is the default
	foundGenerator := false
	foundApproximator := false
	for i := uint64(len(t.rewardsIntervalInfos)); i > 1; i-- {
		info := t.rewardsIntervalInfos[i]
		startInterval, err := info.GetStartInterval(network)
		if err != nil {
			return nil, fmt.Errorf("error getting start interval for rewards period %d: %w", i, err)
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

	// Default to interval 1 if nothing could be found
	if !foundGenerator {
		t.generatorImpl = t.rewardsIntervalInfos[1].generator
	}
	if !foundApproximator {
		t.approximatorImpl = t.rewardsIntervalInfos[1].generator
	}

	return t, nil
}

func (t *TreeGenerator) GenerateTree(context context.Context) (sharedtypes.IRewardsFile, error) {
	return t.generatorImpl.generateTree(context, t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPool(context context.Context) (*big.Int, error) {
	return t.approximatorImpl.approximateStakerShareOfSmoothingPool(context, t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) GetGeneratorRulesetVersion() uint64 {
	return t.generatorImpl.getRulesetVersion()
}

func (t *TreeGenerator) GetApproximatorRulesetVersion() uint64 {
	return t.approximatorImpl.getRulesetVersion()
}

func (t *TreeGenerator) GenerateTreeWithRuleset(context context.Context, ruleset uint64) (sharedtypes.IRewardsFile, error) {
	info, exists := t.rewardsIntervalInfos[ruleset]
	if !exists {
		return nil, fmt.Errorf("ruleset v%d does not exist", ruleset)
	}

	if info.generator == nil {
		return nil, fmt.Errorf("ruleset v%d is obsolete and no longer supported by this Smart Node", ruleset)
	}
	return info.generator.generateTree(context, t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPoolWithRuleset(context context.Context, ruleset uint64) (*big.Int, error) {
	info, exists := t.rewardsIntervalInfos[ruleset]
	if !exists {
		return nil, fmt.Errorf("ruleset v%d does not exist", ruleset)
	}

	if info.generator == nil {
		return nil, fmt.Errorf("ruleset v%d is obsolete and no longer supported by this Smart Node", ruleset)
	}
	return info.generator.approximateStakerShareOfSmoothingPool(context, t.rp, t.cfg, t.bc)
}
