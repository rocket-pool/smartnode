package rewards

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipfs/go-cid"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rewards/fees"
	"github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types"
	sszbig "github.com/rocket-pool/smartnode/shared/services/rewards/ssz_types/big"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"golang.org/x/sync/errgroup"
)

// Implementation for tree generator ruleset v9
type treeGeneratorImpl_v11 struct {
	networkState                 *state.NetworkState
	rewardsFile                  *ssz_types.SSZFile_v2
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
	minipoolValidatorIndexMap    map[string]*MinipoolInfo
	megapoolValidatorIndexMap    map[string]*MegapoolInfo
	elStartTime                  time.Time
	elEndTime                    time.Time
	validNetworkCache            map[uint64]bool
	epsilon                      *big.Int
	intervalSeconds              *big.Int
	beaconConfig                 beacon.Eth2Config
	validatorStatusMap           map[rptypes.ValidatorPubkey]beacon.ValidatorStatus
	totalAttestationScore        *big.Int
	totalVoterScore              *big.Int
	totalPdaoScore               *big.Int
	successfulAttestations       uint64
	genesisTime                  time.Time
	invalidNetworkNodes          map[common.Address]uint64
	performanceFile              *PerformanceFile_v1
	nodeRewards                  map[common.Address]*ssz_types.NodeReward_v2
	networkRewards               map[ssz_types.Layer]*ssz_types.NetworkReward
	// Whether the interval is eligible for consensus bonuses
	isEligibleInterval bool

	// fields for RPIP-62 bonus calculations
	// Withdrawals made by a minipool's validator.
	minipoolWithdrawals map[common.Address]*big.Int
}

// Create a new tree generator
func newTreeGeneratorImpl_v11(log *log.ColorLogger, logPrefix string, index uint64, snapshotEnd *SnapshotEnd, elSnapshotHeader *types.Header, intervalsPassed uint64, state *state.NetworkState, isEligibleInterval bool) *treeGeneratorImpl_v11 {
	return &treeGeneratorImpl_v11{
		rewardsFile: &ssz_types.SSZFile_v2{
			RewardsFileVersion: 4,
			RulesetVersion:     11,
			Index:              index,
			IntervalsPassed:    intervalsPassed,
			TotalRewards: &ssz_types.TotalRewards_v2{
				ProtocolDaoRpl:               sszbig.NewUint256(0),
				TotalCollateralRpl:           sszbig.NewUint256(0),
				TotalOracleDaoRpl:            sszbig.NewUint256(0),
				TotalSmoothingPoolEth:        sszbig.NewUint256(0),
				PoolStakerSmoothingPoolEth:   sszbig.NewUint256(0),
				NodeOperatorSmoothingPoolEth: sszbig.NewUint256(0),
				TotalNodeWeight:              sszbig.NewUint256(0),
				TotalVoterShareEth:           sszbig.NewUint256(0),
				SmoothingPoolVoterShareEth:   sszbig.NewUint256(0),
				TotalPdaoShareEth:            sszbig.NewUint256(0),
			},
			NetworkRewards: ssz_types.NetworkRewards{},
			NodeRewards:    ssz_types.NodeRewards_v2{},
		},
		validatorStatusMap:        map[rptypes.ValidatorPubkey]beacon.ValidatorStatus{},
		minipoolValidatorIndexMap: map[string]*MinipoolInfo{},
		elSnapshotHeader:          elSnapshotHeader,
		snapshotEnd:               snapshotEnd,
		log:                       log,
		logPrefix:                 logPrefix,
		totalAttestationScore:     big.NewInt(0),
		totalVoterScore:           big.NewInt(0),
		totalPdaoScore:            big.NewInt(0),
		networkState:              state,
		invalidNetworkNodes:       map[common.Address]uint64{},
		performanceFile: &PerformanceFile_v1{
			Index:               index,
			MinipoolPerformance: map[common.Address]*MinipoolPerformance_v2{},
			MegapoolPerformance: map[common.Address]*MegapoolPerformance_v1{},
		},
		nodeRewards:         map[common.Address]*ssz_types.NodeReward_v2{},
		networkRewards:      map[ssz_types.Layer]*ssz_types.NetworkReward{},
		minipoolWithdrawals: map[common.Address]*big.Int{},
		isEligibleInterval:  isEligibleInterval,
	}
}

// Get the version of the ruleset used by this generator
func (r *treeGeneratorImpl_v11) getRulesetVersion() uint64 {
	return r.rewardsFile.RulesetVersion
}

func (r *treeGeneratorImpl_v11) generateTree(rp RewardsExecutionClient, networkName string, previousRewardsPoolAddresses []common.Address, bc RewardsBeaconClient) (*GenerateTreeResult, error) {

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
	r.performanceFile.Network = networkName
	r.performanceFile.RewardsFileVersion = r.rewardsFile.RewardsFileVersion
	r.performanceFile.RulesetVersion = r.rewardsFile.RulesetVersion

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
		if r.networkState.IsSaturnDeployed {
			// Add the number of megapool validators
			for _, nodeInfo := range r.nodeDetails {
				if nodeInfo.Megapool == nil {
					continue
				}
				r.epsilon.Add(r.epsilon, big.NewInt(int64(len(nodeInfo.Megapool.Validators))))
			}
		}
		// Cumulative error can exceed the validator count
		r.epsilon.Mul(r.epsilon, big.NewInt(2))
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
	for _, minipoolInfo := range r.performanceFile.MinipoolPerformance {
		slices.Sort(minipoolInfo.MissingAttestationSlots)
	}

	for _, megapoolInfo := range r.performanceFile.MegapoolPerformance {
		for _, validatorInfo := range megapoolInfo.ValidatorPerformance {
			slices.Sort(validatorInfo.MissingAttestationSlots)
		}
	}

	return &GenerateTreeResult{
		RulesetVersion:          r.rewardsFile.RulesetVersion,
		RewardsFile:             r.rewardsFile,
		InvalidNetworkNodes:     r.invalidNetworkNodes,
		MinipoolPerformanceFile: r.performanceFile,
	}, nil

}

// Quickly calculates an approximate of the staker's share of the smoothing pool balance without processing Beacon performance
// Used for approximate returns in the rETH ratio update
func (r *treeGeneratorImpl_v11) approximateStakerShareOfSmoothingPool(rp RewardsExecutionClient, networkName string, bc RewardsBeaconClient) (*big.Int, error) {
	r.log.Printlnf("%s Approximating tree using Ruleset v%d.", r.logPrefix, r.rewardsFile.RulesetVersion)

	r.rp = rp
	r.bc = bc
	r.validNetworkCache = map[uint64]bool{
		0: true,
	}

	// Set the network name
	r.rewardsFile.Network, _ = ssz_types.NetworkFromString(networkName)
	r.performanceFile.Network = networkName
	r.performanceFile.RewardsFileVersion = r.rewardsFile.RewardsFileVersion
	r.performanceFile.RulesetVersion = r.rewardsFile.RulesetVersion

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
		if r.networkState.IsSaturnDeployed {
			// Add the number of megapool validators
			for _, nodeInfo := range r.nodeDetails {
				if nodeInfo.Megapool == nil {
					continue
				}
				r.epsilon.Add(r.epsilon, big.NewInt(int64(nodeInfo.Megapool.ActiveValidatorCount)))
			}
		}
	}
	// Cumulative error can exceed the validator count
	r.epsilon.Mul(r.epsilon, big.NewInt(2))

	// Calculate the ETH rewards
	err := r.calculateEthRewards(false)
	if err != nil {
		return nil, fmt.Errorf("error calculating ETH rewards: %w", err)
	}

	return r.rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Int, nil
}

func (r *treeGeneratorImpl_v11) calculateNodeRplRewards(
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
func (r *treeGeneratorImpl_v11) calculateRplRewards() error {
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

					rewardsForNode = ssz_types.NewNodeReward_v2(
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

			rewardsForNode = ssz_types.NewNodeReward_v2(
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
	if totalNodeWeight.Sign() > 0 {
		// Subtract oDAO rewards only in case node operators had rewards, otherwise pDaoRewards is already correct
		pDaoRewards.Sub(pDaoRewards, totalCalculatedOdaoRewards)
	}
	r.rewardsFile.TotalRewards.ProtocolDaoRpl = sszbig.NewUint256(0)
	r.rewardsFile.TotalRewards.ProtocolDaoRpl.Set(pDaoRewards)
	r.log.Printlnf("%s Actual Protocol DAO rewards:  %s to account for truncation", r.logPrefix, pDaoRewards.String())

	// Print total node weight
	r.log.Printlnf("%s Total Node Weight:            %s", r.logPrefix, totalNodeWeight)

	return nil

}

// Calculates the ETH rewards for the given interval
func (r *treeGeneratorImpl_v11) calculateEthRewards(checkBeaconPerformance bool) error {

	// Get the Smoothing Pool contract's balance
	r.smoothingPoolBalance = r.networkState.NetworkDetails.SmoothingPoolBalance
	r.log.Printlnf("%s Smoothing Pool Balance:\t%s\t(%.3f)", r.logPrefix, r.smoothingPoolBalance.String(), eth.WeiToEth(r.smoothingPoolBalance))
	r.log.Printlnf("%s  Earmarked Voter Share:\t%s\t(%.3f)", r.logPrefix, r.networkState.NetworkDetails.SmoothingPoolPendingVoterShare.String(), eth.WeiToEth(r.networkState.NetworkDetails.SmoothingPoolPendingVoterShare))

	if r.rewardsFile.Index == 0 {
		// This is the first interval, Smoothing Pool rewards are ignored on the first interval since it doesn't have a discrete start time
		return nil
	}

	// Get the start time of this interval based on the event from the previous one
	// This must be done even if there are no smoothing pool rewards, to set the start blocks and timestamps
	//previousIntervalEvent, err := GetRewardSnapshotEvent(r.rp, r.cfg, r.rewardsFile.Index-1, r.opts) // This is immutable so querying at the head is fine and mitigates issues around calls for pruned EL state
	previousIntervalEvent, err := r.rp.GetRewardSnapshotEvent(r.previousRewardsPoolAddresses, r.rewardsFile.Index-1, r.opts)
	if err != nil {
		return err
	}
	startElBlockHeader, err := r.getBlocksAndTimesForInterval(previousIntervalEvent)
	if err != nil {
		return err
	}

	// Ignore the ETH calculation if there are no rewards
	if r.smoothingPoolBalance.Cmp(common.Big0) == 0 {
		return nil
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
	megapools := 0
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.IsEligible {
			eligible++
			if nodeInfo.Megapool != nil {
				megapools++
			}
		}
	}
	r.log.Printlnf("%s %d / %d nodes (%d with megapools) were eligible for Smoothing Pool rewards", r.logPrefix, eligible, len(r.nodeDetails), megapools)

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
				_, percentOfBorrowedEth := r.networkState.GetStakedRplValueInEthAndPercentOfBorrowedEth(nodeInfo.MinipoolEligibleBorrowedEth, nodeInfo.LegacyStakedRpl)
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

				// Repeat, for megapools
				if nodeInfo.Megapool != nil {
					megapool := nodeInfo.Megapool
					for _, validator := range megapool.Validators {
						details := r.networkState.MegapoolDetails[megapool.Address]
						bond := details.GetMegapoolBondNormalized()
						nodeFee := r.networkState.NetworkDetails.MegapoolRevenueSplitTimeWeightedAverages.NodeShare
						voterFee := r.networkState.NetworkDetails.MegapoolRevenueSplitTimeWeightedAverages.VoterShare
						pdaoFee := r.networkState.NetworkDetails.MegapoolRevenueSplitTimeWeightedAverages.PdaoShare

						// The megapool score is given by:
						// (bond + effectiveNodeFee*(32-bond)) / 32
						// However, when multiplying eth values, we need to normalize the wei to eth
						// So really it's (bond + (32*fee / 1E) - (32*bond / 1E)) / 32
						// If we multiply the numerator by 1 eth each, we can avoid some
						// integer math inaccuracy, and when we divide by 32 it is removed.
						//
						// (b*1 + 32f - f*b) / 32
						megapoolScore := big.NewInt(0).Mul(oneEth, bond)                           // b*1
						megapoolScore.Add(megapoolScore, big.NewInt(0).Mul(thirtyTwoEth, nodeFee)) // b*1 + 32f
						megapoolScore.Sub(megapoolScore, big.NewInt(0).Mul(nodeFee, bond))         // b*1 + 32f - f*b
						megapoolScore.Div(megapoolScore, thirtyTwoEth)                             // (b*1 + 32f - f*b) / 32

						// Add it to the megapool's score and the total score
						validator.AttestationScore.Add(&validator.AttestationScore.Int, megapoolScore)
						r.totalAttestationScore.Add(r.totalAttestationScore, megapoolScore)

						// Calculate the voter share
						// This is simply (effectiveVoterFee * (32 - bond)) / 32
						// Simplify to (32f - f*b) / 32
						voterScore := big.NewInt(0).Mul(thirtyTwoEth, voterFee)
						voterScore.Sub(voterScore, big.NewInt(0).Mul(voterFee, bond))
						voterScore.Div(voterScore, thirtyTwoEth)
						r.totalVoterScore.Add(r.totalVoterScore, voterScore)

						// Calculate the pdao share
						// Same formula as the voter share
						pdaoScore := big.NewInt(0).Mul(thirtyTwoEth, pdaoFee)
						pdaoScore.Sub(pdaoScore, big.NewInt(0).Mul(pdaoFee, bond))
						pdaoScore.Div(pdaoScore, thirtyTwoEth)
						r.totalPdaoScore.Add(r.totalPdaoScore, pdaoScore)
						r.successfulAttestations++
					}
				}
			}
		}
	}

	// Determine how much ETH each node gets and how much the pool stakers get
	nodeRewards, err := r.calculateNodeRewards()
	if err != nil {
		return err
	}
	if r.rewardsFile.RulesetVersion >= 10 {
		r.performanceFile.BonusScalar = QuotedBigIntFromBigInt(nodeRewards.bonusScalar)
	}

	// Update the rewards maps
	for _, nodeInfo := range r.nodeDetails {
		// First, take care of voter share
		if nodeInfo.VoterShareEth.Cmp(common.Big0) > 0 {
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

				rewardsForNode = ssz_types.NewNodeReward_v2(
					network,
					ssz_types.AddressFromBytes(nodeInfo.Address.Bytes()),
				)
				r.nodeRewards[nodeInfo.Address] = rewardsForNode
			}
			rewardsForNode.VoterShareEth.Add(rewardsForNode.VoterShareEth.Int, nodeInfo.VoterShareEth)

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.networkRewards[rewardsForNode.Network]
			if !exists {
				rewardsForNetwork = ssz_types.NewNetworkReward(rewardsForNode.Network)
				r.networkRewards[rewardsForNode.Network] = rewardsForNetwork
			}
			rewardsForNetwork.SmoothingPoolEth.Add(rewardsForNetwork.SmoothingPoolEth.Int, nodeInfo.VoterShareEth)
		}

		// Next, take care of smoothing pool ETH
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

				rewardsForNode = ssz_types.NewNodeReward_v2(
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
				performance := &MinipoolPerformance_v2{
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
				r.performanceFile.MinipoolPerformance[minipoolInfo.Address] = performance
			}

			// Add megapool rewards to the JSON
			if nodeInfo.Megapool != nil {
				for _, validator := range nodeInfo.Megapool.Validators {
					successfulAttestations := uint64(len(validator.CompletedAttestations))
					missingAttestations := uint64(len(validator.MissingAttestationSlots))
					performance := &MegapoolValidatorPerformance_v1{
						pubkey:                  validator.Pubkey.Hex(),
						SuccessfulAttestations:  successfulAttestations,
						MissedAttestations:      missingAttestations,
						AttestationScore:        validator.AttestationScore,
						EthEarned:               QuotedBigIntFromBigInt(validator.MegapoolValidatorShare),
						MissingAttestationSlots: []uint64{},
					}
					if successfulAttestations+missingAttestations == 0 {
						// Don't include megapools that have zero attestations
						continue
					}
					for slot := range validator.MissingAttestationSlots {
						performance.MissingAttestationSlots = append(performance.MissingAttestationSlots, slot)
					}
					mpPerformance, exists := r.performanceFile.MegapoolPerformance[nodeInfo.Megapool.Address]
					if !exists {
						mpPerformance = &MegapoolPerformance_v1{
							ValidatorPerformance: make(MegapoolPerformanceMap),
						}
						r.performanceFile.MegapoolPerformance[nodeInfo.Megapool.Address] = mpPerformance
					}
					mpPerformance.ValidatorPerformance[validator.Pubkey] = performance
				}
			}

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.networkRewards[rewardsForNode.Network]
			if !exists {
				rewardsForNetwork = ssz_types.NewNetworkReward(rewardsForNode.Network)
				r.networkRewards[rewardsForNode.Network] = rewardsForNetwork
			}
			rewardsForNetwork.SmoothingPoolEth.Add(rewardsForNetwork.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
		}

		// Finally, take care of adding voter share to the performance file
		if nodeInfo.VoterShareEth.Cmp(common.Big0) > 0 {
			performance, exists := r.performanceFile.MegapoolPerformance[nodeInfo.Megapool.Address]
			if !exists {
				performance = &MegapoolPerformance_v1{
					VoterShare: QuotedBigIntFromBigInt(nodeInfo.VoterShareEth),
				}
				r.performanceFile.MegapoolPerformance[nodeInfo.Megapool.Address] = performance
			}
		}
	}

	// Set the totals
	r.rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Set(nodeRewards.poolStakerEth)
	r.rewardsFile.TotalRewards.NodeOperatorSmoothingPoolEth.Set(nodeRewards.nodeOpEth)
	r.rewardsFile.TotalRewards.TotalSmoothingPoolEth.Set(r.smoothingPoolBalance)
	r.rewardsFile.TotalRewards.TotalVoterShareEth.Set(nodeRewards.voterEth)
	r.rewardsFile.TotalRewards.TotalPdaoShareEth.Set(nodeRewards.pdaoEth)
	return nil

}

func (r *treeGeneratorImpl_v11) calculateNodeBonuses() (*big.Int, error) {
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
		_, percentOfBorrowedEth := r.networkState.GetStakedRplValueInEthAndPercentOfBorrowedEth(nsd.MinipoolEligibleBorrowedEth, nsd.LegacyStakedRpl)
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

type nodeRewards struct {
	poolStakerEth *big.Int
	nodeOpEth     *big.Int
	pdaoEth       *big.Int
	voterEth      *big.Int
	bonusScalar   *big.Int
}

// Calculate the distribution of Smoothing Pool ETH to each node
func (r *treeGeneratorImpl_v11) calculateNodeRewards() (*nodeRewards, error) {
	var err error
	bonusScalar := big.NewInt(0).Set(oneEth)

	voterEth := big.NewInt(0)
	pdaoEth := big.NewInt(0)

	// If pdao score is greater than 0, calculate the pdao share
	if r.totalPdaoScore.Cmp(common.Big0) > 0 {
		pdaoEth.Mul(r.smoothingPoolBalance, r.totalPdaoScore)
		pdaoEth.Div(pdaoEth, big.NewInt(int64(r.successfulAttestations)))
		pdaoEth.Div(pdaoEth, oneEth)
	}

	// If voter score is greater than 0, calculate the voter share
	if r.totalVoterScore.Cmp(common.Big0) > 0 {
		voterEth.Mul(r.smoothingPoolBalance, r.totalVoterScore)
		voterEth.Div(voterEth, big.NewInt(int64(r.successfulAttestations)))
		voterEth.Div(voterEth, oneEth)

		// Set the voter share eth in the rewards file
		r.rewardsFile.TotalRewards.SmoothingPoolVoterShareEth.Set(voterEth)

		// Add in the earmarked voter share
		voterEth.Add(voterEth, r.networkState.NetworkDetails.SmoothingPoolPendingVoterShare)
	}

	totalMegapoolVoteEligibleRpl := big.NewInt(0)
	for _, nodeInfo := range r.nodeDetails {
		// Check if the node is eligible for voter share
		if nodeInfo.Megapool == nil {
			continue
		}
		totalMegapoolVoteEligibleRpl.Add(totalMegapoolVoteEligibleRpl, nodeInfo.MegapoolVoteEligibleRpl)
	}
	// Calculate the voter share for each node
	trueVoterEth := big.NewInt(0)
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.Megapool == nil {
			continue
		}
		if nodeInfo.MegapoolVoteEligibleRpl.Cmp(common.Big0) == 0 {
			continue
		}

		// The node's voter share is nodeRpl*voterEth/totalMegapoolVoteEligibleRpl
		nodeInfo.VoterShareEth.Set(nodeInfo.MegapoolVoteEligibleRpl)
		nodeInfo.VoterShareEth.Mul(nodeInfo.VoterShareEth, voterEth)
		nodeInfo.VoterShareEth.Div(nodeInfo.VoterShareEth, totalMegapoolVoteEligibleRpl)
		trueVoterEth.Add(trueVoterEth, nodeInfo.VoterShareEth)
	}

	// If there weren't any successful attestations, everything goes to the pool stakers
	if r.totalAttestationScore.Cmp(common.Big0) == 0 || r.successfulAttestations == 0 {
		r.log.Printlnf("WARNING: Total attestation score = %s, successful attestations = %d... sending the whole smoothing pool balance to the pool stakers.", r.totalAttestationScore.String(), r.successfulAttestations)
		poolStakerEth := big.NewInt(0).Set(r.smoothingPoolBalance)
		poolStakerEth.Sub(poolStakerEth, trueVoterEth)
		poolStakerEth.Sub(poolStakerEth, pdaoEth)
		return &nodeRewards{
			poolStakerEth: poolStakerEth,
			nodeOpEth:     big.NewInt(0),
			pdaoEth:       pdaoEth,
			voterEth:      trueVoterEth,
			bonusScalar:   bonusScalar,
		}, nil
	}

	// Calculate the minipool bonuses
	var totalConsensusBonus *big.Int
	if r.rewardsFile.RulesetVersion >= 10 && r.isEligibleInterval {
		totalConsensusBonus, err = r.calculateNodeBonuses()
		if err != nil {
			return nil, err
		}
	}

	totalEthForMinipools := big.NewInt(0)
	totalEthForMegapools := big.NewInt(0)
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

		// Check megapool eth as well
		if nodeInfo.Megapool != nil {
			for _, validator := range nodeInfo.Megapool.Validators {
				validatorEth := big.NewInt(0).Set(totalNodeOpShare)
				validatorEth.Mul(validatorEth, &validator.AttestationScore.Int)
				validatorEth.Div(validatorEth, r.totalAttestationScore)
				validator.MegapoolValidatorShare = validatorEth
				nodeInfo.SmoothingPoolEth.Add(nodeInfo.SmoothingPoolEth, validatorEth)

				totalEthForMegapools.Add(totalEthForMegapools, validatorEth)
			}
		}
	}

	if r.rewardsFile.RulesetVersion >= 10 {
		remainingBalance := big.NewInt(0).Sub(r.smoothingPoolBalance, totalEthForMinipools)
		remainingBalance.Sub(remainingBalance, totalEthForMegapools)
		remainingBalance.Sub(remainingBalance, pdaoEth)
		if trueVoterEth.Sign() > 0 {
			remainingBalance.Sub(remainingBalance, trueVoterEth)
		} else {
			// Nobody earned voter share.
			// Subtract voter share- it shouldn't be used to pay bonuses, or we could have a deficit later.
			remainingBalance.Sub(remainingBalance, r.networkState.NetworkDetails.SmoothingPoolPendingVoterShare)
		}
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
		} else {
			r.log.Printlnf("%s Smoothing Pool has %s (%.3f) Pool Staker ETH before bonuses which is enough for %s (%.3f) in bonuses.", r.logPrefix, remainingBalance.String(), eth.WeiToEth(remainingBalance), totalConsensusBonus.String(), eth.WeiToEth(totalConsensusBonus))
		}
	}

	// Finally, award the bonuses
	totalEthForBonuses := big.NewInt(0)
	if r.rewardsFile.RulesetVersion >= 10 {
		for _, nsd := range r.nodeDetails {
			nsd.SmoothingPoolEth.Add(nsd.SmoothingPoolEth, nsd.BonusEth)
			totalEthForBonuses.Add(totalEthForBonuses, nsd.BonusEth)
		}
	}

	trueNodeOperatorAmount := big.NewInt(0)
	trueNodeOperatorAmount.Add(trueNodeOperatorAmount, totalEthForMinipools)
	trueNodeOperatorAmount.Add(trueNodeOperatorAmount, totalEthForMegapools)

	delta := big.NewInt(0).Sub(trueNodeOperatorAmount, totalNodeOpShare)
	delta.Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return nil, fmt.Errorf("error calculating smoothing pool ETH: total was %s, but expected %s; error was too large (%s wei)", trueNodeOperatorAmount.String(), totalNodeOpShare.String(), delta.String())
	}

	trueNodeOperatorAmount.Add(trueNodeOperatorAmount, totalEthForBonuses)

	// This is how much actually goes to the pool stakers - it should ideally be equal to poolStakerShare but this accounts for any cumulative floating point errors
	truePoolStakerAmount := big.NewInt(0).Sub(r.smoothingPoolBalance, trueNodeOperatorAmount)
	truePoolStakerAmount.Sub(truePoolStakerAmount, pdaoEth)
	truePoolStakerAmount.Sub(truePoolStakerAmount, trueVoterEth)

	r.log.Printlnf("%s Smoothing Pool ETH:               \t%s\t(%.3f)", r.logPrefix, r.smoothingPoolBalance.String(), eth.WeiToEth(r.smoothingPoolBalance))
	r.log.Printlnf("%s Pool staker ETH:                  \t%s\t(%.3f)", r.logPrefix, truePoolStakerAmount.String(), eth.WeiToEth(truePoolStakerAmount))
	r.log.Printlnf("%s Node Op Eth:                      \t%s\t(%.3f)", r.logPrefix, trueNodeOperatorAmount.String(), eth.WeiToEth(trueNodeOperatorAmount))
	r.log.Printlnf("%s        '--> minipool attestations:\t%s\t(%.3f)", r.logPrefix, totalEthForMinipools.String(), eth.WeiToEth(totalEthForMinipools))
	r.log.Printlnf("%s        '----------------> bonuses:\t%s\t(%.3f)", r.logPrefix, totalEthForBonuses.String(), eth.WeiToEth(totalEthForBonuses))
	r.log.Printlnf("%s        '--> megapool attestations:\t%s\t(%.3f)", r.logPrefix, totalEthForMegapools.String(), eth.WeiToEth(totalEthForMegapools))
	r.log.Printlnf("%s Voter Share:                      \t%s\t(%.3f)", r.logPrefix, trueVoterEth.String(), eth.WeiToEth(trueVoterEth))
	r.log.Printlnf("%s PDAO ETH:                         \t%s\t(%.3f)", r.logPrefix, pdaoEth.String(), eth.WeiToEth(pdaoEth))
	// Sum the actual values to determine how much eth is distributed
	toBeDistributed := big.NewInt(0)
	toBeDistributed.Add(toBeDistributed, truePoolStakerAmount)
	toBeDistributed.Add(toBeDistributed, trueNodeOperatorAmount)
	toBeDistributed.Add(toBeDistributed, trueVoterEth)
	toBeDistributed.Add(toBeDistributed, pdaoEth)
	r.log.Printlnf("%s TOTAL to be distributed:          \t%s\t(%.3f)", r.logPrefix, toBeDistributed.String(), eth.WeiToEth(toBeDistributed))
	r.log.Printlnf("%s (error = %s wei)", r.logPrefix, delta.String())

	return &nodeRewards{
		poolStakerEth: truePoolStakerAmount,
		nodeOpEth:     trueNodeOperatorAmount,
		bonusScalar:   bonusScalar,
		pdaoEth:       pdaoEth,
		voterEth:      trueVoterEth,
	}, nil

}

// Get all of the duties for a range of epochs
func (r *treeGeneratorImpl_v11) processAttestationsBalancesAndWithdrawalsForInterval() error {

	startEpoch := r.rewardsFile.ConsensusStartBlock / r.beaconConfig.SlotsPerEpoch
	endEpoch := r.rewardsFile.ConsensusEndBlock / r.beaconConfig.SlotsPerEpoch

	// Determine the validator indices of each minipool
	err := r.createMinipoolIndexMap()
	if err != nil {
		return err
	}

	err = r.createMegapoolIndexMap()
	if err != nil {
		return err
	}

	// Check all of the attestations for each epoch
	r.log.Printlnf("%s Checking participation of %d minipools and %d megapool validators for epochs %d to %d", r.logPrefix, len(r.minipoolValidatorIndexMap), len(r.megapoolValidatorIndexMap), startEpoch, endEpoch)
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
func (r *treeGeneratorImpl_v11) processEpoch(duringInterval bool, epoch uint64) error {

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
				mpi, exists := r.minipoolValidatorIndexMap[withdrawal.ValidatorIndex]
				if !exists {
					continue
				}
				nnd := r.networkState.NodeDetailsByAddress[mpi.Node.Address]
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
		return fmt.Errorf("error getting committee and attestation records for epoch %d: %w", epoch, err)
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
		attestations := attestationsPerSlot[i]
		if len(attestations) > 0 {
			r.checkAttestations(attestations)
		}
	}

	return nil

}

func (r *treeGeneratorImpl_v11) checkAttestations(attestations []beacon.AttestationInfo) error {

	// Go through the attestations for the block
	for _, attestation := range attestations {
		// Get the RP committees for this attestation's slot and index
		slotInfo, exists := r.intervalDutiesInfo.Slots[attestation.SlotIndex]
		if !exists {
			continue
		}

		for _, committeeIndex := range attestation.CommitteeIndices() {
			rpCommittee, exists := slotInfo.Committees[uint64(committeeIndex)]
			if !exists {
				continue
			}
			blockTime := r.genesisTime.Add(time.Second * time.Duration(r.networkState.BeaconConfig.SecondsPerSlot*attestation.SlotIndex))

			// Check if each RP validator attested successfully
			for position, positionInfo := range rpCommittee.Positions {
				if !attestation.ValidatorAttested(committeeIndex, position, slotInfo.CommitteeSizes) {
					continue
				}

				// This was seen, so remove it from the missing attestations and add it to the completed ones
				delete(rpCommittee.Positions, position)
				if len(rpCommittee.Positions) == 0 {
					delete(slotInfo.Committees, uint64(committeeIndex))
				}
				if len(slotInfo.Committees) == 0 {
					delete(r.intervalDutiesInfo.Slots, attestation.SlotIndex)
				}
				positionInfo.DeleteMissingAttestationSlot(attestation.SlotIndex)

				// Check if this minipool was opted into the SP for this block
				nodeDetails := positionInfo.GetNodeDetails()
				if blockTime.Before(nodeDetails.OptInTime) || blockTime.After(nodeDetails.OptOutTime) {
					// Not opted in
					continue
				}

				// Mark this duty as completed
				positionInfo.MarkAttestationCompleted(attestation.SlotIndex)

				if positionInfo.MinipoolInfo != nil {
					_, percentOfBorrowedEth := r.networkState.GetStakedRplValueInEthAndPercentOfBorrowedEth(nodeDetails.MinipoolEligibleBorrowedEth, nodeDetails.LegacyStakedRpl)

					validator := positionInfo.MinipoolInfo

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
					continue
				}

				megapool := positionInfo.Megapool
				validator := megapool.GetValidator()

				// Get the pseudoscore for this attestation
				details := r.networkState.MegapoolDetails[megapool.Info.Address]
				bond := details.GetMegapoolBondNormalized()
				nodeFee := r.networkState.NetworkDetails.MegapoolRevenueSplitTimeWeightedAverages.NodeShare
				voterFee := r.networkState.NetworkDetails.MegapoolRevenueSplitTimeWeightedAverages.VoterShare
				pdaoFee := r.networkState.NetworkDetails.MegapoolRevenueSplitTimeWeightedAverages.PdaoShare

				// The megapool score is given by:
				// (bond + effectiveNodeFee*(32-bond)) / 32
				// However, when multiplying eth values, we need to normalize the wei to eth
				// So really it's (bond + (32*fee / 1E) - (32*bond / 1E)) / 32
				// If we multiply the numerator by 1 eth each, we can avoid some
				// integer math inaccuracy, and when we divide by 32 it is removed.
				//
				// (b*1 + 32f - f*b) / 32
				megapoolScore := big.NewInt(0).Mul(oneEth, bond)                           // b*1
				megapoolScore.Add(megapoolScore, big.NewInt(0).Mul(thirtyTwoEth, nodeFee)) // b*1 + 32f
				megapoolScore.Sub(megapoolScore, big.NewInt(0).Mul(nodeFee, bond))         // b*1 + 32f - f*b
				megapoolScore.Div(megapoolScore, thirtyTwoEth)                             // (b*1 + 32f - f*b) / 32

				// Add it to the megapool's score and the total score
				validator.AttestationScore.Add(&validator.AttestationScore.Int, megapoolScore)
				r.totalAttestationScore.Add(r.totalAttestationScore, megapoolScore)

				// Calculate the voter share
				// This is simply (effectiveVoterFee * (32 - bond)) / 32
				// Simplify to (32f - f*b) / 32
				voterScore := big.NewInt(0).Mul(thirtyTwoEth, voterFee)
				voterScore.Sub(voterScore, big.NewInt(0).Mul(voterFee, bond))
				voterScore.Div(voterScore, thirtyTwoEth)
				r.totalVoterScore.Add(r.totalVoterScore, voterScore)

				// Calculate the pdao share
				// Same formula as the voter share
				pdaoScore := big.NewInt(0).Mul(thirtyTwoEth, pdaoFee)
				pdaoScore.Sub(pdaoScore, big.NewInt(0).Mul(pdaoFee, bond))
				pdaoScore.Div(pdaoScore, thirtyTwoEth)
				r.totalPdaoScore.Add(r.totalPdaoScore, pdaoScore)
				r.successfulAttestations++
			}
		}
	}

	return nil

}

// Maps out the attestation duties for the given epoch
func (r *treeGeneratorImpl_v11) getDutiesForEpoch(committees beacon.Committees) error {

	// Crawl the committees
	for idx := 0; idx < committees.Count(); idx++ {
		slotIndex := committees.Slot(idx)
		if slotIndex < r.rewardsFile.ConsensusStartBlock || slotIndex > r.rewardsFile.ConsensusEndBlock {
			// Ignore slots that are out of bounds
			continue
		}
		blockTime := r.genesisTime.Add(time.Second * time.Duration(r.beaconConfig.SecondsPerSlot*slotIndex))
		committeeIndex := committees.Index(idx)

		// Add the committee size to the list, for calculating offset in post-electra aggregation_bits
		slotInfo, exists := r.intervalDutiesInfo.Slots[slotIndex]
		if !exists {
			slotInfo = &SlotInfo{
				Index:          slotIndex,
				Committees:     map[uint64]*CommitteeInfo{},
				CommitteeSizes: map[uint64]int{},
			}
			r.intervalDutiesInfo.Slots[slotIndex] = slotInfo
		}
		slotInfo.CommitteeSizes[committeeIndex] = committees.ValidatorCount(idx)

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*PositionInfo{}
		for position, validator := range committees.Validators(idx) {
			minipoolInfo, miniExists := r.minipoolValidatorIndexMap[validator]
			megapoolInfo, megaExists := r.megapoolValidatorIndexMap[validator]
			if !miniExists && !megaExists {
				// This isn't an RP validator, so ignore it
				continue
			}

			// Check if this validator was opted into the SP for this block
			var nodeDetails *rpstate.NativeNodeDetails
			if miniExists {
				nodeDetails = r.networkState.NodeDetailsByAddress[minipoolInfo.Node.Address]
			} else {
				nodeDetails = r.networkState.NodeDetailsByAddress[megapoolInfo.Node.Address]
			}

			isOptedIn := nodeDetails.SmoothingPoolRegistrationState
			spRegistrationTime := time.Unix(nodeDetails.SmoothingPoolRegistrationChanged.Int64(), 0)
			if (isOptedIn && blockTime.Sub(spRegistrationTime) < 0) || // If this block occurred before the node opted in, ignore it
				(!isOptedIn && spRegistrationTime.Sub(blockTime) < 0) { // If this block occurred after the node opted out, ignore it
				continue
			}

			// Check if this validator was in the `staking` state during this time
			if miniExists {
				mpd := r.networkState.MinipoolDetailsByAddress[minipoolInfo.Address]
				statusChangeTime := time.Unix(mpd.StatusTime.Int64(), 0)
				if mpd.Status != rptypes.Staking || blockTime.Sub(statusChangeTime) < 0 {
					continue
				}

				// This was a legal RP validator opted into the SP during this slot so add it
				rpValidators[position] = &PositionInfo{
					MinipoolInfo: minipoolInfo,
				}
				minipoolInfo.MissingAttestationSlots[slotIndex] = true
				continue
			}
			megapoolInfo, ok := r.megapoolValidatorIndexMap[validator]
			if !ok {
				return fmt.Errorf("megapool not found indexed by validator %s", validator)
			}

			validatorInfo, exists := megapoolInfo.ValidatorIndexMap[validator]
			if !exists {
				return fmt.Errorf("validator %s not found indexed in megapool %s", validator, megapoolInfo.Address.Hex())
			}

			if !validatorInfo.NativeValidatorInfo.ValidatorInfo.Staked {
				continue
			}

			rpValidators[position] = &PositionInfo{
				Megapool: &MegapoolPositionInfo{
					Info:           megapoolInfo,
					ValidatorIndex: validator,
				},
			}
			validatorInfo.MissingAttestationSlots[slotIndex] = true
		}

		// If there are some RP validators, add this committee to the map
		if len(rpValidators) > 0 {
			slotInfo.Committees[committeeIndex] = &CommitteeInfo{
				Index:     committeeIndex,
				Positions: rpValidators,
			}
		}
	}

	return nil

}

// Maps all megapools to their validator indices and creates map of validator indices to megapool info
func (r *treeGeneratorImpl_v11) createMegapoolIndexMap() error {

	// Get the status for all uncached megapool validators and add them to the cache
	r.megapoolValidatorIndexMap = map[string]*MegapoolInfo{}
	for _, details := range r.nodeDetails {
		if !details.IsEligible {
			continue
		}
		if details.Megapool == nil {
			continue
		}
		for _, validatorInfo := range details.Megapool.Validators {
			status, exists := r.networkState.MegapoolValidatorDetails[validatorInfo.Pubkey]
			if !exists {
				validatorInfo.WasActive = false
				continue
			}

			switch status.Status {

			case beacon.ValidatorState_PendingInitialized, beacon.ValidatorState_PendingQueued:
				// Remove megapool validators that don't have indices yet since they're not actually viable
				//r.log.Printlnf("NOTE: megapool %s (index %s, pubkey %s) was in state %s; removing it", megapoolInfo.Address.Hex(), status.Index, validatorInfo.Pubkey.Hex(), string(status.Status))
				validatorInfo.WasActive = false
			default:
				// Get the validator index
				validatorInfo.Index = status.Index
				r.megapoolValidatorIndexMap[validatorInfo.Index] = details.Megapool

				// Get the validator's activation start and end slots

				// Get the validator's activation start and end slots
				startSlot := status.ActivationEpoch * r.beaconConfig.SlotsPerEpoch
				endSlot := status.ExitEpoch * r.beaconConfig.SlotsPerEpoch

				// Verify this megapool has already started
				if status.ActivationEpoch == FarEpoch {
					//r.log.Printlnf("NOTE: megapool %s hasn't been scheduled for activation yet; removing it", megapoolInfo.Address.Hex())
					validatorInfo.WasActive = false
				} else if startSlot > r.rewardsFile.ConsensusEndBlock {
					//r.log.Printlnf("NOTE: megapool %s activates on slot %d which is after interval end %d; removing it", megapoolInfo.Address.Hex(), startSlot, r.rewardsFile.ConsensusEndBlock)
					validatorInfo.WasActive = false
				}

				// Check if the megapool exited before this interval
				if status.ExitEpoch != FarEpoch && endSlot < r.rewardsFile.ConsensusStartBlock {
					//r.log.Printlnf("NOTE: megapool %s exited on slot %d which was before interval start %d; removing it", megapoolInfo.Address.Hex(), endSlot, r.rewardsFile.ConsensusStartBlock)
					validatorInfo.WasActive = false
				}
			}
		}
	}

	return nil
}

// Maps all minipools to their validator indices and creates a map of indices to minipool info
func (r *treeGeneratorImpl_v11) createMinipoolIndexMap() error {

	// Get the status for all uncached minipool validators and add them to the cache
	r.minipoolValidatorIndexMap = map[string]*MinipoolInfo{}
	for _, details := range r.nodeDetails {
		if !details.IsEligible {
			continue
		}
		for _, minipoolInfo := range details.Minipools {
			status, exists := r.networkState.MinipoolValidatorDetails[minipoolInfo.ValidatorPubkey]
			if !exists {
				// Remove minipools that don't have indices yet since they're not actually viable
				//r.log.Printlnf("NOTE: minipool %s (pubkey %s) didn't exist at this slot; removing it", minipoolInfo.Address.Hex(), minipoolInfo.ValidatorPubkey.Hex())
				minipoolInfo.WasActive = false
				continue
			}

			switch status.Status {
			case beacon.ValidatorState_PendingInitialized, beacon.ValidatorState_PendingQueued:
				// Remove minipools that don't have indices yet since they're not actually viable
				//r.log.Printlnf("NOTE: minipool %s (index %s, pubkey %s) was in state %s; removing it", minipoolInfo.Address.Hex(), status.Index, minipoolInfo.ValidatorPubkey.Hex(), string(status.Status))
				minipoolInfo.WasActive = false
			default:
				// Get the validator index
				minipoolInfo.ValidatorIndex = status.Index
				r.minipoolValidatorIndexMap[minipoolInfo.ValidatorIndex] = minipoolInfo

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

	return nil
}

// Get the details for every node that was opted into the Smoothing Pool for at least some portion of this interval
func (r *treeGeneratorImpl_v11) getSmoothingPoolNodeDetails() error {

	// For each NO, get their opt-in status and time of last change in batches
	r.log.Printlnf("%s Getting details of nodes for Smoothing Pool calculation...", r.logPrefix)
	nodeCount := uint64(len(r.networkState.NodeDetails))
	r.nodeDetails = make([]*NodeSmoothingDetails, nodeCount)
	for batchStartIndex := uint64(0); batchStartIndex < nodeCount; batchStartIndex += SmoothingPoolDetailsBatchSize {

		// Get batch start & end index
		iterationStartIndex := batchStartIndex
		iterationEndIndex := min(batchStartIndex+SmoothingPoolDetailsBatchSize, nodeCount)

		// Load details
		var wg errgroup.Group
		for iterationIndex := iterationStartIndex; iterationIndex < iterationEndIndex; iterationIndex++ {
			iterationIndex := iterationIndex
			wg.Go(func() error {
				nativeNodeDetails := r.networkState.NodeDetails[iterationIndex]
				nodeDetails := &NodeSmoothingDetails{
					Index:                   iterationIndex,
					Address:                 nativeNodeDetails.NodeAddress,
					Minipools:               []*MinipoolInfo{},
					SmoothingPoolEth:        big.NewInt(0),
					BonusEth:                big.NewInt(0),
					RewardsNetwork:          nativeNodeDetails.RewardNetwork.Uint64(),
					LegacyStakedRpl:         nativeNodeDetails.LegacyStakedRPL,
					MegapoolStakedRpl:       nativeNodeDetails.MegapoolStakedRPL,
					MegapoolVoteEligibleRpl: big.NewInt(0),
					VoterShareEth:           big.NewInt(0),
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
							Address:                 mpd.MinipoolAddress,
							ValidatorPubkey:         mpd.Pubkey,
							Node:                    nodeDetails,
							Fee:                     nativeMinipoolDetails.NodeFee,
							MissingAttestationSlots: map[uint64]bool{},
							CompletedAttestations:   map[uint64]bool{},
							WasActive:               true,
							AttestationScore:        NewQuotedBigInt(0),
							NodeOperatorBond:        nativeMinipoolDetails.NodeDepositBalance,
						})
					}
				}

				if nativeNodeDetails.MegapoolDeployed {
					// Get the megapool details
					megapoolAddress := nativeNodeDetails.MegapoolAddress
					nativeMegapoolDetails := r.networkState.MegapoolDetails[megapoolAddress]
					validators := r.networkState.MegapoolToPubkeysMap[megapoolAddress]

					mpInfo := &MegapoolInfo{
						Address:              megapoolAddress,
						Node:                 nodeDetails,
						Validators:           []*MegapoolValidatorInfo{},
						ValidatorIndexMap:    make(map[string]*MegapoolValidatorInfo),
						ActiveValidatorCount: nativeMegapoolDetails.ActiveValidatorCount,
					}

					for _, validator := range validators {
						status, exists := r.networkState.MegapoolValidatorDetails[validator]
						if !exists {
							continue
						}

						nativeValidatorInfo, exists := r.networkState.MegapoolValidatorInfo[validator]
						if !exists {
							continue
						}

						v := &MegapoolValidatorInfo{
							Pubkey:                  validator,
							Index:                   status.Index,
							MissingAttestationSlots: map[uint64]bool{},
							AttestationScore:        NewQuotedBigInt(0),
							CompletedAttestations:   map[uint64]bool{},
							NativeValidatorInfo:     nativeValidatorInfo,
						}

						mpInfo.Validators = append(mpInfo.Validators, v)
						mpInfo.ValidatorIndexMap[v.Index] = v

						// Check if the megapool has staked RPL
						if nativeNodeDetails.MegapoolStakedRPL.Sign() > 0 {
							// The megapool's eligible staked RPL is defined by
							// min(1.5*RPL value of megapool bonded_eth, megapool staked rpl)
							bondedEth := nativeNodeDetails.MegapoolEthBonded
							rplPrice := r.networkState.NetworkDetails.RplPrice
							// Price is eth per rpl, so to calculate the rpl value of the bonded eth,
							// we need to divide the bonded eth by the price. This nukes the 1eth unit, so
							// multiply by 1.5 eth first.
							bondedEthRplValue := big.NewInt(0).Mul(bondedEth, big.NewInt(15e17))
							bondedEthRplValue.Div(bondedEthRplValue, rplPrice)
							// Now take the minimum of the node's actual rpl vs bondedEthRplValue
							if nativeNodeDetails.MegapoolStakedRPL.Cmp(bondedEthRplValue) < 0 {
								nodeDetails.MegapoolVoteEligibleRpl = nativeNodeDetails.MegapoolStakedRPL
							} else {
								nodeDetails.MegapoolVoteEligibleRpl = bondedEthRplValue
							}
						}
					}

					nodeDetails.Megapool = mpInfo
					// The node is eligible if it has a megapool or minipools
					nodeDetails.IsEligible = len(validators) > 0 || len(nodeDetails.Minipools) > 0
				} else {
					// The node is eligible if it has minipools
					nodeDetails.IsEligible = len(nodeDetails.Minipools) > 0
				}
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
		nodeDetails.MinipoolEligibleBorrowedEth = r.networkState.GetMinipoolEligibleBorrowedEth(nnd)
	}

	return nil

}

// Validates that the provided network is legal
func (r *treeGeneratorImpl_v11) validateNetwork(network uint64) (bool, error) {
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
func (r *treeGeneratorImpl_v11) getBlocksAndTimesForInterval(previousIntervalEvent rewards.RewardsEvent) (*types.Header, error) {
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
	r.performanceFile.StartTime = startTime

	r.rewardsFile.EndTime = endTime
	r.performanceFile.EndTime = endTime

	r.rewardsFile.ConsensusStartBlock = nextEpoch * r.beaconConfig.SlotsPerEpoch
	r.performanceFile.ConsensusStartBlock = r.rewardsFile.ConsensusStartBlock

	r.rewardsFile.ConsensusEndBlock = r.snapshotEnd.ConsensusBlock
	r.performanceFile.ConsensusEndBlock = r.snapshotEnd.ConsensusBlock

	r.rewardsFile.ExecutionEndBlock = r.snapshotEnd.ExecutionBlock
	r.performanceFile.ExecutionEndBlock = r.snapshotEnd.ExecutionBlock

	// Get the first block that isn't missing
	var elBlockNumber uint64
	for {
		beaconBlock, exists, err := r.bc.GetBeaconBlock(fmt.Sprint(r.rewardsFile.ConsensusStartBlock))
		if err != nil {
			return nil, fmt.Errorf("error getting EL data for BC slot %d: %w", r.rewardsFile.ConsensusStartBlock, err)
		}
		if !exists {
			r.rewardsFile.ConsensusStartBlock++
			r.performanceFile.ConsensusStartBlock++
		} else {
			elBlockNumber = beaconBlock.ExecutionBlockNumber
			break
		}
	}

	var startElHeader *types.Header
	if elBlockNumber == 0 {
		// We are pre-merge, so get the first block after the one from the previous interval
		r.rewardsFile.ExecutionStartBlock = previousIntervalEvent.ExecutionBlock.Uint64() + 1
		r.performanceFile.ExecutionStartBlock = r.rewardsFile.ExecutionStartBlock
		startElHeader, err = r.rp.HeaderByNumber(context.Background(), big.NewInt(int64(r.rewardsFile.ExecutionStartBlock)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL start block %d: %w", r.rewardsFile.ExecutionStartBlock, err)
		}
	} else {
		// We are post-merge, so get the EL block corresponding to the BC block
		r.rewardsFile.ExecutionStartBlock = elBlockNumber
		r.performanceFile.ExecutionStartBlock = r.rewardsFile.ExecutionStartBlock
		startElHeader, err = r.rp.HeaderByNumber(context.Background(), big.NewInt(int64(elBlockNumber)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL header for block %d: %w", elBlockNumber, err)
		}
	}

	return startElHeader, nil
}

func (r *treeGeneratorImpl_v11) saveFiles(smartnode *config.SmartnodeConfig, treeResult *GenerateTreeResult, nodeTrusted bool) (cid.Cid, map[string]cid.Cid, error) {
	return saveRewardsArtifacts(smartnode, treeResult, nodeTrusted)
}
