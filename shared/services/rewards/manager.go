// Experimenting with an alternate language style - named return params
package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/klauspost/compress/zstd"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

const (
	scanningWindowSize uint64 = 10000
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
		bucketBig := big.NewInt(int64(i))
		bucketBytes := [32]byte{}
		bucketBig.FillBytes(bucketBytes[:])

		var bitmap *big.Int
		bitmap, err = rp.RocketStorage.GetUint(nil, crypto.Keccak256Hash([]byte("rewards.interval.claimed"), nodeAddress.Bytes(), bucketBytes[:]))
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
	var event rewards.RewardsEvent

	// Get the event details for this interval
	event, err = GetRewardSnapshotEvent(rp, cfg, interval)
	if err != nil {
		return
	}

	info.CID = event.MerkleTreeCID
	info.StartTime = event.IntervalStartTime
	info.EndTime = event.IntervalEndTime
	merkleRootCanon := event.MerkleRoot

	// Check if the tree file exists
	info.TreeFilePath = cfg.Smartnode.GetRewardsTreePath(interval, true)
	_, err = os.Stat(info.TreeFilePath)
	if os.IsNotExist(err) {
		info.TreeFileExists = false
		err = nil
		return
	}
	info.TreeFileExists = true

	// Unmarshal it
	fileBytes, err := ioutil.ReadFile(info.TreeFilePath)
	if err != nil {
		err = fmt.Errorf("error reading %s: %w", info.TreeFilePath, err)
		return
	}
	var proofWrapper RewardsFile
	err = json.Unmarshal(fileBytes, &proofWrapper)
	if err != nil {
		err = fmt.Errorf("error deserializing %s: %w", info.TreeFilePath, err)
		return
	}

	// Make sure the Merkle root has the expected value
	merkleRootFromFile := common.HexToHash(proofWrapper.MerkleRoot)
	if merkleRootCanon != merkleRootFromFile {
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

		var proof []common.Hash
		proof, err = rewards.GetMerkleProof()
		if err != nil {
			err = fmt.Errorf("error deserializing merkle proof for %s, node %s: %w", info.TreeFilePath, nodeAddress.Hex(), err)
			return
		}
		info.MerkleProof = proof
	}

	return
}

// Get the event for a rewards snapshot
func GetRewardSnapshotEvent(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, interval uint64) (rewards.RewardsEvent, error) {

	var event rewards.RewardsEvent
	var err error

	// Get the event log interval
	eventLogInterval, err := cfg.GetEventLogInterval()
	if err != nil {
		return rewards.RewardsEvent{}, err
	}

	// Check if the interval is already recorded
	prerecordedIntervals := cfg.Smartnode.GetRewardsSubmissionBlockMaps()
	if uint64(len(prerecordedIntervals)) > interval {
		// This already recorded so just use that block number
		blockNumber := big.NewInt(0).SetUint64(prerecordedIntervals[interval])

		// Get the event details for this interval
		return GetUpgradedRewardSnapshotEvent(cfg, rp, interval, big.NewInt(1), blockNumber, blockNumber)
	} else {
		var latestKnownBlock uint64
		var numberOfIntervalsPassed uint64
		if len(prerecordedIntervals) == 0 {
			// If there aren't any prerecorded intervals, start from the deployment block
			deployBlockHash := crypto.Keccak256Hash([]byte("deploy.block"))
			latestKnownBlockBig, err := rp.RocketStorage.GetUint(nil, deployBlockHash)
			if err != nil {
				return rewards.RewardsEvent{}, fmt.Errorf("error getting Rocket Pool deployment block: %w", err)
			}
			latestKnownBlock = latestKnownBlockBig.Uint64()
			numberOfIntervalsPassed = interval + 1
		} else {
			// Grab the latest known one - there will always be at least one of these
			latestKnownInterval := len(prerecordedIntervals) - 1
			latestKnownBlock = prerecordedIntervals[latestKnownInterval]
			numberOfIntervalsPassed = interval - uint64(latestKnownInterval)
		}

		var currentBlock *types.Header
		currentBlock, err = rp.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return event, err
		}

		// Get the current interval time
		var intervalTime time.Duration
		intervalTime, err = rewards.GetClaimIntervalTime(rp, nil)
		if err != nil {
			err = fmt.Errorf("error getting claim interval time: %w", err)
			return event, err
		}

		// Get the time of the latest block
		var latestKnownBlockHeader *types.Header
		latestKnownBlockHeader, err = rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(latestKnownBlock)))
		if err != nil {
			return event, err
		}

		// Traverse multiples of the interval until we find it
		headerToCheck := latestKnownBlockHeader
		timeToCheck := time.Unix(int64(latestKnownBlockHeader.Time), 0).Add(intervalTime * time.Duration(numberOfIntervalsPassed))
		scanningWindow := big.NewInt(0).SetUint64(scanningWindowSize)
		found := false

		for headerToCheck.Number.Uint64() < currentBlock.Number.Uint64() {
			// Get the approximate next header to check
			headerToCheck, err = GetELBlockHeaderForTime(timeToCheck, rp)
			if err != nil {
				return event, err
			}
			// Scan the window around that block
			startBlock := big.NewInt(0).Sub(headerToCheck.Number, scanningWindow)
			endBlock := big.NewInt(0).Add(headerToCheck.Number, scanningWindow)
			if endBlock.Uint64() > currentBlock.Number.Uint64() {
				endBlock = big.NewInt(0).Set(currentBlock.Number)
			}
			event, err = GetUpgradedRewardSnapshotEvent(cfg, rp, interval, big.NewInt(int64(eventLogInterval)), startBlock, endBlock)
			if err != nil {
				if err.Error() == fmt.Sprintf("reward snapshot for interval %d not found", interval) {
					// This isn't a great way to check if an event wasn't found, but it'll do for now
					err = nil
					timeToCheck = timeToCheck.Add(intervalTime) // Try the next interval
					continue
				} else {
					return event, err
				}
			} else {
				found = true
				break
			}

		}

		if !found {
			err = fmt.Errorf("rewards event for interval %d could not be found", interval)
			return event, err
		}
	}

	return event, nil

}

// Get the number of the latest EL block that was created before the given timestamp
func GetELBlockHeaderForTime(targetTime time.Time, rp *rocketpool.RocketPool) (*types.Header, error) {

	// Get the latest block's timestamp
	latestBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting latest block header: %w", err)
	}
	latestBlock := latestBlockHeader.Number

	// Get the block that Rocket Pool deployed to the chain on, use that as the search start
	deployBlockHash := crypto.Keccak256Hash([]byte("deploy.block"))
	deployBlock, err := rp.RocketStorage.GetUint(nil, deployBlockHash)
	if err != nil {
		return nil, fmt.Errorf("error getting Rocket Pool deployment block: %w", err)
	}

	// Get half the distance between the protocol deployment and right now
	delta := big.NewInt(0).Sub(latestBlock, deployBlock)
	delta.Div(delta, big.NewInt(2))

	// Start at the halfway point
	candidateBlockNumber := big.NewInt(0).Sub(latestBlock, delta)
	candidateBlock, err := rp.Client.HeaderByNumber(context.Background(), candidateBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("error getting EL block %d: %w", candidateBlock, err)
	}
	bestBlock := candidateBlock
	pivotSize := candidateBlock.Number.Uint64()
	minimumDistance := +math.Inf(1)
	targetTimeUnix := float64(targetTime.Unix())

	for {
		// Get the distance from the candidate block to the target time
		candidateTime := float64(candidateBlock.Time)
		delta := targetTimeUnix - candidateTime
		distance := math.Abs(delta)

		// If it's better, replace the best candidate with it
		if distance < minimumDistance {
			minimumDistance = distance
			bestBlock = candidateBlock
		} else if pivotSize == 1 {
			// If the pivot is down to size 1 and we didn't find anything better after another iteration, this is the best block!
			for candidateTime > targetTimeUnix {
				// Get the previous block if this one happened after the target time
				candidateBlockNumber.Sub(candidateBlockNumber, big.NewInt(1))
				candidateBlock, err = rp.Client.HeaderByNumber(context.Background(), candidateBlockNumber)
				if err != nil {
					return nil, fmt.Errorf("error getting EL block %d: %w", candidateBlock, err)
				}
				candidateTime = float64(candidateBlock.Time)
				bestBlock = candidateBlock
			}
			return bestBlock, nil
		}

		// Iterate over the correct half, setting the pivot to the halfway point of that half (rounded up)
		pivotSize = uint64(math.Ceil(float64(pivotSize) / 2))
		if delta < 0 {
			// Go left
			candidateBlockNumber.Sub(candidateBlockNumber, big.NewInt(int64(pivotSize)))
		} else {
			// Go right
			candidateBlockNumber.Add(candidateBlockNumber, big.NewInt(int64(pivotSize)))
		}

		// Clamp the new candidate to the latest block
		if candidateBlockNumber.Uint64() > (latestBlock.Uint64() - 1) {
			candidateBlockNumber.SetUint64(latestBlock.Uint64() - 1)
		}

		candidateBlock, err = rp.Client.HeaderByNumber(context.Background(), candidateBlockNumber)
		if err != nil {
			return nil, fmt.Errorf("error getting EL block %d: %w", candidateBlock, err)
		}
	}
}

// Downloads a single rewards file
func DownloadRewardsFile(cfg *config.RocketPoolConfig, interval uint64, cid string, isDaemon bool) error {

	// Determine file name and path
	rewardsTreePath, err := homedir.Expand(cfg.Smartnode.GetRewardsTreePath(interval, isDaemon))
	if err != nil {
		return fmt.Errorf("error expanding rewards tree path: %w", err)
	}
	rewardsTreeFilename := filepath.Base(rewardsTreePath)
	ipfsFilename := rewardsTreeFilename + config.RewardsTreeIpfsExtension

	// Create URL list
	urls := []string{
		fmt.Sprintf(config.PrimaryRewardsFileUrl, cid, ipfsFilename),
		fmt.Sprintf(config.SecondaryRewardsFileUrl, cid, ipfsFilename),
	}

	// Attempt downloads
	errBuilder := strings.Builder{}
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			errBuilder.WriteString(fmt.Sprintf("Downloading %s failed (%s)\n", url, err.Error()))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errBuilder.WriteString(fmt.Sprintf("Downloading %s failed with status %s\n", url, resp.Status))
			continue
		} else {
			// If we got here, we have a successful download
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("Error reading response bytes from %s: %s\n", url, err.Error()))
				continue
			}

			// Decompress it
			decompressedBytes, err := decompressFile(bytes)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("Error decompressing %s: %s\n", url, err.Error()))
				continue
			}

			// Write the file
			err = ioutil.WriteFile(rewardsTreePath, decompressedBytes, 0644)
			if err != nil {
				return fmt.Errorf("error saving interval %d file to %s: %w", interval, rewardsTreePath, err)
			}
			return nil
		}
	}

	return fmt.Errorf(errBuilder.String())

}

// Decompresses a rewards file
func decompressFile(compressedBytes []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating compression decoder: %w", err)
	}

	decompressedBytes, err := decoder.DecodeAll(compressedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("error decompressing rewards file: %w", err)
	}

	return decompressedBytes, nil
}
