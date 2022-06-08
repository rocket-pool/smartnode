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
	SmoothingPoolDetailsBatchSize uint64 = 20
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
func CalculateRplRewards(rp *rocketpool.RocketPool, snapshotBlockHeader *types.Header, rewardsInterval time.Duration, nodeAddresses []common.Address) (map[common.Address]NodeRewards, map[uint64]NodeRewards, *QuotedBigInt, map[common.Address]uint64, error) {

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
func CalculateEthRewards(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client, index uint64, snapshotBlockHeader *types.Header, snapshotBeaconBlock uint64, nodeRewardsMap map[common.Address]NodeRewards, networkRewardsMap map[uint64]NodeRewards, invalidNetworkNodes map[common.Address]uint64, nodeAddresses []common.Address) (*big.Int, error) {

	// Get services
	opts := &bind.CallOpts{
		BlockNumber: snapshotBlockHeader.Number,
	}
	ec := rp.Client

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool")
	if err != nil {
		return nil, fmt.Errorf("error getting smoothing pool contract: %w", err)
	}

	smoothingPoolBalance, err := rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, snapshotBlockHeader.Number)
	if err != nil {
		return nil, fmt.Errorf("error getting smoothing pool balance: %w", err)
	}

	// Ignore the ETH calculation if there are no rewards
	if smoothingPoolBalance.Cmp(big.NewInt(0)) == 0 {
		return nil, nil
	}

	if index == 0 {
		// This is the first interval, Smoothing Pool rewards are ignored on the first interval since it doesn't have a discrete start time
		return nil, nil
	}

	// Get the event log interval
	var eventLogInterval int
	eventLogInterval, err = cfg.GetEventLogInterval()
	if err != nil {
		return nil, err
	}

	// Get the start time of this interval based on the event from the previous one
	previousIntervalEvent, err := rewards.GetRewardSnapshotEvent(rp, index-1, big.NewInt(int64(eventLogInterval)), nil)
	if err != nil {
		return nil, err
	}
	startElBlockNumber := big.NewInt(0).Add(previousIntervalEvent.ExecutionBlock, big.NewInt(1))
	startElBlockHeader, err := ec.HeaderByNumber(context.Background(), startElBlockNumber)
	if err != nil {
		return nil, err
	}
	intervalStartTime := time.Unix(int64(startElBlockHeader.Time), 0)
	intervalEndTime := time.Unix(int64(snapshotBlockHeader.Time), 0)

	// Get the details for nodes eligible for Smoothing Pool rewards
	// This should be all of the eth1 calls, so do them all at the start of Smoothing Pool calculation to prevent the need for an archive node during normal operations
	nodeDetails, err := getSmoothingPoolNodeDetails(rp, opts, intervalStartTime, intervalEndTime, nodeAddresses)
	if err != nil {
		return nil, err
	}

	// Determine the validator indices of each minipool
	validatorIndexMap, err := createMinipoolIndexMap(bc, nodeDetails)
	if err != nil {
		return nil, err
	}

	// Process the attestation performance for each minipool during this interval
	intervalDutiesInfo := &IntervalDutiesInfo{
		Index: index,
		Slots: map[uint64]*SlotInfo{},
	}
	err = processAttestationsForInterval(bc, validatorIndexMap, intervalDutiesInfo, previousIntervalEvent.ConsensusBlock.Uint64()+1, snapshotBeaconBlock, nodeDetails, *smoothingPoolContract.Address)
	if err != nil {
		return nil, err
	}

	// Determine how much ETH each node gets and how much the pool stakers get
	poolStakerETH := calculateNodeRewards(nodeDetails, smoothingPoolBalance)

	// Update the rewards maps
	validNetworkCache := map[uint64]bool{
		0: true,
	}
	for _, nodeInfo := range nodeDetails {
		if nodeInfo.IsEligible && nodeInfo.SmoothingPoolEth.Cmp(big.NewInt(0)) > 0 {
			rewardsForNode, exists := nodeRewardsMap[nodeInfo.Address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(rp, nodeInfo.Address, opts)
				if err != nil {
					return nil, err
				}
				validNetwork, err := ValidateNetwork(rp, network, validNetworkCache)
				if err != nil {
					return nil, err
				}
				if !validNetwork {
					invalidNetworkNodes[nodeInfo.Address] = network
					network = 0
				}

				rewardsForNode = NodeRewards{
					RewardNetwork:    network,
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
			}
			rewardsForNode.SmoothingPoolEth.Add(&rewardsForNode.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
			nodeRewardsMap[nodeInfo.Address] = rewardsForNode

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
			rewardsForNetwork.SmoothingPoolEth.Add(&rewardsForNetwork.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
			networkRewardsMap[rewardsForNode.RewardNetwork] = rewardsForNetwork
		}
	}

	return poolStakerETH, nil

}

// Calculate the distribution of Smoothing Pool ETH to each node
func calculateNodeRewards(nodeDetails []NodeSmoothingDetails, smoothingPoolBalance *big.Int) *big.Int {

	// Get the average fee for all eligible minipools and calculate their weighted share
	one := big.NewInt(1e18) // 100%, used for dividing percentages properly
	feeTotal := big.NewInt(0)
	minipoolCount := int64(0)
	minipoolShareTotal := big.NewInt(0)
	for _, nodeInfo := range nodeDetails {
		if nodeInfo.IsEligible {
			for _, minipool := range nodeInfo.Minipools {
				// Used for average fee calculation
				feeTotal.Add(feeTotal, minipool.Fee)
				minipoolCount++

				// Minipool share calculation
				minipoolShare := big.NewInt(0).Add(one, minipool.Fee) // Start with 1 + fee
				if nodeInfo.EligibilityFactor != 1.0 {
					// Scale the total shares by the eligibility factor based on how long the node has been opted in
					minipoolShare.Mul(minipoolShare, eth.EthToWei(nodeInfo.EligibilityFactor))
					minipoolShare.Div(minipoolShare, one)
				}
				if minipool.MissedAttestations > 0 && minipool.GoodAttestations > 0 {
					// Calculate the participation rate if there are any missed attestations
					participationRate := float64(minipool.GoodAttestations) / float64(minipool.GoodAttestations+minipool.MissedAttestations)
					minipoolShare.Mul(minipoolShare, eth.EthToWei(participationRate))
					minipoolShare.Div(minipoolShare, one)
				}
				minipoolShareTotal.Add(minipoolShareTotal, minipoolShare)
				minipool.MinipoolShare = minipoolShare
			}
		}
	}
	averageFee := big.NewInt(0).Div(feeTotal, big.NewInt(minipoolCount))

	// Calculate the staking pool share and the node op share
	halfSmoothingPool := big.NewInt(0).Div(smoothingPoolBalance, big.NewInt(2))
	commission := big.NewInt(0)
	commission.Mul(halfSmoothingPool, averageFee)
	commission.Div(commission, one)
	poolStakerShare := big.NewInt(0).Sub(halfSmoothingPool, commission)
	nodeOpShare := big.NewInt(0).Sub(smoothingPoolBalance, poolStakerShare)

	// Calculate the amount of ETH to give each minipool based on their share
	totalEthForMinipools := big.NewInt(0)
	for _, nodeInfo := range nodeDetails {
		nodeInfo.SmoothingPoolEth = big.NewInt(0)
		if nodeInfo.IsEligible {
			for _, minipool := range nodeInfo.Minipools {
				// Minipool ETH = NO amount * minipool share / total minipool share
				minipoolEth := big.NewInt(0).Set(nodeOpShare)
				minipoolEth.Mul(minipoolEth, minipool.MinipoolShare)
				minipoolEth.Div(minipoolEth, minipoolShareTotal)
				nodeInfo.SmoothingPoolEth.Add(nodeInfo.SmoothingPoolEth, minipoolEth)
			}
			totalEthForMinipools.Add(totalEthForMinipools, nodeInfo.SmoothingPoolEth)
		}
	}

	// This is how much actually goes to the pool stakers - it should ideally be equal to poolStakerShare but this accounts for any cumulative floating point errors
	truePoolStakerAmount := big.NewInt(0).Sub(smoothingPoolBalance, totalEthForMinipools)
	return truePoolStakerAmount

}

// Get all of the duties for a range of epochs
func processAttestationsForInterval(bc beacon.Client, validatorIndexMap map[uint64]*MinipoolInfo, intervalDutiesInfo *IntervalDutiesInfo, startSlot uint64, endSlot uint64, nodeDetails []NodeSmoothingDetails, smoothingPoolAddress common.Address) error {

	// Determine the start and end epochs to check
	beaconConfig, err := bc.GetEth2Config()
	if err != nil {
		return err
	}
	genesisTime := time.Unix(int64(beaconConfig.GenesisTime), 0)
	slotLength := time.Duration(beaconConfig.SecondsPerSlot) * time.Second

	startEpoch := startSlot / beaconConfig.SlotsPerEpoch
	endEpoch := endSlot / beaconConfig.SlotsPerEpoch

	// Check all of the attestations for each epoch
	for epoch := startEpoch; epoch < endEpoch+1; epoch++ {
		startTime := time.Now()
		// Get all of the expected duties for the epoch
		err := getDutiesForEpoch(bc, epoch, startSlot, endSlot, validatorIndexMap, intervalDutiesInfo)
		if err != nil {
			return fmt.Errorf("Error getting duties for epoch %d: %w", epoch, err)
		}

		// Process all of the slots in the epoch
		for i := uint64(0); i < 32; i++ {
			checkDutiesForSlot(bc, epoch*32+i, validatorIndexMap, intervalDutiesInfo, nodeDetails, smoothingPoolAddress, genesisTime, slotLength)
		}

		// Report on missed duties
		fmt.Printf("Epoch %d complete in %s\n", epoch, time.Since(startTime))
	}

	// Check all of the slots in the epoch after the end too
	for i := uint64(0); i < 32; i++ {
		checkDutiesForSlot(bc, (endSlot+1)*32+i, validatorIndexMap, intervalDutiesInfo, nodeDetails, smoothingPoolAddress, genesisTime, slotLength)
	}

	return nil

}

// Handle all of the attestations in the given slot
func checkDutiesForSlot(bc beacon.Client, slot uint64, validatorIndexMap map[uint64]*MinipoolInfo, intervalDutiesInfo *IntervalDutiesInfo, nodeDetails []NodeSmoothingDetails, smoothingPoolAddress common.Address, genesisTime time.Time, slotLength time.Duration) error {

	block, found, err := bc.GetBeaconBlock(fmt.Sprint(slot))
	if !found {
		// Ignore missing blocks
		return nil
	} else if err != nil {
		return err
	}

	// Check if the minipool cheated
	checkIfMinipoolCheated(block, validatorIndexMap, nodeDetails, smoothingPoolAddress, genesisTime, slotLength)

	// Go through the attestations for the block
	for _, attestation := range block.Attestations {

		// Get the RP committees for this attestation's slot and index
		slotInfo, exists := intervalDutiesInfo.Slots[attestation.SlotIndex]
		if exists {
			rpCommittee, exists := slotInfo.Committees[attestation.CommitteeIndex]
			if exists {
				// Check if each RP validator attested successfully
				for position, validator := range rpCommittee.Positions {
					if attestation.AggregationBits.BitAt(uint64(position)) {
						// We have a winner - remove this duty and update the scores
						delete(rpCommittee.Positions, position)
						if len(rpCommittee.Positions) == 0 {
							delete(slotInfo.Committees, attestation.CommitteeIndex)
						}
						if len(slotInfo.Committees) == 0 {
							delete(intervalDutiesInfo.Slots, attestation.SlotIndex)
						}
						validator.MissedAttestations--
						validator.GoodAttestations++
					}
				}
			}
		}
	}

	return nil

}

func checkIfMinipoolCheated(block beacon.BeaconBlock, validatorIndexMap map[uint64]*MinipoolInfo, nodeDetails []NodeSmoothingDetails, smoothingPoolAddress common.Address, genesisTime time.Time, slotLength time.Duration) {

	mpInfo, exists := validatorIndexMap[block.ProposerIndex]
	if !exists {
		// This proposer isn't a minipool
		return
	}

	// Check if the fee recipient was the Smoothing Pool
	if block.FeeRecipient == smoothingPoolAddress {
		return
	}

	// If this node wasn't part of the Smoothing Pool this interval, ignore it
	nodeInfo := nodeDetails[mpInfo.NodeIndex]
	if !nodeInfo.IsEligible {
		return
	}

	// Get the time this slot occurred
	slotTime := genesisTime.Add(slotLength * time.Duration(block.Slot))

	// If the node was opted in at the end but this block happened before they opted in, ignore it
	if nodeInfo.IsOptedIn && nodeInfo.StatusChangeTime.Sub(slotTime) >= 0 {
		return
	}

	// If the node was opted out at the end but this block happened after they opted out, ignore it
	if !nodeInfo.IsOptedIn && slotTime.Sub(nodeInfo.StatusChangeTime) >= 0 {
		return
	}

	// If we got here, the block happened while they were opted in so they cheated
	nodeDetails[mpInfo.NodeIndex].IsEligible = false
	nodeDetails[mpInfo.NodeIndex].CheaterInfo.CheatingDetected = true
	nodeDetails[mpInfo.NodeIndex].CheaterInfo.OffendingSlot = block.Slot
	nodeDetails[mpInfo.NodeIndex].CheaterInfo.FeeRecipient = block.FeeRecipient
	nodeDetails[mpInfo.NodeIndex].CheaterInfo.Minipool = mpInfo.Address

}

// Maps out the attestaion duties for the given epoch
func getDutiesForEpoch(bc beacon.Client, epoch uint64, startSlot uint64, endSlot uint64, validatorIndexMap map[uint64]*MinipoolInfo, intervalDutiesInfo *IntervalDutiesInfo) error {

	// Get the committees for the epoch
	committees, err := bc.GetCommitteesForEpoch(&epoch)
	if err != nil {
		return err
	}

	// Crawl the committees
	for _, committee := range committees {
		slotIndex := committee.Slot
		if slotIndex < startSlot || slotIndex > endSlot {
			// Ignore slots that are out of bounds
			continue
		}
		committeeIndex := committee.Index

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*MinipoolInfo{}
		for position, validator := range committee.Validators {
			minipoolInfo, exists := validatorIndexMap[validator]
			if exists {
				rpValidators[position] = minipoolInfo
				minipoolInfo.MissedAttestations += 1 // Consider this attestation missed until it's seen later
			}
		}

		// If there are some RP validators, add this committee to the map
		if len(rpValidators) > 0 {
			slotInfo, exists := intervalDutiesInfo.Slots[slotIndex]
			if !exists {
				slotInfo = &SlotInfo{
					Index:      slotIndex,
					Committees: map[uint64]*CommitteeInfo{},
				}
				intervalDutiesInfo.Slots[slotIndex] = slotInfo
			}
			slotInfo.Committees[committeeIndex] = &CommitteeInfo{
				Index:     committeeIndex,
				Positions: rpValidators,
			}
		}
	}

	return nil

}

// Maps all minipools to their validator indices and creates a map of indices to minipool info
func createMinipoolIndexMap(bc beacon.Client, nodeDetails []NodeSmoothingDetails) (map[uint64]*MinipoolInfo, error) {

	// Make a slice of all minipool pubkeys
	minipoolPubkeys := []rptypes.ValidatorPubkey{}
	for _, details := range nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				minipoolPubkeys = append(minipoolPubkeys, minipoolInfo.ValidatorPubkey)
			}
		}
	}

	// Get indices for all minipool validators
	validatorIndexMap := map[uint64]*MinipoolInfo{}
	statusMap, err := bc.GetValidatorStatuses(minipoolPubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator statuses: %w", err)
	}
	for _, details := range nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				minipoolInfo.ValidatorIndex = statusMap[minipoolInfo.ValidatorPubkey].Index
				validatorIndexMap[minipoolInfo.ValidatorIndex] = minipoolInfo
			}
		}
	}

	return validatorIndexMap, nil

}

// Get the details for every node that was opted into the Smoothing Pool for at least some portion of this interval
func getSmoothingPoolNodeDetails(rp *rocketpool.RocketPool, opts *bind.CallOpts, intervalStartTime time.Time, intervalEndTime time.Time, nodeAddresses []common.Address) ([]NodeSmoothingDetails, error) {

	intervalDuration := float64(intervalEndTime.Sub(intervalStartTime))

	// For each NO, get their opt-in status and time of last change in batches
	nodeCount := uint64(len(nodeAddresses))
	details := make([]NodeSmoothingDetails, nodeCount)
	for batchStartIndex := uint64(0); batchStartIndex < nodeCount; batchStartIndex += SmoothingPoolDetailsBatchSize {

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
					Minipools: []*MinipoolInfo{},
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
					nodeDetails.IsEligible = false
					nodeDetails.EligibilityFactor = 0
					details[iterationIndex] = nodeDetails
					return nil
				}

				// Get the node's total active factor
				if nodeDetails.IsOptedIn {
					nodeDetails.EligibilityFactor = float64(intervalEndTime.Sub(nodeDetails.StatusChangeTime)) / intervalDuration
				} else {
					nodeDetails.EligibilityFactor = float64(nodeDetails.StatusChangeTime.Sub(intervalStartTime)) / intervalDuration
				}
				if nodeDetails.EligibilityFactor > 1 {
					nodeDetails.EligibilityFactor = 1
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
							fee, err := mp.GetNodeFeeRaw(opts)
							if err != nil {
								return fmt.Errorf("Error getting fee for minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
							}
							nodeDetails.Minipools = append(nodeDetails.Minipools, &MinipoolInfo{
								Address:         mpd.Address,
								ValidatorPubkey: mpd.Pubkey,
								NodeAddress:     nodeDetails.Address,
								NodeIndex:       iterationIndex,
								Fee:             fee,
							})
						}
					}
				}

				nodeDetails.IsEligible = len(nodeDetails.Minipools) > 0
				details[iterationIndex] = nodeDetails
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return nil, err
		}
	}

	return details, nil

}
