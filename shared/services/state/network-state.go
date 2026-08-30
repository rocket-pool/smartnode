package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
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

// MegapoolValidatorKey uniquely identifies a megapool validator.
type MegapoolValidatorKey struct {
	MegapoolAddress common.Address
	Pubkey          types.ValidatorPubkey
}

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

	// Minipool details
	MinipoolDetails []rpstate.NativeMinipoolDetails `json:"minipool_details"`

	// Stores validator details from all megapools
	// The json tag "megapool_validator_global_index" is kept for backwards compatibility
	MegapoolValidators []megapool.MegapoolValidatorInfo                 `json:"megapool_validator_global_index"`
	MegapoolDetails    map[common.Address]rpstate.NativeMegapoolDetails `json:"megapool_details"`

	// Validator details
	// NetworkState was updated to support megapools, so the old json tag "validator_details" is needed to decode rp-network-state-mainnet-20.json.gz
	MinipoolValidatorDetails ValidatorDetailsMap `json:"validator_details"`
	MegapoolValidatorDetails ValidatorDetailsMap `json:"megapool_validator_details"`

	// Oracle DAO details
	OracleDaoMemberDetails []rpstate.OracleDaoMemberDetails `json:"oracle_dao_member_details"`

	// Protocol DAO proposals
	ProtocolDaoProposalDetails []protocol.ProtocolDaoProposalDetails `json:"protocol_dao_proposal_details,omitempty"`
}

func (s NetworkState) MarshalJSON() ([]byte, error) {
	// No changes needed
	type Alias NetworkState
	a := (*Alias)(&s)
	return json.Marshal(a)
}

func (s *NetworkState) UnmarshalJSON(data []byte) error {
	type Alias NetworkState
	var a Alias
	err := json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	*s = NetworkState(a)
	return s.Validate()
}

type NetworkStateIndex struct {
	*NetworkState
	NodeDetailsByAddress     map[common.Address]*rpstate.NativeNodeDetails
	MegapoolToPubkeysMap     map[common.Address][]types.ValidatorPubkey
	MinipoolDetailsByAddress map[common.Address]*rpstate.NativeMinipoolDetails
	MinipoolDetailsByNode    map[common.Address][]*rpstate.NativeMinipoolDetails
	MegapoolValidatorInfo    map[MegapoolValidatorKey]*megapool.MegapoolValidatorInfo
	NodeFeeDetailsByAddress  map[common.Address]*rpstate.NodeFeeDetails
}

func (s *NetworkState) ToIndexedNetworkState() *NetworkStateIndex {
	out := &NetworkStateIndex{
		NetworkState: s,
	}
	// Rebuild the node details by address index
	out.NodeDetailsByAddress = make(map[common.Address]*rpstate.NativeNodeDetails)
	for i, details := range s.NodeDetails {
		// N.B. &details is not the same as &s.NodeDetails[i]
		// &details is the address of the current element in the loop
		// &s.NodeDetails[i] is the address of the struct in the slice
		out.NodeDetailsByAddress[details.NodeAddress] = &s.NodeDetails[i]
	}

	// Rebuild the minipool details by address index
	out.MinipoolDetailsByAddress = make(map[common.Address]*rpstate.NativeMinipoolDetails)
	for i, details := range s.MinipoolDetails {
		// N.B. &details is not the same as &s.MinipoolDetails[i]
		// &details is the address of the current element in the loop
		// &s.MinipoolDetails[i] is the address of the struct in the slice
		out.MinipoolDetailsByAddress[details.MinipoolAddress] = &s.MinipoolDetails[i]
	}

	// Rebuild the minipool details by node index
	out.MinipoolDetailsByNode = make(map[common.Address][]*rpstate.NativeMinipoolDetails)
	for i, details := range s.MinipoolDetails {
		// See comments in above loops as to why we're using &s.MinipoolDetails[i]
		currentDetails := &s.MinipoolDetails[i]
		nodeList, exists := out.MinipoolDetailsByNode[details.NodeAddress]
		if !exists {
			out.MinipoolDetailsByNode[details.NodeAddress] = []*rpstate.NativeMinipoolDetails{currentDetails}
			continue
		}
		// See comments in other loops
		out.MinipoolDetailsByNode[details.NodeAddress] = append(nodeList, currentDetails)
	}

	out.MegapoolToPubkeysMap = make(map[common.Address][]types.ValidatorPubkey)
	out.MegapoolValidatorInfo = make(map[MegapoolValidatorKey]*megapool.MegapoolValidatorInfo)
	for i := range s.MegapoolValidators {
		validator := &s.MegapoolValidators[i]
		if len(validator.Pubkey) > 0 {
			pubkey := types.ValidatorPubkey(validator.Pubkey)
			out.MegapoolToPubkeysMap[validator.MegapoolAddress] = append(
				out.MegapoolToPubkeysMap[validator.MegapoolAddress], pubkey,
			)
			out.MegapoolValidatorInfo[MegapoolValidatorKey{MegapoolAddress: validator.MegapoolAddress, Pubkey: pubkey}] = validator
		}
	}

	// Calculate avg node fees and distributor shares
	out.NodeFeeDetailsByAddress = make(map[common.Address]*rpstate.NodeFeeDetails)
	for _, details := range s.NodeDetails {
		out.NodeFeeDetailsByAddress[details.NodeAddress] = &rpstate.NodeFeeDetails{
			DistributorBalanceNodeETH: big.NewInt(0),
			DistributorBalanceUserETH: big.NewInt(0),
			AverageNodeFee:            big.NewInt(0),
		}
		out.NodeFeeDetailsByAddress[details.NodeAddress].CalculateAverageFeeAndDistributorShares(&details, out.MinipoolDetailsByNode[details.NodeAddress])
	}

	return out
}

func (s NetworkStateIndex) MarshalJSON() ([]byte, error) {
	return s.NetworkState.MarshalJSON()
}

func (s *NetworkStateIndex) UnmarshalJSON(data []byte) error {
	var inner NetworkState
	err := json.Unmarshal(data, &inner)
	if err != nil {
		return err
	}

	*s = *inner.ToIndexedNetworkState()

	return nil
}

func (s *NetworkState) GetUniqueMegapoolPubkeys() []types.ValidatorPubkey {
	pubkeys := make([]types.ValidatorPubkey, 0, len(s.MegapoolValidators))
	seen := make(map[types.ValidatorPubkey]bool, len(s.MegapoolValidators))
	for _, validator := range s.MegapoolValidators {
		if len(validator.Pubkey) > 0 {
			pubkey := types.ValidatorPubkey(validator.Pubkey)
			if !seen[pubkey] {
				seen[pubkey] = true
				pubkeys = append(pubkeys, pubkey)
			}
		}
	}
	return pubkeys
}

func (s *NetworkState) GetMinipoolPubkeys() []types.ValidatorPubkey {
	pubkeys := make([]types.ValidatorPubkey, 0, len(s.MinipoolDetails))
	emptyPubkey := types.ValidatorPubkey{}
	for _, mpd := range s.MinipoolDetails {
		if !bytes.Equal(mpd.Pubkey[:], emptyPubkey[:]) {
			pubkeys = append(pubkeys, mpd.Pubkey)
		}
	}
	return pubkeys
}

func (s *NetworkState) getMegapoolAddresses() []common.Address {
	seen := make(map[common.Address]bool)
	addresses := make([]common.Address, 0, len(s.MegapoolValidators))
	for _, megapool := range s.MegapoolValidators {
		if len(megapool.Pubkey) == 0 {
			continue
		}
		if seen[megapool.MegapoolAddress] {
			continue
		}
		seen[megapool.MegapoolAddress] = true
		addresses = append(addresses, megapool.MegapoolAddress)
	}
	return addresses
}

func (s *NetworkState) Validate() error {

	// Check for duplicate node details
	seenNodes := make(map[common.Address]bool)
	for _, node := range s.NodeDetails {
		if seenNodes[node.NodeAddress] {
			return fmt.Errorf("duplicate node details for address %s", node.NodeAddress.Hex())
		}
		seenNodes[node.NodeAddress] = true
	}

	// Check for duplicate minipool details
	seenMinipools := make(map[common.Address]bool)
	for _, mpd := range s.MinipoolDetails {
		if seenMinipools[mpd.MinipoolAddress] {
			return fmt.Errorf("duplicate minipool details for address %s", mpd.MinipoolAddress.Hex())
		}
		seenMinipools[mpd.MinipoolAddress] = true
	}

	return nil
}

// Returns the validator info for the given megapool and pubkey.
func (s *NetworkStateIndex) GetMegapoolValidatorInfo(megapoolAddress common.Address, pubkey types.ValidatorPubkey) (*megapool.MegapoolValidatorInfo, bool) {
	info, exists := s.MegapoolValidatorInfo[MegapoolValidatorKey{MegapoolAddress: megapoolAddress, Pubkey: pubkey}]
	return info, exists
}

// Creates a snapshot of the Rocket Pool network state, on both the Execution and Consensus layers.
// If nodeAddresses is nil, all nodes are queried. Otherwise, only the specified nodes are included.
func (m *NetworkStateManager) createNetworkState(slotNumber uint64, nodeAddresses []common.Address) (*NetworkState, error) {
	allNodes := len(nodeAddresses) == 0
	steps := 9
	currentStep := 0

	// Get the execution block for the given slot
	beaconBlock, exists, err := m.bc.GetBeaconBlock(fmt.Sprintf("%d", slotNumber))
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon block for slot %d: %w", slotNumber, err)
	}
	if !exists {
		return nil, fmt.Errorf("slot %d did not have a Beacon block", slotNumber)
	}

	// Get the corresponding block on the EL.
	elBlockNumber, found, err := beacon.ResolveExecutionBlockNumber(context.Background(), m.rp.Client, beaconBlock)
	if err != nil {
		return nil, err
	}
	if !found && beaconBlock.HasExecutionPayload {
		return nil, fmt.Errorf("slot %d has an execution payload association but no resolvable EL block number", slotNumber)
	}
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlockNumber),
	}

	beaconConfig, err := m.getBeaconConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon config: %w", err)
	}

	// Create the state wrapper
	state := &NetworkState{
		BeaconSlotNumber: slotNumber,
		ElBlockNumber:    elBlockNumber,
		BeaconConfig:     *beaconConfig,
	}

	m.logLine("Getting network state for EL block %d, Beacon slot %d", elBlockNumber, slotNumber)
	start := time.Now()

	// Network contracts and details
	contracts, err := rpstate.NewNetworkContracts(m.rp, m.multicaller, m.balanceBatcher, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting network contracts: %w", err)
	}
	state.NetworkDetails, err = rpstate.NewNetworkDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting network details: %w", err)
	}
	currentStep++
	m.logLine("%d/%d - Retrieved network details (%s so far)", currentStep, steps, time.Since(start))

	// Node details
	if allNodes {
		state.NodeDetails, err = rpstate.GetAllNativeNodeDetails(m.rp, contracts)
		if err != nil {
			return nil, fmt.Errorf("error getting all node details: %w", err)
		}
	} else {
		state.NodeDetails = make([]rpstate.NativeNodeDetails, 0, len(nodeAddresses))
		for _, addr := range nodeAddresses {
			nodeDetails, err := rpstate.GetNativeNodeDetails(m.rp, contracts, addr)
			if err != nil {
				return nil, fmt.Errorf("error getting node details for %s: %w", addr.Hex(), err)
			}
			state.NodeDetails = append(state.NodeDetails, nodeDetails)
		}
	}
	currentStep++
	m.logLine("%d/%d - Retrieved node details (%s so far)", currentStep, steps, time.Since(start))

	// Minipool details
	if allNodes {
		state.MinipoolDetails, err = rpstate.GetAllNativeMinipoolDetails(m.rp, contracts)
		if err != nil {
			return nil, fmt.Errorf("error getting all minipool details: %w", err)
		}
	} else {
		state.MinipoolDetails = []rpstate.NativeMinipoolDetails{}
		for _, addr := range nodeAddresses {
			nodeMinipools, err := rpstate.GetNodeNativeMinipoolDetails(m.rp, contracts, addr)
			if err != nil {
				return nil, fmt.Errorf("error getting minipool details for node %s: %w", addr.Hex(), err)
			}
			state.MinipoolDetails = append(state.MinipoolDetails, nodeMinipools...)
		}
	}
	currentStep++
	m.logLine("%d/%d - Retrieved minipool details (%s so far)", currentStep, steps, time.Since(start))

	if allNodes {
		state.MegapoolValidators, err = rpstate.GetAllMegapoolValidators(m.rp, contracts)
		if err != nil {
			return nil, fmt.Errorf("error getting all megapool validator details: %w", err)
		}
	} else {
		state.MegapoolValidators = []megapool.MegapoolValidatorInfo{}
		for _, nd := range state.NodeDetails {
			if !nd.MegapoolDeployed {
				continue
			}
			validators, err := rpstate.GetNodeMegapoolValidators(m.rp, contracts, nd.MegapoolAddress)
			if err != nil {
				return nil, fmt.Errorf("error getting megapool validator details for %s: %w", nd.MegapoolAddress.Hex(), err)
			}
			state.MegapoolValidators = append(state.MegapoolValidators, validators...)
		}
	}
	currentStep++
	m.logLine("%d/%d - Retrieved megapool validator global index (%s so far)", currentStep, steps, time.Since(start))

	megapoolValidatorPubkeys := state.GetUniqueMegapoolPubkeys()
	megapoolAddresses := state.getMegapoolAddresses()

	// Fetch beacon validator statuses and EL megapool details in parallel
	var megapoolWg errgroup.Group
	megapoolWg.Go(func() error {
		statusMap, err := m.bc.GetValidatorStatuses(megapoolValidatorPubkeys, &beacon.ValidatorStatusOptions{
			Slot: &slotNumber,
		})
		if err != nil {
			return err
		}
		state.MegapoolValidatorDetails = statusMap
		return nil
	})
	megapoolWg.Go(func() error {
		var err error
		state.MegapoolDetails, err = rpstate.GetBulkMegapoolDetails(m.rp, contracts, megapoolAddresses)
		return err
	})
	if err := megapoolWg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting megapool details: %w", err)
	}
	currentStep++
	m.logLine("%d/%d - Retrieved megapool validator details (%s so far)", currentStep, steps, time.Since(start))

	// Oracle DAO member details
	state.OracleDaoMemberDetails, err = rpstate.GetAllOracleDaoMemberDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting Oracle DAO details: %w", err)
	}
	currentStep++
	m.logLine("%d/%d - Retrieved Oracle DAO details (%s so far)", currentStep, steps, time.Since(start))

	// Get the validator stats from Beacon
	minipoolPubkeys := state.GetMinipoolPubkeys()
	statusMap, err := m.bc.GetValidatorStatuses(minipoolPubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, err
	}
	state.MinipoolValidatorDetails = statusMap
	currentStep++
	m.logLine("%d/%d - Retrieved validator details (%s so far)", currentStep, steps, time.Since(start))

	// Get the complete node and user shares
	mpds := make([]*rpstate.NativeMinipoolDetails, len(state.MinipoolDetails))
	beaconBalances := make([]*big.Int, len(state.MinipoolDetails))
	for i, mpd := range state.MinipoolDetails {
		mpds[i] = &state.MinipoolDetails[i]
		validator := state.MinipoolValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			beaconBalances[i] = big.NewInt(0)
		} else {
			beaconBalances[i] = math.GweiToWei(float64(validator.Balance))
		}
	}
	err = rpstate.CalculateCompleteMinipoolShares(m.rp, contracts, mpds, beaconBalances)
	if err != nil {
		return nil, err
	}
	state.MinipoolValidatorDetails = statusMap
	currentStep++
	m.logLine("%d/%d - Calculated complete node and user balance shares (%s so far)", currentStep, steps, time.Since(start))

	// Protocol DAO proposals
	state.ProtocolDaoProposalDetails, err = rpstate.GetAllProtocolDaoProposalDetails(m.rp, contracts)
	if err != nil {
		return nil, fmt.Errorf("error getting Protocol DAO proposal details: %w", err)
	}
	currentStep++
	m.logLine("%d/%d - Retrieved Protocol DAO proposals (total time: %s)", currentStep, steps, time.Since(start))

	return state, state.Validate()
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

// Get the node's total borrowed ETH that counts towards RPL rewards (minipool + megapool)
func (s *NetworkStateIndex) GetEligibleBorrowedEth(node *rpstate.NativeNodeDetails) *big.Int {
	eligibleBorrowedEth := s.GetMinipoolEligibleBorrowedEth(node)
	eligibleBorrowedEth.Add(eligibleBorrowedEth, s.GetMegapoolEligibleBorrowedEth(node))
	return eligibleBorrowedEth
}

// Get the node's total staked RPL that counts towards RPL rewards (legacy + megapool)
func (s *NetworkState) GetRewardsEligibleRplStake(node *rpstate.NativeNodeDetails) *big.Int {
	rplStake := big.NewInt(0).Set(node.LegacyStakedRPL)
	// Megapool staked RPL counts towards RPL rewards
	rplStake.Add(rplStake, node.MegapoolStakedRPL)
	return rplStake
}

// Get the node's weight before scaling on participation
func (s *NetworkStateIndex) GetUnscaledNodeWeight(node *rpstate.NativeNodeDetails) *big.Int {
	eligibleBorrowedEth := s.GetEligibleBorrowedEth(node)
	if eligibleBorrowedEth.Sign() <= 0 {
		return big.NewInt(0)
	}
	return s.GetNodeWeight(eligibleBorrowedEth, s.GetRewardsEligibleRplStake(node))
}

// Starting in v8, RPL stake is phased out and replaced with weight.
// scaleByParticipation and allowRplForUnstartedValidators are hard-coded true here, since
// only v8 cares about weight.
func (s *NetworkStateIndex) CalculateNodeWeights() (map[common.Address]*big.Int, *big.Int, error) {
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
		wg.Go(func() error {
			// Calculate the weight
			nodeWeight := s.GetUnscaledNodeWeight(&node)
			if nodeWeight.Sign() <= 0 {
				weightSlice[i] = nodeWeight
				return nil
			}

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

func (s *NetworkStateIndex) GetMinipoolEligibleBorrowedEth(node *rpstate.NativeNodeDetails) *big.Int {
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
	return eligibleBorrowedEth
}

func (s *NetworkStateIndex) GetMegapoolEligibleBorrowedEth(node *rpstate.NativeNodeDetails) *big.Int {
	if !node.MegapoolDeployed {
		return big.NewInt(0)
	}

	megapool, exists := s.MegapoolDetails[node.MegapoolAddress]
	if !exists {
		return big.NewInt(0)
	}
	eligibleBorrowedEth := big.NewInt(0).Set(megapool.UserCapital)

	// Iterate over the validators
	for _, validator := range s.MegapoolToPubkeysMap[node.MegapoolAddress] {
		megapoolValidatorInfo, exists := s.GetMegapoolValidatorInfo(node.MegapoolAddress, validator)
		if !exists || !megapoolValidatorInfo.ValidatorInfo.InPrestake {
			continue
		}

		validatorTotalEth := big.NewInt(0).Set(math.MilliEthToWei(float64(megapoolValidatorInfo.ValidatorInfo.LastRequestedValue)))
		validatorBondedEth := big.NewInt(0).Set(math.MilliEthToWei(float64(megapoolValidatorInfo.ValidatorInfo.LastRequestedBond)))
		validatorUserEth := big.NewInt(0).Sub(validatorTotalEth, validatorBondedEth)
		eligibleBorrowedEth.Sub(eligibleBorrowedEth, validatorUserEth)
	}

	return eligibleBorrowedEth
}
