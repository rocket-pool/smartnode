package services

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/bindings/storage"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	"golang.org/x/sync/errgroup"
)

const MAX_WITHDRAWAL_SLOT_DISTANCE = 144000 // 20 days.

func GetValidatorProof(c *cli.Context, slot uint64, wallet wallet.Wallet, eth2Config beacon.Eth2Config, megapoolAddress common.Address, validatorPubkey types.ValidatorPubkey, beaconState eth2.BeaconState) (megapool.ValidatorProof, uint64, megapool.SlotProof, error) {

	bc, err := GetBeaconClient(c)
	if err != nil {
		return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, err
	}

	// Get the validator index on the beacon chain
	validatorIndex, err := bc.GetValidatorIndex(validatorPubkey)
	if err != nil {
		return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, err
	}

	validatorIndex64, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, err
	}
	if beaconState == nil {
		beaconState, err = GetBeaconState(bc)
		if err != nil {
			return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, err
		}
	}

	slotProofBytes, err := beaconState.SlotProof(beaconState.GetSlot())
	if err != nil {
		return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, err
	}

	proofBytes, err := beaconState.ValidatorProof(validatorIndex64)
	if err != nil {
		return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, err
	}

	validators := beaconState.GetValidators()

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := ConvertToFixedSize(proofBytes)

	withdrawalCredentials := validators[validatorIndex64].WithdrawalCredentials
	// Convert the WithdrawalCredentials to the fixed size [32]byte
	var withdrawalCredentialsFixed [32]byte
	copy(withdrawalCredentialsFixed[:], withdrawalCredentials[:])

	val := megapool.ProvedValidator{
		Pubkey:                     validatorPubkey[:],
		WithdrawalCredentials:      withdrawalCredentialsFixed,
		EffectiveBalance:           validators[validatorIndex64].EffectiveBalance,
		Slashed:                    validators[validatorIndex64].Slashed,
		ActivationEligibilityEpoch: validators[validatorIndex64].ActivationEligibilityEpoch,
		ActivationEpoch:            validators[validatorIndex64].ActivationEpoch,
		ExitEpoch:                  validators[validatorIndex64].ExitEpoch,
		WithdrawableEpoch:          validators[validatorIndex64].WithdrawableEpoch,
	}
	proof := megapool.ValidatorProof{
		ValidatorIndex: big.NewInt(int64(validatorIndex64)),
		Validator:      val,
		Witnesses:      proofWithFixedSize,
	}

	slotProof := megapool.SlotProof{
		Slot:      beaconState.GetSlot(),
		Witnesses: ConvertToFixedSize(slotProofBytes),
	}

	slotTimestamp, err := GetChildBlockTimestampForSlot(c, beaconState.GetSlot())
	if err != nil {
		return megapool.ValidatorProof{}, 0, megapool.SlotProof{}, fmt.Errorf("Error getting the slotTimestamp: %w", err)
	}

	return proof, slotTimestamp, slotProof, err
}

func GetWithdrawableEpochProof(c *cli.Context, wallet *wallet.Wallet, eth2Config beacon.Eth2Config, megapoolAddress common.Address, validatorPubkey types.ValidatorPubkey) (api.ValidatorWithdrawableEpochProof, error) {
	bc, err := GetBeaconClient(c)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	// Get the validator index on the beacon chain
	validatorIndex, err := bc.GetValidatorIndex(validatorPubkey)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	validatorIndex64, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	// Get the head block, requesting the previous one until we have an execution payload
	blockToRequest := "head"
	var block beacon.BeaconBlock
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		block, _, err = bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return api.ValidatorWithdrawableEpochProof{}, err
		}

		if block.HasExecutionPayload {
			break
		}
		if attempts == maxAttempts-1 {
			return api.ValidatorWithdrawableEpochProof{}, fmt.Errorf("failed to find a block with execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", block.Slot-1)
	}

	// Get the beacon state for that slot
	beaconStateResponse, err := bc.GetBeaconStateSSZ(block.Slot)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	beaconState, err := eth2.NewBeaconState(beaconStateResponse.Data, beaconStateResponse.Fork)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	withdrawableEpoch := beaconState.GetValidators()[validatorIndex64].WithdrawableEpoch
	if withdrawableEpoch == math.MaxUint64 {
		return api.ValidatorWithdrawableEpochProof{}, fmt.Errorf("validator %d is not withdrawable", validatorIndex64)
	}

	proofBytes, err := beaconState.ValidatorProof(validatorIndex64)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := ConvertToFixedSize(proofBytes)

	proof := api.ValidatorWithdrawableEpochProof{
		Slot:              block.Slot,
		ValidatorIndex:    new(big.Int).SetUint64(validatorIndex64),
		Pubkey:            validatorPubkey[:],
		WithdrawableEpoch: withdrawableEpoch,
		Witnesses:         proofWithFixedSize,
	}

	return proof, err
}

func ConvertToFixedSize(proofBytes [][]byte) [][32]byte {
	var proofWithFixedSize [][32]byte
	for _, b := range proofBytes {
		if len(b) != 32 {
			panic("each byte slice must be exactly 32 bytes long")
		}
		var arr [32]byte
		copy(arr[:], b)
		proofWithFixedSize = append(proofWithFixedSize, arr)
	}
	return proofWithFixedSize
}

func validateDepositInfo(eth2Config beacon.Eth2Config, depositAmount uint64, pubkey rptypes.ValidatorPubkey, withdrawalCredentials common.Hash, signature rptypes.ValidatorSignature) error {

	// Get the deposit domain based on the eth2 config
	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return err
	}

	// Create the deposit struct
	depositData := new(ethpb.Deposit_Data)
	depositData.Amount = depositAmount
	depositData.PublicKey = pubkey.Bytes()
	depositData.WithdrawalCredentials = withdrawalCredentials.Bytes()
	depositData.Signature = signature.Bytes()

	// Validate the signature
	err = prdeposit.VerifyDepositSignature(depositData, depositDomain)
	return err

}

func CalculateMegapoolWithdrawalCredentials(megapoolAddress common.Address) common.Hash {
	// Convert the address to a uint160 (20 bytes) and then to a uint256 (32 bytes)
	addressBigInt := new(big.Int)
	addressBigInt.SetString(megapoolAddress.Hex()[2:], 16) // Remove the "0x" prefix and convert from hex

	// Shift 0x01 left by 248 bits
	shiftedValue := new(big.Int).Lsh(big.NewInt(0x01), 248)

	// Perform the bitwise OR operation
	result := new(big.Int).Or(shiftedValue, addressBigInt)

	// Convert the result to a 32-byte array (bytes32)
	var bytes32 [32]byte
	resultBytes := result.Bytes()
	copy(bytes32[32-len(resultBytes):], resultBytes)

	return common.BytesToHash(resultBytes)

}

// Get all node megapool details
func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAccount common.Address) (api.MegapoolDetails, error) {

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	// Sync
	var wg errgroup.Group
	details := api.MegapoolDetails{Address: megapoolAddress}

	// Return if megapool isn't deployed
	details.Deployed, err = megapool.GetMegapoolDeployed(rp, nodeAccount, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	if !details.Deployed {
		return details, nil
	}

	// Load the megapool contract
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	details.EffectiveDelegateAddress, err = mega.GetEffectiveDelegate(nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	details.DelegateAddress, err = mega.GetDelegate(nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	// Return if delegate is expired
	details.DelegateExpired, err = mega.GetDelegateExpired(rp, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	if details.DelegateExpired {
		return details, nil
	}

	details.LastDistributionTime, err = mega.GetLastDistributionTime(nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	// Don't calculate the revenue split if there are no staked validators
	if details.LastDistributionTime != 0 {
		wg.Go(func() error {
			var err error
			details.RevenueSplit, err = network.CalculateSplit(rp, details.LastDistributionTime, nil)
			return err
		})
	}
	wg.Go(func() error {
		var err error
		details.NodeShare, err = network.GetCurrentNodeShare(rp, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeDebt, err = mega.GetDebt(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.RefundValue, err = mega.GetRefundValue(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ValidatorCount, err = mega.GetValidatorCount(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ActiveValidatorCount, err = mega.GetActiveValidatorCount(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ExitingValidatorCount, err = mega.GetExitingValidatorCount(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.LockedValidatorCount, err = mega.GetLockedValidatorCount(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.PendingRewards, err = mega.GetPendingRewards(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.UseLatestDelegate, err = mega.GetUseLatestDelegate(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.DelegateExpiry, err = megapool.GetMegapoolDelegateExpiry(rp, details.DelegateAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeExpressTicketCount, err = node.GetExpressTicketCount(rp, nodeAccount, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.AssignedValue, err = mega.GetAssignedValue(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeBond, err = mega.GetNodeBond(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.UserCapital, err = mega.GetUserCapital(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Balances, err = tokens.GetBalances(rp, megapoolAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting megapool %s balances: %w", megapoolAddress.Hex(), err)
		}
		return err
	})
	wg.Go(func() error {
		var err error
		details.PendingRewardSplit, err = mega.CalculatePendingRewards(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ReducedBond, err = protocol.GetReducedBondRaw(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return details, err
	}

	details.BondRequirement, err = node.GetBondRequirement(rp, big.NewInt(int64(details.ActiveValidatorCount)), nil)
	if err != nil {
		return details, err
	}

	details.Validators, err = GetMegapoolValidatorDetails(rp, bc, mega, megapoolAddress, uint32(details.ValidatorCount))
	if err != nil {
		return details, err
	}

	return details, nil
}

func GetMegapoolQueueDetails(rp *rocketpool.RocketPool) (api.QueueDetails, error) {

	// Sync
	var wg errgroup.Group
	queueDetails := api.QueueDetails{}

	wg.Go(func() error {
		var err error
		queueDetails.ExpressQueueLength, err = storage.GetListLength(rp, crypto.Keccak256Hash([]byte("deposit.queue.express")), nil)
		return err
	})
	wg.Go(func() error {
		var err error
		queueDetails.StandardQueueLength, err = storage.GetListLength(rp, crypto.Keccak256Hash([]byte("deposit.queue.standard")), nil)
		return err
	})
	wg.Go(func() error {
		var err error
		queueDetails.QueueIndex, err = rp.RocketStorage.GetUint(nil, crypto.Keccak256Hash([]byte("megapool.queue.index")))
		return err
	})
	wg.Go(func() error {
		var err error
		queueDetails.ExpressQueueRate, err = protocol.GetExpressQueueRate(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return queueDetails, err
	}
	return queueDetails, nil

}

func CalculateRewards(rp *rocketpool.RocketPool, amount *big.Int, nodeAccount common.Address) (api.MegapoolRewardSplitResponse, error) {

	rewards := api.MegapoolRewardSplitResponse{}

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
	if err != nil {
		return api.MegapoolRewardSplitResponse{}, err
	}

	// Return if megapool isn't deployed
	deployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount, nil)
	if err != nil {
		return api.MegapoolRewardSplitResponse{}, err
	}
	if !deployed {
		return rewards, nil
	}

	// Load the megapool contract
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return rewards, err
	}

	// Return if delegate is expired
	delegateExpired, err := mega.GetDelegateExpired(rp, nil)
	if err != nil {
		return api.MegapoolRewardSplitResponse{}, err
	}
	if delegateExpired {
		return rewards, nil
	}

	// Get the rewards split
	rewards.RewardSplit, err = mega.CalculateRewards(amount, nil)
	if err != nil {
		return rewards, err
	}

	return rewards, nil

}

func GetMegapoolValidatorDetails(rp *rocketpool.RocketPool, bc beacon.Client, mp megapool.Megapool, megapoolAddress common.Address, validatorCount uint32) ([]api.MegapoolValidatorDetails, error) {

	details := []api.MegapoolValidatorDetails{}

	var wg errgroup.Group
	var lock sync.Mutex
	var currentEpoch uint64

	queueDetails, err := GetMegapoolQueueDetails(rp)
	if err != nil {
		return details, fmt.Errorf("Error getting the megapool queue details: %w", err)
	}

	head, err := bc.GetBeaconHead()
	if err == nil {
		currentEpoch = head.Epoch
	}

	for i := uint32(0); i < validatorCount; i++ {
		i := i
		wg.Go(func() error {
			validatorDetails, err := mp.GetValidatorInfoAndPubkey(i, nil)
			if err != nil {
				return fmt.Errorf("Error retrieving validator %d details: %v\n", i, err)
			}
			validator := api.MegapoolValidatorDetails{
				ValidatorId:        i,
				PubKey:             types.BytesToValidatorPubkey(validatorDetails.Pubkey),
				LastAssignmentTime: time.Unix(int64(validatorDetails.LastAssignmentTime), 0),
				LastRequestedValue: validatorDetails.LastRequestedValue,
				LastRequestedBond:  validatorDetails.LastRequestedBond,
				DepositValue:       validatorDetails.DepositValue,
				Staked:             validatorDetails.Staked,
				Exited:             validatorDetails.Exited,
				InQueue:            validatorDetails.InQueue,
				InPrestake:         validatorDetails.InPrestake,
				ExpressUsed:        validatorDetails.ExpressUsed,
				Dissolved:          validatorDetails.Dissolved,
				Exiting:            validatorDetails.Exiting,
				Locked:             validatorDetails.Locked,
				ExitBalance:        validatorDetails.ExitBalance,
			}

			// Try to fetch the validator status. If it fails, we assume the first deposit was not processed yet
			validator.BeaconStatus, _ = bc.GetValidatorStatus(validator.PubKey, nil)
			if validator.Staked {
				if currentEpoch > validator.BeaconStatus.ActivationEpoch {
					validator.Activated = true
					validatorIndex, err := strconv.ParseUint(validator.BeaconStatus.Index, 10, 64)
					if err != nil {
						return fmt.Errorf("Error parsing the validator index")
					}
					validator.ValidatorIndex = validatorIndex
					validator.WithdrawableEpoch = validator.BeaconStatus.WithdrawableEpoch
				}
			}

			// Compute the queue position
			if validator.InQueue {
				var queueKey string
				if validator.ExpressUsed {
					queueKey = "deposit.queue.express"
				} else {
					queueKey = "deposit.queue.standard"
				}
				validator.QueuePosition, err = calculatePositionInQueue(rp, queueDetails, megapoolAddress, validator.ValidatorId, queueKey)
				if err != nil {
					return fmt.Errorf("error getting queue position for validator ID %d: %w", validator.ValidatorId, err)
				}
			}
			lock.Lock()
			details = append(details, validator)
			lock.Unlock()
			return nil
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return details, err
	}

	return details, nil

}

func findInQueue(rp *rocketpool.RocketPool, megapoolAddress common.Address, validatorId uint32, queueKey string, indexOffset *big.Int, positionOffset *big.Int) (*big.Int, error) {
	var maxSliceLength = big.NewInt(100)

	slice, err := storage.Scan(rp, crypto.Keccak256Hash([]byte(queueKey)), indexOffset, maxSliceLength, nil)
	if err != nil {
		return nil, err
	}

	for _, entry := range slice.Entries {
		if entry.Receiver == megapoolAddress {
			if entry.ValidatorID == validatorId {
				return positionOffset, err
			}
		}
		positionOffset.Add(positionOffset, big.NewInt(1))
	}
	if slice.NextIndex.Cmp(big.NewInt(0)) == 0 {
		return nil, nil
	} else {
		return findInQueue(rp, megapoolAddress, validatorId, queueKey, slice.NextIndex, positionOffset)
	}

}

func calculatePositionInQueue(rp *rocketpool.RocketPool, queueDetails api.QueueDetails, megapoolAddress common.Address, validatorId uint32, queueKey string) (*big.Int, error) {

	position, err := findInQueue(rp, megapoolAddress, validatorId, queueKey, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return nil, fmt.Errorf("Could not find position in queue %s for validatorId %d: %w", queueKey, validatorId, err)
	}
	if position == nil {
		return nil, nil
	}

	pos := position.Uint64() + 1

	queueIndex := queueDetails.QueueIndex.Uint64()
	expressQueueLength := queueDetails.ExpressQueueLength.Uint64()
	expressQueueRate := queueDetails.ExpressQueueRate
	standardQueueLength := queueDetails.StandardQueueLength.Uint64()

	queueInterval := expressQueueRate + 1

	var overallPosition uint64
	if queueKey == "deposit.queue.express" {
		standardEntriesBefore := (pos + (queueIndex % queueInterval)) / expressQueueRate
		if standardEntriesBefore > standardQueueLength {
			standardEntriesBefore = standardQueueLength
		}
		overallPosition = pos + standardEntriesBefore
	} else {
		expressEntriesbefore := (pos * expressQueueLength) + (expressQueueRate - (queueIndex % queueInterval))
		if expressEntriesbefore > expressQueueLength {
			expressEntriesbefore = expressQueueLength
		}
		overallPosition = pos + expressEntriesbefore
	}

	return new(big.Int).SetUint64(overallPosition), err

}

func GetWithdrawalProofForSlot(c *cli.Context, slot uint64, validatorIndex uint64) (megapool.FinalBalanceProof, uint64, error) {
	// Create a new response
	response := megapool.FinalBalanceProof{}
	response.ValidatorIndex = validatorIndex
	// Get services
	if err := RequireNodeRegistered(c); err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}
	bc, err := GetBeaconClient(c)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}

	// Get the head block, requesting the previous one until we have an execution payload
	blockToRequest := "head"
	var recentBlock beacon.BeaconBlock
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		recentBlock, _, err = bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}

		if recentBlock.HasExecutionPayload {
			break
		}
		if attempts == maxAttempts-1 {
			return megapool.FinalBalanceProof{}, 0, fmt.Errorf("failed to find a block with execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", recentBlock.Slot-1)
	}

	// Find the most recent withdrawal to slot.
	// Keep track of 404s- if we get 24 missing slots in a row, assume we don't have full history.
	notFounds := 0
	var foundWithdrawal bool
	var block eth2.SignedBeaconBlock
	for candidateSlot := slot; candidateSlot <= slot+MAX_WITHDRAWAL_SLOT_DISTANCE; candidateSlot++ {
		// Get the block at the candidate slot.
		blockResponse, found, err := bc.GetBeaconBlockSSZ(candidateSlot)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}
		if !found {
			notFounds++
			if notFounds >= 64 {
				return megapool.FinalBalanceProof{}, 0, fmt.Errorf("2 epochs of missing slots detected. It is likely that the Beacon Client was checkpoint synced after the most recent withdrawal to slot %d, and does not have the history required to generate a withdrawal proof", slot)
			}
			continue
		} else {
			notFounds = 0
		}

		beaconBlock, err := eth2.NewSignedBeaconBlock(blockResponse.Data, blockResponse.Fork)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}

		if !beaconBlock.HasExecutionPayload() {
			continue
		}

		foundWithdrawal := false

		// Check the block for a withdrawal for the given validator index.
		for i, withdrawal := range beaconBlock.Withdrawals() {
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
			block = beaconBlock
			break
		}
	}

	if !foundWithdrawal {
		return megapool.FinalBalanceProof{}, 0, fmt.Errorf("no withdrawal found for validator index %d within %d slots of slot %d", validatorIndex, MAX_WITHDRAWAL_SLOT_DISTANCE, slot)
	}

	// Start by proving from the withdrawal to the block_root
	proof, err := block.ProveWithdrawal(uint64(response.IndexInWithdrawalsArray))
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}

	// Get beacon state
	stateResponse, err := bc.GetBeaconStateSSZ(recentBlock.Slot)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}

	state, err := eth2.NewBeaconState(stateResponse.Data, stateResponse.Fork)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}

	var summaryProof [][]byte

	var stateProof [][]byte
	if response.WithdrawalSlot+generic.SlotsPerHistoricalRoot > recentBlock.Slot {
		stateProof, err = state.BlockRootProof(response.WithdrawalSlot)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}
	} else {
		stateProof, err = state.HistoricalSummaryProof(response.WithdrawalSlot)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}

		// Additionally, we need to prove from the block_root in the historical summary
		// up to the beginning of the above proof, which is the entry in the historical summaries vector.
		blockRootsStateSlot := generic.SlotsPerHistoricalRoot + ((response.WithdrawalSlot / generic.SlotsPerHistoricalRoot) * generic.SlotsPerHistoricalRoot)
		// get the state that has the block roots tree
		blockRootsStateResponse, err := bc.GetBeaconStateSSZ(blockRootsStateSlot)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}
		blockRootsState, err := eth2.NewBeaconState(blockRootsStateResponse.Data, blockRootsStateResponse.Fork)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}
		summaryProof, err = blockRootsState.HistoricalSummaryBlockRootProof(int(response.WithdrawalSlot))
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, err
		}

	}

	withdrawalProof := append(proof, summaryProof...)
	withdrawalProof = append(withdrawalProof, stateProof...)

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := ConvertToFixedSize(withdrawalProof)
	response.Witnesses = proofWithFixedSize

	return response, recentBlock.Slot, nil
}

func GetChildBlockTimestampForSlot(c *cli.Context, slot uint64) (uint64, error) {
	bc, err := GetBeaconClient(c)
	if err != nil {
		return 0, err
	}
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return 0, err
	}

	// Find the timestamp of the child block starting at slot + 1, crawl up to two epochs
	for candidateSlot := slot + 1; candidateSlot <= slot+64; candidateSlot++ {
		_, found, err := bc.GetBeaconBlockSSZ(candidateSlot)
		if err != nil {
			return 0, fmt.Errorf("Error getting the beacon block for slot %d: %w", slot, err)
		}
		if !found {
			continue
		}

		slotTimestamp := uint64(eth2Config.GetSlotTime(candidateSlot).Unix())
		return slotTimestamp, nil
	}

	return 0, fmt.Errorf("Error finding a non-skipped slot for 2 epochs (64 slots) after slot: %d. It is likely that the Beacon Client was checkpoint synced after slot %d", slot, slot)
}
