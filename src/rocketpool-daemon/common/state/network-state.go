package state

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	rpstate "github.com/rocket-pool/rocketpool-go/v2/utils/state"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	"golang.org/x/sync/errgroup"
)

const (
	threadLimit int = 6
)

var two = big.NewInt(2)
var oneHundred = big.NewInt(100)

var oneEth = big.NewInt(1e18)
var oneHundredEth = big.NewInt(0).Mul(oneHundred, oneEth)
var fifteenEth = big.NewInt(0).Mul(big.NewInt(15), oneEth)
var _13_6137_Eth = big.NewInt(0).Mul(big.NewInt(136137), big.NewInt(1e14))
var _13_Eth = big.NewInt(0).Mul(big.NewInt(13), oneEth)

type NetworkState struct {
	// Block / slot for this state
	ElBlockNumber    uint64
	BeaconSlotNumber uint64
	BeaconConfig     beacon.Eth2Config

	// Network details
	NetworkDetails *rpstate.NetworkDetails

	// Node details
	NodeDetails          []rpstate.NativeNodeDetails
	NodeDetailsByAddress map[common.Address]*rpstate.NativeNodeDetails

	// Minipool details
	MinipoolDetails          []rpstate.NativeMinipoolDetails
	MinipoolDetailsByAddress map[common.Address]*rpstate.NativeMinipoolDetails
	MinipoolDetailsByNode    map[common.Address][]*rpstate.NativeMinipoolDetails

	// Validator details
	ValidatorDetails map[beacon.ValidatorPubkey]beacon.ValidatorStatus

	// Oracle DAO details
	OracleDaoMemberDetails []rpstate.OracleDaoMemberDetails

	// Protocol DAO proposals
	ProtocolDaoProposalDetails []*protocol.ProtocolDaoProposal

	// Internal fields
	logger *slog.Logger
}

// Creates a snapshot of the entire Rocket Pool network state, on both the Execution and Consensus layers
func CreateNetworkState(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, ec eth.IExecutionClient, bc beacon.IBeaconClient, logger *slog.Logger, slotNumber uint64, beaconConfig beacon.Eth2Config, context context.Context) (*NetworkState, error) {
	// Get the relevant network contracts
	resources := cfg.GetNetworkResources()
	multicallerAddress := resources.MulticallAddress
	balanceBatcherAddress := resources.BalanceBatcherAddress

	// Get the execution block for the given slot
	beaconBlock, exists, err := bc.GetBeaconBlock(context, fmt.Sprintf("%d", slotNumber))
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon block for slot %d: %w", slotNumber, err)
	}
	if !exists {
		return nil, fmt.Errorf("slot %d did not have a Beacon block", slotNumber)
	}

	// Get the corresponding block on the EL
	elBlockNumber := beaconBlock.ExecutionBlockNumber
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlockNumber),
	}

	// Create the state wrapper
	state := &NetworkState{
		NodeDetailsByAddress:     map[common.Address]*rpstate.NativeNodeDetails{},
		MinipoolDetailsByAddress: map[common.Address]*rpstate.NativeMinipoolDetails{},
		MinipoolDetailsByNode:    map[common.Address][]*rpstate.NativeMinipoolDetails{},
		BeaconSlotNumber:         slotNumber,
		ElBlockNumber:            elBlockNumber,
		BeaconConfig:             beaconConfig,
		logger:                   logger,
	}

	logger.Info("Getting network state...", slog.Uint64(keys.BlockKey, elBlockNumber), slog.Uint64(keys.SlotKey, slotNumber))
	start := time.Now()

	// Network contracts and details
	contracts, err := rpstate.NewNetworkContracts(rp, multicallerAddress, balanceBatcherAddress, true, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting network contracts: %w", err)
	}
	state.NetworkDetails, err = rpstate.NewNetworkDetails(rp, contracts, true)
	if err != nil {
		return nil, fmt.Errorf("error getting network details: %w", err)
	}
	logger.Info("1/6 - Retrieved network details", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Node details
	state.NodeDetails, err = rpstate.GetAllNativeNodeDetails(rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting all node details: %w", err)
	}
	logger.Info("2/6 - Retrieved node details", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Minipool details
	state.MinipoolDetails, err = rpstate.GetAllNativeMinipoolDetails(rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting all minipool details: %w", err)
	}
	logger.Info("3/6 - Retrieved minipool details", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Create the node lookup
	for i, details := range state.NodeDetails {
		state.NodeDetailsByAddress[details.NodeAddress] = &state.NodeDetails[i]
	}

	// Create the minipool lookups
	pubkeys := make([]beacon.ValidatorPubkey, 0, len(state.MinipoolDetails))
	emptyPubkey := beacon.ValidatorPubkey{}
	for i, details := range state.MinipoolDetails {
		state.MinipoolDetailsByAddress[details.MinipoolAddress] = &state.MinipoolDetails[i]
		if details.Pubkey != emptyPubkey {
			pubkeys = append(pubkeys, details.Pubkey)
		}

		// The map of nodes to minipools
		nodeList, exists := state.MinipoolDetailsByNode[details.NodeAddress]
		if !exists {
			nodeList = []*rpstate.NativeMinipoolDetails{}
		}
		nodeList = append(nodeList, &state.MinipoolDetails[i])
		state.MinipoolDetailsByNode[details.NodeAddress] = nodeList
	}

	// Calculate avg node fees and distributor shares
	for _, details := range state.NodeDetails {
		rpstate.CalculateAverageFeeAndDistributorShares(rp, contracts, details, state.MinipoolDetailsByNode[details.NodeAddress])
	}

	// Oracle DAO member details
	state.OracleDaoMemberDetails, err = rpstate.GetAllOracleDaoMemberDetails(rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting Oracle DAO details: %w", err)
	}
	logger.Info("4/6 - Retrieved Oracle DAO details", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Get the validator stats from Beacon
	statusMap, err := bc.GetValidatorStatuses(context, pubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, err
	}
	state.ValidatorDetails = statusMap
	logger.Info("5/6 - Retrieved validator details", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Get the complete node and user shares
	mpds := make([]*rpstate.NativeMinipoolDetails, len(state.MinipoolDetails))
	beaconBalances := make([]*big.Int, len(state.MinipoolDetails))
	for i, mpd := range state.MinipoolDetails {
		mpds[i] = &state.MinipoolDetails[i]
		validator := state.ValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			beaconBalances[i] = big.NewInt(0)
		} else {
			beaconBalances[i] = eth.GweiToWei(float64(validator.Balance))
		}
	}
	err = rpstate.CalculateCompleteMinipoolShares(rp, contracts, mpds, beaconBalances)
	if err != nil {
		return nil, err
	}
	state.ValidatorDetails = statusMap
	logger.Info("6/6 - Calculated complete node and user balance shares", slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	return state, nil
}

// Creates a snapshot of the Rocket Pool network, but only for a single node
// Also gets the total effective RPL stake of the network for convenience since this is required by several node routines
func CreateNetworkStateForNode(cfg *config.SmartNodeConfig, rp *rocketpool.RocketPool, ec eth.IExecutionClient, bc beacon.IBeaconClient, logger *slog.Logger, slotNumber uint64, beaconConfig beacon.Eth2Config, nodeAddress common.Address, calculateTotalEffectiveStake bool, context context.Context) (*NetworkState, *big.Int, error) {
	steps := 6
	if calculateTotalEffectiveStake {
		steps++
	}

	// Get the relevant network contracts
	resources := cfg.GetNetworkResources()
	multicallerAddress := resources.MulticallAddress
	balanceBatcherAddress := resources.BalanceBatcherAddress

	// Get the execution block for the given slot
	beaconBlock, exists, err := bc.GetBeaconBlock(context, fmt.Sprintf("%d", slotNumber))
	if err != nil {
		return nil, nil, fmt.Errorf("error getting Beacon block for slot %d: %w", slotNumber, err)
	}
	if !exists {
		return nil, nil, fmt.Errorf("slot %d did not have a Beacon block", slotNumber)
	}

	// Get the corresponding block on the EL
	elBlockNumber := beaconBlock.ExecutionBlockNumber
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlockNumber),
	}

	// Create the state wrapper
	state := &NetworkState{
		NodeDetailsByAddress:     map[common.Address]*rpstate.NativeNodeDetails{},
		MinipoolDetailsByAddress: map[common.Address]*rpstate.NativeMinipoolDetails{},
		MinipoolDetailsByNode:    map[common.Address][]*rpstate.NativeMinipoolDetails{},
		BeaconSlotNumber:         slotNumber,
		ElBlockNumber:            elBlockNumber,
		BeaconConfig:             beaconConfig,
		logger:                   logger,
	}

	logger.Info("Getting network state...", slog.Uint64(keys.BlockKey, elBlockNumber), slog.Uint64(keys.SlotKey, slotNumber))
	start := time.Now()

	// Network contracts and details
	contracts, err := rpstate.NewNetworkContracts(rp, multicallerAddress, balanceBatcherAddress, true, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting network contracts: %w", err)
	}
	state.NetworkDetails, err = rpstate.NewNetworkDetails(rp, contracts, true)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting network details: %w", err)
	}
	logger.Info(fmt.Sprintf("1/%d - Retrieved network details", steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Node details
	nodeDetails, err := rpstate.GetNativeNodeDetails(rp, contracts, nodeAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting node details: %w", err)
	}
	state.NodeDetails = []rpstate.NativeNodeDetails{nodeDetails}
	logger.Info(fmt.Sprintf("2/%d - Retrieved node details", steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Minipool details
	state.MinipoolDetails, err = rpstate.GetNodeNativeMinipoolDetails(rp, contracts, nodeAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting all minipool details: %w", err)
	}
	logger.Info(fmt.Sprintf("3/%d - Retrieved minipool details", steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))

	// Create the node lookup
	for i, details := range state.NodeDetails {
		state.NodeDetailsByAddress[details.NodeAddress] = &state.NodeDetails[i]
	}

	// Create the minipool lookups
	pubkeys := make([]beacon.ValidatorPubkey, 0, len(state.MinipoolDetails))
	emptyPubkey := beacon.ValidatorPubkey{}
	for i, details := range state.MinipoolDetails {
		state.MinipoolDetailsByAddress[details.MinipoolAddress] = &state.MinipoolDetails[i]
		if details.Pubkey != emptyPubkey {
			pubkeys = append(pubkeys, details.Pubkey)
		}

		// The map of nodes to minipools
		nodeList, exists := state.MinipoolDetailsByNode[details.NodeAddress]
		if !exists {
			nodeList = []*rpstate.NativeMinipoolDetails{}
		}
		nodeList = append(nodeList, &state.MinipoolDetails[i])
		state.MinipoolDetailsByNode[details.NodeAddress] = nodeList
	}

	// Calculate avg node fees and distributor shares
	for _, details := range state.NodeDetails {
		rpstate.CalculateAverageFeeAndDistributorShares(rp, contracts, details, state.MinipoolDetailsByNode[details.NodeAddress])
	}

	// Get the total network effective RPL stake
	currentStep := 4
	var totalEffectiveStake *big.Int
	if calculateTotalEffectiveStake {
		totalEffectiveStake, err = rpstate.GetTotalEffectiveRplStake(rp, contracts)
		if err != nil {
			return nil, nil, fmt.Errorf("error calculating total effective RPL stake for the network: %w", err)
		}
		logger.Info(fmt.Sprintf("%d/%d - Calculated total effective stake", currentStep, steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))
		currentStep++
	}

	// Get the validator stats from Beacon
	statusMap, err := bc.GetValidatorStatuses(context, pubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, nil, err
	}
	state.ValidatorDetails = statusMap
	logger.Info(fmt.Sprintf("%d/%d - Retrieved validator details", currentStep, steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))
	currentStep++

	// Get the complete node and user shares
	mpds := make([]*rpstate.NativeMinipoolDetails, len(state.MinipoolDetails))
	beaconBalances := make([]*big.Int, len(state.MinipoolDetails))
	for i, mpd := range state.MinipoolDetails {
		mpds[i] = &state.MinipoolDetails[i]
		validator := state.ValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			beaconBalances[i] = big.NewInt(0)
		} else {
			beaconBalances[i] = eth.GweiToWei(float64(validator.Balance))
		}
	}
	err = rpstate.CalculateCompleteMinipoolShares(rp, contracts, mpds, beaconBalances)
	if err != nil {
		return nil, nil, err
	}
	state.ValidatorDetails = statusMap
	logger.Info(fmt.Sprintf("%d/%d - Calculated complete node and user balance shares", currentStep, steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))
	currentStep++

	// Get the protocol DAO proposals
	state.ProtocolDaoProposalDetails, err = rpstate.GetAllProtocolDaoProposalDetails(rp, contracts)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting Protocol DAO proposal details: %w", err)
	}
	logger.Info(fmt.Sprintf("%d/%d - Retrieved Protocol DAO proposals", currentStep, steps), slog.Duration(keys.TotalElapsedKey, time.Since(start)))
	currentStep++

	return state, totalEffectiveStake, nil
}

func (s *NetworkState) GetNodeWeight(eligibleBorrowedEth *big.Int, nodeStake *big.Int) *big.Int {
	rplPrice := s.NetworkDetails.RplPrice

	// stakedRplValueInEth := nodeStake * ratio / 1 Eth
	stakedRplValueInEth := big.NewInt(0)
	stakedRplValueInEth.Mul(nodeStake, rplPrice)
	stakedRplValueInEth.Quo(stakedRplValueInEth, oneEth)

	// percentOfBorrowedEth := stakedRplValueInEth * 100 Eth / eligibleBorrowedEth
	percentOfBorrowedEth := big.NewInt(0)
	percentOfBorrowedEth.Mul(stakedRplValueInEth, oneHundredEth)
	percentOfBorrowedEth.Quo(percentOfBorrowedEth, eligibleBorrowedEth)

	// If at or under 15%, return 100 * stakedRplValueInEth
	if percentOfBorrowedEth.Cmp(fifteenEth) <= 0 {
		stakedRplValueInEth.Mul(stakedRplValueInEth, oneHundred)
		return stakedRplValueInEth
	}

	// Otherwise, return ((13.6137 Eth + 2 * ln(percentOfBorrowedEth - 13 Eth)) * eligibleBorrowedEth) / 1 Eth
	lnArgs := big.NewInt(0).Sub(percentOfBorrowedEth, _13_Eth)
	return big.NewInt(0).Quo(
		big.NewInt(0).Mul(
			big.NewInt(0).Add(
				_13_6137_Eth,
				big.NewInt(0).Mul(
					two,
					ethNaturalLog(lnArgs),
				),
			),
			eligibleBorrowedEth,
		),
		oneEth,
	)
}

// Starting in v8, RPL stake is phased out and replaced with weight.
// scaleByParticipation and allowRplForUnstartedValidators are hard-coded true here, since
// only v8 cares about weight.
func (s *NetworkState) CalculateNodeWeights() (map[common.Address]*big.Int, *big.Int, error) {
	weights := make(map[common.Address]*big.Int, len(s.NodeDetails))
	totalWeight := big.NewInt(0)
	intervalDurationBig := big.NewInt(int64(s.NetworkDetails.IntervalDuration.Seconds()))
	genesisTime := time.Unix(int64(s.BeaconConfig.GenesisTime), 0)
	slotOffset := time.Duration(s.BeaconSlotNumber*s.BeaconConfig.SecondsPerSlot) * time.Second
	slotTime := genesisTime.Add(slotOffset)

	nodeCount := uint64(len(s.NodeDetails))
	weightSlice := make([]*big.Int, nodeCount)

	// Get the weight for each node
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	for i, node := range s.NodeDetails {
		i := i
		node := node
		wg.Go(func() error {
			eligibleBorrowedEth := s.GetEligibleBorrowedEth(&node)

			// minCollateral := borrowedEth * minCollateralFraction / ratio
			// NOTE: minCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
			minCollateral := big.NewInt(0).Mul(eligibleBorrowedEth, s.NetworkDetails.MinCollateralFraction)
			minCollateral.Div(minCollateral, s.NetworkDetails.RplPrice)

			// Calculate the weight
			nodeWeight := big.NewInt(0)
			if node.RplStake.Cmp(minCollateral) == -1 || eligibleBorrowedEth.Sign() <= 0 {
				weightSlice[i] = nodeWeight
				return nil
			}

			nodeWeight.Set(s.GetNodeWeight(eligibleBorrowedEth, node.RplStake))

			// Scale the node weight by the participation in the current interval
			// Get the timestamp of the node's registration
			regTimeBig := node.RegistrationTime
			regTime := time.Unix(regTimeBig.Int64(), 0)

			// Get the actual node weight, scaled based on participation
			eligibleDuration := slotTime.Sub(regTime)
			if eligibleDuration < s.NetworkDetails.IntervalDuration {
				eligibleSeconds := big.NewInt(int64(eligibleDuration / time.Second))
				nodeWeight.Mul(nodeWeight, eligibleSeconds)
				nodeWeight.Div(nodeWeight, intervalDurationBig)
			}

			weightSlice[i] = nodeWeight
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, nil, err
	}

	// Tally everything up and make the node stake map
	for i, nodeWeight := range weightSlice {
		node := s.NodeDetails[i]
		weights[node.NodeAddress] = nodeWeight
		totalWeight.Add(totalWeight, nodeWeight)
	}

	return weights, totalWeight, nil
}

func (s *NetworkState) GetEligibleBorrowedEth(node *rpstate.NativeNodeDetails) *big.Int {
	eligibleBorrowedEth := big.NewInt(0)

	for _, mpd := range s.MinipoolDetailsByNode[node.NodeAddress] {

		// It must exist and be staking
		if !mpd.Exists || mpd.Status != types.MinipoolStatus_Staking {
			continue
		}

		// Doesn't exist on Beacon yet
		validatorStatus, exists := s.ValidatorDetails[mpd.Pubkey]
		if !exists {
			//s.logLine("NOTE: minipool %s (pubkey %s) didn't exist, ignoring it in effective RPL calculation", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex())
			continue
		}

		intervalEndEpoch := s.BeaconSlotNumber / s.BeaconConfig.SlotsPerEpoch

		// Already exited
		if validatorStatus.ExitEpoch <= intervalEndEpoch {
			//s.logLine("NOTE: Minipool %s exited on epoch %d which is not after interval epoch %d so it's not eligible for RPL rewards", mpd.MinipoolAddress.Hex(), validatorStatus.ExitEpoch, intervalEndEpoch)
			continue
		}

		// It's eligible, so add up the borrowed and bonded amounts
		eligibleBorrowedEth.Add(eligibleBorrowedEth, mpd.UserDepositBalance)
	}
	return eligibleBorrowedEth
}

// Calculate the true effective stakes of all nodes in the state, using the validator status
// on Beacon as a reference for minipool eligibility instead of the EL-based minipool status
func (s *NetworkState) CalculateTrueEffectiveStakes(scaleByParticipation bool, allowRplForUnstartedValidators bool) (map[common.Address]*big.Int, *big.Int, error) {
	effectiveStakes := make(map[common.Address]*big.Int, len(s.NodeDetails))
	totalEffectiveStake := big.NewInt(0)
	intervalDurationBig := big.NewInt(int64(s.NetworkDetails.IntervalDuration.Seconds()))
	genesisTime := time.Unix(int64(s.BeaconConfig.GenesisTime), 0)
	slotOffset := time.Duration(s.BeaconSlotNumber*s.BeaconConfig.SecondsPerSlot) * time.Second
	slotTime := genesisTime.Add(slotOffset)

	nodeCount := uint64(len(s.NodeDetails))
	effectiveStakeSlice := make([]*big.Int, nodeCount)

	// Get the effective stake for each node
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	for i, node := range s.NodeDetails {
		i := i
		node := node
		wg.Go(func() error {
			eligibleBorrowedEth := big.NewInt(0)
			eligibleBondedEth := big.NewInt(0)
			for _, mpd := range s.MinipoolDetailsByNode[node.NodeAddress] {
				// It must exist and be staking
				if mpd.Exists && mpd.Status == types.MinipoolStatus_Staking {
					// Doesn't exist on Beacon yet
					validatorStatus, exists := s.ValidatorDetails[mpd.Pubkey]
					if !exists {
						//s.logLine("NOTE: minipool %s (pubkey %s) didn't exist, ignoring it in effective RPL calculation", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex())
						continue
					}

					intervalEndEpoch := s.BeaconSlotNumber / s.BeaconConfig.SlotsPerEpoch
					if !allowRplForUnstartedValidators {
						// Starts too late
						if validatorStatus.ActivationEpoch > intervalEndEpoch {
							//s.logLine("NOTE: Minipool %s starts on epoch %d which is after interval epoch %d so it's not eligible for RPL rewards", mpd.MinipoolAddress.Hex(), validatorStatus.ActivationEpoch, intervalEndEpoch)
							continue
						}

					}
					// Already exited
					if validatorStatus.ExitEpoch <= intervalEndEpoch {
						//s.logLine("NOTE: Minipool %s exited on epoch %d which is not after interval epoch %d so it's not eligible for RPL rewards", mpd.MinipoolAddress.Hex(), validatorStatus.ExitEpoch, intervalEndEpoch)
						continue
					}
					// It's eligible, so add up the borrowed and bonded amounts
					eligibleBorrowedEth.Add(eligibleBorrowedEth, mpd.UserDepositBalance)
					eligibleBondedEth.Add(eligibleBondedEth, mpd.NodeDepositBalance)
				}
			}

			// minCollateral := borrowedEth * minCollateralFraction / ratio
			// NOTE: minCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
			minCollateral := big.NewInt(0).Mul(eligibleBorrowedEth, s.NetworkDetails.MinCollateralFraction)
			minCollateral.Div(minCollateral, s.NetworkDetails.RplPrice)

			// maxCollateral := bondedEth * maxCollateralFraction / ratio
			// NOTE: maxCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
			maxCollateral := big.NewInt(0).Mul(eligibleBondedEth, s.NetworkDetails.MaxCollateralFraction)
			maxCollateral.Div(maxCollateral, s.NetworkDetails.RplPrice)

			// Calculate the effective stake
			nodeStake := big.NewInt(0).Set(node.RplStake)
			if nodeStake.Cmp(minCollateral) == -1 {
				// Under min collateral
				nodeStake.SetUint64(0)
			} else if nodeStake.Cmp(maxCollateral) == 1 {
				// Over max collateral
				nodeStake.Set(maxCollateral)
			}

			// Scale the effective stake by the participation in the current interval
			if scaleByParticipation {
				// Get the timestamp of the node's registration
				regTimeBig := node.RegistrationTime
				regTime := time.Unix(regTimeBig.Int64(), 0)

				// Get the actual effective stake, scaled based on participation
				eligibleDuration := slotTime.Sub(regTime)
				if eligibleDuration < s.NetworkDetails.IntervalDuration {
					eligibleSeconds := big.NewInt(int64(eligibleDuration / time.Second))
					nodeStake.Mul(nodeStake, eligibleSeconds)
					nodeStake.Div(nodeStake, intervalDurationBig)
				}
			}

			effectiveStakeSlice[i] = nodeStake
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, nil, err
	}

	// Tally everything up and make the node stake map
	for i, nodeStake := range effectiveStakeSlice {
		node := s.NodeDetails[i]
		effectiveStakes[node.NodeAddress] = nodeStake
		totalEffectiveStake.Add(totalEffectiveStake, nodeStake)
	}

	return effectiveStakes, totalEffectiveStake, nil

}

// Returns the index of the Most Significant Bit of n, or UINT_MAX if the input is 0
// The index of the Least Significant Bit is 0.
func indexOfMSB(n *big.Int) uint {
	copyN := big.NewInt(0).Set(n)
	var out uint
	for copyN.Cmp(big.NewInt(0)) > 0 {
		copyN.Rsh(copyN, 1)
		out++
	}

	// 0-index
	return out - 1
}

func log2(x *big.Int) *big.Int {
	out := big.NewInt(0)

	// Calculate the integer part of the logarithm
	copyX := big.NewInt(0).Set(x)
	copyX.Quo(x, oneEth)
	// The input is always over 2 Eth, so we do not need to worry about
	// overflowing indexOfMSB
	n := indexOfMSB(copyX)

	// Add integer part of the logarithm
	out.Mul(oneEth, big.NewInt(int64(n)))

	// Calculate y = x * 2**-n
	y := big.NewInt(0).Rsh(big.NewInt(0).Set(x), n)

	// If y is the unit number, the fractional part is zero.
	if y.Cmp(oneEth) == 0 {
		return out
	}

	doubleUnit := big.NewInt(0).Mul(big.NewInt(2), oneEth)
	delta := big.NewInt(0).Rsh(oneEth, 1)
	for i := 0; i < 60; i++ {
		y.Mul(y, y)
		y.Quo(y, oneEth)

		if y.Cmp(doubleUnit) >= 0 {
			out.Add(out, delta)
			y.Rsh(y, 1)
		}

		delta.Rsh(delta, 1)
	}

	return out
}

func ethNaturalLog(x *big.Int) *big.Int {
	log2e := big.NewInt(1_442695040888963407)
	log2x := log2(x)

	numerator := big.NewInt(0).Mul(oneEth, log2x)
	return numerator.Quo(numerator, log2e)
}
