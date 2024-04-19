package rewards

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
	merkletree "github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

// Claimer operator rewards
type ClaimerRewardsInfo_v4 struct {
	RewardNetwork    uint64                    `json:"rewardNetwork"`
	CollateralRpl    *sharedtypes.QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *sharedtypes.QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *sharedtypes.QuotedBigInt `json:"smoothingPoolEth"`
	MerkleData       []byte                    `json:"-"`
	MerkleProof      []string                  `json:"merkleProof"`
}

func (i *ClaimerRewardsInfo_v4) GetRewardNetwork() uint64 {
	return i.RewardNetwork
}
func (i *ClaimerRewardsInfo_v4) GetCollateralRpl() *sharedtypes.QuotedBigInt {
	return i.CollateralRpl
}
func (i *ClaimerRewardsInfo_v4) GetOracleDaoRpl() *sharedtypes.QuotedBigInt {
	return i.OracleDaoRpl
}
func (i *ClaimerRewardsInfo_v4) GetSmoothingPoolEth() *sharedtypes.QuotedBigInt {
	return i.SmoothingPoolEth
}
func (n *ClaimerRewardsInfo_v4) GetMerkleProof() ([]common.Hash, error) {
	proof := []common.Hash{}
	for _, proofLevel := range n.MerkleProof {
		proof = append(proof, common.HexToHash(proofLevel))
	}
	return proof, nil
}

// JSON struct for a complete rewards file
type RewardsFile_v4 struct {
	*sharedtypes.RewardsFileHeader
	ClaimerRewards          map[common.Address]*ClaimerRewardsInfo_v4 `json:"claimerRewards"`
	MinipoolPerformanceFile MinipoolPerformanceFile_v3                `json:"-"`
}

// Serialize a rewards file into bytes
func (f *RewardsFile_v4) Serialize() ([]byte, error) {
	return json.Marshal(f)
}

// Deserialize a rewards file from bytes
func (f *RewardsFile_v4) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, &f)
}

// Get the rewards file's header
func (f *RewardsFile_v4) GetHeader() *sharedtypes.RewardsFileHeader {
	return f.RewardsFileHeader
}

// Get all of the claimer addresses with rewards in this file
// NOTE: the order of claimer addresses is not guaranteed to be stable, so don't rely on it
func (f *RewardsFile_v4) GetClaimerAddresses() []common.Address {
	addresses := make([]common.Address, len(f.ClaimerRewards))
	i := 0
	for address := range f.ClaimerRewards {
		addresses[i] = address
		i++
	}
	return addresses
}

// Get info about a claimer's rewards
func (f *RewardsFile_v4) GetClaimerRewardsInfo(address common.Address) (sharedtypes.IClaimerRewardsInfo, bool) {
	rewards, exists := f.ClaimerRewards[address]
	return rewards, exists
}

// Gets the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v4) GetMinipoolPerformanceFile() sharedtypes.IMinipoolPerformanceFile {
	return &f.MinipoolPerformanceFile
}

// Sets the CID of the minipool performance file corresponding to this rewards file
func (f *RewardsFile_v4) SetMinipoolPerformanceFileCID(cid string) {
	f.MinipoolPerformanceFileCID = cid
}

// Generates a merkle tree from the provided rewards map
func (f *RewardsFile_v4) GenerateMerkleTree() error {
	// Generate the leaf data for each claimer
	totalData := make([][]byte, 0, len(f.ClaimerRewards))
	for address, rewardsForClaimers := range f.ClaimerRewards {
		// Ignore claimers that didn't receive any rewards
		if rewardsForClaimers.CollateralRpl.Cmp(common.Big0) == 0 && rewardsForClaimers.OracleDaoRpl.Cmp(common.Big0) == 0 && rewardsForClaimers.SmoothingPoolEth.Cmp(common.Big0) == 0 {
			continue
		}

		// Claimer data is address[20] :: network[32] :: RPL[32] :: ETH[32]
		claimerData := make([]byte, 0, 20+32*3)

		// Claimer address
		addressBytes := address.Bytes()
		claimerData = append(claimerData, addressBytes...)

		// Claimer network
		network := big.NewInt(0).SetUint64(rewardsForClaimers.RewardNetwork)
		networkBytes := make([]byte, 32)
		network.FillBytes(networkBytes)
		claimerData = append(claimerData, networkBytes...)

		// RPL rewards
		rplRewards := big.NewInt(0)
		rplRewards.Add(&rewardsForClaimers.CollateralRpl.Int, &rewardsForClaimers.OracleDaoRpl.Int)
		rplRewardsBytes := make([]byte, 32)
		rplRewards.FillBytes(rplRewardsBytes)
		claimerData = append(claimerData, rplRewardsBytes...)

		// ETH rewards
		ethRewardsBytes := make([]byte, 32)
		rewardsForClaimers.SmoothingPoolEth.FillBytes(ethRewardsBytes)
		claimerData = append(claimerData, ethRewardsBytes...)

		// Assign it to the claimer rewards tracker and add it to the leaf data slice
		rewardsForClaimers.MerkleData = claimerData
		totalData = append(totalData, claimerData)
	}

	// Generate the tree
	tree, err := merkletree.NewUsing(totalData, keccak256.New(), false, true)
	if err != nil {
		return fmt.Errorf("error generating Merkle Tree: %w", err)
	}

	// Generate the proofs for each claimer
	for address, rewardsForClaimer := range f.ClaimerRewards {
		// Get the proof
		proof, err := tree.GenerateProof(rewardsForClaimer.MerkleData, 0)
		if err != nil {
			return fmt.Errorf("error generating proof for claimer %s: %w", address.Hex(), err)
		}

		// Convert the proof into hex strings
		proofStrings := make([]string, len(proof.Hashes))
		for i, hash := range proof.Hashes {
			proofStrings[i] = fmt.Sprintf("0x%s", hex.EncodeToString(hash))
		}

		// Assign the hex strings to the claimer rewards struct
		rewardsForClaimer.MerkleProof = proofStrings
	}

	f.MerkleTree = tree
	f.MerkleRoot = common.BytesToHash(tree.Root()).Hex()
	return nil
}
