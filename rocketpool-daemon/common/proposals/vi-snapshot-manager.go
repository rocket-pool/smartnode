package proposals

import (
	"fmt"
	"math/big"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/blang/semver/v4"
	batchquery "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	votingInfoSnapshotPathFolderName  string = "vi-info"
	votingInfoSnapshotFilenameFormat  string = "vi-%d.json.zst"
	votingInfoSnapshotFilenamePattern string = ".*\\-(?P<block>\\d+)\\.json\\.zst"
	nodeInfoBatchCount                int    = 500
)

// ============================
// === Voting Info Snapshot ===
// ============================

// A network voting tree
type VotingInfoSnapshot struct {
	SmartnodeVersion string                 `json:"smartnodeVersion"`
	Network          cfgtypes.Network       `json:"network"`
	BlockNumber      uint32                 `json:"blockNumber"`
	Info             []types.NodeVotingInfo `json:"info"`
}

// Get the filename of the network voting tree, including the block number it's built against
func (t VotingInfoSnapshot) GetFilename() string {
	return fmt.Sprintf(votingInfoSnapshotFilenameFormat, t.BlockNumber)
}

// ===================================
// === Voting Info Snapshot Manager ===
// ===================================

// Struct for voting info snapshots
type VotingInfoSnapshotManager struct {
	log       *log.ColorLogger
	logPrefix string
	cfg       *config.RocketPoolConfig
	rp        *rocketpool.RocketPool

	filenameRegex           *regexp.Regexp
	latestCompatibleVersion *semver.Version
	checksumManager         *ChecksumManager[uint32, VotingInfoSnapshot]
}

// Create a new VotingInfoSnapshotManager instance
func NewVotingInfoSnapshotManager(log *log.ColorLogger, cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool) (*VotingInfoSnapshotManager, error) {
	// Create the snapshot filename regex
	logPrefix := "[Voting Info Snapshot]"
	filenameRegex := regexp.MustCompile(votingInfoSnapshotFilenamePattern)

	// Create the latest compatible snapshot version
	latestCompatibleVersion, err := semver.New(latestCompatibleVersionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing latest compatible version string [%s]: %w", latestCompatibleVersionString, err)
	}

	manager := &VotingInfoSnapshotManager{
		log:                     log,
		logPrefix:               logPrefix,
		cfg:                     cfg,
		rp:                      rp,
		filenameRegex:           filenameRegex,
		latestCompatibleVersion: latestCompatibleVersion,
	}

	votingPath := cfg.Smartnode.GetVotingPath()
	checksumFilename := filepath.Join(votingPath, votingInfoSnapshotPathFolderName, config.ChecksumTableFilename)
	checksumManager, err := NewChecksumManager[uint32, VotingInfoSnapshot](checksumFilename, manager)
	if err != nil {
		return nil, fmt.Errorf("error creating checksum manager: %w", err)
	}

	manager.checksumManager = checksumManager
	return manager, nil
}

// Create a voting info snapshot from the given block
func (m *VotingInfoSnapshotManager) CreateVotingInfoSnapshot(blockNumber uint32) (*VotingInfoSnapshot, error) {
	nodeMgr, err := node.NewNodeManager(m.rp)
	if err != nil {
		return nil, fmt.Errorf("error creating node manager binding: %w", err)
	}
	networkMgr, err := network.NewNetworkManager(m.rp)
	if err != nil {
		return nil, fmt.Errorf("error creating network manager binding: %w", err)
	}

	// Get the node addresses
	var nodeCount *big.Int
	err = m.rp.Query(func(mc *batchquery.MultiCaller) error {
		networkMgr.GetVotingNodeCountAtBlock(mc, &nodeCount, blockNumber)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting voting node count at block %d: %w", blockNumber, err)
	}
	nodeAddresses, err := nodeMgr.GetNodeAddresses(nodeMgr.NodeCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	// Get the voting info
	count := int(nodeCount.Uint64())
	infos := make([]types.NodeVotingInfo, count)
	err = m.rp.BatchQuery(count, nodeInfoBatchCount, func(mc *batchquery.MultiCaller, i int) error {
		address := nodeAddresses[i]
		node, err := node.NewNode(m.rp, address)
		if err != nil {
			return fmt.Errorf("error creating node binding for %s: %w", address.Hex(), err)
		}

		info := &infos[i]
		info.NodeAddress = address
		node.GetVotingDelegateAtBlock(mc, &info.Delegate, blockNumber)
		node.GetVotingPowerAtBlock(mc, &info.VotingPower, blockNumber)
		return nil
	}, nil)

	return &VotingInfoSnapshot{
		SmartnodeVersion: shared.RocketPoolVersion,
		Network:          m.cfg.Smartnode.Network.Value.(cfgtypes.Network),
		BlockNumber:      blockNumber,
		Info:             infos,
	}, nil
}

// Save a snapshot to a file
func (m *VotingInfoSnapshotManager) SaveToFile(snapshot *VotingInfoSnapshot) error {
	return SaveToFile(m.checksumManager, snapshot)
}

// Load the snapshot for the provided block from disk if it exists
func (m *VotingInfoSnapshotManager) LoadFromDisk(blockNumber uint32) (*VotingInfoSnapshot, error) {
	tree, filename, err := LoadFromFile(m.checksumManager, blockNumber)
	if err != nil {
		m.logMessage("%s WARNING: error loading network tree for block %d from disk: %s... it must be regenerated", m.logPrefix, blockNumber, err.Error())
		return nil, nil
	}
	if tree == nil {
		m.logMessage("%s Couldn't load network tree for block %d from disk, so it must be regenerated.", m.logPrefix, blockNumber)
		return nil, nil
	}

	m.logMessage("%s Loaded file [%s].", m.logPrefix, filename)
	return tree, nil
}

// Return true if the first filename represents a block number that's lower than the second filename's block number
func (m *VotingInfoSnapshotManager) Less(firstFilename string, secondFilename string) (bool, error) {
	firstBlock, err := m.getBlockNumberFromFilename(firstFilename)
	if err != nil {
		return false, err
	}

	secondBlock, err := m.getBlockNumberFromFilename(secondFilename)
	if err != nil {
		return false, err
	}

	return (firstBlock < secondBlock), nil
}

// Return true if the filename matches the block number provided in the context
func (m *VotingInfoSnapshotManager) ShouldLoadEntry(filename string, context uint32) (bool, error) {
	// Extract the block number for this file
	blockNumber, err := m.getBlockNumberFromFilename(filename)
	if err != nil {
		return false, fmt.Errorf("error parsing filename (%s): %w", filename, err)
	}

	return blockNumber == context, nil
}

// Return true if the loaded snapshot can be used for processing
func (m *VotingInfoSnapshotManager) IsDataValid(data *VotingInfoSnapshot, filename string, context uint32) (bool, error) {
	// Check if it has the proper network
	if data.Network != m.cfg.Smartnode.Network.Value.(cfgtypes.Network) {
		m.logMessage("%s File [%s] is for network %s instead of %s so it cannot be used.", m.logPrefix, filename, data.Network, string(m.cfg.Smartnode.Network.Value.(cfgtypes.Network)))
		return false, nil
	}

	// Check if it's using a compatible version
	snapshotVersion, err := semver.New(data.SmartnodeVersion)
	if err != nil {
		m.logMessage("%s Failed to parse the version info for file [%s] so it cannot be used.", m.logPrefix, filename)
		return false, nil
	}
	if snapshotVersion.LT(*m.latestCompatibleVersion) {
		m.logMessage("%s File [%s] was made with Smartnode v%s which is not compatible (lowest compatible = v%s) so it cannot be used.", m.logPrefix, filename, data.SmartnodeVersion, latestCompatibleVersionString)
		return false, nil
	}

	return true, nil
}

// Get the block number from a snapshot filename
func (m *VotingInfoSnapshotManager) getBlockNumberFromFilename(filename string) (uint32, error) {
	matches := m.filenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return 0, fmt.Errorf("filename (%s) did not match the expected format", filename)
	}
	blockIndex := m.filenameRegex.SubexpIndex("block")
	if blockIndex == -1 {
		return 0, fmt.Errorf("block number not found in filename (%s)", filename)
	}
	blockString := matches[blockIndex]
	blockNumber, err := strconv.ParseUint(blockString, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("block number (%s) could not be parsed to a number", blockString)
	}

	return uint32(blockNumber), nil
}

// Log a message to the logger
func (m *VotingInfoSnapshotManager) logMessage(message string, args ...any) {
	if m.log != nil {
		m.log.Printlnf(message, args)
	}
}
