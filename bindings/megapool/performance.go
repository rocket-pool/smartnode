package megapool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
)

// Estimate the gas to call ChallengeMegapool
func EstimateChallengeMegapoolGas(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, startEpoch uint64, participation []*big.Int, slotTimestamp uint64, slotProof SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return rocketNetworkParticipation.GetTransactionGasInfo(opts, "challengeMegapool", megapoolAddress, []uint32{validatorId}, startEpoch, participation, slotTimestamp, slotProof)
}

// Challenge the megapool
func ChallengeMegapool(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, startEpoch uint64, participation []*big.Int, slotTimestamp uint64, slotProof SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "challengeMegapool", megapoolAddress, []uint32{validatorId}, startEpoch, participation, slotTimestamp, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error challenging megapool: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas to call RespondWithParticipation
func EstimateRespondWithParticipationGas(rp *rocketpool.RocketPool, challengeId *big.Int, validatorId uint32, offset uint64, challengeLeaf *big.Int, challengeWitness []common.Hash, slotTimestamp uint64, validatorProof ValidatorProof, participationProof ParticipationProof, slotProof SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return rocketNetworkParticipation.GetTransactionGasInfo(opts, "respondWithMegapoolParticipation", challengeId, validatorId, offset, challengeLeaf, challengeWitness, slotTimestamp, validatorProof, participationProof, slotProof)
}

// Respond to a performance challenge with a proof that the validator did
// participate in one of the challenged epochs
func RespondWithParticipation(rp *rocketpool.RocketPool, challengeId *big.Int, validatorId uint32, offset uint64, challengeLeaf *big.Int, challengeWitness []common.Hash, slotTimestamp uint64, validatorProof ValidatorProof, participationProof ParticipationProof, slotProof SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "respondWithMegapoolParticipation", challengeId, validatorId, offset, challengeLeaf, challengeWitness, slotTimestamp, validatorProof, participationProof, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error responding to challenge: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas to call RespondWithValidator
func EstimateRespondWithValidatorGas(rp *rocketpool.RocketPool, challengeId *big.Int, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return rocketNetworkParticipation.GetTransactionGasInfo(opts, "respondWithMegapoolValidator", challengeId, validatorId, slotTimestamp, validatorProof, slotProof)
}

// Respond to a performance challenge with proof that a listed validator
// activated after the challenge's start epoch.
func RespondWithValidator(rp *rocketpool.RocketPool, challengeId *big.Int, validatorId uint32, slotTimestamp uint64, validatorProof ValidatorProof, slotProof SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "respondWithMegapoolValidator", challengeId, validatorId, slotTimestamp, validatorProof, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error responding to challenge: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas to call FinaliseChallenge
func EstimateFinaliseChallengeGas(rp *rocketpool.RocketPool, challengeId *big.Int, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return gaslimit.Limits{}, err
	}
	return rocketNetworkParticipation.GetTransactionGasInfo(opts, "finaliseChallenge", challengeId)
}

func FinaliseChallenge(rp *rocketpool.RocketPool, challengeId *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "finaliseChallenge", challengeId)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error finalising challenge: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketNetworkParticipationLock sync.Mutex

func getRocketNetworkParticipation(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkParticipationLock.Lock()
	defer rocketNetworkParticipationLock.Unlock()
	return rp.GetContract("rocketNetworkParticipation", opts)
}
