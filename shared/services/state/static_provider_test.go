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

	nsi := ns.ToIndexedNetworkState()

	if nsi.ElBlockNumber != 24866136 {
		t.Errorf("ElBlockNumber: got %d, want 24866136", nsi.ElBlockNumber)
	}
	if nsi.BeaconSlotNumber != 14100211 {
		t.Errorf("BeaconSlotNumber: got %d, want 14100211", nsi.BeaconSlotNumber)
	}

	// Verify index maps were rebuilt by UnmarshalJSON
	if len(nsi.NodeDetails) != 1 {
		t.Fatalf("NodeDetails count: got %d, want 1", len(nsi.NodeDetails))
	}
	nodeAddr := nsi.NodeDetails[0].NodeAddress
	if _, ok := nsi.NodeDetailsByAddress[nodeAddr]; !ok {
		t.Errorf("NodeDetailsByAddress missing %s", nodeAddr.Hex())
	}

	if len(nsi.MinipoolDetails) != 1 {
		t.Fatalf("MinipoolDetails count: got %d, want 1", len(nsi.MinipoolDetails))
	}
	mpAddr := nsi.MinipoolDetails[0].MinipoolAddress
	if _, ok := nsi.MinipoolDetailsByAddress[mpAddr]; !ok {
		t.Errorf("MinipoolDetailsByAddress missing %s", mpAddr.Hex())
	}

	if len(nsi.MinipoolValidatorDetails) != 1 {
		t.Errorf("MinipoolValidatorDetails count: got %d, want 1", len(nsi.MinipoolValidatorDetails))
	}
	if len(nsi.MegapoolValidatorDetails) != 1 {
		t.Errorf("MegapoolValidatorDetails count: got %d, want 1", len(nsi.MegapoolValidatorDetails))
	}
	if len(nsi.MegapoolValidators) != 1 {
		t.Errorf("MegapoolValidatorGlobalIndex count: got %d, want 1", len(nsi.MegapoolValidators))
	}
	if len(nsi.OracleDaoMemberDetails) != 1 {
		t.Errorf("OracleDaoMemberDetails count: got %d, want 1", len(nsi.OracleDaoMemberDetails))
	}
	if len(nsi.ProtocolDaoProposalDetails) != 1 {
		t.Errorf("ProtocolDaoProposalDetails count: got %d, want 1", len(nsi.ProtocolDaoProposalDetails))
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

func TestStaticProviderMegapoolToPubkeysMap(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	nsi := ns.ToIndexedNetworkState()

	// MegapoolToPubkeysMap must be rebuilt from MegapoolValidatorGlobalIndex
	if nsi.MegapoolToPubkeysMap == nil {
		t.Fatal("MegapoolToPubkeysMap is nil after loading from JSON")
	}

	// Every pubkey in the map must have a corresponding MegapoolValidatorInfo entry
	for addr, pubkeys := range nsi.MegapoolToPubkeysMap {
		for _, pk := range pubkeys {
			if _, ok := nsi.GetMegapoolValidatorInfo(addr, pk); !ok {
				t.Errorf("pubkey from MegapoolToPubkeysMap[%s] not found in MegapoolValidatorInfo", addr.Hex())
			}
		}
	}

	// Total pubkeys across all megapools must equal the non-empty entries in MegapoolValidatorGlobalIndex
	totalPubkeys := 0
	for _, pks := range nsi.MegapoolToPubkeysMap {
		totalPubkeys += len(pks)
	}

	expectedCount := 0
	for _, v := range nsi.MegapoolValidators {
		if len(v.Pubkey) > 0 {
			expectedCount++
		}
	}
	if totalPubkeys != expectedCount {
		t.Errorf("MegapoolToPubkeysMap total pubkeys: got %d, want %d", totalPubkeys, expectedCount)
	}
}

func TestStaticProviderMegapoolValidatorInfo(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	nsi := ns.ToIndexedNetworkState()

	if nsi.MegapoolValidatorInfo == nil {
		t.Fatal("MegapoolValidatorInfo is nil after loading from JSON")
	}

	// Every entry in MegapoolValidatorInfo must point back into MegapoolValidatorGlobalIndex
	for key, info := range nsi.MegapoolValidatorInfo {
		found := false
		for i := range nsi.MegapoolValidators {
			candidate := &nsi.MegapoolValidators[i]
			if candidate == info {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("MegapoolValidatorInfo[%s/%x] does not point into MegapoolValidatorGlobalIndex", key.MegapoolAddress.Hex(), key.Pubkey[:4])
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
	nsi := ns.ToIndexedNetworkState()
	provider := NewStaticNetworkStateProvider(nsi)

	got, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}
	if got != nsi {
		t.Error("GetHeadState returned a different pointer than the one provided")
	}
}
