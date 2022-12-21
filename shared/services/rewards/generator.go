package rewards

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Settings
const (
	SmoothingPoolDetailsBatchSize uint64 = 8

	// Mainnet intervals
	MainnetV2Interval uint64 = 4
	MainnetV3Interval uint64 = 5

	// Prater intervals
	PraterV2Interval uint64 = 37
	PraterV3Interval uint64 = 49
)

type TreeGenerator struct {
	rewardsIntervalInfos map[uint64]rewardsIntervalInfo
	logger               log.ColorLogger
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
	generateTree(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*RewardsFile, error)
	approximateStakerShareOfSmoothingPool(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*big.Int, error)
	getRulesetVersion() uint64
}

func NewTreeGenerator(logger log.ColorLogger, logPrefix string, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client, index uint64, startTime time.Time, endTime time.Time, consensusBlock uint64, elSnapshotHeader *types.Header, intervalsPassed uint64) (*TreeGenerator, error) {
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

	// Create the interval wrappers
	rewardsIntervalInfos := []rewardsIntervalInfo{
		{
			rewardsRulesetVersion: 3,
			mainnetStartInterval:  MainnetV3Interval,
			praterStartInterval:   PraterV3Interval,
			generator:             newTreeGeneratorImpl_v3(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed),
		}, {
			rewardsRulesetVersion: 2,
			mainnetStartInterval:  MainnetV2Interval,
			praterStartInterval:   PraterV2Interval,
			generator:             newTreeGeneratorImpl_v2(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed),
		}, {
			rewardsRulesetVersion: 1,
			mainnetStartInterval:  0,
			praterStartInterval:   0,
			generator:             newTreeGeneratorImpl_v1(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed),
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

func (t *TreeGenerator) GenerateTree() (*RewardsFile, error) {
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

func (t *TreeGenerator) GenerateTreeWithRuleset(ruleset uint64) (*RewardsFile, error) {
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
