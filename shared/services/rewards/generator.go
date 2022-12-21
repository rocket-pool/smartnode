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
	MainnetV2Interval int64 = 4
	MainnetV3Interval int64 = 5

	// Prater intervals
	PraterV2Interval int64 = 37
	PraterV3Interval int64 = 49
)

type TreeGenerator struct {
	logger           log.ColorLogger
	logPrefix        string
	rp               *rocketpool.RocketPool
	cfg              *config.RocketPoolConfig
	bc               beacon.Client
	index            uint64
	startTime        time.Time
	endTime          time.Time
	consensusBlock   uint64
	elSnapshotHeader *types.Header
	intervalsPassed  uint64
	generatorImpl    treeGeneratorImpl
	approximatorImpl treeGeneratorImpl
}

type treeGeneratorImpl interface {
	generateTree(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*RewardsFile, error)
	approximateStakerShareOfSmoothingPool(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*big.Int, error)
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

	// Get the start intervals for this network
	var startIntervals []int64
	switch t.cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Mainnet:
		startIntervals = []int64{MainnetV3Interval, MainnetV2Interval}

	case cfgtypes.Network_Prater:
		startIntervals = []int64{PraterV3Interval, PraterV2Interval}

	case cfgtypes.Network_Devnet:
		startIntervals = []int64{-1} // Always use the latest for the devnet

	default:
		return nil, fmt.Errorf("unknown network: %s", string(t.cfg.Smartnode.Network.Value.(cfgtypes.Network)))
	}

	// Default to ruleset v1
	t.generatorImpl = newTreeGeneratorImpl_v1(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed)
	t.approximatorImpl = t.generatorImpl

	// Create newer generators
	generators := []treeGeneratorImpl{
		newTreeGeneratorImpl_v3(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed),
		newTreeGeneratorImpl_v2(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed),
	}

	// Tree generator
	for i, intervalStart := range startIntervals {
		if int64(t.index) >= intervalStart {
			t.generatorImpl = generators[i]
			break
		}
	}

	// Approximator
	for i, intervalStart := range startIntervals {
		if int64(t.index) > intervalStart {
			t.approximatorImpl = generators[i]
			break
		}
	}

	return t, nil
}

func (t *TreeGenerator) GenerateTree() (*RewardsFile, error) {
	return t.generatorImpl.generateTree(t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPool() (*big.Int, error) {
	return t.approximatorImpl.approximateStakerShareOfSmoothingPool(t.rp, t.cfg, t.bc)
}
