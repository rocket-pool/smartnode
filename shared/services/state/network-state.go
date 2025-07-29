package state

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
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

type ValidatorDetailsMap map[types.ValidatorPubkey]beacon.ValidatorStatus

func (vdm ValidatorDetailsMap) MarshalJSON() ([]byte, error) {
	// Marshal as a slice of ValidatorStatus
	out := make([]beacon.ValidatorStatus, 0, len(vdm))
	for _, v := range vdm {
		out = append(out, v)
	}
	return json.Marshal(out)
}

func (vdm *ValidatorDetailsMap) UnmarshalJSON(data []byte) error {
	// Unmarshal as a slice of ValidatorStatus
	var inp []beacon.ValidatorStatus
	err := json.Unmarshal(data, &inp)
	if err != nil {
		return err
	}

	*vdm = make(ValidatorDetailsMap, len(inp))

	// Convert back to a map
	for _, v := range inp {
		// Return an error if the pubkey is already in the map
		if _, exists := (*vdm)[v.Pubkey]; exists {
			return fmt.Errorf("duplicate validator details for pubkey %s", v.Pubkey.Hex())
		}
		(*vdm)[v.Pubkey] = v
	}
	return nil
}

type NetworkState struct {
	// Network version

	// Block / slot for this state
	ElBlockNumber    uint64            `json:"el_block_number"`
	BeaconSlotNumber uint64            `json:"beacon_slot_number"`
	BeaconConfig     beacon.Eth2Config `json:"beacon_config"`

	// Network details
	NetworkDetails *rpstate.NetworkDetails `json:"network_details"`

	// Node details
	NodeDetails []rpstate.NativeNodeDetails `json:"node_details"`
	// NodeDetailsByAddress is an index over NodeDetails and is ignored when marshaling to JSON
	// it is rebuilt when unmarshaling from JSON.
	NodeDetailsByAddress map[common.Address]*rpstate.NativeNodeDetails `json:"-"`

	// Minipool details
	MinipoolDetails []rpstate.NativeMinipoolDetails `json:"minipool_details"`

	// Stores validator details from all megapools
	MegapoolValidatorGlobalIndex []megapool.ValidatorInfoFromGlobalIndex `json:"megapool_validator_global_index"`

	// Map megapool addresses to the pubkeys of its validators
	MegapoolToPubkeysMap map[common.Address][]types.ValidatorPubkey `json:"-"`

	MegapoolDetails map[common.Address]rpstate.NativeMegapoolDetails `json:"-"`

	// These next two fields are indexes over MinipoolDetails and are ignored when marshaling to JSON
	// they are rebuilt when unmarshaling from JSON.
	MinipoolDetailsByAddress map[common.Address]*rpstate.NativeMinipoolDetails   `json:"-"`
	MinipoolDetailsByNode    map[common.Address][]*rpstate.NativeMinipoolDetails `json:"-"`

	// Validator details
	// NetworkState was updated to support megapools, so the old json tag "validator_details" is needed to decode rp-network-state-mainnet-20.json.gz
	MinipoolValidatorDetails ValidatorDetailsMap `json:"validator_details"`
	MegapoolValidatorDetails ValidatorDetailsMap `json:"megapool_validator_details"`

	MegapoolValidatorInfo map[types.ValidatorPubkey]*megapool.ValidatorInfoFromGlobalIndex `json:"-"`

	// Oracle DAO details
	OracleDaoMemberDetails []rpstate.OracleDaoMemberDetails `json:"oracle_dao_member_details"`

	// Protocol DAO proposals
	ProtocolDaoProposalDetails []protocol.ProtocolDaoProposalDetails `json:"protocol_dao_proposal_details,omitempty"`

	IsSaturnDeployed bool
}

func (ns NetworkState) MarshalJSON() ([]byte, error) {
	// No changes needed
	type Alias NetworkState
	a := (*Alias)(&ns)
	return json.Marshal(a)
}

func (ns *NetworkState) UnmarshalJSON(data []byte) error {
	type Alias NetworkState
	var a Alias
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	*ns = NetworkState(a)
	// Rebuild the node details by address index
	ns.NodeDetailsByAddress = make(map[common.Address]*rpstate.NativeNodeDetails)
	for i, details := range ns.NodeDetails {
		if _, ok := ns.NodeDetailsByAddress[details.NodeAddress]; ok {
			return fmt.Errorf("duplicate node details for address %s", details.NodeAddress.Hex())
		}
		// N.B. &details is not the same as &ns.NodeDetails[i]
		// &details is the address of the current element in the loop
		// &ns.NodeDetails[i] is the address of the struct in the slice
		ns.NodeDetailsByAddress[details.NodeAddress] = &ns.NodeDetails[i]
	}

	// Rebuild the minipool details by address index
	ns.MinipoolDetailsByAddress = make(map[common.Address]*rpstate.NativeMinipoolDetails)
	for i, details := range ns.MinipoolDetails {
		if _, ok := ns.MinipoolDetailsByAddress[details.MinipoolAddress]; ok {
			return fmt.Errorf("duplicate minipool details for address %s", details.MinipoolAddress.Hex())
		}

		// N.B. &details is not the same as &ns.MinipoolDetails[i]
		// &details is the address of the current element in the loop
		// &ns.MinipoolDetails[i] is the address of the struct in the slice
		ns.MinipoolDetailsByAddress[details.MinipoolAddress] = &ns.MinipoolDetails[i]
	}

	// Rebuild the minipool details by node index
	ns.MinipoolDetailsByNode = make(map[common.Address][]*rpstate.NativeMinipoolDetails)
	for i, details := range ns.MinipoolDetails {
		// See comments in above loops as to why we're using &ns.MinipoolDetails[i]
		currentDetails := &ns.MinipoolDetails[i]
		nodeList, exists := ns.MinipoolDetailsByNode[details.NodeAddress]
		if !exists {
			ns.MinipoolDetailsByNode[details.NodeAddress] = []*rpstate.NativeMinipoolDetails{currentDetails}
			continue
		}
		// See comments in other loops
		ns.MinipoolDetailsByNode[details.NodeAddress] = append(nodeList, currentDetails)
	}

	return nil
}

// Creates a snapshot of the entire Rocket Pool network state, on both the Execution and Consensus layers
func (m *NetworkStateManager) createNetworkState(slotNumber uint64) (*NetworkState, error) {

	// Get the execution block for the given slot
	beaconBlock, exists, err := m.bc.GetBeaconBlock(fmt.Sprintf("%d", slotNumber))
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

	isSaturnDeployed, err := IsSaturnDeployed(m.rp, opts)
	if err != nil {
		return nil, err
	}
	beaconConfig, err := m.getBeaconConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon config: %w", err)
	}

	// Create the state wrapper
	state := &NetworkState{
		NodeDetailsByAddress:     map[common.Address]*rpstate.NativeNodeDetails{},
		MinipoolDetailsByAddress: map[common.Address]*rpstate.NativeMinipoolDetails{},
		MinipoolDetailsByNode:    map[common.Address][]*rpstate.NativeMinipoolDetails{},
		BeaconSlotNumber:         slotNumber,
		ElBlockNumber:            elBlockNumber,
		BeaconConfig:             *beaconConfig,
		IsSaturnDeployed:         isSaturnDeployed,
	}

	m.logLine("Getting network state for EL block %d, Beacon slot %d", elBlockNumber, slotNumber)
	start := time.Now()

	// Network contracts and details
	contracts, err := rpstate.NewNetworkContracts(m.rp, isSaturnDeployed, m.multicaller, m.balanceBatcher, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting network contracts: %w", err)
	}
	state.NetworkDetails, err = rpstate.NewNetworkDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting network details: %w", err)
	}
	m.logLine("1/6 - Retrieved network details (%s so far)", time.Since(start))

	// Node details
	state.NodeDetails, err = rpstate.GetAllNativeNodeDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting all node details: %w", err)
	}
	m.logLine("2/6 - Retrieved node details (%s so far)", time.Since(start))

	// Minipool details
	state.MinipoolDetails, err = rpstate.GetAllNativeMinipoolDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting all minipool details: %w", err)
	}
	m.logLine("3/6 - Retrieved minipool details (%s so far)", time.Since(start))

	// Create the node lookup
	for i, details := range state.NodeDetails {
		state.NodeDetailsByAddress[details.NodeAddress] = &state.NodeDetails[i]
	}

	// Create the minipool lookups
	pubkeys := make([]types.ValidatorPubkey, 0, len(state.MinipoolDetails))
	emptyPubkey := types.ValidatorPubkey{}
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

	if isSaturnDeployed {
		state.MegapoolValidatorGlobalIndex, err = rpstate.GetAllMegapoolValidators(m.rp, contracts)
		if err != nil {
			return nil, fmt.Errorf("error getting all megapool validator details: %w", err)
		}
		megapoolValidatorPubkeys := make([]types.ValidatorPubkey, 0, len(state.MegapoolValidatorGlobalIndex))
		// Iterate over the megapool validators to add their pubkey to the list of pubkeys
		megapoolAddressMap := make(map[common.Address][]types.ValidatorPubkey)
		megapoolValidatorInfo := make(map[types.ValidatorPubkey]*megapool.ValidatorInfoFromGlobalIndex)
		for _, validator := range state.MegapoolValidatorGlobalIndex {
			// Add the megapool address to a set
			if len(validator.Pubkey) > 0 { // TODO CHECK  validators without a pubkey
				megapoolAddressMap[validator.MegapoolAddress] = append(megapoolAddressMap[validator.MegapoolAddress], types.ValidatorPubkey(validator.Pubkey))
				megapoolValidatorPubkeys = append(megapoolValidatorPubkeys, types.ValidatorPubkey(validator.Pubkey))
				megapoolValidatorInfo[types.ValidatorPubkey(validator.Pubkey)] = &validator
			}
		}
		state.MegapoolToPubkeysMap = megapoolAddressMap
		statusMap, err := m.bc.GetValidatorStatuses(megapoolValidatorPubkeys, &beacon.ValidatorStatusOptions{
			Slot: &slotNumber,
		})
		if err != nil {
			return nil, err
		}
		state.MegapoolValidatorDetails = statusMap
		state.MegapoolValidatorInfo = megapoolValidatorInfo

		// initialize state.MegapoolDetails
		state.MegapoolDetails = make(map[common.Address]rpstate.NativeMegapoolDetails)
		// Sync
		var wg errgroup.Group
		// Iterate the maps and query megapool details
		for megapoolAddress := range megapoolAddressMap {
			megapoolAddress := megapoolAddress
			wg.Go(func() error {

				// Load the megapool
				mp, err := megapool.NewMegaPoolV1(m.rp, megapoolAddress, opts)
				if err != nil {
					return err
				}
				nodeAddress, err := mp.GetNodeAddress(opts)
				if err != nil {
					return err
				}
				megapoolDetails, err := rpstate.GetNodeMegapoolDetails(m.rp, nodeAddress)
				if err != nil {
					return err
				}
				state.MegapoolDetails[megapoolAddress] = megapoolDetails
				return nil
			})
			if err := wg.Wait(); err != nil {
				return nil, fmt.Errorf("error getting all megapool details: %w", err)
			}
		}
	}

	// Calculate avg node fees and distributor shares
	for _, details := range state.NodeDetails {
		details.CalculateAverageFeeAndDistributorShares(state.MinipoolDetailsByNode[details.NodeAddress])
	}

	// Oracle DAO member details
	state.OracleDaoMemberDetails, err = rpstate.GetAllOracleDaoMemberDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting Oracle DAO details: %w", err)
	}
	m.logLine("4/6 - Retrieved Oracle DAO details (%s so far)", time.Since(start))

	// Get the validator stats from Beacon
	statusMap, err := m.bc.GetValidatorStatuses(pubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, err
	}
	state.MinipoolValidatorDetails = statusMap
	m.logLine("5/6 - Retrieved validator details (total time: %s)", time.Since(start))

	// Get the complete node and user shares
	mpds := make([]*rpstate.NativeMinipoolDetails, len(state.MinipoolDetails))
	beaconBalances := make([]*big.Int, len(state.MinipoolDetails))
	for i, mpd := range state.MinipoolDetails {
		mpds[i] = &state.MinipoolDetails[i]
		validator := state.MinipoolValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			beaconBalances[i] = big.NewInt(0)
		} else {
			beaconBalances[i] = eth.GweiToWei(float64(validator.Balance))
		}
	}
	err = rpstate.CalculateCompleteMinipoolShares(m.rp, contracts, mpds, beaconBalances)
	if err != nil {
		return nil, err
	}
	state.MinipoolValidatorDetails = statusMap
	m.logLine("6/6 - Calculated complete node and user balance shares (total time: %s)", time.Since(start))

	return state, nil
}

// Creates a snapshot of the Rocket Pool network, but only for a single node
func (m *NetworkStateManager) createNetworkStateForNode(slotNumber uint64, nodeAddress common.Address) (*NetworkState, error) {
	steps := 5

	// Get the execution block for the given slot
	beaconBlock, exists, err := m.bc.GetBeaconBlock(fmt.Sprintf("%d", slotNumber))
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

	isSaturnDeployed, err := IsSaturnDeployed(m.rp, opts)
	if err != nil {
		return nil, err
	}
	beaconConfig, err := m.getBeaconConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon config: %w", err)
	}

	// Create the state wrapper
	state := &NetworkState{
		NodeDetailsByAddress:     map[common.Address]*rpstate.NativeNodeDetails{},
		MinipoolDetailsByAddress: map[common.Address]*rpstate.NativeMinipoolDetails{},
		MinipoolDetailsByNode:    map[common.Address][]*rpstate.NativeMinipoolDetails{},
		BeaconSlotNumber:         slotNumber,
		ElBlockNumber:            elBlockNumber,
		BeaconConfig:             *beaconConfig,
		IsSaturnDeployed:         isSaturnDeployed,
	}

	m.logLine("Getting network state for EL block %d, Beacon slot %d", elBlockNumber, slotNumber)
	start := time.Now()

	// Network contracts and details
	contracts, err := rpstate.NewNetworkContracts(m.rp, isSaturnDeployed, m.multicaller, m.balanceBatcher, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting network contracts: %w", err)
	}
	state.NetworkDetails, err = rpstate.NewNetworkDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting network details: %w", err)
	}
	m.logLine("1/%d - Retrieved network details (%s so far)", steps, time.Since(start))

	// Node details
	nodeDetails, err := rpstate.GetNativeNodeDetails(m.rp, contracts, nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting node details: %w", err)
	}
	state.NodeDetails = []rpstate.NativeNodeDetails{nodeDetails}
	m.logLine("2/%d - Retrieved node details (%s so far)", steps, time.Since(start))

	// Minipool details
	state.MinipoolDetails, err = rpstate.GetNodeNativeMinipoolDetails(m.rp, contracts, nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting all minipool details: %w", err)
	}
	m.logLine("3/%d - Retrieved minipool details (%s so far)", steps, time.Since(start))

	// Create the node lookup
	for i, details := range state.NodeDetails {
		state.NodeDetailsByAddress[details.NodeAddress] = &state.NodeDetails[i]
	}

	// Create the minipool lookups
	pubkeys := make([]types.ValidatorPubkey, 0, len(state.MinipoolDetails))
	emptyPubkey := types.ValidatorPubkey{}
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
		details.CalculateAverageFeeAndDistributorShares(state.MinipoolDetailsByNode[details.NodeAddress])
	}

	// Get the total network effective RPL stake
	currentStep := 4

	// Get the validator stats from Beacon
	statusMap, err := m.bc.GetValidatorStatuses(pubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, err
	}
	state.MinipoolValidatorDetails = statusMap
	m.logLine("%d/%d - Retrieved validator details (total time: %s)", currentStep, steps, time.Since(start))
	currentStep++

	// Get the complete node and user shares
	mpds := make([]*rpstate.NativeMinipoolDetails, len(state.MinipoolDetails))
	beaconBalances := make([]*big.Int, len(state.MinipoolDetails))
	for i, mpd := range state.MinipoolDetails {
		mpds[i] = &state.MinipoolDetails[i]
		validator := state.MinipoolValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			beaconBalances[i] = big.NewInt(0)
		} else {
			beaconBalances[i] = eth.GweiToWei(float64(validator.Balance))
		}
	}
	err = rpstate.CalculateCompleteMinipoolShares(m.rp, contracts, mpds, beaconBalances)
	if err != nil {
		return nil, err
	}
	state.MinipoolValidatorDetails = statusMap
	m.logLine("%d/%d - Calculated complete node and user balance shares (total time: %s)", currentStep, steps, time.Since(start))
	currentStep++

	// Get the protocol DAO proposals
	state.ProtocolDaoProposalDetails, err = rpstate.GetAllProtocolDaoProposalDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting Protocol DAO proposal details: %w", err)
	}
	m.logLine("%d/%d - Retrieved Protocol DAO proposals (total time: %s)", currentStep, steps, time.Since(start))
	currentStep++

	return state, nil
}

func (s *NetworkState) GetStakedRplValueInEthAndPercentOfBorrowedEth(eligibleBorrowedEth *big.Int, nodeStake *big.Int) (*big.Int, *big.Int) {

	rplPrice := s.NetworkDetails.RplPrice

	// stakedRplValueInEth := nodeStake * ratio / 1 Eth
	stakedRplValueInEth := big.NewInt(0)
	stakedRplValueInEth.Mul(nodeStake, rplPrice)
	stakedRplValueInEth.Quo(stakedRplValueInEth, oneEth)

	// Avoid division by zero
	if eligibleBorrowedEth.Sign() == 0 {
		return stakedRplValueInEth, big.NewInt(0)
	}

	// percentOfBorrowedEth := stakedRplValueInEth * 100 Eth / eligibleBorrowedEth
	percentOfBorrowedEth := big.NewInt(0)
	percentOfBorrowedEth.Mul(stakedRplValueInEth, oneHundredEth)
	percentOfBorrowedEth.Quo(percentOfBorrowedEth, eligibleBorrowedEth)

	return stakedRplValueInEth, percentOfBorrowedEth
}

func (s *NetworkState) GetNodeWeight(eligibleBorrowedEth *big.Int, nodeStake *big.Int) *big.Int {
	stakedRplValueInEth, percentOfBorrowedEth := s.GetStakedRplValueInEthAndPercentOfBorrowedEth(eligibleBorrowedEth, nodeStake)

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
	intervalEndEpoch := s.BeaconSlotNumber / s.BeaconConfig.SlotsPerEpoch

	for _, mpd := range s.MinipoolDetailsByNode[node.NodeAddress] {

		// It must exist and be staking
		if !mpd.Exists || mpd.Status != types.Staking {
			continue
		}

		// Doesn't exist on Beacon yet
		validatorStatus, exists := s.MinipoolValidatorDetails[mpd.Pubkey]
		if !exists {
			//s.logLine("NOTE: minipool %s (pubkey %s) didn't exist, ignoring it in effective RPL calculation", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex())
			continue
		}

		// Already exited
		if validatorStatus.ExitEpoch <= intervalEndEpoch {
			//s.logLine("NOTE: Minipool %s exited on epoch %d which is not after interval epoch %d so it's not eligible for RPL rewards", mpd.MinipoolAddress.Hex(), validatorStatus.ExitEpoch, intervalEndEpoch)
			continue
		}

		// It's eligible, so add up the borrowed and bonded amounts
		eligibleBorrowedEth.Add(eligibleBorrowedEth, mpd.UserDepositBalance)
	}

	if node.MegapoolDeployed {
		megapool := s.MegapoolDetails[node.MegapoolAddress]
		validators := s.MegapoolToPubkeysMap[node.MegapoolAddress]
		activeValidators := 0
		for _, validator := range validators {
			validatorStatus, exists := s.MegapoolValidatorDetails[validator]
			if !exists {
				continue
			}

			if validatorStatus.ExitEpoch <= intervalEndEpoch {
				continue
			}

			activeValidators += 1
		}
		totalValidators := megapool.ValidatorCount
		userCapital := big.NewInt(0).Mul(megapool.UserCapital, big.NewInt(int64(activeValidators)))
		// Scale the userCapital by active validators / total validators
		userCapital.Mul(userCapital, big.NewInt(int64(activeValidators)))
		userCapital.Quo(userCapital, big.NewInt(int64(totalValidators)))
		eligibleBorrowedEth.Add(eligibleBorrowedEth, userCapital)
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
				if mpd.Exists && mpd.Status == types.Staking {
					// Doesn't exist on Beacon yet
					validatorStatus, exists := s.MinipoolValidatorDetails[mpd.Pubkey]
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
