package megapool

import (
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/contracts"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

func performanceContractABI(t *testing.T) (abi.ABI, string) {
	t.Helper()
	raw, err := os.ReadFile("testdata/performance-v1.5-dev.abi.json")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := abi.JSON(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	return parsed, string(raw)
}

func TestDefensePayloadMatchesContractABI(t *testing.T) {
	contractABI, _ := performanceContractABI(t)
	for _, mini := range []bool{false, true} {
		for _, participation := range []bool{false, true} {
			d := PerformanceChallengeDefense{
				ChallengeId:      new(big.Int).Lsh(big.NewInt(1), 100),
				ValidatorId:      9,
				SlotTimestamp:    123456,
				Validator:        ValidatorProof{ValidatorIndex: big.NewInt(567)},
				Slot:             SlotProof{Slot: 98765},
				Offset:           257,
				ChallengeLeaf:    big.NewInt(2),
				ChallengeWitness: []common.Hash{{1}, {2}},
			}
			pool := "Megapool"
			if mini {
				d.MinipoolAddress = common.HexToAddress("0x1234")
				pool = "Minipool"
			}
			suffix := "Validator"
			if participation {
				d.Participation = &ParticipationProof{ValidatorIndex: big.NewInt(567)}
				suffix = "Participation"
			}
			method, args := d.call()
			expected := "respondWith" + pool + suffix
			if method != expected {
				t.Fatalf("got %s, want %s", method, expected)
			}
			payload, err := contractABI.Pack(method, args...)
			if err != nil {
				t.Fatalf("%s: %v", method, err)
			}
			decoded, err := contractABI.Methods[method].Inputs.Unpack(payload[4:])
			if err != nil {
				t.Fatal(err)
			}
			if decoded[0].(*big.Int).Cmp(d.ChallengeId) != 0 {
				t.Fatal("truncated challenge ID")
			}
			if mini && decoded[1].(common.Address) != d.MinipoolAddress {
				t.Fatal("incorrect minipool member")
			}
			if !mini && decoded[1].(uint32) != d.ValidatorId {
				t.Fatal("incorrect megapool member")
			}
			if participation && (decoded[2].(uint64) != d.Offset || decoded[3].(*big.Int).Cmp(d.ChallengeLeaf) != 0) {
				t.Fatal("incorrect challenge leaf or offset")
			}
		}
	}
}

func TestDiscoverBothPerformanceChallengeTypes(t *testing.T) {
	networkABI, raw := performanceContractABI(t)
	storageABI, err := abi.JSON(strings.NewReader(contracts.RocketStorageABI))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := rocketpool.EncodeAbiStr(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, mini := range []bool{false, true} {
		name := "MegapoolChallenged"
		var members interface{} = []uint32{0, 31}
		if mini {
			name, members = "MinipoolsChallenged", []common.Address{common.HexToAddress("0x1234"), common.HexToAddress("0x5678")}
		}
		event := networkABI.Events[name]
		data, err := event.Inputs.NonIndexed().Pack(members, uint64(50), common.Hash{}, []*big.Int{big.NewInt(3)})
		if err != nil {
			t.Fatal(err)
		}
		id, owner := new(big.Int).Lsh(big.NewInt(1), 100), common.HexToAddress("0xabcd")
		client := &challengeStatusClient{
			storageABI:       storageABI,
			participationABI: networkABI,
			encodedABI:       encoded,
			challengeLog: types.Log{
				Address: common.HexToAddress("0x05"),
				Topics:  []common.Hash{event.ID, common.BytesToHash(owner.Bytes()), common.BigToHash(id)},
				Data:    data,
			},
		}
		rp, err := rocketpool.NewRocketPool(client, common.HexToAddress("0x01"))
		if err != nil {
			t.Fatal(err)
		}
		got, err := GetPerformanceChallenges(rp, id, big.NewInt(100), big.NewInt(700), big.NewInt(777), &bind.CallOpts{BlockNumber: big.NewInt(777)})
		if err != nil || len(got) != 1 {
			t.Fatalf("%s: %v, %v", name, got, err)
		}
		if got[0].ChallengeId.Cmp(id) != 0 || got[0].StartEpoch != 50 || got[0].Participation[0].Int64() != 3 {
			t.Fatal("incorrect event payload")
		}
		if mini && (got[0].NodeAddress != owner || !reflect.DeepEqual(got[0].MinipoolAddresses, members)) {
			t.Fatal("incorrect minipool members")
		}
		if !mini && (got[0].MegapoolAddress != owner || !reflect.DeepEqual(got[0].ValidatorIds, members)) {
			t.Fatal("incorrect megapool members")
		}
		query := client.queries[1]
		if len(query.Topics[0]) != 2 || len(query.Topics[1]) != 0 || query.Topics[2][0] != common.BigToHash(id) {
			t.Fatalf("incorrect ID/topic filter: %v", query.Topics)
		}
		if !reflect.DeepEqual(query.Addresses, []common.Address{common.HexToAddress("0x02"), common.HexToAddress("0x05")}) {
			t.Fatal("lost previous contract address")
		}
	}
}

func TestPerformanceChallengeDeadlineAndSettlement(t *testing.T) {
	proposer, defender := common.HexToAddress("0x01"), common.HexToAddress("0x02")
	s := PerformanceChallengeStatus{Proposer: proposer, ResponseDeadline: big.NewInt(100)}
	if !s.CanDefend(defender, 100) || s.CanDefend(proposer, 100) || s.CanFinalise(100) {
		t.Fatal("incorrect exact deadline or proposer rule")
	}
	if s.CanDefend(defender, 101) || !s.CanFinalise(101) {
		t.Fatal("incorrect expired rules")
	}
	s.BondSettled = true
	if !s.CanFinalise(101) {
		t.Fatal("releasing bond must not block finalisation")
	}
	s.BondSettled, s.Finalised = false, true
	if !s.CanReleaseBond(101) || s.CanFinalise(101) || s.CanDefend(defender, 99) {
		t.Fatal("finalisation must not block bond release or allow repeat actions")
	}
	s.Finalised, s.Responded, s.Responder = false, true, defender
	if s.CanFinalise(101) || s.CanReleaseBond(101) || s.CanDefend(defender, 99) || !s.CanClaimReward(defender) {
		t.Fatal("incorrect defeated challenge rules")
	}
}
