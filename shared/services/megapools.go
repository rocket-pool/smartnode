package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ssz "github.com/ferranbt/fastssz"
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
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	"github.com/rocket-pool/smartnode/shared/types/eth2/fork/fulu"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	"golang.org/x/sync/errgroup"
)

const MAX_WITHDRAWAL_SLOT_DISTANCE = 144000 // 20 days.

// API URL for the withdrawal proofs (base URL + network + withdrawal slot + validator index)
const apiURL = "https://api.rocketpool.net/%s/withdrawals/proofs/%d/%d/%d"

type Withdrawal struct {
	Index                 uint64         `json:"index"`
	ValidatorIndex        uint64         `json:"validatorIndex"`
	WithdrawalCredentials common.Address `json:"withdrawalCredentials"`
	AmountInGwei          uint64         `json:"amountInGwei"`
}
type WithdrawalProofResponse struct {
	Slot           uint64        `json:"slot"`
	WithdrawalSlot uint64        `json:"withdrawalSlot"`
	WithdrawalNum  uint16        `json:"withdrawalNum"`
	Withdrawal     Withdrawal    `json:"withdrawal"`
	Witnesses      []common.Hash `json:"witnesses"`
}

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
func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, bc beacon.Client, nodeAccount common.Address, opts *bind.CallOpts) (api.MegapoolDetails, error) {

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	// Sync
	var wg errgroup.Group
	details := api.MegapoolDetails{Address: megapoolAddress}

	// Return if megapool isn't deployed
	details.Deployed, err = megapool.GetMegapoolDeployed(rp, nodeAccount, opts)
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

	details.EffectiveDelegateAddress, err = mega.GetEffectiveDelegate(opts)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	details.DelegateAddress, err = mega.GetDelegate(opts)
	if err != nil {
		return api.MegapoolDetails{}, err
	}

	// Return if delegate is expired
	details.DelegateExpired, err = mega.GetDelegateExpired(rp, opts)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	if details.DelegateExpired {
		return details, nil
	}

	details.LastDistributionTime, err = mega.GetLastDistributionTime(opts)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	// Don't calculate the revenue split if there are no staked validators
	if details.LastDistributionTime != 0 {
		wg.Go(func() error {
			var err error
			details.RevenueSplit, err = network.CalculateSplit(rp, details.LastDistributionTime, opts)
			return err
		})
	}
	wg.Go(func() error {
		var err error
		details.NodeShare, err = network.GetCurrentNodeShare(rp, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeDebt, err = mega.GetDebt(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.PendingRewards, err = mega.GetPendingRewards(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.RefundValue, err = mega.GetRefundValue(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ValidatorCount, err = mega.GetValidatorCount(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ActiveValidatorCount, err = mega.GetActiveValidatorCount(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ExitingValidatorCount, err = mega.GetExitingValidatorCount(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.LockedValidatorCount, err = mega.GetLockedValidatorCount(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.PendingRewards, err = mega.GetPendingRewards(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.UseLatestDelegate, err = mega.GetUseLatestDelegate(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.DelegateExpiry, err = megapool.GetMegapoolDelegateExpiry(rp, details.DelegateAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeExpressTicketCount, err = node.GetExpressTicketCount(rp, nodeAccount, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.AssignedValue, err = mega.GetAssignedValue(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeBond, err = mega.GetNodeBond(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.UserCapital, err = mega.GetUserCapital(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.Balances, err = tokens.GetBalances(rp, megapoolAddress, opts)
		if err != nil {
			return fmt.Errorf("error getting megapool %s balances: %w", megapoolAddress.Hex(), err)
		}
		return err
	})
	wg.Go(func() error {
		var err error
		details.PendingRewardSplit, err = mega.CalculatePendingRewards(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		details.ReducedBond, err = protocol.GetReducedBondRaw(rp, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return details, err
	}

	details.BondRequirement, err = node.GetBondRequirement(rp, big.NewInt(int64(details.ActiveValidatorCount)), opts)
	if err != nil {
		return details, err
	}

	details.Validators, err = GetMegapoolValidatorDetails(rp, bc, mega, megapoolAddress, uint32(details.ValidatorCount), opts)
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

func GetMegapoolValidatorDetails(rp *rocketpool.RocketPool, bc beacon.Client, mp megapool.Megapool, megapoolAddress common.Address, validatorCount uint32, opts *bind.CallOpts) ([]api.MegapoolValidatorDetails, error) {

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
	if opts != nil {
		currentEpoch = head.FinalizedEpoch
	}

	for i := uint32(0); i < validatorCount; i++ {
		i := i
		wg.Go(func() error {
			validatorDetails, err := mp.GetValidatorInfoAndPubkey(i, opts)
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

func GetWithdrawalProofForSlotFromAPI(c *cli.Context, finalizedSlot uint64, withdrawalSlot uint64, validatorIndex uint64, network cfgtypes.Network) (megapool.FinalBalanceProof, uint64, error) {

	/*  API calls follow this format:
	    https://api.rocketpool.net/<network>/withdrawals/proofs/<finalized_slot>/<withdrawal_slot>/<validator_index>

	    An example API response:

	    {"slot":2277023,"withdrawalSlot":2244697,"withdrawalNum":5,"withdrawal":{"index":33227778,"validatorIndex":1265923,"withdrawalCredentials":"0xf24c70772544b8ae60af28ddd34f13cc9c428ee7","amountInGwei":31995839800},"witnesses":["0x3410c5feb39b3f702d9b56634136cfb904a1f86c8f367e16fe2d4f969c898b0f","0xe2ba28dfa59acd70227d134be96c2e77e9ea1ccf7d27f94b7e3aff50eb344a53","0xf217264812ed46e949aaa1d8006b05409bbdb1438daa70cef95938140e412062","0x7fb96870d35763e6ab832250c9f8b1091f2383cde611d161b8564437f7910fc4","0x1000000000000000000000000000000000000000000000000000000000000000","0x0000100000000000000000000000000000000000000000000000000000000000","0x3f700c8ef3d83fe50eeb18e4f3df37ad98fe6f380f72c238c90d14c2cb18eefd","0x1085626fa4323b8bbe2acad06da450d8a9edc3cfcc0efe5b2830718a06d66a10","0x9c9bd327030be8c943838fd7083cb0d93d5f9c8408f78f3cd339e0f28e41c370","0xca7e70f57e9b781e937fa905ad6f090076d01db09959e31aa5b181a0d1c8569a","0x8ea7caad7a8238c943f26be0b020fb17ff83b12de0eb5977d763d2a11a21fc2c","0xaa3e128e91825440f40be45202d472ac6b70275a7999d5aa09e7b03a7f10e1d7","0x6dd3b9955d892d92338b19976fd07084bfe88a76c3063482b7f30ee60feb2a58","0xd7a3d1a54f5271259f8a74bad234ebbd73c9ce97727177da2d2a244a114c7852","0x0000000000000000000000000000000000000000000000000000000000000000","0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b","0xdd165f288f4232e65aa7873f0bd8c12cefbea56dc663a30c06d3190922b1fd24","0x4dd9e68fabfffa19a6765e8aa9a1130a4c89aa587b8fe3a94026ca1b8523f679","0x1854ce5a71bcceb3c268c12bbeefbe5668bcb3f4a7e81a4856a9f43c10049859","0xbca9d545006824d4eb061f4e0b4908edffc6a50c25fbed6133e64141b86bc8b3","0xf13045a0e245b3959fa463a60028529f1a558ddf8d971c867d6a15218dd8279a","0xcc50292a5848698d71cae30bf5e6f897c8bef60dfc3e02ad6f1942e2044d8c69","0xec92853cdf50b7de0437d100cdf905e1114edd570f787eacf99fbd597c4300c1","0x982a2bee4c6798f97538986ed8bdf4327aa3ed5559421e693ca1c5513cc6d20b","0x06feb5a3ce20048e0b918726ed5cf15c4a098d4271035b1b58e975fde1cc8381","0xcd846f1c9978a061d010eb7eec8b3bd11a85b791a27e045eaf92f0ec3d7453f7","0xc33805f67b05f7f76a3db10096d2399f19cbf68be4d83e03c44fe8bc4c7f8620","0x8f9d27bbc7ec8fc85f61ac7f3bdbdcf9e7489bc3d55014dc7cdabc79d489cfe4","0xa97d47c50569d287b611735eeeb108004c0a4b64fc35b5c9d28adaa78c42b434","0x25a68400f09df83d1f6db6e9eea1db5b834bb26eb411dddf30f970994a278c8a","0xc80d5f6f2940ee052aac33fe0078ff5b52a64112c08617c8f71313a5e5f6b3de","0x114ef5c4f4bbc789f42d3367156029782916ab844a9ba92b5eac4bc7c0a1e6ae","0x5816ed9efb563c976395c126fd210ab89b7ba9f3b6336f830d4d052833fecbde","0xf9c4052c9567a598edae2a270cf8933b4b24323e49c01e1a3071ec3c56396de8","0xc78009fdf07fc56a11f122370658a353aaa542ed63e44c4bc15ff4cd105ab33c","0xa6c0f739071376d5fae6e5713a77f43bfb900e7e926a272a1fce2c046eb1af15","0x9efde052aa15429fae05bad4d0b1d7c64da64d03d7a1854a588c2cb8430c0d30","0xd88ddfeed400a8755596b21942c1497e114c302e6118290f91e6772976041fa1","0x87eb0ddba57e35f6d286673802a4af5975e22506c7cf4c64bb6be5ee11527f2c","0x92700a3530f06301b0ea071690a88c7723d5ebde9a5505c994351821be155309","0x506d86582d252405b840018792cad2bf1259f1ef5aa5f887e13cb2f0094f51e1","0xffff0ad7e659772f9534c195c815efc4014ef1e1daed4404c06385d11192e92b","0x6cf04127db05441cd833107a52be852868890e4317e6a02ab47683aa75964220","0xb7d05f875f140027ef5118a2247bbb84ce8f2f0f1123623085daf7960c329f5f","0xdf6af5f5bbdb6be9ef8aa618e4bf8073960867171e29676f8b284dea6a08a85e","0xb58d900f5e182e3c50ef74969ea16c7726c549757cc23523c369587da7293784","0xd49a7502ffcfb0340b1d7885688500ca308161a7f96b62df9d083b71fcc8f2bb","0x8fe6b1689256c0d385f42f5bbe2027a22c1996e110ba97c171d3e5948de92beb","0x8d0d63c39ebade8509e0ae3c9c3876fb5fa112be18f905ecacfecb92057603ab","0x95eec8b2e541cad4e91de38385f2e046619f54496c2382cb6cacd5b98c26f5a4","0xf893e908917775b62bff23294dbbe3a1cd8e6cc1c35b4801887b646a6f81f17f","0xcddba7b592e3133393c16194fac7431abf2f5485ed711db282183c819e08ebaa","0x8a8d7fe3af8caa085a7639a832001457dfb9128a8061142ad0335629ff23ff9c","0xfeb3c337d7a51a6fbf00b9e34c52e1c9195c969bd4e7a0bfd51d5c5bed9c1167","0xe71f0aa83cc32edfbefa9f4d3e0174ca85182eec9f3a09f6a6c0df6377a510d7","0x1501000000000000000000000000000000000000000000000000000000000000","0xbb4a0d0000000000000000000000000000000000000000000000000000000000","0x675afa2a5607e4b52f90aa67641f4d2a87d1d37e894912832fcb7879dd43b2bd","0xe0a37ef350f23459204e32961caab1b635c7349b8c21d0b170e64c760af8ccb7","0x95bee184cc5eab4b78d31c50ee452486edfedf473864c7fac86cef1de9b56d21","0x639a4850fe20dbbb573079ef58139feb5f6e7fd4d6e5319c4bd54448402a4c4a","0x91a46d0e5f2f963abc96b45ebb1867df4044298fa87235091e392b6c02aa9584","0x6c9497c4b3fb8b1a7e499db515726638d9468baad37c8b712f57636afb20b40d","0x44ce5bfa4c2c3c9e3f04c74c3cdb56404fb6b7a83bbf386f11e7305dc9aade0d","0x664471f37724d6eff852d4a3bc3b39aa1d3b37334e6184805a5364530433f860"]}/
	*/

	url := fmt.Sprintf(apiURL, network, finalizedSlot, withdrawalSlot, validatorIndex)
	response, err := http.Get(url)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, err
	}
	// unmarshal the response into the WithdrawalProofResponse type
	var withdrawalProofResponse WithdrawalProofResponse
	json.Unmarshal([]byte(body), &withdrawalProofResponse)

	// Convert []common.Hash to [][32]byte
	witnesses := make([][32]byte, len(withdrawalProofResponse.Witnesses))
	for i, w := range withdrawalProofResponse.Witnesses {
		witnesses[i] = [32]byte(w)
	}

	// Convert to fixed size
	return megapool.FinalBalanceProof{
		IndexInWithdrawalsArray: uint(withdrawalProofResponse.WithdrawalNum),
		WithdrawalIndex:         withdrawalProofResponse.Withdrawal.Index,
		WithdrawalAddress:       withdrawalProofResponse.Withdrawal.WithdrawalCredentials,
		WithdrawalSlot:          withdrawalProofResponse.WithdrawalSlot,
		ValidatorIndex:          validatorIndex,
		Amount:                  big.NewInt(0).SetUint64(withdrawalProofResponse.Withdrawal.AmountInGwei),
		Witnesses:               witnesses,
	}, withdrawalProofResponse.Slot, nil
}

func GetWithdrawalProofForSlot(c *cli.Context, slot uint64, validatorIndex uint64) (megapool.FinalBalanceProof, uint64, eth2.BeaconState, error) {
	// Create a new response
	response := megapool.FinalBalanceProof{}
	response.ValidatorIndex = validatorIndex
	// Get services
	if err := RequireNodeRegistered(c); err != nil {
		return megapool.FinalBalanceProof{}, 0, nil, err
	}
	bc, err := GetBeaconClient(c)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, nil, err
	}

	withdrawalSlot, block, indexInWithdrawalsArray, withdrawal, finalizedBlock, err := FindWithdrawalBlockAndArrayPosition(slot, validatorIndex, bc)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, nil, err
	}

	response.WithdrawalSlot = withdrawalSlot
	response.Amount = ConvertWithdrawalAmount(withdrawal.Amount)
	response.Amount = big.NewInt(0).SetUint64(withdrawal.Amount)
	response.IndexInWithdrawalsArray = uint(indexInWithdrawalsArray)
	response.WithdrawalIndex = withdrawal.Index
	response.WithdrawalAddress = withdrawal.Address

	// Start by proving from the withdrawal to the block_root
	withdrawalProof, err := block.ProveWithdrawal(uint64(response.IndexInWithdrawalsArray))
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, nil, err
	}

	// Get beacon state
	stateResponse, err := bc.GetBeaconStateSSZ(finalizedBlock.Slot)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, nil, err
	}

	beaconState, err := eth2.NewBeaconState(stateResponse.Data, stateResponse.Fork)
	if err != nil {
		return megapool.FinalBalanceProof{}, 0, nil, err
	}

	fuluState, ok := beaconState.(*fulu.BeaconState)
	if !ok {
		return megapool.FinalBalanceProof{}, 0, nil, fmt.Errorf("expected fulu.BeaconState, got %T", beaconState)
	}

	// Generate proofs separately
	// 1. Withdrawal proof (withdrawal -> block_root)
	// 2. Block roots proof (block_root -> state)
	// 3. Block header proof (state_root in block header)
	// Final order: [withdrawal, block_roots, block_header]

	var blockRootsProof [][]byte
	var blockHeaderProof [][]byte
	var summaryProof [][]byte
	var historicalSummaryProof [][]byte
	var finalProof [][]byte

	if response.WithdrawalSlot+generic.SlotsPerHistoricalRoot > finalizedBlock.Slot {
		// Recent slot: use block_roots
		// Get the block_roots proof separately
		blockRootsProof, err = beaconState.BlockRootProof(response.WithdrawalSlot)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, err
		}
		blockHeaderProof, err = beaconState.BlockHeaderProof()
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, err
		}

		finalProof = append(finalProof, withdrawalProof...)
		finalProof = append(finalProof, blockRootsProof...)
		finalProof = append(finalProof, blockHeaderProof...)

	} else {
		// Historical slot: use historical_summaries
		// Get historical summary block root proof
		blockRootsStateSlot := generic.SlotsPerHistoricalRoot + ((response.WithdrawalSlot / generic.SlotsPerHistoricalRoot) * generic.SlotsPerHistoricalRoot)
		blockRootsStateResponse, err := bc.GetBeaconStateSSZ(blockRootsStateSlot)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, err
		}
		blockRootsState, err := eth2.NewBeaconState(blockRootsStateResponse.Data, blockRootsStateResponse.Fork)
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, err
		}
		summaryProof, err = blockRootsState.HistoricalSummaryBlockRootProof(int(response.WithdrawalSlot))
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, err
		}

		// Get historical summary proof
		var tree *ssz.Node
		tree, err = fuluState.GetTree()
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, fmt.Errorf("could not get state tree: %w", err)
		}

		// Navigate to the historical_summaries (matching HistoricalSummaryProof logic)
		beaconStateChunkCeil := uint64(64)
		gid := uint64(1)
		gid = gid*beaconStateChunkCeil + generic.BeaconStateHistoricalSummariesFieldIndex
		arrayIndex := (response.WithdrawalSlot / generic.SlotsPerHistoricalRoot)
		gid = gid*2*generic.BeaconStateHistoricalSummariesMaxLength + arrayIndex

		proof, err := tree.Prove(int(gid))
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, fmt.Errorf("could not get proof for historical summary: %w", err)
		}
		historicalSummaryProof = proof.Hashes

		blockHeaderProof, err = fuluState.BlockHeaderProof()
		if err != nil {
			return megapool.FinalBalanceProof{}, 0, nil, err
		}
		// Concatenate in order: [withdrawal, summary_block_root, historical_summary, block_header]
		finalProof = append(finalProof, withdrawalProof...)
		finalProof = append(finalProof, summaryProof...)
		finalProof = append(finalProof, historicalSummaryProof...)
		finalProof = append(finalProof, blockHeaderProof...)
	}

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := ConvertToFixedSize(finalProof)
	response.Witnesses = proofWithFixedSize

	return response, finalizedBlock.Slot, beaconState, nil
}

func ConvertWithdrawalAmount(amount uint64) *big.Int {
	amountBigInt := big.NewInt(int64(amount))

	// amount is in Gwei, but we want wei
	amountBigInt.Mul(amountBigInt, big.NewInt(1e9))
	return amountBigInt
}

func FindWithdrawalBlockAndArrayPosition(slot uint64, validatorIndex uint64, bc beacon.Client) (uint64, eth2.SignedBeaconBlock, int, *generic.Withdrawal, *beacon.BeaconBlock, error) {

	// cant use head here as we need to grab the next slot timestamp
	blockToRequest := "finalized"
	var finalizedBlock beacon.BeaconBlock
	var err error
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		finalizedBlock, _, err = bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return 0, nil, 0, nil, nil, err
		}

		if finalizedBlock.HasExecutionPayload {
			break
		}
		if attempts == maxAttempts-1 {
			return 0, nil, 0, nil, nil, fmt.Errorf("failed to find a block with execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", finalizedBlock.Slot-1)
	}

	// Find the most recent withdrawal to slot.
	// Keep track of 404s- if we get 24 missing slots in a row, assume we don't have full history.
	notFounds := 0
	for candidateSlot := slot; candidateSlot <= slot+MAX_WITHDRAWAL_SLOT_DISTANCE; candidateSlot++ {
		// Get the block at the candidate slot.
		blockResponse, found, err := bc.GetBeaconBlockSSZ(candidateSlot)
		if err != nil {
			return 0, nil, 0, nil, nil, err
		}
		if !found {
			notFounds++
			if notFounds >= 64 {
				return 0, nil, 0, nil, nil, fmt.Errorf("2 epochs of missing slots detected. It is likely that the Beacon Client was checkpoint synced after the most recent withdrawal to slot %d, and does not have the history required to generate a withdrawal proof", slot)
			}
			continue
		} else {
			notFounds = 0
		}

		beaconBlock, err := eth2.NewSignedBeaconBlock(blockResponse.Data, blockResponse.Fork)
		if err != nil {
			return 0, nil, 0, nil, nil, err
		}

		if !beaconBlock.HasExecutionPayload() {
			continue
		}

		// Check the block for a withdrawal for the given validator index.
		for i, withdrawal := range beaconBlock.Withdrawals() {
			if withdrawal.ValidatorIndex != validatorIndex {
				continue
			}

			return candidateSlot, beaconBlock, i, withdrawal, &finalizedBlock, nil
		}
	}
	return 0, nil, 0, nil, nil, fmt.Errorf("no withdrawal found for validator index %d within %d slots of slot %d", validatorIndex, MAX_WITHDRAWAL_SLOT_DISTANCE, slot)
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
