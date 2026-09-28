package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

type performanceChallengeSource interface {
	header(uint64) (*types.Header, error)
	events(common.Address, *big.Int, *big.Int) ([]megapool.PerformanceChallenge, error)
	status(*big.Int, *big.Int) (megapool.PerformanceChallengeStatus, error)
}

type onchainPerformanceChallenges struct {
	rp       *rocketpool.RocketPool
	interval *big.Int
}

func (s onchainPerformanceChallenges) header(block uint64) (*types.Header, error) {
	return s.rp.Client.HeaderByNumber(context.Background(), new(big.Int).SetUint64(block))
}

func (s onchainPerformanceChallenges) events(address common.Address, from, to *big.Int) ([]megapool.PerformanceChallenge, error) {
	return megapool.GetMegapoolPerformanceChallenges(s.rp, address, s.interval, from, to, &bind.CallOpts{BlockNumber: to})
}

func (s onchainPerformanceChallenges) status(id, block *big.Int) (megapool.PerformanceChallengeStatus, error) {
	return megapool.GetPerformanceChallengeStatus(s.rp, id, &bind.CallOpts{BlockNumber: block})
}

// Retain unresolved challenges without a time cutoff. Restarting or detecting a
// reorg rebuilds the cache from historical events; ordinary runs scan new blocks.
type performanceChallengeDiscovery struct {
	checkpoint *types.Header
	address    common.Address
	unresolved []megapool.PerformanceChallenge
}

func (d *performanceChallengeDiscovery) discover(source performanceChallengeSource, address common.Address, block uint64) ([]megapoolPerformanceChallenge, uint64, error) {
	head, err := source.header(block)
	if err != nil {
		return nil, 0, err
	}
	var from *big.Int
	var candidates []megapool.PerformanceChallenge
	if d.checkpoint != nil && d.address == address && d.checkpoint.Number.Cmp(head.Number) <= 0 {
		previous, err := source.header(d.checkpoint.Number.Uint64())
		if err != nil {
			return nil, 0, err
		}
		if previous.Hash() == d.checkpoint.Hash() {
			from = new(big.Int).Add(d.checkpoint.Number, big.NewInt(1))
			candidates = append(candidates, d.unresolved...)
		}
	}
	if from == nil || from.Cmp(head.Number) <= 0 {
		events, err := source.events(address, from, head.Number)
		if err != nil {
			return nil, 0, fmt.Errorf("error discovering performance challenges: %w", err)
		}
		candidates = append(candidates, events...)
	}
	active := make([]megapoolPerformanceChallenge, 0, len(candidates))
	unresolved := make([]megapool.PerformanceChallenge, 0, len(candidates))
	seen := make(map[string]bool)
	for _, event := range candidates {
		if event.MegapoolAddress != address || seen[event.ChallengeId.String()] {
			continue
		}
		seen[event.ChallengeId.String()] = true
		status, err := source.status(event.ChallengeId, head.Number)
		if err != nil {
			return nil, 0, err
		}
		if status.Responded || status.Finalised {
			continue
		}
		unresolved = append(unresolved, event)
		active = append(active, megapoolPerformanceChallenge{
			challengeId: event.ChallengeId, megapoolAddress: event.MegapoolAddress,
			validatorIds: event.ValidatorIds, startEpoch: event.StartEpoch,
			participationCallData: event.Participation, responseDeadline: status.ResponseDeadline,
			proposer: status.Proposer,
		})
	}
	// Do not advance the cursor on a partial scan or a reorg during the reads.
	confirmed, err := source.header(block)
	if err != nil {
		return nil, 0, err
	}
	if confirmed.Hash() != head.Hash() {
		return nil, 0, fmt.Errorf("execution chain changed during performance challenge discovery")
	}
	d.checkpoint, d.address, d.unresolved = head, address, unresolved
	return active, head.Time, nil
}

func (c megapoolPerformanceChallenge) expired(timestamp uint64) bool {
	// finaliseChallenge requires deadline < block.timestamp; a response is
	// still valid at the exact deadline.
	return c.responseDeadline.Cmp(new(big.Int).SetUint64(timestamp)) < 0
}
