package rewards

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipfs/go-cid"
	"github.com/rocket-pool/rocketpool-go/rewards"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rewards/fees"
	"github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types"
	sszbig "github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types/big"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"golang.org/x/sync/errgroup"
)

// Type assertion to ensure SSZFile_v1 is IRewardsFile
var _ IRewardsFile = (*ssz_types.SSZFile_v1)(nil)

// Implementation for tree generator ruleset v9
type treeGeneratorImpl_v9_v10 struct {
	networkState                 *state.NetworkState
	rewardsFile                  *ssz_types.SSZFile_v1
	elSnapshotHeader             *types.Header
	snapshotEnd                  *SnapshotEnd
	log                          *log.ColorLogger
	logPrefix                    string
	rp                           RewardsExecutionClient
	previousRewardsPoolAddresses []common.Address
	bc                           RewardsBeaconClient
	opts                         *bind.CallOpts
	nodeDetails                  []*NodeSmoothingDetails
	smoothingPoolBalance         *big.Int
	intervalDutiesInfo           *IntervalDutiesInfo
	slotsPerEpoch                uint64
	validatorIndexMap            map[string]*MinipoolInfo
	elStartTime                  time.Time
	elEndTime                    time.Time
	validNetworkCache            map[uint64]bool
	epsilon                      *big.Int
	intervalSeconds              *big.Int
	beaconConfig                 beacon.Eth2Config
	validatorStatusMap           map[rptypes.ValidatorPubkey]beacon.ValidatorStatus
	totalAttestationScore        *big.Int
	successfulAttestations       uint64
	genesisTime                  time.Time
	invalidNetworkNodes          map[common.Address]uint64
	minipoolPerformanceFile      *MinipoolPerformanceFile_v2
	nodeRewards                  map[common.Address]*ssz_types.NodeReward
	networkRewards               map[ssz_types.Layer]*ssz_types.NetworkReward

	// fields for RPIP-62 bonus calculations
	// Withdrawals made by a minipool's validator.
	minipoolWithdrawals map[common.Address]*big.Int
}

// Create a new tree generator
func newTreeGeneratorImpl_v9_v10(rulesetVersion uint64, log *log.ColorLogger, logPrefix string, index uint64, snapshotEnd *SnapshotEnd, elSnapshotHeader *types.Header, intervalsPassed uint64, state *state.NetworkState) *treeGeneratorImpl_v9_v10 {
	return &treeGeneratorImpl_v9_v10{
		rewardsFile: &ssz_types.SSZFile_v1{
			RewardsFileVersion: 3,
			RulesetVersion:     rulesetVersion,
			Index:              index,
			IntervalsPassed:    intervalsPassed,
			TotalRewards: &ssz_types.TotalRewards{
				ProtocolDaoRpl:               sszbig.NewUint256(0),
				TotalCollateralRpl:           sszbig.NewUint256(0),
				TotalOracleDaoRpl:            sszbig.NewUint256(0),
				TotalSmoothingPoolEth:        sszbig.NewUint256(0),
				PoolStakerSmoothingPoolEth:   sszbig.NewUint256(0),
				NodeOperatorSmoothingPoolEth: sszbig.NewUint256(0),
				TotalNodeWeight:              sszbig.NewUint256(0),
			},
			NetworkRewards: ssz_types.NetworkRewards{},
			NodeRewards:    ssz_types.NodeRewards{},
		},
		validatorStatusMap:    map[rptypes.ValidatorPubkey]beacon.ValidatorStatus{},
		validatorIndexMap:     map[string]*MinipoolInfo{},
		elSnapshotHeader:      elSnapshotHeader,
		snapshotEnd:           snapshotEnd,
		log:                   log,
		logPrefix:             logPrefix,
		totalAttestationScore: big.NewInt(0),
		networkState:          state,
		invalidNetworkNodes:   map[common.Address]uint64{},
		minipoolPerformanceFile: &MinipoolPerformanceFile_v2{
			Index:               index,
			MinipoolPerformance: map[common.Address]*SmoothingPoolMinipoolPerformance_v2{},
		},
		nodeRewards:         map[common.Address]*ssz_types.NodeReward{},
		networkRewards:      map[ssz_types.Layer]*ssz_types.NetworkReward{},
		minipoolWithdrawals: map[common.Address]*big.Int{},
	}
}

// Get the version of the ruleset used by this generator
func (r *treeGeneratorImpl_v9_v10) getRulesetVersion() uint64 {
	return r.rewardsFile.RulesetVersion
}

func (r *treeGeneratorImpl_v9_v10) generateTree(rp RewardsExecutionClient, networkName string, previousRewardsPoolAddresses []common.Address, bc RewardsBeaconClient) (*GenerateTreeResult, error) {

	r.log.Printlnf("%s Generating tree using Ruleset v%d.", r.logPrefix, r.rewardsFile.RulesetVersion)

	// Provision some struct params
	r.rp = rp
	r.previousRewardsPoolAddresses = previousRewardsPoolAddresses
	r.bc = bc
	r.validNetworkCache = map[uint64]bool{
		0: true,
	}

	// Set the network name
	r.rewardsFile.Network, _ = ssz_types.NetworkFromString(networkName)
	r.minipoolPerformanceFile.Network = networkName
	r.minipoolPerformanceFile.RewardsFileVersion = r.rewardsFile.RewardsFileVersion
	r.minipoolPerformanceFile.RulesetVersion = r.rewardsFile.RulesetVersion

	// Get the Beacon config
	r.beaconConfig = r.networkState.BeaconConfig
	r.slotsPerEpoch = r.beaconConfig.SlotsPerEpoch
	r.genesisTime = time.Unix(int64(r.beaconConfig.GenesisTime), 0)

	// Set the EL client call opts
	r.opts = &bind.CallOpts{
		BlockNumber: r.elSnapshotHeader.Number,
	}

	r.log.Printlnf("%s Creating tree for %d nodes", r.logPrefix, len(r.networkState.NodeDetails))

	// Get the max of node count and minipool count - this will be used for an error epsilon due to division truncation
	nodeCount := len(r.networkState.NodeDetails)
	minipoolCount := len(r.networkState.MinipoolDetails)
	if nodeCount > minipoolCount {
		r.epsilon = big.NewInt(int64(nodeCount))
	} else {
		r.epsilon = big.NewInt(int64(minipoolCount))
	}

	// Calculate the RPL rewards
	err := r.calculateRplRewards()
	if err != nil {
		return nil, fmt.Errorf("error calculating RPL rewards: %w", err)
	}

	// Calculate the ETH rewards
	err = r.calculateEthRewards(true)
	if err != nil {
		return nil, fmt.Errorf("error calculating ETH rewards: %w", err)
	}

	// Sort and assign the maps to the ssz file lists
	for nodeAddress, nodeReward := range r.nodeRewards {
		copy(nodeReward.Address[:], nodeAddress[:])
		r.rewardsFile.NodeRewards = append(r.rewardsFile.NodeRewards, nodeReward)
	}

	for layer, networkReward := range r.networkRewards {
		networkReward.Network = layer
		r.rewardsFile.NetworkRewards = append(r.rewardsFile.NetworkRewards, networkReward)
	}

	// Generate the Merkle Tree
	err = r.rewardsFile.GenerateMerkleTree()
	if err != nil {
		return nil, fmt.Errorf("error generating Merkle tree: %w", err)
	}

	// Sort all of the missed attestations so the files are always generated in the same state
	for _, minipoolInfo := range r.minipoolPerformanceFile.MinipoolPerformance {
		sort.Slice(minipoolInfo.MissingAttestationSlots, func(i, j int) bool {
			return minipoolInfo.MissingAttestationSlots[i] < minipoolInfo.MissingAttestationSlots[j]
		})
	}

	return &GenerateTreeResult{
		RewardsFile:             r.rewardsFile,
		InvalidNetworkNodes:     r.invalidNetworkNodes,
		MinipoolPerformanceFile: r.minipoolPerformanceFile,
	}, nil

}

// Quickly calculates an approximate of the staker's share of the smoothing pool balance without processing Beacon performance
// Used for approximate returns in the rETH ratio update
func (r *treeGeneratorImpl_v9_v10) approximateStakerShareOfSmoothingPool(rp RewardsExecutionClient, networkName string, bc RewardsBeaconClient) (*big.Int, error) {
	r.log.Printlnf("%s Approximating tree using Ruleset v%d.", r.logPrefix, r.rewardsFile.RulesetVersion)

	r.rp = rp
	r.bc = bc
	r.validNetworkCache = map[uint64]bool{
		0: true,
	}

	// Set the network name
	r.rewardsFile.Network, _ = ssz_types.NetworkFromString(networkName)
	r.minipoolPerformanceFile.Network = networkName
	r.minipoolPerformanceFile.RewardsFileVersion = r.rewardsFile.RewardsFileVersion
	r.minipoolPerformanceFile.RulesetVersion = r.rewardsFile.RulesetVersion

	// Get the Beacon config
	r.beaconConfig = r.networkState.BeaconConfig
	r.slotsPerEpoch = r.beaconConfig.SlotsPerEpoch
	r.genesisTime = time.Unix(int64(r.beaconConfig.GenesisTime), 0)

	// Set the EL client call opts
	r.opts = &bind.CallOpts{
		BlockNumber: r.elSnapshotHeader.Number,
	}

	r.log.Printlnf("%s Creating tree for %d nodes", r.logPrefix, len(r.networkState.NodeDetails))

	// Get the max of node count and minipool count - this will be used for an error epsilon due to division truncation
	nodeCount := len(r.networkState.NodeDetails)
	minipoolCount := len(r.networkState.MinipoolDetails)
	if nodeCount > minipoolCount {
		r.epsilon = big.NewInt(int64(nodeCount))
	} else {
		r.epsilon = big.NewInt(int64(minipoolCount))
	}

	// Calculate the ETH rewards
	err := r.calculateEthRewards(false)
	if err != nil {
		return nil, fmt.Errorf("error calculating ETH rewards: %w", err)
	}

	return r.rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Int, nil
}

func (r *treeGeneratorImpl_v9_v10) calculateNodeRplRewards(
	collateralRewards *big.Int,
	nodeWeight *big.Int,
	totalNodeWeight *big.Int,
) *big.Int {

	if nodeWeight.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (collateralRewards * nodeWeight / totalNodeWeight)
	rpip30Rewards := big.NewInt(0).Mul(collateralRewards, nodeWeight)
	rpip30Rewards.Quo(rpip30Rewards, totalNodeWeight)

	return rpip30Rewards
}

// Calculates the RPL rewards for the given interval
func (r *treeGeneratorImpl_v9_v10) calculateRplRewards() error {
	pendingRewards := r.networkState.NetworkDetails.PendingRPLRewards
	r.log.Printlnf("%s Pending RPL rewards: %s (%.3f)", r.logPrefix, pendingRewards.String(), eth.WeiToEth(pendingRewards))
	if pendingRewards.Cmp(common.Big0) == 0 {
		return fmt.Errorf("there are no pending RPL rewards, so this interval cannot be used for rewards submission")
	}

	// Get baseline Protocol DAO rewards
	pDaoPercent := r.networkState.NetworkDetails.ProtocolDaoRewardsPercent
	pDaoRewards := big.NewInt(0)
	pDaoRewards.Mul(pendingRewards, pDaoPercent)
	pDaoRewards.Div(pDaoRewards, oneEth)
	r.log.Printlnf("%s Expected Protocol DAO rewards: %s (%.3f)", r.logPrefix, pDaoRewards.String(), eth.WeiToEth(pDaoRewards))

	// Get node operator rewards
	nodeOpPercent := r.networkState.NetworkDetails.NodeOperatorRewardsPercent
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, oneEth)
	r.log.Printlnf("%s Approx. total collateral RPL rewards: %s (%.3f)", r.logPrefix, totalNodeRewards.String(), eth.WeiToEth(totalNodeRewards))

	// Calculate the RPIP-30 weight of each node, scaling by their participation in this interval
	nodeWeights, totalNodeWeight, err := r.networkState.CalculateNodeWeights()
	if err != nil {
		return fmt.Errorf("error calculating node weights: %w", err)
	}

	// Operate normally if any node has rewards
	if totalNodeWeight.Sign() > 0 {
		// Make sure to record totalNodeWeight in the rewards file
		r.rewardsFile.TotalRewards.TotalNodeWeight.Set(totalNodeWeight)

		r.log.Printlnf("%s Calculating individual collateral rewards...", r.logPrefix)
		for i, nodeDetails := range r.networkState.NodeDetails {
			// Get how much RPL goes to this node
			nodeRplRewards := r.calculateNodeRplRewards(
				totalNodeRewards,
				nodeWeights[nodeDetails.NodeAddress],
				totalNodeWeight,
			)

			// If there are pending rewards, add it to the map
			if nodeRplRewards.Sign() == 1 {
				rewardsForNode, exists := r.nodeRewards[nodeDetails.NodeAddress]
				if !exists {
					// Get the network the rewards should go to
					network := r.networkState.NodeDetails[i].RewardNetwork.Uint64()
					validNetwork, err := r.validateNetwork(network)
					if err != nil {
						return err
					}
					if !validNetwork {
						network = 0
					}

					rewardsForNode = ssz_types.NewNodeReward(
						network,
						ssz_types.AddressFromBytes(nodeDetails.NodeAddress.Bytes()),
					)
					r.nodeRewards[nodeDetails.NodeAddress] = rewardsForNode
				}
				rewardsForNode.CollateralRpl.Add(rewardsForNode.CollateralRpl.Int, nodeRplRewards)

				// Add the rewards to the running total for the specified network
				rewardsForNetwork, exists := r.networkRewards[rewardsForNode.Network]
				if !exists {
					rewardsForNetwork = ssz_types.NewNetworkReward(rewardsForNode.Network)
					r.networkRewards[rewardsForNode.Network] = rewardsForNetwork
				}
				rewardsForNetwork.CollateralRpl.Int.Add(rewardsForNetwork.CollateralRpl.Int, nodeRplRewards)
			}
		}

		// Sanity check to make sure we arrived at the correct total
		delta := big.NewInt(0)
		totalCalculatedNodeRewards := big.NewInt(0)
		for _, networkRewards := range r.networkRewards {
			totalCalculatedNodeRewards.Add(totalCalculatedNodeRewards, networkRewards.CollateralRpl.Int)
		}
		delta.Sub(totalNodeRewards, totalCalculatedNodeRewards).Abs(delta)
		if delta.Cmp(r.epsilon) == 1 {
			return fmt.Errorf("error calculating collateral RPL: total was %s, but expected %s; error was too large", totalCalculatedNodeRewards.String(), totalNodeRewards.String())
		}
		r.rewardsFile.TotalRewards.TotalCollateralRpl.Int.Set(totalCalculatedNodeRewards)
		r.log.Printlnf("%s Calculated rewards:           %s (error = %s wei)", r.logPrefix, totalCalculatedNodeRewards.String(), delta.String())
		pDaoRewards.Sub(pendingRewards, totalCalculatedNodeRewards)
	} else {
		// In this situation, none of the nodes in the network had eligible rewards so send it all to the pDAO
		pDaoRewards.Add(pDaoRewards, totalNodeRewards)
		r.log.Printlnf("%s None of the nodes were eligible for collateral rewards, sending everything to the pDAO; now at %s (%.3f)", r.logPrefix, pDaoRewards.String(), eth.WeiToEth(pDaoRewards))
	}

	// Handle Oracle DAO rewards
	oDaoPercent := r.networkState.NetworkDetails.TrustedNodeOperatorRewardsPercent
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, oneEth)
	r.log.Printlnf("%s Total Oracle DAO RPL rewards: %s (%.3f)", r.logPrefix, totalODaoRewards.String(), eth.WeiToEth(totalODaoRewards))

	oDaoDetails := r.networkState.OracleDaoMemberDetails

	// Calculate the true effective time of each oDAO node based on their participation in this interval
	totalODaoNodeTime := big.NewInt(0)
	trueODaoNodeTimes := map[common.Address]*big.Int{}
	for _, details := range oDaoDetails {
		// Get the timestamp of the node joining the oDAO
		joinTime := details.JoinedTime

		// Get the actual effective time, scaled based on participation
		intervalDuration := r.networkState.NetworkDetails.IntervalDuration
		intervalDurationBig := big.NewInt(int64(intervalDuration.Seconds()))
		participationTime := big.NewInt(0).Set(intervalDurationBig)
		snapshotBlockTime := time.Unix(int64(r.elSnapshotHeader.Time), 0)
		eligibleDuration := snapshotBlockTime.Sub(joinTime)
		if eligibleDuration < intervalDuration {
			participationTime = big.NewInt(int64(eligibleDuration.Seconds()))
		}
		trueODaoNodeTimes[details.Address] = participationTime

		// Add it to the total
		totalODaoNodeTime.Add(totalODaoNodeTime, participationTime)
	}

	for _, details := range oDaoDetails {
		address := details.Address

		// Calculate the oDAO rewards for the node: (participation time) * (total oDAO rewards) / (total participation time)
		individualOdaoRewards := big.NewInt(0)
		individualOdaoRewards.Mul(trueODaoNodeTimes[address], totalODaoRewards)
		individualOdaoRewards.Div(individualOdaoRewards, totalODaoNodeTime)

		rewardsForNode, exists := r.nodeRewards[address]
		if !exists {
			// Get the network the rewards should go to
			network := r.networkState.NodeDetailsByAddress[address].RewardNetwork.Uint64()
			validNetwork, err := r.validateNetwork(network)
			if err != nil {
				return err
			}
			if !validNetwork {
				r.invalidNetworkNodes[address] = network
				network = 0
			}

			rewardsForNode = ssz_types.NewNodeReward(
				network,
				ssz_types.AddressFromBytes(address.Bytes()),
			)
			r.nodeRewards[address] = rewardsForNode

		}
		rewardsForNode.OracleDaoRpl.Add(rewardsForNode.OracleDaoRpl.Int, individualOdaoRewards)

		// Add the rewards to the running total for the specified network
		rewardsForNetwork, exists := r.networkRewards[rewardsForNode.Network]
		if !exists {
			rewardsForNetwork = ssz_types.NewNetworkReward(rewardsForNode.Network)
			r.networkRewards[rewardsForNode.Network] = rewardsForNetwork
		}
		rewardsForNetwork.OracleDaoRpl.Add(rewardsForNetwork.OracleDaoRpl.Int, individualOdaoRewards)
	}

	// Sanity check to make sure we arrived at the correct total
	totalCalculatedOdaoRewards := big.NewInt(0)
	delta := big.NewInt(0)
	for _, networkRewards := range r.networkRewards {
		totalCalculatedOdaoRewards.Add(totalCalculatedOdaoRewards, networkRewards.OracleDaoRpl.Int)
	}
	delta.Sub(totalODaoRewards, totalCalculatedOdaoRewards).Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return fmt.Errorf("error calculating ODao RPL: total was %s, but expected %s; error was too large", totalCalculatedOdaoRewards.String(), totalODaoRewards.String())
	}
	r.rewardsFile.TotalRewards.TotalOracleDaoRpl.Int.Set(totalCalculatedOdaoRewards)
	r.log.Printlnf("%s Calculated rewards:           %s (error = %s wei)", r.logPrefix, totalCalculatedOdaoRewards.String(), delta.String())

	// Get actual protocol DAO rewards
	pDaoRewards.Sub(pDaoRewards, totalCalculatedOdaoRewards)
	r.rewardsFile.TotalRewards.ProtocolDaoRpl = sszbig.NewUint256(0)
	r.rewardsFile.TotalRewards.ProtocolDaoRpl.Set(pDaoRewards)
	r.log.Printlnf("%s Actual Protocol DAO rewards:  %s to account for truncation", r.logPrefix, pDaoRewards.String())

	// Print total node weight
	r.log.Printlnf("%s Total Node Weight:            %s", r.logPrefix, totalNodeWeight)

	return nil

}

// Calculates the ETH rewards for the given interval
func (r *treeGeneratorImpl_v9_v10) calculateEthRewards(checkBeaconPerformance bool) error {

	// Get the Smoothing Pool contract's balance
	r.smoothingPoolBalance = r.networkState.NetworkDetails.SmoothingPoolBalance
	r.log.Printlnf("%s Smoothing Pool Balance: %s (%.3f)", r.logPrefix, r.smoothingPoolBalance.String(), eth.WeiToEth(r.smoothingPoolBalance))

	// Ignore the ETH calculation if there are no rewards
	if r.smoothingPoolBalance.Cmp(common.Big0) == 0 {
		return nil
	}

	if r.rewardsFile.Index == 0 {
		// This is the first interval, Smoothing Pool rewards are ignored on the first interval since it doesn't have a discrete start time
		return nil
	}

	// Get the start time of this interval based on the event from the previous one
	//previousIntervalEvent, err := GetRewardSnapshotEvent(r.rp, r.cfg, r.rewardsFile.Index-1, r.opts) // This is immutable so querying at the head is fine and mitigates issues around calls for pruned EL state
	previousIntervalEvent, err := r.rp.GetRewardSnapshotEvent(r.previousRewardsPoolAddresses, r.rewardsFile.Index-1, r.opts)
	if err != nil {
		return err
	}
	startElBlockHeader, err := r.getBlocksAndTimesForInterval(previousIntervalEvent)
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
		Index: r.rewardsFile.Index,
		Slots: map[uint64]*SlotInfo{},
	}
	if checkBeaconPerformance {
		err = r.processAttestationsBalancesAndWithdrawalsForInterval()
		if err != nil {
			return err
		}
	} else {
		// Attestation processing is disabled, just give each minipool 1 good attestation and complete slot activity so they're all scored the same
		// Used for approximating rETH's share during balances calculation
		validatorReq := big.NewInt(0).Set(thirtyTwoEth)
		for _, nodeInfo := range r.nodeDetails {
			// Check if the node is currently opted in for simplicity
			if nodeInfo.IsEligible && nodeInfo.IsOptedIn && r.elEndTime.After(nodeInfo.OptInTime) {
				eligibleBorrowedEth := nodeInfo.EligibleBorrowedEth
				_, percentOfBorrowedEth := r.networkState.GetStakedRplValueInEthAndPercentOfBorrowedEth(eligibleBorrowedEth, nodeInfo.RplStake)
				for _, minipool := range nodeInfo.Minipools {
					minipool.CompletedAttestations = map[uint64]bool{0: true}

					// Make up an attestation
					details := r.networkState.MinipoolDetailsByAddress[minipool.Address]
					bond, fee := details.GetMinipoolBondAndNodeFee(r.elEndTime)
					if r.rewardsFile.RulesetVersion >= 10 {
						fee = fees.GetMinipoolFeeWithBonus(bond, fee, percentOfBorrowedEth)
					}
					minipoolScore := big.NewInt(0).Sub(oneEth, fee) // 1 - fee
					minipoolScore.Mul(minipoolScore, bond)          // Multiply by bond
					minipoolScore.Div(minipoolScore, validatorReq)  // Divide by 32 to get the bond as a fraction of a total validator
					minipoolScore.Add(minipoolScore, fee)           // Total = fee + (bond/32)(1 - fee)

					// Add it to the minipool's score and the total score
					minipool.AttestationScore.Add(&minipool.AttestationScore.Int, minipoolScore)
					r.totalAttestationScore.Add(r.totalAttestationScore, minipoolScore)

					r.successfulAttestations++
				}
			}
		}
	}

	// Determine how much ETH each node gets and how much the pool stakers get
	poolStakerETH, nodeOpEth, bonusScalar, err := r.calculateNodeRewards()
	if err != nil {
		return err
	}
	if r.rewardsFile.RulesetVersion >= 10 {
		r.minipoolPerformanceFile.BonusScalar = QuotedBigIntFromBigInt(bonusScalar)
	}

	// Update the rewards maps
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.IsEligible && nodeInfo.SmoothingPoolEth.Cmp(common.Big0) > 0 {
			rewardsForNode, exists := r.nodeRewards[nodeInfo.Address]
			if !exists {
				network := nodeInfo.RewardsNetwork
				validNetwork, err := r.validateNetwork(network)
				if err != nil {
					return err
				}
				if !validNetwork {
					r.invalidNetworkNodes[nodeInfo.Address] = network
					network = 0
				}

				rewardsForNode = ssz_types.NewNodeReward(
					network,
					ssz_types.AddressFromBytes(nodeInfo.Address.Bytes()),
				)
				r.nodeRewards[nodeInfo.Address] = rewardsForNode
			}
			rewardsForNode.SmoothingPoolEth.Add(rewardsForNode.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)

			// Add minipool rewards to the JSON
			for _, minipoolInfo := range nodeInfo.Minipools {
				successfulAttestations := uint64(len(minipoolInfo.CompletedAttestations))
				missingAttestations := uint64(len(minipoolInfo.MissingAttestationSlots))
				performance := &SmoothingPoolMinipoolPerformance_v2{
					Pubkey:                  minipoolInfo.ValidatorPubkey.Hex(),
					SuccessfulAttestations:  successfulAttestations,
					MissedAttestations:      missingAttestations,
					AttestationScore:        minipoolInfo.AttestationScore,
					EthEarned:               QuotedBigIntFromBigInt(minipoolInfo.MinipoolShare),
					BonusEthEarned:          QuotedBigIntFromBigInt(minipoolInfo.MinipoolBonus),
					ConsensusIncome:         minipoolInfo.ConsensusIncome,
					EffectiveCommission:     QuotedBigIntFromBigInt(minipoolInfo.TotalFee),
					MissingAttestationSlots: []uint64{},
				}
				if successfulAttestations+missingAttestations == 0 {
					// Don't include minipools that have zero attestations
					continue
				}
				for slot := range minipoolInfo.MissingAttestationSlots {
					performance.MissingAttestationSlots = append(performance.MissingAttestationSlots, slot)
				}
				r.minipoolPerformanceFile.MinipoolPerformance[minipoolInfo.Address] = performance
			}

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.networkRewards[rewardsForNode.Network]
			if !exists {
				rewardsForNetwork = ssz_types.NewNetworkReward(rewardsForNode.Network)
				r.networkRewards[rewardsForNode.Network] = rewardsForNetwork
			}
			rewardsForNetwork.SmoothingPoolEth.Add(rewardsForNetwork.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
		}
	}

	// Set the totals
	r.rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Set(poolStakerETH)
	r.rewardsFile.TotalRewards.NodeOperatorSmoothingPoolEth.Set(nodeOpEth)
	r.rewardsFile.TotalRewards.TotalSmoothingPoolEth.Set(r.smoothingPoolBalance)
	return nil

}

var oneEth = big.NewInt(1000000000000000000)
var eightEth = big.NewInt(0).Mul(oneEth, big.NewInt(8))
var fourteenPercentEth = big.NewInt(14e16)
var thirtyTwoEth = big.NewInt(0).Mul(oneEth, big.NewInt(32))

func (r *treeGeneratorImpl_v9_v10) calculateNodeBonuses() (*big.Int, error) {
	totalConsensusBonus := big.NewInt(0)
	for _, nsd := range r.nodeDetails {
		if !nsd.IsEligible {
			continue
		}

		nodeDetails := r.networkState.NodeDetailsByAddress[nsd.Address]
		eligible, _, eligibleEnd := nodeDetails.IsEligibleForBonuses(r.elStartTime, r.elEndTime)
		if !eligible {
			continue
		}

		// Get the nodeDetails from the network state
		eligibleBorrowedEth := nsd.EligibleBorrowedEth
		_, percentOfBorrowedEth := r.networkState.GetStakedRplValueInEthAndPercentOfBorrowedEth(eligibleBorrowedEth, nsd.RplStake)
		for _, mpd := range nsd.Minipools {
			mpi := r.networkState.MinipoolDetailsByAddress[mpd.Address]
			if !mpi.IsEligibleForBonuses(eligibleEnd) {
				continue
			}
			bond, fee := mpi.GetMinipoolBondAndNodeFee(eligibleEnd)
			feeWithBonus := fees.GetMinipoolFeeWithBonus(bond, fee, percentOfBorrowedEth)
			if fee.Cmp(feeWithBonus) >= 0 {
				// This minipool won't get any bonuses, so skip it
				continue
			}
			// This minipool will get a bonus
			// It is safe to populate the optional fields from here on.

			fee = feeWithBonus
			// Save fee as totalFee for the Minipool
			mpd.TotalFee = fee

			// Total fee for a minipool with a bonus shall never exceed 14%
			if fee.Cmp(fourteenPercentEth) > 0 {
				r.log.Printlnf("WARNING: Minipool %s has a fee of %s, which is greater than the maximum allowed of 14%", mpd.Address.Hex(), fee.String())
				r.log.Printlnf("WARNING: Aborting.")
				return nil, fmt.Errorf("minipool %s has a fee of %s, which is greater than the maximum allowed of 14%%", mpd.Address.Hex(), fee.String())
			}
			bonusFee := big.NewInt(0).Set(fee)
			bonusFee.Sub(bonusFee, mpi.NodeFee)
			withdrawalTotal := r.minipoolWithdrawals[mpd.Address]
			if withdrawalTotal == nil {
				withdrawalTotal = big.NewInt(0)
			}
			consensusIncome := big.NewInt(0).Set(withdrawalTotal)
			mpd.ConsensusIncome = &QuotedBigInt{Int: *(big.NewInt(0).Set(consensusIncome))}
			bonusShare := bonusFee.Mul(bonusFee, big.NewInt(0).Sub(thirtyTwoEth, mpi.NodeDepositBalance))
			bonusShare.Div(bonusShare, thirtyTwoEth)
			minipoolBonus := consensusIncome.Mul(consensusIncome, bonusShare)
			minipoolBonus.Div(minipoolBonus, oneEth)
			if minipoolBonus.Sign() == -1 {
				minipoolBonus = big.NewInt(0)
			}
			mpd.MinipoolBonus = minipoolBonus
			totalConsensusBonus.Add(totalConsensusBonus, minipoolBonus)
			nsd.BonusEth.Add(nsd.BonusEth, minipoolBonus)
		}
	}
	return totalConsensusBonus, nil
}

// Calculate the distribution of Smoothing Pool ETH to each node
func (r *treeGeneratorImpl_v9_v10) calculateNodeRewards() (*big.Int, *big.Int, *big.Int, error) {
	var err error
	bonusScalar := big.NewInt(0).Set(oneEth)

	// If there weren't any successful attestations, everything goes to the pool stakers
	if r.totalAttestationScore.Cmp(common.Big0) == 0 || r.successfulAttestations == 0 {
		r.log.Printlnf("WARNING: Total attestation score = %s, successful attestations = %d... sending the whole smoothing pool balance to the pool stakers.", r.totalAttestationScore.String(), r.successfulAttestations)
		return r.smoothingPoolBalance, big.NewInt(0), bonusScalar, nil
	}

	// Calculate the minipool bonuses
	isEligibleInterval := true // TODO - check on-chain for saturn 1
	var totalConsensusBonus *big.Int
	if r.rewardsFile.RulesetVersion >= 10 && isEligibleInterval {
		totalConsensusBonus, err = r.calculateNodeBonuses()
		if err != nil {
			return nil, nil, nil, err
		}
	}

	totalEthForMinipools := big.NewInt(0)
	totalNodeOpShare := big.NewInt(0)
	totalNodeOpShare.Mul(r.smoothingPoolBalance, r.totalAttestationScore)
	totalNodeOpShare.Div(totalNodeOpShare, big.NewInt(int64(r.successfulAttestations)))
	totalNodeOpShare.Div(totalNodeOpShare, oneEth)

	for _, nodeInfo := range r.nodeDetails {
		nodeInfo.SmoothingPoolEth = big.NewInt(0)
		if !nodeInfo.IsEligible {
			continue
		}
		for _, minipool := range nodeInfo.Minipools {
			if len(minipool.CompletedAttestations)+len(minipool.MissingAttestationSlots) == 0 || !minipool.WasActive {
				// Ignore minipools that weren't active for the interval
				minipool.WasActive = false
				minipool.MinipoolShare = big.NewInt(0)
				continue
			}

			minipoolEth := big.NewInt(0).Set(totalNodeOpShare)
			minipoolEth.Mul(minipoolEth, &minipool.AttestationScore.Int)
			minipoolEth.Div(minipoolEth, r.totalAttestationScore)
			minipool.MinipoolShare = minipoolEth
			nodeInfo.SmoothingPoolEth.Add(nodeInfo.SmoothingPoolEth, minipoolEth)
		}
		totalEthForMinipools.Add(totalEthForMinipools, nodeInfo.SmoothingPoolEth)
	}

	if r.rewardsFile.RulesetVersion >= 10 {
		remainingBalance := big.NewInt(0).Sub(r.smoothingPoolBalance, totalEthForMinipools)
		if remainingBalance.Cmp(totalConsensusBonus) < 0 {
			r.log.Printlnf("WARNING: Remaining balance is less than total consensus bonus... Balance = %s, total consensus bonus = %s", remainingBalance.String(), totalConsensusBonus.String())
			// Scale bonuses down to fit the remaining balance
			bonusScalar.Div(big.NewInt(0).Mul(remainingBalance, oneEth), totalConsensusBonus)
			for _, nsd := range r.nodeDetails {
				nsd.BonusEth.Mul(nsd.BonusEth, remainingBalance)
				nsd.BonusEth.Div(nsd.BonusEth, totalConsensusBonus)
				// Calculate the reduced bonus for each minipool
				// Because of integer division, this will be less than the actual bonus by up to 1 wei
				for _, mpd := range nsd.Minipools {
					if mpd.MinipoolBonus == nil {
						continue
					}
					mpd.MinipoolBonus.Mul(mpd.MinipoolBonus, remainingBalance)
					mpd.MinipoolBonus.Div(mpd.MinipoolBonus, totalConsensusBonus)
				}
			}
		}
	}

	// Sanity check the totalNodeOpShare before bonuses are awarded
	delta := big.NewInt(0).Sub(totalEthForMinipools, totalNodeOpShare)
	delta.Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return nil, nil, nil, fmt.Errorf("error calculating smoothing pool ETH: total was %s, but expected %s; error was too large (%s wei)", totalEthForMinipools.String(), totalNodeOpShare.String(), delta.String())
	}

	// Finally, award the bonuses
	if r.rewardsFile.RulesetVersion >= 10 {
		for _, nsd := range r.nodeDetails {
			nsd.SmoothingPoolEth.Add(nsd.SmoothingPoolEth, nsd.BonusEth)
			totalEthForMinipools.Add(totalEthForMinipools, nsd.BonusEth)
		}
	}

	// This is how much actually goes to the pool stakers - it should ideally be equal to poolStakerShare but this accounts for any cumulative floating point errors
	truePoolStakerAmount := big.NewInt(0).Sub(r.smoothingPoolBalance, totalEthForMinipools)

	// Calculate the staking pool share and the node op share
	poolStakerShareBeforeBonuses := big.NewInt(0).Sub(r.smoothingPoolBalance, totalNodeOpShare)

	r.log.Printlnf("%s Pool staker ETH before bonuses:    %s (%.3f)", r.logPrefix, poolStakerShareBeforeBonuses.String(), eth.WeiToEth(poolStakerShareBeforeBonuses))
	r.log.Printlnf("%s Pool staker ETH after bonuses:     %s (%.3f)", r.logPrefix, truePoolStakerAmount.String(), eth.WeiToEth(truePoolStakerAmount))
	r.log.Printlnf("%s Node Op ETH before bonuses:        %s (%.3f)", r.logPrefix, totalNodeOpShare.String(), eth.WeiToEth(totalNodeOpShare))
	r.log.Printlnf("%s Node Op ETH after bonuses:         %s (%.3f)", r.logPrefix, totalEthForMinipools.String(), eth.WeiToEth(totalEthForMinipools))
	r.log.Printlnf("%s (error = %s wei)", r.logPrefix, delta.String())
	r.log.Printlnf("%s Adjusting pool staker ETH to %s to account for truncation", r.logPrefix, truePoolStakerAmount.String())

	return truePoolStakerAmount, totalEthForMinipools, bonusScalar, nil

}

// Get all of the duties for a range of epochs
func (r *treeGeneratorImpl_v9_v10) processAttestationsBalancesAndWithdrawalsForInterval() error {

	startEpoch := r.rewardsFile.ConsensusStartBlock / r.beaconConfig.SlotsPerEpoch
	endEpoch := r.rewardsFile.ConsensusEndBlock / r.beaconConfig.SlotsPerEpoch

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
			r.log.Printlnf("%s On Epoch %d of %d (%.2f%%)... (%s so far)", r.logPrefix, epoch, endEpoch, float64(epoch-startEpoch)/float64(endEpoch-startEpoch)*100.0, timeTaken)
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
func (r *treeGeneratorImpl_v9_v10) processEpoch(duringInterval bool, epoch uint64) error {

	// Get the committee info and attestation records for this epoch
	var committeeData beacon.Committees
	attestationsPerSlot := make([][]beacon.AttestationInfo, r.slotsPerEpoch)
	var wg errgroup.Group

	if duringInterval {
		wg.Go(func() error {
			var err error
			committeeData, err = r.bc.GetCommitteesForEpoch(&epoch)
			return err
		})
	}

	withdrawalsLock := &sync.Mutex{}
	for i := uint64(0); i < r.slotsPerEpoch; i++ {
		// Get the beacon block for this slot
		i := i
		slot := epoch*r.slotsPerEpoch + i
		slotTime := r.networkState.BeaconConfig.GetSlotTime(slot)
		wg.Go(func() error {
			beaconBlock, found, err := r.bc.GetBeaconBlock(fmt.Sprint(slot))
			if err != nil {
				return err
			}
			if found {
				attestationsPerSlot[i] = beaconBlock.Attestations
			}

			// If we don't need withdrawal amounts because we're using ruleset 9,
			// return early
			if r.rewardsFile.RulesetVersion < 10 || !duringInterval {
				return nil
			}

			for _, withdrawal := range beaconBlock.Withdrawals {
				// Ignore non-RP validators
				mpi, exists := r.validatorIndexMap[withdrawal.ValidatorIndex]
				if !exists {
					continue
				}
				nnd := r.networkState.NodeDetailsByAddress[mpi.NodeAddress]
				nmd := r.networkState.MinipoolDetailsByAddress[mpi.Address]

				// Check that the node is opted into the SP during this slot
				if !nnd.WasOptedInAt(slotTime) {
					continue
				}

				// Check that the minipool's bond is eligible for bonuses at this slot
				if eligible := nmd.IsEligibleForBonuses(slotTime); !eligible {
					continue
				}

				// If the withdrawal is in or after the minipool's withdrawable epoch, adjust it.
				withdrawalAmount := withdrawal.Amount
				validatorInfo := r.networkState.MinipoolValidatorDetails[mpi.ValidatorPubkey]
				if slot >= r.networkState.BeaconConfig.FirstSlotOfEpoch(validatorInfo.WithdrawableEpoch) {
					// Subtract 32 ETH from the withdrawal amount
					withdrawalAmount = big.NewInt(0).Sub(withdrawalAmount, thirtyTwoEth)
					// max(withdrawalAmount, 0)
					if withdrawalAmount.Sign() < 0 {
						withdrawalAmount.SetInt64(0)
					}
				}

				// Create the minipool's withdrawal sum big.Int if it doesn't exist
				withdrawalsLock.Lock()
				if r.minipoolWithdrawals[mpi.Address] == nil {
					r.minipoolWithdrawals[mpi.Address] = big.NewInt(0)
				}
				// Add the withdrawal amount
				r.minipoolWithdrawals[mpi.Address].Add(r.minipoolWithdrawals[mpi.Address], withdrawalAmount)
				withdrawalsLock.Unlock()
			}
			return nil
		})
	}
	err := wg.Wait()
	// Return preallocated memory to the pool if it exists
	if committeeData != nil {
		defer committeeData.Release()
	}
	if err != nil {
		return fmt.Errorf("error getting committee and attestaion records for epoch %d: %w", epoch, err)
	}

	if duringInterval {
		// Get all of the expected duties for the epoch
		err = r.getDutiesForEpoch(committeeData)
		if err != nil {
			return fmt.Errorf("error getting duties for epoch %d: %w", epoch, err)
		}
	}

	// Process all of the slots in the epoch
	for i := uint64(0); i < r.slotsPerEpoch; i++ {
		inclusionSlot := epoch*r.slotsPerEpoch + i
		attestations := attestationsPerSlot[i]
		if len(attestations) > 0 {
			r.checkAttestations(attestations, inclusionSlot)
		}
	}

	return nil

}

func (r *treeGeneratorImpl_v9_v10) checkAttestations(attestations []beacon.AttestationInfo, inclusionSlot uint64) error {

	// Go through the attestations for the block
	for _, attestation := range attestations {
		// Get the RP committees for this attestation's slot and index
		slotInfo, exists := r.intervalDutiesInfo.Slots[attestation.SlotIndex]
		if !exists {
			continue
		}
		// Ignore attestations delayed by more than 32 slots
		if inclusionSlot-attestation.SlotIndex > r.beaconConfig.SlotsPerEpoch {
			continue
		}
		rpCommittee, exists := slotInfo.Committees[attestation.CommitteeIndex]
		if !exists {
			continue
		}
		blockTime := r.genesisTime.Add(time.Second * time.Duration(r.networkState.BeaconConfig.SecondsPerSlot*attestation.SlotIndex))

		// Check if each RP validator attested successfully
		for position, validator := range rpCommittee.Positions {
			if !attestation.AggregationBits.BitAt(uint64(position)) {
				continue
			}

			// This was seen, so remove it from the missing attestations and add it to the completed ones
			delete(rpCommittee.Positions, position)
			if len(rpCommittee.Positions) == 0 {
				delete(slotInfo.Committees, attestation.CommitteeIndex)
			}
			if len(slotInfo.Committees) == 0 {
				delete(r.intervalDutiesInfo.Slots, attestation.SlotIndex)
			}
			delete(validator.MissingAttestationSlots, attestation.SlotIndex)

			// Check if this minipool was opted into the SP for this block
			nodeDetails := r.nodeDetails[validator.NodeIndex]
			if blockTime.Before(nodeDetails.OptInTime) || blockTime.After(nodeDetails.OptOutTime) {
				// Not opted in
				continue
			}

			eligibleBorrowedEth := nodeDetails.EligibleBorrowedEth
			_, percentOfBorrowedEth := r.networkState.GetStakedRplValueInEthAndPercentOfBorrowedEth(eligibleBorrowedEth, nodeDetails.RplStake)

			// Mark this duty as completed
			validator.CompletedAttestations[attestation.SlotIndex] = true

			// Get the pseudoscore for this attestation
			details := r.networkState.MinipoolDetailsByAddress[validator.Address]
			bond, fee := details.GetMinipoolBondAndNodeFee(blockTime)

			if r.rewardsFile.RulesetVersion >= 10 {
				fee = fees.GetMinipoolFeeWithBonus(bond, fee, percentOfBorrowedEth)
			}

			minipoolScore := big.NewInt(0).Sub(oneEth, fee) // 1 - fee
			minipoolScore.Mul(minipoolScore, bond)          // Multiply by bond
			minipoolScore.Div(minipoolScore, thirtyTwoEth)  // Divide by 32 to get the bond as a fraction of a total validator
			minipoolScore.Add(minipoolScore, fee)           // Total = fee + (bond/32)(1 - fee)

			// Add it to the minipool's score and the total score
			validator.AttestationScore.Add(&validator.AttestationScore.Int, minipoolScore)
			r.totalAttestationScore.Add(r.totalAttestationScore, minipoolScore)
			r.successfulAttestations++
		}
	}

	return nil

}

// Maps out the attestaion duties for the given epoch
func (r *treeGeneratorImpl_v9_v10) getDutiesForEpoch(committees beacon.Committees) error {

	// Crawl the committees
	for idx := 0; idx < committees.Count(); idx++ {
		slotIndex := committees.Slot(idx)
		if slotIndex < r.rewardsFile.ConsensusStartBlock || slotIndex > r.rewardsFile.ConsensusEndBlock {
			// Ignore slots that are out of bounds
			continue
		}
		blockTime := r.genesisTime.Add(time.Second * time.Duration(r.beaconConfig.SecondsPerSlot*slotIndex))
		committeeIndex := committees.Index(idx)

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*MinipoolInfo{}
		for position, validator := range committees.Validators(idx) {
			minipoolInfo, exists := r.validatorIndexMap[validator]
			if !exists {
				// This isn't an RP validator, so ignore it
				continue
			}

			// Check if this minipool was opted into the SP for this block
			nodeDetails := r.networkState.NodeDetailsByAddress[minipoolInfo.NodeAddress]
			isOptedIn := nodeDetails.SmoothingPoolRegistrationState
			spRegistrationTime := time.Unix(nodeDetails.SmoothingPoolRegistrationChanged.Int64(), 0)
			if (isOptedIn && blockTime.Sub(spRegistrationTime) < 0) || // If this block occurred before the node opted in, ignore it
				(!isOptedIn && spRegistrationTime.Sub(blockTime) < 0) { // If this block occurred after the node opted out, ignore it
				continue
			}

			// Check if this minipool was in the `staking` state during this time
			mpd := r.networkState.MinipoolDetailsByAddress[minipoolInfo.Address]
			statusChangeTime := time.Unix(mpd.StatusTime.Int64(), 0)
			if mpd.Status != rptypes.Staking || blockTime.Sub(statusChangeTime) < 0 {
				continue
			}

			// This was a legal RP validator opted into the SP during this slot so add it
			rpValidators[position] = minipoolInfo
			minipoolInfo.MissingAttestationSlots[slotIndex] = true
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
func (r *treeGeneratorImpl_v9_v10) createMinipoolIndexMap() error {

	// Get the status for all uncached minipool validators and add them to the cache
	r.validatorIndexMap = map[string]*MinipoolInfo{}
	for _, details := range r.nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				status, exists := r.networkState.MinipoolValidatorDetails[minipoolInfo.ValidatorPubkey]
				if !exists {
					// Remove minipools that don't have indices yet since they're not actually viable
					//r.log.Printlnf("NOTE: minipool %s (pubkey %s) didn't exist at this slot; removing it", minipoolInfo.Address.Hex(), minipoolInfo.ValidatorPubkey.Hex())
					minipoolInfo.WasActive = false
				} else {
					switch status.Status {
					case beacon.ValidatorState_PendingInitialized, beacon.ValidatorState_PendingQueued:
						// Remove minipools that don't have indices yet since they're not actually viable
						//r.log.Printlnf("NOTE: minipool %s (index %s, pubkey %s) was in state %s; removing it", minipoolInfo.Address.Hex(), status.Index, minipoolInfo.ValidatorPubkey.Hex(), string(status.Status))
						minipoolInfo.WasActive = false
					default:
						// Get the validator index
						minipoolInfo.ValidatorIndex = status.Index
						r.validatorIndexMap[minipoolInfo.ValidatorIndex] = minipoolInfo

						// Get the validator's activation start and end slots
						startSlot := status.ActivationEpoch * r.beaconConfig.SlotsPerEpoch
						endSlot := status.ExitEpoch * r.beaconConfig.SlotsPerEpoch

						// Verify this minipool has already started
						if status.ActivationEpoch == FarEpoch {
							//r.log.Printlnf("NOTE: minipool %s hasn't been scheduled for activation yet; removing it", minipoolInfo.Address.Hex())
							minipoolInfo.WasActive = false
							continue
						} else if startSlot > r.rewardsFile.ConsensusEndBlock {
							//r.log.Printlnf("NOTE: minipool %s activates on slot %d which is after interval end %d; removing it", minipoolInfo.Address.Hex(), startSlot, r.rewardsFile.ConsensusEndBlock)
							minipoolInfo.WasActive = false
						}

						// Check if the minipool exited before this interval
						if status.ExitEpoch != FarEpoch && endSlot < r.rewardsFile.ConsensusStartBlock {
							//r.log.Printlnf("NOTE: minipool %s exited on slot %d which was before interval start %d; removing it", minipoolInfo.Address.Hex(), endSlot, r.rewardsFile.ConsensusStartBlock)
							minipoolInfo.WasActive = false
							continue
						}
					}
				}
			}
		}
	}

	return nil

}

var farFutureTimestamp int64 = 1000000000000000000 // Far into the future
var farPastTimestamp int64 = 0

// Get the details for every node that was opted into the Smoothing Pool for at least some portion of this interval
func (r *treeGeneratorImpl_v9_v10) getSmoothingPoolNodeDetails() error {

	// For each NO, get their opt-in status and time of last change in batches
	r.log.Printlnf("%s Getting details of nodes for Smoothing Pool calculation...", r.logPrefix)
	nodeCount := uint64(len(r.networkState.NodeDetails))
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
				nativeNodeDetails := r.networkState.NodeDetails[iterationIndex]
				nodeDetails := &NodeSmoothingDetails{
					Address:          nativeNodeDetails.NodeAddress,
					Minipools:        []*MinipoolInfo{},
					SmoothingPoolEth: big.NewInt(0),
					BonusEth:         big.NewInt(0),
					RewardsNetwork:   nativeNodeDetails.RewardNetwork.Uint64(),
					RplStake:         nativeNodeDetails.RplStake,
				}

				nodeDetails.IsOptedIn = nativeNodeDetails.SmoothingPoolRegistrationState
				statusChangeTimeBig := nativeNodeDetails.SmoothingPoolRegistrationChanged
				statusChangeTime := time.Unix(statusChangeTimeBig.Int64(), 0)

				if nodeDetails.IsOptedIn {
					nodeDetails.OptInTime = statusChangeTime
					nodeDetails.OptOutTime = time.Unix(farFutureTimestamp, 0)
				} else {
					nodeDetails.OptOutTime = statusChangeTime
					nodeDetails.OptInTime = time.Unix(farPastTimestamp, 0)
				}

				// Get the details for each minipool in the node
				for _, mpd := range r.networkState.MinipoolDetailsByNode[nodeDetails.Address] {
					if mpd.Exists && mpd.Status == rptypes.Staking {
						nativeMinipoolDetails := r.networkState.MinipoolDetailsByAddress[mpd.MinipoolAddress]
						penaltyCount := nativeMinipoolDetails.PenaltyCount.Uint64()
						if penaltyCount >= 3 {
							// This node is a cheater
							nodeDetails.IsEligible = false
							nodeDetails.Minipools = []*MinipoolInfo{}
							r.nodeDetails[iterationIndex] = nodeDetails
							return nil
						}

						// This minipool is below the penalty count, so include it
						nodeDetails.Minipools = append(nodeDetails.Minipools, &MinipoolInfo{
							Address:         mpd.MinipoolAddress,
							ValidatorPubkey: mpd.Pubkey,
							NodeAddress:     nodeDetails.Address,
							NodeIndex:       iterationIndex,
							Fee:             nativeMinipoolDetails.NodeFee,
							//MissedAttestations:      0,
							//GoodAttestations:        0,
							MissingAttestationSlots: map[uint64]bool{},
							CompletedAttestations:   map[uint64]bool{},
							WasActive:               true,
							AttestationScore:        NewQuotedBigInt(0),
							NodeOperatorBond:        nativeMinipoolDetails.NodeDepositBalance,
						})
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

	// Populate the eligible borrowed ETH field for all nodes
	for _, nodeDetails := range r.nodeDetails {
		nnd := r.networkState.NodeDetailsByAddress[nodeDetails.Address]
		nodeDetails.EligibleBorrowedEth = r.networkState.GetEligibleBorrowedEth(nnd)
	}

	return nil

}

// Validates that the provided network is legal
func (r *treeGeneratorImpl_v9_v10) validateNetwork(network uint64) (bool, error) {
	valid, exists := r.validNetworkCache[network]
	if !exists {
		var err error
		valid, err = r.rp.GetNetworkEnabled(big.NewInt(int64(network)), r.opts)
		if err != nil {
			return false, err
		}
		r.validNetworkCache[network] = valid
	}

	return valid, nil
}

// Gets the start blocks for the given interval
func (r *treeGeneratorImpl_v9_v10) getBlocksAndTimesForInterval(previousIntervalEvent rewards.RewardsEvent) (*types.Header, error) {
	// Sanity check to confirm the BN can access the block from the previous interval
	_, exists, err := r.bc.GetBeaconBlock(previousIntervalEvent.ConsensusBlock.String())
	if err != nil {
		return nil, fmt.Errorf("error verifying block from previous interval: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("couldn't retrieve CL block from previous interval (slot %d); this likely means you checkpoint sync'd your Beacon Node and it has not backfilled to the previous interval yet so it cannot be used for tree generation", previousIntervalEvent.ConsensusBlock.Uint64())
	}

	previousEpoch := previousIntervalEvent.ConsensusBlock.Uint64() / r.beaconConfig.SlotsPerEpoch
	nextEpoch := previousEpoch + 1

	consensusStartSlot := nextEpoch * r.beaconConfig.SlotsPerEpoch
	startTime := r.beaconConfig.GetSlotTime(consensusStartSlot)
	endTime := r.beaconConfig.GetSlotTime(r.snapshotEnd.Slot)

	r.rewardsFile.StartTime = startTime
	r.minipoolPerformanceFile.StartTime = startTime

	r.rewardsFile.EndTime = endTime
	r.minipoolPerformanceFile.EndTime = endTime

	r.rewardsFile.ConsensusStartBlock = nextEpoch * r.beaconConfig.SlotsPerEpoch
	r.minipoolPerformanceFile.ConsensusStartBlock = r.rewardsFile.ConsensusStartBlock

	r.rewardsFile.ConsensusEndBlock = r.snapshotEnd.ConsensusBlock
	r.minipoolPerformanceFile.ConsensusEndBlock = r.snapshotEnd.ConsensusBlock

	r.rewardsFile.ExecutionEndBlock = r.snapshotEnd.ExecutionBlock
	r.minipoolPerformanceFile.ExecutionEndBlock = r.snapshotEnd.ExecutionBlock

	// Get the first block that isn't missing
	var elBlockNumber uint64
	for {
		beaconBlock, exists, err := r.bc.GetBeaconBlock(fmt.Sprint(r.rewardsFile.ConsensusStartBlock))
		if err != nil {
			return nil, fmt.Errorf("error getting EL data for BC slot %d: %w", r.rewardsFile.ConsensusStartBlock, err)
		}
		if !exists {
			r.rewardsFile.ConsensusStartBlock++
			r.minipoolPerformanceFile.ConsensusStartBlock++
		} else {
			elBlockNumber = beaconBlock.ExecutionBlockNumber
			break
		}
	}

	var startElHeader *types.Header
	if elBlockNumber == 0 {
		// We are pre-merge, so get the first block after the one from the previous interval
		r.rewardsFile.ExecutionStartBlock = previousIntervalEvent.ExecutionBlock.Uint64() + 1
		r.minipoolPerformanceFile.ExecutionStartBlock = r.rewardsFile.ExecutionStartBlock
		startElHeader, err = r.rp.HeaderByNumber(context.Background(), big.NewInt(int64(r.rewardsFile.ExecutionStartBlock)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL start block %d: %w", r.rewardsFile.ExecutionStartBlock, err)
		}
	} else {
		// We are post-merge, so get the EL block corresponding to the BC block
		r.rewardsFile.ExecutionStartBlock = elBlockNumber
		r.minipoolPerformanceFile.ExecutionStartBlock = r.rewardsFile.ExecutionStartBlock
		startElHeader, err = r.rp.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL header for block %d: %w", elBlockNumber, err)
		}
	}

	return startElHeader, nil
}

func (r *treeGeneratorImpl_v9_v10) saveFiles(smartnode *config.SmartnodeConfig, treeResult *GenerateTreeResult, nodeTrusted bool) (cid.Cid, map[string]cid.Cid, error) {
	return saveRewardsArtifacts(smartnode, treeResult, nodeTrusted)
}
