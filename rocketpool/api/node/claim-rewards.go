package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func getRewardsInfo(c *cli.Context) (*api.NodeGetRewardsInfoResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeGetRewardsInfoResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Make the map of eligible intervals
	response.Intervals = map[uint64]api.IntervalInfo{}

	// Get the current interval
	currentIndex, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return nil, err
	}

	// Get the claim status of every interval that's happened so far
	indexInt := currentIndex.Uint64() // This is guaranteed to be from 0 to 65535 so the conversion is legal
	if indexInt == 0 {
		// If we're still in the first interval, there's nothing to report.
		return &response, nil
	}

	bucket := indexInt / 256
	for i := uint64(0); i <= bucket; i++ {
		bitmap, err := rewards.ClaimedBitMap(rp, nodeAccount.Address, big.NewInt(0).SetUint64(i), nil)
		if err != nil {
			return nil, err
		}

		one := big.NewInt(1)
		for j := uint64(0); j < 256; j++ {
			mask := big.NewInt(0)
			mask.Lsh(one, uint(j))
			maskedBitmap := big.NewInt(0)
			maskedBitmap.And(bitmap, mask)

			if maskedBitmap.Cmp(mask) != 0 {
				// This bit was not flipped, so this period has not been claimed yet
				targetIndex := i*256 + j
				response.Intervals[targetIndex] = api.IntervalInfo{
					Index: targetIndex,
				}
			}
		}
	}

	// Populate the interval info for each one
	for index, intervalInfo := range response.Intervals {

		// Check if the tree file exists
		path := cfg.Smartnode.GetRewardsTreePath(index)
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			intervalInfo.TreeFileExists = false
			continue
		}
		intervalInfo.TreeFileExists = true

		// Unmarshal it
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", path, err)
		}
		var proofWrapper types.ProofWrapper
		err = json.Unmarshal(bytes, &proofWrapper)
		if err != nil {
			return nil, fmt.Errorf("error deserializing %s: %w", path, err)
		}

		// Get the rewards from it
		rewards, exists := proofWrapper.NodeRewards[nodeAccount.Address]
		if exists {
			intervalInfo.CollateralRplAmount = rewards.CollateralRpl
			intervalInfo.ODaoRplAmount = rewards.OracleDaoRpl
			intervalInfo.SmoothingPoolEthAmount = rewards.SmoothingPoolEth
			proof, err := rewards.GetMerkleProof()
			if err != nil {
				return nil, fmt.Errorf("error deserializing merkle proof for %s, node %s: %w", path, nodeAccount.Address.Hex(), err)
			}
			intervalInfo.MerkleProof = proof
		}
	}

	// Get collateral info for restaking
	var totalMinipools int
	var finalizedMinipools int
	details, err := getNodeMinipoolCountDetails(rp, nodeAccount.Address)
	if err == nil {
		totalMinipools = len(details)
		for _, mpDetails := range details {
			if mpDetails.Finalised {
				finalizedMinipools++
			}
		}
	}
	response.RplStake, err = node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.RplPrice, err = network.GetRPLPrice(rp, nil)
	if err != nil {
		return nil, err
	}
	response.ActiveMinipools = totalMinipools - finalizedMinipools

	return &response, nil
}

func canClaimRewards(c *cli.Context, indicesString string) (*api.CanNodeClaimRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeClaimRewardsResponse{}

	// Get the rewards
	indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervals(c, indicesString)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := rewards.EstimateClaimGas(rp, indices, amountRPL, amountETH, merkleProofs, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	return &response, nil

}

func claimRewards(c *cli.Context, indicesString string) (*api.NodeClaimRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeClaimRewardsResponse{}

	// Get the rewards
	indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervals(c, indicesString)
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Claim rewards
	hash, err := rewards.Claim(rp, indices, amountRPL, amountETH, merkleProofs, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canClaimAndStakeRewards(c *cli.Context, indicesString string, stakeAmount *big.Int) (*api.CanNodeClaimAndStakeRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeClaimAndStakeRewardsResponse{}

	// Get the rewards
	indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervals(c, indicesString)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := rewards.EstimateClaimAndStakeGas(rp, indices, amountRPL, amountETH, merkleProofs, stakeAmount, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	return &response, nil

}

func claimAndStakeRewards(c *cli.Context, indicesString string, stakeAmount *big.Int) (*api.NodeClaimAndStakeRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeClaimAndStakeRewardsResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the rewards
	indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervals(c, indicesString)
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Claim rewards
	hash, err := rewards.ClaimAndStake(rp, indices, amountRPL, amountETH, merkleProofs, stakeAmount, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

// Get the rewards for the provided interval indices
func getRewardsForIntervals(c *cli.Context, indicesString string) ([]*big.Int, []*big.Int, []*big.Int, [][][]byte, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get the indices
	elements := strings.Split(indicesString, ",")
	indices := []*big.Int{}
	for _, element := range elements {
		index, err := strconv.ParseUint(element, 0, 64)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("cannot convert index %s to a number: %w", element, err)
		}
		indices = append(indices, big.NewInt(0).SetUint64(index))
	}

	// Read the tree files to get the details
	amountRPL := []*big.Int{}
	amountETH := []*big.Int{}
	merkleProofs := [][][]byte{}

	// Populate the interval info for each one
	for _, index := range indices {

		// Check if the tree file exists
		path := cfg.Smartnode.GetRewardsTreePath(index.Uint64())
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			return nil, nil, nil, nil, fmt.Errorf("rewards tree file '%s' doesn't exist", path)
		}

		// Unmarshal it
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error reading %s: %w", path, err)
		}
		var proofWrapper types.ProofWrapper
		err = json.Unmarshal(bytes, &proofWrapper)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error deserializing %s: %w", path, err)
		}

		// Get the rewards from it
		rewards, exists := proofWrapper.NodeRewards[nodeAccount.Address]
		if exists {
			// Append RPL
			rpl := big.NewInt(0)
			rpl.Add(rpl, rewards.CollateralRpl)
			rpl.Add(rpl, rewards.OracleDaoRpl)
			amountRPL = append(amountRPL, rpl)

			// Append ETH
			amountETH = append(amountETH, rewards.SmoothingPoolEth)

			// Append Merkle proof
			proof, err := rewards.GetMerkleProof()
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("error deserializing merkle proof for %s, node %s: %w", path, nodeAccount.Address.Hex(), err)
			}
			merkleProofs = append(merkleProofs, proof)
		}
	}

	// Return
	return indices, amountRPL, amountETH, merkleProofs, nil

}
