package state

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/types"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/units"
)

func newTestPubkey(b byte) types.ValidatorPubkey {
	var pk types.ValidatorPubkey
	pk[0] = b
	return pk
}

func buildTestState() *NetworkState {
	nodeAddrA := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	nodeAddrB := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	mpAddrA1 := common.HexToAddress("0xA100000000000000000000000000000000000000")
	mpAddrA2 := common.HexToAddress("0xA200000000000000000000000000000000000000")
	mpAddrB1 := common.HexToAddress("0xB100000000000000000000000000000000000000")
	megapoolAddrA := common.HexToAddress("0xAA00000000000000000000000000000000000000")

	pubkeyA1 := newTestPubkey(0xA1)
	pubkeyA2 := newTestPubkey(0xA2)
	pubkeyB1 := newTestPubkey(0xB1)
	megapoolPubkey := newTestPubkey(0xCC)

	nodeDetails := []rpstate.NativeNodeDetails{
		{
			Exists:                           true,
			NodeAddress:                      nodeAddrA,
			RegistrationTime:                 big.NewInt(1000),
			RewardNetwork:                    big.NewInt(0),
			LegacyStakedRPL: units.NewWei(big.NewInt(100)),
			EffectiveRPLStake: units.NewWei(big.NewInt(100)),
			MinimumRPLStake: units.NewWei(big.NewInt(10)),
			MaximumRPLStake: units.NewWei(big.NewInt(1000)),
			EthBorrowed: units.NewWei(big.NewInt(0)),
			EthBorrowedLimit: units.NewWei(big.NewInt(0)),
			MegapoolETHBorrowed: units.NewWei(big.NewInt(0)),
			MinipoolETHBorrowed: units.NewWei(big.NewInt(0)),
			EthBonded: units.NewWei(big.NewInt(0)),
			MegapoolEthBonded: units.NewWei(big.NewInt(0)),
			MinipoolETHBonded: units.NewWei(big.NewInt(0)),
			MegapoolStakedRPL: units.NewWei(big.NewInt(0)),
			UnstakingRPL: units.NewWei(big.NewInt(0)),
			LockedRPL: units.NewWei(big.NewInt(0)),
			MinipoolCount:                    big.NewInt(2),
			BalanceETH: units.NewWei(big.NewInt(0)),
			BalanceRETH: units.NewWei(big.NewInt(0)),
			BalanceRPL: units.NewWei(big.NewInt(0)),
			BalanceOldRPL: units.NewWei(big.NewInt(0)),
			DepositCreditBalance: units.NewWei(big.NewInt(0)),
			SmoothingPoolRegistrationChanged: big.NewInt(0),
			CollateralisationRatio: units.NewWei(big.NewInt(0)),
			DistributorBalance: units.NewWei(big.NewInt(0)),
			MegapoolAddress:                  megapoolAddrA,
			MegapoolDeployed:                 true,
		},
		{
			Exists:                           true,
			NodeAddress:                      nodeAddrB,
			RegistrationTime:                 big.NewInt(2000),
			RewardNetwork:                    big.NewInt(0),
			LegacyStakedRPL: units.NewWei(big.NewInt(200)),
			EffectiveRPLStake: units.NewWei(big.NewInt(200)),
			MinimumRPLStake: units.NewWei(big.NewInt(20)),
			MaximumRPLStake: units.NewWei(big.NewInt(2000)),
			EthBorrowed: units.NewWei(big.NewInt(0)),
			EthBorrowedLimit: units.NewWei(big.NewInt(0)),
			MegapoolETHBorrowed: units.NewWei(big.NewInt(0)),
			MinipoolETHBorrowed: units.NewWei(big.NewInt(0)),
			EthBonded: units.NewWei(big.NewInt(0)),
			MegapoolEthBonded: units.NewWei(big.NewInt(0)),
			MinipoolETHBonded: units.NewWei(big.NewInt(0)),
			MegapoolStakedRPL: units.NewWei(big.NewInt(0)),
			UnstakingRPL: units.NewWei(big.NewInt(0)),
			LockedRPL: units.NewWei(big.NewInt(0)),
			MinipoolCount:                    big.NewInt(1),
			BalanceETH: units.NewWei(big.NewInt(0)),
			BalanceRETH: units.NewWei(big.NewInt(0)),
			BalanceRPL: units.NewWei(big.NewInt(0)),
			BalanceOldRPL: units.NewWei(big.NewInt(0)),
			DepositCreditBalance: units.NewWei(big.NewInt(0)),
			SmoothingPoolRegistrationChanged: big.NewInt(0),
			CollateralisationRatio: units.NewWei(big.NewInt(0)),
			DistributorBalance: units.NewWei(big.NewInt(0)),
		},
	}

	minipoolDetails := []rpstate.NativeMinipoolDetails{
		{
			Exists:                            true,
			MinipoolAddress:                   mpAddrA1,
			Pubkey:                            pubkeyA1,
			NodeAddress:                       nodeAddrA,
			NodeFee: units.NewWei(big.NewInt(1e17)),
			NodeDepositBalance: units.NewWei(big.NewInt(16_000_000_000)),
			UserDepositBalance: units.NewWei(big.NewInt(16_000_000_000)),
			StatusTime:                        big.NewInt(1000),
			StatusBlock:                       big.NewInt(100),
			Balance: units.NewWei(big.NewInt(0)),
			DistributableBalance: units.NewWei(big.NewInt(0)),
			NodeShareOfBalance: units.NewWei(big.NewInt(0)),
			UserShareOfBalance: units.NewWei(big.NewInt(0)),
			NodeRefundBalance: units.NewWei(big.NewInt(0)),
			PenaltyCount:                      big.NewInt(0),
			PenaltyRate:                       units.NewWei(big.NewInt(0)),
			UserDepositAssignedTime:           big.NewInt(1000),
			NodeShareOfBalanceIncludingBeacon: units.NewWei(big.NewInt(0)),
			UserShareOfBalanceIncludingBeacon: units.NewWei(big.NewInt(0)),
			NodeShareOfBeaconBalance: units.NewWei(big.NewInt(0)),
			UserShareOfBeaconBalance: units.NewWei(big.NewInt(0)),
			LastBondReductionTime:             big.NewInt(0),
			LastBondReductionPrevValue: units.NewWei(big.NewInt(0)),
			LastBondReductionPrevNodeFee: units.NewWei(big.NewInt(0)),
			ReduceBondTime:                    big.NewInt(0),
			ReduceBondValue: units.NewWei(big.NewInt(0)),
			PreMigrationBalance: units.NewWei(big.NewInt(0)),
			Status:                            types.Staking,
		},
		{
			Exists:                            true,
			MinipoolAddress:                   mpAddrA2,
			Pubkey:                            pubkeyA2,
			NodeAddress:                       nodeAddrA,
			NodeFee: units.NewWei(big.NewInt(1e17)),
			NodeDepositBalance: units.NewWei(big.NewInt(8_000_000_000)),
			UserDepositBalance: units.NewWei(big.NewInt(24_000_000_000)),
			StatusTime:                        big.NewInt(1100),
			StatusBlock:                       big.NewInt(110),
			Balance: units.NewWei(big.NewInt(0)),
			DistributableBalance: units.NewWei(big.NewInt(0)),
			NodeShareOfBalance: units.NewWei(big.NewInt(0)),
			UserShareOfBalance: units.NewWei(big.NewInt(0)),
			NodeRefundBalance: units.NewWei(big.NewInt(0)),
			PenaltyCount:                      big.NewInt(0),
			PenaltyRate:                       units.NewWei(big.NewInt(0)),
			UserDepositAssignedTime:           big.NewInt(1100),
			NodeShareOfBalanceIncludingBeacon: units.NewWei(big.NewInt(0)),
			UserShareOfBalanceIncludingBeacon: units.NewWei(big.NewInt(0)),
			NodeShareOfBeaconBalance: units.NewWei(big.NewInt(0)),
			UserShareOfBeaconBalance: units.NewWei(big.NewInt(0)),
			LastBondReductionTime:             big.NewInt(0),
			LastBondReductionPrevValue: units.NewWei(big.NewInt(0)),
			LastBondReductionPrevNodeFee: units.NewWei(big.NewInt(0)),
			ReduceBondTime:                    big.NewInt(0),
			ReduceBondValue: units.NewWei(big.NewInt(0)),
			PreMigrationBalance: units.NewWei(big.NewInt(0)),
			Status:                            types.Staking,
		},
		{
			Exists:                            true,
			MinipoolAddress:                   mpAddrB1,
			Pubkey:                            pubkeyB1,
			NodeAddress:                       nodeAddrB,
			NodeFee: units.NewWei(big.NewInt(5e16)),
			NodeDepositBalance: units.NewWei(big.NewInt(16_000_000_000)),
			UserDepositBalance: units.NewWei(big.NewInt(16_000_000_000)),
			StatusTime:                        big.NewInt(2000),
			StatusBlock:                       big.NewInt(200),
			Balance: units.NewWei(big.NewInt(0)),
			DistributableBalance: units.NewWei(big.NewInt(0)),
			NodeShareOfBalance: units.NewWei(big.NewInt(0)),
			UserShareOfBalance: units.NewWei(big.NewInt(0)),
			NodeRefundBalance: units.NewWei(big.NewInt(0)),
			PenaltyCount:                      big.NewInt(0),
			PenaltyRate:                       units.NewWei(big.NewInt(0)),
			UserDepositAssignedTime:           big.NewInt(2000),
			NodeShareOfBalanceIncludingBeacon: units.NewWei(big.NewInt(0)),
			UserShareOfBalanceIncludingBeacon: units.NewWei(big.NewInt(0)),
			NodeShareOfBeaconBalance: units.NewWei(big.NewInt(0)),
			UserShareOfBeaconBalance: units.NewWei(big.NewInt(0)),
			LastBondReductionTime:             big.NewInt(0),
			LastBondReductionPrevValue: units.NewWei(big.NewInt(0)),
			LastBondReductionPrevNodeFee: units.NewWei(big.NewInt(0)),
			ReduceBondTime:                    big.NewInt(0),
			ReduceBondValue: units.NewWei(big.NewInt(0)),
			PreMigrationBalance: units.NewWei(big.NewInt(0)),
			Status:                            types.Staking,
		},
	}

	megapoolValidatorGlobalIndex := []megapool.MegapoolValidatorInfo{
		{
			Pubkey:          megapoolPubkey[:],
			MegapoolAddress: megapoolAddrA,
			ValidatorId:     1,
			ValidatorInfo: megapool.ValidatorInfo{
				Staked: true,
			},
		},
	}

	megapoolDetails := map[common.Address]rpstate.NativeMegapoolDetails{
		megapoolAddrA: {
			Address:              megapoolAddrA,
			Deployed:             true,
			ActiveValidatorCount: 1,
			UserCapital: units.NewWei(big.NewInt(24_000_000_000)),
			NodeBond: units.NewWei(big.NewInt(8_000_000_000)),
			NodeDebt: units.NewWei(big.NewInt(0)),
			RefundValue: units.NewWei(big.NewInt(0)),
			AssignedValue: units.NewWei(big.NewInt(0)),
			BondRequirement: units.NewWei(big.NewInt(0)),
			EthBalance: units.NewWei(big.NewInt(0)),
			PendingRewards: units.NewWei(big.NewInt(0)),
			NodeQueuedBond: units.NewWei(big.NewInt(0)),
		},
	}

	return &NetworkState{
		ElBlockNumber:    1000,
		BeaconSlotNumber: 32000,
		BeaconConfig: beacon.Eth2Config{
			GenesisTime:    1600000000,
			SecondsPerSlot: 12,
			SlotsPerEpoch:  32,
		},
		NetworkDetails: &rpstate.NetworkDetails{
			RplPrice: units.NewWei(big.NewInt(1e16)),
			MinCollateralFraction:             units.NewWei(big.NewInt(1e17)),
			MaxCollateralFraction:             units.NewWei(big.NewInt(15e17)),
			IntervalDuration:                  28 * 24 * time.Hour,
			NodeOperatorRewardsPercent: units.NewWei(big.NewInt(0)),
			TrustedNodeOperatorRewardsPercent: units.NewWei(big.NewInt(0)),
			ProtocolDaoRewardsPercent: units.NewWei(big.NewInt(0)),
		},
		NodeDetails:     nodeDetails,
		MinipoolDetails: minipoolDetails,
		MinipoolValidatorDetails: ValidatorDetailsMap{
			pubkeyA1: {Pubkey: pubkeyA1, Index: "1", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
			pubkeyA2: {Pubkey: pubkeyA2, Index: "2", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
			pubkeyB1: {Pubkey: pubkeyB1, Index: "3", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
		},
		MegapoolValidatorDetails: ValidatorDetailsMap{
			megapoolPubkey: {Pubkey: megapoolPubkey, Index: "4", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
		},
		MegapoolValidators: megapoolValidatorGlobalIndex,
		MegapoolDetails:    megapoolDetails,
		OracleDaoMemberDetails: []rpstate.OracleDaoMemberDetails{
			{
				Address:          nodeAddrA,
				Exists:           true,
				ID:               "odao-a",
				RPLBondAmount:    units.NewWei(big.NewInt(1000)),
				JoinedTime:       time.Unix(1000, 0),
				LastProposalTime: time.Unix(1000, 0),
			},
		},
		ProtocolDaoProposalDetails: []protocol.ProtocolDaoProposalDetails{
			{
				ID:                   1,
				ProposerAddress:      nodeAddrA,
				VotingPowerRequired:  units.NewWei(big.NewInt(100)),
				VotingPowerFor:       units.NewWei(big.NewInt(80)),
				VotingPowerAgainst:   units.NewWei(big.NewInt(10)),
				VotingPowerAbstained: units.NewWei(big.NewInt(10)),
			},
		},
	}
}

// TestNetworkStateJSONRoundtrip verifies that marshaling and unmarshaling
// a NetworkState preserves all serialized fields and correctly rebuilds
// the index maps (NodeDetailsByAddress, MinipoolDetailsByAddress,
// MinipoolDetailsByNode) that are excluded from JSON.
func TestNetworkStateJSONRoundtrip(t *testing.T) {
	original := buildTestState().ToIndexedNetworkState()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored NetworkStateIndex
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Scalar fields
	if restored.ElBlockNumber != original.ElBlockNumber {
		t.Errorf("ElBlockNumber mismatch: got %d, want %d", restored.ElBlockNumber, original.ElBlockNumber)
	}
	if restored.BeaconSlotNumber != original.BeaconSlotNumber {
		t.Errorf("BeaconSlotNumber mismatch: got %d, want %d", restored.BeaconSlotNumber, original.BeaconSlotNumber)
	}

	// Node details
	if len(restored.NodeDetails) != len(original.NodeDetails) {
		t.Fatalf("NodeDetails count: got %d, want %d", len(restored.NodeDetails), len(original.NodeDetails))
	}

	// NodeDetailsByAddress rebuilt
	if len(restored.NodeDetailsByAddress) != len(original.NodeDetailsByAddress) {
		t.Fatalf("NodeDetailsByAddress count: got %d, want %d", len(restored.NodeDetailsByAddress), len(original.NodeDetailsByAddress))
	}
	for addr := range original.NodeDetailsByAddress {
		if _, ok := restored.NodeDetailsByAddress[addr]; !ok {
			t.Errorf("NodeDetailsByAddress missing key %s", addr.Hex())
		}
	}

	// Verify NodeDetailsByAddress points into the NodeDetails slice
	for addr, ptr := range restored.NodeDetailsByAddress {
		found := false
		for i := range restored.NodeDetails {
			if &restored.NodeDetails[i] == ptr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("NodeDetailsByAddress[%s] does not point into NodeDetails slice", addr.Hex())
		}
	}

	// Minipool details
	if len(restored.MinipoolDetails) != len(original.MinipoolDetails) {
		t.Fatalf("MinipoolDetails count: got %d, want %d", len(restored.MinipoolDetails), len(original.MinipoolDetails))
	}

	// MinipoolDetailsByAddress rebuilt
	if len(restored.MinipoolDetailsByAddress) != len(original.MinipoolDetailsByAddress) {
		t.Fatalf("MinipoolDetailsByAddress count: got %d, want %d", len(restored.MinipoolDetailsByAddress), len(original.MinipoolDetailsByAddress))
	}
	for addr := range original.MinipoolDetailsByAddress {
		if _, ok := restored.MinipoolDetailsByAddress[addr]; !ok {
			t.Errorf("MinipoolDetailsByAddress missing key %s", addr.Hex())
		}
	}

	// MinipoolDetailsByNode rebuilt
	nodeAddrA := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	nodeAddrB := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	if len(restored.MinipoolDetailsByNode) != 2 {
		t.Fatalf("MinipoolDetailsByNode count: got %d, want 2", len(restored.MinipoolDetailsByNode))
	}
	if len(restored.MinipoolDetailsByNode[nodeAddrA]) != 2 {
		t.Errorf("MinipoolDetailsByNode[nodeA] count: got %d, want 2", len(restored.MinipoolDetailsByNode[nodeAddrA]))
	}
	if len(restored.MinipoolDetailsByNode[nodeAddrB]) != 1 {
		t.Errorf("MinipoolDetailsByNode[nodeB] count: got %d, want 1", len(restored.MinipoolDetailsByNode[nodeAddrB]))
	}

	// Minipool validator details
	if len(restored.MinipoolValidatorDetails) != len(original.MinipoolValidatorDetails) {
		t.Errorf("MinipoolValidatorDetails count: got %d, want %d",
			len(restored.MinipoolValidatorDetails), len(original.MinipoolValidatorDetails))
	}

	// Megapool validator details
	if len(restored.MegapoolValidatorDetails) != len(original.MegapoolValidatorDetails) {
		t.Errorf("MegapoolValidatorDetails count: got %d, want %d",
			len(restored.MegapoolValidatorDetails), len(original.MegapoolValidatorDetails))
	}

	// Megapool validator global index
	if len(restored.MegapoolValidators) != len(original.MegapoolValidators) {
		t.Errorf("MegapoolValidatorGlobalIndex count: got %d, want %d",
			len(restored.MegapoolValidators), len(original.MegapoolValidators))
	}

	// Oracle DAO member details
	if len(restored.OracleDaoMemberDetails) != len(original.OracleDaoMemberDetails) {
		t.Errorf("OracleDaoMemberDetails count: got %d, want %d",
			len(restored.OracleDaoMemberDetails), len(original.OracleDaoMemberDetails))
	}
	if restored.OracleDaoMemberDetails[0].ID != "odao-a" {
		t.Errorf("OracleDaoMemberDetails[0].ID: got %q, want %q", restored.OracleDaoMemberDetails[0].ID, "odao-a")
	}

	// Protocol DAO proposal details
	if len(restored.ProtocolDaoProposalDetails) != len(original.ProtocolDaoProposalDetails) {
		t.Errorf("ProtocolDaoProposalDetails count: got %d, want %d",
			len(restored.ProtocolDaoProposalDetails), len(original.ProtocolDaoProposalDetails))
	}
	if restored.ProtocolDaoProposalDetails[0].ID != 1 {
		t.Errorf("ProtocolDaoProposalDetails[0].ID: got %d, want 1", restored.ProtocolDaoProposalDetails[0].ID)
	}
}

// TestUnmarshalDuplicateNodeErrors verifies that unmarshaling a
// NetworkState with duplicate node addresses produces an error.
func TestUnmarshalDuplicateNodeErrors(t *testing.T) {
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	state := &NetworkState{
		BeaconConfig: beacon.Eth2Config{},
		NetworkDetails: &rpstate.NetworkDetails{
			RplPrice: units.NewWei(big.NewInt(0)),
			MinCollateralFraction:             units.NewWei(big.NewInt(0)),
			MaxCollateralFraction:             units.NewWei(big.NewInt(0)),
			NodeOperatorRewardsPercent: units.NewWei(big.NewInt(0)),
			TrustedNodeOperatorRewardsPercent: units.NewWei(big.NewInt(0)),
			ProtocolDaoRewardsPercent: units.NewWei(big.NewInt(0)),
		},
		NodeDetails: []rpstate.NativeNodeDetails{
			{NodeAddress: addr, RegistrationTime: big.NewInt(0), RewardNetwork: big.NewInt(0), LegacyStakedRPL: units.NewWei(big.NewInt(0)), EffectiveRPLStake: units.NewWei(big.NewInt(0)), MinimumRPLStake: units.NewWei(big.NewInt(0)), MaximumRPLStake: units.NewWei(big.NewInt(0)), EthBorrowed: units.NewWei(big.NewInt(0)), EthBorrowedLimit: units.NewWei(big.NewInt(0)), MegapoolETHBorrowed: units.NewWei(big.NewInt(0)), MinipoolETHBorrowed: units.NewWei(big.NewInt(0)), EthBonded: units.NewWei(big.NewInt(0)), MegapoolEthBonded: units.NewWei(big.NewInt(0)), MinipoolETHBonded: units.NewWei(big.NewInt(0)), MegapoolStakedRPL: units.NewWei(big.NewInt(0)), UnstakingRPL: units.NewWei(big.NewInt(0)), LockedRPL: units.NewWei(big.NewInt(0)), MinipoolCount: big.NewInt(0), BalanceETH: units.NewWei(big.NewInt(0)), BalanceRETH: units.NewWei(big.NewInt(0)), BalanceRPL: units.NewWei(big.NewInt(0)), BalanceOldRPL: units.NewWei(big.NewInt(0)), DepositCreditBalance: units.NewWei(big.NewInt(0)), SmoothingPoolRegistrationChanged: big.NewInt(0), CollateralisationRatio: units.NewWei(big.NewInt(0)), DistributorBalance: units.NewWei(big.NewInt(0))},
			{NodeAddress: addr, RegistrationTime: big.NewInt(0), RewardNetwork: big.NewInt(0), LegacyStakedRPL: units.NewWei(big.NewInt(0)), EffectiveRPLStake: units.NewWei(big.NewInt(0)), MinimumRPLStake: units.NewWei(big.NewInt(0)), MaximumRPLStake: units.NewWei(big.NewInt(0)), EthBorrowed: units.NewWei(big.NewInt(0)), EthBorrowedLimit: units.NewWei(big.NewInt(0)), MegapoolETHBorrowed: units.NewWei(big.NewInt(0)), MinipoolETHBorrowed: units.NewWei(big.NewInt(0)), EthBonded: units.NewWei(big.NewInt(0)), MegapoolEthBonded: units.NewWei(big.NewInt(0)), MinipoolETHBonded: units.NewWei(big.NewInt(0)), MegapoolStakedRPL: units.NewWei(big.NewInt(0)), UnstakingRPL: units.NewWei(big.NewInt(0)), LockedRPL: units.NewWei(big.NewInt(0)), MinipoolCount: big.NewInt(0), BalanceETH: units.NewWei(big.NewInt(0)), BalanceRETH: units.NewWei(big.NewInt(0)), BalanceRPL: units.NewWei(big.NewInt(0)), BalanceOldRPL: units.NewWei(big.NewInt(0)), DepositCreditBalance: units.NewWei(big.NewInt(0)), SmoothingPoolRegistrationChanged: big.NewInt(0), CollateralisationRatio: units.NewWei(big.NewInt(0)), DistributorBalance: units.NewWei(big.NewInt(0))},
		},
		MinipoolDetails:          []rpstate.NativeMinipoolDetails{},
		MinipoolValidatorDetails: ValidatorDetailsMap{},
		MegapoolValidatorDetails: ValidatorDetailsMap{},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored NetworkStateIndex
	err = json.Unmarshal(data, &restored)
	if err == nil {
		t.Fatal("expected error for duplicate node address, got nil")
	}
}

// TestUnmarshalDuplicateMinipoolErrors verifies that unmarshaling a
// NetworkState with duplicate minipool addresses produces an error.
func TestUnmarshalDuplicateMinipoolErrors(t *testing.T) {
	mpAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	nodeAddr := common.HexToAddress("0x3333333333333333333333333333333333333333")
	pk1 := newTestPubkey(0x01)
	pk2 := newTestPubkey(0x02)

	zeroInt := func() *big.Int { return big.NewInt(0) }
	zeroWei := func() units.Wei { return units.Wei{} }

	state := &NetworkState{
		BeaconConfig: beacon.Eth2Config{},
		NetworkDetails: &rpstate.NetworkDetails{
			RplPrice:                          zeroWei(),
			MinCollateralFraction:             zeroWei(),
			MaxCollateralFraction:             zeroWei(),
			NodeOperatorRewardsPercent:        zeroWei(),
			TrustedNodeOperatorRewardsPercent: zeroWei(),
			ProtocolDaoRewardsPercent:         zeroWei(),
		},
		NodeDetails: []rpstate.NativeNodeDetails{
			{NodeAddress: nodeAddr, RegistrationTime: zeroInt(), RewardNetwork: zeroInt(), LegacyStakedRPL: zeroWei(), EffectiveRPLStake: zeroWei(), MinimumRPLStake: zeroWei(), MaximumRPLStake: zeroWei(), EthBorrowed: zeroWei(), EthBorrowedLimit: zeroWei(), MegapoolETHBorrowed: zeroWei(), MinipoolETHBorrowed: zeroWei(), EthBonded: zeroWei(), MegapoolEthBonded: zeroWei(), MinipoolETHBonded: zeroWei(), MegapoolStakedRPL: zeroWei(), UnstakingRPL: zeroWei(), LockedRPL: zeroWei(), MinipoolCount: zeroInt(), BalanceETH: zeroWei(), BalanceRETH: zeroWei(), BalanceRPL: zeroWei(), BalanceOldRPL: zeroWei(), DepositCreditBalance: zeroWei(), SmoothingPoolRegistrationChanged: zeroInt(), CollateralisationRatio: zeroWei()},
		},
		MinipoolDetails: []rpstate.NativeMinipoolDetails{
			{Exists: true, MinipoolAddress: mpAddr, Pubkey: pk1, NodeAddress: nodeAddr, NodeFee: zeroWei(), NodeDepositBalance: zeroWei(), UserDepositBalance: zeroWei(), StatusTime: zeroInt(), StatusBlock: zeroInt(), Balance: zeroWei(), DistributableBalance: zeroWei(), NodeShareOfBalance: zeroWei(), UserShareOfBalance: zeroWei(), NodeRefundBalance: zeroWei(), PenaltyCount: zeroInt(), PenaltyRate: zeroWei(), UserDepositAssignedTime: zeroInt(), NodeShareOfBalanceIncludingBeacon: zeroWei(), UserShareOfBalanceIncludingBeacon: zeroWei(), NodeShareOfBeaconBalance: zeroWei(), UserShareOfBeaconBalance: zeroWei(), LastBondReductionTime: zeroInt(), LastBondReductionPrevValue: zeroWei(), LastBondReductionPrevNodeFee: zeroWei(), ReduceBondTime: zeroInt(), ReduceBondValue: zeroWei(), PreMigrationBalance: zeroWei()},
			{Exists: true, MinipoolAddress: mpAddr, Pubkey: pk2, NodeAddress: nodeAddr, NodeFee: zeroWei(), NodeDepositBalance: zeroWei(), UserDepositBalance: zeroWei(), StatusTime: zeroInt(), StatusBlock: zeroInt(), Balance: zeroWei(), DistributableBalance: zeroWei(), NodeShareOfBalance: zeroWei(), UserShareOfBalance: zeroWei(), NodeRefundBalance: zeroWei(), PenaltyCount: zeroInt(), PenaltyRate: zeroWei(), UserDepositAssignedTime: zeroInt(), NodeShareOfBalanceIncludingBeacon: zeroWei(), UserShareOfBalanceIncludingBeacon: zeroWei(), NodeShareOfBeaconBalance: zeroWei(), UserShareOfBeaconBalance: zeroWei(), LastBondReductionTime: zeroInt(), LastBondReductionPrevValue: zeroWei(), LastBondReductionPrevNodeFee: zeroWei(), ReduceBondTime: zeroInt(), ReduceBondValue: zeroWei(), PreMigrationBalance: zeroWei()},
		},
		MinipoolValidatorDetails: ValidatorDetailsMap{},
		MegapoolValidatorDetails: ValidatorDetailsMap{},
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored NetworkState
	err = json.Unmarshal(data, &restored)
	if err == nil {
		t.Fatal("expected error for duplicate minipool address, got nil")
	}
}

func TestDuplicatePubkeyAcrossMegapools(t *testing.T) {
	megapoolAddrA := common.HexToAddress("0xAA00000000000000000000000000000000000000")
	megapoolAddrB := common.HexToAddress("0xBB00000000000000000000000000000000000000")
	pubkey := newTestPubkey(0xDD)

	state := &NetworkState{
		MinipoolValidatorDetails: ValidatorDetailsMap{},
		MegapoolValidatorDetails: ValidatorDetailsMap{},
		MegapoolValidators: []megapool.MegapoolValidatorInfo{
			{
				Pubkey:          pubkey[:],
				MegapoolAddress: megapoolAddrA,
				ValidatorId:     1,
				ValidatorInfo: megapool.ValidatorInfo{
					Staked: true,
				},
			},
			{
				// A validator in a different megapool reusing the same pubkey,
				// dequeued immediately (all status flags false)
				Pubkey:          pubkey[:],
				MegapoolAddress: megapoolAddrB,
				ValidatorId:     0,
				ValidatorInfo:   megapool.ValidatorInfo{},
			},
		},
	}

	stateIndex := state.ToIndexedNetworkState()

	checkState := func(t *testing.T, s *NetworkStateIndex) {
		t.Helper()

		if len(s.MegapoolValidatorInfo) != 2 {
			t.Errorf("MegapoolValidatorInfo count: got %d, want 2", len(s.MegapoolValidatorInfo))
		}

		infoA, exists := s.GetMegapoolValidatorInfo(megapoolAddrA, pubkey)
		if !exists {
			t.Fatal("validator info for megapool A not found")
		}
		if !infoA.ValidatorInfo.Staked {
			t.Error("megapool A validator lost Staked=true (overwritten by the duplicate pubkey)")
		}
		if infoA.ValidatorId != 1 {
			t.Errorf("megapool A ValidatorId: got %d, want 1", infoA.ValidatorId)
		}

		infoB, exists := s.GetMegapoolValidatorInfo(megapoolAddrB, pubkey)
		if !exists {
			t.Fatal("validator info for megapool B not found")
		}
		if infoB.ValidatorInfo.Staked {
			t.Error("megapool B validator should not be staked")
		}

		if len(s.MegapoolToPubkeysMap[megapoolAddrA]) != 1 {
			t.Errorf("MegapoolToPubkeysMap[A] count: got %d, want 1", len(s.MegapoolToPubkeysMap[megapoolAddrA]))
		}
		if len(s.MegapoolToPubkeysMap[megapoolAddrB]) != 1 {
			t.Errorf("MegapoolToPubkeysMap[B] count: got %d, want 1", len(s.MegapoolToPubkeysMap[megapoolAddrB]))
		}
	}

	numMegapoolPubkeys := len(stateIndex.GetUniqueMegapoolPubkeys())
	if numMegapoolPubkeys != 1 {
		t.Errorf("numMegapoolPubkeys: got %d, want 1", numMegapoolPubkeys)
	}
	checkState(t, stateIndex)

	// Round-trip through JSON to exercise the UnmarshalJSON rebuild path
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var restored NetworkState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	checkState(t, restored.ToIndexedNetworkState())
}
