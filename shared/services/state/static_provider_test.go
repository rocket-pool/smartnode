package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func TestStaticProviderMegapoolToPubkeysMap(t *testing.T) {
	provider, err := NewStaticNetworkStateProviderFromFile(smallStatePath)
	if err != nil {
		t.Fatalf("NewStaticNetworkStateProviderFromFile: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	// MegapoolToPubkeysMap must be rebuilt from MegapoolValidatorGlobalIndex
	if ns.MegapoolToPubkeysMap == nil {
		t.Fatal("MegapoolToPubkeysMap is nil after loading from JSON")
	}

	// Every pubkey in the map must have a corresponding MegapoolValidatorInfo entry
	for addr, pubkeys := range ns.MegapoolToPubkeysMap {
		for _, pk := range pubkeys {
			if _, ok := ns.MegapoolValidatorInfo[pk]; !ok {
				t.Errorf("pubkey from MegapoolToPubkeysMap[%s] not found in MegapoolValidatorInfo", addr.Hex())
			}
		}
	}

	// Total pubkeys across all megapools must equal the non-empty entries in MegapoolValidatorGlobalIndex
	totalPubkeys := 0
	for _, pks := range ns.MegapoolToPubkeysMap {
		totalPubkeys += len(pks)
	}

	expectedCount := 0
	for _, v := range ns.MegapoolValidatorGlobalIndex {
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

	if ns.MegapoolValidatorInfo == nil {
		t.Fatal("MegapoolValidatorInfo is nil after loading from JSON")
	}

	// Every entry in MegapoolValidatorInfo must point back into MegapoolValidatorGlobalIndex
	for pk, info := range ns.MegapoolValidatorInfo {
		found := false
		for i := range ns.MegapoolValidatorGlobalIndex {
			candidate := &ns.MegapoolValidatorGlobalIndex[i]
			if candidate == info {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("MegapoolValidatorInfo[%x] does not point into MegapoolValidatorGlobalIndex", pk[:4])
		}
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
