package node

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const stateFixture = "../../shared/services/state/testdata/small_state.json"

// --- stubs ---------------------------------------------------------------

// stubBeaconClient satisfies beacon.Client; only GetBeaconBlock is called by
// the challenge flow, so the rest panic if invoked.
type stubBeaconClient struct {
	beacon.Client // embedded nil – will panic on any un-overridden method
	block         beacon.BeaconBlock
}

func (s *stubBeaconClient) GetBeaconBlock(_ string) (beacon.BeaconBlock, bool, error) {
	return s.block, true, nil
}

// stubNodeGetter returns a fixed on-chain root for every proposal.
type stubNodeGetter struct {
	root types.VotingTreeNode
}

func (s *stubNodeGetter) GetNode(_ uint64, _ uint64) (types.VotingTreeNode, error) {
	return s.root, nil
}

// stubTreeProvider returns a fixed local tree for every target block.
type stubTreeProvider struct {
	tree *proposals.NetworkVotingTree
}

func (s *stubTreeProvider) GetNetworkTree(_ uint32) (*proposals.NetworkVotingTree, error) {
	return s.tree, nil
}

// stubChallengeStateGetter returns a preconfigured state per (proposal, index).
type stubChallengeStateGetter struct {
	states map[[2]uint64]types.ChallengeState
}

func (s *stubChallengeStateGetter) GetChallengeState(proposalID uint64, index uint64) (types.ChallengeState, error) {
	key := [2]uint64{proposalID, index}
	if cs, ok := s.states[key]; ok {
		return cs, nil
	}
	return types.ChallengeState_Unchallenged, nil
}

// stubEventProvider returns pre-configured RootSubmitted events.
type stubEventProvider struct {
	events []protocol.RootSubmitted
}

func (s *stubEventProvider) GetRootSubmittedEvents(_ []uint64, _, _ *big.Int) ([]protocol.RootSubmitted, error) {
	return s.events, nil
}

// stubArtifactChecker simulates finding a challengeable child index.
type stubArtifactChecker struct {
	childIndex     uint64
	challengedNode types.VotingTreeNode
	proof          []types.VotingTreeNode
}

func (s *stubArtifactChecker) CheckForChallengeableArtifacts(_ protocol.RootSubmitted) (uint64, types.VotingTreeNode, []types.VotingTreeNode, error) {
	return s.childIndex, s.challengedNode, s.proof, nil
}

// --- helpers -------------------------------------------------------------

func makeLocalRoot() types.VotingTreeNode {
	return types.VotingTreeNode{
		Sum:  big.NewInt(1000),
		Hash: common.HexToHash("0xaaaa"),
	}
}

func makeMismatchedOnChainRoot() types.VotingTreeNode {
	return types.VotingTreeNode{
		Sum:  big.NewInt(9999),
		Hash: common.HexToHash("0xbbbb"),
	}
}

// --- tests ---------------------------------------------------------------

// TestChallengePathNoEligible verifies that when our node address matches the
// proposer of the only pending proposal, no challenges are produced.
func TestChallengePathNoEligible(t *testing.T) {
	provider, err := state.NewStaticNetworkStateProviderFromFile(stateFixture)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}
	ns, _ := provider.GetHeadState()

	// Find the pending proposal's proposer and use it as our node address
	var proposerAddr common.Address
	for _, p := range ns.ProtocolDaoProposalDetails {
		if p.State == types.ProtocolDaoProposalState_Pending && p.ID != 0 {
			proposerAddr = p.ProposerAddress
			break
		}
	}

	logger := log.NewColorLogger(0)
	challenges, defeats, err := getChallengesFromState(
		ns, proposerAddr, &logger,
		&stubBeaconClient{block: beacon.BeaconBlock{ExecutionBlockNumber: ns.ElBlockNumber}},
		&stubNodeGetter{root: makeLocalRoot()},
		&stubTreeProvider{},
		&stubChallengeStateGetter{},
		&stubEventProvider{},
		&stubArtifactChecker{},
		map[uint64]bool{},
		map[uint64]map[uint64]*protocol.RootSubmitted{},
		nil,
	)
	if err != nil {
		t.Fatalf("getChallengesFromState: %v", err)
	}
	if len(challenges) != 0 {
		t.Errorf("expected 0 challenges when node is the proposer, got %d", len(challenges))
	}
	if len(defeats) != 0 {
		t.Errorf("expected 0 defeats when node is the proposer, got %d", len(defeats))
	}
}

// TestChallengePathMatchingRoot verifies that a pending proposal whose
// on-chain root matches the local tree produces no challenges.
func TestChallengePathMatchingRoot(t *testing.T) {
	provider, err := state.NewStaticNetworkStateProviderFromFile(stateFixture)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}
	ns, _ := provider.GetHeadState()

	localRoot := makeLocalRoot()
	localTree := &proposals.NetworkVotingTree{
		VotingTree: &proposals.VotingTree{
			Nodes: []*types.VotingTreeNode{&localRoot},
		},
	}

	nodeAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	logger := log.NewColorLogger(0)
	validCache := map[uint64]bool{}

	challenges, defeats, err := getChallengesFromState(
		ns, nodeAddr, &logger,
		&stubBeaconClient{block: beacon.BeaconBlock{ExecutionBlockNumber: ns.ElBlockNumber}},
		&stubNodeGetter{root: localRoot},
		&stubTreeProvider{tree: localTree},
		&stubChallengeStateGetter{},
		&stubEventProvider{},
		&stubArtifactChecker{},
		validCache,
		map[uint64]map[uint64]*protocol.RootSubmitted{},
		nil,
	)
	if err != nil {
		t.Fatalf("getChallengesFromState: %v", err)
	}
	if len(challenges) != 0 {
		t.Errorf("expected 0 challenges for matching root, got %d", len(challenges))
	}
	if len(defeats) != 0 {
		t.Errorf("expected 0 defeats for matching root, got %d", len(defeats))
	}

	// Matching proposal should be cached as valid
	for _, p := range ns.ProtocolDaoProposalDetails {
		if p.State == types.ProtocolDaoProposalState_Pending && p.ID != 0 {
			if !validCache[p.ID] {
				t.Errorf("proposal %d should have been cached as valid", p.ID)
			}
		}
	}
}

// TestChallengePathMismatchProducesChallenge verifies that a root mismatch
// causes the flow to produce a challenge against the identified child index.
func TestChallengePathMismatchProducesChallenge(t *testing.T) {
	provider, err := state.NewStaticNetworkStateProviderFromFile(stateFixture)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}
	ns, _ := provider.GetHeadState()

	// Find the pending proposal
	var pendingProp protocol.ProtocolDaoProposalDetails
	for _, p := range ns.ProtocolDaoProposalDetails {
		if p.State == types.ProtocolDaoProposalState_Pending && p.ID != 0 {
			pendingProp = p
			break
		}
	}

	// Local root differs from on-chain root
	localRoot := makeLocalRoot()
	localTree := &proposals.NetworkVotingTree{
		VotingTree: &proposals.VotingTree{
			Nodes: []*types.VotingTreeNode{&localRoot},
		},
	}

	// Simulate a RootSubmitted event at index 1 (root) for this proposal
	rootEvent := protocol.RootSubmitted{
		ProposalID:  big.NewInt(int64(pendingProp.ID)),
		Proposer:    pendingProp.ProposerAddress,
		BlockNumber: pendingProp.TargetBlock,
		Index:       big.NewInt(1),
		Root:        makeMismatchedOnChainRoot(),
	}

	challengedChildNode := types.VotingTreeNode{
		Sum:  big.NewInt(500),
		Hash: common.HexToHash("0xcccc"),
	}
	proofNodes := []types.VotingTreeNode{
		{Sum: big.NewInt(500), Hash: common.HexToHash("0xdddd")},
	}

	nodeAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	logger := log.NewColorLogger(0)

	challenges, defeats, err := getChallengesFromState(
		ns, nodeAddr, &logger,
		&stubBeaconClient{block: beacon.BeaconBlock{ExecutionBlockNumber: ns.ElBlockNumber}},
		&stubNodeGetter{root: makeMismatchedOnChainRoot()},
		&stubTreeProvider{tree: localTree},
		&stubChallengeStateGetter{},
		&stubEventProvider{events: []protocol.RootSubmitted{rootEvent}},
		&stubArtifactChecker{childIndex: 2, challengedNode: challengedChildNode, proof: proofNodes},
		map[uint64]bool{},
		map[uint64]map[uint64]*protocol.RootSubmitted{},
		nil,
	)
	if err != nil {
		t.Fatalf("getChallengesFromState: %v", err)
	}
	if len(challenges) != 1 {
		t.Fatalf("expected 1 challenge, got %d", len(challenges))
	}
	if challenges[0].proposalID != pendingProp.ID {
		t.Errorf("challenge proposalID: got %d, want %d", challenges[0].proposalID, pendingProp.ID)
	}
	if challenges[0].challengedIndex != 2 {
		t.Errorf("challenge index: got %d, want 2", challenges[0].challengedIndex)
	}
	if challenges[0].challengedNode.Hash != challengedChildNode.Hash {
		t.Errorf("challenge node hash mismatch")
	}
	if len(challenges[0].witness) != 1 {
		t.Errorf("challenge proof length: got %d, want 1", len(challenges[0].witness))
	}
	if len(defeats) != 0 {
		t.Errorf("expected 0 defeats, got %d", len(defeats))
	}
	t.Logf("Challenge produced for proposal %d at index %d", challenges[0].proposalID, challenges[0].challengedIndex)
}

// TestChallengePathAlreadyChallengedWaits verifies that when the target index
// is already in Challenged state but the challenge window has NOT expired
// (relative to the state's slot time), the flow returns nothing — we wait
// for the proposer to respond.
func TestChallengePathAlreadyChallengedWaits(t *testing.T) {
	provider, err := state.NewStaticNetworkStateProviderFromFile(stateFixture)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}
	ns, _ := provider.GetHeadState()

	var pendingProp protocol.ProtocolDaoProposalDetails
	for _, p := range ns.ProtocolDaoProposalDetails {
		if p.State == types.ProtocolDaoProposalState_Pending && p.ID != 0 {
			pendingProp = p
			break
		}
	}

	localRoot := makeLocalRoot()
	localTree := &proposals.NetworkVotingTree{
		VotingTree: &proposals.VotingTree{
			Nodes: []*types.VotingTreeNode{&localRoot},
		},
	}

	rootEvent := protocol.RootSubmitted{
		ProposalID:  big.NewInt(int64(pendingProp.ID)),
		Proposer:    pendingProp.ProposerAddress,
		BlockNumber: pendingProp.TargetBlock,
		Index:       big.NewInt(1),
		Root:        makeMismatchedOnChainRoot(),
	}

	// Index 2 has already been challenged
	csGetter := &stubChallengeStateGetter{
		states: map[[2]uint64]types.ChallengeState{
			{pendingProp.ID, 2}: types.ChallengeState_Challenged,
		},
	}

	nodeAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	logger := log.NewColorLogger(0)

	challenges, defeats, err := getChallengesFromState(
		ns, nodeAddr, &logger,
		&stubBeaconClient{block: beacon.BeaconBlock{ExecutionBlockNumber: ns.ElBlockNumber}},
		&stubNodeGetter{root: makeMismatchedOnChainRoot()},
		&stubTreeProvider{tree: localTree},
		csGetter,
		&stubEventProvider{events: []protocol.RootSubmitted{rootEvent}},
		&stubArtifactChecker{childIndex: 2, challengedNode: types.VotingTreeNode{Sum: big.NewInt(500), Hash: common.HexToHash("0xcccc")}},
		map[uint64]bool{},
		map[uint64]map[uint64]*protocol.RootSubmitted{},
		nil,
	)
	if err != nil {
		t.Fatalf("getChallengesFromState: %v", err)
	}
	if len(challenges) != 0 {
		t.Errorf("expected 0 challenges (waiting for response), got %d", len(challenges))
	}
	if len(defeats) != 0 {
		t.Errorf("expected 0 defeats (still within challenge window), got %d", len(defeats))
	}
}

// TestChallengePathRespondedDelveDeeper verifies that when a challenged index
// has been responded to, the flow continues deeper into the tree.
func TestChallengePathRespondedDelveDeeper(t *testing.T) {
	provider, err := state.NewStaticNetworkStateProviderFromFile(stateFixture)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}
	ns, _ := provider.GetHeadState()

	var pendingProp protocol.ProtocolDaoProposalDetails
	for _, p := range ns.ProtocolDaoProposalDetails {
		if p.State == types.ProtocolDaoProposalState_Pending && p.ID != 0 {
			pendingProp = p
			break
		}
	}

	localRoot := makeLocalRoot()
	localTree := &proposals.NetworkVotingTree{
		VotingTree: &proposals.VotingTree{
			Nodes: []*types.VotingTreeNode{&localRoot},
		},
	}

	// Events for both indices 1 (root) and 2 (child responded)
	rootEvent := protocol.RootSubmitted{
		ProposalID: big.NewInt(int64(pendingProp.ID)),
		Proposer:   pendingProp.ProposerAddress,
		Index:      big.NewInt(1),
		Root:       makeMismatchedOnChainRoot(),
	}
	childEvent := protocol.RootSubmitted{
		ProposalID: big.NewInt(int64(pendingProp.ID)),
		Proposer:   pendingProp.ProposerAddress,
		Index:      big.NewInt(2),
		Root:       types.VotingTreeNode{Sum: big.NewInt(500), Hash: common.HexToHash("0xcccc")},
	}

	// Index 2 has been responded to, so the flow should continue to index 4
	csGetter := &stubChallengeStateGetter{
		states: map[[2]uint64]types.ChallengeState{
			{pendingProp.ID, 2}: types.ChallengeState_Responded,
		},
	}

	// The artifact checker returns different results depending on call count
	callCount := 0
	artifactChecker := &sequentialArtifactChecker{
		results: []artifactResult{
			{childIndex: 2, node: types.VotingTreeNode{Sum: big.NewInt(500), Hash: common.HexToHash("0xcccc")}},
			{childIndex: 4, node: types.VotingTreeNode{Sum: big.NewInt(250), Hash: common.HexToHash("0xeeee")}},
		},
		callCount: &callCount,
	}

	nodeAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	logger := log.NewColorLogger(0)

	challenges, defeats, err := getChallengesFromState(
		ns, nodeAddr, &logger,
		&stubBeaconClient{block: beacon.BeaconBlock{ExecutionBlockNumber: ns.ElBlockNumber}},
		&stubNodeGetter{root: makeMismatchedOnChainRoot()},
		&stubTreeProvider{tree: localTree},
		csGetter,
		&stubEventProvider{events: []protocol.RootSubmitted{rootEvent, childEvent}},
		artifactChecker,
		map[uint64]bool{},
		map[uint64]map[uint64]*protocol.RootSubmitted{},
		nil,
	)
	if err != nil {
		t.Fatalf("getChallengesFromState: %v", err)
	}
	if len(challenges) != 1 {
		t.Fatalf("expected 1 challenge at deeper level, got %d", len(challenges))
	}
	if challenges[0].challengedIndex != 4 {
		t.Errorf("challenge index: got %d, want 4 (deeper level)", challenges[0].challengedIndex)
	}
	if len(defeats) != 0 {
		t.Errorf("expected 0 defeats, got %d", len(defeats))
	}
	t.Logf("Challenge produced at deeper index %d after index 2 was responded to", challenges[0].challengedIndex)
}

// sequentialArtifactChecker returns different results on successive calls,
// simulating the tree traversal going deeper on each iteration.
type sequentialArtifactChecker struct {
	results   []artifactResult
	callCount *int
}

type artifactResult struct {
	childIndex uint64
	node       types.VotingTreeNode
	proof      []types.VotingTreeNode
}

func (s *sequentialArtifactChecker) CheckForChallengeableArtifacts(_ protocol.RootSubmitted) (uint64, types.VotingTreeNode, []types.VotingTreeNode, error) {
	idx := *s.callCount
	*s.callCount++
	if idx >= len(s.results) {
		return 0, types.VotingTreeNode{}, nil, nil
	}
	r := s.results[idx]
	return r.childIndex, r.node, r.proof, nil
}
