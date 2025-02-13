package megapool

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/storage"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"golang.org/x/sync/errgroup"
)

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
		details.NodeCapital, err = mega.GetNodeCapital(nil)
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
				Staked:             validatorDetails.Staked,
				Exited:             validatorDetails.Exited,
				InQueue:            validatorDetails.InQueue,
				InPrestake:         validatorDetails.InPrestake,
				ExpressUsed:        validatorDetails.ExpressUsed,
				Dissolved:          validatorDetails.Dissolved,
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

	chunk, err := storage.Scan(rp, crypto.Keccak256Hash([]byte(queueKey)), indexOffset, maxSliceLength, nil)
	if err != nil {
		return nil, err
	}

	for _, entry := range chunk.Entries {
		if entry.Receiver == megapoolAddress {
			if entry.ValidatorID == validatorId {
				return positionOffset, err
			} else {
				positionOffset.Add(positionOffset, big.NewInt(1))
			}
		}
	}
	if chunk.NextIndex.Cmp(big.NewInt(0)) == 0 {
		return nil, err
	} else {
		return findInQueue(rp, megapoolAddress, validatorId, queueKey, chunk.NextIndex, positionOffset)
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
