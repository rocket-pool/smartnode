package megapool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Challenge the megapool
func ChallengeMegapool(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorIds []uint32, startEpoch uint64, participation []*big.Int, slotTimestamp uint64, slotProof SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "challengeMegapool", megapoolAddress, validatorIds, startEpoch, participation, slotTimestamp, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error challenging megapool: %w", err)
	}
	return tx.Hash(), nil
}

func Respond(rp *rocketpool.RocketPool, challengeId uint64, offset uint64, challengeLeaf *big.Int, challengeWitness []common.Hash, slotTimestamp uint64, participationProof ParticipationProof, slotProof SlotProof, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkParticipation, err := getRocketNetworkParticipation(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkParticipation.Transact(opts, "respond", challengeId, offset, challengeLeaf, challengeWitness, slotTimestamp, participationProof, slotProof)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error responding to challenge: %w", err)
	}
	return tx.Hash(), nil
}

func FinaliseChallenge(rp *rocketpool.RocketPool, challengeId uint64, opts *bind.TransactOpts) (common.Hash, error) {
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
