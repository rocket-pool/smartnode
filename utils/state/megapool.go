package state

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"golang.org/x/sync/errgroup"
)

const (
	megapoolValidatorsBatchSize int = 1000
)

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

		wg.Go(func() error {
			for j := i; j < max; j++ {
				validators[j], err = megapool.GetValidatorInfo(rp, uint32(j), opts)
				if err != nil {
					return fmt.Errorf("error executing GetValidatorInfo with global index %d", j)
				}
			}

			// var err error
			// mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			// if err != nil {
			// 	return err
			// }
			// for j := i; j < max; j++ {
			// 	mc.AddCall(contracts.RocketMegapoolManager, &validators[j], "getValidatorInfo", big.NewInt(int64(j)))
			// }
			// _, err = mc.FlexibleCall(true, opts)
			// if err != nil {
			// 	return fmt.Errorf("error executing multicall: %w", err)
			// }
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting all megapool validators: %w", err)
	}

	return validators, nil
}
