package megapool

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/contracts"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

const performanceChallengeEventABI = `[{"type":"event","name":"MegapoolChallenged","inputs":[
 {"name":"megapoolAddress","type":"address","indexed":true},
 {"name":"_validatorIds","type":"uint32[]"},
 {"name":"_challengeId","type":"uint256","indexed":true},
 {"name":"_startEpoch","type":"uint64"},
 {"name":"_root","type":"bytes32"},
 {"name":"_participation","type":"uint256[]"}]}]`

func TestDecodePerformanceChallenge(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(performanceChallengeEventABI))
	if err != nil {
		t.Fatal(err)
	}
	event := parsed.Events["MegapoolChallenged"]
	pool := common.HexToAddress("0x1234")
	id := new(big.Int).Lsh(big.NewInt(1), 100) // Preserve Solidity uint256 IDs.
	participation := []*big.Int{big.NewInt(5), new(big.Int).Lsh(big.NewInt(1), 255)}
	data, err := event.Inputs.NonIndexed().Pack([]uint32{3, 9}, uint64(105000), common.Hash{}, participation)
	if err != nil {
		t.Fatal(err)
	}
	entry := types.Log{Topics: []common.Hash{event.ID, common.BytesToHash(pool.Bytes()), common.BigToHash(id)}, Data: data}
	got, err := decodePerformanceChallenge(event, entry)
	if err != nil {
		t.Fatal(err)
	}
	if got.ChallengeId.Cmp(id) != 0 || got.MegapoolAddress != pool || got.StartEpoch != 105000 || !reflect.DeepEqual(got.ValidatorIds, []uint32{3, 9}) || !reflect.DeepEqual(got.Participation, participation) {
		t.Fatalf("event data not preserved: %+v", got)
	}
	entry.Topics = entry.Topics[:2]
	if _, err := decodePerformanceChallenge(event, entry); err == nil {
		t.Fatal("accepted event without indexed challenge ID")
	}
	entry.Topics = append(entry.Topics, common.BigToHash(id))
	entry.Data = []byte{1}
	if _, err := decodePerformanceChallenge(event, entry); err == nil {
		t.Fatal("accepted truncated event data")
	}
}

type challengeStatusClient struct {
	rocketpool.ExecutionClient
	storageABI, participationABI abi.ABI
	encodedABI                   string
	challengeId                  *big.Int
	responded, finalised         bool
	queries                      []ethereum.FilterQuery
	challengeLog                 types.Log
}

const challengeUpgradeEventABI = `[{"type":"event","name":"ContractUpgraded","inputs":[{"name":"name","type":"bytes32","indexed":true},{"name":"oldAddress","type":"address","indexed":true},{"name":"newAddress","type":"address","indexed":true},{"name":"time","type":"uint256"}]}]`

func (c *challengeStatusClient) FilterLogs(_ context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	c.queries = append(c.queries, query)
	if len(query.Addresses) == 1 && query.Addresses[0] == common.HexToAddress("0x04") {
		return []types.Log{{Topics: []common.Hash{query.Topics[0][0], query.Topics[1][0], common.BigToHash(big.NewInt(5)), common.BigToHash(big.NewInt(2))}}}, nil
	}
	return []types.Log{c.challengeLog}, nil
}

func (c *challengeStatusClient) CallContract(_ context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if block == nil || block.Int64() != 777 {
		return nil, fmt.Errorf("challenge status must be pinned to block 777")
	}
	if *call.To == common.HexToAddress("0x02") {
		method, err := c.participationABI.MethodById(call.Data)
		if err != nil {
			return nil, err
		}
		args, err := method.Inputs.Unpack(call.Data[4:])
		if err != nil || args[0].(*big.Int).Cmp(c.challengeId) != 0 {
			return nil, fmt.Errorf("incorrect challenge ID: %v", err)
		}
		// A released/settled bond does not mean the exits have been finalised.
		return method.Outputs.Pack(common.HexToAddress("0x03"), common.Address{}, big.NewInt(100), big.NewInt(123456), true)
	}
	method, err := c.storageABI.MethodById(call.Data)
	if err != nil {
		return nil, err
	}
	args, err := method.Inputs.Unpack(call.Data[4:])
	if err != nil {
		return nil, err
	}
	key := common.Hash(args[0].([32]byte))
	switch {
	case method.Name == "getAddress" && key == crypto.Keccak256Hash([]byte("contract.addressrocketDAONodeTrustedUpgrade")):
		return method.Outputs.Pack(common.HexToAddress("0x04"))
	case method.Name == "getString" && key == crypto.Keccak256Hash([]byte("contract.abirocketDAONodeTrustedUpgrade")):
		encoded, err := rocketpool.EncodeAbiStr(challengeUpgradeEventABI)
		if err != nil {
			return nil, err
		}
		return method.Outputs.Pack(encoded)
	case method.Name == "getAddress" && key == crypto.Keccak256Hash([]byte("contract.addressrocketNetworkParticipation")):
		return method.Outputs.Pack(common.HexToAddress("0x02"))
	case method.Name == "getString" && key == crypto.Keccak256Hash([]byte("contract.abirocketNetworkParticipation")):
		return method.Outputs.Pack(c.encodedABI)
	case method.Name == "getBool" && key == crypto.Keccak256Hash([]byte("participation.challenge.responded"), common.LeftPadBytes(c.challengeId.Bytes(), 32)):
		return method.Outputs.Pack(c.responded)
	case method.Name == "getBool" && key == crypto.Keccak256Hash([]byte("participation.challenge.finalised"), common.LeftPadBytes(c.challengeId.Bytes(), 32)):
		return method.Outputs.Pack(c.finalised)
	default:
		return nil, fmt.Errorf("unexpected storage call %s %s", method.Name, key)
	}
}

func TestChallengeEventsIncludePreviousContractAndBoundedUpgradeScan(t *testing.T) {
	networkABI, err := abi.JSON(strings.NewReader(performanceChallengeEventABI))
	if err != nil {
		t.Fatal(err)
	}
	storageABI, err := abi.JSON(strings.NewReader(contracts.RocketStorageABI))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := rocketpool.EncodeAbiStr(performanceChallengeEventABI)
	if err != nil {
		t.Fatal(err)
	}
	pool := common.HexToAddress("0x1234")
	event := networkABI.Events["MegapoolChallenged"]
	data, err := event.Inputs.NonIndexed().Pack([]uint32{1, 2}, uint64(50), common.Hash{}, []*big.Int{big.NewInt(3)})
	if err != nil {
		t.Fatal(err)
	}
	client := &challengeStatusClient{
		storageABI: storageABI, participationABI: networkABI, encodedABI: encoded,
		challengeLog: types.Log{Address: common.HexToAddress("0x05"), Topics: []common.Hash{event.ID, common.BytesToHash(pool.Bytes()), common.BigToHash(big.NewInt(9))}, Data: data},
	}
	rp, err := rocketpool.NewRocketPool(client, common.HexToAddress("0x01"))
	if err != nil {
		t.Fatal(err)
	}
	challenges, err := GetMegapoolPerformanceChallenges(rp, pool, big.NewInt(100), big.NewInt(700), big.NewInt(777), &bind.CallOpts{BlockNumber: big.NewInt(777)})
	if err != nil || len(challenges) != 1 || challenges[0].ChallengeId.Int64() != 9 {
		t.Fatalf("did not discover old contract event: %v, %v", challenges, err)
	}
	if len(client.queries) != 2 {
		t.Fatalf("expected one upgrade scan and one challenge scan, got %d", len(client.queries))
	}
	for _, query := range client.queries {
		if query.FromBlock.Int64() != 700 || query.ToBlock.Int64() != 777 {
			t.Fatalf("scan escaped incremental bounds: %+v", query)
		}
	}
	query := client.queries[1]
	if !reflect.DeepEqual(query.Addresses, []common.Address{common.HexToAddress("0x02"), common.HexToAddress("0x05")}) || query.Topics[1][0] != common.BytesToHash(pool.Bytes()) {
		t.Fatalf("incorrect contract or megapool filter: %+v", query)
	}
}

func TestGetPerformanceChallengeStatus(t *testing.T) {
	const bondABI = `[{"name":"getChallengeBondDetails","type":"function","stateMutability":"view","inputs":[{"type":"uint256"}],"outputs":[{"type":"address"},{"type":"address"},{"type":"uint256"},{"type":"uint256"},{"type":"bool"}]}]`
	parsed, err := abi.JSON(strings.NewReader(bondABI))
	if err != nil {
		t.Fatal(err)
	}
	storageABI, err := abi.JSON(strings.NewReader(contracts.RocketStorageABI))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := rocketpool.EncodeAbiStr(bondABI)
	if err != nil {
		t.Fatal(err)
	}
	client := &challengeStatusClient{storageABI: storageABI, participationABI: parsed, encodedABI: encoded, challengeId: new(big.Int).Lsh(big.NewInt(1), 100)}
	rp, err := rocketpool.NewRocketPool(client, common.HexToAddress("0x01"))
	if err != nil {
		t.Fatal(err)
	}
	for _, flags := range [][2]bool{{false, false}, {true, false}, {false, true}} {
		client.responded, client.finalised = flags[0], flags[1]
		status, err := GetPerformanceChallengeStatus(rp, client.challengeId, &bind.CallOpts{BlockNumber: big.NewInt(777)})
		if err != nil {
			t.Fatal(err)
		}
		if status.Proposer != common.HexToAddress("0x03") || status.ResponseDeadline.Int64() != 123456 || status.Responded != flags[0] || status.Finalised != flags[1] {
			t.Fatalf("incorrect status: %+v", status)
		}
	}
}
