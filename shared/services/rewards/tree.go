package rewards

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
)

// Create the JSON file with the interval rewards and Merkle proof information for each node
func GenerateTreeJson(treeRoot []byte, nodeRewardsMap map[common.Address]NodeRewards, networkRewardsMap map[uint64]NodeRewards) *ProofWrapper {

	totalCollateralRpl := big.NewInt(0)
	totalODaoRpl := big.NewInt(0)
	totalSmoothingPoolEth := big.NewInt(0)

	networkCollateralRplMap := map[uint64]*big.Int{}
	networkODaoRplMap := map[uint64]*big.Int{}
	networkSmoothingPoolEthMap := map[uint64]*big.Int{}

	for network, rewardsForNetwork := range networkRewardsMap {
		networkCollateralRplMap[network] = rewardsForNetwork.CollateralRpl
		networkODaoRplMap[network] = rewardsForNetwork.OracleDaoRpl
		networkSmoothingPoolEthMap[network] = rewardsForNetwork.SmoothingPoolEth

		totalCollateralRpl.Add(totalCollateralRpl, rewardsForNetwork.CollateralRpl)
		totalODaoRpl.Add(totalODaoRpl, rewardsForNetwork.OracleDaoRpl)
		totalSmoothingPoolEth.Add(totalSmoothingPoolEth, rewardsForNetwork.SmoothingPoolEth)
	}

	wrapper := &ProofWrapper{
		MerkleRoot:  fmt.Sprintf("0x%s", hex.EncodeToString(treeRoot)),
		NodeRewards: nodeRewardsMap,
	}
	wrapper.NetworkRewards.CollateralRplPerNetwork = networkCollateralRplMap
	wrapper.NetworkRewards.OracleDaoRplPerNetwork = networkODaoRplMap
	wrapper.NetworkRewards.SmoothingPoolEthPerNetwork = networkSmoothingPoolEthMap
	wrapper.TotalRewards.TotalCollateralRpl = totalCollateralRpl
	wrapper.TotalRewards.TotalOracleDaoRpl = totalODaoRpl
	wrapper.TotalRewards.TotalSmoothingPoolEth = totalSmoothingPoolEth

	return wrapper

}

// Generates a merkle tree from the provided rewards map
func GenerateMerkleTree(nodeRewardsMap map[common.Address]NodeRewards) (*merkletree.MerkleTree, error) {

	// Generate the leaf data for each node
	totalData := make([][]byte, 0, len(nodeRewardsMap))
	for address, rewardsForNode := range nodeRewardsMap {
		// Ignore nodes that didn't receive any rewards
		zero := big.NewInt(0)
		if rewardsForNode.CollateralRpl.Cmp(zero) == 0 && rewardsForNode.OracleDaoRpl.Cmp(zero) == 0 && rewardsForNode.SmoothingPoolEth.Cmp(zero) == 0 {
			continue
		}

		// Node data is address[20] :: network[32] :: RPL[32] :: ETH[32]
		nodeData := make([]byte, 0, 20+32*3)

		// Node address
		addressBytes := address.Bytes()
		nodeData = append(nodeData, addressBytes...)

		// Node network
		network := big.NewInt(0).SetUint64(rewardsForNode.RewardNetwork)
		networkBytes := make([]byte, 32)
		network.FillBytes(networkBytes)
		nodeData = append(nodeData, networkBytes...)

		// RPL rewards
		rplRewards := big.NewInt(0)
		rplRewards.Add(rewardsForNode.CollateralRpl, rewardsForNode.OracleDaoRpl)
		rplRewardsBytes := make([]byte, 32)
		rplRewards.FillBytes(rplRewardsBytes)
		nodeData = append(nodeData, rplRewardsBytes...)

		// ETH rewards
		ethRewardsBytes := make([]byte, 32)
		rewardsForNode.SmoothingPoolEth.FillBytes(ethRewardsBytes)
		nodeData = append(nodeData, ethRewardsBytes...)

		// Assign it to the node rewards tracker and add it to the leaf data slice
		rewardsForNode.MerkleData = nodeData
		nodeRewardsMap[address] = rewardsForNode
		totalData = append(totalData, nodeData)
	}

	// Generate the tree
	tree, err := merkletree.NewUsing(totalData, keccak256.New(), false, true)
	if err != nil {
		return nil, fmt.Errorf("error generating Merkle Tree: %w", err)
	}

	// Generate the proofs for each node
	for address, rewardsForNode := range nodeRewardsMap {
		// Get the proof
		proof, err := tree.GenerateProof(rewardsForNode.MerkleData, 0)
		if err != nil {
			return nil, fmt.Errorf("error generating proof for node %s: %w", address.Hex(), err)
		}

		// Convert the proof into hex strings
		proofStrings := make([]string, len(proof.Hashes))
		for i, hash := range proof.Hashes {
			proofStrings[i] = fmt.Sprintf("0x%s", hex.EncodeToString(hash))
		}

		// Assign the hex strings to the node rewards struct
		rewardsForNode.MerkleProof = proofStrings
		nodeRewardsMap[address] = rewardsForNode
	}

	return tree, nil

}

// Calculates the RPL rewards for the given interval
func CalculateRplRewards(rp *rocketpool.RocketPool, snapshotBlockHeader *types.Header, rewardsInterval time.Duration) (map[common.Address]NodeRewards, map[uint64]NodeRewards, map[common.Address]uint64, error) {

	nodeRewardsMap := map[common.Address]NodeRewards{}
	networkRewardsMap := map[uint64]NodeRewards{}
	invalidNetworkNodes := map[common.Address]uint64{}
	opts := &bind.CallOpts{
		BlockNumber: snapshotBlockHeader.Number,
	}
	snapshotBlockTime := time.Unix(int64(snapshotBlockHeader.Time), 0)

	// Handle node operator rewards
	nodeOpPercent, err := rewards.GetNodeOperatorRewardsPercent(rp, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	pendingRewards, err := rewards.GetPendingRPLRewards(rp, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, eth.EthToWei(1))

	nodeAddresses, err := node.GetNodeAddresses(rp, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	totalRplStake, err := node.GetTotalEffectiveRPLStake(rp, opts)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, address := range nodeAddresses {
		// Make sure this node is eligible for rewards
		regTime, err := node.GetNodeRegistrationTime(rp, address, opts)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error getting registration time for node %s: %w", err)
		}
		if snapshotBlockTime.Sub(regTime) < rewardsInterval {
			continue
		}

		// Get how much RPL goes to this node: effective stake / total stake * total RPL rewards for nodes
		nodeStake, err := node.GetNodeEffectiveRPLStake(rp, address, opts)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error getting effective stake for node %s: %w", address.Hex(), err)
		}
		nodeRplRewards := big.NewInt(0)
		nodeRplRewards.Mul(nodeStake, totalNodeRewards)
		nodeRplRewards.Div(nodeRplRewards, totalRplStake)

		// If there are pending rewards, add it to the map
		if nodeRplRewards.Cmp(big.NewInt(0)) == 1 {
			rewardsForNode, exists := nodeRewardsMap[address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(rp, address, opts)
				if err != nil {
					return nil, nil, nil, err
				}
				if !ValidateNetwork(network) {
					invalidNetworkNodes[address] = network
					continue
				}

				rewardsForNode = NodeRewards{
					RewardNetwork:    network,
					CollateralRpl:    big.NewInt(0),
					OracleDaoRpl:     big.NewInt(0),
					SmoothingPoolEth: big.NewInt(0),
				}
			}
			rewardsForNode.CollateralRpl.Add(rewardsForNode.CollateralRpl, nodeRplRewards)
			nodeRewardsMap[address] = rewardsForNode

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := networkRewardsMap[rewardsForNode.RewardNetwork]
			if !exists {
				rewardsForNetwork = NodeRewards{
					RewardNetwork:    rewardsForNode.RewardNetwork,
					CollateralRpl:    big.NewInt(0),
					OracleDaoRpl:     big.NewInt(0),
					SmoothingPoolEth: big.NewInt(0),
				}
			}
			rewardsForNetwork.CollateralRpl.Add(rewardsForNetwork.CollateralRpl, nodeRplRewards)
			networkRewardsMap[rewardsForNode.RewardNetwork] = rewardsForNetwork
		}
	}

	// Handle Oracle DAO rewards
	oDaoPercent, err := rewards.GetTrustedNodeOperatorRewardsPercent(rp, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, eth.EthToWei(1))

	oDaoAddresses, err := trustednode.GetMemberAddresses(rp, opts)
	if err != nil {
		return nil, nil, nil, err
	}
	memberCount := big.NewInt(int64(len(oDaoAddresses)))
	individualOdaoRewards := big.NewInt(0)
	individualOdaoRewards.Div(totalODaoRewards, memberCount)

	for _, address := range oDaoAddresses {
		rewardsForNode, exists := nodeRewardsMap[address]
		if !exists {
			// Get the network the rewards should go to
			network, err := node.GetRewardNetwork(rp, address, opts)
			if err != nil {
				return nil, nil, nil, err
			}
			if !ValidateNetwork(network) {
				invalidNetworkNodes[address] = network
				continue
			}

			rewardsForNode = NodeRewards{
				RewardNetwork:    network,
				CollateralRpl:    big.NewInt(0),
				OracleDaoRpl:     big.NewInt(0),
				SmoothingPoolEth: big.NewInt(0),
			}
		}
		rewardsForNode.OracleDaoRpl.Add(rewardsForNode.OracleDaoRpl, individualOdaoRewards)
		nodeRewardsMap[address] = rewardsForNode

		// Add the rewards to the running total for the specified network
		rewardsForNetwork, exists := networkRewardsMap[rewardsForNode.RewardNetwork]
		if !exists {
			rewardsForNetwork = NodeRewards{
				RewardNetwork:    rewardsForNode.RewardNetwork,
				CollateralRpl:    big.NewInt(0),
				OracleDaoRpl:     big.NewInt(0),
				SmoothingPoolEth: big.NewInt(0),
			}
		}
		rewardsForNetwork.OracleDaoRpl.Add(rewardsForNetwork.OracleDaoRpl, individualOdaoRewards)
		networkRewardsMap[rewardsForNode.RewardNetwork] = rewardsForNetwork
	}

	// Return the rewards maps
	return nodeRewardsMap, networkRewardsMap, invalidNetworkNodes, nil
}

// Validates that the provided network is legal
func ValidateNetwork(network uint64) bool {

	// TODO: add more of these as we add L2 support
	switch network {
	case 0:
		return true
	default:
		return false
	}

}
