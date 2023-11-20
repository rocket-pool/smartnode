package proposals

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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

	stateMgr, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bc, log)
	if err != nil {
		return nil, fmt.Errorf("error creating network state manager: %w", err)
	}

	logPrefix := "[PDAO Proposals]"
	return &ProposalManager{
		viSnapshotMgr:  viSnapshotMgr,
		networkTreeMgr: networkMgr,
		nodeTreeMgr:    nodeMgr,
		stateMgr:       stateMgr,

		log:       log,
		logPrefix: logPrefix,
		cfg:       cfg,
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
	tree, err := m.GetNetworkTree(blockNumber)
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
	tree, err := m.GetNetworkTree(blockNumber)
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

func (m *ProposalManager) GetNetworkTree(blockNumber uint32) (*NetworkVotingTree, error) {
	// Try to load the network tree from disk
	tree, err := m.networkTreeMgr.LoadFromDisk(blockNumber)
	if err != nil {
		m.logMessage("Loading network tree for block %d failed: %s; regenerating tree.", blockNumber, err.Error())
	} else if tree != nil {
		return tree, nil
	}

	// Try to load the voting info snapshot from disk or create it
	m.logMessage("Network tree for block %d didn't exist, creating one.", blockNumber)
	snapshot, err := m.GetVotingInfoSnapshot(blockNumber)
	if err != nil {
		return nil, err
	}

	// Generate the tree
	tree = m.networkTreeMgr.CreateNetworkVotingTree(snapshot)
	err = m.networkTreeMgr.SaveToFile(tree)
	if err != nil {
		return nil, fmt.Errorf("error saving tree for block %d: %w", blockNumber, err)
	}
	return tree, nil
}

func (m *ProposalManager) GetNodeTree(blockNumber uint32, nodeIndex uint64) (*NodeVotingTree, error) {
	// Try to load the node tree from disk
	tree, err := m.nodeTreeMgr.LoadFromDisk(blockNumber, nodeIndex)
	if err != nil {
		m.logMessage("Loading node tree for block %d, node index %d failed: %s; regenerating tree.", blockNumber, nodeIndex, err.Error())
	} else if tree != nil {
		return tree, nil
	}

	// Try to load the voting info snapshot from disk or create it
	m.logMessage("Node tree for block %d, node index %d didn't exist, creating one.", blockNumber, nodeIndex)
	snapshot, err := m.GetVotingInfoSnapshot(blockNumber)
	if err != nil {
		return nil, err
	}

	// Generate the tree
	treeIndex := getTreeNodeIndexFromRPNodeIndex(snapshot, nodeIndex)
	tree = m.nodeTreeMgr.CreateNodeVotingTree(snapshot, nodeIndex, treeIndex)
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

	// Get the tree
	tree, err := m.GetNodeTree(blockNumber, nodeIndex)
	if err != nil {
		return nil, 0, nil, err
	}

	// Get the artifacts
	totalDelegatedVp := tree.Nodes[0].Sum
	treeIndex := getTreeNodeIndexFromRPNodeIndex(snapshot, nodeIndex)
	proofPtrs := tree.generateMerkleProof(treeIndex)

	proof := make([]types.VotingTreeNode, len(proofPtrs))
	for i := range proofPtrs {
		proof[i] = *proofPtrs[i]
	}
	return totalDelegatedVp, nodeIndex, proof, nil
}

// Log a message to the logger
func (m *ProposalManager) logMessage(message string, args ...any) {
	if m.log != nil {
		m.log.Printlnf(fmt.Sprintf("%s %s", m.logPrefix, message), args)
	}
}
