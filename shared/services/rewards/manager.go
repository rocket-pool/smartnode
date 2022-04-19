// Experimenting with an alternate language style - named return params
package rewards

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

// Gets the intervals the node can claim and the intervals that have already been claimed
func GetClaimStatus(rp *rocketpool.RocketPool, nodeAddress common.Address) (unclaimed []uint64, claimed []uint64, err error) {
	// Get the current interval
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return
	}

	currentIndex := currentIndexBig.Uint64() // This is guaranteed to be from 0 to 65535 so the conversion is legal
	if currentIndex == 0 {
		// If we're still in the first interval, there's nothing to report.
		return
	}

	// Get the claim status of every interval that's happened so far
	one := big.NewInt(1)
	bucket := currentIndex / 256
	for i := uint64(0); i <= bucket; i++ {
		var bitmap *big.Int
		bitmap, err = rewards.ClaimedBitMap(rp, nodeAddress, big.NewInt(0).SetUint64(i), nil)
		if err != nil {
			return
		}

		for j := uint64(0); j < 256; j++ {
			targetIndex := i*256 + j
			if targetIndex >= currentIndex {
				// End once we've hit the current interval
				break
			}

			mask := big.NewInt(0)
			mask.Lsh(one, uint(j))
			maskedBitmap := big.NewInt(0)
			maskedBitmap.And(bitmap, mask)

			if maskedBitmap.Cmp(mask) == 0 {
				// This bit was flipped, so it's been claimed already
				claimed = append(claimed, targetIndex)
			} else {
				// This bit was not flipped, so it hasn't been claimed yet
				unclaimed = append(unclaimed, targetIndex)
			}
		}
	}

	return
}

// Gets the information for an interval including the file status, the validity, and the node's rewards
func GetIntervalInfo(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, nodeAddress common.Address, interval uint64) (info IntervalInfo, err error) {
	info.Index = interval

	// Check if the tree file exists
	info.TreeFilePath = cfg.Smartnode.GetRewardsTreePath(interval)
	_, err = os.Stat(info.TreeFilePath)
	if os.IsNotExist(err) {
		info.TreeFileExists = false
		return
	}
	info.TreeFileExists = true

	// Unmarshal it
	fileBytes, err := ioutil.ReadFile(info.TreeFilePath)
	if err != nil {
		err = fmt.Errorf("error reading %s: %w", info.TreeFilePath, err)
		return
	}
	var proofWrapper ProofWrapper
	err = json.Unmarshal(fileBytes, &proofWrapper)
	if err != nil {
		err = fmt.Errorf("error deserializing %s: %w", info.TreeFilePath, err)
		return
	}

	// Make sure the Merkle root has the expected value
	merkleRootCanon, err := rewards.MerkleRoots(rp, big.NewInt(int64(interval)), nil)
	if err != nil {
		return
	}
	merkleRootFromFile := common.Hex2Bytes(hexutil.RemovePrefix(proofWrapper.MerkleRoot))
	if !bytes.Equal(merkleRootCanon, merkleRootFromFile) {
		info.MerkleRootValid = false
		return
	}
	info.MerkleRootValid = true

	// Get the rewards from it
	rewards, exists := proofWrapper.NodeRewards[nodeAddress]
	info.NodeExists = exists
	if exists {
		info.CollateralRplAmount = rewards.CollateralRpl
		info.ODaoRplAmount = rewards.OracleDaoRpl
		info.SmoothingPoolEthAmount = rewards.SmoothingPoolEth

		var proof [][]byte
		proof, err = rewards.GetMerkleProof()
		if err != nil {
			err = fmt.Errorf("error deserializing merkle proof for %s, node %s: %w", info.TreeFilePath, nodeAddress.Hex(), err)
			return
		}
		info.MerkleProof = proof
	}

	return
}
