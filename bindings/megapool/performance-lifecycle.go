package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/logs"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

// MaxPerformanceChallengeValidators is the contract's per-list limit.
const MaxPerformanceChallengeValidators = 32

// EstimateChallengeMinipoolsGas estimates the gas to challenge a minipool list.
func EstimateChallengeMinipoolsGas(rp *rocketpool.RocketPool, minipoolAddresses []common.Address, startEpoch uint64, participation []*big.Int, slotTimestamp uint64, slotProof SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return rocketNetworkParticipation.GetTransactionGasInfo(opts, "challengeMinipools", minipoolAddresses, startEpoch, participation, slotTimestamp, slotProof)
}

// ChallengeMinipools challenges the target-vote performance of a minipool list.
func ChallengeMinipools(rp *rocketpool.RocketPool, minipoolAddresses []common.Address, startEpoch uint64, participation []*big.Int, slotTimestamp uint64, slotProof SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "challengeMinipools", minipoolAddresses, startEpoch, participation, slotTimestamp, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error challenging minipools: %w", err)
	}
	return tx.Hash(), nil
}

// PerformanceChallengeDefense represents either permitted proof for either pool
// type. A nonzero MinipoolAddress selects the minipool entry point; otherwise
// ValidatorId identifies a member of the challenged megapool.
type PerformanceChallengeDefense struct {
	ChallengeId      *big.Int
	ValidatorId      uint32
	MinipoolAddress  common.Address
	SlotTimestamp    uint64
	Validator        ValidatorProof
	Slot             SlotProof
	Participation    *ParticipationProof
	Offset           uint64
	ChallengeLeaf    *big.Int
	ChallengeWitness []common.Hash
}

func (d PerformanceChallengeDefense) call() (string, []interface{}) {
	validatorMethod := "respondWithMegapoolValidator"
	participationMethod := "respondWithMegapoolParticipation"
	var member interface{} = d.ValidatorId
	if d.MinipoolAddress != (common.Address{}) {
		validatorMethod = "respondWithMinipoolValidator"
		participationMethod = "respondWithMinipoolParticipation"
		member = d.MinipoolAddress
	}
	if d.Participation != nil {
		return participationMethod, []interface{}{
			d.ChallengeId, member, d.Offset, d.ChallengeLeaf, d.ChallengeWitness,
			d.SlotTimestamp, d.Validator, *d.Participation, d.Slot,
		}
	}
	return validatorMethod, []interface{}{d.ChallengeId, member, d.SlotTimestamp, d.Validator, d.Slot}
}

func (d PerformanceChallengeDefense) EstimateGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	method, args := d.call()
	return rocketNetworkParticipation.GetTransactionGasInfo(opts, method, args...)
}

func (d PerformanceChallengeDefense) Submit(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	method, args := d.call()
	tx, err := rocketNetworkParticipation.Transact(opts, method, args...)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

func (s PerformanceChallengeStatus) CanFinalise(timestamp uint64) bool {
	return !s.Responded && !s.Finalised && s.ResponseDeadline.Cmp(new(big.Int).SetUint64(timestamp)) < 0
}

func (s PerformanceChallengeStatus) CanDefend(caller common.Address, timestamp uint64) bool {
	return !s.Responded && !s.Finalised && s.Proposer != caller && s.ResponseDeadline.Cmp(new(big.Int).SetUint64(timestamp)) >= 0
}

// GetPerformanceChallenges discovers both challenge types across upgrades.
// challengeId optionally limits the event query to one challenge.
func GetPerformanceChallenges(rp *rocketpool.RocketPool, challengeId, interval, fromBlock, toBlock *big.Int, opts *bind.CallOpts) ([]PerformanceChallenge, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, opts)
	if err != nil {
		return nil, err
	}
	megapoolEvent, ok := rocketNetworkParticipation.ABI.Events["MegapoolChallenged"]
	if !ok {
		return nil, fmt.Errorf("MegapoolChallenged event missing")
	}
	minipoolEvent, ok := rocketNetworkParticipation.ABI.Events["MinipoolsChallenged"]
	if !ok {
		return nil, fmt.Errorf("MinipoolsChallenged event missing")
	}
	addresses, err := participationContractAddresses(rp, rocketNetworkParticipation, interval, fromBlock, toBlock, opts)
	if err != nil {
		return nil, err
	}
	topics := [][]common.Hash{{megapoolEvent.ID, minipoolEvent.ID}}
	if challengeId != nil {
		topics = append(topics, nil, []common.Hash{common.BigToHash(challengeId)})
	}
	entries, err := logs.GetLogs(rp, addresses, topics, interval, fromBlock, toBlock, nil)
	if err != nil {
		return nil, err
	}
	challenges := make([]PerformanceChallenge, 0, len(entries))
	for _, entry := range entries {
		if entry.Removed {
			continue
		}
		if len(entry.Topics) != 3 {
			return nil, fmt.Errorf("invalid performance challenge event topics")
		}
		var challenge PerformanceChallenge
		switch entry.Topics[0] {
		case megapoolEvent.ID:
			challenge, err = decodePerformanceChallenge(megapoolEvent, entry)
		case minipoolEvent.ID:
			challenge, err = decodeMinipoolPerformanceChallenge(minipoolEvent, entry)
		default:
			return nil, fmt.Errorf("unknown performance challenge event")
		}
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, nil
}

func decodeMinipoolPerformanceChallenge(event abi.Event, entry types.Log) (PerformanceChallenge, error) {
	if len(entry.Topics) != 3 || entry.Topics[0] != event.ID {
		return PerformanceChallenge{}, fmt.Errorf("invalid MinipoolsChallenged event topics")
	}
	values, err := event.Inputs.NonIndexed().Unpack(entry.Data)
	if err != nil {
		return PerformanceChallenge{}, fmt.Errorf("error unpacking MinipoolsChallenged: %w", err)
	}
	if len(values) != 4 {
		return PerformanceChallenge{}, fmt.Errorf("invalid MinipoolsChallenged event data")
	}
	minipoolAddresses, addressesOK := values[0].([]common.Address)
	startEpoch, startOK := values[1].(uint64)
	participation, participationOK := values[3].([]*big.Int)
	if !addressesOK || !startOK || !participationOK || len(minipoolAddresses) == 0 || len(participation) == 0 {
		return PerformanceChallenge{}, fmt.Errorf("invalid MinipoolsChallenged minipool list or participation bitmap")
	}
	return PerformanceChallenge{
		ChallengeId:       entry.Topics[2].Big(),
		NodeAddress:       common.BytesToAddress(entry.Topics[1].Bytes()),
		MinipoolAddresses: minipoolAddresses,
		StartEpoch:        startEpoch,
		Participation:     participation,
	}, nil
}
