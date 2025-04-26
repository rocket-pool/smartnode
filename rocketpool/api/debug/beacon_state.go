package debug

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

const MAX_WITHDRAWAL_SLOT_DISTANCE = 432000 // 60 days.

func getBeaconStateForSlot(c *cli.Context, slot uint64, validatorIndex uint64) error {
	// Create a new response
	response := api.BeaconStateResponse{}

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return err
	}

	// Get beacon state
	beaconState, err := bc.GetBeaconState(slot)
	if err != nil {
		return err
	}

	proof, err := beaconState.ValidatorCredentialsProof(validatorIndex)
	if err != nil {
		return err
	}

	// Convert the proof to a list of 0x-prefixed hex strings
	response.Proof = make([]string, 0, len(proof))
	for _, hash := range proof {
		response.Proof = append(response.Proof, hexutil.EncodeToString(hash))
	}

	// Render response json
	json, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Println(string(json))

	return nil
}

func getWithdrawalProofForSlot(c *cli.Context, slot uint64, validatorIndex uint64) error {
	// Create a new response
	response := api.WithdrawalProofResponse{}
	response.ValidatorIndex = validatorIndex
	response.Slot = slot
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return err
	}

	// Find the most recent withdrawal to slot.
	// Keep track of 404s- if we get 24 missing slots in a row, assume we don't have full history.
	notFounds := 0
	var beaconBlockDeneb *eth2.SignedBeaconBlockDeneb
	for candidateSlot := slot; candidateSlot >= slot-MAX_WITHDRAWAL_SLOT_DISTANCE; candidateSlot-- {
		// Get the block at the candidate slot.
		block, found, err := bc.GetBeaconBlockDeneb(candidateSlot)
		if err != nil {
			return err
		}
		if !found {
			notFounds++
			if notFounds >= 64 {
				return fmt.Errorf("2 epochs of missing slots detected. It is likely that the Beacon Client was checkpoint synced after the most recent withdrawal to slot %d, and does not have the history required to generate a withdrawal proof", slot)
			}
			continue
		} else {
			notFounds = 0
		}

		if block.Block.Body.ExecutionPayload == nil {
			continue
		}

		foundWithdrawal := false

		// Check the block for a withdrawal for the given validator index.
		for i, withdrawal := range block.Block.Body.ExecutionPayload.Withdrawals {
			if withdrawal.ValidatorIndex != validatorIndex {
				continue
			}
			response.WithdrawalSlot = candidateSlot
			response.Amount = big.NewInt(0).SetUint64(withdrawal.Amount)
			foundWithdrawal = true
			response.IndexInWithdrawalsArray = uint(i)
			response.WithdrawalIndex = withdrawal.Index
			response.WithdrawalAddress = withdrawal.Address
			break
		}

		if foundWithdrawal {
			beaconBlockDeneb = block
			break
		}
	}

	if response.Slot == 0 {
		return fmt.Errorf("no withdrawal found for validator index %d within %d slots of slot %d", validatorIndex, MAX_WITHDRAWAL_SLOT_DISTANCE, slot)
	}

	// Start by proving from the withdrawal to the block_root
	proof, err := beaconBlockDeneb.Block.ProveWithdrawal(uint64(response.IndexInWithdrawalsArray))
	if err != nil {
		return err
	}

	// Get beacon state
	state, err := bc.GetBeaconState(slot)
	if err != nil {
		return err
	}

	var summaryProof [][]byte

	var stateProof [][]byte
	if response.WithdrawalSlot+eth2.SlotsPerHistoricalRoot > state.Slot {
		stateProof, err = state.BlockRootProof(response.WithdrawalSlot)
		if err != nil {
			return err
		}
	} else {
		stateProof, err = state.HistoricalSummaryProof(response.WithdrawalSlot)
		if err != nil {
			return err
		}

		// Additionally, we need to prove from the block_root in the historical summary
		// up to the beginning of the above proof, which is the entry in the historical summaries vector.
		blockRootsStateSlot := eth2.SlotsPerHistoricalRoot + ((response.WithdrawalSlot / eth2.SlotsPerHistoricalRoot) * eth2.SlotsPerHistoricalRoot)
		// get the state that has the block roots tree
		blockRootsState, err := bc.GetBeaconState(blockRootsStateSlot)
		if err != nil {
			return err
		}
		summaryProof, err = blockRootsState.HistoricalSummaryBlockRootProof(int(response.WithdrawalSlot))
		if err != nil {
			return err
		}

	}

	// Convert the proof to a list of 0x-prefixed hex strings
	response.Proof = make([]string, 0, len(proof)+len(stateProof)+len(summaryProof))
	// First we prove from the withdrawal to the block_root
	for _, hash := range proof {
		response.Proof = append(response.Proof, hexutil.EncodeToString(hash))
	}

	// Then, if summaryProof has rows, we add them to prove from the block_root to the historical_summary row
	for _, hash := range summaryProof {
		response.Proof = append(response.Proof, hexutil.EncodeToString(hash))
	}

	// Finally, we prove either from the historical_summary or the block_root to the state_root
	for _, hash := range stateProof {
		response.Proof = append(response.Proof, hexutil.EncodeToString(hash))
	}

	// Render response json
	json, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Println(string(json))

	return nil
}
