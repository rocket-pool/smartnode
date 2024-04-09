package proposals

import (
	"fmt"
	"log/slog"
	"math/big"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/blang/semver/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	nodePathFolderName            string = "node-trees"
	nodeVotingTreeFilenameFormat  string = "node-tree-%d-%s-%d.json.zst"
	nodeVotingTreeFilenamePattern string = ".*\\-(?P<block>\\d+)\\-(?P<address>0x[0-9a-fA-F]{40})\\-(?P<index>\\d+)\\.json\\.zst"
)

// ========================
// === Node Voting Tree ===
// ========================

// A node voting tree
type NodeVotingTree struct {
	Address   common.Address `json:"address"`
	NodeIndex uint64         `json:"nodeIndex"`
	*VotingTree
}

// Get the filename of the node voting tree, including the block number it's built against, the node's address, and the node's index
func (t NodeVotingTree) GetFilename() string {
	return fmt.Sprintf(nodeVotingTreeFilenameFormat, t.BlockNumber, t.Address, t.NodeIndex)
}

// ========================================
// === Node Voting Tree Loading Context ===
// ========================================

type NodeVotingTreeLoadingContext struct {
	blockNumber uint32
	nodeIndex   uint64
}

// ================================
// === Node Voting Tree Manager ===
// ================================

// Struct for node voting trees
type NodeTreeManager struct {
	logger *slog.Logger
	cfg    *config.SmartNodeConfig

	filenameRegex           *regexp.Regexp
	latestCompatibleVersion *semver.Version
	checksumManager         *ChecksumManager[NodeVotingTreeLoadingContext, NodeVotingTree]
}

// Create a new NodeTreeManager instance
func NewNodeTreeManager(logger *slog.Logger, cfg *config.SmartNodeConfig) (*NodeTreeManager, error) {
	// Create the snapshot filename regex
	sublogger := logger.With(slog.String(keys.RoutineKey, "Node Tree"))
	filenameRegex := regexp.MustCompile(nodeVotingTreeFilenamePattern)

	// Create the latest compatible snapshot version
	latestCompatibleVersion, err := semver.New(latestCompatibleVersionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing latest compatible version string [%s]: %w", latestCompatibleVersionString, err)
	}

	manager := &NodeTreeManager{
		logger:                  sublogger,
		cfg:                     cfg,
		filenameRegex:           filenameRegex,
		latestCompatibleVersion: latestCompatibleVersion,
	}

	votingPath := cfg.GetVotingPath()
	checksumFilename := filepath.Join(votingPath, nodePathFolderName, config.ChecksumTableFilename)
	checksumManager, err := NewChecksumManager[NodeVotingTreeLoadingContext, NodeVotingTree](checksumFilename, manager)
	if err != nil {
		return nil, fmt.Errorf("error creating checksum manager: %w", err)
	}

	manager.checksumManager = checksumManager
	return manager, nil
}

// Create a node voting tree from a voting info snapshot and the node's index
func (m *NodeTreeManager) CreateNodeVotingTree(snapshot *VotingInfoSnapshot, rpNodeIndex uint64, networkTreeNodeIndex uint64, depthPerRound uint64) *NodeVotingTree {
	address := &snapshot.Info[rpNodeIndex].NodeAddress
	leaves := make([]*types.VotingTreeNode, len(snapshot.Info))
	zeroHash := getHashForBalance(common.Big0)
	for i, info := range snapshot.Info {
		if info.Delegate == *address {
			leaves[i] = &types.VotingTreeNode{
				Sum:  info.VotingPower,
				Hash: getHashForBalance(info.VotingPower),
			}
		} else {
			leaves[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0),
				Hash: zeroHash,
			}
		}
	}

	// Make the tree
	network := m.cfg.Network.Value
	tree := CreateTreeFromLeaves(snapshot.BlockNumber, network, leaves, networkTreeNodeIndex, depthPerRound)
	return &NodeVotingTree{
		Address:    *address,
		NodeIndex:  rpNodeIndex,
		VotingTree: tree,
	}
}

// Save a node voting tree to a file
func (m *NodeTreeManager) SaveToFile(tree *NodeVotingTree) error {
	return SaveToFile(m.checksumManager, tree)
}

// Load the node tree for the provided block from disk if it exists
func (m *NodeTreeManager) LoadFromDisk(blockNumber uint32, rpIndex uint64) (*NodeVotingTree, error) {
	context := NodeVotingTreeLoadingContext{
		blockNumber: blockNumber,
		nodeIndex:   rpIndex,
	}
	tree, filename, err := LoadFromFile(m.checksumManager, context)
	if err != nil {
		m.logger.Warn("Loading node tree from disk failed, it must be regenerated.", slog.Uint64(keys.BlockKey, uint64(blockNumber)), slog.Uint64(keys.NodeIndexKey, rpIndex), log.Err(err))
		return nil, nil
	}
	if tree == nil {
		m.logger.Warn("Node tree must be regenerated.", slog.Uint64(keys.BlockKey, uint64(blockNumber)), slog.Uint64(keys.NodeIndexKey, rpIndex))
		return nil, nil
	}

	m.logger.Info("Loaded node tree from disk.", slog.Uint64(keys.BlockKey, uint64(blockNumber)), slog.Uint64(keys.NodeIndexKey, rpIndex), slog.String(keys.FileKey, filename))
	return tree, nil
}

// Sort the checksum file entries by their block number
func (m *NodeTreeManager) Less(firstFilename string, secondFilename string) (bool, error) {
	firstBlock, firstNodeIndex, err := m.getInfoFromFilename(firstFilename)
	if err != nil {
		return false, err
	}

	secondBlock, secondNodeIndex, err := m.getInfoFromFilename(secondFilename)
	if err != nil {
		return false, err
	}

	if firstBlock < secondBlock {
		return true, nil
	} else if firstBlock == secondBlock {
		return firstNodeIndex < secondNodeIndex, nil
	}
	return false, nil
}

// Get the checksum, the filename, and the block number from a checksum entry.
func (m *NodeTreeManager) ShouldLoadEntry(filename string, context NodeVotingTreeLoadingContext) (bool, error) {
	// Extract the block number for this file
	blockNumber, nodeIndex, err := m.getInfoFromFilename(filename)
	if err != nil {
		return false, fmt.Errorf("error parsing filename (%s): %w", filename, err)
	}

	return (blockNumber == context.blockNumber && nodeIndex == context.nodeIndex), nil
}

// Return true if the loaded node tree can be used for processing
func (m *NodeTreeManager) IsDataValid(data *NodeVotingTree, filename string, context NodeVotingTreeLoadingContext) (bool, error) {
	// Check if it has the proper network
	if data.Network != m.cfg.Network.Value {
		m.logger.Warn("Node tree on disk is for the wrong network so it cannot be used.", slog.String(keys.FileKey, filename), slog.String(keys.CurrentNetworkKey, string(m.cfg.Network.Value)), slog.String(keys.FileNetworkKey, string(data.Network)))
		return false, nil
	}

	// Check if it's using a compatible version
	snapshotVersion, err := semver.New(data.SmartnodeVersion)
	if err != nil {
		m.logger.Error("Parsing node tree version failed so it cannot be used.", slog.String(keys.FileKey, filename), log.Err(err))
		return false, nil
	}
	if snapshotVersion.LT(*m.latestCompatibleVersion) {
		m.logger.Warn("Node tree was made with an incompatible Smart Node so it cannot be used.", slog.String(keys.FileKey, filename), slog.String(keys.FileVersionKey, data.SmartnodeVersion), slog.String(keys.HighestCompatibleKey, latestCompatibleVersionString))
		return false, nil
	}

	return true, nil
}

// Get the block number and node index from a snapshot filename
func (m *NodeTreeManager) getInfoFromFilename(filename string) (uint32, uint64, error) {
	matches := m.filenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return 0, 0, fmt.Errorf("filename (%s) did not match the expected format", filename)
	}

	// Block number
	blockIndex := m.filenameRegex.SubexpIndex("block")
	if blockIndex == -1 {
		return 0, 0, fmt.Errorf("block number not found in filename (%s)", filename)
	}
	blockString := matches[blockIndex]
	blockNumber, err := strconv.ParseUint(blockString, 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("block number (%s) could not be parsed to a number", blockString)
	}

	// Node index
	indexOfNodeIndex := m.filenameRegex.SubexpIndex("index")
	if indexOfNodeIndex == -1 {
		return 0, 0, fmt.Errorf("node index not found in filename (%s)", filename)
	}
	nodeString := matches[indexOfNodeIndex]
	nodeIndex, err := strconv.ParseUint(nodeString, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("node index (%s) could not be parsed to a number", blockString)
	}

	return uint32(blockNumber), nodeIndex, nil
}
