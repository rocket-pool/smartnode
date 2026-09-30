package services

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/performance"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

type PerformanceChallengeRequest struct {
	MegapoolAddress   common.Address
	ValidatorIds      []uint32
	MinipoolAddresses []common.Address
	StartEpoch        uint64
	Participation     []*big.Int
}

// PreparedPerformanceChallenge contains the payload and slot proof used for submission.
type PreparedPerformanceChallenge struct {
	Request       PerformanceChallengeRequest
	SlotTimestamp uint64
	Slot          megapool.SlotProof
}

func (r PerformanceChallengeRequest) Validate() error {
	if len(r.MinipoolAddresses) > 0 {
		if len(r.ValidatorIds) > 0 || r.MegapoolAddress != (common.Address{}) {
			return fmt.Errorf("choose either minipools or megapool validators")
		}
		return validateChallengeMembers(r.MinipoolAddresses)
	}
	if len(r.ValidatorIds) == 0 || len(r.ValidatorIds) > megapool.MaxPerformanceChallengeValidators {
		return fmt.Errorf("a challenge requires 1 to 32 validators")
	}
	seen := map[uint32]bool{}
	for _, id := range r.ValidatorIds {
		if seen[id] {
			return fmt.Errorf("duplicate validator ID %d", id)
		}
		seen[id] = true
	}
	return nil
}

func validateChallengeMembers(members []common.Address) error {
	if len(members) == 0 || len(members) > megapool.MaxPerformanceChallengeValidators {
		return fmt.Errorf("a challenge requires 1 to 32 minipools")
	}
	seen := map[common.Address]bool{}
	for _, address := range members {
		if address == (common.Address{}) || seen[address] {
			return fmt.Errorf("zero or duplicate minipool address %s", address)
		}
		seen[address] = true
	}
	return nil
}

// AvailablePerformanceBond mirrors RocketNodeStaking.lockRPL without underflow.
func AvailablePerformanceBond(staked, locked *big.Int) *big.Int {
	available := new(big.Int).Sub(staked, locked)
	if available.Sign() < 0 {
		available.SetInt64(0)
	}
	return available
}

// PreparePerformanceChallenge validates the exact payload and estimates at current
// state for both preflight and submission. Contracts remain authoritative for
// membership, staking status, same-node minipools and outstanding exit requests.
func PreparePerformanceChallenge(c *cli.Command, request PerformanceChallengeRequest, opts *bind.TransactOpts) (*api.CanChallengeMegapoolPerformanceResponse, *PreparedPerformanceChallenge, error) {
	result := &api.CanChallengeMegapoolPerformanceResponse{}
	if err := request.Validate(); err != nil {
		return nil, nil, err
	}
	if err := RequireNodeRegistered(c); err != nil {
		return nil, nil, err
	}
	if err := RequireBeaconClientSynced(c); err != nil {
		return nil, nil, err
	}
	rp, err := GetRocketPool(c)
	if err != nil {
		return nil, nil, err
	}
	deployed, err := state.IsSaturn2Deployed(rp, nil)
	if err != nil {
		return nil, nil, err
	}
	if !deployed {
		return nil, nil, fmt.Errorf("performance challenges require Saturn 2")
	}
	result.ChallengeBond, err = protocol.GetPerformanceChallengeBond(rp, nil)
	if err != nil {
		return nil, nil, err
	}
	stake, err := node.GetNodeStakedRPL(rp, opts.From, nil)
	if err != nil {
		return nil, nil, err
	}
	locked, err := node.GetNodeLockedRPL(rp, opts.From, nil)
	if err != nil {
		return nil, nil, err
	}
	result.RplBalance = AvailablePerformanceBond(stake, locked)
	result.RplLockingAllowed, err = node.GetRPLLockedAllowed(rp, opts.From, nil)
	if err != nil {
		return nil, nil, err
	}
	if !result.RplLockingAllowed {
		result.Reason = "Enable RPL locking before submitting a performance challenge."
	}
	if result.RplBalance.Cmp(result.ChallengeBond) < 0 {
		result.InsufficientRplBalance = true
		result.Reason = "Insufficient unlocked staked RPL for the challenge bond."
	}
	if result.Reason != "" {
		return result, nil, nil
	}
	var pubkey types.ValidatorPubkey
	if len(request.MinipoolAddresses) > 0 {
		for _, address := range request.MinipoolAddresses {
			mp, err := minipool.NewMinipool(rp, address, nil)
			if err != nil {
				return nil, nil, err
			}
			status, err := mp.GetStatus(nil)
			if err != nil {
				return nil, nil, err
			}
			if status != types.Staking {
				result.Reason = fmt.Sprintf("Minipool %s is not staking.", address.Hex())
				return result, nil, nil
			}
		}
		pubkey, err = minipool.GetMinipoolPubkey(rp, request.MinipoolAddresses[0], nil)
	} else {
		if request.MegapoolAddress == (common.Address{}) {
			request.MegapoolAddress, err = node.GetMegapoolAddress(rp, opts.From, nil)
			if err != nil {
				return nil, nil, err
			}
		}
		var mp megapool.Megapool
		mp, err = megapool.NewMegapool(rp, request.MegapoolAddress, nil)
		if err == nil {
			for _, validatorId := range request.ValidatorIds {
				validator, err := mp.GetValidatorInfo(validatorId, nil)
				if err != nil {
					return nil, nil, err
				}
				if !validator.Staked {
					result.Reason = fmt.Sprintf("Validator %d is not staked.", validatorId)
					return result, nil, nil
				}
			}
			pubkey, err = mp.GetValidatorPubkey(request.ValidatorIds[0], nil)
		}
	}
	if err != nil {
		return nil, nil, err
	}
	bc, err := GetBeaconClient(c)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, nil, err
	}
	w, err := GetWallet(c)
	if err != nil {
		return nil, nil, err
	}
	_, timestamp, slot, err := GetValidatorProof(c, 0, w, cfg, pubkey, nil)
	if err != nil {
		return nil, nil, err
	}
	params, err := performance.GetChallengeParams(rp)
	if err != nil {
		return nil, nil, err
	}
	if err := performance.ValidateChallengeBitmap(params, slot.Slot/32, request.StartEpoch, request.Participation); err != nil {
		return nil, nil, err
	}
	opts.Value = nil
	if len(request.MinipoolAddresses) > 0 {
		result.GasLimits, err = megapool.EstimateChallengeMinipoolsGas(rp, request.MinipoolAddresses, request.StartEpoch, request.Participation, timestamp, slot, opts)
	} else {
		result.GasLimits, err = megapool.EstimateChallengeMegapoolGas(rp, request.MegapoolAddress, request.ValidatorIds, request.StartEpoch, request.Participation, timestamp, slot, opts)
	}
	if err != nil {
		return nil, nil, err
	}
	result.CanChallenge = true
	prepared := &PreparedPerformanceChallenge{
		Request:       request,
		SlotTimestamp: timestamp,
		Slot:          slot,
	}
	return result, prepared, nil
}

func SubmitPerformanceChallenge(rp *rocketpool.RocketPool, challenge *PreparedPerformanceChallenge, opts *bind.TransactOpts) (common.Hash, error) {
	opts.Value = nil
	request := challenge.Request
	if len(request.MinipoolAddresses) > 0 {
		return megapool.ChallengeMinipools(rp, request.MinipoolAddresses, request.StartEpoch, request.Participation, challenge.SlotTimestamp, challenge.Slot, opts)
	}
	return megapool.ChallengeMegapool(rp, request.MegapoolAddress, request.ValidatorIds, request.StartEpoch, request.Participation, challenge.SlotTimestamp, challenge.Slot, opts)
}

// BuildPerformanceChallengeDefense tries all members, using an activation
// proof when possible before fetching historical data. A failed member must
// not hide another valid defense.
func BuildPerformanceChallengeDefense(c *cli.Command, challenge megapool.PerformanceChallenge) (megapool.PerformanceChallengeDefense, error) {
	rp, err := GetRocketPool(c)
	if err != nil {
		return megapool.PerformanceChallengeDefense{}, err
	}
	bc, err := GetBeaconClient(c)
	if err != nil {
		return megapool.PerformanceChallengeDefense{}, err
	}
	cfg, err := bc.GetEth2Config()
	if err != nil {
		return megapool.PerformanceChallengeDefense{}, err
	}
	w, err := GetWallet(c)
	if err != nil {
		return megapool.PerformanceChallengeDefense{}, err
	}
	epochs := []uint64{}
	for wordIndex, word := range challenge.Participation {
		for bit := 0; bit < word.BitLen(); bit++ {
			if word.Bit(bit) != 0 {
				epochs = append(epochs, challenge.StartEpoch+uint64(wordIndex)*256+uint64(bit))
			}
		}
	}
	count := len(challenge.ValidatorIds)
	if len(challenge.MinipoolAddresses) > 0 {
		count = len(challenge.MinipoolAddresses)
	}
	var failures []error
	for i := 0; i < count; i++ {
		defense := megapool.PerformanceChallengeDefense{ChallengeId: challenge.ChallengeId}
		var pubkey types.ValidatorPubkey
		if len(challenge.MinipoolAddresses) > 0 {
			defense.MinipoolAddress = challenge.MinipoolAddresses[i]
			pubkey, err = minipool.GetMinipoolPubkey(rp, defense.MinipoolAddress, nil)
		} else {
			defense.ValidatorId = challenge.ValidatorIds[i]
			var mp megapool.Megapool
			mp, err = megapool.NewMegapool(rp, challenge.MegapoolAddress, nil)
			if err == nil {
				pubkey, err = mp.GetValidatorPubkey(defense.ValidatorId, nil)
			}
		}
		if err != nil {
			failures = append(failures, err)
			continue
		}
		status, statusErr := bc.GetValidatorStatus(pubkey, nil)
		if statusErr != nil {
			failures = append(failures, statusErr)
			continue
		}
		if !status.Exists {
			continue
		}
		if status.ActivationEpoch > challenge.StartEpoch {
			defense.Validator, defense.SlotTimestamp, defense.Slot, err = GetValidatorProof(c, 0, w, cfg, pubkey, nil)
			if err == nil {
				return defense, nil
			}
			failures = append(failures, err)
			continue
		}
		index, parseErr := strconv.ParseUint(status.Index, 10, 64)
		if parseErr != nil {
			failures = append(failures, parseErr)
			continue
		}
		// Missing historical data for one epoch must not hide a later defense.
		proofs, searchErr := findParticipationDefense(epochs, func(epoch uint64) (bool, error) {
			_, _, found, err := performance.FindFirstTimelyTargetVote(bc, cfg, []uint64{index}, []uint64{epoch})
			return found, err
		}, func(epoch uint64) (PerformanceDefenseProofs, error) {
			return GetParticipationProof(c, index, pubkey, epoch, challenge.StartEpoch, challenge.Participation)
		})
		if searchErr != nil {
			failures = append(failures, searchErr)
			continue
		}
		if proofs != nil {
			defense.SlotTimestamp = proofs.SlotTimestamp
			defense.Validator = proofs.Validator
			defense.Slot = proofs.Slot
			defense.Participation = &proofs.Participation
			defense.Offset = proofs.Offset
			defense.ChallengeLeaf = proofs.ChallengeLeaf
			defense.ChallengeWitness = proofs.ChallengeWitness
			return defense, nil
		}
	}
	if len(failures) > 0 {
		return megapool.PerformanceChallengeDefense{}, fmt.Errorf("could not complete defense search: %w", errors.Join(failures...))
	}
	return megapool.PerformanceChallengeDefense{}, fmt.Errorf("no activation or timely-target defense found for challenge %s", challenge.ChallengeId)
}

// findParticipationDefense keeps searching when a beacon endpoint lacks one
// epoch or a proof cannot be built. A positive vote alone is not a usable proof.
func findParticipationDefense(epochs []uint64, find func(uint64) (bool, error), build func(uint64) (PerformanceDefenseProofs, error)) (*PerformanceDefenseProofs, error) {
	var failures []error
	for _, epoch := range epochs {
		found, err := find(epoch)
		if err != nil {
			failures = append(failures, fmt.Errorf("epoch %d: %w", epoch, err))
			continue
		}
		if !found {
			continue
		}
		proof, err := build(epoch)
		if err != nil {
			failures = append(failures, fmt.Errorf("epoch %d: %w", epoch, err))
			continue
		}
		return &proof, nil
	}
	return nil, errors.Join(failures...)
}
