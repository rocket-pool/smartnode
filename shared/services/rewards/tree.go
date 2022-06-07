package rewards

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
	"golang.org/x/sync/errgroup"
)

// Settings
const (
	SmoothingPoolDetailsBatchSize int    = 20
	RewardsFileVersion            uint64 = 1
)

// Create the JSON file with the interval rewards and Merkle proof information for each node
func GenerateTreeJson(treeRoot common.Hash, nodeRewardsMap map[common.Address]NodeRewards, networkRewardsMap map[uint64]NodeRewards, protocolDaoRpl *QuotedBigInt, index uint64, consensusBlock uint64, executionBlock uint64, intervalsPassed uint64) *ProofWrapper {

	totalCollateralRpl := NewQuotedBigInt(0)
	totalODaoRpl := NewQuotedBigInt(0)
	totalSmoothingPoolEth := NewQuotedBigInt(0)

	networkCollateralRplMap := map[uint64]*QuotedBigInt{}
	networkODaoRplMap := map[uint64]*QuotedBigInt{}
	networkSmoothingPoolEthMap := map[uint64]*QuotedBigInt{}

	// Get the highest network index with valid rewards
	highestNetworkIndex := uint64(0)
	for network, _ := range networkRewardsMap {
		if network > highestNetworkIndex {
			highestNetworkIndex = network
		}
	}

	// Create the map for each network, including unused ones
	for network := uint64(0); network <= highestNetworkIndex; network++ {
		rewardsForNetwork, exists := networkRewardsMap[network]
		if !exists {
			rewardsForNetwork = NodeRewards{
				CollateralRpl:    NewQuotedBigInt(0),
				OracleDaoRpl:     NewQuotedBigInt(0),
				SmoothingPoolEth: NewQuotedBigInt(0),
			}
		}

		networkCollateralRplMap[network] = rewardsForNetwork.CollateralRpl
		networkODaoRplMap[network] = rewardsForNetwork.OracleDaoRpl
		networkSmoothingPoolEthMap[network] = rewardsForNetwork.SmoothingPoolEth

		totalCollateralRpl.Add(&totalCollateralRpl.Int, &rewardsForNetwork.CollateralRpl.Int)
		totalODaoRpl.Add(&totalODaoRpl.Int, &rewardsForNetwork.OracleDaoRpl.Int)
		totalSmoothingPoolEth.Add(&totalSmoothingPoolEth.Int, &rewardsForNetwork.SmoothingPoolEth.Int)
	}

	wrapper := &ProofWrapper{
		RewardsFileVersion: RewardsFileVersion,
		Index:              index,
		ConsensusBlock:     consensusBlock,
		ExecutionBlock:     executionBlock,
		IntervalsPassed:    intervalsPassed,
		MerkleRoot:         treeRoot.Hex(),
		NodeRewards:        nodeRewardsMap,
	}
	wrapper.NetworkRewards.CollateralRplPerNetwork = networkCollateralRplMap
	wrapper.NetworkRewards.OracleDaoRplPerNetwork = networkODaoRplMap
	wrapper.NetworkRewards.SmoothingPoolEthPerNetwork = networkSmoothingPoolEthMap
	wrapper.TotalRewards.ProtocolDaoRpl = protocolDaoRpl
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
		rplRewards.Add(&rewardsForNode.CollateralRpl.Int, &rewardsForNode.OracleDaoRpl.Int)
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
func CalculateRplRewards(rp *rocketpool.RocketPool, snapshotBlockHeader *types.Header, rewardsInterval time.Duration) (map[common.Address]NodeRewards, map[uint64]NodeRewards, *QuotedBigInt, map[common.Address]uint64, error) {

	validNetworkCache := map[uint64]bool{
		0: true,
	}

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
		return nil, nil, nil, nil, err
	}
	pendingRewards, err := rewards.GetPendingRPLRewards(rp, opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, eth.EthToWei(1))

	nodeAddresses, err := node.GetNodeAddresses(rp, opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	totalRplStake, err := node.GetTotalEffectiveRPLStake(rp, opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for _, address := range nodeAddresses {
		// Make sure this node is eligible for rewards
		regTime, err := node.GetNodeRegistrationTime(rp, address, opts)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error getting registration time for node %s: %w", err)
		}
		if snapshotBlockTime.Sub(regTime) < rewardsInterval {
			continue
		}

		// Get how much RPL goes to this node: effective stake / total stake * total RPL rewards for nodes
		nodeStake, err := node.GetNodeEffectiveRPLStake(rp, address, opts)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error getting effective stake for node %s: %w", address.Hex(), err)
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
					return nil, nil, nil, nil, err
				}
				validNetwork, err := ValidateNetwork(rp, network, validNetworkCache)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				if !validNetwork {
					invalidNetworkNodes[address] = network
					network = 0
				}

				rewardsForNode = NodeRewards{
					RewardNetwork:    network,
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
			}
			rewardsForNode.CollateralRpl.Add(&rewardsForNode.CollateralRpl.Int, nodeRplRewards)
			nodeRewardsMap[address] = rewardsForNode

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := networkRewardsMap[rewardsForNode.RewardNetwork]
			if !exists {
				rewardsForNetwork = NodeRewards{
					RewardNetwork:    rewardsForNode.RewardNetwork,
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
			}
			rewardsForNetwork.CollateralRpl.Add(&rewardsForNetwork.CollateralRpl.Int, nodeRplRewards)
			networkRewardsMap[rewardsForNode.RewardNetwork] = rewardsForNetwork
		}
	}

	// Handle Oracle DAO rewards
	oDaoPercent, err := rewards.GetTrustedNodeOperatorRewardsPercent(rp, opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, eth.EthToWei(1))

	oDaoAddresses, err := trustednode.GetMemberAddresses(rp, opts)
	if err != nil {
		return nil, nil, nil, nil, err
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
				return nil, nil, nil, nil, err
			}
			validNetwork, err := ValidateNetwork(rp, network, validNetworkCache)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if !validNetwork {
				invalidNetworkNodes[address] = network
				network = 0
			}

			rewardsForNode = NodeRewards{
				RewardNetwork:    network,
				CollateralRpl:    NewQuotedBigInt(0),
				OracleDaoRpl:     NewQuotedBigInt(0),
				SmoothingPoolEth: NewQuotedBigInt(0),
			}
		}
		rewardsForNode.OracleDaoRpl.Add(&rewardsForNode.OracleDaoRpl.Int, individualOdaoRewards)
		nodeRewardsMap[address] = rewardsForNode

		// Add the rewards to the running total for the specified network
		rewardsForNetwork, exists := networkRewardsMap[rewardsForNode.RewardNetwork]
		if !exists {
			rewardsForNetwork = NodeRewards{
				RewardNetwork:    rewardsForNode.RewardNetwork,
				CollateralRpl:    NewQuotedBigInt(0),
				OracleDaoRpl:     NewQuotedBigInt(0),
				SmoothingPoolEth: NewQuotedBigInt(0),
			}
		}
		rewardsForNetwork.OracleDaoRpl.Add(&rewardsForNetwork.OracleDaoRpl.Int, individualOdaoRewards)
		networkRewardsMap[rewardsForNode.RewardNetwork] = rewardsForNetwork
	}

	// Handle Protocol DAO rewards
	pDaoPercent, err := rewards.GetProtocolDaoRewardsPercent(rp, opts)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	pDaoRewards := NewQuotedBigInt(0)
	pDaoRewards.Mul(pendingRewards, pDaoPercent)
	pDaoRewards.Div(&pDaoRewards.Int, eth.EthToWei(1))

	// Return the rewards maps
	return nodeRewardsMap, networkRewardsMap, pDaoRewards, invalidNetworkNodes, nil
}

// Validates that the provided network is legal
func ValidateNetwork(rp *rocketpool.RocketPool, network uint64, validNetworkCache map[uint64]bool) (bool, error) {
	valid, exists := validNetworkCache[network]
	if !exists {
		var err error
		valid, err = tnsettings.GetNetworkEnabled(rp, big.NewInt(int64(network)), nil)
		if err != nil {
			return false, err
		}
		validNetworkCache[network] = valid
	}

	return valid, nil
}

// Calculates the ETH rewards for the given interval
func CalculateEthRewards(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc *beacon.Client, index uint64, snapshotBlockHeader *types.Header, rewardsInterval time.Duration, nodeRewardsMap map[common.Address]NodeRewards, networkRewardsMap map[uint64]NodeRewards, invalidNetworkNodes map[common.Address]uint64) error {

	// Get services
	opts := &bind.CallOpts{
		BlockNumber: snapshotBlockHeader.Number,
	}
	ec := rp.Client

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool")
	if err != nil {
		return fmt.Errorf("error getting smoothing pool contract: %w", err)
	}

	smoothingPoolBalance, err := rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, snapshotBlockHeader.Number)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}

	// Ignore the ETH calculation if there are no rewards
	if smoothingPoolBalance.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if index == 0 {
		// This is the first interval, Smoothing Pool rewards are ignored on the first interval since it doesn't have a discrete start time
		return nil
	}

	// Get the event log interval
	var eventLogInterval int
	eventLogInterval, err = cfg.GetEventLogInterval()
	if err != nil {
		return err
	}

	// Get the start time of this interval based on the event from the previous one
	previousIntervalEvent, err := rewards.GetRewardSnapshotEvent(rp, index-1, big.NewInt(int64(eventLogInterval)), nil)
	if err != nil {
		return err
	}
	startElBlockNumber := big.NewInt(0).Add(previousIntervalEvent.ExecutionBlock, big.NewInt(1))
	startElBlockHeader, err := ec.HeaderByNumber(context.Background(), startElBlockNumber)
	if err != nil {
		return err
	}
	intervalStartTime := time.Unix(int64(startElBlockHeader.Time), 0)

	// Get all of the registered nodes
	nodeAddresses, err := node.GetNodeAddresses(rp, opts)
	if err != nil {
		return err
	}

	// For each NO, get their opt-in status and time of last change in batches
	nodeCount := len(nodeAddresses)
	details := make([]NodeSmoothingDetails, nodeCount)
	for batchStartIndex := 0; batchStartIndex < nodeCount; batchStartIndex += SmoothingPoolDetailsBatchSize {

		// Get batch start & end index
		iterationStartIndex := batchStartIndex
		iterationEndIndex := batchStartIndex + SmoothingPoolDetailsBatchSize
		if iterationEndIndex > nodeCount {
			iterationEndIndex = nodeCount
		}

		// Load details
		var wg errgroup.Group
		for iterationIndex := iterationStartIndex; iterationIndex < iterationEndIndex; iterationIndex++ {
			iterationIndex := iterationIndex
			wg.Go(func() error {
				var err error
				nodeDetails := NodeSmoothingDetails{
					Address:   nodeAddresses[iterationIndex],
					Minipools: map[common.Address]rptypes.ValidatorPubkey{},
				}

				// Check if the node is opted into the smoothing pool
				nodeDetails.IsOptedIn, err = node.GetSmoothingPoolRegistrationState(rp, nodeDetails.Address, opts)
				if err != nil {
					return fmt.Errorf("Error getting smoothing pool registration state for node %s: %w", nodeDetails.Address.Hex(), err)
				}

				// Get the time of the last registration change
				nodeDetails.StatusChangeTime, err = node.GetSmoothingPoolRegistrationChanged(rp, nodeDetails.Address, opts)
				if err != nil {
					return fmt.Errorf("Error getting smoothing pool registration change time for node %s: %w", nodeDetails.Address.Hex(), err)
				}

				// If the node isn't opted into the Smoothing Pool and they didn't opt out during this interval, ignore them
				if intervalStartTime.Sub(nodeDetails.StatusChangeTime) > 0 && !nodeDetails.IsOptedIn {
					return nil
				}

				// Get the details for each minipool in the node
				minipoolDetails, err := minipool.GetNodeMinipools(rp, nodeDetails.Address, opts)
				if err != nil {
					return fmt.Errorf("Error getting minipool details for node %s: %w", nodeDetails.Address, err)
				}
				for _, mpd := range minipoolDetails {
					if mpd.Exists {
						mp, err := minipool.NewMinipool(rp, mpd.Address)
						if err != nil {
							return fmt.Errorf("Error creating minipool wrapper for minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
						}
						status, err := mp.GetStatus(opts)
						if err != nil {
							return fmt.Errorf("Error getting status of minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
						}
						if status == rptypes.Staking {
							nodeDetails.Minipools[mpd.Address] = mpd.Pubkey
						}
					}
				}

				details[iterationIndex] = nodeDetails
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return err
		}

	}

	return nil
}
