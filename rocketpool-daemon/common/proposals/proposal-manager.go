package proposals

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

type ProposalManager struct {
	viSnapshotMgr  *VotingInfoSnapshotManager
	networkTreeMgr *NetworkTreeManager
	nodeTreeMgr    *NodeTreeManager
	stateMgr       *state.NetworkStateManager

	logger *slog.Logger
	cfg    *config.SmartNodeConfig
	rp     *rocketpool.RocketPool
	bc     beacon.IBeaconClient
}

func NewProposalManager(context context.Context, logger *slog.Logger, cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, bc beacon.IBeaconClient) (*ProposalManager, error) {
	viSnapshotMgr, err := NewVotingInfoSnapshotManager(logger, cfg, rp)
	if err != nil {
		return nil, fmt.Errorf("error creating voting info manager: %w", err)
	}

	networkMgr, err := NewNetworkTreeManager(logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating network tree manager: %w", err)
	}

	nodeMgr, err := NewNodeTreeManager(logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating node tree manager: %w", err)
	}

	stateMgr, err := state.NewNetworkStateManager(context, rp, cfg, rp.Client, bc, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating network state manager: %w", err)
	}

	sublogger := logger.With(slog.String(keys.TaskKey, "PDAO Proposals"))
	return &ProposalManager{
		viSnapshotMgr:  viSnapshotMgr,
		networkTreeMgr: networkMgr,
		nodeTreeMgr:    nodeMgr,
		stateMgr:       stateMgr,

		logger: sublogger,
		cfg:    cfg,
		rp:     rp,
		bc:     bc,
	}, nil
}

func (m *ProposalManager) CreateLatestFinalizedTree(context context.Context) (uint32, *NetworkVotingTree, error) {
	// Get the latest finalized block
	block, err := m.stateMgr.GetLatestFinalizedBeaconBlock(context)
	if err != nil {
		return 0, nil, fmt.Errorf("error determining latest finalized block: %w", err)
	}
	blockNumber := uint32(block.ExecutionBlockNumber)

	// Get the network tree for the block
	tree, err := m.GetNetworkTree(blockNumber, nil)
	if err != nil {
		return 0, nil, err
	}

	return blockNumber, tree, nil
}

func (m *ProposalManager) CreatePollardForProposal(context context.Context) (uint32, []*types.VotingTreeNode, error) {
	blockNumber, tree, err := m.CreateLatestFinalizedTree(context)
	if err != nil {
		return 0, nil, err
	}

	_, pollard := tree.GetPollardForProposal()
	return blockNumber, pollard, nil
}

func (m *ProposalManager) GetPollardForProposal(blockNumber uint32) ([]*types.VotingTreeNode, error) {
	tree, err := m.GetNetworkTree(blockNumber, nil)
	if err != nil {
		return nil, err
	}

	_, pollard := tree.GetPollardForProposal()
	return pollard, nil
}

func (m *ProposalManager) GetVotingInfoSnapshot(blockNumber uint32) (*VotingInfoSnapshot, error) {
	snapshot, _ := m.viSnapshotMgr.LoadFromDisk(blockNumber)
	if snapshot != nil {
		return snapshot, nil
	}

	// Generate the snapshot
	m.logger.Info("Creating voting info snapshot...", slog.Uint64(keys.BlockKey, uint64(blockNumber)))
	var err error
	snapshot, err = m.viSnapshotMgr.CreateVotingInfoSnapshot(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error creating voting info snapshot for block %d: %w", blockNumber, err)
	}
	err = m.viSnapshotMgr.SaveToFile(snapshot)
	if err != nil {
		return nil, fmt.Errorf("error saving voting info snapshot for block %d: %w", blockNumber, err)
	}
	return snapshot, nil
}

func (m *ProposalManager) GetNetworkTree(blockNumber uint32, snapshot *VotingInfoSnapshot) (*NetworkVotingTree, error) {
	// Try to load the network tree from disk
	tree, _ := m.networkTreeMgr.LoadFromDisk(blockNumber)
	if tree != nil {
		return tree, nil
	}

	// Try to load the voting info snapshot from disk or create it
	m.logger.Info("Creating network tree..", slog.Uint64(keys.BlockKey, uint64(blockNumber)))
	if snapshot == nil {
		var err error
		snapshot, err = m.GetVotingInfoSnapshot(blockNumber)
		if err != nil {
			return nil, err
		}
	}

	// Get the depth per round
	depthPerRound, err := m.getDepthPerRound()
	if err != nil {
		return nil, err
	}

	// Generate the tree
	tree = m.networkTreeMgr.CreateNetworkVotingTree(snapshot, depthPerRound)
	err = m.networkTreeMgr.SaveToFile(tree)
	if err != nil {
		return nil, fmt.Errorf("error saving tree for block %d: %w", blockNumber, err)
	}
	return tree, nil
}

func (m *ProposalManager) GetNodeTree(blockNumber uint32, nodeIndex uint64, snapshot *VotingInfoSnapshot) (*NodeVotingTree, error) {
	// Try to load the node tree from disk
	tree, err := m.nodeTreeMgr.LoadFromDisk(blockNumber, nodeIndex)
	if tree != nil {
		return tree, nil
	}
	if err != nil {
		return nil, err
	}

	// Try to load the voting info snapshot from disk or create it
	m.logger.Info("Creating node tree...", slog.Uint64(keys.BlockKey, uint64(blockNumber)), slog.Uint64(keys.NodeIndexKey, nodeIndex))
	if snapshot == nil {
		snapshot, err = m.GetVotingInfoSnapshot(blockNumber)
		if err != nil {
			return nil, err
		}
	}

	// Get the depth per round
	depthPerRound, err := m.getDepthPerRound()
	if err != nil {
		return nil, err
	}

	// Generate the tree
	treeIndex := getTreeNodeIndexFromRPNodeIndex(snapshot, nodeIndex)
	tree = m.nodeTreeMgr.CreateNodeVotingTree(snapshot, nodeIndex, treeIndex, depthPerRound)
	err = m.nodeTreeMgr.SaveToFile(tree)
	if err != nil {
		return nil, fmt.Errorf("error saving tree for block %d, node index %d: %w", blockNumber, nodeIndex, err)
	}
	return tree, nil
}

// Get the artifacts required for voting on a proposal: the node's total delegated voting power, the node index, and a Merkle proof for the node's
// corresponding leaf index in the network tree
func (m *ProposalManager) GetArtifactsForVoting(blockNumber uint32, nodeAddress common.Address) (*big.Int, uint64, []types.VotingTreeNode, error) {
	// Get the voting info snapshot
	snapshot, err := m.GetVotingInfoSnapshot(blockNumber)
	if err != nil {
		return nil, 0, nil, err
	}

	// Get the node nodeIndex
	nodeIndex, err := getRPNodeIndexFromSnapshot(snapshot, nodeAddress)
	if err != nil {
		return nil, 0, nil, err
	}

	// Get the networkTree - used to build the merkle proof needed to vote
	networkTree, err := m.GetNetworkTree(blockNumber, snapshot)
	if err != nil {
		return nil, 0, nil, err
	}

	// Get the networkTree - used to fetch the totalDelegatedVp for the node
	nodeTree, err := m.GetNodeTree(blockNumber, nodeIndex, snapshot)
	if err != nil {
		return nil, 0, nil, err
	}

	// Get the artifacts
	totalDelegatedVp := nodeTree.Nodes[0].Sum
	treeIndex := getTreeNodeIndexFromRPNodeIndex(snapshot, nodeIndex)
	proofPtrs := networkTree.generateMerkleProof(treeIndex)

	proof := make([]types.VotingTreeNode, len(proofPtrs))
	for i := range proofPtrs {
		proof[i] = *proofPtrs[i]
	}
	return totalDelegatedVp, nodeIndex, proof, nil
}

// Gets the root node and pollard for a proposer's response to a challenge against a tree node
func (m *ProposalManager) GetArtifactsForChallengeResponse(blockNumber uint32, challengedIndex uint64) (types.VotingTreeNode, []types.VotingTreeNode, error) {
	// Load the voting info snapshot
	snapshot, err := m.GetVotingInfoSnapshot(blockNumber)
	if err != nil {
		return types.VotingTreeNode{}, nil, err
	}

	// Get the proper tree
	rpNodeIndex := getRPNodeIndexFromTreeNodeIndex(snapshot, challengedIndex)
	var tree *VotingTree
	if rpNodeIndex == nil {
		// This is a node in the network tree
		networkTree, err := m.GetNetworkTree(blockNumber, snapshot)
		if err != nil {
			return types.VotingTreeNode{}, nil, err
		}
		tree = networkTree.VotingTree
	} else {
		// This is a node in a node tree
		nodeTree, err := m.GetNodeTree(blockNumber, *rpNodeIndex, snapshot)
		if err != nil {
			return types.VotingTreeNode{}, nil, err
		}
		tree = nodeTree.VotingTree
	}

	// Create the artifacts
	rootPtr, pollardPtrs := tree.GetArtifactsForChallengeResponse(challengedIndex)
	pollard := make([]types.VotingTreeNode, len(pollardPtrs))
	for i := range pollardPtrs {
		pollard[i] = *pollardPtrs[i]
	}
	return *rootPtr, pollard, nil
}

// Checks a RootSubmitted event against the local artifacts to see if there's a mismatch at an index; if so, returns the index, the node, and the proof
func (m *ProposalManager) CheckForChallengeableArtifacts(event protocol.RootSubmitted) (uint64, types.VotingTreeNode, []types.VotingTreeNode, error) {
	// Load the voting info snapshot
	blockNumber := event.BlockNumber
	index := event.Index.Uint64()
	snapshot, err := m.GetVotingInfoSnapshot(blockNumber)
	if err != nil {
		return 0, types.VotingTreeNode{}, nil, err
	}

	// Get the proper tree
	rpNodeIndex := getRPNodeIndexFromTreeNodeIndex(snapshot, index)
	var tree *VotingTree
	if rpNodeIndex == nil {
		// This is a node in the network tree
		networkTree, err := m.GetNetworkTree(blockNumber, snapshot)
		if err != nil {
			return 0, types.VotingTreeNode{}, nil, err
		}
		tree = networkTree.VotingTree
	} else {
		// This is a node in a node tree
		nodeTree, err := m.GetNodeTree(blockNumber, *rpNodeIndex, snapshot)
		if err != nil {
			return 0, types.VotingTreeNode{}, nil, err
		}
		tree = nodeTree.VotingTree
	}

	// Check for artifacts
	challengedIndex, challengedNode, proofPtrs, err := tree.CheckForChallengeableArtifacts(index, event.TreeNodes)
	if err != nil {
		return 0, types.VotingTreeNode{}, nil, fmt.Errorf("error checking for challengeable artifacts: %w", err)
	}
	if challengedIndex == 0 {
		// Nothing to challenge, the trees match
		return 0, types.VotingTreeNode{}, nil, nil
	}

	proof := make([]types.VotingTreeNode, len(proofPtrs))
	for i := range proofPtrs {
		proof[i] = *proofPtrs[i]
	}
	return challengedIndex, *challengedNode, proof, nil
}

func (m *ProposalManager) getDepthPerRound() (uint64, error) {
	// Create a pDAO manager binding
	pdaoMgr, err := protocol.NewProtocolDaoManager(m.rp)
	if err != nil {
		return 0, fmt.Errorf("error creating Protocol DAO manager binding: %w", err)
	}

	// Get the depth per round
	err = m.rp.Query(nil, nil, pdaoMgr.DepthPerRound)
	if err != nil {
		return 0, fmt.Errorf("error getting depth per round: %w", err)
	}

	depthPerRound := pdaoMgr.DepthPerRound.Formatted()
	return depthPerRound, nil
}
