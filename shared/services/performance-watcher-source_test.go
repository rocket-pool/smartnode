package services

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/contracts"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
)

type watcherOwnershipClient struct {
	rocketpool.ExecutionClient
	responses map[string][]byte
}

func (c *watcherOwnershipClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if number == nil || number.Int64() != -3 {
		return nil, fmt.Errorf("ownership must use finalized execution state")
	}
	return &types.Header{Number: big.NewInt(123)}, nil
}

func (c *watcherOwnershipClient) CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, ok := ctx.Deadline(); !ok || block == nil || block.Uint64() != 123 {
		return nil, fmt.Errorf("ownership read is not bounded and pinned to finalized block")
	}
	key := call.To.Hex() + string(call.Data)
	if result, ok := c.responses[key]; ok {
		return result, nil
	}
	return nil, fmt.Errorf("unexpected ownership RPC to %s: %x", call.To.Hex(), call.Data)
}

func TestWatcherNodePoolOwnershipBeforeSaturn2(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		count                int
		factory, failFactory bool
	}{
		{"legacy minipools", 2, false, false},
		{"minipools and megapool with bounded batches", 101, true, false},
		{"empty node", 0, false, false},
		{"ownership RPC failure", 1, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage, manager, factory, multi, owner := common.HexToAddress("0x01"), common.HexToAddress("0x02"), common.HexToAddress("0x03"), common.HexToAddress("0x04"), common.HexToAddress("0x05")
			mega := common.HexToAddress("0x06")
			client := &watcherOwnershipClient{responses: map[string][]byte{}}
			parseABI := func(raw string) abi.ABI {
				parsed, err := abi.JSON(strings.NewReader(raw))
				if err != nil {
					t.Fatal(err)
				}
				return parsed
			}
			storageABI := parseABI(contracts.RocketStorageABI)
			managerRaw := `[{"type":"function","name":"getNodeMinipoolCount","inputs":[{"type":"address"}],"outputs":[{"type":"uint256"}]},{"type":"function","name":"getNodeMinipoolAt","inputs":[{"type":"address"},{"type":"uint256"}],"outputs":[{"type":"address"}]}]`
			factoryRaw := `[{"type":"function","name":"getExpectedAddress","inputs":[{"type":"address"}],"outputs":[{"type":"address"}]}]`
			managerABI, factoryABI, multiABI := parseABI(managerRaw), parseABI(factoryRaw), parseABI(multicall.MulticallABI)
			pack := func(a abi.ABI, method string, args ...any) []byte {
				data, err := a.Pack(method, args...)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			output := func(a abi.ABI, method string, values ...any) []byte {
				data, err := a.Methods[method].Outputs.Pack(values...)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			add := func(address common.Address, data, result []byte) {
				client.responses[address.Hex()+string(data)] = result
			}
			contract := func(name string, address common.Address, raw string) {
				add(storage, pack(storageABI, "getAddress", crypto.Keccak256Hash([]byte("contract.address"+name))), output(storageABI, "getAddress", address))
				encoded, err := rocketpool.EncodeAbiStr(raw)
				if err != nil {
					t.Fatal(err)
				}
				add(storage, pack(storageABI, "getString", crypto.Keccak256Hash([]byte("contract.abi"+name))), output(storageABI, "getString", encoded))
			}
			contract("rocketMinipoolManager", manager, managerRaw)
			add(manager, pack(managerABI, "getNodeMinipoolCount", owner), output(managerABI, "getNodeMinipoolCount", big.NewInt(int64(tc.count))))
			expected := []string{}
			for start := 0; start < tc.count; start += 100 {
				calls := []multicall.MultiCall{}
				results := []struct {
					Success    bool
					ReturnData []byte
				}{}
				for i := start; i < min(start+100, tc.count); i++ {
					pool := common.BigToAddress(big.NewInt(int64(1000 + i)))
					expected = append(expected, pool.Hex())
					calls = append(calls, multicall.MultiCall{Target: manager, CallData: pack(managerABI, "getNodeMinipoolAt", owner, big.NewInt(int64(i)))})
					results = append(results, struct {
						Success    bool
						ReturnData []byte
					}{true, output(managerABI, "getNodeMinipoolAt", pool)})
				}
				add(multi, pack(multiABI, "tryAggregate", true, calls), output(multiABI, "tryAggregate", results))
			}
			if tc.factory {
				contract("rocketMegapoolFactory", factory, factoryRaw)
				if !tc.failFactory {
					add(factory, pack(factoryABI, "getExpectedAddress", owner), output(factoryABI, "getExpectedAddress", mega))
				}
				expected = append(expected, mega.Hex())
			} else {
				add(storage, pack(storageABI, "getAddress", crypto.Keccak256Hash([]byte("contract.addressrocketMegapoolFactory"))), output(storageABI, "getAddress", common.Address{}))
			}
			source := &watcherSource{ec: client, storage: storage, multicall: multi}
			pools, err := source.NodePools(context.Background(), owner.Hex())
			if tc.failFactory {
				if err == nil {
					t.Fatal("ownership RPC failure returned a partial report")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(pools, expected) {
				t.Fatalf("ownership mismatch: %v %v, want %v", pools, err, expected)
			}
		})
	}
}
