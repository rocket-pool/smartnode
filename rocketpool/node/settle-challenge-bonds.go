package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/transactions"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
)

type challengeBondSource interface {
	header(uint64) (*types.Header, error)
	bondEvents(common.Address, *big.Int, *big.Int) ([]*big.Int, error)
	status(*big.Int, *big.Int) (megapool.PerformanceChallengeStatus, error)
}

func (s onchainPerformanceChallenges) bondEvents(node common.Address, from, to *big.Int) ([]*big.Int, error) {
	return megapool.GetChallengeBondCandidates(s.rp, node, s.interval, from, to, &bind.CallOpts{BlockNumber: to})
}

type challengeBond struct {
	id     *big.Int
	status megapool.PerformanceChallengeStatus
}

type challengeBondDiscovery struct {
	checkpoint *types.Header
	node       common.Address
	pending    []*big.Int
}

// Bonds outlive challenge finalisation/defeat. Keep tracking them until settled
// and rebuild history on restart or reorg, without a deadline-based cutoff.
func (d *challengeBondDiscovery) discover(source challengeBondSource, node common.Address, block uint64) ([]challengeBond, uint64, error) {
	head, err := source.header(block)
	if err != nil {
		return nil, 0, err
	}
	var from *big.Int
	var candidates []*big.Int
	if d.checkpoint != nil && d.node == node && d.checkpoint.Number.Cmp(head.Number) <= 0 {
		previous, err := source.header(d.checkpoint.Number.Uint64())
		if err != nil {
			return nil, 0, err
		}
		if previous.Hash() == d.checkpoint.Hash() {
			from = new(big.Int).Add(d.checkpoint.Number, big.NewInt(1))
			candidates = append(candidates, d.pending...)
		}
	}
	if from == nil || from.Cmp(head.Number) <= 0 {
		ids, err := source.bondEvents(node, from, head.Number)
		if err != nil {
			return nil, 0, err
		}
		candidates = append(candidates, ids...)
	}
	seen := map[string]bool{}
	pending := []*big.Int{}
	bonds := []challengeBond{}
	for _, id := range candidates {
		if seen[id.String()] {
			continue
		}
		seen[id.String()] = true
		status, err := source.status(id, head.Number)
		if err != nil {
			return nil, 0, err
		}
		if !keepChallengeBond(status, node) {
			continue
		}
		pending = append(pending, id)
		bonds = append(bonds, challengeBond{id: id, status: status})
	}
	confirmed, err := source.header(block)
	if err != nil {
		return nil, 0, err
	}
	if confirmed.Hash() != head.Hash() {
		return nil, 0, fmt.Errorf("execution chain changed during challenge bond discovery")
	}
	d.checkpoint, d.node, d.pending = head, node, pending
	return bonds, head.Time, nil
}

func keepChallengeBond(status megapool.PerformanceChallengeStatus, node common.Address) bool {
	if status.BondSettled {
		return false
	}
	if status.Responded {
		return status.Responder == node
	}
	return status.Proposer == node
}

func (t *defendChallengePerformance) settleChallengeBonds(node common.Address, block uint64) error {
	bonds, timestamp, err := t.bondDiscovery.discover(onchainPerformanceChallenges{rp: t.rp, interval: t.intervalSize}, node, block)
	if err != nil {
		return err
	}
	for _, bond := range bonds {
		claim := bond.status.CanClaimReward(node)
		if !claim && !bond.status.CanReleaseBond(timestamp) {
			continue
		}
		if err := t.settleChallengeBond(bond.id, claim); err != nil {
			t.log.Printlnf("error settling challenge %s bond: %v", bond.id, err)
		}
	}
	return nil
}

func (t *defendChallengePerformance) settleChallengeBond(id *big.Int, claim bool) error {
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}
	estimate := megapool.EstimateReleaseChallengeBondGas
	submit := megapool.ReleaseChallengeBond
	action := "release bond"
	if claim {
		estimate, submit, action = megapool.EstimateClaimChallengeRewardGas, megapool.ClaimChallengeReward, "claim defender reward as staked RPL"
	}
	gasInfo, err := estimate(t.rp, id, opts)
	if err != nil {
		return err
	}
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Sign() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return err
		}
	}
	if !gasInfo.PrintAndCheck(true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return nil
	}
	opts.GasFeeCap, opts.GasTipCap, opts.GasLimit = maxFee, GetPriorityFee(t.maxPriorityFee, maxFee), gasInfo.Safe
	t.log.Printlnf("Challenge %s: %s.", id, action)
	hash, err := submit(t.rp, id, opts)
	if err != nil {
		return err
	}
	if err := transactions.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log); err != nil {
		return err
	}
	t.log.Printlnf("Challenge %s: successfully completed %s.", id, action)
	return nil
}
