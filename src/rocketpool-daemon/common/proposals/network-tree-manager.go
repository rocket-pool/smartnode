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
	networkPathFolderName            string = "network-trees"
	networkVotingTreeFilenameFormat  string = "network-tree-%d.json.zst"
	networkVotingTreeFilenamePattern string = ".*\\-(?P<block>\\d+)\\.json\\.zst"
)

// ===========================
// === Network Voting Tree ===
// ===========================

// A network voting tree
type NetworkVotingTree struct {
	*VotingTree
}

// Get the filename of the network voting tree, including the block number it's built against
func (t NetworkVotingTree) GetFilename() string {
	return fmt.Sprintf(networkVotingTreeFilenameFormat, t.BlockNumber)
}

// ===================================
// === Network Voting Tree Manager ===
// ===================================

// Struct for network voting trees
type NetworkTreeManager struct {
	logger *slog.Logger
	cfg    *config.SmartNodeConfig

	filenameRegex           *regexp.Regexp
	latestCompatibleVersion *semver.Version
	checksumManager         *ChecksumManager[uint32, NetworkVotingTree]
}

// Create a new NetworkTreeManager instance
func NewNetworkTreeManager(logger *slog.Logger, cfg *config.SmartNodeConfig) (*NetworkTreeManager, error) {
	// Create the snapshot filename regex
	sublogger := logger.With(slog.String(keys.RoutineKey, "Network Tree"))
	filenameRegex := regexp.MustCompile(networkVotingTreeFilenamePattern)

	// Create the latest compatible snapshot version
	latestCompatibleVersion, err := semver.New(latestCompatibleVersionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing latest compatible version string [%s]: %w", latestCompatibleVersionString, err)
	}

	manager := &NetworkTreeManager{
		logger:                  sublogger,
		cfg:                     cfg,
		filenameRegex:           filenameRegex,
		latestCompatibleVersion: latestCompatibleVersion,
	}

	votingPath := cfg.GetVotingPath()
	checksumFilename := filepath.Join(votingPath, networkPathFolderName, config.ChecksumTableFilename)
	checksumManager, err := NewChecksumManager[uint32, NetworkVotingTree](checksumFilename, manager)
	if err != nil {
		return nil, fmt.Errorf("error creating checksum manager: %w", err)
	}

	manager.checksumManager = checksumManager
	return manager, nil
}

// Create a network voting tree from a voting info snapshot
func (m *NetworkTreeManager) CreateNetworkVotingTree(snapshot *VotingInfoSnapshot, depthPerRound uint64) *NetworkVotingTree {
	// Create a map of the voting power of each node, accounting for delegation
	votingPower := map[common.Address]*big.Int{}
	for _, info := range snapshot.Info {
		delegateVp, exists := votingPower[info.Delegate]
		if !exists {
			delegateVp = big.NewInt(0)
			votingPower[info.Delegate] = delegateVp
		}
		delegateVp.Add(delegateVp, info.VotingPower)
	}

	// Make the tree leaves
	leaves := make([]*types.VotingTreeNode, len(snapshot.Info))
	zeroHash := getHashForBalance(common.Big0)
	for i, info := range snapshot.Info {
		vp, exists := votingPower[info.NodeAddress]
		if !exists || vp.Cmp(common.Big0) == 0 {
			leaves[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0),
				Hash: zeroHash,
			}
		} else {
			leaves[i] = &types.VotingTreeNode{
				Sum:  big.NewInt(0).Set(vp),
				Hash: getHashForBalance(vp),
			}
		}
	}

	// Make the tree
	network := m.cfg.Network.Value
	tree := CreateTreeFromLeaves(snapshot.BlockNumber, network, leaves, 1, depthPerRound)
	return &NetworkVotingTree{
		VotingTree: tree,
	}
}

// Save a network voting tree to a file
func (m *NetworkTreeManager) SaveToFile(tree *NetworkVotingTree) error {
	return SaveToFile(m.checksumManager, tree)
}

// Load the network tree for the provided block from disk if it exists
func (m *NetworkTreeManager) LoadFromDisk(blockNumber uint32) (*NetworkVotingTree, error) {
	tree, filename, err := LoadFromFile(m.checksumManager, blockNumber)
	if err != nil {
		m.logger.Warn("Loading network tree from disk failed, it must be regenerated.", slog.Uint64(keys.BlockKey, uint64(blockNumber)), log.Err(err))
		return nil, nil
	}
	if tree == nil {
		m.logger.Warn("Network tree must be regenerated.", slog.Uint64(keys.BlockKey, uint64(blockNumber)))
		return nil, nil
	}

	m.logger.Info("Loaded network tree from disk.", slog.Uint64(keys.BlockKey, uint64(blockNumber)), slog.String(keys.FileKey, filename))
	return tree, nil
}

// Return true if the first filename represents a block number that's lower than the second filename's block number
func (m *NetworkTreeManager) Less(firstFilename string, secondFilename string) (bool, error) {
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
func (m *NetworkTreeManager) ShouldLoadEntry(filename string, context uint32) (bool, error) {
	// Extract the block number for this file
	blockNumber, err := m.getBlockNumberFromFilename(filename)
	if err != nil {
		return false, fmt.Errorf("error parsing filename (%s): %w", filename, err)
	}

	return blockNumber == context, nil
}

// Return true if the loaded network tree can be used for processing
func (m *NetworkTreeManager) IsDataValid(data *NetworkVotingTree, filename string, context uint32) (bool, error) {
	// Check if it has the proper network
	if data.Network != m.cfg.Network.Value {
		m.logger.Warn("Network tree on disk is for the wrong network so it cannot be used.", slog.String(keys.FileKey, filename), slog.String(keys.CurrentNetworkKey, string(m.cfg.Network.Value)), slog.String(keys.FileNetworkKey, string(data.Network)))
		return false, nil
	}

	// Check if it's using a compatible version
	snapshotVersion, err := semver.New(data.SmartnodeVersion)
	if err != nil {
		m.logger.Error("Parsing network tree version failed so it cannot be used.", slog.String(keys.FileKey, filename), log.Err(err))
		return false, nil
	}
	if snapshotVersion.LT(*m.latestCompatibleVersion) {
		m.logger.Warn("Network tree was made with an incompatible Smart Node so it cannot be used.", slog.String(keys.FileKey, filename), slog.String(keys.FileVersionKey, data.SmartnodeVersion), slog.String(keys.HighestCompatibleKey, latestCompatibleVersionString))
		return false, nil
	}

	return true, nil
}

// Get the block number from a snapshot filename
func (m *NetworkTreeManager) getBlockNumberFromFilename(filename string) (uint32, error) {
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
