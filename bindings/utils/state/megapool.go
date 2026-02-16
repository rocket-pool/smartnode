package state

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"golang.org/x/sync/errgroup"
)

const (
	megapoolValidatorsBatchSize int = 1000
)

type NativeMegapoolDetails struct {
	Address                  common.Address `json:"address"`
	DelegateAddress          common.Address `json:"delegate"`
	EffectiveDelegateAddress common.Address `json:"effectiveDelegateAddress"`
	Deployed                 bool           `json:"deployed"`
	ValidatorCount           uint32         `json:"validatorCount"`
	ActiveValidatorCount     uint32         `json:"activeValidatorCount"`
	LockedValidatorCount     uint32         `json:"lockedValidatorCount"`
	NodeDebt                 *big.Int       `json:"nodeDebt"`
	RefundValue              *big.Int       `json:"refundValue"`
	DelegateExpiry           uint64         `json:"delegateExpiry"`
	DelegateExpired          bool           `json:"delegateExpired"`
	NodeExpressTicketCount   uint64         `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool           `json:"useLatestDelegate"`
	AssignedValue            *big.Int       `json:"assignedValue"`
	NodeBond                 *big.Int       `json:"nodeBond"`
	UserCapital              *big.Int       `json:"userCapital"`
	BondRequirement          *big.Int       `json:"bondRequirement"`
	EthBalance               *big.Int       `json:"ethBalance"`
	LastDistributionTime     uint64         `json:"lastDistributionTime"`
	PendingRewards           *big.Int       `json:"pendingRewards"`
	NodeQueuedBond           *big.Int       `json:"nodeQueuedBond"`
}

// Get the normalized bond per 32 eth validator
// This is used in treegen to calculate attestation scores
func (m *NativeMegapoolDetails) GetMegapoolBondNormalized() *big.Int {
	if m.ActiveValidatorCount == 0 {
		return big.NewInt(0)
	}
	return big.NewInt(0).Div(m.NodeBond, big.NewInt(int64(m.ActiveValidatorCount)))
}

// Get all megapool validators using the multicaller
func GetAllMegapoolValidators(rp *rocketpool.RocketPool, contracts *NetworkContracts) ([]megapool.ValidatorInfoFromGlobalIndex, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get megapool validators count
	megapoolValidatorsCount, err := megapool.GetValidatorCount(rp, opts)
	if err != nil {
		return []megapool.ValidatorInfoFromGlobalIndex{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	validators := make([]megapool.ValidatorInfoFromGlobalIndex, megapoolValidatorsCount)

	// Run the getters in batches
	count := int(megapoolValidatorsCount)
	for i := 0; i < count; i += megapoolValidatorsBatchSize {
		i := i
		max := i + megapoolValidatorsBatchSize
		if max > count {
			max = count
		}

		for j := i; j < max; j++ {
			j := j // Create a new variable `j` scoped to the loop iteration
			validators[j], err = megapool.GetValidatorInfo(rp, uint32(j), opts)
			if err != nil {
				return nil, fmt.Errorf("error executing GetValidatorInfo with global index %d", j)
			}
		}
	}

	return validators, nil
}

func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, nodeAccount common.Address, opts *bind.CallOpts) (NativeMegapoolDetails, error) {

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}

	// Sync
	var wg errgroup.Group
	details := NativeMegapoolDetails{Address: megapoolAddress}

	// Return if megapool isn't deployed
	details.Deployed, err = megapool.GetMegapoolDeployed(rp, nodeAccount, opts)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}
	if !details.Deployed {
		return details, nil
	}

	// Load the megapool contract
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}

	details.EffectiveDelegateAddress, err = mega.GetEffectiveDelegate(opts)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}
	details.DelegateAddress, err = mega.GetDelegate(opts)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}

	// Return if delegate is expired
	details.DelegateExpired, err = mega.GetDelegateExpired(rp, opts)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}
	if details.DelegateExpired {
		return details, nil
	}

	details.LastDistributionTime, err = mega.GetLastDistributionTime(opts)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}
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
		details.LockedValidatorCount, err = mega.GetLockedValidatorCount(opts)
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
		details.EthBalance, err = rp.Client.BalanceAt(context.Background(), details.Address, opts.BlockNumber)
		return err
	})
	wg.Go(func() error {
		var err error
		details.NodeQueuedBond, err = mega.GetNodeQueuedBond(opts)
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
	return details, nil
}
