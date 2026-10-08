package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	statebindings "github.com/rocket-pool/smartnode/bindings/utils/state"
	watcher "github.com/rocket-pool/smartnode/shared/services/performancewatcher"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

// Some legacy binding helpers use context.Background. A private contract
// manager ensures those calls still share the worker/request cancellation.
type watcherExecutionClient struct {
	rocketpool.ExecutionClient
	ctx context.Context
}

func (e watcherExecutionClient) CallContract(_ context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()
	return e.ExecutionClient.CallContract(ctx, call, block)
}
func (e watcherExecutionClient) CodeAt(_ context.Context, address common.Address, block *big.Int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()
	return e.ExecutionClient.CodeAt(ctx, address, block)
}

type watcherSource struct {
	ec                 rocketpool.ExecutionClient
	storage, multicall common.Address
	chainID            uint64
}

func (s *watcherSource) contracts(ctx context.Context) (*rocketpool.RocketPool, error) {
	return rocketpool.NewRocketPool(watcherExecutionClient{s.ec, ctx}, s.storage)
}
func (s *watcherSource) Snapshot(ctx context.Context, finalized bool) (watcher.Parameters, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var p watcher.Parameters
	chain, err := s.ec.ChainID(ctx)
	if err != nil {
		return p, err
	}
	if !chain.IsUint64() || chain.Uint64() != s.chainID {
		return p, fmt.Errorf("execution client chain ID does not match watcher network")
	}
	sync, err := s.ec.SyncProgress(ctx)
	if err != nil {
		return p, err
	}
	if sync != nil {
		return p, watcher.ErrSyncing
	}
	var number *big.Int
	if finalized {
		number = big.NewInt(-3)
	}
	h, err := s.ec.HeaderByNumber(ctx, number)
	if err != nil {
		return p, err
	}
	p.Block = h.Number.Uint64()
	p.BlockHash = h.Hash().Hex()
	rp, err := s.contracts(ctx)
	if err != nil {
		return p, err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: h.Number}
	p.Deployed, err = state.IsSaturn2Deployed(rp, opts)
	if err != nil || !p.Deployed {
		return p, err
	}
	p.ExitsEnabled, err = protocol.GetPerformanceExitsEnabled(rp, opts)
	if err != nil {
		return p, err
	}
	p.Period, err = protocol.GetPerformancePeriod(rp, opts)
	if err != nil {
		return p, err
	}
	p.Buffer, err = protocol.GetProofBuffer(rp, opts)
	if err != nil {
		return p, err
	}
	p.Threshold, err = protocol.GetPerformanceThreshold(rp, opts)
	if err != nil {
		return p, err
	}
	_, err = p.RequiredMisses()
	return p, err
}

func (s *watcherSource) Registry(ctx context.Context, p watcher.Parameters) ([]watcher.Validator, error) {
	rp, err := s.contracts(ctx)
	if err != nil {
		return nil, err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(p.Block)}
	manager, err := rp.GetContract("rocketMinipoolManager", opts)
	if err != nil {
		return nil, err
	}
	statusABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"getStatus","stateMutability":"view","inputs":[],"outputs":[{"type":"uint8"}]}]`))
	if err != nil {
		return nil, err
	}
	count, err := minipool.GetMinipoolCount(rp, opts)
	if err != nil {
		return nil, err
	}
	var out []watcher.Validator
	for start := uint64(0); start < count; start += 100 {
		n := min(uint64(100), count-start)
		addresses := make([]common.Address, n)
		mc, err := multicall.NewMultiCaller(rp.Client, s.multicall)
		if err != nil {
			return nil, err
		}
		for i := range addresses {
			if err = mc.AddCall(manager, &addresses[i], "getMinipoolAt", new(big.Int).SetUint64(start+uint64(i))); err != nil {
				return nil, err
			}
		}
		if _, err = mc.FlexibleCall(true, opts); err != nil {
			return nil, err
		}
		pubkeys := make([]types.ValidatorPubkey, n)
		statuses := make([]uint8, n)
		for i := range addresses {
			contract := &rocketpool.Contract{Address: &addresses[i], ABI: &statusABI, Client: rp.Client}
			if err = mc.AddCall(manager, &pubkeys[i], "getMinipoolPubkey", addresses[i]); err != nil {
				return nil, err
			}
			if err = mc.AddCall(contract, &statuses[i], "getStatus"); err != nil {
				return nil, err
			}
		}
		if _, err = mc.FlexibleCall(true, opts); err != nil {
			return nil, err
		}
		for i, address := range addresses {
			if statuses[i] == uint8(types.Staking) {
				out = append(out, watcher.Validator{Pubkey: "0x" + hex.EncodeToString(pubkeys[i][:]), Pool: address.Hex(), Kind: "minipool", Eligible: true})
			}
		}
	}
	// The registry can exist before Saturn 2; its absence before deployment is
	// expected. An RPC failure, however, must not silently drop existing pools.
	address, err := rp.GetAddress("rocketMegapoolManager", opts)
	if err != nil {
		return nil, err
	}
	if *address == (common.Address{}) {
		return out, nil
	}
	mega, err := rp.GetContract("rocketMegapoolManager", opts)
	if err != nil {
		return nil, err
	}
	megaCount, err := megapool.GetValidatorCount(rp, opts)
	if err != nil {
		return nil, err
	}
	for start := uint32(0); start < megaCount; start += 100 {
		mc, err := multicall.NewMultiCaller(rp.Client, s.multicall)
		if err != nil {
			return nil, err
		}
		end := min(uint64(start)+100, uint64(megaCount))
		for i := uint64(start); i < end; i++ {
			if err = mc.AddCall(mega, nil, "getValidatorInfo", new(big.Int).SetUint64(i)); err != nil {
				return nil, err
			}
		}
		rows, err := mc.Execute(true, opts)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			v, err := statebindings.DecodeMegapoolValidatorInfo(mega, row.ReturnDataRaw)
			if err != nil {
				return nil, err
			}
			if v.ValidatorInfo.Staked && !v.ValidatorInfo.Exited && !v.ValidatorInfo.Dissolved && !v.ValidatorInfo.Exiting {
				out = append(out, watcher.Validator{Pubkey: "0x" + hex.EncodeToString(v.Pubkey), Pool: v.MegapoolAddress.Hex(), Kind: "megapool", ValidatorID: v.ValidatorId, Eligible: true})
			}
		}
	}
	return out, nil
}

// NodePools resolves only this node's pools, without the performance discovery
// cache or any requirement for Saturn 2 challenge contracts. Every read uses
// the same finalized execution block and the watcher's shared RPC budget.
func (s *watcherSource) NodePools(ctx context.Context, nodeAddress string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if !common.IsHexAddress(nodeAddress) || common.HexToAddress(nodeAddress) == (common.Address{}) {
		return nil, fmt.Errorf("invalid node address %q", nodeAddress)
	}
	address := common.HexToAddress(nodeAddress)
	h, err := s.ec.HeaderByNumber(ctx, big.NewInt(-3))
	if err != nil {
		return nil, err
	}
	rp, err := s.contracts(ctx)
	if err != nil {
		return nil, err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: h.Number}
	manager, err := rp.GetContract("rocketMinipoolManager", opts)
	if err != nil {
		return nil, err
	}
	count := new(big.Int)
	if err = manager.Call(opts, &count, "getNodeMinipoolCount", address); err != nil {
		return nil, err
	}
	if !count.IsUint64() {
		return nil, fmt.Errorf("invalid node minipool count")
	}
	pools := []string{}
	for start := uint64(0); start < count.Uint64(); {
		n := min(uint64(100), count.Uint64()-start)
		addresses := make([]common.Address, n)
		mc, err := multicall.NewMultiCaller(rp.Client, s.multicall)
		if err != nil {
			return nil, err
		}
		for i := range addresses {
			if err = mc.AddCall(manager, &addresses[i], "getNodeMinipoolAt", address, new(big.Int).SetUint64(start+uint64(i))); err != nil {
				return nil, err
			}
		}
		if _, err = mc.FlexibleCall(true, opts); err != nil {
			return nil, err
		}
		for _, pool := range addresses {
			pools = append(pools, pool.Hex())
		}
		start += n
	}
	// Older networks may have no megapool factory yet. RPC failures must still
	// surface instead of silently producing an incomplete ownership report.
	factory, err := rp.GetAddress("rocketMegapoolFactory", opts)
	if err != nil {
		return nil, err
	}
	if *factory != (common.Address{}) {
		pool, err := megapool.GetMegapoolExpectedAddress(rp, address, opts)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool.Hex())
	}
	return pools, nil
}

func (s *watcherSource) Eligible(ctx context.Context, p watcher.Parameters, v watcher.Validator) (bool, error) {
	rp, err := s.contracts(ctx)
	if err != nil {
		return false, err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: new(big.Int).SetUint64(p.Block)}
	if v.Kind == "minipool" {
		mp, err := minipool.NewMinipool(rp, common.HexToAddress(v.Pool), opts)
		if err != nil {
			return false, err
		}
		status, err := mp.GetStatus(opts)
		if err != nil {
			return false, err
		}
		pubkey, err := minipool.GetMinipoolPubkey(rp, common.HexToAddress(v.Pool), opts)
		return status == types.Staking && "0x"+hex.EncodeToString(pubkey[:]) == v.Pubkey, err
	}
	mp, err := megapool.NewMegapool(rp, common.HexToAddress(v.Pool), opts)
	if err != nil {
		return false, err
	}
	info, err := mp.GetValidatorInfo(v.ValidatorID, opts)
	if err != nil {
		return false, err
	}
	pubkey, err := mp.GetValidatorPubkey(v.ValidatorID, opts)
	return info.Staked && !info.Exited && !info.Exiting && !info.Dissolved && "0x"+hex.EncodeToString(pubkey[:]) == v.Pubkey, err
}
