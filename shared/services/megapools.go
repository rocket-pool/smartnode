package services

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/storage"
	"github.com/rocket-pool/rocketpool-go/types"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	"golang.org/x/sync/errgroup"
)

func GetStakeValidatorInfo(c *cli.Context, wallet *wallet.Wallet, eth2Config beacon.Eth2Config, megapoolAddress common.Address, validatorPubkey types.ValidatorPubkey) (megapool.ValidatorProof, error) {
	// Get validator private key
	validatorKey, err := wallet.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	withdrawalCredentials := CalculateMegapoolWithdrawalCredentials(megapoolAddress)

	depositAmount := uint64(31e9) // 31 ETH in gwei

	depositData, _, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}
	signature := types.BytesToValidatorSignature(depositData.Signature)

	bc, err := GetBeaconClient(c)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	// Get the validator index on the beacon chain
	validatorIndex, err := bc.GetValidatorIndex(validatorPubkey)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	validatorIndex64, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	err = validateDepositInfo(eth2Config, uint64(depositAmount), validatorPubkey, withdrawalCredentials, signature)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	// Get the finalized block, requesting the next one until we have an execution payload
	blockToRequest := "finalized"
	var block beacon.BeaconBlock
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		block, _, err := bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return megapool.ValidatorProof{}, err
		}

		if block.HasExecutionPayload {
			break
		}
		if attempts == maxAttempts-1 {
			return megapool.ValidatorProof{}, fmt.Errorf("failed to find a block with execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", block.Slot+1)
	}

	// Get the beacon state for that slot
	beaconState, err := bc.GetBeaconState(block.Slot)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	proofBytes, err := beaconState.ValidatorCredentialsProof(validatorIndex64)
	if err != nil {
		return megapool.ValidatorProof{}, err
	}

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := convertToFixedSize(proofBytes)

	proof := megapool.ValidatorProof{
		Slot:                  block.Slot,
		ValidatorIndex:        new(big.Int).SetUint64(validatorIndex64),
		Pubkey:                validatorPubkey[:],
		WithdrawalCredentials: withdrawalCredentials,
		Witnesses:             proofWithFixedSize,
	}

	return proof, err
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

	// Get the finalized block, requesting the next one until we have an execution payload
	blockToRequest := "finalized"
	var block beacon.BeaconBlock
	const maxAttempts = 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		block, _, err := bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return api.ValidatorWithdrawableEpochProof{}, err
		}

		if block.HasExecutionPayload {
			break
		}
		if attempts == maxAttempts-1 {
			return api.ValidatorWithdrawableEpochProof{}, fmt.Errorf("failed to find a block with execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", block.Slot+1)
	}

	// Get the beacon state for that slot
	beaconState, err := bc.GetBeaconState(block.Slot)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	withdrawableEpoch := beaconState.Validators[validatorIndex64].WithdrawableEpoch

	proofBytes, err := beaconState.ValidatorWithdrawableEpochProof(validatorIndex64)
	if err != nil {
		return api.ValidatorWithdrawableEpochProof{}, err
	}

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := convertToFixedSize(proofBytes)

	proof := api.ValidatorWithdrawableEpochProof{
		Slot:              block.Slot,
		ValidatorIndex:    new(big.Int).SetUint64(validatorIndex64),
		Pubkey:            validatorPubkey[:],
		WithdrawableEpoch: withdrawableEpoch,
		Witnesses:         proofWithFixedSize,
	}

	return proof, err
}

func convertToFixedSize(proofBytes [][]byte) [][32]byte {
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

	details.LastDistributionBlock, err = mega.GetLastDistributionBlock(nil)
	if err != nil {
		return api.MegapoolDetails{}, err
	}
	// Don't calculate the revenue split if there are no staked validators
	if details.LastDistributionBlock != 0 {
		wg.Go(func() error {
			var err error
			details.RevenueSplit, err = network.CalculateSplit(rp, details.LastDistributionBlock, nil)
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
			validatorDetails, err := mp.GetValidatorInfo(i, nil)
			if err != nil {
				return fmt.Errorf("Error retrieving validator %d details: %v\n", i, err)
			}
			validator := api.MegapoolValidatorDetails{
				ValidatorId:        i,
				PubKey:             types.BytesToValidatorPubkey(validatorDetails.PubKey),
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
				ValidatorIndex:     validatorDetails.ValidatorIndex,
				ExitBalance:        validatorDetails.ExitBalance,
			}
			if validator.Staked {
				validator.BeaconStatus, err = bc.GetValidatorStatus(validator.PubKey, nil)
				if err != nil {
					return fmt.Errorf("error getting beacon status for validator ID %d: %w", validator.ValidatorId, err)
				}
				if currentEpoch > validator.BeaconStatus.ActivationEpoch {
					validator.Activated = true
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
		return nil, err
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
		return nil, err
	}

	pos := position.Uint64()

	queueIndex := queueDetails.QueueIndex.Uint64()
	expressQueueLength := queueDetails.ExpressQueueLength.Uint64()
	expressQueueRate := queueDetails.ExpressQueueRate
	standardQueueLength := queueDetails.StandardQueueLength.Uint64()

	queueInterval := expressQueueRate + 1

	var overallPosition uint64
	if queueKey == "deposit.queue.express" {
		standardEntriesBefore := (pos + (queueIndex%queueInterval)/expressQueueRate)
		if standardEntriesBefore > standardQueueLength {
			standardEntriesBefore = standardQueueLength
		}
		overallPosition = pos + standardEntriesBefore
	} else {
		expressEntriesbefore := (pos*expressQueueLength + (expressQueueRate - (queueIndex % queueInterval)))
		if expressEntriesbefore > expressQueueLength {
			expressEntriesbefore = expressQueueLength
		}
		overallPosition = pos + expressEntriesbefore
	}

	return new(big.Int).SetUint64(overallPosition), err

}
