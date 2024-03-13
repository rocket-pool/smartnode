// === WARNING ===
// This is kept around for legacy / reference purposes.
// It is NOT optimized to work with rocketpool-go v2 and will run noticeably slower than the other intervals!

package rewards

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/config"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"golang.org/x/sync/errgroup"
)

// Implementation for tree generator ruleset v4
type treeGeneratorImpl_v4 struct {
	rewardsFile            *RewardsFile_v1
	elSnapshotHeader       *types.Header
	log                    *log.ColorLogger
	logPrefix              string
	rp                     *rocketpool.RocketPool
	cfg                    *config.SmartNodeConfig
	bc                     beacon.IBeaconClient
	opts                   *bind.CallOpts
	nodeAddresses          []common.Address
	nodeDetails            []*NodeSmoothingDetails
	smoothingPoolBalance   *big.Int
	smoothingPoolAddress   common.Address
	intervalDutiesInfo     *IntervalDutiesInfo
	slotsPerEpoch          uint64
	validatorIndexMap      map[string]*MinipoolInfo
	elStartTime            time.Time
	elEndTime              time.Time
	validNetworkCache      map[uint64]bool
	epsilon                *big.Int
	intervalSeconds        *big.Int
	beaconConfig           beacon.Eth2Config
	stakingMinipoolMap     map[common.Address][]MinipoolDetails
	validatorStatusMap     map[beacon.ValidatorPubkey]beacon.ValidatorStatus
	rplPrice               *big.Int
	minCollateralFraction  *big.Int
	maxCollateralFraction  *big.Int
	stakingMinipoolPubkeys []beacon.ValidatorPubkey
	nodeStakes             []*big.Int
}

// Create a new tree generator
func newTreeGeneratorImpl_v4(log *log.ColorLogger, logPrefix string, index uint64, startTime time.Time, endTime time.Time, consensusBlock uint64, elSnapshotHeader *types.Header, intervalsPassed uint64) *treeGeneratorImpl_v4 {
	return &treeGeneratorImpl_v4{
		rewardsFile: &RewardsFile_v1{
			RewardsFileHeader: &sharedtypes.RewardsFileHeader{
				RewardsFileVersion:  1,
				RulesetVersion:      4,
				Index:               index,
				StartTime:           startTime.UTC(),
				EndTime:             endTime.UTC(),
				ConsensusEndBlock:   consensusBlock,
				ExecutionEndBlock:   elSnapshotHeader.Number.Uint64(),
				IntervalsPassed:     intervalsPassed,
				InvalidNetworkNodes: map[common.Address]uint64{},
				TotalRewards: &sharedtypes.TotalRewards{
					ProtocolDaoRpl:               sharedtypes.NewQuotedBigInt(0),
					TotalCollateralRpl:           sharedtypes.NewQuotedBigInt(0),
					TotalOracleDaoRpl:            sharedtypes.NewQuotedBigInt(0),
					TotalSmoothingPoolEth:        sharedtypes.NewQuotedBigInt(0),
					PoolStakerSmoothingPoolEth:   sharedtypes.NewQuotedBigInt(0),
					NodeOperatorSmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
				},
				NetworkRewards: map[uint64]*sharedtypes.NetworkRewardsInfo{},
			},
			NodeRewards: map[common.Address]*NodeRewardsInfo_v1{},
			MinipoolPerformanceFile: MinipoolPerformanceFile_v1{
				Index:               index,
				StartTime:           startTime.UTC(),
				EndTime:             endTime.UTC(),
				ConsensusEndBlock:   consensusBlock,
				ExecutionEndBlock:   elSnapshotHeader.Number.Uint64(),
				MinipoolPerformance: map[common.Address]*SmoothingPoolMinipoolPerformance_v1{},
			},
		},
		stakingMinipoolMap: map[common.Address][]MinipoolDetails{},
		validatorStatusMap: map[beacon.ValidatorPubkey]beacon.ValidatorStatus{},
		elSnapshotHeader:   elSnapshotHeader,
		log:                log,
		logPrefix:          logPrefix,
	}
}

// Get the version of the ruleset used by this generator
func (r *treeGeneratorImpl_v4) getRulesetVersion() uint64 {
	return r.rewardsFile.RulesetVersion
}

func (r *treeGeneratorImpl_v4) generateTree(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient) (sharedtypes.IRewardsFile, error) {

	r.log.Printlnf("%s Generating tree using Ruleset v%d.", r.logPrefix, r.rewardsFile.RulesetVersion)

	// Provision some struct params
	r.rp = rp
	r.cfg = cfg
	r.bc = bc
	r.validNetworkCache = map[uint64]bool{
		0: true,
	}

	// Set the network name
	r.rewardsFile.Network = fmt.Sprint(cfg.Network.Value)
	r.rewardsFile.MinipoolPerformanceFile.Network = r.rewardsFile.Network

	// Get the Beacon config
	var err error
	r.beaconConfig, err = r.bc.GetEth2Config(context)
	if err != nil {
		return nil, err
	}
	r.slotsPerEpoch = r.beaconConfig.SlotsPerEpoch

	// Get the addresses for all nodes
	r.opts = &bind.CallOpts{
		BlockNumber: r.elSnapshotHeader.Number,
	}

	// Create the bindings
	nodeMgr, err := node.NewNodeManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating node manager binding: %w", err)
	}
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool manager binding: %w", err)
	}

	// Query the state
	err = rp.Query(nil, r.opts, nodeMgr.NodeCount, mpMgr.MinipoolCount)
	if err != nil {
		return nil, fmt.Errorf("error getting initial contract state: %w", err)
	}

	nodeAddresses, err := nodeMgr.GetNodeAddresses(nodeMgr.NodeCount.Formatted(), r.opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}
	r.log.Printlnf("%s Creating tree for %d nodes", r.logPrefix, len(nodeAddresses))
	r.nodeAddresses = nodeAddresses

	// Get the minipool count - this will be used for an error epsilon due to division truncation
	minipoolCount := mpMgr.MinipoolCount.Formatted()
	r.epsilon = big.NewInt(int64(minipoolCount))

	// Create the minipool details cache
	err = r.cacheMinipoolDetails()
	if err != nil {
		return nil, fmt.Errorf("error caching minipool details: %w", err)
	}

	// Calculate the RPL rewards
	err = r.calculateRplRewards(context)
	if err != nil {
		return nil, fmt.Errorf("error calculating RPL rewards: %w", err)
	}

	// Calculate the ETH rewards
	err = r.calculateEthRewards(context, true)
	if err != nil {
		return nil, fmt.Errorf("error calculating ETH rewards: %w", err)
	}

	// Calculate the network reward map and the totals
	r.updateNetworksAndTotals()

	// Generate the Merkle Tree
	err = r.rewardsFile.GenerateMerkleTree()
	if err != nil {
		return nil, fmt.Errorf("error generating Merkle tree: %w", err)
	}

	// Sort all of the missed attestations so the files are always generated in the same state
	for _, minipoolInfo := range r.rewardsFile.MinipoolPerformanceFile.MinipoolPerformance {
		sort.Slice(minipoolInfo.MissingAttestationSlots, func(i, j int) bool {
			return minipoolInfo.MissingAttestationSlots[i] < minipoolInfo.MissingAttestationSlots[j]
		})
	}

	return r.rewardsFile, nil

}

// Quickly calculates an approximate of the staker's share of the smoothing pool balance without processing Beacon performance
// Used for approximate returns in the rETH ratio update
func (r *treeGeneratorImpl_v4) approximateStakerShareOfSmoothingPool(context context.Context, rp *rocketpool.RocketPool, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient) (*big.Int, error) {
	r.log.Printlnf("%s Approximating tree using Ruleset v%d.", r.logPrefix, r.rewardsFile.RulesetVersion)

	r.rp = rp
	r.cfg = cfg
	r.bc = bc
	r.validNetworkCache = map[uint64]bool{
		0: true,
	}

	// Set the network name
	r.rewardsFile.Network = fmt.Sprint(cfg.Network.Value)
	r.rewardsFile.MinipoolPerformanceFile.Network = r.rewardsFile.Network

	// Get the Beacon config
	var err error
	r.beaconConfig, err = r.bc.GetEth2Config(context)
	if err != nil {
		return nil, err
	}
	r.slotsPerEpoch = r.beaconConfig.SlotsPerEpoch

	// Get the addresses for all nodes
	r.opts = &bind.CallOpts{
		BlockNumber: r.elSnapshotHeader.Number,
	}

	// Create the bindings
	nodeMgr, err := node.NewNodeManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating node manager binding: %w", err)
	}
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool manager binding: %w", err)
	}

	// Query the state
	err = rp.Query(nil, r.opts, nodeMgr.NodeCount, mpMgr.MinipoolCount)
	if err != nil {
		return nil, fmt.Errorf("error getting initial contract state: %w", err)
	}

	nodeAddresses, err := nodeMgr.GetNodeAddresses(nodeMgr.NodeCount.Formatted(), r.opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}
	r.log.Printlnf("%s Creating tree for %d nodes", r.logPrefix, len(nodeAddresses))
	r.nodeAddresses = nodeAddresses

	// Get the minipool count - this will be used for an error epsilon due to division truncation
	minipoolCount := mpMgr.MinipoolCount.Formatted()
	r.epsilon = big.NewInt(int64(minipoolCount))

	// Create the minipool details cache
	err = r.cacheMinipoolDetails()
	if err != nil {
		return nil, fmt.Errorf("error caching minipool details: %w", err)
	}

	// Calculate the ETH rewards
	err = r.calculateEthRewards(context, false)
	if err != nil {
		return nil, fmt.Errorf("error calculating ETH rewards: %w", err)
	}

	return &r.rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Int, nil
}

// Calculates the per-network distribution amounts and the total reward amounts
func (r *treeGeneratorImpl_v4) updateNetworksAndTotals() {

	// Get the highest network index with valid rewards
	highestNetworkIndex := uint64(0)
	for network := range r.rewardsFile.NetworkRewards {
		if network > highestNetworkIndex {
			highestNetworkIndex = network
		}
	}

	// Create the map for each network, including unused ones
	for network := uint64(0); network <= highestNetworkIndex; network++ {
		_, exists := r.rewardsFile.NetworkRewards[network]
		if !exists {
			rewardsForNetwork := &sharedtypes.NetworkRewardsInfo{
				CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
				OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
				SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
			}
			r.rewardsFile.NetworkRewards[network] = rewardsForNetwork
		}
	}

}

// Calculates the RPL rewards for the given interval
func (r *treeGeneratorImpl_v4) calculateRplRewards(context context.Context) error {
	// Create the bindings
	rewardsPool, err := rewards.NewRewardsPool(r.rp)
	if err != nil {
		return fmt.Errorf("error creating rewards pool binding: %w", err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(r.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	pSettings := pMgr.Settings
	networkMgr, err := network.NewNetworkManager(r.rp)
	if err != nil {
		return fmt.Errorf("error creating network manager binding: %w", err)
	}

	// Get the state
	var percentages protocol.RplRewardsPercentages
	err = r.rp.Query(func(mc *batch.MultiCaller) error {
		pMgr.GetRewardsPercentages(mc, &percentages)
		eth.AddQueryablesToMulticall(mc,
			rewardsPool.PendingRplRewards,
			rewardsPool.IntervalDuration,
			networkMgr.RplPrice,
			pSettings.Node.MinimumPerMinipoolStake,
			pSettings.Node.MaximumPerMinipoolStake,
		)
		return nil
	}, r.opts)
	if err != nil {
		return fmt.Errorf("error getting rewards pool details: %w", err)
	}

	// Get the RPL min and max collateral stats
	r.rplPrice = networkMgr.RplPrice.Raw()
	r.minCollateralFraction = pSettings.Node.MinimumPerMinipoolStake.Raw()
	r.maxCollateralFraction = pSettings.Node.MaximumPerMinipoolStake.Raw()

	// Handle node operator rewards
	snapshotBlockTime := time.Unix(int64(r.elSnapshotHeader.Time), 0)
	intervalDuration := rewardsPool.IntervalDuration.Formatted()
	nodeOpPercent := percentages.NodePercentage
	pendingRewards := rewardsPool.PendingRplRewards.Get()

	r.log.Printlnf("%s Pending RPL rewards: %s (%.3f)", r.logPrefix, pendingRewards.String(), eth.WeiToEth(pendingRewards))
	totalNodeRewards := big.NewInt(0)
	totalNodeRewards.Mul(pendingRewards, nodeOpPercent)
	totalNodeRewards.Div(totalNodeRewards, eth.EthToWei(1))
	r.log.Printlnf("%s Approx. total collateral RPL rewards: %s (%.3f)", r.logPrefix, totalNodeRewards.String(), eth.WeiToEth(totalNodeRewards))

	// Get the effective stakes of each node
	effectiveStakes, err := r.getNodeEffectiveRPLStakes(context)
	if err != nil {
		return fmt.Errorf("error calculating effective RPL stakes: %w", err)
	}

	// Calculate the true effective stake of each node based on their participation in this interval
	totalNodeEffectiveStake := big.NewInt(0)
	trueNodeEffectiveStakes := map[common.Address]*big.Int{}
	intervalDurationBig := big.NewInt(int64(intervalDuration.Seconds()))
	r.log.Printlnf("%s Calculating true total collateral rewards (progress is reported every 100 nodes)", r.logPrefix)
	nodesDone := 0
	startTime := time.Now()
	nodeCount := len(r.nodeAddresses)

	// Get node details
	nodes := map[common.Address]*node.Node{}
	for _, address := range r.nodeAddresses {
		// Create the node binding
		node, err := node.NewNode(r.rp, address)
		if err != nil {
			return fmt.Errorf("error creating node %s binding: %w", address.Hex(), err)
		}
		nodes[address] = node
	}
	err = r.rp.BatchQuery(nodeCount, LegacyDetailsBatchCount, func(mc *batch.MultiCaller, i int) error {
		address := r.nodeAddresses[i]
		node := nodes[address]
		eth.AddQueryablesToMulticall(mc,
			node.EffectiveRplStake,
			node.RegistrationTime,
			node.RewardNetwork,
		)
		return nil
	}, r.opts)
	if err != nil {
		return fmt.Errorf("error getting node details: %w", err)
	}

	for i, address := range r.nodeAddresses {
		if nodesDone == 100 {
			timeTaken := time.Since(startTime)
			r.log.Printlnf("%s On Node %d of %d (%.2f%%)... (%s so far)", r.logPrefix, i, nodeCount, float64(i)/float64(nodeCount)*100.0, timeTaken)
			nodesDone = 0
		}

		// Get the details
		node := nodes[address]
		regTime := node.RegistrationTime.Formatted()

		// Get the node's effective stake
		nodeStake := effectiveStakes[i]

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

		nodesDone++
	}

	r.log.Printlnf("%s Calculating individual collateral rewards (progress is reported every 100 nodes)", r.logPrefix)
	nodesDone = 0
	startTime = time.Now()
	for i, address := range r.nodeAddresses {
		if nodesDone == 100 {
			timeTaken := time.Since(startTime)
			r.log.Printlnf("%s On Node %d of %d (%.2f%%)... (%s so far)", r.logPrefix, i, nodeCount, float64(i)/float64(nodeCount)*100.0, timeTaken)
			nodesDone = 0
		}
		node := nodes[address]

		// Get how much RPL goes to this node: (true effective stake) * (total node rewards) / (total true effective stake)
		nodeRplRewards := big.NewInt(0)
		nodeRplRewards.Mul(trueNodeEffectiveStakes[address], totalNodeRewards)
		if totalNodeEffectiveStake.Cmp(big.NewInt(0)) > 0 {
			nodeRplRewards.Div(nodeRplRewards, totalNodeEffectiveStake)
		}

		// If there are pending rewards, add it to the map
		if nodeRplRewards.Cmp(big.NewInt(0)) == 1 {
			rewardsForNode, exists := r.rewardsFile.NodeRewards[address]
			if !exists {
				// Get the network the rewards should go to
				network := node.RewardNetwork.Formatted()
				validNetwork, err := r.validateNetwork(network)
				if err != nil {
					return err
				}
				if !validNetwork {
					r.rewardsFile.InvalidNetworkNodes[address] = network
					network = 0
				}

				rewardsForNode = &NodeRewardsInfo_v1{
					RewardNetwork:    network,
					CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
					OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
					SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
				}
				r.rewardsFile.NodeRewards[address] = rewardsForNode
			}
			rewardsForNode.CollateralRpl.Add(&rewardsForNode.CollateralRpl.Int, nodeRplRewards)

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.rewardsFile.NetworkRewards[rewardsForNode.RewardNetwork]
			if !exists {
				rewardsForNetwork = &sharedtypes.NetworkRewardsInfo{
					CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
					OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
					SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
				}
				r.rewardsFile.NetworkRewards[rewardsForNode.RewardNetwork] = rewardsForNetwork
			}
			rewardsForNetwork.CollateralRpl.Add(&rewardsForNetwork.CollateralRpl.Int, nodeRplRewards)
		}

		nodesDone++
	}

	// Sanity check to make sure we arrived at the correct total

	delta := big.NewInt(0)
	totalCalculatedNodeRewards := big.NewInt(0)
	for _, networkRewards := range r.rewardsFile.NetworkRewards {
		totalCalculatedNodeRewards.Add(totalCalculatedNodeRewards, &networkRewards.CollateralRpl.Int)
	}
	delta.Sub(totalNodeRewards, totalCalculatedNodeRewards).Abs(delta)
	if totalNodeEffectiveStake.Cmp(big.NewInt(0)) > 0 {
		if delta.Cmp(r.epsilon) == 1 {
			return fmt.Errorf("error calculating collateral RPL: total was %s, but expected %s; error was too large", totalCalculatedNodeRewards.String(), totalNodeRewards.String())
		}
	}
	r.rewardsFile.TotalRewards.TotalCollateralRpl.Int = *totalCalculatedNodeRewards
	r.log.Printlnf("%s Calculated rewards:           %s (error = %s wei)", r.logPrefix, totalCalculatedNodeRewards.String(), delta.String())

	// Handle Oracle DAO rewards
	oDaoPercent := percentages.OdaoPercentage
	totalODaoRewards := big.NewInt(0)
	totalODaoRewards.Mul(pendingRewards, oDaoPercent)
	totalODaoRewards.Div(totalODaoRewards, eth.EthToWei(1))
	r.log.Printlnf("%s Total Oracle DAO RPL rewards: %s (%.3f)", r.logPrefix, totalODaoRewards.String(), eth.WeiToEth(totalODaoRewards))

	// Create the bindings
	odaoMgr, err := oracle.NewOracleDaoManager(r.rp)
	if err != nil {
		return fmt.Errorf("error getting DNT binding: %w", err)
	}

	// Get the contract state
	err = r.rp.Query(nil, r.opts, odaoMgr.MemberCount)
	if err != nil {
		return fmt.Errorf("error getting oDAO member count: %w", err)
	}

	// Get the oDAO member addresses
	oDaoAddresses, err := odaoMgr.GetMemberAddresses(odaoMgr.MemberCount.Formatted(), r.opts)
	if err != nil {
		return err
	}

	// Calculate the true effective time of each oDAO node based on their participation in this interval
	totalODaoNodeTime := big.NewInt(0)
	trueODaoNodeTimes := map[common.Address]*big.Int{}
	for _, address := range oDaoAddresses {
		node := nodes[address]
		// Get the timestamp of the node's registration
		regTime := node.RegistrationTime.Formatted()

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

		rewardsForNode, exists := r.rewardsFile.NodeRewards[address]
		if !exists {
			node := nodes[address]
			// Get the network the rewards should go to
			network := node.RewardNetwork.Formatted()
			validNetwork, err := r.validateNetwork(network)
			if err != nil {
				return err
			}
			if !validNetwork {
				r.rewardsFile.InvalidNetworkNodes[address] = network
				network = 0
			}

			rewardsForNode = &NodeRewardsInfo_v1{
				RewardNetwork:    network,
				CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
				OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
				SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
			}
			r.rewardsFile.NodeRewards[address] = rewardsForNode

		}
		rewardsForNode.OracleDaoRpl.Add(&rewardsForNode.OracleDaoRpl.Int, individualOdaoRewards)

		// Add the rewards to the running total for the specified network
		rewardsForNetwork, exists := r.rewardsFile.NetworkRewards[rewardsForNode.RewardNetwork]
		if !exists {
			rewardsForNetwork = &sharedtypes.NetworkRewardsInfo{
				CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
				OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
				SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
			}
			r.rewardsFile.NetworkRewards[rewardsForNode.RewardNetwork] = rewardsForNetwork
		}
		rewardsForNetwork.OracleDaoRpl.Add(&rewardsForNetwork.OracleDaoRpl.Int, individualOdaoRewards)
	}

	// Sanity check to make sure we arrived at the correct total
	if big.NewInt(int64(len(oDaoAddresses))).Cmp(r.epsilon) > 0 {
		r.epsilon.SetUint64(uint64(len(oDaoAddresses)))
	}
	totalCalculatedOdaoRewards := big.NewInt(0)
	delta = big.NewInt(0)
	for _, networkRewards := range r.rewardsFile.NetworkRewards {
		totalCalculatedOdaoRewards.Add(totalCalculatedOdaoRewards, &networkRewards.OracleDaoRpl.Int)
	}
	delta.Sub(totalODaoRewards, totalCalculatedOdaoRewards).Abs(delta)
	if delta.Cmp(r.epsilon) == 1 {
		return fmt.Errorf("error calculating ODao RPL: total was %s, but expected %s; error was too large", totalCalculatedOdaoRewards.String(), totalODaoRewards.String())
	}
	r.rewardsFile.TotalRewards.TotalOracleDaoRpl.Int = *totalCalculatedOdaoRewards
	r.log.Printlnf("%s Calculated rewards:           %s (error = %s wei)", r.logPrefix, totalCalculatedOdaoRewards.String(), delta.String())

	// Get expected Protocol DAO rewards
	pDaoPercent := percentages.PdaoPercentage
	pDaoRewards := sharedtypes.NewQuotedBigInt(0)
	pDaoRewards.Mul(pendingRewards, pDaoPercent)
	pDaoRewards.Div(&pDaoRewards.Int, eth.EthToWei(1))
	r.log.Printlnf("%s Expected Protocol DAO rewards: %s (%.3f)", r.logPrefix, pDaoRewards.String(), eth.WeiToEth(&pDaoRewards.Int))

	// Get actual protocol DAO rewards
	pDaoRewards.Sub(pendingRewards, totalCalculatedNodeRewards)
	pDaoRewards.Sub(&pDaoRewards.Int, totalCalculatedOdaoRewards)
	r.rewardsFile.TotalRewards.ProtocolDaoRpl = pDaoRewards
	r.log.Printlnf("%s Actual Protocol DAO rewards:   %s to account for truncation", r.logPrefix, pDaoRewards.String())

	return nil

}

// Calculates the ETH rewards for the given interval
func (r *treeGeneratorImpl_v4) calculateEthRewards(context context.Context, checkBeaconPerformance bool) error {

	// Get the Smoothing Pool contract's balance
	smoothingPoolContract, err := r.rp.GetContract(rocketpool.ContractName_RocketSmoothingPool)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool contract: %w", err)
	}
	r.smoothingPoolAddress = smoothingPoolContract.Address

	r.smoothingPoolBalance, err = r.rp.Client.BalanceAt(context, smoothingPoolContract.Address, r.elSnapshotHeader.Number)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}
	r.log.Printlnf("%s Smoothing Pool Balance: %s (%.3f)", r.logPrefix, r.smoothingPoolBalance.String(), eth.WeiToEth(r.smoothingPoolBalance))

	// Ignore the ETH calculation if there are no rewards
	if r.smoothingPoolBalance.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if r.rewardsFile.Index == 0 {
		// This is the first interval, Smoothing Pool rewards are ignored on the first interval since it doesn't have a discrete start time
		return nil
	}

	// Get the start time of this interval based on the event from the previous one
	previousIntervalEvent, err := GetRewardSnapshotEvent(r.rp, r.cfg, r.rewardsFile.Index-1, r.opts)
	if err != nil {
		return err
	}
	startElBlockHeader, err := r.getStartBlocksForInterval(context, previousIntervalEvent)
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
		err = r.processAttestationsForInterval(context)
		if err != nil {
			return err
		}
	} else {
		// Attestation processing is disabled, just give each minipool 1 good attestation and complete slot activity so they're all scored the same
		// Used for approximating rETH's share during balances calculation
		for _, nodeInfo := range r.nodeDetails {
			if nodeInfo.IsEligible {
				for _, minipool := range nodeInfo.Minipools {
					minipool.GoodAttestations = 1
					minipool.StartSlot = r.rewardsFile.ConsensusStartBlock
					minipool.EndSlot = r.rewardsFile.ConsensusEndBlock
				}
			}
		}
	}

	// Determine how much ETH each node gets and how much the pool stakers get
	poolStakerETH, nodeOpEth, err := r.calculateNodeRewards()
	if err != nil {
		return err
	}

	// Update the rewards maps
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.IsEligible && nodeInfo.SmoothingPoolEth.Cmp(big.NewInt(0)) > 0 {
			rewardsForNode, exists := r.rewardsFile.NodeRewards[nodeInfo.Address]
			if !exists {
				network := nodeInfo.RewardsNetwork
				validNetwork, err := r.validateNetwork(network)
				if err != nil {
					return err
				}
				if !validNetwork {
					r.rewardsFile.InvalidNetworkNodes[nodeInfo.Address] = network
					network = 0
				}

				rewardsForNode = &NodeRewardsInfo_v1{
					RewardNetwork:    network,
					CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
					OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
					SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
				}
				r.rewardsFile.NodeRewards[nodeInfo.Address] = rewardsForNode
			}
			rewardsForNode.SmoothingPoolEth.Add(&rewardsForNode.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
			rewardsForNode.SmoothingPoolEligibilityRate = float64(nodeInfo.EndSlot-nodeInfo.StartSlot) / float64(r.rewardsFile.ConsensusEndBlock-r.rewardsFile.ConsensusStartBlock)

			// Add minipool rewards to the JSON
			for _, minipoolInfo := range nodeInfo.Minipools {
				performance := &SmoothingPoolMinipoolPerformance_v1{
					Pubkey:                  minipoolInfo.ValidatorPubkey.Hex(),
					StartSlot:               minipoolInfo.StartSlot,
					EndSlot:                 minipoolInfo.EndSlot,
					ActiveFraction:          float64(minipoolInfo.EndSlot-minipoolInfo.StartSlot) / float64(r.rewardsFile.ConsensusEndBlock-r.rewardsFile.ConsensusStartBlock),
					SuccessfulAttestations:  minipoolInfo.GoodAttestations,
					MissedAttestations:      minipoolInfo.MissedAttestations,
					EthEarned:               eth.WeiToEth(minipoolInfo.MinipoolShare),
					MissingAttestationSlots: []uint64{},
				}
				if minipoolInfo.GoodAttestations+minipoolInfo.MissedAttestations == 0 {
					performance.ParticipationRate = 0
				} else {
					performance.ParticipationRate = float64(minipoolInfo.GoodAttestations) / float64(minipoolInfo.GoodAttestations+minipoolInfo.MissedAttestations)
				}
				for slot := range minipoolInfo.MissingAttestationSlots {
					performance.MissingAttestationSlots = append(performance.MissingAttestationSlots, slot)
				}
				r.rewardsFile.MinipoolPerformanceFile.MinipoolPerformance[minipoolInfo.Address] = performance
			}

			// Add the rewards to the running total for the specified network
			rewardsForNetwork, exists := r.rewardsFile.NetworkRewards[rewardsForNode.RewardNetwork]
			if !exists {
				rewardsForNetwork = &sharedtypes.NetworkRewardsInfo{
					CollateralRpl:    sharedtypes.NewQuotedBigInt(0),
					OracleDaoRpl:     sharedtypes.NewQuotedBigInt(0),
					SmoothingPoolEth: sharedtypes.NewQuotedBigInt(0),
				}
				r.rewardsFile.NetworkRewards[rewardsForNode.RewardNetwork] = rewardsForNetwork
			}
			rewardsForNetwork.SmoothingPoolEth.Add(&rewardsForNetwork.SmoothingPoolEth.Int, nodeInfo.SmoothingPoolEth)
		}
	}

	// Set the totals
	r.rewardsFile.TotalRewards.PoolStakerSmoothingPoolEth.Int = *poolStakerETH
	r.rewardsFile.TotalRewards.NodeOperatorSmoothingPoolEth.Int = *nodeOpEth
	r.rewardsFile.TotalRewards.TotalSmoothingPoolEth.Int = *r.smoothingPoolBalance
	return nil

}

// Calculate the distribution of Smoothing Pool ETH to each node
func (r *treeGeneratorImpl_v4) calculateNodeRewards() (*big.Int, *big.Int, error) {

	// Get the average fee for all eligible minipools and calculate their weighted share
	one := big.NewInt(1e18) // 100%, used for dividing percentages properly
	feeTotal := big.NewInt(0)
	minipoolCount := int64(0)
	minipoolShareTotal := big.NewInt(0)
	intervalSlots := r.rewardsFile.ConsensusEndBlock - r.rewardsFile.ConsensusStartBlock
	intervalSlotsBig := big.NewInt(int64(intervalSlots))
	for _, nodeInfo := range r.nodeDetails {
		if nodeInfo.IsEligible {
			for _, minipool := range nodeInfo.Minipools {
				if minipool.GoodAttestations+minipool.MissedAttestations == 0 || !minipool.WasActive {
					// Ignore minipools that weren't active for the interval
					minipool.StartSlot = 0
					minipool.EndSlot = 0
					minipool.WasActive = false
					minipool.MinipoolShare = big.NewInt(0)
					continue
				}
				// Used for average fee calculation
				feeTotal.Add(feeTotal, minipool.Fee)
				minipoolCount++

				// Minipool share calculation
				minipoolShare := big.NewInt(0).Add(one, minipool.Fee) // Start with 1 + fee
				if uint64(minipool.EndSlot-minipool.StartSlot) < intervalSlots {
					// Prorate the minipool based on its number of active slots
					activeSlots := big.NewInt(int64(minipool.EndSlot - minipool.StartSlot))
					minipoolShare.Mul(minipoolShare, activeSlots)
					minipoolShare.Div(minipoolShare, intervalSlotsBig)
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
	r.log.Printlnf("%s Fee Total:          %s (%.3f)", r.logPrefix, feeTotal.String(), eth.WeiToEth(feeTotal))
	r.log.Printlnf("%s Minipool Count:     %d", r.logPrefix, minipoolCount)
	r.log.Printlnf("%s Average Fee:        %s (%.3f)", r.logPrefix, averageFee.String(), eth.WeiToEth(averageFee))

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
				if minipool.EndSlot-minipool.StartSlot == 0 {
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

	r.log.Printlnf("%s Pool staker ETH:    %s (%.3f)", r.logPrefix, poolStakerShare.String(), eth.WeiToEth(poolStakerShare))
	r.log.Printlnf("%s Node Op ETH:        %s (%.3f)", r.logPrefix, nodeOpShare.String(), eth.WeiToEth(nodeOpShare))
	r.log.Printlnf("%s Calculated NO ETH:  %s (error = %s wei)", r.logPrefix, totalEthForMinipools.String(), delta.String())
	r.log.Printlnf("%s Adjusting pool staker ETH to %s to account for truncation", r.logPrefix, truePoolStakerAmount.String())

	return truePoolStakerAmount, totalEthForMinipools, nil

}

// Get all of the duties for a range of epochs
func (r *treeGeneratorImpl_v4) processAttestationsForInterval(context context.Context) error {

	startEpoch := r.rewardsFile.ConsensusStartBlock / r.beaconConfig.SlotsPerEpoch
	endEpoch := r.rewardsFile.ConsensusEndBlock / r.beaconConfig.SlotsPerEpoch

	// Determine the validator indices of each minipool
	err := r.createMinipoolIndexMap(context)
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

		err := r.processEpoch(context, true, epoch)
		if err != nil {
			return err
		}

		epochsDone++
	}

	// Check the epoch after the end of the interval for any lingering attestations
	epoch := endEpoch + 1
	err = r.processEpoch(context, false, epoch)
	if err != nil {
		return err
	}

	r.log.Printlnf("%s Finished participation check (total time = %s)", r.logPrefix, time.Since(reportStartTime))
	return nil

}

// Process an epoch, optionally getting the duties for all eligible minipools in it and checking each one's attestation performance
func (r *treeGeneratorImpl_v4) processEpoch(context context.Context, getDuties bool, epoch uint64) error {

	// Get the committee info and attestation records for this epoch
	var committeeData beacon.Committees
	attestationsPerSlot := make([][]beacon.AttestationInfo, r.slotsPerEpoch)
	var wg errgroup.Group

	if getDuties {
		wg.Go(func() error {
			var err error
			committeeData, err = r.bc.GetCommitteesForEpoch(context, &epoch)
			return err
		})
	}

	for i := uint64(0); i < r.slotsPerEpoch; i++ {
		i := i
		slot := epoch*r.slotsPerEpoch + i
		wg.Go(func() error {
			attestations, found, err := r.bc.GetAttestations(context, fmt.Sprint(slot))
			if err != nil {
				return err
			}
			if found {
				attestationsPerSlot[i] = attestations
			} else {
				attestationsPerSlot[i] = []beacon.AttestationInfo{}
			}
			return nil
		})
	}
	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("error getting committee and attestaion records for epoch %d: %w", epoch, err)
	}

	if getDuties {
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
			r.checkDutiesForSlot(attestations)
		}
	}

	return nil

}

// Handle all of the attestations in the given slot
func (r *treeGeneratorImpl_v4) checkDutiesForSlot(attestations []beacon.AttestationInfo) error {

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
func (r *treeGeneratorImpl_v4) getDutiesForEpoch(committees beacon.Committees) error {

	defer committees.Release()

	// Crawl the committees
	for idx := 0; idx < committees.Count(); idx++ {
		slotIndex := committees.Slot(idx)
		if slotIndex < r.rewardsFile.ConsensusStartBlock || slotIndex > r.rewardsFile.ConsensusEndBlock {
			// Ignore slots that are out of bounds
			continue
		}
		committeeIndex := committees.Index(idx)

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*MinipoolInfo{}
		for position, validator := range committees.Validators(idx) {
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
func (r *treeGeneratorImpl_v4) createMinipoolIndexMap(context context.Context) error {

	// Make a slice of all minipool pubkeys
	uncachedMinipoolPubkeys := []beacon.ValidatorPubkey{}
	for _, details := range r.nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				_, exists := r.validatorStatusMap[minipoolInfo.ValidatorPubkey]
				if !exists {
					uncachedMinipoolPubkeys = append(uncachedMinipoolPubkeys, minipoolInfo.ValidatorPubkey)
				}
			}
		}
	}

	// Get the status for all uncached minipool validators and add them to the cache
	r.validatorIndexMap = map[string]*MinipoolInfo{}
	statusMap, err := r.bc.GetValidatorStatuses(context, uncachedMinipoolPubkeys, &beacon.ValidatorStatusOptions{
		Slot: &r.rewardsFile.ConsensusEndBlock,
	})
	for pubkey, status := range r.validatorStatusMap {
		statusMap[pubkey] = status
	}
	r.validatorStatusMap = statusMap

	if err != nil {
		return fmt.Errorf("error getting validator statuses: %w", err)
	}
	for _, details := range r.nodeDetails {
		if details.IsEligible {
			for _, minipoolInfo := range details.Minipools {
				status, exists := statusMap[minipoolInfo.ValidatorPubkey]
				if !exists {
					// Remove minipools that don't have indices yet since they're not actually viable
					r.log.Printlnf("NOTE: minipool %s (pubkey %s) didn't exist at this slot; removing it", minipoolInfo.Address.Hex(), minipoolInfo.ValidatorPubkey.Hex())
					minipoolInfo.StartSlot = 0
					minipoolInfo.EndSlot = 0
					minipoolInfo.WasActive = false
				} else {
					switch status.Status {
					case beacon.ValidatorState_PendingInitialized, beacon.ValidatorState_PendingQueued:
						// Remove minipools that don't have indices yet since they're not actually viable
						r.log.Printlnf("NOTE: minipool %s (index %s, pubkey %s) was in state %s; removing it", minipoolInfo.Address.Hex(), status.Index, minipoolInfo.ValidatorPubkey.Hex(), string(status.Status))
						minipoolInfo.StartSlot = 0
						minipoolInfo.EndSlot = 0
						minipoolInfo.WasActive = false
					default:
						// Get the validator index
						minipoolInfo.ValidatorIndex = statusMap[minipoolInfo.ValidatorPubkey].Index
						r.validatorIndexMap[minipoolInfo.ValidatorIndex] = minipoolInfo

						// Get the validator's activation start and end slots
						startSlot := status.ActivationEpoch * r.beaconConfig.SlotsPerEpoch
						endSlot := status.ExitEpoch * r.beaconConfig.SlotsPerEpoch

						// Verify this minipool has already started
						if status.ActivationEpoch == FarEpoch {
							minipoolInfo.StartSlot = 0
							minipoolInfo.EndSlot = 0
							minipoolInfo.WasActive = false
							continue
						}

						// Check if the minipool exited before this interval
						if status.ExitEpoch != FarEpoch && endSlot < r.rewardsFile.ConsensusStartBlock {
							r.log.Printlnf("NOTE: minipool %s exited on slot %d which was before interval start %d; removing it", minipoolInfo.Address.Hex(), endSlot, r.rewardsFile.ConsensusStartBlock)
							minipoolInfo.StartSlot = 0
							minipoolInfo.EndSlot = 0
							minipoolInfo.WasActive = false
							continue
						}

						if startSlot > details.EndSlot {
							// This minipool was activated after the node's window ended, so don't count it
							r.log.Printlnf("NOTE: minipool %s was activated on slot %d which was after the node's end slot of %d; removing it", minipoolInfo.Address.Hex(), details.EndSlot, r.rewardsFile.ConsensusStartBlock)
							minipoolInfo.StartSlot = 0
							minipoolInfo.EndSlot = 0
							minipoolInfo.WasActive = false
							continue
						}

						// If this minipool was activated after its node-based start slot, update the start slot
						if startSlot > minipoolInfo.StartSlot {
							minipoolInfo.StartSlot = startSlot
						}

						// If this minipool exited before its node-based end slot, update the end slot
						if status.ExitEpoch != FarEpoch && endSlot < minipoolInfo.EndSlot {
							minipoolInfo.EndSlot = endSlot
						}
					}
				}
			}
		}
	}

	return nil

}

// Get the details for every node that was opted into the Smoothing Pool for at least some portion of this interval
func (r *treeGeneratorImpl_v4) getSmoothingPoolNodeDetails() error {

	genesisTime := time.Unix(int64(r.beaconConfig.GenesisTime), 0)

	nodesDone := uint64(0)
	startTime := time.Now()
	r.log.Printlnf("%s Getting details of nodes for Smoothing Pool calculation (progress is reported every 100 nodes)", r.logPrefix)

	// Get node details
	nodeCount := uint64(len(r.nodeAddresses))
	nodes := map[common.Address]*node.Node{}
	for _, address := range r.nodeAddresses {
		// Create the node binding
		node, err := node.NewNode(r.rp, address)
		if err != nil {
			return fmt.Errorf("error creating node %s binding: %w", address.Hex(), err)
		}
		nodes[address] = node
	}
	err := r.rp.BatchQuery(int(nodeCount), LegacyDetailsBatchCount, func(mc *batch.MultiCaller, i int) error {
		address := r.nodeAddresses[i]
		node := nodes[address]
		eth.AddQueryablesToMulticall(mc,
			node.SmoothingPoolRegistrationState,
			node.SmoothingPoolRegistrationChanged,
			node.RewardNetwork,
			node.MinipoolCount,
		)
		return nil
	}, r.opts)
	if err != nil {
		return fmt.Errorf("error getting node details: %w", err)
	}

	// For each NO, get their opt-in status and time of last change in batches
	r.nodeDetails = make([]*NodeSmoothingDetails, nodeCount)
	for batchStartIndex := uint64(0); batchStartIndex < nodeCount; batchStartIndex += SmoothingPoolDetailsBatchSize {

		// Get batch start & end index
		iterationStartIndex := batchStartIndex
		iterationEndIndex := batchStartIndex + SmoothingPoolDetailsBatchSize
		if iterationEndIndex > nodeCount {
			iterationEndIndex = nodeCount
		}

		if nodesDone >= 100 {
			timeTaken := time.Since(startTime)
			r.log.Printlnf("%s On Node %d of %d (%.2f%%)... (%s so far)", r.logPrefix, iterationStartIndex, nodeCount, float64(iterationStartIndex)/float64(nodeCount)*100.0, timeTaken)
			nodesDone = 0
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
				node := nodes[nodeDetails.Address]

				// Get some details
				nodeDetails.RewardsNetwork = node.RewardNetwork.Formatted()
				nodeDetails.IsOptedIn = node.SmoothingPoolRegistrationState.Get()
				nodeDetails.StatusChangeTime = node.SmoothingPoolRegistrationChanged.Formatted()
				var changeSlot uint64
				if err != nil {
					return fmt.Errorf("error getting smoothing pool registration change time for node %s: %w", nodeDetails.Address.Hex(), err)
				}
				if nodeDetails.StatusChangeTime == time.Unix(0, 0) {
					changeSlot = 0
				} else {
					changeSlot = uint64(nodeDetails.StatusChangeTime.Sub(genesisTime).Seconds()) / r.beaconConfig.SecondsPerSlot
				}

				// If the node isn't opted into the Smoothing Pool and they didn't opt out during this interval, ignore them
				if r.rewardsFile.ConsensusStartBlock > changeSlot && !nodeDetails.IsOptedIn {
					nodeDetails.IsEligible = false
					nodeDetails.EligibleSeconds = big.NewInt(0)
					nodeDetails.StartSlot = 0
					nodeDetails.EndSlot = 0
					r.nodeDetails[iterationIndex] = nodeDetails
					return nil
				}

				// Get the node's total active factor
				if nodeDetails.IsOptedIn {
					nodeDetails.StartSlot = changeSlot
					nodeDetails.EndSlot = r.rewardsFile.ConsensusEndBlock
					// Clamp to this interval
					if nodeDetails.StartSlot < r.rewardsFile.ConsensusStartBlock {
						nodeDetails.StartSlot = r.rewardsFile.ConsensusStartBlock
					}
				} else {
					nodeDetails.StartSlot = r.rewardsFile.ConsensusStartBlock
					nodeDetails.EndSlot = changeSlot
					// Clamp to this interval
					if nodeDetails.EndSlot > r.rewardsFile.ConsensusEndBlock {
						nodeDetails.EndSlot = r.rewardsFile.ConsensusEndBlock
					}
				}

				// Get the details for each minipool in the node
				minipoolDetails, exists := r.stakingMinipoolMap[nodeDetails.Address]
				if !exists {
					return fmt.Errorf("attempted to get the minipool details for node %s, but that node's minipools were missing from the cache", nodeDetails.Address.Hex())
				}
				for _, mpd := range minipoolDetails {
					penaltyCount := mpd.PenaltyCount
					if penaltyCount >= 3 {
						// This node is a cheater
						nodeDetails.IsEligible = false
						nodeDetails.EligibleSeconds = big.NewInt(0)
						nodeDetails.StartSlot = 0
						nodeDetails.EndSlot = 0
						nodeDetails.Minipools = []*MinipoolInfo{}
						r.nodeDetails[iterationIndex] = nodeDetails
						return nil
					}

					// This minipool is below the penalty count, so include it
					fee := mpd.NodeFee
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
						StartSlot:               nodeDetails.StartSlot,
						EndSlot:                 nodeDetails.EndSlot,
					})
				}

				nodeDetails.IsEligible = len(nodeDetails.Minipools) > 0
				r.nodeDetails[iterationIndex] = nodeDetails
				return nil
			})
		}
		if err := wg.Wait(); err != nil {
			return err
		}

		nodesDone += SmoothingPoolDetailsBatchSize
	}

	return nil

}

// Validates that the provided network is legal
func (r *treeGeneratorImpl_v4) validateNetwork(network uint64) (bool, error) {
	valid, exists := r.validNetworkCache[network]
	if !exists {
		// Make the oDAO settings binding
		oMgr, err := oracle.NewOracleDaoManager(r.rp)
		if err != nil {
			return false, fmt.Errorf("error creating oDAO manager binding: %w", err)
		}
		oSettings := oMgr.Settings

		// Get the contract state
		err = r.rp.Query(func(mc *batch.MultiCaller) error {
			oSettings.GetNetworkEnabled(mc, &valid, network)
			return nil
		}, r.opts)
		if err != nil {
			return false, fmt.Errorf("error checking if network %d is enabled: %w", network, err)
		}
		r.validNetworkCache[network] = valid
	}

	return valid, nil
}

// Gets the start blocks for the given interval
func (r *treeGeneratorImpl_v4) getStartBlocksForInterval(context context.Context, previousIntervalEvent rewards.RewardsEvent) (*types.Header, error) {
	// Sanity check to confirm the BN can access the block from the previous interval
	_, exists, err := r.bc.GetBeaconBlock(context, previousIntervalEvent.ConsensusBlock.String())
	if err != nil {
		return nil, fmt.Errorf("error verifying block from previous interval: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("couldn't retrieve CL block from previous interval (slot %d); this likely means you checkpoint sync'd your Beacon Node and it has not backfilled to the previous interval yet so it cannot be used for tree generation", previousIntervalEvent.ConsensusBlock.Uint64())
	}

	previousEpoch := previousIntervalEvent.ConsensusBlock.Uint64() / r.beaconConfig.SlotsPerEpoch
	nextEpoch := previousEpoch + 1
	r.rewardsFile.ConsensusStartBlock = nextEpoch * r.beaconConfig.SlotsPerEpoch
	r.rewardsFile.MinipoolPerformanceFile.ConsensusStartBlock = r.rewardsFile.ConsensusStartBlock

	// Get the first block that isn't missing
	var elBlockNumber uint64
	for {
		beaconBlock, exists, err := r.bc.GetBeaconBlock(context, fmt.Sprint(r.rewardsFile.ConsensusStartBlock))
		if err != nil {
			return nil, fmt.Errorf("error getting EL data for BC slot %d: %w", r.rewardsFile.ConsensusStartBlock, err)
		}
		if !exists {
			r.rewardsFile.ConsensusStartBlock++
			r.rewardsFile.MinipoolPerformanceFile.ConsensusStartBlock++
		} else {
			elBlockNumber = beaconBlock.ExecutionBlockNumber
			break
		}
	}

	var startElHeader *types.Header
	if elBlockNumber == 0 {
		// We are pre-merge, so get the first block after the one from the previous interval
		r.rewardsFile.ExecutionStartBlock = previousIntervalEvent.ExecutionBlock.Uint64() + 1
		r.rewardsFile.MinipoolPerformanceFile.ExecutionStartBlock = r.rewardsFile.ExecutionStartBlock
		startElHeader, err = r.rp.Client.HeaderByNumber(context, big.NewInt(int64(r.rewardsFile.ExecutionStartBlock)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL start block %d: %w", r.rewardsFile.ExecutionStartBlock, err)
		}
	} else {
		// We are post-merge, so get the EL block corresponding to the BC block
		r.rewardsFile.ExecutionStartBlock = elBlockNumber
		r.rewardsFile.MinipoolPerformanceFile.ExecutionStartBlock = r.rewardsFile.ExecutionStartBlock
		startElHeader, err = r.rp.Client.HeaderByNumber(context, big.NewInt(int64(elBlockNumber)))
		if err != nil {
			return nil, fmt.Errorf("error getting EL header for block %d: %w", elBlockNumber, err)
		}
	}

	return startElHeader, nil
}

// Create a cache of the minipool details for each node
func (r *treeGeneratorImpl_v4) cacheMinipoolDetails() error {

	r.stakingMinipoolPubkeys = []beacon.ValidatorPubkey{}
	nodesDone := uint64(0)
	startTime := time.Now()
	r.log.Printlnf("%s Querying minipool info for nodes (progress is reported every 100 nodes)", r.logPrefix)

	nodeCount := uint64(len(r.nodeAddresses))
	stakingMinipoolDetailsList := make([][]MinipoolDetails, nodeCount)
	pubkeyList := make([][]beacon.ValidatorPubkey, nodeCount)
	r.nodeStakes = make([]*big.Int, nodeCount)

	// Get node details
	nodes := map[common.Address]*node.Node{}
	for _, address := range r.nodeAddresses {
		// Create the node binding
		node, err := node.NewNode(r.rp, address)
		if err != nil {
			return fmt.Errorf("error creating node %s binding: %w", address.Hex(), err)
		}
		nodes[address] = node
	}
	err := r.rp.BatchQuery(int(nodeCount), LegacyDetailsBatchCount, func(mc *batch.MultiCaller, i int) error {
		address := r.nodeAddresses[i]
		node := nodes[address]
		eth.AddQueryablesToMulticall(mc,
			node.RplStake,
			node.MinipoolCount,
		)
		return nil
	}, r.opts)
	if err != nil {
		return fmt.Errorf("error getting node details: %w", err)
	}

	mpMgr, err := minipool.NewMinipoolManager(r.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager binding: %w", err)
	}

	// Get the details for each minipool in each node
	for batchStartIndex := uint64(0); batchStartIndex < nodeCount; batchStartIndex += SmoothingPoolDetailsBatchSize {

		// Get batch start & end index
		iterationStartIndex := batchStartIndex
		iterationEndIndex := batchStartIndex + SmoothingPoolDetailsBatchSize
		if iterationEndIndex > nodeCount {
			iterationEndIndex = nodeCount
		}

		if nodesDone >= 100 {
			timeTaken := time.Since(startTime)
			r.log.Printlnf("%s On Node %d of %d (%.2f%%)... (%s so far)", r.logPrefix, iterationStartIndex, nodeCount, float64(iterationStartIndex)/float64(nodeCount)*100.0, timeTaken)
			nodesDone = 0
		}

		// Load details
		var wg errgroup.Group
		for iterationIndex := iterationStartIndex; iterationIndex < iterationEndIndex; iterationIndex++ {
			iterationIndex := iterationIndex
			wg.Go(func() error {
				address := r.nodeAddresses[iterationIndex]
				node := nodes[address]

				// Get the details for each minipool in the node
				mpAddresses, err := node.GetMinipoolAddresses(node.MinipoolCount.Formatted(), r.opts)
				if err != nil {
					return fmt.Errorf("error getting node %s minipool addreses: %w", node.Address.Hex(), err)
				}
				mps, err := mpMgr.CreateMinipoolsFromAddresses(mpAddresses, false, r.opts)
				if err != nil {
					return fmt.Errorf("error getting node %s minipools: %w", node.Address.Hex(), err)
				}
				err = r.rp.BatchQuery(len(mps), LegacyDetailsBatchCount, func(mc *batch.MultiCaller, i int) error {
					mpCommon := mps[i].Common()
					eth.AddQueryablesToMulticall(mc,
						mpCommon.Exists,
						mpCommon.Status,
						mpCommon.PenaltyCount,
						mpCommon.NodeFee,
						mpCommon.Pubkey,
					)
					return nil
				}, r.opts)
				if err != nil {
					return fmt.Errorf("error getting node %s minipool details: %w", node.Address.Hex(), err)
				}
				minipoolDetails := make([]MinipoolDetails, len(mps))
				for i, mp := range mps {
					mpCommon := mp.Common()
					mpd := MinipoolDetails{
						Address:      mpCommon.Address,
						Exists:       mpCommon.Exists.Get(),
						Status:       mpCommon.Status.Formatted(),
						Pubkey:       mpCommon.Pubkey.Get(),
						PenaltyCount: mpCommon.PenaltyCount.Formatted(),
						NodeFee:      mpCommon.NodeFee.Raw(),
					}
					minipoolDetails[i] = mpd
				}

				stakingMinipools := make([]MinipoolDetails, 0, len(minipoolDetails))
				minipoolPubkeys := make([]beacon.ValidatorPubkey, 0, len(minipoolDetails))
				for _, mpd := range minipoolDetails {
					if mpd.Exists {
						status := mpd.Status
						if status == rptypes.MinipoolStatus_Staking {
							stakingMinipools = append(stakingMinipools, mpd)
							minipoolPubkeys = append(minipoolPubkeys, mpd.Pubkey)
						}
					}
				}
				stakingMinipoolDetailsList[iterationIndex] = stakingMinipools
				pubkeyList[iterationIndex] = minipoolPubkeys

				// Cache the node stake since we're processing each node here anyway
				nodeStake := node.RplStake.Get()
				r.nodeStakes[iterationIndex] = nodeStake

				return nil
			})
		}

		if err := wg.Wait(); err != nil {
			return err
		}

		nodesDone += SmoothingPoolDetailsBatchSize
	}

	// Cache the minipool details and aggregate the pubkeys
	for i, address := range r.nodeAddresses {
		r.stakingMinipoolMap[address] = stakingMinipoolDetailsList[i]
		r.stakingMinipoolPubkeys = append(r.stakingMinipoolPubkeys, pubkeyList[i]...)
	}

	return nil

}

// Get the effective stake of a node based on the status of its validators
func (r *treeGeneratorImpl_v4) getNodeEffectiveRPLStakes(context context.Context) ([]*big.Int, error) {

	// Get the status for all staking minipool validators
	r.log.Printlnf("%s Getting validator statuses for all eligible minipools", r.logPrefix)
	r.validatorIndexMap = map[string]*MinipoolInfo{}
	statusMap, err := r.bc.GetValidatorStatuses(context, r.stakingMinipoolPubkeys, &beacon.ValidatorStatusOptions{
		Slot: &r.rewardsFile.ConsensusEndBlock,
	})
	if err != nil {
		return nil, fmt.Errorf("can't get validator statuses: %w", err)
	}

	// Add them to the cache for later
	for pubkey, status := range statusMap {
		r.validatorStatusMap[pubkey] = status
	}

	// Get the effective stake per node
	effectiveStakes := make([]*big.Int, len(r.nodeAddresses))
	for i, address := range r.nodeAddresses {

		// Get the number of eligible minipools
		eligibleMinipools := 0
		for _, mpd := range r.stakingMinipoolMap[address] {
			validatorStatus, exists := statusMap[mpd.Pubkey]
			if !exists {
				r.log.Printlnf("NOTE: minipool %s (pubkey %s) didn't exist, ignoring it in effective RPL calculation", mpd.Address.Hex(), mpd.Pubkey.Hex())
				continue
			}
			intervalEndEpoch := r.rewardsFile.ConsensusEndBlock / r.slotsPerEpoch

			if validatorStatus.ActivationEpoch > intervalEndEpoch {
				r.log.Printlnf("NOTE: Minipool %s starts on epoch %d which is after interval epoch %d so it's not eligible for RPL rewards", mpd.Address.Hex(), validatorStatus.ActivationEpoch, intervalEndEpoch)
				continue
			}
			if validatorStatus.ExitEpoch <= intervalEndEpoch {
				r.log.Printlnf("NOTE: Minipool %s exited on epoch %d which is not after interval epoch %d so it's not eligible for RPL rewards", mpd.Address.Hex(), validatorStatus.ExitEpoch, intervalEndEpoch)
				continue
			}
			eligibleMinipools++
		}

		// Calculate the min and max RPL collateral based on the number of eligible minipools
		_16Eth := eth.EthToWei(16)
		eligibleMinipoolsBig := big.NewInt(int64(eligibleMinipools))

		// minCollateral := 16 * minCollateralFraction * eligibleMinipools / ratio
		// NOTE: minCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
		minCollateral := big.NewInt(0).Mul(_16Eth, r.minCollateralFraction)
		minCollateral.Mul(minCollateral, eligibleMinipoolsBig).Div(minCollateral, r.rplPrice)

		// maxCollateral := 16 * maxCollateralFraction * eligibleMinipools / ratio
		// NOTE: maxCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
		maxCollateral := big.NewInt(0).Mul(_16Eth, r.maxCollateralFraction)
		maxCollateral.Mul(maxCollateral, eligibleMinipoolsBig).Div(maxCollateral, r.rplPrice)

		// Calculate the effective stake
		nodeStake := r.nodeStakes[i]
		if nodeStake.Cmp(minCollateral) == -1 {
			effectiveStakes[i] = big.NewInt(0)
		} else if nodeStake.Cmp(maxCollateral) == 1 {
			effectiveStakes[i] = maxCollateral
		} else {
			effectiveStakes[i] = nodeStake
		}

	}

	return effectiveStakes, nil

}
