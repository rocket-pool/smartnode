package proposals

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

type ProposalManager struct {
	viSnapshotMgr  *VotingInfoSnapshotManager
	networkTreeMgr *NetworkTreeManager
	nodeTreeMgr    *NodeTreeManager
	stateMgr       *state.NetworkStateManager

	log       *log.ColorLogger
	logPrefix string
	cfg       *config.RocketPoolConfig
	rp        *rocketpool.RocketPool
	bc        beacon.Client
}

func NewProposalManager(log *log.ColorLogger, cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, bc beacon.Client) (*ProposalManager, error) {
	viSnapshotMgr, err := NewVotingInfoSnapshotManager(log, cfg, rp)
	if err != nil {
		return nil, fmt.Errorf("error creating voting info manager: %w", err)
	}

	networkMgr, err := NewNetworkTreeManager(log, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating network tree manager: %w", err)
	}

	nodeMgr, err := NewNodeTreeManager(log, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating node tree manager: %w", err)
	}

	stateMgr := state.NewNetworkStateManager(rp, cfg.Smartnode.GetStateManagerContracts(), bc, log)

	logPrefix := "[PDAO Proposals]"
	return &ProposalManager{
		viSnapshotMgr:  viSnapshotMgr,
		networkTreeMgr: networkMgr,
		nodeTreeMgr:    nodeMgr,
		stateMgr:       stateMgr,

		log:       log,
		logPrefix: logPrefix,
		cfg:       cfg,
		rp:        rp,
		bc:        bc,
	}, nil
}

func (m *ProposalManager) CreateLatestFinalizedTree() (uint32, *NetworkVotingTree, error) {
	// Get the latest finalized block
	block, err := m.stateMgr.GetLatestFinalizedBeaconBlock()
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

func (m *ProposalManager) CreatePollardForProposal() (uint32, []*types.VotingTreeNode, error) {
	blockNumber, tree, err := m.CreateLatestFinalizedTree()
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
	snapshot, err := m.viSnapshotMgr.LoadFromDisk(blockNumber)
	if err != nil {
		m.logMessage("Loading voting info snapshot for block %d failed: %s; regenerating snapshot.", blockNumber, err.Error())
	} else if snapshot != nil {
		return snapshot, nil
	}

	// Generate the snapshot
	m.logMessage("Voting info snapshot for block %d didn't exist, creating one.", blockNumber)
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
	tree, err := m.networkTreeMgr.LoadFromDisk(blockNumber)
	if err != nil {
		m.logMessage("Loading network tree for block %d failed: %s; regenerating tree.", blockNumber, err.Error())
	} else if tree != nil {
		return tree, nil
	}

	// Try to load the voting info snapshot from disk or create it
	m.logMessage("Network tree for block %d didn't exist, creating one.", blockNumber)
	if snapshot == nil {
		snapshot, err = m.GetVotingInfoSnapshot(blockNumber)
		if err != nil {
			return nil, err
		}
	}

	// Get the depth per round
	depthPerRound, err := protocol.GetDepthPerRound(m.rp, nil)
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
	if err != nil {
		m.logMessage("Loading node tree for block %d, node index %d failed: %s; regenerating tree.", blockNumber, nodeIndex, err.Error())
	} else if tree != nil {
		return tree, nil
	}

	// Try to load the voting info snapshot from disk or create it
	m.logMessage("Node tree for block %d, node index %d didn't exist, creating one.", blockNumber, nodeIndex)
	if snapshot == nil {
		snapshot, err = m.GetVotingInfoSnapshot(blockNumber)
		if err != nil {
			return nil, err
		}
	}

	// Get the depth per round
	depthPerRound, err := protocol.GetDepthPerRound(m.rp, nil)
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
	if totalDelegatedVp == nil {
		totalDelegatedVp = big.NewInt(0)
	}
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

// Log a message to the logger
func (m *ProposalManager) logMessage(message string, args ...any) {
	if m.log != nil {
		m.log.Printlnf(fmt.Sprintf("%s %s", m.logPrefix, message), args...)
	}
}
