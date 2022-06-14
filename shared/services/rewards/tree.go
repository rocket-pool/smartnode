package rewards

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
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
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/wealdtech/go-merkletree"
	"github.com/wealdtech/go-merkletree/keccak256"
	"golang.org/x/sync/errgroup"
)

// Settings
const (
	SmoothingPoolDetailsBatchSize uint64 = 20
	RewardsFileVersion            uint64 = 1
)

// Minipool stats
type SmoothingPoolMinipoolPerformance struct {
	SuccessfulAttestations  uint64   `json:"successfulAttestations"`
	MissedAttestations      uint64   `json:"missedAttestations"`
	ParticipationRate       float64  `json:"participationRate"`
	MissingAttestationSlots []uint64 `json:"missingAttestationSlots"`
	ETHEarned               float64  `json:"ethEarned"`
}

// Node operator rewards
type NodeRewardsInfo struct {
	RewardNetwork       uint64                                               `json:"rewardNetwork"`
	CollateralRpl       *QuotedBigInt                                        `json:"collateralRpl"`
	OracleDaoRpl        *QuotedBigInt                                        `json:"oracleDaoRpl"`
	SmoothingPoolEth    *QuotedBigInt                                        `json:"smoothingPoolEth"`
	MerkleData          []byte                                               `json:"-"`
	MerkleProof         []string                                             `json:"merkleProof"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance `json:"minipoolPerformance,omitempty"`
}

// Rewards per network
type NetworkRewardsInfo struct {
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
}

// Total cumulative rewards for an interval
type TotalRewards struct {
	ProtocolDaoRpl               *QuotedBigInt `json:"protocolDaoRpl"`
	TotalCollateralRpl           *QuotedBigInt `json:"totalCollateralRpl"`
	TotalOracleDaoRpl            *QuotedBigInt `json:"totalOracleDaoRpl"`
	TotalSmoothingPoolEth        *QuotedBigInt `json:"totalSmoothingPoolEth"`
	PoolStakerSmoothingPoolEth   *QuotedBigInt `json:"poolStakerSmoothingPoolEth"`
	NodeOperatorSmoothingPoolEth *QuotedBigInt `json:"nodeOperatorSmoothingPoolEth"`
}

// JSON struct for a complete rewards file
type RewardsFile struct {
	// Serialized fields
	RewardsFileVersion  uint64                              `json:"rewardsFileVersion"`
	Index               uint64                              `json:"index"`
	Network             string                              `json:"network"`
	StartTime           time.Time                           `json:"startTime,omitempty"`
	EndTime             time.Time                           `json:"endTime"`
	ConsensusStartBlock uint64                              `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock   uint64                              `json:"consensusEndBlock"`
	ExecutionStartBlock uint64                              `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock   uint64                              `json:"executionEndBlock"`
	IntervalsPassed     uint64                              `json:"intervalsPassed"`
	MerkleRoot          string                              `json:"merkleRoot,omitempty"`
	TotalRewards        *TotalRewards                       `json:"totalRewards"`
	NetworkRewards      map[uint64]*NetworkRewardsInfo      `json:"networkRewards"`
	NodeRewards         map[common.Address]*NodeRewardsInfo `json:"nodeRewards"`

	// Non-serialized fields
	MerkleTree          *merkletree.MerkleTree    `json:"-"`
	InvalidNetworkNodes map[common.Address]uint64 `json:"-"`
	elSnapshotHeader    *types.Header             `json:"-"`
	log                 log.ColorLogger           `json:"-"`
	logPrefix           string                    `json:"-"`
}

// Create a new rewards file
func NewRewardsFile(log log.ColorLogger, logPrefix string, index uint64, startTime time.Time, endTime time.Time, consensusBlock uint64, elSnapshotHeader *types.Header, intervalsPassed uint64) *RewardsFile {
	return &RewardsFile{
		RewardsFileVersion: RewardsFileVersion,
		Index:              index,
		StartTime:          startTime,
		EndTime:            endTime,
		ConsensusEndBlock:  consensusBlock,
		ExecutionEndBlock:  elSnapshotHeader.Number.Uint64(),
		IntervalsPassed:    intervalsPassed,
		TotalRewards: &TotalRewards{
			ProtocolDaoRpl:               NewQuotedBigInt(0),
			TotalCollateralRpl:           NewQuotedBigInt(0),
			TotalOracleDaoRpl:            NewQuotedBigInt(0),
			TotalSmoothingPoolEth:        NewQuotedBigInt(0),
			PoolStakerSmoothingPoolEth:   NewQuotedBigInt(0),
			NodeOperatorSmoothingPoolEth: NewQuotedBigInt(0),
		},
		NetworkRewards:      map[uint64]*NetworkRewardsInfo{},
		NodeRewards:         map[common.Address]*NodeRewardsInfo{},
		InvalidNetworkNodes: map[common.Address]uint64{},
		elSnapshotHeader:    elSnapshotHeader,
		log:                 log,
		logPrefix:           logPrefix,
	}
}

func (r *RewardsFile) GenerateTree(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) error {

	// Set the network name
	r.Network = fmt.Sprint(cfg.Smartnode.Network.Value)

	// Get the addresses for all nodes
	opts := &bind.CallOpts{
		BlockNumber: r.elSnapshotHeader.Number,
	}
	nodeAddresses, err := node.GetNodeAddresses(rp, opts)
	if err != nil {
		return fmt.Errorf("Error getting node addresses: %w", err)
	}
	r.log.Printlnf("%s Creating tree for %d nodes", r.logPrefix, len(nodeAddresses))

	// Calculate the RPL rewards
	err = r.calculateRplRewards(rp, nodeAddresses, opts)
	if err != nil {
		return fmt.Errorf("Error calculating RPL rewards: %w", err)
	}

	// Calculate the ETH rewards
	err = r.calculateEthRewards(rp, cfg, bc, nodeAddresses, opts)
	if err != nil {
		return fmt.Errorf("Error calculating ETH rewards: %w", err)
	}

	// Calculate the network reward map and the totals
	r.updateNetworksAndTotals()

	// Generate the Merkle Tree
	err = r.generateMerkleTree()
	if err != nil {
		return fmt.Errorf("Error generating Merkle tree: %w", err)
	}

	// Sort all of the missed attestations so the files are always generated in the same state
	for _, nodeInfo := range r.NodeRewards {
		for _, minipoolInfo := range nodeInfo.MinipoolPerformance {
			sort.Slice(minipoolInfo.MissingAttestationSlots, func(i, j int) bool {
				return minipoolInfo.MissingAttestationSlots[i] < minipoolInfo.MissingAttestationSlots[j]
			})
		}
	}

	return nil

}

// Generates a merkle tree from the provided rewards map
func (r *RewardsFile) generateMerkleTree() error {

	// Generate the leaf data for each node
	totalData := make([][]byte, 0, len(r.NodeRewards))
	for address, rewardsForNode := range r.NodeRewards {
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
		totalData = append(totalData, nodeData)
	}

	// Generate the tree
	tree, err := merkletree.NewUsing(totalData, keccak256.New(), false, true)
	if err != nil {
		return fmt.Errorf("error generating Merkle Tree: %w", err)
	}

	// Generate the proofs for each node
	for address, rewardsForNode := range r.NodeRewards {
		// Get the proof
		proof, err := tree.GenerateProof(rewardsForNode.MerkleData, 0)
		if err != nil {
			return fmt.Errorf("error generating proof for node %s: %w", address.Hex(), err)
		}

		// Convert the proof into hex strings
		proofStrings := make([]string, len(proof.Hashes))
		for i, hash := range proof.Hashes {
			proofStrings[i] = fmt.Sprintf("0x%s", hex.EncodeToString(hash))
		}

		// Assign the hex strings to the node rewards struct
		rewardsForNode.MerkleProof = proofStrings
	}

	r.MerkleTree = tree
	r.MerkleRoot = common.BytesToHash(tree.Root()).Hex()
	return nil

}

func (r *RewardsFile) updateNetworksAndTotals() {

	// Get the highest network index with valid rewards
	highestNetworkIndex := uint64(0)
	for network := range r.NetworkRewards {
		if network > highestNetworkIndex {
			highestNetworkIndex = network
		}
	}

	// Create the map for each network, including unused ones
	for network := uint64(0); network <= highestNetworkIndex; network++ {
		rewardsForNetwork, exists := r.NetworkRewards[network]
		if !exists {
			rewardsForNetwork = &NetworkRewardsInfo{
				CollateralRpl:    NewQuotedBigInt(0),
				OracleDaoRpl:     NewQuotedBigInt(0),
				SmoothingPoolEth: NewQuotedBigInt(0),
			}
			r.NetworkRewards[network] = rewardsForNetwork
		}

		// Calculate the total RPL
		r.TotalRewards.TotalCollateralRpl.Add(&r.TotalRewards.TotalCollateralRpl.Int, &rewardsForNetwork.CollateralRpl.Int)
		r.TotalRewards.TotalOracleDaoRpl.Add(&r.TotalRewards.TotalOracleDaoRpl.Int, &rewardsForNetwork.OracleDaoRpl.Int)
	}

}

// Calculates the RPL rewards for the given interval
func (r *RewardsFile) calculateRplRewards(rp *rocketpool.RocketPool, nodeAddresses []common.Address, opts *bind.CallOpts) error {

	validNetworkCache := map[uint64]bool{
		0: true,
	}

	snapshotBlockTime := time.Unix(int64(r.elSnapshotHeader.Time), 0)
	requiredRegistrationLength, err := rewards.GetClaimIntervalTime(rp, opts)
	if err != nil {
		return fmt.Errorf("error getting required registration time: %w", err)
	}

	// Handle node operator rewards
	nodeOpPercent, err := rewards.GetNodeOperatorRewardsPercent(rp, opts)
	if err != nil {
		return err
	}
	pendingRewards, err := rewards.GetPendingRPLRewards(rp, opts)
	if err != nil {
		return err
	}
	r.log.Printlnf("%s Pending RPL rewards: %s (%.3f)", r.logPrefix, pendingRewards.String(), eth.WeiToEth(pendingRewards))
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, eth.EthToWei(1))
	r.log.Printlnf("%s Total collateral RPL rewards: %s (%.3f)", r.logPrefix, totalNodeRewards.String(), eth.WeiToEth(totalNodeRewards))

	totalRplStake, err := node.GetTotalEffectiveRPLStake(rp, opts)
	if err != nil {
		return err
	}

	for _, address := range nodeAddresses {
		// Make sure this node is eligible for rewards
		regTime, err := node.GetNodeRegistrationTime(rp, address, opts)
		if err != nil {
			return fmt.Errorf("error getting registration time for node %s: %w", address, err)
		}
		if snapshotBlockTime.Sub(regTime) < requiredRegistrationLength {
			continue
		}

		// Get how much RPL goes to this node: effective stake / total stake * total RPL rewards for nodes
		nodeStake, err := node.GetNodeEffectiveRPLStake(rp, address, opts)
		if err != nil {
			return fmt.Errorf("error getting effective stake for node %s: %w", address.Hex(), err)
		}
		nodeRplRewards := big.NewInt(0)
		nodeRplRewards.Mul(nodeStake, totalNodeRewards)
		nodeRplRewards.Div(nodeRplRewards, totalRplStake)

		// If there are pending rewards, add it to the map
		if nodeRplRewards.Cmp(big.NewInt(0)) == 1 {
			rewardsForNode, exists := r.NodeRewards[address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(rp, address, opts)
				if err != nil {
					return err
				}
				validNetwork, err := validateNetwork(rp, network, validNetworkCache)
				if err != nil {
					return err
				}
				if !validNetwork {
					r.InvalidNetworkNodes[address] = network
					network = 0
				}

				rewardsForNode = &NodeRewardsInfo{
					RewardNetwork:    network,
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
				r.NodeRewards[address] = rewardsForNode
			}
			rewardsForNode.CollateralRpl.Add(&rewardsForNode.CollateralRpl.Int, nodeRplRewards)

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.NetworkRewards[rewardsForNode.RewardNetwork]
			if !exists {
				rewardsForNetwork = &NetworkRewardsInfo{
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
				r.NetworkRewards[rewardsForNode.RewardNetwork] = rewardsForNetwork
			}
			rewardsForNetwork.CollateralRpl.Add(&rewardsForNetwork.CollateralRpl.Int, nodeRplRewards)
		}
	}

	// Handle Oracle DAO rewards
	oDaoPercent, err := rewards.GetTrustedNodeOperatorRewardsPercent(rp, opts)
	if err != nil {
		return err
	}
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, eth.EthToWei(1))
	r.log.Printlnf("%s Total Oracle DAO RPL rewards: %s (%.3f)", r.logPrefix, totalODaoRewards.String(), eth.WeiToEth(totalODaoRewards))

	oDaoAddresses, err := trustednode.GetMemberAddresses(rp, opts)
	if err != nil {
		return err
	}
	memberCount := big.NewInt(int64(len(oDaoAddresses)))
	individualOdaoRewards := big.NewInt(0)
	individualOdaoRewards.Div(totalODaoRewards, memberCount)

	for _, address := range oDaoAddresses {
		// Make sure this node is eligible for rewards
		regTime, err := node.GetNodeRegistrationTime(rp, address, opts)
		if err != nil {
			return fmt.Errorf("error getting registration time for node %s: %w", address, err)
		}
		if snapshotBlockTime.Sub(regTime) < requiredRegistrationLength {
			continue
		}

		rewardsForNode, exists := r.NodeRewards[address]
		if !exists {
			// Get the network the rewards should go to
			network, err := node.GetRewardNetwork(rp, address, opts)
			if err != nil {
				return err
			}
			validNetwork, err := validateNetwork(rp, network, validNetworkCache)
			if err != nil {
				return err
			}
			if !validNetwork {
				r.InvalidNetworkNodes[address] = network
				network = 0
			}

			rewardsForNode = &NodeRewardsInfo{
				RewardNetwork:    network,
				CollateralRpl:    NewQuotedBigInt(0),
				OracleDaoRpl:     NewQuotedBigInt(0),
				SmoothingPoolEth: NewQuotedBigInt(0),
			}
			r.NodeRewards[address] = rewardsForNode

		}
		rewardsForNode.OracleDaoRpl.Add(&rewardsForNode.OracleDaoRpl.Int, individualOdaoRewards)

		// Add the rewards to the running total for the specified network
		rewardsForNetwork, exists := r.NetworkRewards[rewardsForNode.RewardNetwork]
		if !exists {
			rewardsForNetwork = &NetworkRewardsInfo{
				CollateralRpl:    NewQuotedBigInt(0),
				OracleDaoRpl:     NewQuotedBigInt(0),
				SmoothingPoolEth: NewQuotedBigInt(0),
			}
			r.NetworkRewards[rewardsForNode.RewardNetwork] = rewardsForNetwork
		}
		rewardsForNetwork.OracleDaoRpl.Add(&rewardsForNetwork.OracleDaoRpl.Int, individualOdaoRewards)
	}

	// Handle Protocol DAO rewards
	pDaoPercent, err := rewards.GetProtocolDaoRewardsPercent(rp, opts)
	if err != nil {
		return err
	}
	pDaoRewards := NewQuotedBigInt(0)
	pDaoRewards.Mul(pendingRewards, pDaoPercent)
	pDaoRewards.Div(&pDaoRewards.Int, eth.EthToWei(1))
	r.TotalRewards.ProtocolDaoRpl = pDaoRewards
	r.log.Printlnf("%s Total Protocol DAO rewards: %s (%.3f)", r.logPrefix, pDaoRewards.String(), eth.WeiToEth(&pDaoRewards.Int))

	return nil

}

// Calculates the ETH rewards for the given interval
func (r *RewardsFile) calculateEthRewards(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client, nodeAddresses []common.Address, opts *bind.CallOpts) error {

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool")
	if err != nil {
		return fmt.Errorf("error getting smoothing pool contract: %w", err)
	}

	smoothingPoolBalance, err := rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, r.elSnapshotHeader.Number)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}
	r.log.Printlnf("%s Smoothing Pool Balance: %s (%.3f)", r.logPrefix, smoothingPoolBalance.String(), eth.WeiToEth(smoothingPoolBalance))

	// Ignore the ETH calculation if there are no rewards
	if smoothingPoolBalance.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if r.Index == 0 {
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
	previousIntervalEvent, err := rewards.GetRewardSnapshotEventWithUpgrades(rp, r.Index-1, big.NewInt(int64(eventLogInterval)), nil, cfg.Smartnode.GetPreviousRewardsPoolAddresses())
	if err != nil {
		return err
	}
	startElBlockNumber := big.NewInt(0).Add(previousIntervalEvent.ExecutionBlock, big.NewInt(1))
	startElBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), startElBlockNumber)
	if err != nil {
		return err
	}

	r.ConsensusStartBlock = previousIntervalEvent.ConsensusBlock.Uint64() + 1
	r.ExecutionStartBlock = startElBlockNumber.Uint64()
	elStartTime := time.Unix(int64(startElBlockHeader.Time), 0)
	elEndTime := time.Unix(int64(r.elSnapshotHeader.Time), 0)

	// Get the details for nodes eligible for Smoothing Pool rewards
	// This should be all of the eth1 calls, so do them all at the start of Smoothing Pool calculation to prevent the need for an archive node during normal operations
	nodeDetails, err := getSmoothingPoolNodeDetails(rp, opts, elStartTime, elEndTime, nodeAddresses)
	if err != nil {
		return err
	}
	eligible := 0
	for _, nodeInfo := range nodeDetails {
		if nodeInfo.IsEligible {
			eligible++
		}
	}
	r.log.Printlnf("%s %d / %d nodes were eligible for Smoothing Pool rewards", r.logPrefix, eligible, len(nodeDetails))

	// Determine the validator indices of each minipool
	validatorIndexMap, err := createMinipoolIndexMap(bc, nodeDetails)
	if err != nil {
		return err
	}

	// Process the attestation performance for each minipool during this interval
	intervalDutiesInfo := &IntervalDutiesInfo{
		Index: r.Index,
		Slots: map[uint64]*SlotInfo{},
	}
	err = r.processAttestationsForInterval(bc, validatorIndexMap, intervalDutiesInfo, previousIntervalEvent.ConsensusBlock.Uint64()+1, r.ConsensusEndBlock, nodeDetails, *smoothingPoolContract.Address)
	if err != nil {
		return err
	}

	// Determine how much ETH each node gets and how much the pool stakers get
	poolStakerETH, nodeOpEth := calculateNodeRewards(nodeDetails, smoothingPoolBalance)
	r.log.Printlnf("%s Pool staker ETH: %s (%.3f)", r.logPrefix, poolStakerETH.String(), eth.WeiToEth(poolStakerETH))
	r.log.Printlnf("%s Node Op ETH:     %s (%.3f)", r.logPrefix, nodeOpEth.String(), eth.WeiToEth(nodeOpEth))

	// Update the rewards maps
	validNetworkCache := map[uint64]bool{
		0: true,
	}
	for _, nodeInfo := range nodeDetails {
		if nodeInfo.IsEligible && nodeInfo.SmoothingPoolEth.Cmp(big.NewInt(0)) > 0 {
			rewardsForNode, exists := r.NodeRewards[nodeInfo.Address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(rp, nodeInfo.Address, opts)
				if err != nil {
					return err
				}
				validNetwork, err := validateNetwork(rp, network, validNetworkCache)
				if err != nil {
					return err
				}
				if !validNetwork {
					r.InvalidNetworkNodes[nodeInfo.Address] = network
					network = 0
				}

				rewardsForNode = &NodeRewardsInfo{
					RewardNetwork:    network,
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
				r.NodeRewards[nodeInfo.Address] = rewardsForNode
			}
			rewardsForNode.SmoothingPoolEth.Add(&rewardsForNode.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)

			// Add minipool rewards to the JSON
			rewardsForNode.MinipoolPerformance = map[common.Address]*SmoothingPoolMinipoolPerformance{}
			for _, minipoolInfo := range nodeInfo.Minipools {
				if !minipoolInfo.WasActive {
					continue
				}
				performance := &SmoothingPoolMinipoolPerformance{
					SuccessfulAttestations:  minipoolInfo.GoodAttestations,
					MissedAttestations:      minipoolInfo.MissedAttestations,
					ParticipationRate:       float64(minipoolInfo.GoodAttestations) / float64(minipoolInfo.GoodAttestations+minipoolInfo.MissedAttestations),
					ETHEarned:               eth.WeiToEth(minipoolInfo.MinipoolShare),
					MissingAttestationSlots: []uint64{},
				}
				for slot := range minipoolInfo.MissingAttestationSlots {
					performance.MissingAttestationSlots = append(performance.MissingAttestationSlots, slot)
				}
				rewardsForNode.MinipoolPerformance[minipoolInfo.Address] = performance
			}

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.NetworkRewards[rewardsForNode.RewardNetwork]
			if !exists {
				rewardsForNetwork = &NetworkRewardsInfo{
					CollateralRpl:    NewQuotedBigInt(0),
					OracleDaoRpl:     NewQuotedBigInt(0),
					SmoothingPoolEth: NewQuotedBigInt(0),
				}
				r.NetworkRewards[rewardsForNode.RewardNetwork] = rewardsForNetwork
			}
			rewardsForNetwork.SmoothingPoolEth.Add(&rewardsForNetwork.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
		}
	}

	// Set the totals
	r.TotalRewards.PoolStakerSmoothingPoolEth.Int = *poolStakerETH
	r.TotalRewards.NodeOperatorSmoothingPoolEth.Int = *nodeOpEth
	r.TotalRewards.TotalSmoothingPoolEth.Int = *smoothingPoolBalance
	return nil

}

// Calculate the distribution of Smoothing Pool ETH to each node
func calculateNodeRewards(nodeDetails []*NodeSmoothingDetails, smoothingPoolBalance *big.Int) (*big.Int, *big.Int) {

	// Get the average fee for all eligible minipools and calculate their weighted share
	one := big.NewInt(1e18) // 100%, used for dividing percentages properly
	feeTotal := big.NewInt(0)
	minipoolCount := int64(0)
	minipoolShareTotal := big.NewInt(0)
	for _, nodeInfo := range nodeDetails {
		if nodeInfo.IsEligible {
			for _, minipool := range nodeInfo.Minipools {
				if minipool.GoodAttestations+minipool.MissedAttestations == 0 {
					// Ignore minipools that weren't active for the interval
					minipool.WasActive = false
					continue
				}
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
				if !minipool.WasActive {
					continue
				}
				// Minipool ETH = NO amount * minipool share / total minipool share
				minipoolEth := big.NewInt(0).Set(nodeOpShare)
				minipoolEth.Mul(minipoolEth, minipool.MinipoolShare)
				minipoolEth.Div(minipoolEth, minipoolShareTotal)
				nodeInfo.SmoothingPoolEth.Add(nodeInfo.SmoothingPoolEth, minipoolEth)
				minipool.MinipoolShare = minipoolEth // Set the minipool share to the normalized fraction for the JSON
			}
			totalEthForMinipools.Add(totalEthForMinipools, nodeInfo.SmoothingPoolEth)
		}
	}

	// This is how much actually goes to the pool stakers - it should ideally be equal to poolStakerShare but this accounts for any cumulative floating point errors
	truePoolStakerAmount := big.NewInt(0).Sub(smoothingPoolBalance, totalEthForMinipools)
	return truePoolStakerAmount, totalEthForMinipools

}

// Get all of the duties for a range of epochs
func (r *RewardsFile) processAttestationsForInterval(bc beacon.Client, validatorIndexMap map[uint64]*MinipoolInfo, intervalDutiesInfo *IntervalDutiesInfo, startSlot uint64, endSlot uint64, nodeDetails []*NodeSmoothingDetails, smoothingPoolAddress common.Address) error {

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
	r.log.Printlnf("%s Checking participation of %d minipools for epochs %d to %d", r.logPrefix, len(validatorIndexMap), startEpoch, endEpoch)
	r.log.Printlnf("%s NOTE: this will take a long time, progress is reported every 100 epochs", r.logPrefix)
	epochsDone := 0
	reportStartTime := time.Now()
	for epoch := startEpoch; epoch < endEpoch+1; epoch++ {
		if epochsDone == 100 {
			timeTaken := time.Since(reportStartTime)
			r.log.Printlnf("%s On Epoch %d... (%s so far)", r.logPrefix, epoch, timeTaken)
			epochsDone = 0
		}
		// Get all of the expected duties for the epoch
		err := getDutiesForEpoch(bc, epoch, startSlot, endSlot, validatorIndexMap, intervalDutiesInfo)
		if err != nil {
			return fmt.Errorf("Error getting duties for epoch %d: %w", epoch, err)
		}

		// Process all of the slots in the epoch
		for i := uint64(0); i < 32; i++ {
			checkDutiesForSlot(bc, epoch*32+i, validatorIndexMap, intervalDutiesInfo, nodeDetails, smoothingPoolAddress, genesisTime, slotLength)
		}
		epochsDone++
	}

	// Check all of the slots in the epoch after the end too
	for i := uint64(0); i < 32; i++ {
		checkDutiesForSlot(bc, (endSlot+1)*32+i, validatorIndexMap, intervalDutiesInfo, nodeDetails, smoothingPoolAddress, genesisTime, slotLength)
	}

	r.log.Printlnf("%s Finished participation check (total time = %s)", r.logPrefix, time.Since(reportStartTime))
	return nil

}

// Handle all of the attestations in the given slot
func checkDutiesForSlot(bc beacon.Client, slot uint64, validatorIndexMap map[uint64]*MinipoolInfo, intervalDutiesInfo *IntervalDutiesInfo, nodeDetails []*NodeSmoothingDetails, smoothingPoolAddress common.Address, genesisTime time.Time, slotLength time.Duration) error {

	block, found, err := bc.GetBeaconBlock(fmt.Sprint(slot))
	if !found {
		// Ignore missing blocks
		return nil
	} else if err != nil {
		return err
	}

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
						delete(validator.MissingAttestationSlots, attestation.SlotIndex)
					}
				}
			}
		}
	}

	return nil

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
				minipoolInfo.MissingAttestationSlots[slotIndex] = true
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
func createMinipoolIndexMap(bc beacon.Client, nodeDetails []*NodeSmoothingDetails) (map[uint64]*MinipoolInfo, error) {

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
func getSmoothingPoolNodeDetails(rp *rocketpool.RocketPool, opts *bind.CallOpts, elStartTime time.Time, elEndTime time.Time, nodeAddresses []common.Address) ([]*NodeSmoothingDetails, error) {

	intervalDuration := float64(elEndTime.Sub(elStartTime))

	// For each NO, get their opt-in status and time of last change in batches
	nodeCount := uint64(len(nodeAddresses))
	details := make([]*NodeSmoothingDetails, nodeCount)
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
				nodeDetails := &NodeSmoothingDetails{
					Address:          nodeAddresses[iterationIndex],
					Minipools:        []*MinipoolInfo{},
					SmoothingPoolEth: big.NewInt(0),
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
				if elStartTime.Sub(nodeDetails.StatusChangeTime) > 0 && !nodeDetails.IsOptedIn {
					nodeDetails.IsEligible = false
					nodeDetails.EligibilityFactor = 0
					details[iterationIndex] = nodeDetails
					return nil
				}

				// Get the node's total active factor
				if nodeDetails.IsOptedIn {
					nodeDetails.EligibilityFactor = float64(elEndTime.Sub(nodeDetails.StatusChangeTime)) / intervalDuration
				} else {
					nodeDetails.EligibilityFactor = float64(nodeDetails.StatusChangeTime.Sub(elStartTime)) / intervalDuration
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
							penaltyCount, err := minipool.GetMinipoolPenaltyCount(rp, mpd.Address, opts)
							if err != nil {
								return fmt.Errorf("Error getting penalty count for minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
							}
							if penaltyCount >= 3 {
								// This node is a cheater
								nodeDetails.IsEligible = false
								nodeDetails.EligibilityFactor = 0
								nodeDetails.Minipools = []*MinipoolInfo{}
								details[iterationIndex] = nodeDetails
								return nil
							}

							// This minipool is below the penalty count, so include it
							fee, err := mp.GetNodeFeeRaw(opts)
							if err != nil {
								return fmt.Errorf("Error getting fee for minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
							}
							nodeDetails.Minipools = append(nodeDetails.Minipools, &MinipoolInfo{
								Address:                 mpd.Address,
								ValidatorPubkey:         mpd.Pubkey,
								NodeAddress:             nodeDetails.Address,
								NodeIndex:               iterationIndex,
								Fee:                     fee,
								MissedAttestations:      0,
								GoodAttestations:        0,
								MissingAttestationSlots: map[uint64]bool{},
								WasActive:               true,
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

// Validates that the provided network is legal
func validateNetwork(rp *rocketpool.RocketPool, network uint64, validNetworkCache map[uint64]bool) (bool, error) {
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
