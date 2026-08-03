package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
)

const minimalStatePath = "./testdata/minimal_state.json"

func TestStaticProviderFromFile(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(minimalStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	if ns.ElBlockNumber != 24866136 {
		t.Errorf("ElBlockNumber: got %d, want 24866136", ns.ElBlockNumber)
	}
	if ns.BeaconSlotNumber != 14100211 {
		t.Errorf("BeaconSlotNumber: got %d, want 14100211", ns.BeaconSlotNumber)
	}

	// Verify index maps were rebuilt by UnmarshalJSON
	if len(ns.NodeDetails) != 1 {
		t.Fatalf("NodeDetails count: got %d, want 1", len(ns.NodeDetails))
	}
	nodeAddr := ns.NodeDetails[0].NodeAddress
	if _, ok := ns.NodeDetailsByAddress[nodeAddr]; !ok {
		t.Errorf("NodeDetailsByAddress missing %s", nodeAddr.Hex())
	}

	if len(ns.MinipoolDetails) != 1 {
		t.Fatalf("MinipoolDetails count: got %d, want 1", len(ns.MinipoolDetails))
	}
	mpAddr := ns.MinipoolDetails[0].MinipoolAddress
	if _, ok := ns.MinipoolDetailsByAddress[mpAddr]; !ok {
		t.Errorf("MinipoolDetailsByAddress missing %s", mpAddr.Hex())
	}

	if len(ns.MinipoolValidatorDetails) != 1 {
		t.Errorf("MinipoolValidatorDetails count: got %d, want 1", len(ns.MinipoolValidatorDetails))
	}
	if len(ns.MegapoolValidatorDetails) != 1 {
		t.Errorf("MegapoolValidatorDetails count: got %d, want 1", len(ns.MegapoolValidatorDetails))
	}
	if len(ns.MegapoolValidatorGlobalIndex) != 1 {
		t.Errorf("MegapoolValidatorGlobalIndex count: got %d, want 1", len(ns.MegapoolValidatorGlobalIndex))
	}
	if len(ns.OracleDaoMemberDetails) != 1 {
		t.Errorf("OracleDaoMemberDetails count: got %d, want 1", len(ns.OracleDaoMemberDetails))
	}
	if len(ns.ProtocolDaoProposalDetails) != 1 {
		t.Errorf("ProtocolDaoProposalDetails count: got %d, want 1", len(ns.ProtocolDaoProposalDetails))
	}
}

func TestStaticProviderGetHeadStateForNode(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(minimalStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	// Address is ignored for the static provider, but the call must succeed.
	ns, err := provider.GetHeadStateForNode(common.HexToAddress("0x1234"))
	if err != nil {
		t.Fatalf("GetHeadStateForNode: %v", err)
	}
	if ns.ElBlockNumber != 24866136 {
		t.Errorf("ElBlockNumber: got %d, want 24866136", ns.ElBlockNumber)
	}
}

func TestStaticProviderGetStateForSlot(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(minimalStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetStateForSlot(999)
	if err != nil {
		t.Fatalf("GetStateForSlot: %v", err)
	}
	if ns.BeaconSlotNumber != 14100211 {
		t.Errorf("BeaconSlotNumber: got %d, want 14100211", ns.BeaconSlotNumber)
	}
}

func TestStaticProviderGetLatestBeaconBlock(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(minimalStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	block, err := provider.GetLatestBeaconBlock()
	if err != nil {
		t.Fatalf("GetLatestBeaconBlock: %v", err)
	}
	if block.Slot != 14100211 {
		t.Errorf("Slot: got %d, want 14100211", block.Slot)
	}
	if block.ExecutionBlockNumber != 24866136 {
		t.Errorf("ExecutionBlockNumber: got %d, want 24866136", block.ExecutionBlockNumber)
	}
	if !block.HasExecutionPayload {
		t.Error("HasExecutionPayload: got false, want true")
	}
}

func TestStaticProviderGetLatestFinalizedBeaconBlock(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(minimalStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	block, err := provider.GetLatestFinalizedBeaconBlock()
	if err != nil {
		t.Fatalf("GetLatestFinalizedBeaconBlock: %v", err)
	}
	if block.Slot != 14100211 {
		t.Errorf("Slot: got %d, want 14100211", block.Slot)
	}
}

const smallStatePath = "./testdata/small_state.json"

func TestStaticProviderMegapoolDetails(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	// MegapoolDetails must be non-nil and populated from the JSON
	if ns.MegapoolDetails == nil {
		t.Fatal("MegapoolDetails is nil after loading from JSON")
	}
	if len(ns.MegapoolDetails) == 0 {
		t.Fatal("MegapoolDetails is empty after loading from JSON")
	}

	// Every entry's map key must match its Address field
	for addr, details := range ns.MegapoolDetails {
		if addr != details.Address {
			t.Errorf("MegapoolDetails key %s does not match Address field %s", addr.Hex(), details.Address.Hex())
		}
	}

	// Spot-check: all loaded megapools must be deployed (per the fixture data)
	for addr, details := range ns.MegapoolDetails {
		if !details.Deployed {
			t.Errorf("MegapoolDetails[%s].Deployed is false, expected true", addr.Hex())
		}
	}
}

func TestStaticProviderMegapoolValidatorsByAddress(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	// MegapoolValidatorsByAddress must be rebuilt from MegapoolValidatorGlobalIndex
	if ns.MegapoolValidatorsByAddress == nil {
		t.Fatal("MegapoolValidatorsByAddress is nil after loading from JSON")
	}

	// Every record must be filed under the megapool that actually owns it
	for addr, validators := range ns.MegapoolValidatorsByAddress {
		for _, v := range validators {
			if v.MegapoolAddress != addr {
				t.Errorf("MegapoolValidatorsByAddress[%s] contains a record owned by %s", addr.Hex(), v.MegapoolAddress.Hex())
			}
		}
	}

	// Total validators across all megapools must equal the entries in MegapoolValidatorGlobalIndex
	// that carry a full-length pubkey
	totalValidators := 0
	for _, validators := range ns.MegapoolValidatorsByAddress {
		totalValidators += len(validators)
	}

	expectedCount := 0
	for _, v := range ns.MegapoolValidatorGlobalIndex {
		if len(v.Pubkey) == len(types.ValidatorPubkey{}) {
			expectedCount++
		}
	}
	if totalValidators != expectedCount {
		t.Errorf("MegapoolValidatorsByAddress total validators: got %d, want %d", totalValidators, expectedCount)
	}
}

func TestStaticProviderMegapoolValidatorAliasing(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	if ns.MegapoolValidatorsByAddress == nil {
		t.Fatal("MegapoolValidatorsByAddress is nil after loading from JSON")
	}

	// Every entry must point back into MegapoolValidatorGlobalIndex rather than at a copy
	for addr, validators := range ns.MegapoolValidatorsByAddress {
		for _, info := range validators {
			found := false
			for i := range ns.MegapoolValidatorGlobalIndex {
				candidate := &ns.MegapoolValidatorGlobalIndex[i]
				if candidate == info {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("MegapoolValidatorsByAddress[%s] entry does not point into MegapoolValidatorGlobalIndex", addr.Hex())
			}
		}
	}
}

func TestStaticProviderChallengeableProposal(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	if len(ns.ProtocolDaoProposalDetails) == 0 {
		t.Fatal("ProtocolDaoProposalDetails is empty")
	}

	// Find proposals in Pending state (challengeable)
	pendingCount := 0
	for _, prop := range ns.ProtocolDaoProposalDetails {
		if prop.State == types.ProtocolDaoProposalState_Pending && prop.ID != 0 {
			pendingCount++

			// Compute slot time from beacon config
			slotTime := ns.BeaconConfig.GenesisTime + ns.BeaconSlotNumber*ns.BeaconConfig.SecondsPerSlot
			challengeDeadline := prop.CreatedTime.Add(prop.ChallengeWindow)

			// The proposal must still be within its challenge window relative to slot time
			if uint64(challengeDeadline.Unix()) <= slotTime {
				t.Errorf("Pending proposal %d: challenge window expired before slot time (deadline %s, slot time %d)",
					prop.ID, challengeDeadline, slotTime)
			}

			// Proposal bond and challenge bond must be positive
			if prop.ProposalBond == nil || prop.ProposalBond.Sign() <= 0 {
				t.Errorf("Pending proposal %d has non-positive ProposalBond", prop.ID)
			}
			if prop.ChallengeBond == nil || prop.ChallengeBond.Sign() <= 0 {
				t.Errorf("Pending proposal %d has non-positive ChallengeBond", prop.ID)
			}

			t.Logf("Found challengeable proposal: id=%d, proposer=%s, message=%q, challengeDeadline=%s",
				prop.ID, prop.ProposerAddress.Hex(), prop.Message, challengeDeadline)
		}
	}

	if pendingCount == 0 {
		t.Error("No Pending (challengeable) proposals found in fixture")
	}
}

func TestStaticProviderFromConstructor(t *testing.T) {
	ns := buildTestState()
	provider := NewStaticNetworkStateProvider(ns)

	got, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}
	if got != ns {
		t.Error("GetHeadState returned a different pointer than the one provided")
	}
}
