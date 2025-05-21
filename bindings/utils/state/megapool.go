package state

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/network"
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
	NodeDebt                 *big.Int       `json:"nodeDebt"`
	RefundValue              *big.Int       `json:"refundValue"`
	DelegateExpiry           uint64         `json:"delegateExpiry"`
	DelegateExpired          bool           `json:"delegateExpired"`
	NodeExpressTicketCount   uint64         `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool           `json:"useLatestDelegate"`
	AssignedValue            *big.Int       `json:"assignedValue"`
	NodeBond                 *big.Int       `json:"nodeBond"`
	UserCapital              *big.Int       `json:"userCapital"`
	NodeShare                *big.Int       `json:"nodeShare"`
	BondRequirement          *big.Int       `json:"bondRequirement"`
	EthBalance               *big.Int       `json:"ethBalance"`
	LastDistributionBlock    uint64         `json:"lastDistributionBlock"`
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
			wg.Go(func() error {
				validators[j], err = megapool.GetValidatorInfo(rp, uint32(j), opts)
				if err != nil {
					return fmt.Errorf("error executing GetValidatorInfo with global index %d", j)
				}
				return nil
			})
		}
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting all megapool validators: %w", err)
	}

	return validators, nil
}

func GetNodeMegapoolDetails(rp *rocketpool.RocketPool, nodeAccount common.Address) (NativeMegapoolDetails, error) {

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount, nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}

	// Sync
	var wg errgroup.Group
	details := NativeMegapoolDetails{Address: megapoolAddress}

	// Return if megapool isn't deployed
	details.Deployed, err = megapool.GetMegapoolDeployed(rp, nodeAccount, nil)
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

	details.EffectiveDelegateAddress, err = mega.GetEffectiveDelegate(nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}
	details.DelegateAddress, err = mega.GetDelegate(nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}

	// Return if delegate is expired
	details.DelegateExpired, err = mega.GetDelegateExpired(rp, nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
	}
	if details.DelegateExpired {
		return details, nil
	}

	details.LastDistributionBlock, err = mega.GetLastDistributionBlock(nil)
	if err != nil {
		return NativeMegapoolDetails{}, err
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
		details.EthBalance, err = rp.Client.BalanceAt(context.Background(), details.Address, nil)
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
	return details, nil
}
