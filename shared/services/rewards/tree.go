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

// Holds information
type MinipoolPerformanceFile struct {
	Index               uint64                                               `json:"index"`
	Network             string                                               `json:"network"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance `json:"minipoolPerformance"`
}

// Minipool stats
type SmoothingPoolMinipoolPerformance struct {
	Pubkey                  string   `json:"pubkey"`
	SuccessfulAttestations  uint64   `json:"successfulAttestations"`
	MissedAttestations      uint64   `json:"missedAttestations"`
	ParticipationRate       float64  `json:"participationRate"`
	MissingAttestationSlots []uint64 `json:"missingAttestationSlots"`
	EthEarned               float64  `json:"ethEarned"`
}

// Node operator rewards
type NodeRewardsInfo struct {
	RewardNetwork                uint64        `json:"rewardNetwork"`
	CollateralRpl                *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl                 *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth             *QuotedBigInt `json:"smoothingPoolEth"`
	SmoothingPoolEligibilityRate float64       `json:"smoothingPoolEligibilityRate"`
	MerkleData                   []byte        `json:"-"`
	MerkleProof                  []string      `json:"merkleProof"`
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
	RewardsFileVersion         uint64                              `json:"rewardsFileVersion"`
	Index                      uint64                              `json:"index"`
	Network                    string                              `json:"network"`
	StartTime                  time.Time                           `json:"startTime,omitempty"`
	EndTime                    time.Time                           `json:"endTime"`
	ConsensusStartBlock        uint64                              `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock          uint64                              `json:"consensusEndBlock"`
	ExecutionStartBlock        uint64                              `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock          uint64                              `json:"executionEndBlock"`
	IntervalsPassed            uint64                              `json:"intervalsPassed"`
	MerkleRoot                 string                              `json:"merkleRoot,omitempty"`
	MinipoolPerformanceFileCID string                              `json:"minipoolPerformanceFileCid,omitempty"`
	TotalRewards               *TotalRewards                       `json:"totalRewards"`
	NetworkRewards             map[uint64]*NetworkRewardsInfo      `json:"networkRewards"`
	NodeRewards                map[common.Address]*NodeRewardsInfo `json:"nodeRewards"`
	MinipoolPerformanceFile    MinipoolPerformanceFile             `json:"-"`

	// Non-serialized fields
	MerkleTree           *merkletree.MerkleTree    `json:"-"`
	InvalidNetworkNodes  map[common.Address]uint64 `json:"-"`
	elSnapshotHeader     *types.Header             `json:"-"`
	log                  log.ColorLogger           `json:"-"`
	logPrefix            string                    `json:"-"`
	rp                   *rocketpool.RocketPool    `json:"-"`
	cfg                  *config.RocketPoolConfig  `json:"-"`
	bc                   beacon.Client             `json:"-"`
	opts                 *bind.CallOpts            `json:"-"`
	nodeAddresses        []common.Address          `json:"-"`
	nodeDetails          []*NodeSmoothingDetails   `json:"-"`
	smoothingPoolBalance *big.Int                  `json:"-"`
	smoothingPoolAddress common.Address            `json:"-"`
	intervalDutiesInfo   *IntervalDutiesInfo       `json:"-"`
	slotsPerEpoch        uint64                    `json:"-"`
	validatorIndexMap    map[uint64]*MinipoolInfo  `json:"-"`
	elStartTime          time.Time                 `json:"-"`
	elEndTime            time.Time                 `json:"-"`
	validNetworkCache    map[uint64]bool           `json:"-"`
	epsilon              *big.Int                  `json:"-"`
	intervalSeconds      *big.Int                  `json:"-"`
	beaconConfig         beacon.Eth2Config         `json:"-"`
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
		MinipoolPerformanceFile: MinipoolPerformanceFile{
			Index:               index,
			MinipoolPerformance: map[common.Address]*SmoothingPoolMinipoolPerformance{},
		},
		elSnapshotHeader: elSnapshotHeader,
		log:              log,
		logPrefix:        logPrefix,
	}
}

func (r *RewardsFile) GenerateTree(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) error {

	// Provision some struct params
	r.rp = rp
	r.cfg = cfg
	r.bc = bc
	r.validNetworkCache = map[uint64]bool{
		0: true,
	}

	// Set the network name
	r.Network = fmt.Sprint(cfg.Smartnode.Network.Value)
	r.MinipoolPerformanceFile.Network = r.Network

	// Get the addresses for all nodes
	r.opts = &bind.CallOpts{
		BlockNumber: r.elSnapshotHeader.Number,
	}
	nodeAddresses, err := node.GetNodeAddresses(rp, r.opts)
	if err != nil {
		return fmt.Errorf("Error getting node addresses: %w", err)
	}
	r.log.Printlnf("%s Creating tree for %d nodes", r.logPrefix, len(nodeAddresses))
	r.nodeAddresses = nodeAddresses

	// Get the minipool count - this will be used for an error epsilon due to division truncation
	minipoolCount, err := minipool.GetMinipoolCount(rp, r.opts)
	if err != nil {
		return fmt.Errorf("Error getting minipool count: %w", err)
	}
	r.epsilon = big.NewInt(int64(minipoolCount))

	// Calculate the RPL rewards
	err = r.calculateRplRewards()
	if err != nil {
		return fmt.Errorf("Error calculating RPL rewards: %w", err)
	}

	// Calculate the ETH rewards
	err = r.calculateEthRewards()
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
	for _, minipoolInfo := range r.MinipoolPerformanceFile.MinipoolPerformance {
		sort.Slice(minipoolInfo.MissingAttestationSlots, func(i, j int) bool {
			return minipoolInfo.MissingAttestationSlots[i] < minipoolInfo.MissingAttestationSlots[j]
		})
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

// Calculates the per-network distribution amounts and the total reward amounts
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
	}

}

// Calculates the RPL rewards for the given interval
func (r *RewardsFile) calculateRplRewards() error {

	snapshotBlockTime := time.Unix(int64(r.elSnapshotHeader.Time), 0)
	intervalDuration, err := rewards.GetClaimIntervalTime(r.rp, r.opts)
	if err != nil {
		return fmt.Errorf("error getting required registration time: %w", err)
	}

	// Handle node operator rewards
	nodeOpPercent, err := rewards.GetNodeOperatorRewardsPercent(r.rp, r.opts)
	if err != nil {
		return err
	}
	pendingRewards, err := rewards.GetPendingRPLRewards(r.rp, r.opts)
	if err != nil {
		return err
	}
	r.log.Printlnf("%s Pending RPL rewards: %s (%.3f)", r.logPrefix, pendingRewards.String(), eth.WeiToEth(pendingRewards))
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, eth.EthToWei(1))
	r.log.Printlnf("%s Total collateral RPL rewards: %s (%.3f)", r.logPrefix, totalNodeRewards.String(), eth.WeiToEth(totalNodeRewards))

	// Calculate the true effective stake of each node based on their participation in this interval
	totalNodeEffectiveStake := big.NewInt(0)
	trueNodeEffectiveStakes := map[common.Address]*big.Int{}
	intervalDurationBig := big.NewInt(int64(intervalDuration.Seconds()))
	for _, address := range r.nodeAddresses {
		// Get the node's effective stake
		nodeStake, err := node.GetNodeEffectiveRPLStake(r.rp, address, r.opts)
		if err != nil {
			return fmt.Errorf("error getting effective stake for node %s: %w", address.Hex(), err)
		}

		// Get the timestamp of the node's registration
		regTime, err := node.GetNodeRegistrationTime(r.rp, address, r.opts)
		if err != nil {
			return fmt.Errorf("error getting registration time for node %s: %w", address, err)
		}

		// Get the actual effective stake, scaled based on participation
		eligibleDuration := snapshotBlockTime.Sub(regTime)
		if eligibleDuration < intervalDuration {
			eligibleSeconds := big.NewInt(int64(eligibleDuration / time.Second))
			nodeStake.Mul(nodeStake, eligibleSeconds)
			nodeStake.Div(nodeStake, intervalDurationBig)
		}
		trueNodeEffectiveStakes[address] = nodeStake

		// Add it to the total
		totalNodeEffectiveStake.Add(totalNodeEffectiveStake, nodeStake)
	}

	for _, address := range r.nodeAddresses {
		// Get how much RPL goes to this node: (true effective stake) * (total node rewards) / (total true effective stake)
		nodeRplRewards := big.NewInt(0)
		nodeRplRewards.Mul(trueNodeEffectiveStakes[address], totalNodeRewards)
		nodeRplRewards.Div(nodeRplRewards, totalNodeEffectiveStake)

		// If there are pending rewards, add it to the map
		if nodeRplRewards.Cmp(big.NewInt(0)) == 1 {
			rewardsForNode, exists := r.NodeRewards[address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(r.rp, address, r.opts)
				if err != nil {
					return err
				}
				validNetwork, err := r.validateNetwork(network)
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

	// Sanity check to make sure we arrived at the correct total
	delta := big.NewInt(0)
	totalCalculatedNodeRewards := big.NewInt(0)
	for _, networkRewards := range r.NetworkRewards {
		totalCalculatedNodeRewards.Add(totalCalculatedNodeRewards, &networkRewards.CollateralRpl.Int)
	}
	delta.Sub(totalNodeRewards, totalCalculatedNodeRewards).Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return fmt.Errorf("error calculating collateral RPL: total was %s, but expected %s; error was too large", totalCalculatedNodeRewards.String(), totalNodeRewards.String())
	}
	r.TotalRewards.TotalCollateralRpl.Int = *totalCalculatedNodeRewards
	r.log.Printlnf("%s Calculated rewards:           %s (error = %s wei)", r.logPrefix, totalCalculatedNodeRewards.String(), delta.String())

	// Handle Oracle DAO rewards
	oDaoPercent, err := rewards.GetTrustedNodeOperatorRewardsPercent(r.rp, r.opts)
	if err != nil {
		return err
	}
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, eth.EthToWei(1))
	r.log.Printlnf("%s Total Oracle DAO RPL rewards: %s (%.3f)", r.logPrefix, totalODaoRewards.String(), eth.WeiToEth(totalODaoRewards))

	oDaoAddresses, err := trustednode.GetMemberAddresses(r.rp, r.opts)
	if err != nil {
		return err
	}

	// Calculate the true effective time of each oDAO node based on their participation in this interval
	totalODaoNodeTime := big.NewInt(0)
	trueODaoNodeTimes := map[common.Address]*big.Int{}
	for _, address := range oDaoAddresses {
		// Get the timestamp of the node's registration
		regTime, err := node.GetNodeRegistrationTime(r.rp, address, r.opts)
		if err != nil {
			return fmt.Errorf("error getting registration time for node %s: %w", address, err)
		}

		// Get the actual effective time, scaled based on participation
		participationTime := big.NewInt(0).Set(intervalDurationBig)
		eligibleDuration := snapshotBlockTime.Sub(regTime)
		if eligibleDuration < intervalDuration {
			participationTime = big.NewInt(int64(eligibleDuration.Seconds()))
		}
		trueODaoNodeTimes[address] = participationTime

		// Add it to the total
		totalODaoNodeTime.Add(totalODaoNodeTime, participationTime)
	}

	for _, address := range oDaoAddresses {
		// Calculate the oDAO rewards for the node: (participation time) * (total oDAO rewards) / (total participation time)
		individualOdaoRewards := big.NewInt(0)
		individualOdaoRewards.Mul(trueODaoNodeTimes[address], totalODaoRewards)
		individualOdaoRewards.Div(individualOdaoRewards, totalODaoNodeTime)

		rewardsForNode, exists := r.NodeRewards[address]
		if !exists {
			// Get the network the rewards should go to
			network, err := node.GetRewardNetwork(r.rp, address, r.opts)
			if err != nil {
				return err
			}
			validNetwork, err := r.validateNetwork(network)
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

	// Sanity check to make sure we arrived at the correct total
	totalCalculatedOdaoRewards := big.NewInt(0)
	delta = big.NewInt(0)
	for _, networkRewards := range r.NetworkRewards {
		totalCalculatedOdaoRewards.Add(totalCalculatedOdaoRewards, &networkRewards.OracleDaoRpl.Int)
	}
	delta.Sub(totalODaoRewards, totalCalculatedOdaoRewards).Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return fmt.Errorf("error calculating ODao RPL: total was %s, but expected %s; error was too large", totalCalculatedOdaoRewards.String(), totalODaoRewards.String())
	}
	r.TotalRewards.TotalOracleDaoRpl.Int = *totalCalculatedOdaoRewards
	r.log.Printlnf("%s Calculated rewards:           %s (error = %s wei)", r.logPrefix, totalCalculatedOdaoRewards.String(), delta.String())

	// Get expected Protocol DAO rewards
	pDaoPercent, err := rewards.GetProtocolDaoRewardsPercent(r.rp, r.opts)
	if err != nil {
		return err
	}
	pDaoRewards := NewQuotedBigInt(0)
	pDaoRewards.Mul(pendingRewards, pDaoPercent)
	pDaoRewards.Div(&pDaoRewards.Int, eth.EthToWei(1))
	r.log.Printlnf("%s Expected Protocol DAO rewards: %s (%.3f)", r.logPrefix, pDaoRewards.String(), eth.WeiToEth(&pDaoRewards.Int))

	// Get actual protocol DAO rewards
	pDaoRewards.Sub(pendingRewards, totalCalculatedNodeRewards)
	pDaoRewards.Sub(&pDaoRewards.Int, totalCalculatedOdaoRewards)
	r.TotalRewards.ProtocolDaoRpl = pDaoRewards
	r.log.Printlnf("%s Actual Protocol DAO rewards:   %s to account for truncation", r.logPrefix, pDaoRewards.String())

	return nil

}

// Calculates the ETH rewards for the given interval
func (r *RewardsFile) calculateEthRewards() error {

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := r.rp.GetContract("rocketSmoothingPool")
	if err != nil {
		return fmt.Errorf("error getting smoothing pool contract: %w", err)
	}
	r.smoothingPoolAddress = *smoothingPoolContract.Address

	r.smoothingPoolBalance, err = r.rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, r.elSnapshotHeader.Number)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}
	r.log.Printlnf("%s Smoothing Pool Balance: %s (%.3f)", r.logPrefix, r.smoothingPoolBalance.String(), eth.WeiToEth(r.smoothingPoolBalance))

	// Ignore the ETH calculation if there are no rewards
	if r.smoothingPoolBalance.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if r.Index == 0 {
		// This is the first interval, Smoothing Pool rewards are ignored on the first interval since it doesn't have a discrete start time
		return nil
	}

	// Get the event log interval
	var eventLogInterval int
	eventLogInterval, err = r.cfg.GetEventLogInterval()
	if err != nil {
		return err
	}

	// Get the Beacon config
	r.beaconConfig, err = r.bc.GetEth2Config()
	if err != nil {
		return err
	}
	r.slotsPerEpoch = r.beaconConfig.SlotsPerEpoch

	// Get the start time of this interval based on the event from the previous one
	previousIntervalEvent, err := rewards.GetRewardSnapshotEventWithUpgrades(r.rp, r.Index-1, big.NewInt(int64(eventLogInterval)), nil, r.cfg.Smartnode.GetPreviousRewardsPoolAddresses())
	if err != nil {
		return err
	}
	startElBlockHeader, err := r.getStartBlocksForInterval(previousIntervalEvent)
	if err != nil {
		return err
	}

	r.elStartTime = time.Unix(int64(startElBlockHeader.Time), 0)
	r.elEndTime = time.Unix(int64(r.elSnapshotHeader.Time), 0)
	r.intervalSeconds = big.NewInt(int64(r.elEndTime.Sub(r.elStartTime) / time.Second))

	// Get the details for nodes eligible for Smoothing Pool rewards
	// This should be all of the eth1 calls, so do them all at the start of Smoothing Pool calculation to prevent the need for an archive node during normal operations
	err = r.getSmoothingPoolNodeDetails()
	if err != nil {
		return err
	}
	eligible := 0
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.IsEligible {
			eligible++
		}
	}
	r.log.Printlnf("%s %d / %d nodes were eligible for Smoothing Pool rewards", r.logPrefix, eligible, len(r.nodeDetails))

	// Process the attestation performance for each minipool during this interval
	r.intervalDutiesInfo = &IntervalDutiesInfo{
		Index: r.Index,
		Slots: map[uint64]*SlotInfo{},
	}
	err = r.processAttestationsForInterval()
	if err != nil {
		return err
	}

	// Determine how much ETH each node gets and how much the pool stakers get
	poolStakerETH, nodeOpEth, err := r.calculateNodeRewards()
	if err != nil {
		return err
	}

	// Update the rewards maps
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.IsEligible && nodeInfo.SmoothingPoolEth.Cmp(big.NewInt(0)) > 0 {
			rewardsForNode, exists := r.NodeRewards[nodeInfo.Address]
			if !exists {
				// Get the network the rewards should go to
				network, err := node.GetRewardNetwork(r.rp, nodeInfo.Address, r.opts)
				if err != nil {
					return err
				}
				validNetwork, err := r.validateNetwork(network)
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
			rewardsForNode.SmoothingPoolEligibilityRate = float64(nodeInfo.EligibleSeconds.Uint64()) / float64(r.intervalSeconds.Uint64())

			// Add minipool rewards to the JSON
			for _, minipoolInfo := range nodeInfo.Minipools {
				if !minipoolInfo.WasActive {
					continue
				}
				performance := &SmoothingPoolMinipoolPerformance{
					Pubkey:                  minipoolInfo.ValidatorPubkey.Hex(),
					SuccessfulAttestations:  minipoolInfo.GoodAttestations,
					MissedAttestations:      minipoolInfo.MissedAttestations,
					ParticipationRate:       float64(minipoolInfo.GoodAttestations) / float64(minipoolInfo.GoodAttestations+minipoolInfo.MissedAttestations),
					EthEarned:               eth.WeiToEth(minipoolInfo.MinipoolShare),
					MissingAttestationSlots: []uint64{},
				}
				for slot := range minipoolInfo.MissingAttestationSlots {
					performance.MissingAttestationSlots = append(performance.MissingAttestationSlots, slot)
				}
				r.MinipoolPerformanceFile.MinipoolPerformance[minipoolInfo.Address] = performance
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
	r.TotalRewards.TotalSmoothingPoolEth.Int = *r.smoothingPoolBalance
	return nil

}

// Calculate the distribution of Smoothing Pool ETH to each node
func (r *RewardsFile) calculateNodeRewards() (*big.Int, *big.Int, error) {

	// Get the average fee for all eligible minipools and calculate their weighted share
	one := big.NewInt(1e18) // 100%, used for dividing percentages properly
	feeTotal := big.NewInt(0)
	minipoolCount := int64(0)
	minipoolShareTotal := big.NewInt(0)
	for _, nodeInfo := range r.nodeDetails {
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
				if r.intervalSeconds.Cmp(nodeInfo.EligibleSeconds) == 1 {
					// Scale the total shares by the eligibility factor based on how long the node has been opted in
					minipoolShare.Mul(minipoolShare, nodeInfo.EligibleSeconds)
					minipoolShare.Div(minipoolShare, r.intervalSeconds)
				}
				if minipool.MissedAttestations > 0 {
					// Calculate the participation rate if there are any missed attestations
					goodCount := big.NewInt(int64(minipool.GoodAttestations))
					missedCount := big.NewInt(int64(minipool.MissedAttestations))
					totalCount := big.NewInt(0).Add(goodCount, missedCount)
					minipoolShare.Mul(minipoolShare, goodCount)
					minipoolShare.Div(minipoolShare, totalCount)
				}
				minipoolShareTotal.Add(minipoolShareTotal, minipoolShare)
				minipool.MinipoolShare = minipoolShare
			}
		}
	}
	averageFee := big.NewInt(0).Div(feeTotal, big.NewInt(minipoolCount))

	// Calculate the staking pool share and the node op share
	halfSmoothingPool := big.NewInt(0).Div(r.smoothingPoolBalance, big.NewInt(2))
	commission := big.NewInt(0)
	commission.Mul(halfSmoothingPool, averageFee)
	commission.Div(commission, one)
	poolStakerShare := big.NewInt(0).Sub(halfSmoothingPool, commission)
	nodeOpShare := big.NewInt(0).Sub(r.smoothingPoolBalance, poolStakerShare)

	// Calculate the amount of ETH to give each minipool based on their share
	totalEthForMinipools := big.NewInt(0)
	for _, nodeInfo := range r.nodeDetails {
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
	truePoolStakerAmount := big.NewInt(0).Sub(r.smoothingPoolBalance, totalEthForMinipools)

	// Sanity check to make sure we arrived at the correct total
	delta := big.NewInt(0).Sub(totalEthForMinipools, nodeOpShare)
	delta.Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return nil, nil, fmt.Errorf("error calculating smoothing pool ETH: total was %s, but expected %s; error was too large (%s wei)", totalEthForMinipools.String(), nodeOpShare.String(), delta.String())
	}

	r.log.Printlnf("%s Pool staker ETH:    %s (%.3f)", r.logPrefix, poolStakerShare.String(), eth.WeiToEth(truePoolStakerAmount))
	r.log.Printlnf("%s Node Op ETH:        %s (%.3f)", r.logPrefix, nodeOpShare.String(), eth.WeiToEth(nodeOpShare))
	r.log.Printlnf("%s Calculated NO ETH:  %s (error = %s wei)", r.logPrefix, totalEthForMinipools.String(), delta.String())
	r.log.Printlnf("%s Adjusting pool staker ETH to %s to acount for truncation", r.logPrefix, truePoolStakerAmount.String())

	return truePoolStakerAmount, totalEthForMinipools, nil

}

// Get all of the duties for a range of epochs
func (r *RewardsFile) processAttestationsForInterval() error {

	startEpoch := r.ConsensusStartBlock / r.beaconConfig.SlotsPerEpoch
	endEpoch := r.ConsensusEndBlock / r.beaconConfig.SlotsPerEpoch

	// Determine the validator indices of each minipool
	err := r.createMinipoolIndexMap()
	if err != nil {
		return err
	}

	// Check all of the attestations for each epoch
	r.log.Printlnf("%s Checking participation of %d minipools for epochs %d to %d", r.logPrefix, len(r.validatorIndexMap), startEpoch, endEpoch)
	r.log.Printlnf("%s NOTE: this will take a long time, progress is reported every 100 epochs", r.logPrefix)

	epochsDone := 0
	reportStartTime := time.Now()
	for epoch := startEpoch; epoch < endEpoch+1; epoch++ {
		if epochsDone == 100 {
			timeTaken := time.Since(reportStartTime)
			r.log.Printlnf("%s On Epoch %d... (%s so far)", r.logPrefix, epoch, timeTaken)
			epochsDone = 0
		}

		err := r.processEpoch(true, epoch)
		if err != nil {
			return err
		}

		epochsDone++
	}

	// Check the epoch after the end of the interval for any lingering attestations
	epoch := endEpoch + 1
	err = r.processEpoch(false, epoch)
	if err != nil {
		return err
	}

	r.log.Printlnf("%s Finished participation check (total time = %s)", r.logPrefix, time.Since(reportStartTime))
	return nil

}

// Process an epoch, optionally getting the duties for all eligible minipools in it and checking each one's attestation performance
func (r *RewardsFile) processEpoch(getDuties bool, epoch uint64) error {

	// Get the committee info and attestation records for this epoch
	var committeeData []beacon.Committee
	attestationsPerSlot := make([][]beacon.AttestationInfo, r.slotsPerEpoch)
	var wg errgroup.Group

	if getDuties {
		wg.Go(func() error {
			var err error
			committeeData, err = r.bc.GetCommitteesForEpoch(&epoch)
			return err
		})
	}

	for i := uint64(0); i < r.slotsPerEpoch; i++ {
		i := i
		slot := epoch*r.slotsPerEpoch + i
		wg.Go(func() error {
			block, found, err := r.bc.GetBeaconBlock(fmt.Sprint(slot))
			if err != nil {
				return err
			}
			if found {
				attestationsPerSlot[i] = block.Attestations
			} else {
				attestationsPerSlot[i] = []beacon.AttestationInfo{}
			}
			return nil
		})
	}
	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("Error getting committee and attestaion records for epoch %d: %w", epoch, err)
	}

	if getDuties {
		// Get all of the expected duties for the epoch
		err = r.getDutiesForEpoch(committeeData)
		if err != nil {
			return fmt.Errorf("Error getting duties for epoch %d: %w", epoch, err)
		}
	}

	// Process all of the slots in the epoch
	for i := uint64(0); i < r.slotsPerEpoch; i++ {
		attestations := attestationsPerSlot[i]
		if len(attestations) > 0 {
			r.checkDutiesForSlot(attestations)
		}
	}

	return nil

}

// Handle all of the attestations in the given slot
func (r *RewardsFile) checkDutiesForSlot(attestations []beacon.AttestationInfo) error {

	// Go through the attestations for the block
	for _, attestation := range attestations {

		// Get the RP committees for this attestation's slot and index
		slotInfo, exists := r.intervalDutiesInfo.Slots[attestation.SlotIndex]
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
							delete(r.intervalDutiesInfo.Slots, attestation.SlotIndex)
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
func (r *RewardsFile) getDutiesForEpoch(committees []beacon.Committee) error {

	// Crawl the committees
	for _, committee := range committees {
		slotIndex := committee.Slot
		if slotIndex < r.ConsensusStartBlock || slotIndex > r.ConsensusEndBlock {
			// Ignore slots that are out of bounds
			continue
		}
		committeeIndex := committee.Index

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*MinipoolInfo{}
		for position, validator := range committee.Validators {
			minipoolInfo, exists := r.validatorIndexMap[validator]
			if exists {
				rpValidators[position] = minipoolInfo
				minipoolInfo.MissedAttestations += 1 // Consider this attestation missed until it's seen later
				minipoolInfo.MissingAttestationSlots[slotIndex] = true
			}
		}

		// If there are some RP validators, add this committee to the map
		if len(rpValidators) > 0 {
			slotInfo, exists := r.intervalDutiesInfo.Slots[slotIndex]
			if !exists {
				slotInfo = &SlotInfo{
					Index:      slotIndex,
					Committees: map[uint64]*CommitteeInfo{},
				}
				r.intervalDutiesInfo.Slots[slotIndex] = slotInfo
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
func (r *RewardsFile) createMinipoolIndexMap() error {

	// Make a slice of all minipool pubkeys
	minipoolPubkeys := []rptypes.ValidatorPubkey{}
	for _, details := range r.nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				minipoolPubkeys = append(minipoolPubkeys, minipoolInfo.ValidatorPubkey)
			}
		}
	}

	// Get indices for all minipool validators
	r.validatorIndexMap = map[uint64]*MinipoolInfo{}
	statusMap, err := r.bc.GetValidatorStatuses(minipoolPubkeys, &beacon.ValidatorStatusOptions{
		Slot: &r.ConsensusEndBlock,
	})
	if err != nil {
		return fmt.Errorf("Error getting validator statuses: %w", err)
	}
	for _, details := range r.nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				minipoolInfo.ValidatorIndex = statusMap[minipoolInfo.ValidatorPubkey].Index
				r.validatorIndexMap[minipoolInfo.ValidatorIndex] = minipoolInfo
			}
		}
	}

	return nil

}

// Get the details for every node that was opted into the Smoothing Pool for at least some portion of this interval
func (r *RewardsFile) getSmoothingPoolNodeDetails() error {

	// For each NO, get their opt-in status and time of last change in batches
	nodeCount := uint64(len(r.nodeAddresses))
	r.nodeDetails = make([]*NodeSmoothingDetails, nodeCount)
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
					Address:          r.nodeAddresses[iterationIndex],
					Minipools:        []*MinipoolInfo{},
					SmoothingPoolEth: big.NewInt(0),
				}

				// Check if the node is opted into the smoothing pool
				nodeDetails.IsOptedIn, err = node.GetSmoothingPoolRegistrationState(r.rp, nodeDetails.Address, r.opts)
				if err != nil {
					return fmt.Errorf("Error getting smoothing pool registration state for node %s: %w", nodeDetails.Address.Hex(), err)
				}

				// Get the time of the last registration change
				nodeDetails.StatusChangeTime, err = node.GetSmoothingPoolRegistrationChanged(r.rp, nodeDetails.Address, r.opts)
				if err != nil {
					return fmt.Errorf("Error getting smoothing pool registration change time for node %s: %w", nodeDetails.Address.Hex(), err)
				}

				// If the node isn't opted into the Smoothing Pool and they didn't opt out during this interval, ignore them
				if r.elStartTime.Sub(nodeDetails.StatusChangeTime) > 0 && !nodeDetails.IsOptedIn {
					nodeDetails.IsEligible = false
					nodeDetails.EligibleSeconds = big.NewInt(0)
					r.nodeDetails[iterationIndex] = nodeDetails
					return nil
				}

				// Get the node's total active factor
				if nodeDetails.IsOptedIn {
					nodeDetails.EligibleSeconds = big.NewInt(int64(r.elEndTime.Sub(nodeDetails.StatusChangeTime) / time.Second))
				} else {
					nodeDetails.EligibleSeconds = big.NewInt(int64(nodeDetails.StatusChangeTime.Sub(r.elStartTime) / time.Second))
				}
				if nodeDetails.EligibleSeconds.Cmp(r.intervalSeconds) == 1 {
					nodeDetails.EligibleSeconds.Set(r.intervalSeconds)
				}

				// Get the details for each minipool in the node
				minipoolDetails, err := minipool.GetNodeMinipools(r.rp, nodeDetails.Address, r.opts)
				if err != nil {
					return fmt.Errorf("Error getting minipool details for node %s: %w", nodeDetails.Address, err)
				}
				for _, mpd := range minipoolDetails {
					if mpd.Exists {
						mp, err := minipool.NewMinipool(r.rp, mpd.Address)
						if err != nil {
							return fmt.Errorf("Error creating minipool wrapper for minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
						}
						status, err := mp.GetStatus(r.opts)
						if err != nil {
							return fmt.Errorf("Error getting status of minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
						}
						if status == rptypes.Staking {
							penaltyCount, err := minipool.GetMinipoolPenaltyCount(r.rp, mpd.Address, r.opts)
							if err != nil {
								return fmt.Errorf("Error getting penalty count for minipool %s on node %s: %w", mpd.Address.Hex(), nodeDetails.Address.Hex(), err)
							}
							if penaltyCount >= 3 {
								// This node is a cheater
								nodeDetails.IsEligible = false
								nodeDetails.EligibleSeconds = big.NewInt(0)
								nodeDetails.Minipools = []*MinipoolInfo{}
								r.nodeDetails[iterationIndex] = nodeDetails
								return nil
							}

							// This minipool is below the penalty count, so include it
							fee, err := mp.GetNodeFeeRaw(r.opts)
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
				r.nodeDetails[iterationIndex] = nodeDetails
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return err
		}
	}

	return nil

}

// Validates that the provided network is legal
func (r *RewardsFile) validateNetwork(network uint64) (bool, error) {
	valid, exists := r.validNetworkCache[network]
	if !exists {
		var err error
		valid, err = tnsettings.GetNetworkEnabled(r.rp, big.NewInt(int64(network)), r.opts)
		if err != nil {
			return false, err
		}
		r.validNetworkCache[network] = valid
	}

	return valid, nil
}

// Gets the start blocks for the given interval
func (r *RewardsFile) getStartBlocksForInterval(previousIntervalEvent rewards.RewardsEvent) (*types.Header, error) {
	previousEpoch := previousIntervalEvent.ConsensusBlock.Uint64() / r.beaconConfig.SlotsPerEpoch
	nextEpoch := previousEpoch + 1
	r.ConsensusStartBlock = nextEpoch * r.beaconConfig.SlotsPerEpoch

	// Get the first block that isn't missing
	var elBlockNumber uint64
	for {
		beaconBlock, exists, err := r.bc.GetBeaconBlock(fmt.Sprint(r.ConsensusStartBlock))
		if err != nil {
			return nil, fmt.Errorf("error getting EL data for BC slot %d: %w", r.ConsensusStartBlock, err)
		}
		if !exists {
			r.ConsensusStartBlock++
		} else {
			elBlockNumber = beaconBlock.ExecutionBlockNumber
			break
		}
	}

	var startElHeader *types.Header
	var err error
	if elBlockNumber == 0 {
		// We are pre-merge, so get the first block after the one from the previous interval
		r.ExecutionStartBlock = previousIntervalEvent.ExecutionBlock.Uint64() + 1
		startElHeader, err = r.rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(r.ExecutionStartBlock)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL start block %d: %w", r.ExecutionStartBlock, err)
		}
	} else {
		// We are post-merge, so get the EL block corresponding to the BC block
		r.ExecutionStartBlock = elBlockNumber
		startElHeader, err = r.rp.Client.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL header for block %d: %w", elBlockNumber, err)
		}
	}

	return startElHeader, nil
}
