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
	MainnetV2Interval             uint64 = 4
	PraterV2Interval              uint64 = 37
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
	impl             treeGeneratorImpl
}

type treeGeneratorImpl interface {
	generateTree(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*RewardsFile, error)
	approximateStakerShareOfSmoothingPool(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (*big.Int, error)
}

func NewTreeGenerator(log log.ColorLogger, logPrefix string, rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client, index uint64, startTime time.Time, endTime time.Time, consensusBlock uint64, elSnapshotHeader *types.Header, intervalsPassed uint64) (*TreeGenerator, error) {
	t := &TreeGenerator{
		logger:           log,
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

	switch t.cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Mainnet:
		if t.index < MainnetV2Interval {
			t.impl = newTreeGeneratorImpl_v1(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed)
		} else {
			t.impl = newTreeGeneratorImpl_v2(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed)
		}
	case cfgtypes.Network_Prater:
		if t.index < PraterV2Interval {
			t.impl = newTreeGeneratorImpl_v1(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed)
		} else {
			t.impl = newTreeGeneratorImpl_v2(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed)
		}
	case cfgtypes.Network_Devnet:
		t.impl = newTreeGeneratorImpl_v2(t.logger, t.logPrefix, t.index, t.startTime, t.endTime, t.consensusBlock, t.elSnapshotHeader, t.intervalsPassed)
	default:
		return nil, fmt.Errorf("unknown network: %s", string(t.cfg.Smartnode.Network.Value.(cfgtypes.Network)))
	}

	return t, nil
}

func (t *TreeGenerator) GenerateTree() (*RewardsFile, error) {
	return t.impl.generateTree(t.rp, t.cfg, t.bc)
}

func (t *TreeGenerator) ApproximateStakerShareOfSmoothingPool() (*big.Int, error) {
	return t.impl.approximateStakerShareOfSmoothingPool(t.rp, t.cfg, t.bc)
}
