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
			LegacyStakedRPL:                  big.NewInt(100),
			EffectiveRPLStake:                big.NewInt(100),
			MinimumRPLStake:                  big.NewInt(10),
			MaximumRPLStake:                  big.NewInt(1000),
			EthBorrowed:                      big.NewInt(0),
			EthBorrowedLimit:                 big.NewInt(0),
			MegapoolETHBorrowed:              big.NewInt(0),
			MinipoolETHBorrowed:              big.NewInt(0),
			EthBonded:                        big.NewInt(0),
			MegapoolEthBonded:                big.NewInt(0),
			MinipoolETHBonded:                big.NewInt(0),
			MegapoolStakedRPL:                big.NewInt(0),
			UnstakingRPL:                     big.NewInt(0),
			LockedRPL:                        big.NewInt(0),
			MinipoolCount:                    big.NewInt(2),
			BalanceETH:                       big.NewInt(0),
			BalanceRETH:                      big.NewInt(0),
			BalanceRPL:                       big.NewInt(0),
			BalanceOldRPL:                    big.NewInt(0),
			DepositCreditBalance:             big.NewInt(0),
			DistributorBalanceUserETH:        big.NewInt(0),
			DistributorBalanceNodeETH:        big.NewInt(0),
			SmoothingPoolRegistrationChanged: big.NewInt(0),
			AverageNodeFee:                   big.NewInt(0),
			CollateralisationRatio:           big.NewInt(0),
			DistributorBalance:               big.NewInt(0),
			MegapoolAddress:                  megapoolAddrA,
			MegapoolDeployed:                 true,
		},
		{
			Exists:                           true,
			NodeAddress:                      nodeAddrB,
			RegistrationTime:                 big.NewInt(2000),
			RewardNetwork:                    big.NewInt(0),
			LegacyStakedRPL:                  big.NewInt(200),
			EffectiveRPLStake:                big.NewInt(200),
			MinimumRPLStake:                  big.NewInt(20),
			MaximumRPLStake:                  big.NewInt(2000),
			EthBorrowed:                      big.NewInt(0),
			EthBorrowedLimit:                 big.NewInt(0),
			MegapoolETHBorrowed:              big.NewInt(0),
			MinipoolETHBorrowed:              big.NewInt(0),
			EthBonded:                        big.NewInt(0),
			MegapoolEthBonded:                big.NewInt(0),
			MinipoolETHBonded:                big.NewInt(0),
			MegapoolStakedRPL:                big.NewInt(0),
			UnstakingRPL:                     big.NewInt(0),
			LockedRPL:                        big.NewInt(0),
			MinipoolCount:                    big.NewInt(1),
			BalanceETH:                       big.NewInt(0),
			BalanceRETH:                      big.NewInt(0),
			BalanceRPL:                       big.NewInt(0),
			BalanceOldRPL:                    big.NewInt(0),
			DepositCreditBalance:             big.NewInt(0),
			DistributorBalanceUserETH:        big.NewInt(0),
			DistributorBalanceNodeETH:        big.NewInt(0),
			SmoothingPoolRegistrationChanged: big.NewInt(0),
			AverageNodeFee:                   big.NewInt(0),
			CollateralisationRatio:           big.NewInt(0),
			DistributorBalance:               big.NewInt(0),
		},
	}

	minipoolDetails := []rpstate.NativeMinipoolDetails{
		{
			Exists:                            true,
			MinipoolAddress:                   mpAddrA1,
			Pubkey:                            pubkeyA1,
			NodeAddress:                       nodeAddrA,
			NodeFee:                           big.NewInt(1e17),
			NodeDepositBalance:                big.NewInt(16_000_000_000),
			UserDepositBalance:                big.NewInt(16_000_000_000),
			StatusTime:                        big.NewInt(1000),
			StatusBlock:                       big.NewInt(100),
			Balance:                           big.NewInt(0),
			DistributableBalance:              big.NewInt(0),
			NodeShareOfBalance:                big.NewInt(0),
			UserShareOfBalance:                big.NewInt(0),
			NodeRefundBalance:                 big.NewInt(0),
			PenaltyCount:                      big.NewInt(0),
			PenaltyRate:                       big.NewInt(0),
			UserDepositAssignedTime:           big.NewInt(1000),
			NodeShareOfBalanceIncludingBeacon: big.NewInt(0),
			UserShareOfBalanceIncludingBeacon: big.NewInt(0),
			NodeShareOfBeaconBalance:          big.NewInt(0),
			UserShareOfBeaconBalance:          big.NewInt(0),
			LastBondReductionTime:             big.NewInt(0),
			LastBondReductionPrevValue:        big.NewInt(0),
			LastBondReductionPrevNodeFee:      big.NewInt(0),
			ReduceBondTime:                    big.NewInt(0),
			ReduceBondValue:                   big.NewInt(0),
			PreMigrationBalance:               big.NewInt(0),
			Status:                            types.Staking,
		},
		{
			Exists:                            true,
			MinipoolAddress:                   mpAddrA2,
			Pubkey:                            pubkeyA2,
			NodeAddress:                       nodeAddrA,
			NodeFee:                           big.NewInt(1e17),
			NodeDepositBalance:                big.NewInt(8_000_000_000),
			UserDepositBalance:                big.NewInt(24_000_000_000),
			StatusTime:                        big.NewInt(1100),
			StatusBlock:                       big.NewInt(110),
			Balance:                           big.NewInt(0),
			DistributableBalance:              big.NewInt(0),
			NodeShareOfBalance:                big.NewInt(0),
			UserShareOfBalance:                big.NewInt(0),
			NodeRefundBalance:                 big.NewInt(0),
			PenaltyCount:                      big.NewInt(0),
			PenaltyRate:                       big.NewInt(0),
			UserDepositAssignedTime:           big.NewInt(1100),
			NodeShareOfBalanceIncludingBeacon: big.NewInt(0),
			UserShareOfBalanceIncludingBeacon: big.NewInt(0),
			NodeShareOfBeaconBalance:          big.NewInt(0),
			UserShareOfBeaconBalance:          big.NewInt(0),
			LastBondReductionTime:             big.NewInt(0),
			LastBondReductionPrevValue:        big.NewInt(0),
			LastBondReductionPrevNodeFee:      big.NewInt(0),
			ReduceBondTime:                    big.NewInt(0),
			ReduceBondValue:                   big.NewInt(0),
			PreMigrationBalance:               big.NewInt(0),
			Status:                            types.Staking,
		},
		{
			Exists:                            true,
			MinipoolAddress:                   mpAddrB1,
			Pubkey:                            pubkeyB1,
			NodeAddress:                       nodeAddrB,
			NodeFee:                           big.NewInt(5e16),
			NodeDepositBalance:                big.NewInt(16_000_000_000),
			UserDepositBalance:                big.NewInt(16_000_000_000),
			StatusTime:                        big.NewInt(2000),
			StatusBlock:                       big.NewInt(200),
			Balance:                           big.NewInt(0),
			DistributableBalance:              big.NewInt(0),
			NodeShareOfBalance:                big.NewInt(0),
			UserShareOfBalance:                big.NewInt(0),
			NodeRefundBalance:                 big.NewInt(0),
			PenaltyCount:                      big.NewInt(0),
			PenaltyRate:                       big.NewInt(0),
			UserDepositAssignedTime:           big.NewInt(2000),
			NodeShareOfBalanceIncludingBeacon: big.NewInt(0),
			UserShareOfBalanceIncludingBeacon: big.NewInt(0),
			NodeShareOfBeaconBalance:          big.NewInt(0),
			UserShareOfBeaconBalance:          big.NewInt(0),
			LastBondReductionTime:             big.NewInt(0),
			LastBondReductionPrevValue:        big.NewInt(0),
			LastBondReductionPrevNodeFee:      big.NewInt(0),
			ReduceBondTime:                    big.NewInt(0),
			ReduceBondValue:                   big.NewInt(0),
			PreMigrationBalance:               big.NewInt(0),
			Status:                            types.Staking,
		},
	}

	megapoolValidatorGlobalIndex := []megapool.ValidatorInfoFromGlobalIndex{
		{
			Pubkey:          megapoolPubkey[:],
			MegapoolAddress: megapoolAddrA,
			ValidatorId:     1,
			ValidatorInfo: megapool.ValidatorInfo{
				Staked: true,
			},
		},
	}

	nodeDetailsByAddress := map[common.Address]*rpstate.NativeNodeDetails{
		nodeAddrA: &nodeDetails[0],
		nodeAddrB: &nodeDetails[1],
	}

	minipoolDetailsByAddress := map[common.Address]*rpstate.NativeMinipoolDetails{
		mpAddrA1: &minipoolDetails[0],
		mpAddrA2: &minipoolDetails[1],
		mpAddrB1: &minipoolDetails[2],
	}

	minipoolDetailsByNode := map[common.Address][]*rpstate.NativeMinipoolDetails{
		nodeAddrA: {&minipoolDetails[0], &minipoolDetails[1]},
		nodeAddrB: {&minipoolDetails[2]},
	}

	megapoolValidatorsByAddress := map[common.Address][]*megapool.ValidatorInfoFromGlobalIndex{
		megapoolAddrA: {&megapoolValidatorGlobalIndex[0]},
	}

	megapoolDetails := map[common.Address]rpstate.NativeMegapoolDetails{
		megapoolAddrA: {
			Address:              megapoolAddrA,
			Deployed:             true,
			ActiveValidatorCount: 1,
			UserCapital:          big.NewInt(24_000_000_000),
			NodeBond:             big.NewInt(8_000_000_000),
			NodeDebt:             big.NewInt(0),
			RefundValue:          big.NewInt(0),
			AssignedValue:        big.NewInt(0),
			BondRequirement:      big.NewInt(0),
			EthBalance:           big.NewInt(0),
			PendingRewards:       big.NewInt(0),
			NodeQueuedBond:       big.NewInt(0),
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
			RplPrice:                          big.NewInt(1e16),
			MinCollateralFraction:             big.NewInt(1e17),
			MaxCollateralFraction:             big.NewInt(15e17),
			IntervalDuration:                  28 * 24 * time.Hour,
			NodeOperatorRewardsPercent:        big.NewInt(0),
			TrustedNodeOperatorRewardsPercent: big.NewInt(0),
			ProtocolDaoRewardsPercent:         big.NewInt(0),
		},
		NodeDetails:              nodeDetails,
		NodeDetailsByAddress:     nodeDetailsByAddress,
		MinipoolDetails:          minipoolDetails,
		MinipoolDetailsByAddress: minipoolDetailsByAddress,
		MinipoolDetailsByNode:    minipoolDetailsByNode,
		MinipoolValidatorDetails: ValidatorDetailsMap{
			pubkeyA1: {Pubkey: pubkeyA1, Index: "1", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
			pubkeyA2: {Pubkey: pubkeyA2, Index: "2", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
			pubkeyB1: {Pubkey: pubkeyB1, Index: "3", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
		},
		MegapoolValidatorDetails: ValidatorDetailsMap{
			megapoolPubkey: {Pubkey: megapoolPubkey, Index: "4", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
		},
		MegapoolValidatorGlobalIndex: megapoolValidatorGlobalIndex,
		MegapoolValidatorsByAddress:  megapoolValidatorsByAddress,
		MegapoolDetails:              megapoolDetails,
		OracleDaoMemberDetails: []rpstate.OracleDaoMemberDetails{
			{
				Address:          nodeAddrA,
				Exists:           true,
				ID:               "odao-a",
				RPLBondAmount:    big.NewInt(1000),
				JoinedTime:       time.Unix(1000, 0),
				LastProposalTime: time.Unix(1000, 0),
			},
		},
		ProtocolDaoProposalDetails: []protocol.ProtocolDaoProposalDetails{
			{
				ID:                   1,
				ProposerAddress:      nodeAddrA,
				VotingPowerRequired:  big.NewInt(100),
				VotingPowerFor:       big.NewInt(80),
				VotingPowerAgainst:   big.NewInt(10),
				VotingPowerAbstained: big.NewInt(10),
			},
		},
	}
}

// TestNetworkStateJSONRoundtrip verifies that marshaling and unmarshaling
// a NetworkState preserves all serialized fields and correctly rebuilds
// the index maps (NodeDetailsByAddress, MinipoolDetailsByAddress,
// MinipoolDetailsByNode) that are excluded from JSON.
func TestNetworkStateJSONRoundtrip(t *testing.T) {
	original := buildTestState()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored NetworkState
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
	if len(restored.MegapoolValidatorGlobalIndex) != len(original.MegapoolValidatorGlobalIndex) {
		t.Errorf("MegapoolValidatorGlobalIndex count: got %d, want %d",
			len(restored.MegapoolValidatorGlobalIndex), len(original.MegapoolValidatorGlobalIndex))
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
			RplPrice:                          big.NewInt(0),
			MinCollateralFraction:             big.NewInt(0),
			MaxCollateralFraction:             big.NewInt(0),
			NodeOperatorRewardsPercent:        big.NewInt(0),
			TrustedNodeOperatorRewardsPercent: big.NewInt(0),
			ProtocolDaoRewardsPercent:         big.NewInt(0),
		},
		NodeDetails: []rpstate.NativeNodeDetails{
			{NodeAddress: addr, RegistrationTime: big.NewInt(0), RewardNetwork: big.NewInt(0), LegacyStakedRPL: big.NewInt(0), EffectiveRPLStake: big.NewInt(0), MinimumRPLStake: big.NewInt(0), MaximumRPLStake: big.NewInt(0), EthBorrowed: big.NewInt(0), EthBorrowedLimit: big.NewInt(0), MegapoolETHBorrowed: big.NewInt(0), MinipoolETHBorrowed: big.NewInt(0), EthBonded: big.NewInt(0), MegapoolEthBonded: big.NewInt(0), MinipoolETHBonded: big.NewInt(0), MegapoolStakedRPL: big.NewInt(0), UnstakingRPL: big.NewInt(0), LockedRPL: big.NewInt(0), MinipoolCount: big.NewInt(0), BalanceETH: big.NewInt(0), BalanceRETH: big.NewInt(0), BalanceRPL: big.NewInt(0), BalanceOldRPL: big.NewInt(0), DepositCreditBalance: big.NewInt(0), DistributorBalanceUserETH: big.NewInt(0), DistributorBalanceNodeETH: big.NewInt(0), SmoothingPoolRegistrationChanged: big.NewInt(0), AverageNodeFee: big.NewInt(0), CollateralisationRatio: big.NewInt(0), DistributorBalance: big.NewInt(0)},
			{NodeAddress: addr, RegistrationTime: big.NewInt(0), RewardNetwork: big.NewInt(0), LegacyStakedRPL: big.NewInt(0), EffectiveRPLStake: big.NewInt(0), MinimumRPLStake: big.NewInt(0), MaximumRPLStake: big.NewInt(0), EthBorrowed: big.NewInt(0), EthBorrowedLimit: big.NewInt(0), MegapoolETHBorrowed: big.NewInt(0), MinipoolETHBorrowed: big.NewInt(0), EthBonded: big.NewInt(0), MegapoolEthBonded: big.NewInt(0), MinipoolETHBonded: big.NewInt(0), MegapoolStakedRPL: big.NewInt(0), UnstakingRPL: big.NewInt(0), LockedRPL: big.NewInt(0), MinipoolCount: big.NewInt(0), BalanceETH: big.NewInt(0), BalanceRETH: big.NewInt(0), BalanceRPL: big.NewInt(0), BalanceOldRPL: big.NewInt(0), DepositCreditBalance: big.NewInt(0), DistributorBalanceUserETH: big.NewInt(0), DistributorBalanceNodeETH: big.NewInt(0), SmoothingPoolRegistrationChanged: big.NewInt(0), AverageNodeFee: big.NewInt(0), CollateralisationRatio: big.NewInt(0), DistributorBalance: big.NewInt(0)},
		},
		MinipoolDetails:          []rpstate.NativeMinipoolDetails{},
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

	state := &NetworkState{
		BeaconConfig: beacon.Eth2Config{},
		NetworkDetails: &rpstate.NetworkDetails{
			RplPrice:                          zeroInt(),
			MinCollateralFraction:             zeroInt(),
			MaxCollateralFraction:             zeroInt(),
			NodeOperatorRewardsPercent:        zeroInt(),
			TrustedNodeOperatorRewardsPercent: zeroInt(),
			ProtocolDaoRewardsPercent:         zeroInt(),
		},
		NodeDetails: []rpstate.NativeNodeDetails{
			{NodeAddress: nodeAddr, RegistrationTime: zeroInt(), RewardNetwork: zeroInt(), LegacyStakedRPL: zeroInt(), EffectiveRPLStake: zeroInt(), MinimumRPLStake: zeroInt(), MaximumRPLStake: zeroInt(), EthBorrowed: zeroInt(), EthBorrowedLimit: zeroInt(), MegapoolETHBorrowed: zeroInt(), MinipoolETHBorrowed: zeroInt(), EthBonded: zeroInt(), MegapoolEthBonded: zeroInt(), MinipoolETHBonded: zeroInt(), MegapoolStakedRPL: zeroInt(), UnstakingRPL: zeroInt(), LockedRPL: zeroInt(), MinipoolCount: zeroInt(), BalanceETH: zeroInt(), BalanceRETH: zeroInt(), BalanceRPL: zeroInt(), BalanceOldRPL: zeroInt(), DepositCreditBalance: zeroInt(), DistributorBalanceUserETH: zeroInt(), DistributorBalanceNodeETH: zeroInt(), SmoothingPoolRegistrationChanged: zeroInt(), AverageNodeFee: zeroInt(), CollateralisationRatio: zeroInt(), DistributorBalance: zeroInt()},
		},
		MinipoolDetails: []rpstate.NativeMinipoolDetails{
			{Exists: true, MinipoolAddress: mpAddr, Pubkey: pk1, NodeAddress: nodeAddr, NodeFee: zeroInt(), NodeDepositBalance: zeroInt(), UserDepositBalance: zeroInt(), StatusTime: zeroInt(), StatusBlock: zeroInt(), Balance: zeroInt(), DistributableBalance: zeroInt(), NodeShareOfBalance: zeroInt(), UserShareOfBalance: zeroInt(), NodeRefundBalance: zeroInt(), PenaltyCount: zeroInt(), PenaltyRate: zeroInt(), UserDepositAssignedTime: zeroInt(), NodeShareOfBalanceIncludingBeacon: zeroInt(), UserShareOfBalanceIncludingBeacon: zeroInt(), NodeShareOfBeaconBalance: zeroInt(), UserShareOfBeaconBalance: zeroInt(), LastBondReductionTime: zeroInt(), LastBondReductionPrevValue: zeroInt(), LastBondReductionPrevNodeFee: zeroInt(), ReduceBondTime: zeroInt(), ReduceBondValue: zeroInt(), PreMigrationBalance: zeroInt()},
			{Exists: true, MinipoolAddress: mpAddr, Pubkey: pk2, NodeAddress: nodeAddr, NodeFee: zeroInt(), NodeDepositBalance: zeroInt(), UserDepositBalance: zeroInt(), StatusTime: zeroInt(), StatusBlock: zeroInt(), Balance: zeroInt(), DistributableBalance: zeroInt(), NodeShareOfBalance: zeroInt(), UserShareOfBalance: zeroInt(), NodeRefundBalance: zeroInt(), PenaltyCount: zeroInt(), PenaltyRate: zeroInt(), UserDepositAssignedTime: zeroInt(), NodeShareOfBalanceIncludingBeacon: zeroInt(), UserShareOfBalanceIncludingBeacon: zeroInt(), NodeShareOfBeaconBalance: zeroInt(), UserShareOfBeaconBalance: zeroInt(), LastBondReductionTime: zeroInt(), LastBondReductionPrevValue: zeroInt(), LastBondReductionPrevNodeFee: zeroInt(), ReduceBondTime: zeroInt(), ReduceBondValue: zeroInt(), PreMigrationBalance: zeroInt()},
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

// buildDuplicatePubkeyState returns a state in which two megapools have registered the same
// pubkey. RocketMegapoolManager only enforces pubkey uniqueness per megapool, so this is a
// state the contracts genuinely permit: megapool A owns and has staked the validator, while
// megapool B deposited using the same pubkey and then dequeued. B is later in the global
// index, so any pubkey-keyed lookup would return B's record for both megapools.
func buildDuplicatePubkeyState() (*NetworkState, common.Address, common.Address, types.ValidatorPubkey) {
	megapoolA := common.HexToAddress("0xAAAA000000000000000000000000000000000001")
	megapoolB := common.HexToAddress("0xBBBB000000000000000000000000000000000002")
	sharedPubkey := newTestPubkey(0x42)

	globalIndex := []megapool.ValidatorInfoFromGlobalIndex{
		{
			Pubkey:          sharedPubkey[:],
			MegapoolAddress: megapoolA,
			ValidatorId:     0,
			ValidatorInfo: megapool.ValidatorInfo{
				Staked:       true,
				DepositValue: 32000,
			},
		},
		{
			Pubkey:          sharedPubkey[:],
			MegapoolAddress: megapoolB,
			ValidatorId:     0,
			ValidatorInfo: megapool.ValidatorInfo{
				Staked:             false,
				InPrestake:         true,
				LastRequestedValue: 32000,
				LastRequestedBond:  4000,
			},
		},
	}

	zeroInt := func() *big.Int { return big.NewInt(0) }
	state := &NetworkState{
		BeaconConfig: beacon.Eth2Config{},
		NetworkDetails: &rpstate.NetworkDetails{
			RplPrice:                          zeroInt(),
			MinCollateralFraction:             zeroInt(),
			MaxCollateralFraction:             zeroInt(),
			NodeOperatorRewardsPercent:        zeroInt(),
			TrustedNodeOperatorRewardsPercent: zeroInt(),
			ProtocolDaoRewardsPercent:         zeroInt(),
		},
		NodeDetails:              []rpstate.NativeNodeDetails{},
		MinipoolDetails:          []rpstate.NativeMinipoolDetails{},
		MinipoolValidatorDetails: ValidatorDetailsMap{},
		MegapoolValidatorDetails: ValidatorDetailsMap{
			sharedPubkey: {Pubkey: sharedPubkey, Index: "7", Exists: true, Balance: 32000000000, ActivationEpoch: 0, ExitEpoch: ^uint64(0)},
		},
		MegapoolValidatorGlobalIndex: globalIndex,
		MegapoolDetails: map[common.Address]rpstate.NativeMegapoolDetails{
			megapoolA: {Address: megapoolA, Deployed: true, ActiveValidatorCount: 1, UserCapital: big.NewInt(28_000_000_000), NodeBond: big.NewInt(4_000_000_000), NodeDebt: zeroInt(), RefundValue: zeroInt(), AssignedValue: zeroInt(), BondRequirement: zeroInt(), EthBalance: zeroInt(), PendingRewards: zeroInt(), NodeQueuedBond: zeroInt()},
			megapoolB: {Address: megapoolB, Deployed: true, ActiveValidatorCount: 0, UserCapital: big.NewInt(28_000_000_000), NodeBond: big.NewInt(4_000_000_000), NodeDebt: zeroInt(), RefundValue: zeroInt(), AssignedValue: zeroInt(), BondRequirement: zeroInt(), EthBalance: zeroInt(), PendingRewards: zeroInt(), NodeQueuedBond: zeroInt()},
		},
	}

	state.MegapoolValidatorsByAddress = make(map[common.Address][]*megapool.ValidatorInfoFromGlobalIndex)
	for i := range state.MegapoolValidatorGlobalIndex {
		v := &state.MegapoolValidatorGlobalIndex[i]
		state.MegapoolValidatorsByAddress[v.MegapoolAddress] = append(state.MegapoolValidatorsByAddress[v.MegapoolAddress], v)
	}

	return state, megapoolA, megapoolB, sharedPubkey
}

// assertDuplicatePubkeyIsolation checks that each megapool sees its own validator record and
// not the other's.
func assertDuplicatePubkeyIsolation(t *testing.T, ns *NetworkState, megapoolA, megapoolB common.Address) {
	t.Helper()

	validatorsA := ns.MegapoolValidatorsByAddress[megapoolA]
	if len(validatorsA) != 1 {
		t.Fatalf("megapool A: got %d validators, want 1", len(validatorsA))
	}
	if !validatorsA[0].ValidatorInfo.Staked {
		t.Error("megapool A's validator reads as not staked; it was overwritten by megapool B's record")
	}
	if validatorsA[0].MegapoolAddress != megapoolA {
		t.Errorf("megapool A's validator is owned by %s", validatorsA[0].MegapoolAddress.Hex())
	}

	validatorsB := ns.MegapoolValidatorsByAddress[megapoolB]
	if len(validatorsB) != 1 {
		t.Fatalf("megapool B: got %d validators, want 1", len(validatorsB))
	}
	if validatorsB[0].ValidatorInfo.Staked {
		t.Error("megapool B's validator reads as staked; it was overwritten by megapool A's record")
	}
	if validatorsB[0].MegapoolAddress != megapoolB {
		t.Errorf("megapool B's validator is owned by %s", validatorsB[0].MegapoolAddress.Hex())
	}
}

// TestDuplicatePubkeyAcrossMegapools verifies that a pubkey registered by two megapools does
// not cause one megapool's validator record to shadow the other's. Previously the state kept a
// single pubkey-keyed map, so the last global index entry won and the victim's validator
// appeared unstaked - which zeroed its Beacon balance in the rETH balance submission.
func TestDuplicatePubkeyAcrossMegapools(t *testing.T) {
	ns, megapoolA, megapoolB, _ := buildDuplicatePubkeyState()
	assertDuplicatePubkeyIsolation(t, ns, megapoolA, megapoolB)
}

// TestDuplicatePubkeyAcrossMegapoolsRoundtrip verifies the same isolation survives a JSON
// round-trip, since UnmarshalJSON rebuilds the index independently of createNetworkState.
func TestDuplicatePubkeyAcrossMegapoolsRoundtrip(t *testing.T) {
	original, megapoolA, megapoolB, _ := buildDuplicatePubkeyState()

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored NetworkState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	assertDuplicatePubkeyIsolation(t, &restored, megapoolA, megapoolB)

	// Entries must alias the restored global index, not a copy of it
	for addr, validators := range restored.MegapoolValidatorsByAddress {
		for _, v := range validators {
			found := false
			for i := range restored.MegapoolValidatorGlobalIndex {
				if &restored.MegapoolValidatorGlobalIndex[i] == v {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("MegapoolValidatorsByAddress[%s] entry does not point into the global index", addr.Hex())
			}
		}
	}
}

// TestGetMegapoolEligibleBorrowedEthWithDuplicatePubkey verifies that megapool B's in-prestake
// validator does not reduce megapool A's eligible borrowed ETH. A's validator is staked, not in
// prestake, so A's eligible borrowed ETH must be its full user capital.
func TestGetMegapoolEligibleBorrowedEthWithDuplicatePubkey(t *testing.T) {
	ns, megapoolA, megapoolB, _ := buildDuplicatePubkeyState()

	nodeA := &rpstate.NativeNodeDetails{MegapoolDeployed: true, MegapoolAddress: megapoolA}
	gotA := ns.GetMegapoolEligibleBorrowedEth(nodeA)
	wantA := ns.MegapoolDetails[megapoolA].UserCapital
	if gotA.Cmp(wantA) != 0 {
		t.Errorf("megapool A eligible borrowed ETH: got %s, want %s (B's prestake validator leaked into A)", gotA, wantA)
	}

	// B's own validator is in prestake, so B's eligible borrowed ETH is reduced by the user
	// portion of the requested value: 32000 - 4000 = 28000 milliEth.
	nodeB := &rpstate.NativeNodeDetails{MegapoolDeployed: true, MegapoolAddress: megapoolB}
	gotB := ns.GetMegapoolEligibleBorrowedEth(nodeB)
	if gotB.Cmp(gotA) >= 0 {
		t.Errorf("megapool B eligible borrowed ETH (%s) should be below A's (%s); B's own prestake validator was not applied", gotB, gotA)
	}
}
