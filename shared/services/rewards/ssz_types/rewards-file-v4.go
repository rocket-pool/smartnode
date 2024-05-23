package ssz_types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	stdbig "math/big"
	"slices"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types/big"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

type Format = uint

const (
	FormatJSON = iota
	FormatSSZ
)

var Magic [4]byte = [4]byte{0x52, 0x50, 0x52, 0x54}

type Address [20]byte
type Hash [32]byte
type NetworkRewards []*NetworkReward
type NodeRewards []*NodeReward
type Network uint64
type MerkleProof []Hash

type SSZFile_v1 struct {
	// Fields specific to ssz encoding are first

	// A magic header. Four bytes. Helps immediately verify what follows is a rewards tree.
	// 0x52505254 - it's RPRT in ASCII and easy to recognize
	Magic [4]byte `ssz-size:"4" json:"-"`
	// Version is first- parsers can check the first 12 bytes of the file to make sure they're
	// parsing a rewards tree and it is a version they know how to parse.
	RewardsFileVersion uint64 `json:"rewardsFileVersion"`

	// Next, we need fields for the rest of the RewardsFileHeader

	// RulesetVersion is the version of the ruleset used to generate the tree, e.g., v9 for the first
	// ruleset to use ssz
	RulesetVersion uint64 `json:"rulesetVersion"`
	// Network is the chain id for which the tree is generated
	Network Network `json:"network"`
	// Index is the rewards interval index
	Index uint64 `json:"index"`
	// StartTime is the time of the first slot of the interval
	StartTime time.Time `json:"startTime"`
	// EndTime is the time fo the last slot of the interval
	EndTime time.Time `json:"endTime"`
	// ConsensusStartBlock is the first non-empty slot of the interval
	ConsensusStartBlock uint64 `json:"consensusStartBlock,omitempty"`
	// ConsensusEndBlock is the last non-empty slot of the interval
	ConsensusEndBlock uint64 `json:"consensusEndBlock"`
	// ExecutionBlock is the execution block number included in ConsensusStartBlock
	ExecutionStartBlock uint64 `json:"executionStartBlock,omitempty"`
	// ExecutionEndBlock is the execution block number included in ConsensusEndBlock
	ExecutionEndBlock uint64 `json:"executionEndBlock"`
	// IntervalsPassed is the number of rewards intervals contained in this tree
	IntervalsPassed uint64 `json:"intervalsPassed"`
	// MerkleRoot is the root of the merkle tree of all the nodes in this tree.
	MerkleRoot Hash `ssz-size:"32" json:"merkleRoot,omitempty"`
	// TotalRewards is aggregate data on how many rewards this tree contains
	TotalRewards *TotalRewards `json:"totalRewards"`
	// NetworkRewards is the destinations and aggregate amounts for each network
	// this tree distributes to.
	// Must be sorted by Chain ID ascending
	NetworkRewards NetworkRewards `ssz-max:"128" json:"networkRewards"`

	// Finally, the actual per-node objects that get merkle-ized

	// NodeRewards are the objects that make up the merkle tree.
	// Must be sorted by Node Address ascending
	NodeRewards NodeRewards `ssz-max:"9223372036854775807" json:"nodeRewards"`

	merkleProofs map[Address]MerkleProof `ssz:"-" json:"-"`
}

func NewSSZFile_v1() *SSZFile_v1 {
	return &SSZFile_v1{
		Magic: Magic,
	}
}

// Check if the NodeRewards field respects unique constraints
func (f *SSZFile_v1) nodeRewardsUnique() bool {
	m := make(map[Address]any, len(f.NodeRewards))

	for _, nr := range f.NodeRewards {
		_, found := m[nr.Address]
		if found {
			return false
		}
		m[nr.Address] = struct{}{}
	}

	return true
}

// Check if the NetworkRewards field respects unique constraints
func (f *SSZFile_v1) networkRewardsUnique() bool {
	m := make(map[uint64]any, len(f.NetworkRewards))

	for _, nr := range f.NetworkRewards {
		_, found := m[nr.Network]
		if found {
			return false
		}
		m[nr.Network] = struct{}{}
	}

	return true
}

// Verify checks that the arrays in the file are appropriately sorted and that
// the merkle proof, if present, matches.
func (f *SSZFile_v1) Verify() error {
	if !sort.IsSorted(f.NodeRewards) {
		return errors.New("ssz file node rewards out of order")
	}

	if !sort.IsSorted(f.NetworkRewards) {
		return errors.New("ssz file network rewards out of order")
	}

	if !f.nodeRewardsUnique() {
		return errors.New("ssz file has duplicate entries in its NodeRewards field")
	}

	if !f.networkRewardsUnique() {
		return errors.New("ssz file has duplicate entries in its NetworkRewards field")
	}

	if f.TotalRewards == nil {
		return errors.New("missing required field TotalRewards")
	}

	if _, err := f.Proofs(); err != nil {
		return err
	}

	return nil
}

// Minipool Performance CID is deprecated, but we must implement this for the interface
func (f *SSZFile_v1) SetMinipoolPerformanceFileCID(cid string) {
}

// The "normal" serialize() call is expected to be JSON by ISerializable in files.go
func (f *SSZFile_v1) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

// Write as SSZ
func (f *SSZFile_v1) SerializeSSZ() ([]byte, error) {
	return f.FinalizeSSZ()
}

func (f *SSZFile_v1) GenerateMerkleTree() error {
	_, err := f.Proofs()
	return err
}

// Marshal wrappers that adds the magic header if absent and sets or validators merkle root
func (f *SSZFile_v1) FinalizeSSZ() ([]byte, error) {

	return f.FinalizeSSZTo(make([]byte, 0, f.SizeSSZ()))
}

func (f *SSZFile_v1) FinalizeSSZTo(buf []byte) ([]byte, error) {
	copy(f.Magic[:], Magic[:])
	if err := f.Verify(); err != nil {
		return nil, err
	}

	return f.MarshalSSZTo(buf)
}

// Parsing wrapper that adds verification to the merkle root and magic header
func ParseSSZFile(buf []byte) (*SSZFile_v1, error) {
	if !bytes.HasPrefix(buf, Magic[:]) {
		return nil, errors.New("magic header not found in reward ssz file")
	}

	f := &SSZFile_v1{}
	if err := f.UnmarshalSSZ(buf); err != nil {
		return nil, err
	}

	if err := f.Verify(); err != nil {
		return nil, err
	}

	return f, nil
}

// This getter lazy-computes the proofs and caches them on the file
func (f *SSZFile_v1) Proofs() (map[Address]MerkleProof, error) {
	if f.merkleProofs != nil {
		return f.merkleProofs, nil
	}

	sort.Sort(f.NodeRewards)
	sort.Sort(f.NetworkRewards)

	nodeDataMap := make(map[Address][]byte, len(f.NodeRewards))
	treeData := make([][]byte, 0, len(f.NodeRewards))
	for _, nr := range f.NodeRewards {
		// 20 bytes for address, 32 each for network/rpl/eth
		address := nr.Address
		network := uint256.NewInt(nr.Network).Bytes32()
		rpl := stdbig.NewInt(0)
		rpl.Add(rpl, nr.CollateralRpl.Int)
		rpl.Add(rpl, nr.OracleDaoRpl.Int)
		rplBytes := make([]byte, 32)
		rplBytes = rpl.FillBytes(rplBytes)
		eth, err := nr.SmoothingPoolEth.Bytes32()
		if err != nil {
			return nil, fmt.Errorf("error converting big.Int to uint256 byte slice: %w", err)
		}

		const dataSize = 20 + 32*3
		nodeData := make([]byte, dataSize)
		copy(nodeData[0:20], address[:])
		copy(nodeData[20:20+32], network[:])
		copy(nodeData[20+32:20+32*2], rplBytes[:])
		copy(nodeData[20+32*2:20+32*3], eth[:])

		treeData = append(treeData, nodeData)
		nodeDataMap[nr.Address] = nodeData
	}

	tree, err := merkletree.NewUsing(treeData, keccak256.New(), false, true)
	if err != nil {
		return nil, fmt.Errorf("error generating Merkle Tree: %w", err)
	}

	// Generate the proofs
	out := make(map[Address]MerkleProof)
	f.merkleProofs = out
	for address, nodeData := range nodeDataMap {
		proof, err := tree.GenerateProof(nodeData, 0)
		if err != nil {
			return nil, fmt.Errorf("error generating proof for node 0x%s: %w", hex.EncodeToString(address[:]), err)
		}

		// Store the proof in the result map
		out[address] = make([]Hash, len(proof.Hashes))
		for i, hash := range proof.Hashes {
			out[address][i] = Hash{}
			copy(out[address][i][:], hash)
		}
	}

	// Populate missing proofs at node level
	for _, nr := range f.NodeRewards {
		if nr.MerkleProof == nil {
			nr.MerkleProof = out[nr.Address]
		}
	}

	// Finally, set the root. If it's already set, and differs, return an error.
	root := Hash{}
	copy(root[:], tree.Root())
	if bytes.Count(f.MerkleRoot[:], []byte{0x00}) >= 32 {
		f.MerkleRoot = root
		return out, nil
	}

	if !bytes.Equal(f.MerkleRoot[:], root[:]) {
		return nil, fmt.Errorf("generated root %s mismatch against existing root %s", root, f.MerkleRoot)
	}

	// The existing root matches the calculated root
	return out, nil
}

type TotalRewards struct {
	// Total amount of RPL sent to the pDAO
	ProtocolDaoRpl big.Uint256 `ssz-size:"32" json:"protocolDaoRpl"`
	// Total amount of RPL sent to Node Operators
	TotalCollateralRpl big.Uint256 `ssz-size:"32" json:"totalCollateralRpl"`
	// Total amount of RPL sent to the oDAO
	TotalOracleDaoRpl big.Uint256 `ssz-size:"32" json:"totalOracleDaoRpl"`
	// Total amount of ETH in the Smoothing Pool
	TotalSmoothingPoolEth big.Uint256 `ssz-size:"32" json:"totalSmoothingPoolEth"`
	// Total amount of Eth sent to the rETH contract
	PoolStakerSmoothingPoolEth big.Uint256 `ssz-size:"32" json:"poolStakerSmoothingPoolEth"`
	// Total amount of Eth sent to Node Operators in the Smoothing Pool
	NodeOperatorSmoothingPoolEth big.Uint256 `ssz-size:"32" json:"nodeOperatorSmoothingPoolEth"`
	// Total Node Weight as defined by RPIP-30
	TotalNodeWeight big.Uint256 `ssz-size:"32" json:"totalNodeWeight,omitempty"`
}

type NetworkReward struct {
	// Chain ID (key)
	Network uint64 `json:"-"`

	// Amount of RPL sent to the network for Node Operators
	CollateralRpl big.Uint256 `ssz-size:"32" json:"collateralRpl"`
	// Amount of RPL sent to the network for oDAO members
	OracleDaoRpl big.Uint256 `ssz-size:"32" json:"oracleDaoRpl"`
	// Amount of Eth sent to the network for Node Operators
	SmoothingPoolEth big.Uint256 `ssz-size:"32" json:"smoothingPoolEth"`
}

// NetworkRewards should implement sort.Interface to make it easier to sort.
func (n NetworkRewards) Len() int {
	return len(n)
}

func (n NetworkRewards) Less(i, j int) bool {
	return n[i].Network < n[j].Network
}

func (n NetworkRewards) Swap(i, j int) {
	tmp := n[i]
	n[i] = n[j]
	n[j] = tmp
}

type NodeReward struct {
	// Address of the Node (key)
	Address Address `ssz-size:"20" json:"-"`

	// Chain ID on which the Node will claim
	Network uint64 `json:"rewardNetwork"`
	// Amount of staking RPL earned by the Node
	CollateralRpl big.Uint256 `ssz-size:"32" json:"collateralRpl"`
	// Amount of oDAO RPL earned by the Node
	OracleDaoRpl big.Uint256 `ssz-size:"32" json:"oracleDaoRpl"`
	// Amount of Smoothing Pool ETH earned by the Node
	SmoothingPoolEth big.Uint256 `ssz-size:"32" json:"smoothingPoolEth"`
	// Merkle proof for the node claim, sorted with the Merkle root last
	MerkleProof MerkleProof `ssz:"-" json:"merkleProof"`
}

// NodeRewards should implement sort.Interface to make it easier to sort.
func (n NodeRewards) Len() int {
	return len(n)
}

func (n NodeRewards) Less(i, j int) bool {
	ia := n[i].Address
	ja := n[j].Address

	if bytes.Compare(ia[:], ja[:]) < 0 {
		return true
	}

	return false
}

func (n NodeRewards) Swap(i, j int) {
	tmp := n[i]
	n[i] = n[j]
	n[j] = tmp
}

func (n NodeRewards) Find(addr Address) *NodeReward {
	idx := slices.IndexFunc(n, func(nr *NodeReward) bool {
		return bytes.Equal(nr.Address[:], addr[:])
	})
	if idx == -1 {
		return nil
	}
	return n[idx]
}

// Functions to implement IRewardsFile
func (f *SSZFile_v1) Deserialize(data []byte) error {
	if bytes.HasPrefix(data, Magic[:]) {
		if err := f.UnmarshalSSZ(data); err != nil {
			return err
		}

		return f.Verify()
	}

	return json.Unmarshal(data, f)
}

func (f *SSZFile_v1) GetIndex() uint64 {
	return f.Index
}

func (f *SSZFile_v1) GetMerkleRoot() string {
	return f.MerkleRoot.String()
}

func (f *SSZFile_v1) GetNodeAddresses() []common.Address {
	out := make([]common.Address, 0, len(f.NodeRewards))

	for _, nr := range f.NodeRewards {
		out = append(out, common.BytesToAddress(nr.Address[:]))
	}
	return out
}

func (f *SSZFile_v1) GetConsensusEndBlock() uint64 {
	return f.ConsensusEndBlock
}

func (f *SSZFile_v1) GetExecutionEndBlock() uint64 {
	return f.ExecutionEndBlock
}

func (f *SSZFile_v1) GetIntervalsPassed() uint64 {
	return f.IntervalsPassed
}

func (f *SSZFile_v1) GetMerkleProof(address common.Address) ([]common.Hash, error) {
	proofs, err := f.Proofs()
	if err != nil {
		return nil, fmt.Errorf("error while calculating proof for %s: %w", address.String(), err)
	}

	var nativeAddress Address
	copy(nativeAddress[:], address[:])
	nativeProofs := proofs[nativeAddress]
	out := make([]common.Hash, 0, len(nativeProofs))
	for _, p := range nativeProofs {
		var h common.Hash
		copy(h[:], p[:])
		out = append(out, h)
	}

	return out, nil
}

func (f *SSZFile_v1) getRewardsForNetwork(network uint64) *NetworkReward {
	for _, nr := range f.NetworkRewards {
		if nr.Network == network {
			return nr
		}
	}
	return nil
}

func (f *SSZFile_v1) HasRewardsForNetwork(network uint64) bool {
	return f.getRewardsForNetwork(network) != nil
}

func (f *SSZFile_v1) GetNetworkCollateralRpl(network uint64) *stdbig.Int {
	nr := f.getRewardsForNetwork(network)
	if nr == nil {
		return stdbig.NewInt(0)
	}

	return nr.CollateralRpl.Int
}

func (f *SSZFile_v1) GetNetworkOracleDaoRpl(network uint64) *stdbig.Int {
	nr := f.getRewardsForNetwork(network)
	if nr == nil {
		return stdbig.NewInt(0)
	}

	return nr.OracleDaoRpl.Int
}

func (f *SSZFile_v1) GetNetworkSmoothingPoolEth(network uint64) *stdbig.Int {
	nr := f.getRewardsForNetwork(network)
	if nr == nil {
		return stdbig.NewInt(0)
	}

	return nr.SmoothingPoolEth.Int
}

func (f *SSZFile_v1) getNodeRewards(addr common.Address) *NodeReward {
	var nativeAddress Address
	copy(nativeAddress[:], addr[:])
	return f.NodeRewards.Find(nativeAddress)
}

func (f *SSZFile_v1) HasRewardsFor(addr common.Address) bool {
	return f.getNodeRewards(addr) != nil
}

func (f *SSZFile_v1) GetNodeCollateralRpl(addr common.Address) *stdbig.Int {
	nr := f.getNodeRewards(addr)
	if nr == nil {
		return stdbig.NewInt(0)
	}

	return nr.CollateralRpl.Int
}

func (f *SSZFile_v1) GetNodeOracleDaoRpl(addr common.Address) *stdbig.Int {
	nr := f.getNodeRewards(addr)
	if nr == nil {
		return stdbig.NewInt(0)
	}

	return nr.OracleDaoRpl.Int
}

func (f *SSZFile_v1) GetNodeSmoothingPoolEth(addr common.Address) *stdbig.Int {
	nr := f.getNodeRewards(addr)
	if nr == nil {
		return stdbig.NewInt(0)
	}

	return nr.SmoothingPoolEth.Int
}

func (f *SSZFile_v1) GetRewardsFileVersion() uint64 {
	return f.RewardsFileVersion
}

func (f *SSZFile_v1) GetTotalCollateralRpl() *stdbig.Int {
	return f.TotalRewards.TotalCollateralRpl.Int
}

func (f *SSZFile_v1) GetTotalNodeOperatorSmoothingPoolEth() *stdbig.Int {
	return f.TotalRewards.NodeOperatorSmoothingPoolEth.Int
}

func (f *SSZFile_v1) GetTotalNodeWeight() *stdbig.Int {
	return f.TotalRewards.TotalNodeWeight.Int
}

func (f *SSZFile_v1) GetTotalOracleDaoRpl() *stdbig.Int {
	return f.TotalRewards.TotalOracleDaoRpl.Int
}

func (f *SSZFile_v1) GetTotalPoolStakerSmoothingPoolEth() *stdbig.Int {
	return f.TotalRewards.PoolStakerSmoothingPoolEth.Int
}

func (f *SSZFile_v1) GetTotalProtocolDaoRpl() *stdbig.Int {
	return f.TotalRewards.ProtocolDaoRpl.Int
}
