package beacon

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type stubHeaderSource struct {
	headers map[common.Hash]*types.Header
	err     error
	calls   int
}

func (s *stubHeaderSource) HeaderByHash(_ context.Context, hash common.Hash) (*types.Header, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if header, ok := s.headers[hash]; ok {
		return header, nil
	}
	return nil, ethereum.NotFound
}

func TestResolveExecutionBlockNumberUsesEmbeddedNumber(t *testing.T) {
	ec := &stubHeaderSource{}
	block := BeaconBlock{
		Slot:                 100,
		HasExecutionPayload:  true,
		ExecutionBlockNumber: 4242,
	}

	number, found, err := ResolveExecutionBlockNumber(context.Background(), ec, block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected found for a block with an embedded number")
	}
	if number != 4242 {
		t.Errorf("expected block 4242, got %d", number)
	}
	// Pre-Gloas blocks must not cost an extra EL round trip, and snapshot mode depends on it
	if ec.calls != 0 {
		t.Errorf("expected no EL calls for an embedded number, got %d", ec.calls)
	}
}

func TestResolveExecutionBlockNumberNoPayload(t *testing.T) {
	ec := &stubHeaderSource{}
	block := BeaconBlock{Slot: 100}

	number, found, err := ResolveExecutionBlockNumber(context.Background(), ec, block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected not found for a block with no payload association")
	}
	if number != 0 {
		t.Errorf("expected block 0, got %d", number)
	}
	if ec.calls != 0 {
		t.Errorf("expected no EL calls with an empty hash, got %d", ec.calls)
	}
}

func TestResolveExecutionBlockNumberFromHash(t *testing.T) {
	hash := common.HexToHash("0xabc123")
	ec := &stubHeaderSource{
		headers: map[common.Hash]*types.Header{
			hash: {Number: big.NewInt(9001)},
		},
	}
	block := BeaconBlock{
		Slot:                100,
		HasExecutionPayload: true,
		ExecutionBlockHash:  hash,
	}

	number, found, err := ResolveExecutionBlockNumber(context.Background(), ec, block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected found for a resolvable payload hash")
	}
	if number != 9001 {
		t.Errorf("expected block 9001, got %d", number)
	}
}

func TestResolveExecutionBlockNumberPayloadWithheld(t *testing.T) {
	ec := &stubHeaderSource{}
	block := BeaconBlock{
		Slot:                100,
		HasExecutionPayload: true,
		ExecutionBlockHash:  common.HexToHash("0xdeadbeef"),
	}

	number, found, err := ResolveExecutionBlockNumber(context.Background(), ec, block)
	if err != nil {
		t.Fatalf("expected no error for a withheld payload, got %v", err)
	}
	if found {
		t.Error("expected not found for a withheld payload")
	}
	if number != 0 {
		t.Errorf("expected block 0, got %d", number)
	}
}

func TestResolveExecutionBlockNumberPropagatesErrors(t *testing.T) {
	sentinel := errors.New("execution client is offline")
	ec := &stubHeaderSource{err: sentinel}
	block := BeaconBlock{
		Slot:                100,
		HasExecutionPayload: true,
		ExecutionBlockHash:  common.HexToHash("0xabc123"),
	}

	_, found, err := ResolveExecutionBlockNumber(context.Background(), ec, block)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the EL error to be wrapped, got %v", err)
	}
	if found {
		t.Error("expected not found on error")
	}
}
