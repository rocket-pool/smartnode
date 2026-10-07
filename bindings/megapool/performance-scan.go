package megapool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Bound initial scans by the first challenge, whose timestamp survives both
// settlement and contract upgrades. Header lookup does not need archive state.
// A nil returned start means there are no challenges in this snapshot.
func performanceChallengeScanRange(rp *rocketpool.RocketPool, from, to *big.Int, opts *bind.CallOpts) (*big.Int, *big.Int, error) {
	if from != nil {
		return from, to, nil
	}
	callOpts := bind.CallOpts{}
	if opts != nil {
		callOpts = *opts
	}
	ctx := callOpts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if to == nil {
		to = callOpts.BlockNumber
		if to == nil {
			block, err := rp.Client.BlockNumber(ctx)
			if err != nil {
				return nil, nil, err
			}
			to = new(big.Int).SetUint64(block)
		}
	}
	if !to.IsUint64() {
		return nil, nil, fmt.Errorf("invalid performance challenge scan end block: %s", to)
	}
	callOpts.BlockNumber, callOpts.BlockHash, callOpts.Pending = to, common.Hash{}, false
	key := crypto.Keccak256Hash([]byte("participation.challenge.time"), common.LeftPadBytes(big.NewInt(1).Bytes(), 32))
	firstTime, err := rp.RocketStorage.GetUint(&callOpts, key)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting first performance challenge timestamp: %w", err)
	}
	if firstTime.Sign() == 0 {
		return nil, to, nil
	}
	if !firstTime.IsUint64() {
		return nil, nil, fmt.Errorf("invalid first performance challenge timestamp: %s", firstTime)
	}
	lo, hi := uint64(0), to.Uint64()
	head, err := rp.Client.HeaderByNumber(ctx, to)
	if err != nil {
		return nil, nil, err
	}
	if head == nil || head.Time < firstTime.Uint64() {
		return nil, nil, fmt.Errorf("first performance challenge timestamp is after scan end block")
	}
	for lo < hi {
		mid := lo + (hi-lo)/2
		header, err := rp.Client.HeaderByNumber(ctx, new(big.Int).SetUint64(mid))
		if err != nil {
			return nil, nil, err
		}
		if header == nil {
			return nil, nil, fmt.Errorf("missing execution header at block %d", mid)
		}
		if header.Time < firstTime.Uint64() {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return new(big.Int).SetUint64(lo), to, nil
}
